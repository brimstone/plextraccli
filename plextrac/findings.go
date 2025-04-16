// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type Finding struct {
	r      *Report
	assets []Asset
	full   bool
	raw    map[string]interface{}

	ID        int
	Status    string
	Name      string
	Published string
	Evidence  string
	tags      []string
}

type findingsResponse struct {
	ID    string   `json:"id"`
	DocID []string `json:"doc_id"`
	Data  []any    `json:"data"`
	// 0 int ID
	// 1 string Severity
	// 2 string Name
	// 3 string Status
	// 4 milliseconds since epoch Updated?
	// 5 null
	// 6 milliseconds since epoch Created?
	// 7 null
	// 8 milliseconds since epoch ???
	// 9 null
	// 10 string Published
	// 11 string empty?
}

func (r *Report) Findings() ([]*Finding, []error, error) {
	var findingsResp []findingsResponse

	var warnings []error

	path := fmt.Sprintf("v1/client/%d/report/%d/flaws", r.c.ID, r.ID)

	_, err := r.ua.apiGet(path, &findingsResp)
	if err != nil {
		return r.findings, nil, err
	}
	//fmt.Printf("Json: %s\n", body) // DEBUG
	for _, f := range findingsResp {
		finding := &Finding{
			r: r,
		}

		// TODO 0 ID
		i := 0
		if v, ok := f.Data[i].(float64); ok {
			finding.ID = int(v)
		} else {
			warnings = append(warnings, fmt.Errorf("couldn't coerce data[%d] %v into an int", i, f.Data[i]))
		}
		// TODO 1 Severity
		// 2 Name
		i = 2
		if n, ok := f.Data[i].(string); ok {
			finding.Name = n
		} else {
			warnings = append(warnings, fmt.Errorf("couldn't coerce data[%d] %v into a string", i, f.Data[i]))
		}

		// 3 Status
		if s, ok := f.Data[3].(string); ok {
			finding.Status = s
		} else {
			warnings = append(warnings, fmt.Errorf("couldn't coerce data[3] %v into a string", f.Data[3]))
		}

		// 4 TODO milliseconds since epoch Updated?
		// 6 TODO milliseconds since epoch Created?
		// 8 string Published
		i = 10
		if p, ok := f.Data[i].(string); ok {
			finding.Published = p
		} else {
			warnings = append(warnings, fmt.Errorf("couldn't coerce data[%d] %v into a string", i, f.Data[i]))
		}

		r.findings = append(r.findings, finding)
	}

	return r.findings, warnings, nil
}

func (r *Report) FindingByPartial(partial string) (*Finding, error) {
	findings, warnings, err := r.Findings()

	var match *Finding

	if err != nil {
		return match, err
	}

	_ = warnings // TODO hide warnings, i guess?
	matches := 0

	for _, f := range findings {
		if strings.Contains(strings.ToLower(f.Name), strings.ToLower(partial)) {
			match = f
			matches++
		}
	}

	if matches == 0 {
		return match, errors.New("finding not found")
	}

	if matches > 1 {
		return match, errors.New("multiple findings match")
	}

	return match, nil
}

func findingAssets(m map[string]interface{}) ([]Asset, []error, error) {
	var assets []Asset

	var warnings []error

	affected_assets, ok := m["affected_assets"].(map[string]interface{})
	if !ok {
		return assets, warnings, errors.New("unable to coerce affected_assets into map[string]interface{}")
	}

	for k, asset := range affected_assets {
		asset_map, ok := asset.(map[string]interface{})
		if !ok {
			return assets, warnings, errors.New("unable to coerce asset into map[string]interface{}")
		}

		asset := Asset{
			ID: k,
		}

		if v, ok := asset_map["asset"].(string); ok {
			asset.Value = v
		} else {
			warnings = append(warnings, fmt.Errorf("unable to coerce asset into string: %#v", asset_map["asset"]))
		}
		// fmt.Printf("Asset %s: %#v\n", k, asset_map["asset"])
		assets = append(assets, asset)
	}

	return assets, warnings, nil
}

