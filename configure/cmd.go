// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package configure

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "configure",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: cmdConfigure,
	}
	// configureCmd represents the configure command
	// cmd.PersistentFlags().String("username", "", "help for foo")
	return cmd
}

func cmdConfigure(cmd *cobra.Command, args []string) error {
	fmt.Printf("username: %s\n", viper.GetString("username"))
	fmt.Printf("password: %s\n", viper.GetString("password"))

	return nil
}
