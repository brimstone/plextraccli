// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package reports

import (
	"errors"
	"fmt"
	"os"
	"plextraccli/plextrac"
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
	p, err := plextrac.New(viper.GetString("username"), viper.GetString("password"), viper.GetString("mfa"), viper.GetString("mfaseed"))
	if err != nil {
		return err
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
			strings.Join(r.Tags, ","),
		},
		)
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
		fmt.Fprintf(os.Stderr, "Warning: %#v\n", warning)
	}

	return nil
}
