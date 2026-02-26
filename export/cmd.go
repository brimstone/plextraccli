// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package export

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/brimstone/plextraccli/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	PTRAC    = "ptrac"
	DOC      = "doc"
	MARKDOWN = "md"
)

var allowedFormats = []string{DOC, PTRAC, MARKDOWN}

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "export",
		Short: "Export reports in various formats",
		Long:  `Export reports in various formats.`,
		RunE:  cmdExport,
	}

	cmd.PersistentFlags().StringP("type", "t", allowedFormats[0], "Format type. One of: "+strings.Join(allowedFormats, ",")+".")
	cmd.PersistentFlags().StringP("out", "o", "", "Output file (default: name of report)")
	cmd.PersistentFlags().StringP("template", "", "", "Export Template name to use (default: template specified in report template)")

	return cmd
}

func cmdExport(cmd *cobra.Command, args []string) error {
	filename := cmd.Flag("out").Value.String()

	format := cmd.Flag("type").Value.String()
	if !slices.Contains(allowedFormats, format) {
		return errors.New("not an allowed export format")
	}

	templateName := cmd.Flag("template").Value.String()

	p, warnings, err := utils.NewPlextrac()
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

	r, warnings2, err := c.ReportByPartial(reportPartial)
	if err != nil {
		return err
	}

	warnings = append(warnings, warnings2...)

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	if filename == "" {
		filename = r.Name
	}

	slog.Debug("Export options",
		"filename", filename,
		"type", format,
	)

	if format != PTRAC && format != DOC && format != MARKDOWN {
		return errors.New("unsupported export format")
	}

	if filename == "-" {
		warnings, err = r.ExportWriter(format, os.Stdout, templateName)
	} else {
		switch format {
		case PTRAC:
			filename = strings.TrimSuffix(filename, ".ptrac")
			filename += ".ptrac"
		case DOC:
			filename = strings.TrimSuffix(filename, ".docx")
			filename += ".docx"
		case MARKDOWN:
			filename = strings.TrimSuffix(filename, ".md")
			filename += ".md"
		default:
			err = errors.New("unsupported export format")
		}

		if err != nil {
			return err
		}

		// check if filename already exists
		// TODO check for force flag
		if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s already exists", filename)
		}

		file, err := os.Create(filename) //nolint:gosec
		if err != nil {
			return fmt.Errorf("while writing file to disk: %w", err)
		}

		switch format {
		case PTRAC:
			warnings, err = r.ExportPtrac(file) //nolint:ineffassign,staticcheck,wastedassign
		case DOC:
			warnings, err = r.ExportDoc(file, templateName) //nolint:ineffassign,staticcheck,wastedassign
		case MARKDOWN:
			warnings, err = r.ExportMarkdown(file) //nolint:ineffassign,staticcheck,wastedassign
		default:
			err = errors.New("unsupported export format") //nolint:ineffassign,staticcheck,wastedassign
		}
	}

	if err != nil {
		return err
	}

	if filename != "-" {
		fmt.Printf("Exporting %q as %s to %s\n", r.Name, format, filename)
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	return nil
}
