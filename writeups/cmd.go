// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package writeups

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/brimstone/plextraccli/plextrac"
	"github.com/brimstone/plextraccli/utils"
	blackfriday "github.com/russross/blackfriday/v2"
	"github.com/spf13/cobra"
)

var defaultCols = []string{"repo", "title"}
var allCols = []string{"Repo", "Title"}

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "writeups",
		Short: "Manage Writeups",
		Long:  `Manage Writesups for plextrac tenant.`,
		RunE:  cmdWriteups,
	}

	// TODO filter by repository
	cmd.PersistentFlags().String("cols", strings.Join(defaultCols, ","), "Columns to show")

	err := cmd.RegisterFlagCompletionFunc("cols", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return utils.LowerCaseHeaders(allCols), 0
	})
	if err != nil {
		panic(err)
	}

	cmd.PersistentFlags().String("writeup", "", "Writeup of interest")

	// Get
	getCmd := &cobra.Command{
		Use:  "get --writeup",
		RunE: cmdWriteupsGet,
	}
	cmd.AddCommand(getCmd)

	getCmd.AddCommand(&cobra.Command{
		Use: "description --writeup",
		Aliases: []string{
			"des",
			"desc",
		},
		RunE: cmdWriteupsGetDescription,
	})
	getCmd.AddCommand(&cobra.Command{
		Use: "recommendations --writeup",
		Aliases: []string{
			"rec",
			"recs",
		},
		RunE: cmdWriteupsGetRecommendations,
	})
	getCmd.AddCommand(&cobra.Command{
		Use: "references --writeup",
		Aliases: []string{
			"ref",
			"refs",
		},
		RunE: cmdWriteupsGetReferences,
	})

	// TODO Export (to a file)

	// Import
	importCmd := &cobra.Command{
		Use: "import",
		Aliases: []string{
			"set",
		},
		RunE: cmdWriteupsImport,
	}
	cmd.AddCommand(importCmd)

	return cmd
}

func cmdWriteups(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	writeups, err := p.Writeups()
	if err != nil {
		return err
	}

	sort.Slice(writeups, func(i, j int) bool { return writeups[i].Title < writeups[j].Title })

	showCols := utils.AggregateCols(defaultCols, cmd.Flag("cols").Value.String())

	var rows [][]string
	for _, w := range writeups {
		rows = append(rows, []string{
			w.RepositoryName,
			w.Title,
		})
	}

	utils.ShowTable(
		allCols,
		rows,
		showCols,
	)

	return nil
}

func findOneWriteup(cmd *cobra.Command) (*plextrac.Writeup, error) {
	writeupName := cmd.Flag("writeup").Value.String()
	if writeupName == "" {
		return nil, errors.New("writeup must be set")
	}

	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return nil, err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	writeups, err := p.Writeups()
	if err != nil {
		return nil, err
	}

	for _, w := range writeups {
		if w.Title == writeupName {
			return w, nil
		}
	}

	return nil, errors.New("writeup not found")
}
func cmdWriteupsGet(cmd *cobra.Command, args []string) error {
	writeup, err := findOneWriteup(cmd)
	if err != nil {
		return err
	}

	fmt.Printf("# %s\n", writeup.Title)
	fmt.Printf("ID: %s\n\n", writeup.Cuid)

	md, err := htmltomarkdown.ConvertString(writeup.Description)
	if err != nil {
		return err
	}

	fmt.Printf("## Description\n%s\n\n", md)

	md, err = htmltomarkdown.ConvertString(writeup.Recommendations)
	if err != nil {
		return err
	}

	fmt.Printf("## Recommendations\n%s\n\n", md)

	md, err = htmltomarkdown.ConvertString(writeup.References)
	if err != nil {
		return err
	}

	fmt.Printf("## References\n%s\n\n", md)

	return nil
}
func cmdWriteupsGetDescription(cmd *cobra.Command, args []string) error {
	writeup, err := findOneWriteup(cmd)
	if err != nil {
		return err
	}

	md, err := htmltomarkdown.ConvertString(writeup.Description)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", md)

	return nil
}
func cmdWriteupsGetRecommendations(cmd *cobra.Command, args []string) error {
	writeup, err := findOneWriteup(cmd)
	if err != nil {
		return err
	}

	md, err := htmltomarkdown.ConvertString(writeup.Recommendations)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", md)

	return nil
}
func cmdWriteupsGetReferences(cmd *cobra.Command, args []string) error {
	writeup, err := findOneWriteup(cmd)
	if err != nil {
		return err
	}

	md, err := htmltomarkdown.ConvertString(writeup.References)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", md)

	return nil
}

func parseMD(document *blackfriday.Node) (plextrac.Writeup, error) {
	var w plextrac.Writeup

	var location string

	document.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if !entering {
			return blackfriday.GoToNext
		}

		switch node.Type { //nolint:exhaustive
		case blackfriday.Heading:
			//fmt.Printf("Heading: %d\n", node.Level)
			if node.Level == 1 {
				w.Title = string(node.FirstChild.Literal)
			}

			if node.Level == 2 {
				location = string(node.FirstChild.Literal)
			}
		case blackfriday.Paragraph:
			//fmt.Printf("Text for %q: %s\n", location, node.FirstChild.Literal)
			switch location {
			case "":
				// TODO parse out the ID
			case "Description":
				w.Description = string(node.FirstChild.Literal)
			case "Recommendations":
				w.Recommendations = string(node.FirstChild.Literal)
			case "References":
				w.References = string(node.FirstChild.Literal)
			}
		}

		return blackfriday.GoToNext
	})

	return w, nil
}

func cmdWriteupsImport(cmd *cobra.Command, args []string) error {
	// TODO handle importing from a file too
	/*
		writeup, p, err := findOneWriteup(cmd)
		if err != nil {
			return err
		}
	*/
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n\n\n", data)
	output := blackfriday.New().Parse(data)

	writeup, err := parseMD(output)
	if err != nil {
		return err
	}

	fmt.Printf("Writeup: %s\n", writeup.Title)

	// TODO If writeup is found
	// TODO		if --replace is not set
	// TODO			return error writeup already exists
	// TODO		update fields in writeup
	// TODO		save writeup

	// TODO create new writeup with fields
	return nil
}
