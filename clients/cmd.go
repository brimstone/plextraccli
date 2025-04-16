// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package clients

import (
	"log/slog"
	"plextraccli/utils"
	"sort"
	"strings"

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

	return cmd
}

func cmdClients(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	clients, err := p.Clients()
	if err != nil {
		return err
	}

	sort.Slice(clients, func(i, j int) bool { return clients[i].Name < clients[j].Name })

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
