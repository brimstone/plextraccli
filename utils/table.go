// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var baseStyle = lipgloss.NewStyle()

type TableColumn = table.Column

type tableModel struct {
	table table.Model
	rows  []table.Row
	cols  []table.Column
}

func NewTable() tableModel {
	var t tableModel

	t.table = table.New()

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = lipgloss.NewStyle()

	t.table.SetStyles(s)

	return t
}

func (m *tableModel) SetColumns(cols []table.Column) {
	m.cols = cols
	for i, j := range m.cols {
		m.cols[i].Width = len(j.Title)
	}

	m.table.SetColumns(cols)
}

func (m *tableModel) AddRow(r []string) {
	m.rows = append(m.rows, r)
	for i, j := range r {
		// Auto size the Name column if it's needed
		if len(j) > m.cols[i].Width {
			m.cols[i].Width = len(j)
		}
	}

	m.table.SetRows(m.rows)
	m.table.SetHeight(len(m.rows) + 2)
}

func (m *tableModel) Render() string {
	// fmt.Printf("%#v\n", m.rows)
	return baseStyle.Render(m.table.View())
}

func LowerCaseHeaders(h []string) []string {
	var x []string
	for _, s := range h {
		x = append(x, strings.ToLower(strings.ReplaceAll(s, " ", "")))
	}

	return x
}

func transposeMatrix(matrix [][]string) [][]string {
	// stolen from https://codesignal.com/learn/courses/multidimensional-arrays-and-their-traversal-in-go/lessons/transposing-matrices-in-go-a-hands-on-tutorial
	rows := len(matrix)
	cols := len(matrix[0])

	result := make([][]string, cols)
	for i := range result {
		result[i] = make([]string, rows)
	}

	for i := range rows {
		for j := range cols {
			result[j][i] = matrix[i][j]
		}
	}

	return result
}

func ShowTable(headers []string, rows [][]string, showCols []string) {
	transposeCols := transposeMatrix(rows)

	var toShowMatrix [][]string

	var toShowHeaders []TableColumn

	// Determine which columns need to be shown
	for _, c := range showCols {
		foundColumn := false

		for i, h := range headers {
			h2 := strings.ToLower(strings.ReplaceAll(h, " ", ""))
			// This column needs to be shown
			if h2 == c {
				foundColumn = true

				toShowMatrix = append(toShowMatrix, transposeCols[i])
				toShowHeaders = append(toShowHeaders, TableColumn{Title: h})
			}
		}

		if !foundColumn {
			fmt.Printf("unsupported column: %s\n", c)

			return
		}
	}

	showRows := transposeMatrix(toShowMatrix)

	if isatty.IsTerminal(os.Stdout.Fd()) {
		t := NewTable()
		t.SetColumns(toShowHeaders)

		for _, r := range showRows {
			t.AddRow(r)
		}

		fmt.Println(t.Render())
	} else {
		for _, r := range showRows {
			fmt.Println(strings.Join(r, " "))
		}
	}
}
