// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

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
