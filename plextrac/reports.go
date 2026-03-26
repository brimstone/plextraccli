// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Report struct {
	ua         *UserAgent
	c          *Client
	findings   []*Finding
	full       bool
	sections   []Section
	templateID string
	raw        map[string]any

	ID               int64     `json:"id"                jsonschema:"Unique report identifier"`
	CreatedAt        time.Time `json:"created_at"        jsonschema:"Report creation timestamp"                upstream:"7"`
	FindingsCount    float64   `json:"findings_count"    jsonschema:"Number of findings in the report"         upstream:"4"`
	FindingsTemplate string    `json:"findings_template" jsonschema:"Template used for findings"               upstream:"12"`
	Name             string    `json:"name"              jsonschema:"Report name/title"                        upstream:"1"`
	ReportTemplate   string    `json:"report_template"   jsonschema:"Template used for the report"             upstream:"11"`
	StartDate        time.Time `json:"start_date"        jsonschema:"Report start date"                        upstream:"8"`
	Status           string    `json:"status"            jsonschema:"Current status of the report"             upstream:"3"`
	StopDate         time.Time `json:"stop_date"         jsonschema:"Report completion date"                   upstream:"9"`
	Operators        []string  `json:"operators"         jsonschema:"List of operators assigned to the report" upstream:"5"`
	Reviewers        []string  `json:"reviewers"         jsonschema:"List of reviewers for the report"         upstream:"6"`
	tags             []string  //`json:"tags"              jsonschema:"Tags associated with the report"          upstream:"10"`
}

type reportResponse struct {
	ID    int64   `json:"id"`
	DocID []int64 `json:"doc_id"`
	Data  []any   `json:"data"`
	// 0: int64 ID
	// 1: string Name of report
	// 2: null
	// 3: enum string Status
	// 4: int64 Findings count
	// 5: []string Operators
	// 6: []string Reviewers
	// 7: int64 created at
	// 8: string start date
	// 9: string stop date
	// 10: []string tags
	// 11: string report template
	// 12: string findings template
	// 13: null?
}

type Section struct {
	ID      string
	Title   string
	Content string
}

type fullReportResponse struct {
	ExecSummary struct {
		CustomFields []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
			Text  string `json:"text"`
		} `json:"custom_fields"`
	} `json:"exec_summary"`
	Template string `json:"template"`
}

func (c *Client) Reports() ([]*Report, []error, error) {
	var warnings []error

	var reportResp []reportResponse

	var reports []*Report

	_, err := c.ua.apiGet(fmt.Sprintf("v1/client/%d/reports", c.ID), &reportResp)
	if err != nil {
		return nil, warnings, fmt.Errorf("unable to get reports: %w", err)
	}

	for _, r := range reportResp {
		report := &Report{
			ID: r.ID,
			ua: c.ua,
			c:  c,
		}

		/*
			for i := range r.Data {
				fmt.Printf("%d: %#v\n", i, r.Data[i])
			}
		*/

		// Name 1
		if n, ok := r.Data[1].(string); ok {
			report.Name = n
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[1] to string for name", r.ID))
		}

		// CreatedAt 7
		if c, ok := r.Data[7].(float64); ok {
			report.CreatedAt = time.Unix(int64(c)/1000, 0).UTC()
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[7] into float64 for CreatedAt: %#v", r.ID, r.Data[7]))
		}

		// Finding 4
		if f, ok := r.Data[4].(float64); ok {
			report.FindingsCount = f
		} else if r.Data[4] == nil {
			report.FindingsCount = 0
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[4] into float64 for findings", r.ID))
		}

		// FindingsTemplate 12
		if f, ok := r.Data[12].(string); ok {
			report.FindingsTemplate = f
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[12] into string for findingstemplate", r.ID))
		}

		// ReportTemplate 11
		if t, ok := r.Data[11].(string); ok {
			report.ReportTemplate = t
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[11] to string for ReportTemplate", r.ID))
		}

		// StartDate 8
		if s, ok := r.Data[8].(string); ok {
			t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
			if err == nil {
				report.StartDate = t.UTC()
			} else {
				warnings = append(warnings, fmt.Errorf("%s: can't parse data[8] into date for StartDate: %#v", report.Name, r.Data[8]))
			}
		} else {
			warnings = append(warnings, fmt.Errorf("%s: can't coerce data[8] into string for StartDate: %#v", report.Name, r.Data[8]))
		}
		// Status 3
		if s, ok := r.Data[3].(string); ok {
			report.Status = s
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[3] into string for status", r.ID))
		}

		// StopDate 9
		if r.Data[9] != nil {
			if s, ok := r.Data[9].(string); ok {
				t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
				if err == nil {
					report.StopDate = t.UTC()
				} else {
					warnings = append(warnings, fmt.Errorf("%d: can't parse data[9] into date for StopDate: %#v", r.ID, r.Data[9]))
				}
			} else {
				warnings = append(warnings, fmt.Errorf("%d: can't coerce data[9] into string for StopDate: %#v", r.ID, r.Data[9]))
			}
		}

		// Operators 5
		if s, ok := r.Data[5].([]any); ok {
			for i, j := range s {
				if o, ok := j.(string); ok {
					report.Operators = append(report.Operators, o)
				} else {
					warnings = append(warnings, fmt.Errorf("can't coerce data[5][%d] into string for slice for Operators: %#v", i, r.Data[5]))
				}
			}
		} else {
			warnings = append(warnings, fmt.Errorf("can't coerce data[5] into string slice for Operator: %#v", r.Data[5]))
		}

		// TODO Reviewers 6

		// Tags 10
		if s, ok := r.Data[10].([]any); ok {
			for i, j := range s {
				if t, ok := j.(string); ok {
					report.tags = append(report.tags, t)
				} else {
					warnings = append(warnings, fmt.Errorf("can't coerce data[10][%d] into string for slice for Tags: %#v", i, r.Data[10]))
				}
			}
		} else {
			warnings = append(warnings, fmt.Errorf("can't coerce data[10] into string slice for Tags: %#v", r.Data[10]))
		}

		reports = append(reports, report)
	}

	return reports, warnings, nil
}

