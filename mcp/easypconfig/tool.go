package easypconfig

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTool adds the v1 configuration description tool to an MCP server.
func RegisterTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolName,
		Description: "Describe four EasyP v1 formats: policy/generation YAML, protobuf.lock JSON Schema, and protobuf.mod text grammar. This reference never reads project files, downloads dependencies, or executes plugins.",
		InputSchema: describeInputSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, input DescribeInput) (*mcp.CallToolResult, DescribeOutput, error) {
		output, err := Describe(input)
		if err != nil {
			return nil, DescribeOutput{}, err
		}
		return nil, output, nil
	})
}
