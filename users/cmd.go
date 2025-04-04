// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package users

import (
	"fmt"
	"plextraccli/plextrac"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "users",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: cmdUsers,
	}
	// usersCmd represents the users command
	cmd.PersistentFlags().String("foo", "", "A help for foo")

	return cmd
}

func cmdUsers(cmd *cobra.Command, args []string) error {
	p, err := plextrac.New(viper.GetString("username"), viper.GetString("password"), viper.GetString("mfa"), viper.GetString("mfaseed"))
	if err != nil {
		return err
	}

	users, err := p.Users()
	if err != nil {
		return err
	}

	for _, r := range users {
		fmt.Printf("%s %s\n",
			r.Name,
			r.Email,
		)
	}
	/*
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "Warning: %#v\n", warning)
		}
	*/
	return nil
}
