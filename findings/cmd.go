// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package findings

import (
	"errors"
	"fmt"
	"os"
	"plextraccli/plextrac"
	"plextraccli/utils"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultCols = []string{"status", "published", "name"}

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "findings",
		Short: "Manage findings for a report",
		Long:  `Manage findings for a report.`,
		RunE:  cmdFindings,
	}
	// findingsCmd represents the findings command
	cmd.PersistentFlags().String("cols", strings.Join(defaultCols, ","), "Columns to show")

	return cmd
}

func cmdFindings(cmd *cobra.Command, args []string) error {
	p, err := plextrac.New(viper.GetString("username"), viper.GetString("password"), viper.GetString("mfa"), viper.GetString("mfaseed"))
	if err != nil {
		return err
	}
	// Get Client
	clientPartial := viper.GetString("client")
	if clientPartial == "" {
		return errors.New("must specify a client")
	}

	c, err := p.ClientByPartial(clientPartial)
	if err != nil {
		return err
	}
	// Get Report
	reportPartial := viper.GetString("report")
	if reportPartial == "" {
		return errors.New("must specify a report")
	}

	r, _, err := c.ReportByPartial(reportPartial)
	if err != nil {
		return err
	}
	// Get Findings
	findings, warnings, err := r.Findings()
	if err != nil {
		return err
	}

	showCols := utils.AggregateCols(defaultCols, cmd.Flag("cols").Value.String())

	var ensure bool
	if slices.Contains(showCols, "tags") {
		ensure = true
	}

	var rows [][]string

	for _, f := range findings {
		if ensure {
			w, err := f.EnsureFull()
			if err != nil {
				return err
			}

			warnings = append(warnings, w...)
		}

		rows = append(rows, []string{
			f.Published,
			f.Status,
			f.Name,
			strings.Join(f.Tags, ","),
		},
		)
	}

	utils.ShowTable(
		[]string{
			"Status",
			"Published",
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
