// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package export

import (
	"errors"
	"fmt"
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
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: cmdExport,
	}
	cmd.PersistentFlags().StringP("type", "t", allowedFormats[0], "Format type. One of: "+strings.Join(allowedFormats, ",")+".")
	cmd.PersistentFlags().StringP("out", "o", "", "Output file")

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
