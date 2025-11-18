// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func (r *Report) ExportWriter(extension string, writer io.Writer, templateName string) ([]error, error) {
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

	_, err = writer.Write([]byte(body))
	if err != nil {
		return nil, fmt.Errorf("while writing file: %w", err)
	}

	return nil, nil
}

func (r *Report) ExportDoc(writer io.Writer, templateName string) ([]error, error) {
	return r.ExportWriter("doc", writer, templateName)
}

// ExportPtrac exports the report in ptrac format to the provided writer.
// It returns any warnings encountered during export and any errors that occurred.
//
// Parameters:
//   - writer: io.Writer to which the ptrac formatted report will be written
//
// Returns:
//   - warnings: slice of warning messages encountered during export
//   - err: error if the export operation fails
func (r *Report) ExportPtrac(writer io.Writer) ([]error, error) {
	// Export the report in ptrac format to a buffer
	var buf bytes.Buffer

	warnings, err := r.ExportWriter("ptrac", &buf, "")
	if err != nil {
		return warnings, err
	}

	// Parse the buffer as JSON
	var jsonData interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
		return warnings, fmt.Errorf("failed to parse ptrac data as JSON: %w", err)
	}

	// Write the JSON data to the provided writer with indentation
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(jsonData); err != nil {
		return warnings, fmt.Errorf("failed to write ptrac data to writer: %w", err)
	}

	return warnings, nil
}
func (r *Report) ExportMarkdown(writer io.Writer) ([]error, error) {
	return r.ExportWriter("md", writer, "")
}