func findingEvidence(m map[string]interface{}) (string, []error, error) {
	fields, ok := m["fields"].(map[string]interface{})
	if !ok {
		return "", nil, errors.New("unable to coerce fields into map[string]interface{}")
	}

	_, ok = fields["evidence"]
	if !ok {
		return "", []error{errors.New("evidence is missing")}, nil
	}

	evidence, ok := fields["evidence"].(map[string]interface{})
	if !ok {
		return "", nil, errors.New("unable to coerce evidence into map[string]interface{}")
	}

	value, ok := evidence["value"].(string)
	if !ok {
		return "", nil, errors.New("unable to coerce evidence value into string")
	}

	// fmt.Printf("Evidence: %#v\n", evidence)
	return value, nil, nil
}

func (f *Finding) EnsureFull() ([]error, error) {
	var warnings []error

	var warningsParsed []error

	var err error

	if f.full {
		return nil, nil
	}

	path := fmt.Sprintf("v1/client/%d/report/%d/flaw/%d", f.r.c.ID, f.r.ID, f.ID)

	_, err = f.r.ua.apiGet(path, &f.raw)
	if err != nil {
		return nil, err
	}

	f.full = true

	// Parse Affected Assets
	f.assets, warningsParsed, err = findingAssets(f.raw)
	if err != nil {
		return nil, err
	}

	warnings = append(warnings, warningsParsed...)

	// Parse Evidence
	f.Evidence, warningsParsed, err = findingEvidence(f.raw)
	if err != nil {
		return nil, err
	}

	warnings = append(warnings, warningsParsed...)

	// Parse Tags
	if tags, ok := f.raw["tags"].([]interface{}); ok {
		for _, t := range tags {
			if tag, ok := t.(string); ok {
				f.tags = append(f.tags, tag)
			} else {
				warnings = append(warnings, fmt.Errorf("unable to coerce %#v into a string", tag))
			}
		}
	} else {
		warnings = append(warnings, fmt.Errorf("unable to coerce %#v into a []string", f.raw["tags"]))
	}

	return warnings, nil
}

func (f *Finding) update() ([]error, error) {
	path := fmt.Sprintf("v1/client/%d/report/%d/flaw/%d", f.r.c.ID, f.r.ID, f.ID)

	body, err := f.r.ua.apiCall(http.MethodPut, path, f.raw, nil)
	if err != nil {
		fmt.Printf("body: %s\n", body)

		return nil, fmt.Errorf("error updating finding: %w", err)
	}

	return nil, nil
}

func (f *Finding) Tags() []string {
	_, _ = f.EnsureFull()

	return f.tags
}

func (f *Finding) AddTags(tags []string) ([]error, error) {
	warnings, err := f.EnsureFull()
	if err != nil {
		return warnings, err
	}

	f.tags = append(f.tags, tags...)
	f.raw["tags"] = f.tags
	warnings2, err := f.update()
	warnings = append(warnings, warnings2...)

	return warnings, err
}

func (f *Finding) RemoveTags(tags []string) ([]error, error) {
	warnings, err := f.EnsureFull()
	if err != nil {
		return warnings, err
	}

	f.tags = slices.DeleteFunc(f.tags, func(t string) bool {
		return slices.Contains(tags, t)
	})
	fmt.Printf("tags: %#v\n", f.tags)
	f.raw["tags"] = f.tags
	warnings2, err := f.update()
	warnings = append(warnings, warnings2...)

	return warnings, err
}

func (f *Finding) SetTags(tags []string) ([]error, error) {
	warnings, err := f.EnsureFull()
	if err != nil {
		return warnings, err
	}

	f.tags = tags
	f.raw["tags"] = f.tags
	warnings2, err := f.update()
	warnings = append(warnings, warnings2...)

	return warnings, err
}
