// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package mcp

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/brimstone/plextraccli/clients"
	"github.com/brimstone/plextraccli/reports"
	"github.com/brimstone/plextraccli/version"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol",
		Long:  `Start a Model Context Protocol server for LLMs`,
		RunE:  cmdMCP,
	}

	return cmd
}

func cmdMCP(cmd *cobra.Command, args []string) error {
	server := mcp.NewServer(&mcp.Implementation{Name: "plextrac", Version: version.Version}, nil)

	clients.MCPTools(server)
	reports.MCPTools(server)

	// Run the server over stdin/stdout, until the client disconnects.
	err := server.Run(context.Background(), &mcp.StdioTransport{})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
