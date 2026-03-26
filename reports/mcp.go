// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package reports

import (
	"context"

	"github.com/brimstone/plextraccli/plextrac"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func MCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "get_reports", Description: "List penetration test reports for a client"}, MCPReports)
}

type MCPReportsInput struct {
	Client string `json:"client" jsonschema:"the name or partial name of the client for which to retrieve reports"`
}

type MCPReportsOutput struct {
	Client  *plextrac.Client   `json:"client"  jsonschema:"the full name of the client"`
	Reports []*plextrac.Report `json:"reports" jsonschema:"the reports for a client"`
}

func MCPReports(ctx context.Context, req *mcp.CallToolRequest, input MCPReportsInput) (*mcp.CallToolResult, MCPReportsOutput, error) {
	var out MCPReportsOutput

	client, reports, _, err := getReports(input.Client)
	if err != nil {
		return nil, out, err
	}

	out.Client = client
	out.Reports = reports

	/*
		data, err := json.Marshal(reports)
		os.WriteFile("/tmp/plextrac.mcp.debug.json", data, 0644)
	*/

	return nil, out, nil
}