func (c *Client) ReportByPartial(partial string) (*Report, []error, error) {
	reports, warnings, err := c.Reports()

	var (
		match      *Report
		exactMatch *Report
	)

	if err != nil {
		return match, warnings, err
	}

	matches := 0

	for _, r := range reports {
		slog.Debug("Found a match by name",
			"partial", partial,
			"name", r.Name,
		)

		if partial == r.Name {
			exactMatch = r
			match = r
			matches = 1

			break
		}

		if strings.Contains(strings.ToLower(r.Name), strings.ToLower(partial)) {
			match = r
			matches++
		}
	}

	if matches == 0 {
		return nil, warnings, errors.New("report not found")
	}

	if matches > 1 {
		if exactMatch != nil {
			return exactMatch, warnings, nil
		}

		return nil, warnings, errors.New("multiple reports match")
	}

	return match, warnings, nil
}

func (c *Client) ReportByID(id int64) (*Report, []error, error) {
	reports, warnings, err := c.Reports()

	var match *Report

	if err != nil {
		return match, warnings, err
	}

	for _, r := range reports {
		slog.Debug("Checking for match",
			"id1", id,
			"id2", r.ID,
		)

		if id == r.ID {
			match = r

			break
		}
	}

	return match, warnings, nil
}

func (r *Report) EnsureFull() ([]error, error) {
	if r.full {
		return nil, nil
	}

	var warnings []error

	var reportResp fullReportResponse

	_, err := r.c.ua.apiGet(fmt.Sprintf("v1/client/%d/report/%d", r.c.ID, r.ID), &r.raw)
	if err != nil {
		return nil, fmt.Errorf("unable to get reports: %w", err)
	}

	// TODO this is icky. What's a better way?
	jsonReport, err := json.Marshal(r.raw)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal report response back to json: %w", err)
	}

	err = json.Unmarshal(jsonReport, &reportResp)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal report json back to struct: %w", err)
	}

	for _, s := range reportResp.ExecSummary.CustomFields {
		r.sections = append(r.sections, Section{
			ID:      s.ID,
			Title:   s.Label,
			Content: s.Text,
		})
	}

	r.templateID = reportResp.Template

	r.full = true

	return warnings, nil
}

func (r *Report) Sections() ([]Section, []error, error) {
	warnings, err := r.EnsureFull()
	if err != nil {
		return nil, nil, err
	}

	return r.sections, warnings, nil
}

func (r *Report) GetTemplateID() (string, []error, error) {
	warnings, err := r.EnsureFull()
	if err != nil {
		return "", nil, err
	}

	return r.templateID, warnings, err
}

func (r *Report) Tags() []string {
	return r.tags
}

func (r *Report) AddTags(tags []string) ([]error, error) {
	warnings, err := r.EnsureFull()
	if err != nil {
		return warnings, err
	}

	r.tags = append(r.tags, tags...)
	r.raw["tags"] = r.tags

	return r.update()
}
func (r *Report) RemoveTags(tags []string) ([]error, error) {
	warnings, err := r.EnsureFull()
	if err != nil {
		return warnings, err
	}

	r.tags = slices.DeleteFunc(r.tags, func(t string) bool {
		return slices.Contains(tags, t)
	})
	fmt.Printf("tags: %#v\n", r.tags)
	r.raw["tags"] = r.tags

	return r.update()
}
func (r *Report) SetTags(tags []string) ([]error, error) {
	warnings, err := r.EnsureFull()
	if err != nil {
		return warnings, err
	}

	r.tags = tags
	r.raw["tags"] = r.tags

	return r.update()
}
func (r *Report) update() ([]error, error) {
	path := fmt.Sprintf("v1/client/%d/report/%d", r.c.ID, r.ID)

	body, err := r.ua.apiCall(http.MethodPut, path, r.raw, nil)
	if err != nil {
		fmt.Printf("body: %s\n", body)

		return nil, fmt.Errorf("error updating report: %w", err)
	}

	return nil, nil
}
