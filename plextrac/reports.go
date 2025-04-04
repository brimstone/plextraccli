// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type reportResponse struct {
	ID    int64         `json:"id"`
	DocID []int64       `json:"doc_id"`
	Data  []interface{} `json:"data"`
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

type Report struct {
	ua         *UserAgent
	c          *Client
	findings   []Finding
	full       bool
	sections   []Section
	templateID string

	ID               int64
	CreatedAt        time.Time // 7
	FindingsCount    float64   // 4
	FindingsTemplate string    // 12
	Name             string    // 1
	ReportTemplate   string    // 11
	StartDate        time.Time // 8
	Status           string    // 3
	StopDate         time.Time // 9
	Operators        []string  // 5
	Reviewers        []string  // 6
	Tags             []string  // 10

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

func (c *Client) Reports() ([]Report, []error, error) {
	var warnings []error

	var reportResp []reportResponse

	var reports []Report

	_, err := c.ua.apiGet(fmt.Sprintf("v1/client/%d/reports", c.ID), &reportResp)
	if err != nil {
		return nil, warnings, fmt.Errorf("unable to get reports: %w", err)
	}

	for _, r := range reportResp {
		report := Report{
			ID: r.ID,
			ua: c.ua,
			c:  c,
		}

		// CreatedAt 7
		if c, ok := r.Data[7].(float64); ok {
			report.CreatedAt = time.Unix(int64(c), 0).UTC()
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

		// Name 1
		if n, ok := r.Data[1].(string); ok {
			report.Name = n
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[1] to string for name", r.ID))
		}

		// ReportTemplate 11
		if t, ok := r.Data[11].(string); ok {
			report.ReportTemplate = t
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[11] to string for ReportTemplate", r.ID))
		}

		// StartDate 8
		if s, ok := r.Data[8].(string); ok {
			if t, err := time.Parse("2006-01-02T15:04:05.999Z", s); err == nil {
				report.StartDate = t.UTC()
			} else {
				warnings = append(warnings, fmt.Errorf("%d: can't parse data[8] into date for StartDate: %#v", r.ID, r.Data[8]))
			}
		} else {
			warnings = append(warnings, fmt.Errorf("%d: can't coerce data[8] into string for StartDate: %#v", r.ID, r.Data[8]))
		}
		// Status 3
		if s, ok := r.Data[3].(string); ok {
			report.Status = s
		} else {
			warnings = append(warnings, errors.New("can't coerce data[3] into string for status"))
		}

		// StopDate 9
		if s, ok := r.Data[9].(string); ok {
			if t, err := time.Parse("2006-01-02T15:04:05.999Z", s); err == nil {
				report.StopDate = t.UTC()
			} else {
				warnings = append(warnings, fmt.Errorf("can't parse data[9] into date for StopDate: %#v", r.Data[9]))
			}
		} else {
			warnings = append(warnings, fmt.Errorf("can't coerce data[9] into string for StopDate: %#v", r.Data[9]))
		}

		// TODO Operators 5

		// TODO Reviewers 6

		// TODO Tags 10
		if s, ok := r.Data[10].([]interface{}); ok {
			for i, j := range s {
				if t, ok := j.(string); ok {
					report.Tags = append(report.Tags, t)
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

func (c *Client) ReportByPartial(partial string) (Report, []error, error) {
	reports, warnings, err := c.Reports()

	var match Report

	if err != nil {
		return match, warnings, err
	}

	matches := 0

	for _, r := range reports {
		if strings.Contains(strings.ToLower(r.Name), strings.ToLower(partial)) {
			match = r
			matches++
		}
	}

	if matches == 0 {
		return match, warnings, errors.New("report not found")
	}

	if matches > 1 {
		return match, warnings, errors.New("multiple reports match")
	}

	return match, warnings, nil
}

func (r *Report) EnsureFull() ([]error, error) {
	if r.full {
		return nil, nil
	}

	var warnings []error

	var reportResp fullReportResponse

	json, err := r.c.ua.apiGet(fmt.Sprintf("v1/client/%d/report/%d", r.c.ID, r.ID), &reportResp)
	_ = json

	if err != nil {
		return nil, fmt.Errorf("unable to get reports: %w", err)
	}
	// fmt.Printf("json: %s\n", json)
	// fmt.Printf("report: %#v\n", reportResp)
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
