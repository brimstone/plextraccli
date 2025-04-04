// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package configure

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "configure",
		Short: "Configure plextraccli",
		Long:  `Configure plextraccli`,
		RunE:  cmdConfigure,
	}

	return cmd
}

// Lifted from viper's source.
func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}

		return home
	}

	return os.Getenv("HOME")
}

func cmdConfigure(cmd *cobra.Command, args []string) error {
	viper.SetConfigFile(filepath.Join(userHomeDir(), ".plextrac.yaml"))
	viper.SetConfigType("yaml")
	fmt.Printf("Writing config to %s\n", viper.ConfigFileUsed())

	err := viper.WriteConfig()
	if err != nil {
		fmt.Println(err)
	}

	return nil
}
