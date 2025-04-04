package update

import (
	"fmt"
	"log/slog"
	"plextraccli/version"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
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
}

func cmdUpdate(cmd *cobra.Command, args []string) error {
	slog.Info("Checking and applying update")
	v := semver.MustParse(version.Version)
	latest, err := selfupdate.UpdateSelf(v, "brimstone/plextraccli")
	if err != nil {
		return fmt.Errorf("binary update failed: %w", err)
	}
	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up to date.
		slog.Info("Current binary is the latest version",
			"version", version,
		)
	} else {
		slog.Info("Successfully updated to version",
			"version", latest.Version,
			"notes", latest.ReleaseNotes,
		)
	}
	return nill
}
