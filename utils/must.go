// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package utils

func Must(f func() error) {
	if err := f(); err != nil {
		panic(err)
	}
}
