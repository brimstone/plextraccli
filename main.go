// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package main

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"plextraccli/assets"
	"plextraccli/clients"
	"plextraccli/configure"
	"plextraccli/export"
	"plextraccli/findings"
	"plextraccli/lint"
	"plextraccli/reports"
	"plextraccli/tags"
	"plextraccli/update"
	"plextraccli/users"
	"plextraccli/utils"
	"strings"

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

	// Report
	rootCmd.PersistentFlags().StringP("report", "r", "", "Partial name of report")

	err = viper.BindPFlag("report", rootCmd.PersistentFlags().Lookup("report"))
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
		viper.SetConfigFile(configFile)

		if viper.ConfigFileUsed() == configFile {
			continue
		}

		if err := viper.MergeInConfig(); err == nil {
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
