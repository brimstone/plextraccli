// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package reports

import (
	"errors"
	"log/slog"
	"plextraccli/utils"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultCols = []string{"status", "startdate", "name"}

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "reports",
		Short: "Manage reports for a client",
		Long:  `Manage reports for a client.`,
		RunE:  cmdReports,
	}
	// reportsCmd represents the reports command
	cmd.PersistentFlags().String("cols", strings.Join(defaultCols, ","), "Columns to show")

	return cmd
}

func cmdReports(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	clientPartial := viper.GetString("client")
	if clientPartial == "" {
		return errors.New("must specify a client")
	}

	c, err := p.ClientByPartial(clientPartial)
	if err != nil {
		return err
	}

	reports, warnings, err := c.Reports()
	if err != nil {
		return err
	}

	showCols := utils.AggregateCols(defaultCols, cmd.Flag("cols").Value.String())

	var rows [][]string
	for _, r := range reports {
		rows = append(rows, []string{
			r.Status,
			r.StartDate.Format(time.DateOnly),
			r.Name,
			strings.Join(r.Tags(), ","),
		})
	}

	utils.ShowTable(
		[]string{
			"Status",
			"Start Date",
			"Name",
			"Tags",
		},
		rows,
		showCols,
	)

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	return nil
}
