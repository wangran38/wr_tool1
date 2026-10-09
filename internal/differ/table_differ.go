package differ

import (
	"strings"

	"wr_tool/internal/cleaner"
	"wr_tool/internal/parser"
)

type CellChange struct {
	Col  int    `json:"col"`
	Kind Kind   `json:"kind"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

type RowChange struct {
	Kind  Kind         `json:"kind"`
	Old   []string     `json:"old"`
	New   []string     `json:"new"`
	Cells []CellChange `json:"cells"`
}

type TableDiff struct {
	Index    int         `json:"index"`
	Kind     Kind        `json:"kind"`
	RowCount int         `json:"rowCount"`
	Rows     []RowChange `json:"rows"`
	Stats    Stats       `json:"stats"`
}

// DiffTables pairs tables by position, aligns their rows, then compares cells.
func DiffTables(oldTables, newTables []parser.Table, o cleaner.Options) []TableDiff {
	count := len(oldTables)
	if len(newTables) > count {
		count = len(newTables)
	}

	diffs := make([]TableDiff, 0, count)
	for i := 0; i < count; i++ {
		if i >= len(oldTables) {
			diffs = append(diffs, wholeTable(i, Insert, newTables[i]))
			continue
		}
		if i >= len(newTables) {
			diffs = append(diffs, wholeTable(i, Delete, oldTables[i]))
			continue
		}
		diffs = append(diffs, diffOneTable(i, oldTables[i], newTables[i], o))
	}
	return diffs
}

func wholeTable(index int, kind Kind, t parser.Table) TableDiff {
	var rows []RowChange
	for _, r := range t.Rows {
		if kind == Insert {
			rows = append(rows, RowChange{Kind: Insert, New: r})
		} else {
			rows = append(rows, RowChange{Kind: Delete, Old: r})
		}
	}
	return TableDiff{Index: index, Kind: kind, RowCount: len(t.Rows), Rows: rows}
}

func diffOneTable(index int, oldT, newT parser.Table, o cleaner.Options) TableDiff {
	oldUnits := rowUnits(oldT.Rows, o)
	newUnits := rowUnits(newT.Rows, o)
	result := Diff(oldUnits, newUnits)

	d := TableDiff{Index: index, Kind: Equal, RowCount: len(newT.Rows)}

	var rows []RowChange
	for _, p := range result.Pairs {
		switch p.Kind {
		case Replace:
			rows = append(rows, RowChange{
				Kind:  Replace,
				Old:   splitCells(p.Old),
				New:   splitCells(p.New),
				Cells: diffCells(splitCells(p.Old), splitCells(p.New), o),
			})
		case Insert:
			rows = append(rows, RowChange{Kind: Insert, New: splitCells(p.New)})
		case Delete:
			rows = append(rows, RowChange{Kind: Delete, Old: splitCells(p.Old)})
		}
	}

	if len(rows) == 0 {
		d.Kind = Equal
	} else {
		d.Kind = Replace
	}
	d.Rows = rows
	d.Stats = result.Stats
	return d
}

// rowCells joins cells so the row survives the cleaner's whitespace rules and can be
// split back apart for per-cell reporting.
const cellSeparator = "\x1f"

func rowUnits(rows [][]string, o cleaner.Options) []cleaner.Unit {
	units := make([]cleaner.Unit, 0, len(rows))
	for _, r := range rows {
		raw := strings.Join(r, cellSeparator)
		keyCells := make([]string, len(r))
		for i, c := range r {
			keyCells[i] = cleaner.Normalize(c, o)
		}
		key := strings.Join(keyCells, cellSeparator)
		if key == "" {
			continue
		}
		units = append(units, cleaner.Unit{Raw: raw, Key: key})
	}
	return units
}

func splitCells(raw string) []string {
	return strings.Split(raw, cellSeparator)
}

func diffCells(oldRow, newRow []string, o cleaner.Options) []CellChange {
	cols := len(oldRow)
	if len(newRow) > cols {
		cols = len(newRow)
	}

	cells := make([]CellChange, 0, cols)
	for i := 0; i < cols; i++ {
		var oldCell, newCell string
		if i < len(oldRow) {
			oldCell = oldRow[i]
		}
		if i < len(newRow) {
			newCell = newRow[i]
		}
		kind := Equal
		switch {
		case oldCell == "" && newCell != "":
			kind = Insert
		case newCell == "" && oldCell != "":
			kind = Delete
		case cleaner.Normalize(oldCell, o) != cleaner.Normalize(newCell, o):
			kind = Replace
		}
		if kind == Equal {
			continue
		}
		cells = append(cells, CellChange{Col: i, Kind: kind, Old: oldCell, New: newCell})
	}
	return cells
}
