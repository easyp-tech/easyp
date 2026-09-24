package easypconfig

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTool adds the v1 configuration description tool to an MCP server.
func RegisterTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolName,
		Description: "Describe EasyP v1 configuration files, fields, JSON Schemas, and examples.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input DescribeInput) (*mcp.CallToolResult, DescribeOutput, error) {
		output, err := Describe(input)
		if err != nil {
			return nil, DescribeOutput{}, err
		}
		return nil, output, nil
	})
}
