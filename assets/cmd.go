// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package assets

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/brimstone/plextraccli/plextrac"
	"github.com/brimstone/plextraccli/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "assets",
		Short: "Manage assets of a finding",
		Long:  `Manage assets of a finding.`,
		RunE:  cmdAssets,
	}

	// Add subcommand
	addCmd := &cobra.Command{
		Use:  "add",
		RunE: cmdAssetsAdd,
	}
	addCmd.Flags().StringP("value", "v", "", "Value")
	cmd.AddCommand(addCmd)

	return cmd
}

func assetArgs() (*plextrac.Finding, error) {
	var f *plextrac.Finding

	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return f, err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	// Get Client
	clientPartial := viper.GetString("client")
	if clientPartial == "" {
		return f, errors.New("must specify a client")
	}

	c, err := p.ClientByPartial(clientPartial)
	if err != nil {
		return f, err
	}

	// Get Report
	reportPartial := viper.GetString("report")
	if reportPartial == "" {
		return f, errors.New("must specify a report")
	}

	r, _, err := c.ReportByPartial(reportPartial)
	if err != nil {
		return f, err
	}
	// Get Finding
	findingPartial := viper.GetString("finding")
	if findingPartial == "" {
		return f, errors.New("must specify a finding")
	}

	f, err = r.FindingByPartial(findingPartial)
	if err != nil {
		return f, err
	}

	return f, nil
}

func cmdAssets(cmd *cobra.Command, args []string) error {
	f, err := assetArgs()
	if err != nil {
		return err
	}

	assets, _, err := f.Assets()
	if err != nil {
		return err
	}

	fmt.Printf("Assets:\n")

	for _, v := range assets {
		fmt.Printf("- %s\n", v.Value)
	}

	return nil
}

func cmdAssetsAdd(cmd *cobra.Command, args []string) error {
	f, err := assetArgs()
	if err != nil {
		return err
	}

	_ = f

	fmt.Printf("Got to add %#v\n", args)

	if len(args) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			asset := scanner.Text()
			fmt.Printf("Adding %s\n", asset)

			err = f.AddAsset(asset)
			if err != nil {
				return err
			}
		}

		if scanner.Err() != nil {
			return err
		}
	}

	return nil
}
