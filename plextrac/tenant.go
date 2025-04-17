// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"fmt"
	"net/http"
)

func (ua *UserAgent) updateTags() ([]error, error) {
	var tagResponse struct {
		Count struct {
			TotalDocs int `json:"totalDocs"`
		} `json:"count"`
		Tags []tenantTag `json:"tags"`
	}

	// TODO handle pagination
	path := fmt.Sprintf("v1/tenant/%d/tag?limit=10000", ua.tenantID)

	_, err := ua.apiGet(path, &tagResponse)
	if err != nil {
		return nil, err
	}

	ua.tags = tagResponse.Tags

	return nil, nil
}

func (ua *UserAgent) Tags() []string {
	var tags []string

	// TODO handle this better
	_, _ = ua.updateTags()

	for _, tag := range ua.tags {
		tags = append(tags, tag.Name)
	}

	return tags
}

func (ua *UserAgent) AddTags(tags []string) ([]error, error) {
	return nil, nil
}

func (ua *UserAgent) RemoveTags(tags []string) ([]error, error) {
	var response struct {
		Deleted bool `json:"deleted"`
	}

	warnings, err := ua.updateTags()
	if err != nil {
		return warnings, err
	}

	for _, t := range tags {
		id := ""

		for _, u := range ua.tags {
			if u.Name == t {
				id = u.ID

				break
			}
		}

		if id == "" {
			return warnings, fmt.Errorf("tag %s not found", t)
		}

		path := fmt.Sprintf("v1/tenant/%d/tag/%s", ua.tenantID, id)

		body, err := ua.apiCall(http.MethodDelete, path, nil, &response)
		if err != nil {
			return nil, err
		}

		if !response.Deleted {
			return nil, fmt.Errorf("error deleting tag: %s", body)
		}
	}

	return nil, nil
}

func (ua *UserAgent) SetTags(tags []string) ([]error, error) {
	return nil, nil
}
