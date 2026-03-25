// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package plextrac

import "fmt"

type Writeup struct {
	CreatedAt   int64  `json:"createdAt"`
	CreatedBy   int64  `json:"createdBy"`
	Cuid        string `json:"cuid"`
	Description string `json:"description"`
	DocID       int64  `json:"doc_id"`
	DocType     string `json:"doc_type"`
	Fields      struct {
		CustomField1 struct {
			ID        string `json:"id"`
			Key       string `json:"key"`
			Label     string `json:"label"`
			SortOrder int64  `json:"sort_order"`
			Value     string `json:"value"`
		} `json:"custom_field_1"`
		CustomField2 struct {
			ID        string `json:"id"`
			Key       string `json:"key"`
			Label     string `json:"label"`
			SortOrder int64  `json:"sort_order"`
			Value     string `json:"value"`
		} `json:"custom_field_2"`
		CustomFieldsToInfinity struct {
			ID        string `json:"id"`
			Key       string `json:"key"`
			Label     string `json:"label"`
			SortOrder int64  `json:"sort_order"`
			Value     string `json:"value"`
		} `json:"custom_fields_to_infinity"`
		CveData struct {
			ID        string `json:"id"`
			Key       string `json:"key"`
			Label     string `json:"label"`
			SortOrder int64  `json:"sort_order"`
			Value     string `json:"value"`
		} `json:"cve_data"`
		Evidence struct {
			ID        string `json:"id"`
			Key       string `json:"key"`
			Label     string `json:"label"`
			SortOrder int64  `json:"sort_order"`
			Value     string `json:"value"`
		} `json:"evidence"`
		Scores struct {
			Cvss struct {
				Calculation string `json:"calculation"`
				Label       string `json:"label"`
				Type        string `json:"type"`
				Value       string `json:"value"`
			} `json:"cvss"`
			Cvss3 struct {
				Calculation string `json:"calculation"`
				Label       string `json:"label"`
				Type        string `json:"type"`
				Value       string `json:"value"`
			} `json:"cvss3"`
			General struct {
				Calculation string `json:"calculation"`
				Label       string `json:"label"`
				Type        string `json:"type"`
				Value       string `json:"value"`
			} `json:"general"`
		} `json:"scores"`
		Synopsis struct {
			ID        string `json:"id"`
			Key       string `json:"key"`
			Label     string `json:"label"`
			SortOrder int64  `json:"sort_order"`
			Value     string `json:"value"`
		} `json:"synopsis"`
	} `json:"fields"`
	ID                  string `json:"id"`
	IsDeleted           bool   `json:"isDeleted"`
	Recommendations     string `json:"recommendations"`
	References          string `json:"references"`
	RepositoryID        string `json:"repositoryId"`
	RepositoryName      string
	Score               string   `json:"score"`
	Severity            string   `json:"severity"`
	Source              string   `json:"source"`
	Tags                []string `json:"tags"`
	TenantID            int64    `json:"tenantId"`
	Title               string   `json:"title"`
	UpdatedAt           int64    `json:"updatedAt"`
	WriteupAbbreviation string   `json:"writeupAbbreviation"`
}

type writeupRepositoriesResponse struct {
	Data []struct {
		Abbreviation    string `json:"abbreviation"`
		CreatedAt       int64  `json:"createdAt"`
		CreatedBy       any    `json:"createdBy"`
		Description     string `json:"description"`
		DocType         string `json:"doc_type"`
		IsDeleted       bool   `json:"isDeleted"`
		Name            string `json:"name"`
		RepositoryID    string `json:"repositoryId"`
		RepositoryType  string `json:"repositoryType"`
		RepositoryUsers []any  `json:"repositoryUsers"`
		TenantID        int64  `json:"tenantId"`
		UpdatedAt       int64  `json:"updatedAt"`
		WriteupsCount   int64  `json:"writeupsCount"`
	} `json:"data"`
	Status string `json:"status"`
}

type writeupsResponse struct {
	Data   []Writeup `json:"data"`
	Status string    `json:"status"`
}

func (ua *UserAgent) Writeups() ([]*Writeup, error) {
	// TODO filter by repository.
	var writeups []*Writeup

	var writeupsRepositoriesResp writeupRepositoriesResponse

	var writeupsResp writeupsResponse

	_, err := ua.apiCall("POST", "v2/repositories/getAllWriteupsRepositories", struct{}{}, &writeupsRepositoriesResp)

	//fmt.Printf("%s\n", body)

	for _, r := range writeupsRepositoriesResp.Data {
		//fmt.Printf("%s %s\n", r.RepositoryID, r.Name)
		_, err := ua.apiCall("POST", fmt.Sprintf("v2/repositories/%s/getWriteups", r.RepositoryID), struct {
			RepositoryID string `json:"repositoryId"`
		}{
			RepositoryID: r.RepositoryID,
		}, &writeupsResp)

		//fmt.Printf("%s\n", body)

		for _, w := range writeupsResp.Data {
			w.RepositoryName = r.Name
			writeups = append(writeups, &w)
		}

		if err != nil {
			return nil, err
		}
	}

	// TODO v2/repositories/cl52lpwl200060unnf99m3ys3/getWriteups

	return writeups, err
}
