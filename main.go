// Command mochi-card-mcp is a Model Context Protocol (MCP) server exposing the
// Mochi Cards API as tools for use with Claude and other MCP clients.
//
// It reads the Mochi API key from the MOCHI_API_KEY environment variable and
// serves over stdio.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dudleycarr/mochi-card-mcp/internal/mochi"
	"github.com/dudleycarr/mochi-card-mcp/internal/server"
)

// version is the server version. It can be overridden at build time with
// -ldflags "-X main.version=...".
var version = "1.0.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mochi-card-mcp:", err)
		os.Exit(1)
	}
}

func run() error {
	apiKey := os.Getenv("MOCHI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("MOCHI_API_KEY environment variable is not set; get an API key from https://app.mochi.cards/settings/api")
	}

	client := mochi.NewClient(apiKey)
	srv := server.New(client, version)

	return srv.Run(context.Background(), &mcp.StdioTransport{})
}
