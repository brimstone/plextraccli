// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package export

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"plextraccli/plextrac"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var allowedFormats = []string{"doc", "ptrac"}

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "export",
		Short: "Export reports in various formats",
		Long:  `Export reports in various formats.`,
		RunE:  cmdExport,
	}

	cmd.PersistentFlags().StringP("type", "t", allowedFormats[0], "Format type. One of: "+strings.Join(allowedFormats, ",")+".")
	cmd.PersistentFlags().StringP("out", "o", "", "Output file (default: name of report)")

	return cmd
}

func cmdExport(cmd *cobra.Command, args []string) error {
	filename := cmd.Flag("out").Value.String()

	format := cmd.Flag("type").Value.String()
	if !slices.Contains(allowedFormats, format) {
		return errors.New("not an allowed export format")
	}

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

	r, warnings, err := c.ReportByPartial(reportPartial)
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		fmt.Printf("Warning: %#v\n", warning)
	}

	slog.Debug("Filename",
		"filename", filename,
	)

	if filename == "" {
		filename = r.Name
	}

	switch format {
	case "ptrac":
		filename = strings.TrimSuffix(filename, ".ptrac")
		filename += ".ptrac"
		warnings, err = r.ExportPtrac(filename)
	case "doc":
		filename = strings.TrimSuffix(filename, ".docx")
		filename += ".docx"
		warnings, err = r.ExportDoc(filename)
	default:
		err = errors.New("unsupported export format")
	}

	if err != nil {
		return err
	}

	fmt.Printf("Exporting %q as %s to %s\n", r.Name, format, filename)

	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %#v\n", warning)
	}

	return nil
}
