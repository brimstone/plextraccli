// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package utils

import (
	"log/slog"
	"plextraccli/plextrac"
	"time"

	"github.com/spf13/viper"
)

var SaveConfigFile string

func NewPlextrac() (*plextrac.UserAgent, []error, error) {
	return plextrac.New(plextrac.NewOptions{
		InstanceURL: viper.GetString("instanceurl"),
		Username:    viper.GetString("username"),
		Password:    viper.GetString("password"),
		MFAToken:    viper.GetString("mfa"),
		MFASeed:     viper.GetString("mfaseed"),
		AuthToken:   viper.GetString("authtoken"),
		OnRenewFunc: func(token string, expires time.Time) error {
			slog.Debug("Got a new token",
				"token", token,
				"expires", expires,
				"config", viper.ConfigFileUsed(),
			)
			v := viper.New()
			v.SetConfigFile(SaveConfigFile)
			err := v.ReadInConfig()
			if err != nil {
				// If there was an error reading the config, then don't worry
				// about saving it and bail early
				//nolint:nilerr
				return nil
			}
			v.Set("authtoken", token)
			err = v.WriteConfig()
			if err != nil {
				return err
			}

			return nil
		},
	})
}
