// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package lint

import (
	"errors"
	"fmt"
	"os"
	"plextraccli/plextrac"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "lint",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: cmdLint,
	}
	// lintCmd represents the lint command
	// cmd.PersistentFlags().String("foo", "", "A help for foo")
	return cmd
}

func lintReport(report plextrac.Report) []error {
	var errs []error
	// General Details
	if report.StartDate.IsZero() {
		errs = append(errs, errors.New("report is missing Start Date"))
	}

	if report.StopDate.IsZero() {
		errs = append(errs, errors.New("report is missing Stop Date"))
	}

	if report.StartDate.After(report.StopDate) {
		errs = append(errs, errors.New("start date is after Stop date"))
	}

	if report.Status == "Draft" {
		errs = append(errs, errors.New("report is still in draft"))
	}

	if len(report.Tags) == 0 {
		errs = append(errs, errors.New("report has no tags"))
	}

	// Sections
	sections, warnings, err := report.Sections()
	if err != nil {
		errs = append(errs, fmt.Errorf("error parsing sections: %w", err))
	}

	for _, w := range warnings {
		errs = append(errs, fmt.Errorf("warning parsing sections: %w", w))
	}

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}

	// TODO move this out to the config file
	requiredSections := []string{
		"Executive Summary",
		"Project Team",
		"Conclusion",
		"Appendix: Vulnerability Scan Results",
	}

	// TODO move this out to the config file
	for _, t := range report.Tags {
		switch t {
		case "scope_ipt":
			requiredSections = append(requiredSections, "Finding Summary: Internal Network Penetration Test")
			requiredSections = append(requiredSections, "Attack Narrative: Internal Network Penetration Test")

			continue
		case "scope_ept":
			requiredSections = append(requiredSections, "Finding Summary: External Network Penetration Test")
			requiredSections = append(requiredSections, "Attack Narrative: External Network Penetration Test")

			continue
		case "scope_wireless":
			requiredSections = append(requiredSections, "Finding Summary: Wireless Penetration Test")
			requiredSections = append(requiredSections, "Attack Narrative: Wireless Penetration Test")

			continue
		default:
			errs = append(errs, fmt.Errorf("report uses unknown tag: %s", t))
		}
	}

	for _, s := range requiredSections {
		j := -1

		for i, t := range titles {
			if t == s {
				j = i
			}
		}

		if j == -1 {
			errs = append(errs, fmt.Errorf("missing section: %s", s))
		} else {
			titles = append(titles[:j], titles[j+1:]...)
		}
	}

	for _, t := range titles {
		errs = append(errs, fmt.Errorf("extra section: %s", t))
	}

	re := regexp.MustCompile(`.{0,10}n't.{0,10}`)

	for _, s := range sections {
		if slices.Contains(titles, s.Title) {
			continue
		}

		matches := re.FindAllString(s.Content, -1)
		for _, m := range matches {
			errs = append(errs, fmt.Errorf("found contraction in section %q: %q", s.Title, m))
		}
	}

	// TODO Check for required custom fields

	return errs
}

func lintFindings(findings []plextrac.Finding) []error {
	var errs []error

	for _, f := range findings {
		// TODO check for past tense
		// Maybe https://github.com/neurosnap/sentences + https://github.com/jdkato/prose
		// TODO check for heading 3, 2 and 1 cuz those look bad
		// TODO check that there is at least one tag set
		// TODO check that the report has those tags too
		assets, warnings, err := f.Assets()
		if err != nil {
			return []error{err}
		}

		for _, w := range warnings {
			fmt.Printf("Warning: %s\n", w)
		}
		// Lint: Check all findings have assets
		if len(assets) == 0 {
			errs = append(errs, fmt.Errorf("finding %q has no assets", f.Name))
		}
		// Lint: Check all findings have evidence
		// fmt.Printf("Evidence: %#v\n", f.Evidence)
		if f.Evidence == "" {
			errs = append(errs, fmt.Errorf("finding %q has no evidence", f.Name))
		} else {
			if !strings.Contains(f.Evidence, "<figure") {
				errs = append(errs, fmt.Errorf("finding %q has no screenshot for evidence", f.Name))
			}
		}

		if strings.Index(f.Evidence, ".</figcaption>") > 0 {
			errs = append(errs, fmt.Errorf("finding %q has at least one caption ending with a period", f.Name))
		}
	}

	return errs
}

// TODO lint narrative

func cmdLint(cmd *cobra.Command, args []string) error {
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

	errors := lintReport(r)
	if len(errors) > 0 {
		fmt.Printf("Errors with report:\n")

		for _, err := range errors {
			fmt.Printf("- %s\n", err)
		}
	}

	// Get Findings
	findings, _, err := r.Findings()
	if err != nil {
		return err
	}

	errors = lintFindings(findings)
	if len(errors) > 0 {
		fmt.Printf("Errors with findings:\n")

		for _, err := range errors {
			fmt.Printf("- %s\n", err)
		}
	}

	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %#v\n", warning)
	}

	return nil
}
