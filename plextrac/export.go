// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func (r *Report) export(extension string, filename string, templateName string) ([]error, error) {
	/*
		There's actually a way to export a report with an export template
		different than what it specifies. This is really only helpful when
		testing new export templates or changes to existing.

			var templateResponse struct {
				ExportTemplateID string `json:"export_template"`
			}
			// TODO get report template ID
			templateID, warnings, err := r.GetTemplateID()
			fmt.Printf("Report template: %s\n", templateID)
			// TODO get export template ID
			json, err := r.c.ua.apiGet(fmt.Sprintf("v1/tenant/%d/report-template/%s", r.c.ua.GetTenantID(), templateID), &templateResponse)
			fmt.Printf("Export Template ID: %s\n", templateResponse.ExportTemplateID)
			//json, err = r.c.ua.apiGet(fmt.Sprintf("v2/tenant/%d/export-templates", r.c.ua.GetTenantID()), &reportResp)
			//json, err := r.c.ua.apiGet(fmt.Sprintf("v2/tenant/%d/export-templates", r.c.ua.GetTenantID()), &reportResp)
			json, err = r.c.ua.apiGet(fmt.Sprintf("v1/client/%d/report/%d/export/word?includeEvidence=false&templateID=%s", r.c.ID, r.ID, templateResponse.ExportTemplateID), &reportResp)
	*/
	type template struct {
		Name string `json:"name"`
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	var templateResponse map[string]template

	slog.Debug("Starting export")

	templateID := ""

	if templateName != "" {
		_, err := r.c.ua.apiGet(fmt.Sprintf("v2/tenant/%d/export-templates", r.c.ua.GetTenantID()), &templateResponse)
		if err != nil {
			return nil, fmt.Errorf("while calling export-templates api: %w", err)
		}

		for id, t := range templateResponse {
			if strings.Contains(t.Name, templateName) {
				if templateID != "" {
					return nil, errors.New("multiple export template names matched")
				}

				templateID = id
			}
		}
	}

	url := fmt.Sprintf("v1/client/%d/report/%d/export/%s?includeEvidence=false", r.c.ID, r.ID, extension)
	if templateID != "" {
		url += "&templateID=" + templateID
	}

	body, err := r.c.ua.apiGet(url, nil)
	if err != nil {
		return nil, fmt.Errorf("while calling export api: %w", err)
	}

	// check if filename already exists
	if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s already exists", filename)
	}

	err = os.WriteFile(filename, []byte(body), 0600)
	if err != nil {
		return nil, fmt.Errorf("while writing file to disk: %w", err)
	}

	return nil, nil
}

func (r *Report) ExportDoc(filename string, templateName string) ([]error, error) {
	return r.export("doc", filename, templateName)
}
func (r *Report) ExportPtrac(filename string) ([]error, error) {
	return r.export("ptrac", filename, "")
}
func (r *Report) ExportMarkdown(filename string) ([]error, error) {
	return r.export("md", filename, "")
}
