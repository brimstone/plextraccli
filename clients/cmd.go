// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package clients

import (
	"errors"
	"log/slog"
	"sort"
	"strings"

	"github.com/brimstone/plextraccli/plextrac"
	"github.com/brimstone/plextraccli/utils"

	"github.com/spf13/cobra"
)

var defaultCols = []string{"name", "poc", "pocemail", "tags"}

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "clients",
		Short: "Manage clients",
		Long:  `Manage clients for plextrac tenant.`,
		RunE:  cmdClients,
	}

	cmd.PersistentFlags().String("cols", strings.Join(defaultCols, ","), "Columns to show")

	var setCmd = &cobra.Command{
		Use:   "set [client] [flags]",
		Short: "Set client properties",
		Long:  `Set client properties such as description, name, poc, etc.`,
		RunE:  cmdSetClient,
	}
	setCmd.Flags().String("description", "", "Description to set for the client")
	// Additional flags can be added here for other client properties
	cmd.AddCommand(setCmd)

	return cmd
}

func getClients() ([]*plextrac.Client, error) {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return nil, err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	clients, err := p.Clients()
	if err != nil {
		return nil, err
	}

	sort.Slice(clients, func(i, j int) bool { return clients[i].Name < clients[j].Name })

	return clients, err
}

func cmdClients(cmd *cobra.Command, args []string) error {
	clients, err := getClients()
	if err != nil {
		return err
	}

	showCols := utils.AggregateCols(defaultCols, cmd.Flag("cols").Value.String())

	var rows [][]string
	for _, r := range clients {
		rows = append(rows, []string{
			r.Name,
			r.POC,
			r.POCEmail,
			strings.Join(r.Tags(), ","),
		})
	}

	utils.ShowTable(
		[]string{
			"Name",
			"POC",
			"POC Email",
			"Tags",
		},
		rows,
		showCols,
	)

	return nil
}

func cmdSetClient(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("client identifier required")
	}

	clientIdentifier := args[0]

	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	client, err := p.ClientByPartial(clientIdentifier)
	if err != nil {
		return err
	}

	// Handle description flag
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if description != "" {
		_, err = client.SetDescription(description)
		if err != nil {
			return err
		}

		slog.Info("Client description updated successfully", "client", client.Name, "description", description)
	}

	// Additional flags can be handled here for other client properties

	return nil
}
