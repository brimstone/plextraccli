// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package main

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/brimstone/plextraccli/assets"
	"github.com/brimstone/plextraccli/clients"
	"github.com/brimstone/plextraccli/configure"
	"github.com/brimstone/plextraccli/export"
	"github.com/brimstone/plextraccli/findings"
	"github.com/brimstone/plextraccli/lint"
	"github.com/brimstone/plextraccli/narrative"
	"github.com/brimstone/plextraccli/reports"
	"github.com/brimstone/plextraccli/tags"
	"github.com/brimstone/plextraccli/update"
	"github.com/brimstone/plextraccli/users"
	"github.com/brimstone/plextraccli/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	// Setup logger
	var programLevel = new(slog.LevelVar) // Info by default
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

	// check for environment variable now
	if os.Getenv("DEBUG") != "" {
		programLevel.Set(slog.LevelDebug)
	}

	me, err := os.Executable()
	if err != nil {
		panic(err)
	}

	var rootCmd = &cobra.Command{
		Use:   filepath.Base(me),
		Short: "CLI to plextrac.com",
		Long:  `CLI to plextrac.com`,
	}
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringP("instanceurl", "i", "", "InstanceURL")

	err = viper.BindPFlag("instanceurl", rootCmd.PersistentFlags().Lookup("instanceurl"))
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringP("username", "u", "", "Username")

	err = viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringP("password", "p", "", "Password")

	err = viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringP("mfa", "m", "", "MFA value")

	err = viper.BindPFlag("mfa", rootCmd.PersistentFlags().Lookup("mfa"))
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().String("mfaseed", "", "MFA Seed to automatically derive MFA value")

	err = viper.BindPFlag("mfaseed", rootCmd.PersistentFlags().Lookup("mfaseed"))
	if err != nil {
		panic(err)
	}

	// Client
	rootCmd.PersistentFlags().StringP("client", "c", "", "Partial name of client")

	err = viper.BindPFlag("client", rootCmd.PersistentFlags().Lookup("client"))
	if err != nil {
		panic(err)
	}

	err = rootCmd.RegisterFlagCompletionFunc("client", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		p, _, err := utils.NewPlextrac()
		if err != nil {
			// TODO something with err?
			return nil, cobra.ShellCompDirectiveError
		}

		clients, err := p.Clients()
		if err != nil {
			// TODO something with err?
			return nil, cobra.ShellCompDirectiveError
		}

		var clientNames []string
		for _, c := range clients {
			clientNames = append(clientNames, c.Name)
		}

		sort.Strings(clientNames)

		return clientNames, 0
	})
	if err != nil {
		panic(err)
	}

	// Report
	rootCmd.PersistentFlags().StringP("report", "r", "", "Partial name of report")

	err = viper.BindPFlag("report", rootCmd.PersistentFlags().Lookup("report"))
	if err != nil {
		panic(err)
	}

	err = rootCmd.RegisterFlagCompletionFunc("report", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		p, _, err := utils.NewPlextrac()
		if err != nil {
			// TODO something with err?
			return nil, cobra.ShellCompDirectiveError
		}

		clientPartial := viper.GetString("client")
		if clientPartial == "" {
			return nil, cobra.ShellCompDirectiveError
		}

		c, err := p.ClientByPartial(clientPartial)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		reports, _, err := c.Reports()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var reportNames []string
		for _, r := range reports {
			reportNames = append(reportNames, r.Name)
		}

		sort.Strings(reportNames)

		return reportNames, 0
	})
	if err != nil {
		panic(err)
	}

	// Finding
	rootCmd.PersistentFlags().StringP("finding", "f", "", "Partial name of finding")

	err = viper.BindPFlag("finding", rootCmd.PersistentFlags().Lookup("finding"))
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(assets.Cmd())
	rootCmd.AddCommand(clients.Cmd())
	rootCmd.AddCommand(configure.Cmd())
	rootCmd.AddCommand(export.Cmd())
	rootCmd.AddCommand(findings.Cmd())
	rootCmd.AddCommand(lint.Cmd())
	rootCmd.AddCommand(narrative.Cmd())
	rootCmd.AddCommand(reports.Cmd())
	rootCmd.AddCommand(tags.Cmd())
	rootCmd.AddCommand(update.Cmd())
	rootCmd.AddCommand(users.Cmd())

	cobra.OnInitialize(initConfig)

	// Execute adds all child commands to the root command and sets flags appropriately.
	// This is called by main.main(). It only needs to happen once to the rootCmd.

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	// First look at the home directory
	viper.SetConfigName(".plextrac")
	viper.AddConfigPath("$HOME")

	if err := viper.MergeInConfig(); err == nil {
		slog.Debug("Using config",
			"configFile", viper.ConfigFileUsed(),
		)
	}

	if utils.SaveConfigFile == "" && (viper.GetString("password") != "" || viper.GetString("authtoken") != "") {
		utils.SaveConfigFile = viper.ConfigFileUsed()
	}

	configPath := []string{
		string(os.PathSeparator),
	}

	// Walk each parent of pwd starting at / looking for configs that overwride the one in $HOME
	pwd, _ := os.Getwd()
	for _, dir := range strings.Split(pwd, "/") {
		configPath = append(configPath, dir)

		configFile := path.Join(append(configPath, ".plextrac.yaml")...)
		slog.Debug("Searching for config file",
			"path", configFile,
		)

		if viper.ConfigFileUsed() == configFile {
			continue
		}

		viper.SetConfigFile(configFile)

		err := viper.MergeInConfig()
		if err == nil {
			slog.Debug("Also using config",
				"configFile", configFile,
			)
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("PLEXTRAC")

	if err := viper.BindEnv("INSTANCEURL"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("USERNAME"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("PASSWORD"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("MFA"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("MFASEED"); err != nil {
		panic(err)
	}

	if err := viper.MergeInConfig(); err == nil {
		slog.Debug("Using environment config")
	}
}
