// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

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

func ShowTable(headers []string, rows [][]string, showCols []string) {
	var cols []TableColumn

	var keepCol []bool

	for _, h := range headers {
		h2 := strings.ToLower(strings.ReplaceAll(h, " ", ""))

		var i int

		// remove the value from showCols
		var c string
		for i, c = range showCols {
			if c == h2 {
				showCols = append(showCols[:i], showCols[i+1:]...)

				break
			}
		}

		if c == h2 { // still
			cols = append(cols, TableColumn{Title: h})
			keepCol = append(keepCol, true)
		} else {
			keepCol = append(keepCol, false)
		}
	}
	// check to see if showCols still has a value, if it does, error with unsupported header
	if len(showCols) > 0 {
		fmt.Printf("unsupported header: %s\n", showCols[0])
	}

	var showRows [][]string

	for _, r := range rows {
		var r3 []string

		for c, r2 := range r {
			if keepCol[c] {
				r3 = append(r3, r2)
			}
		}

		showRows = append(showRows, r3)
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		t := NewTable()
		t.SetColumns(cols)

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
