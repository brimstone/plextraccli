// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package clients

import (
	"fmt"
	"os"
	"plextraccli/plextrac"
	"plextraccli/utils"
	"sort"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Cmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "clients",
		Short: "Manage clients",
		Long:  `Manage clients for plextrac tenant.`,
		RunE:  cmdClients,
	}

	return cmd
}

func cmdClients(cmd *cobra.Command, args []string) error {
	p, err := plextrac.New(viper.GetString("username"), viper.GetString("password"), viper.GetString("mfa"), viper.GetString("mfaseed"))
	if err != nil {
		return err
	}

	clients, err := p.Clients()
	if err != nil {
		return err
	}

	sort.Slice(clients, func(i, j int) bool { return clients[i].Name < clients[j].Name })

	if isatty.IsTerminal(os.Stdout.Fd()) {
		t := utils.NewTable()
		t.SetColumns(
			[]utils.TableColumn{
				{Title: "Name"},
			})

		for _, c := range clients {
			t.AddRow([]string{
				c.Name,
			})
		}

		fmt.Println(t.Render())
	} else {
		for _, c := range clients {
			fmt.Printf("%s\n", c.Name)
		}
	}

	return nil
}
