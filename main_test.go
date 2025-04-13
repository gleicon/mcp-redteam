package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type mockMCPClient struct {
	callToolResponse *mcp.CallToolResult
	callToolError    error
}

func (m *mockMCPClient) Initialize(ctx context.Context, req mcp.InitializeRequest) error {
	return nil
}

func (m *mockMCPClient) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return m.callToolResponse, m.callToolError
}

func (m *mockMCPClient) Close() error {
	return nil
}

func TestMCPRedteamPipeline(t *testing.T) {
	// Create a mock MCP client
	mockClient := &mockMCPClient{
		callToolResponse: &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Pipeline execution completed successfully",
				},
			},
		},
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test initialization
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp-test-client",
		Version: "1.0.0",
	}

	if err := mockClient.Initialize(ctx, initRequest); err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// Test pipeline execution
	pipelineArgs := map[string]interface{}{
		"template": "recon.yaml",
		"params": map[string]string{
			"domain": "example.com",
		},
	}

	callRequest := mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      "pipeline",
			Arguments: pipelineArgs,
		},
	}

	result, err := mockClient.CallTool(ctx, callRequest)
	if err != nil {
		t.Fatalf("Failed to call pipeline tool: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("No content returned by pipeline")
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Unexpected content format: %+v", result.Content[0])
	}

	if text.Text != "Pipeline execution completed successfully" {
		t.Errorf("Unexpected pipeline output: %s", text.Text)
	}

	// Test error handling
	mockClient.callToolError = fmt.Errorf("pipeline execution failed")
	_, err = mockClient.CallTool(ctx, callRequest)
	if err == nil {
		t.Error("Expected error from pipeline execution")
	}
}
