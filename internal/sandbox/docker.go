package sandbox

import (
	"context"
	"fmt"
	"io"
	"strings"

	"gexec-sandbox/internal/api"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func getLanguageImage(language string) string {
	switch strings.ToLower(language) {
	case "python", "py":
		return "python:3.9-slim"
	case "golang", "go":
		return "golang:1.21-alpine"
	default:
		return "python:3.9-slim"
	}
}

func getCommand(language string, filePath string) []string {
	switch strings.ToLower(language) {
	case "python", "py":
		return []string{"python", filePath}
	case "golang", "go":
		return []string{"go", "run", filePath}
	default:
		return []string{"python", filePath}
	}
}

func getExtension(language string) string {
	switch strings.ToLower(language) {
	case "python", "py":
		return ".py"
	case "golang", "go":
		return ".go"
	default:
		return ".py"
	}
}

func RunCodeInSandbox(req api.ExecutionRequest) (api.ExecutionResponse, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	imageName := getLanguageImage(req.Language)
	extension := getExtension(req.Language)
	filePath := "/tmp/main" + extension

	pull, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to pull image: %w", err)
	}
	io.Copy(io.Discard, pull)
	pull.Close()

	execCmd := getCommand(req.Language, filePath)
	fullCmd := fmt.Sprintf("echo '%s' > %s && %s", strings.ReplaceAll(req.SourceCode, "'", "'\\''"), filePath, strings.Join(execCmd, " "))

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           imageName,
		Cmd:             []string{"sh", "-c", fullCmd},
		Tty:             false,
		AttachStdout:    true,
		AttachStderr:    true,
		NetworkDisabled: true,
	}, nil, nil, nil, "")
	if err != nil {
		return api.ExecutionResponse{}, fmt.Errorf("failed to create container: %w", err)
	}

	attachResp, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
		return api.ExecutionResponse{}, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
		return api.ExecutionResponse{}, fmt.Errorf("failed to start container: %w", err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
		return api.ExecutionResponse{}, fmt.Errorf("error waiting for container: %w", err)
	case <-statusCh:
	}

	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
		return api.ExecutionResponse{}, fmt.Errorf("failed to read container output: %w", err)
	}

	inspect, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
		return api.ExecutionResponse{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})

	return api.ExecutionResponse{
		Stdout:   string(output),
		Stderr:   "",
		ExitCode: inspect.State.ExitCode,
		Error:    "",
	}, nil
}
