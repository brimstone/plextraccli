// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import "errors"

type Asset struct {
	ID    string
	Value string
}

func (f *Finding) Assets() ([]Asset, []error, error) {
	var warnings []error

	warnings, err := f.EnsureFull()
	if err != nil {
		return nil, nil, err
	}

	return f.assets, warnings, nil
}

func (f *Finding) AddAsset(value string) error {
	return errors.New("not implemented yet")
}
