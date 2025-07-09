// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package narrative

import (
	"errors"
	"fmt"
	"log/slog"
	"plextraccli/utils"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var formats = []string{
	"txt",
	"html",
	"md",
}

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "narrative",
		Short: "Manage narratives in report",
		Long:  `Manage narratives in a report.`,
		RunE:  cmdNarrative,
	}

	cmd.PersistentFlags().StringP("type", "t", formats[0], "Format type. One of: "+strings.Join(formats, ",")+".")

	return cmd
}

func cmdNarrative(cmd *cobra.Command, args []string) error {
	// Get format
	contentType := cmd.Flag("type").Value.String()

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

	r, _, err := c.ReportByPartial(reportPartial)
	if err != nil {
		return err
	}

	warnings2, err := r.EnsureFull()
	if err != nil {
		return err
	}

	warnings = append(warnings, warnings2...)

	sections, warnings2, err := r.Sections()
	if err != nil {
		return err
	}

	warnings = append(warnings, warnings2...)

	if len(args) == 0 {
		var rows [][]string

		for _, s := range sections {
			rows = append(rows, []string{
				s.Title,
			})
		}

		utils.ShowTable(
			[]string{
				"Title",
			},
			rows,
			[]string{
				"title",
			},
		)
	} else {
		var content string

		for _, s := range sections {
			if strings.Contains(s.Title, args[0]) {
				content = s.Content
			}
		}

		if content == "" {
			return errors.New("narrative section not found or has no content")
		} else {
			switch contentType {
			case "html":
				fmt.Printf("%s\n", content)
			case "txt":
				content = strings.ReplaceAll(content, "<p>", "\n")
				content = strings.ReplaceAll(content, "</p>", "\n")
				content = regexp.MustCompile(`<figure.*?figure>`).ReplaceAllString(content, "")
				content = strings.ReplaceAll(content, "&nbsp;", " ")
				fmt.Printf("%s\n", content)
			case "md":
				fmt.Printf("md\n")
			default:
				fmt.Printf("unsupported content type: %s", contentType)
			}
		}
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	return nil
}
