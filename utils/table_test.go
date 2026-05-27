// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package utils_test

import (
	"testing"

	"github.com/brimstone/plextraccli/utils"
)

func TestLowerCaseHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "uppercase spaces",
			input:    []string{"Start Date", "Is Active"},
			expected: []string{"startdate", "isactive"},
		},
		{
			name:     "already lowercase",
			input:    []string{"name", "status"},
			expected: []string{"name", "status"},
		},
		{
			name:     "mixed case no spaces",
			input:    []string{"StatusCode", "UserName"},
			expected: []string{"statuscode", "username"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "single item",
			input:    []string{"Name"},
			expected: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := utils.LowerCaseHeaders(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("LowerCaseHeaders() length = %d, want %d", len(got), len(tt.expected))

				return
			}

			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("LowerCaseHeaders()[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestTransposeMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    [][]string
		expected [][]string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "2x2",
			input:    [][]string{{"a", "b"}, {"c", "d"}},
			expected: [][]string{{"a", "c"}, {"b", "d"}},
		},
		{
			name:     "single row",
			input:    [][]string{{"a", "b", "c"}},
			expected: [][]string{{"a"}, {"b"}, {"c"}},
		},
		{
			name:     "single column",
			input:    [][]string{{"a"}, {"b"}},
			expected: [][]string{{"a", "b"}},
		},
		{
			name:     "3x2",
			input:    [][]string{{"a", "b"}, {"c", "d"}, {"e", "f"}},
			expected: [][]string{{"a", "c", "e"}, {"b", "d", "f"}},
		},
		{
			name:     "single element",
			input:    [][]string{{"x"}},
			expected: [][]string{{"x"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := utils.TransposeMatrix(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("transposeMatrix() length = %d, want %d", len(got), len(tt.expected))

				return
			}

			for i, row := range got {
				if len(row) != len(tt.expected[i]) {
					t.Errorf("transposeMatrix()[%d] length = %d, want %d", i, len(row), len(tt.expected[i]))

					continue
				}

				for j, v := range row {
					if v != tt.expected[i][j] {
						t.Errorf("transposeMatrix()[%d][%d] = %q, want %q", i, j, v, tt.expected[i][j])
					}
				}
			}
		})
	}
}
