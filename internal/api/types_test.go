package api

import (
	"encoding/json"
	"testing"
)

func TestExecutionTypesRoundTripJSON(t *testing.T) {
	req := ExecutionRequest{
		Language:   "python",
		SourceCode: "print('hello')",
		Stdin:      "input",
		TimeoutMS:  1234,
	}

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var gotReq ExecutionRequest
	if err := json.Unmarshal(raw, &gotReq); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if gotReq.Language != req.Language || gotReq.SourceCode != req.SourceCode || gotReq.Stdin != req.Stdin || gotReq.TimeoutMS != req.TimeoutMS {
		t.Fatalf("round trip request = %+v, want %+v", gotReq, req)
	}

	resp := ExecutionResponse{
		Stdout:   "ok",
		Stderr:   "",
		ExitCode: 0,
		Error:    "",
	}

	raw, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() response error = %v", err)
	}

	var gotResp ExecutionResponse
	if err := json.Unmarshal(raw, &gotResp); err != nil {
		t.Fatalf("json.Unmarshal() response error = %v", err)
	}

	if gotResp != resp {
		t.Fatalf("round trip response = %+v, want %+v", gotResp, resp)
	}
}
