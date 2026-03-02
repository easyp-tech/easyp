package easypconfig

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:         ToolName,
		Description:  "Describe easyp.yaml schema and field usage. Supports full schema or a specific path with examples.",
		InputSchema:  describeInputSchema(),
		OutputSchema: describeOutputSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, input DescribeInput) (*mcp.CallToolResult, DescribeOutput, error) {
		out, err := Describe(input)
		if err != nil {
			return nil, DescribeOutput{}, err
		}
		return nil, out, nil
	})
}
