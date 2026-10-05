package main

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/easyp-tech/easyp/internal/version"
	"github.com/easyp-tech/easyp/mcp/easypconfig"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "easypconfig", Version: version.System()}, nil)
	easypconfig.RegisterTool(server)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
