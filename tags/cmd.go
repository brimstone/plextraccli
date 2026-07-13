// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package tags

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/brimstone/plextraccli/plextrac"
	"github.com/brimstone/plextraccli/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "tags",
		Short: "Manage tags for an object",
		Long:  `Manage tags for an object.`,
		RunE:  cmdTags,
	}
	// Add subcommand
	addCmd := &cobra.Command{
		Use:  "add",
		RunE: cmdTagsAdd,
	}
	cmd.AddCommand(addCmd)
	// Remove subcommand
	removeCmd := &cobra.Command{
		Use: "remove",
		Aliases: []string{
			"rm",
			"delete",
			"del",
		},
		RunE: cmdTagsRemove,
	}
	cmd.AddCommand(removeCmd)
	// Set subcommand
	setCmd := &cobra.Command{
		Use:  "set",
		RunE: cmdTagsSet,
	}
	cmd.AddCommand(setCmd)
	// Search subcommand
	searchCmd := &cobra.Command{
		Use:  "search",
		RunE: cmdTagsSearch,
	}
	cmd.AddCommand(searchCmd)

	return cmd
}

type Tagger interface {
	Tags() []string
	AddTags(tags []string) ([]error, error)
	RemoveTags(tags []string) ([]error, error)
	SetTags(tags []string) ([]error, error)
}

func getTagger(p *plextrac.UserAgent) (Tagger, error) {
	// Get Client
	clientPartial := viper.GetString("client")
	if clientPartial == "" {
		return p, nil
	}

	c, err := p.ClientByPartial(clientPartial)
	if err != nil {
		return nil, err
	}
	// Get Report
	reportPartial := viper.GetString("report")
	if reportPartial == "" {
		return c, nil
	}

	r, _, err := c.ReportByPartial(reportPartial)
	if err != nil {
		return nil, err
	}
	// Get Finding
	findingPartial := viper.GetString("finding")
	if findingPartial == "" {
		return r, nil
	}

	f, err := r.FindingByPartial(findingPartial)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func cmdTags(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	tagger, err := getTagger(p)
	if err != nil {
		return err
	}

	for _, t := range tagger.Tags() {
		fmt.Println(t)
	}

	return nil
}

func cmdTagsAdd(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	tagger, err := getTagger(p)
	if err != nil {
		return err
	}
	// If args is empty, read from stdin until we can't
	tags := args
	if len(tags) == 0 {
		tags, err = utils.StdinToStringSlice()
		if err != nil {
			return err
		}
	}

	warnings, err = tagger.AddTags(tags)
	for _, warning := range warnings {
		slog.Warn(warning.Error())
	}

	return err
}

func cmdTagsRemove(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	tagger, err := getTagger(p)
	if err != nil {
		return err
	}
	// If args is empty, read from stdin until we can't
	tags := args
	if len(tags) == 0 {
		tags, err = utils.StdinToStringSlice()
		if err != nil {
			return err
		}
	}

	warnings, err = tagger.RemoveTags(tags)

	for _, warning := range warnings {
		slog.Warn(warning.Error())
	}

	return err
}

func cmdTagsSet(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	tagger, err := getTagger(p)
	if err != nil {
		return err
	}
	// If args is empty, read from stdin until we can't
	tags := args
	if len(tags) == 0 {
		tags, err = utils.StdinToStringSlice()
		if err != nil {
			return err
		}
	}

	warnings, err = tagger.SetTags(tags)

	for _, warning := range warnings {
		slog.Warn(warning.Error())
	}

	return err
}

func cmdTagsSearch(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("must have exactly one tag")
	}

	tag := args[0]

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
	// TODO can this have a progress bar?
	// TODO maybe count the number of reports a client has, then use that as the total value
	for _, client := range clients {
		for _, t := range client.Tags() {
			if t == tag {
				fmt.Printf("Found tag %s on client %q\n", tag, client.Name)
			}
		}

		reports, _, _ := client.Reports()
		for _, report := range reports {
			for _, t := range report.Tags() {
				if t == tag {
					fmt.Printf("Found tag %s on report %q for client %q\n", tag, report.Name, client.Name)
				}
			}

			findings, _, _ := report.Findings()
			for _, finding := range findings {
				for _, t := range finding.Tags() {
					if t == tag {
						fmt.Printf("Found tag %s on finding %q report %q for client %q\n", tag, finding.Name, report.Name, client.Name)
					}
				}
			}
		}
	}
	// TODO also search WriteupsDB
	return nil
}

// TODO sync tags up from findings -> reports -> clients
