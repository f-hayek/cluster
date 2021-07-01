package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

type Table struct {
	tview.Table
}

func NewTable() *Table {
	tt := tview.NewTable()
	return &Table{*tt}
}

func (t *Table) AddColumnHeader(name string, align int) {
	lines := strings.Split(name, "\n")
	col := t.GetColumnCount()
	for i, v := range lines {
		t.SetCell(i, col,
			tview.NewTableCell(v).SetTextColor(tcell.ColorWhite).SetAlign(align))
	}
}

func (t *Table) AddHeaderSeparator() {
	cols := t.GetColumnCount()
	rowOffset := t.GetRowCount()
	for i := 0; i < cols; i++ {
		t.SetCell(rowOffset, i,
			tview.NewTableCell("────────────"))
	}
}
