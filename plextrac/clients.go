// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type Client struct {
	ua   *UserAgent
	tags []string
	full bool
	raw  map[string]interface{}

	ID       int64
	Name     string
	POC      string
	POCEmail string
}

type clientResponse struct {
	Status string `json:"status"`
	Data   []struct {
		ClientID int      `json:"client_id"`
		Name     string   `json:"name"`
		Tags     []string `json:"tags"`
		POC      string   `json:"poc,omitempty"`
		POCEmail string   `json:"poc_email,omitempty"`
		// TODO users is a dict of:
		// useremail: {role, classificationId}
		// role can be any of:
		// - STD_USER
		// - ADMIN
		// - TENANT_0_ROLE_big_ol_long_name_of_custom_role
	} `json:"data"`
	Meta struct {
		Pagination struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
			Total  int `json:"total"`
		} `json:"pagination"`
		Sort []struct {
			By    string `json:"by"`
			Order string `json:"order"`
		} `json:"sort"`
		Filters []struct {
			By    string `json:"by"`
			Value string `json:"value"`
		} `json:"filters"`
	} `json:"meta"`
}

type clientsRequestSort struct {
	By    string `json:"by"`
	Order string `json:"order"`
}
type clientsRequestFilter struct {
	By    string `json:"by"`
	Value string `json:"value"`
}
type clientsRequest struct {
	Pagination struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	} `json:"pagination"`
	Sort    []clientsRequestSort   `json:"sort"`
	Filters []clientsRequestFilter `json:"filters"`
}

func (ua *UserAgent) Clients() ([]*Client, error) {
	var clientResp clientResponse

	var clientsReq clientsRequest
	clientsReq.Pagination.Limit = 1000
	clientsReq.Sort = []clientsRequestSort{
		{
			By:    "name",
			Order: "ASC",
		},
	}
	clientsReq.Filters = []clientsRequestFilter{}

	// TODO handle pagination

	_, err := ua.apiCall(http.MethodPost, "v2/clients", clientsReq, &clientResp)

	if err != nil {
		// handle err
		return nil, err
	}

	var clients []*Client

	for _, c := range clientResp.Data {
		clients = append(clients, &Client{
			ID:       int64(c.ClientID),
			Name:     c.Name,
			POC:      c.POC,
			POCEmail: c.POCEmail,
			tags:     c.Tags,
			ua:       ua,
		})
	}

	return clients, nil
}

func (ua *UserAgent) ClientByPartial(partial string) (*Client, error) {
	clients, err := ua.Clients()
	if err != nil {
		return &Client{}, err
	}

	matches := 0

	var match *Client

	for _, c := range clients {
		if strings.Contains(strings.ToLower(c.Name), strings.ToLower(partial)) {
			match = c
			matches++
		}
	}

	if matches == 0 {
		return nil, errors.New("client not found")
	}

	if matches > 1 {
		return nil, errors.New("multiple clients match")
	}

	return match, nil
}

func (c *Client) EnsureFull() ([]error, error) {
	if c.full {
		return nil, nil
	}

	path := fmt.Sprintf("v1/client/%d", c.ID)

	_, err := c.ua.apiGet(path, &c.raw)
	if err != nil {
		return nil, err
	}

	for _, k := range []string{
		"client_id",
		"cuid",
		"doc_type",
		"licenseKeys",
		"logo",
		"tenant_id",
		"users",
	} {
		delete(c.raw, k)
	}

	c.full = true
	// TODO parse any parts of this we care about into our c struct

	return nil, nil
}

func (c *Client) update() ([]error, error) {
	path := fmt.Sprintf("v1/client/%d", c.ID)

	body, err := c.ua.apiCall(http.MethodPut, path, c.raw, nil)
	if err != nil {
		fmt.Printf("body: %s\n", body)

		return nil, fmt.Errorf("error updating client: %w", err)
	}

	fmt.Printf("body: %s\n", body)

	return nil, nil
}

func (c *Client) Tags() []string {
	return c.tags
}
func (c *Client) AddTags(tags []string) ([]error, error) {
	warnings, err := c.EnsureFull()
	if err != nil {
		return warnings, err
	}

	c.tags = append(c.tags, tags...)
	c.raw["tags"] = c.tags
	warnings2, err := c.update()
	warnings = append(warnings, warnings2...)

	return warnings, err
}
func (c *Client) RemoveTags(tags []string) ([]error, error) {
	warnings, err := c.EnsureFull()
	if err != nil {
		return warnings, err
	}

	c.tags = slices.DeleteFunc(c.tags, func(t string) bool {
		return slices.Contains(tags, t)
	})
	fmt.Printf("tags: %#v\n", c.tags)
	c.raw["tags"] = c.tags
	warnings2, err := c.update()
	warnings = append(warnings, warnings2...)

	return warnings, err
}
func (c *Client) SetTags(tags []string) ([]error, error) {
	warnings, err := c.EnsureFull()
	if err != nil {
		return warnings, err
	}

	c.tags = tags
	c.raw["tags"] = c.tags
	warnings2, err := c.update()
	warnings = append(warnings, warnings2...)

	return warnings, err
}
