// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"errors"
	"strings"
)

type clientResponse struct {
	Data  []interface{} `json:"data"`
	DocID []int64       `json:"doc_id"`
	ID    string        `json:"id"`
}

type Client struct {
	ua *UserAgent

	ID   int64
	Name string
}

func (ua *UserAgent) Clients() ([]Client, error) {
	var clientResp []clientResponse
	_, err := ua.apiGet("v1/client/list", &clientResp)

	if err != nil {
		// handle err
		return nil, err
	}

	var clients []Client

	for _, c := range clientResp {
		clients = append(clients, Client{
			ID:   c.DocID[0],
			Name: c.Data[1].(string),
			ua:   ua,
		})
	}

	return clients, nil
}

func (ua *UserAgent) ClientByPartial(partial string) (Client, error) {
	clients, err := ua.Clients()
	if err != nil {
		return Client{}, err
	}

	matches := 0

	var match Client

	for _, c := range clients {
		if strings.Contains(strings.ToLower(c.Name), strings.ToLower(partial)) {
			match = c
			matches++
		}
	}

	if matches == 0 {
		return Client{}, errors.New("client not found")
	}

	if matches > 1 {
		return Client{}, errors.New("multiple clients match")
	}

	return match, nil
}
