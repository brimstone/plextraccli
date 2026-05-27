// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package types

type Config struct {
	Username string
	Password string
	TOTP     string
	TOTPSeed string
	Tags     []ConfigTag
}

type ConfigTag struct {
	Name       string
	Narratives []string
}

type LintConfig struct {
	RequiredSections []RequiredSection `mapstructure:"requiredsections"`
}

type RequiredSection struct {
	Section string   `mapstructure:"section"`
	Tags    []string `mapstructure:"tags"`
}
