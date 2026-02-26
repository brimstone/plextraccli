// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"errors"
	"fmt"
	"net/http"
)

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

func (f *Finding) AddAssetBulk(value []string) error {
	type assetStruct struct {
		Asset string `json:"asset"`
	}

	var compareRequest struct {
		PastedAssets []string `json:"pastedAssets"`
		UseRawInput  bool     `json:"useRawInput"`
	}

	var compareResponse struct {
		Status         string        `json:"status"`
		ExistingAssets []any         `json:"existingAssets"`
		NewAssets      []assetStruct `json:"newAssets"`
	}
	/*
		var bulkRequest struct {
			Assets []assetStruct `json:"assets"`
		}
	*/

	// compare what's already there
	compareRequest.PastedAssets = value

	path := fmt.Sprintf("v2/client/%d/assets/compare", f.r.c.ID)

	body, err := f.r.ua.apiCall(http.MethodPost, path, compareRequest, &compareResponse)
	if err != nil {
		fmt.Printf("body: %s\n", body)

		return fmt.Errorf("error comparing assets: %w", err)
	}

	fmt.Printf("body: %#v\n", compareResponse.NewAssets)
	// TODO collect asset ids for updating the finding

	// TODO post the difference to add new assets
	path = fmt.Sprintf("v2/client/%d/bulk/assets", f.r.c.ID)

	body, err = f.r.ua.apiCall(http.MethodPost, path, compareRequest, &compareResponse)
	if err != nil {
		fmt.Printf("body: %s\n", body)

		return fmt.Errorf("error comparing assets: %w", err)
	}
	// TODO collect asset ids for updating the finding

	return errors.New("not implemented yet")
}
