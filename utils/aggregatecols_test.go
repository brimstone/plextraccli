// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package utils_test

import (
	"testing"

	"github.com/brimstone/plextraccli/utils"
)

func TestAggregateCols_add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		def      []string
		modify   string
		expected []string
	}{
		{
			name:     "add new column",
			def:      []string{"name", "status"},
			modify:   "+tags",
			expected: []string{"name", "status", "tags"},
		},
		{
			name:     "add column already present",
			def:      []string{"name", "tags"},
			modify:   "+tags",
			expected: []string{"name", "tags"},
		},
		{
			name:     "add multiple columns",
			def:      []string{"name"},
			modify:   "+tags,status",
			expected: []string{"name", "tags", "status"},
		},
		{
			name:     "add multiple with some already present",
			def:      []string{"name", "tags"},
			modify:   "+tags,severity",
			expected: []string{"name", "tags", "severity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := utils.AggregateCols(tt.def, tt.modify)
			if len(got) != len(tt.expected) {
				t.Errorf("AggregateCols() length = %d, want %d", len(got), len(tt.expected))

				return
			}

			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("AggregateCols()[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestAggregateCols_override(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		def      []string
		modify   string
		expected []string
	}{
		{
			name:     "override single column",
			def:      []string{"name", "status"},
			modify:   "tags",
			expected: []string{"tags"},
		},
		{
			name:     "override with multiple columns",
			def:      []string{"name"},
			modify:   "name,status",
			expected: []string{"name", "status"},
		},
		{
			name:     "dash prefix is a plain override",
			def:      []string{"name", "startdate"},
			modify:   "-startdate",
			expected: []string{"-startdate"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := utils.AggregateCols(tt.def, tt.modify)
			if len(got) != len(tt.expected) {
				t.Errorf("AggregateCols() length = %d, want %d", len(got), len(tt.expected))

				return
			}

			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("AggregateCols()[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}
