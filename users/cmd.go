// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package users

import (
	"log/slog"
	"plextraccli/utils"

	"github.com/spf13/cobra"
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
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
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
