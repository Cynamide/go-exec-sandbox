package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var (
	containers      = make(map[string]*client.Client)
	containersMutex sync.RWMutex
)

func registerContainer(containerID string, cli *client.Client) {
	containersMutex.Lock()
	defer containersMutex.Unlock()
	containers[containerID] = cli
}

func unregisterContainer(containerID string) {
	containersMutex.Lock()
	defer containersMutex.Unlock()
	delete(containers, containerID)
}

func CleanupAllContainers() {
	containersMutex.RLock()
	defer containersMutex.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for containerID, cli := range containers {
		log.Printf("Cleaning up container: %s", containerID)
		cli.ContainerKill(ctx, containerID, "SIGKILL")
		cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		cli.Close()
	}
}

func getCommand(language string, filePath string, cfg config.Config) []string {
	lowerLang := strings.ToLower(language)
	if strings.HasPrefix(lowerLang, "py") {
		return []string{"python", filePath}
	}
	if strings.HasPrefix(lowerLang, "go") {
		return []string{"go", "run", filePath}
	}
	return []string{language, filePath}
}

func getExtension(language string, cfg config.Config) string {
	lowerLang := strings.ToLower(language)
	if strings.HasPrefix(lowerLang, "py") {
		return ".py"
	}
	if strings.HasPrefix(lowerLang, "go") {
		return ".go"
	}
	return ".txt"
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func readAttachedOutput(reader io.Reader) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	_, err := stdcopy.StdCopy(&stdout, &stderr, reader)
	if err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}

func RunCodeInSandbox(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	imageName, ok := cfg.Languages[req.Language]
	if !ok {
		return api.ExecutionResponse{Error: fmt.Sprintf("unsupported language: %s", req.Language)}, fmt.Errorf("unsupported language: %s", req.Language)
	}
	extension := getExtension(req.Language, cfg)
	filePath := "/tmp/main" + extension

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.TimeoutMS)*time.Millisecond)
	defer cancel()

	pull, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		if ctx.Err() != nil {
			return api.ExecutionResponse{}, fmt.Errorf("execution timed out")
		}
		return api.ExecutionResponse{}, fmt.Errorf("failed to pull image: %w", err)
	}
	defer pull.Close()

	if _, err := io.Copy(io.Discard, pull); err != nil {
		if ctx.Err() != nil {
			return api.ExecutionResponse{}, fmt.Errorf("execution timed out")
		}
		return api.ExecutionResponse{}, fmt.Errorf("failed to read image pull output: %w", err)
	}

	execCmd := getCommand(req.Language, filePath, cfg)
	sourceCmd := fmt.Sprintf("printf %%s %s > %s", shellQuote(req.SourceCode), shellQuote(filePath))
	runCmd := strings.Join(execCmd, " ")
	if req.Stdin != "" {
		runCmd = fmt.Sprintf("printf %%s %s | %s", shellQuote(req.Stdin), runCmd)
	}
	fullCmd := fmt.Sprintf("%s && %s", sourceCmd, runCmd)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           imageName,
		Cmd:             []string{"sh", "-c", fullCmd},
		Tty:             false,
		AttachStdout:    true,
		AttachStderr:    true,
		NetworkDisabled: true,
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory:   int64(cfg.MaxMemoryMB) * 1024 * 1024,
			CPUQuota: 50000,
		},
	}, nil, nil, "")
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID
	registerContainer(containerID, cli)

	defer func() {
		cli.ContainerKill(context.Background(), containerID, "SIGKILL")
		cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})
		unregisterContainer(containerID)
	}()

	attachResp, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to start container: %w", err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return api.ExecutionResponse{}, fmt.Errorf("error waiting for container: %w", err)
	case <-ctx.Done():
		return api.ExecutionResponse{}, fmt.Errorf("execution timed out")
	case <-statusCh:
	}

	stdout, stderr, err := readAttachedOutput(attachResp.Reader)
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to read container output: %w", err)
	}

	inspect, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	return api.ExecutionResponse{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: inspect.State.ExitCode,
		Error:    "",
	}, nil
}
