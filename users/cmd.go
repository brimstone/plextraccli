// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package users

import (
	"plextraccli/plextrac"
	"plextraccli/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "users",
		Short: "Manage users",
		Long:  `Manage users to plextrac tenant.`,
		RunE:  cmdUsers,
	}

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

	var rows [][]string
	for _, user := range users {
		rows = append(rows, []string{
			user.Name,
			user.Email,
		})
	}

	utils.ShowTable(
		[]string{
			"Name",
			"Email",
		},
		rows,
		[]string{
			"name",
			"email",
		},
	)

	return nil
}
