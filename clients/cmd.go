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
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: cmdClients,
	}
	// clientsCmd represents the clients command
	// cmd.PersistentFlags().String("foo", "", "A help for foo")
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
