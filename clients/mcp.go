// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package clients

import (
	"context"
	"log/slog"

	"github.com/brimstone/plextraccli/plextrac"
	"github.com/brimstone/plextraccli/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func MCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "get_all_clients", Description: "List all clients"}, MCPAllClients)
	mcp.AddTool(server, &mcp.Tool{Name: "get_client", Description: "List one specific clients"}, MCPClient)
}

type MCPAllClientsInput struct {
}
type MCPAllClientsOutput struct {
	Clients []*plextrac.Client `json:"clients" jsonschema:"the name of the client"`
}

func MCPAllClients(ctx context.Context, req *mcp.CallToolRequest, input MCPAllClientsInput) (*mcp.CallToolResult, MCPAllClientsOutput, error) {
	var out MCPAllClientsOutput

	clients, err := getClients()
	if err != nil {
		return nil, out, err
	}

	out.Clients = clients

	return nil, out, nil
}

type MCPClientInput struct {
	Name string `json:"name" jsonschema:"the name or partial name of the clients to list"`
}
type MCPClientOutput struct {
	Client *plextrac.Client `json:"client" jsonschema:"the name of the client"`
}

func MCPClient(ctx context.Context, req *mcp.CallToolRequest, input MCPClientInput) (*mcp.CallToolResult, MCPClientOutput, error) {
	var out MCPClientOutput

	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return nil, out, err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	client, err := p.ClientByPartial(input.Name)

	out.Client = client

	return nil, out, err
}
