// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package users

import (
	"errors"
	"fmt"
	"log/slog"
	"plextraccli/plextrac"
	"plextraccli/utils"
	"strings"

	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "users",
		Short: "Manage users",
		Long:  `Manage users to plextrac tenant.`,
		RunE:  cmdUsers,
	}

	cmd.PersistentFlags().StringP("filter", "", "", "Filter users that match this")

	// Reset-password subcommand
	cmdReset := &cobra.Command{
		Use:  "reset-password",
		RunE: cmdUsersReset,
	}

	cmdReset.Flags().StringP("email", "", "", "Email address of user to reset")

	cmd.AddCommand(cmdReset)

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

	filter := cmd.Flag("filter").Value.String()

	users, err := p.Users()
	if err != nil {
		return err
	}

	var rows [][]string

	for _, user := range users {
		if filter == "" || strings.Contains(strings.ToLower(user.Name), strings.ToLower(filter)) ||
			strings.Contains(strings.ToLower(user.Email), strings.ToLower(filter)) {
			rows = append(rows, []string{
				user.Name,
				user.Email,
			})
		}
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

func cmdUsersReset(cmd *cobra.Command, args []string) error {
	p, warnings, err := utils.NewPlextrac()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("Warning while creating plextrac instance",
			"warning", warning,
		)
	}

	email := cmd.Flag("email").Value.String()

	users, err := p.Users()
	if err != nil {
		return err
	}

	var u *plextrac.User

	for _, user := range users {
		if user.Email == email {
			if u != nil {
				return errors.New("more than one user with that email found")
			}

			u = user
		}
	}

	if u == nil {
		return errors.New("user not found")
	}

	warnings, err = u.Reset()
	if err != nil {
		return err
	}

	for _, warning := range warnings {
		slog.Warn("warnings while resetting user password",
			"warning", warning,
		)
	}

	fmt.Printf("User password reset for %s\n", u)

	return nil
}
