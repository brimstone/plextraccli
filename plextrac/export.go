// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func (r *Report) export(extension string, filename string) ([]error, error) {
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
	slog.Debug("Starting export")

	body, err := r.c.ua.apiGet(fmt.Sprintf("v1/client/%d/report/%d/export/%s?includeEvidence=false", r.c.ID, r.ID, extension), nil)
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

func (r *Report) ExportDoc(filename string) ([]error, error) {
	return r.export("doc", filename)
}
func (r *Report) ExportPtrac(filename string) ([]error, error) {
	return r.export("ptrac", filename)
}
