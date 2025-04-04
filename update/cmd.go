// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package update

import (
	"fmt"
	"plextraccli/version"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Update binary from github",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: cmdUpdate,
	}

	return cmd
}

func cmdUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking and applying update")

	v := semver.MustParse(version.Version)

	latest, err := selfupdate.UpdateSelf(v, "brimstone/plextraccli")
	if err != nil {
		return fmt.Errorf("binary update failed: %w", err)
	}

	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up to date.
		fmt.Printf("Current binary is the latest version, %s\n", version.Version)
	} else {
		fmt.Printf("Successfully updated to version %s\n", latest.Version)
		fmt.Println(latest.ReleaseNotes)
	}

	return nil
}
