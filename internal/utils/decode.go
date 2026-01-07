package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

type Header struct {
	AWSRequestID      string `json:"aws-request-id,omitempty"`
	GCPRequestID      string `json:"function-execution-id,omitempty"`
	AZUREInvocationID string `json:"azure-invocation-id,omitempty"`
	ALIBABARequestID  string `json:"ali-request-id,omitempty"`
}

type BenchmarkResponse struct {
	Header Header `json:"header"`
	Body   any    `json:"body"`
}

func DecodeBenchmarkResponse(resp *http.Response) (*BenchmarkResponse, error) {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var b any
	if err := json.Unmarshal(bodyBytes, &b); err != nil {
		return nil, err
	}

	hdr := Header{
		AWSRequestID:      resp.Header.Get("aws-request-id"),
		GCPRequestID:      resp.Header.Get("Function-Execution-Id"),
		AZUREInvocationID: resp.Header.Get("azure-invocation-id"),
		ALIBABARequestID:  resp.Header.Get("ali-request-id"),
	}

	return &BenchmarkResponse{
		Header: hdr,
		Body:   b,
	}, nil
}

func (r *BenchmarkResponse) ToString() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
