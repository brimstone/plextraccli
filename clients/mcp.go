// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package clients

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"

	"github.com/brimstone/plextraccli/plextrac"
	"github.com/brimstone/plextraccli/utils"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func MCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "get_all_clients", Description: "List all clients"}, MCPAllClients)
	mcp.AddTool(server, &mcp.Tool{Name: "get_client", Description: "List one specific client"}, MCPClient)
	mcp.AddTool(server, &mcp.Tool{Name: "set_client_description", Description: "Set client description"}, MCPSetClientDescription(server))
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

type MCPSetClientDescriptionInput struct {
	Name        string `json:"name"        jsonschema:"the name or partial name of the client to update"`
	Description string `json:"description" jsonschema:"the description to set for the client"`
}
type MCPSetClientDescriptionOutput struct {
	Success bool `json:"success" jsonschema:"whether the update was successful"`
}

func MCPSetClientDescription(server *mcp.Server) func(ctx context.Context, req *mcp.CallToolRequest, input MCPSetClientDescriptionInput) (*mcp.CallToolResult, MCPSetClientDescriptionOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input MCPSetClientDescriptionInput) (*mcp.CallToolResult, MCPSetClientDescriptionOutput, error) {
		var out MCPSetClientDescriptionOutput

		configSchema := &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"Description": {Type: "string", Description: "Client Description", ReadOnly: true, Default: json.RawMessage(`"This client..."`)},
			},
			Required: []string{"serverEndpoint"},
		}

		var serverSession *mcp.ServerSession
		for serverSession = range server.Sessions() { // FIXME this feels weird, surely there's a better way
			break
		}

		result, err := serverSession.Elicit(ctx, &mcp.ElicitParams{
			Message:         "Please provide your configuration settings",
			RequestedSchema: configSchema,
		})
		if err != nil {
			log.Fatal(err)
		}

		if result.Action != "accept" {
			return nil, out, err
		}

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
		if err != nil {
			return nil, out, err
		}

		warnings2, err := client.SetDescription(input.Description)
		if err != nil {
			return nil, out, err
		}
		// warnings are currently unused but captured for completeness
		_ = warnings2

		out.Success = true

		return nil, out, nil
	}
}
