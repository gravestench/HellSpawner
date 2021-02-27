package stringtablewidget

import (
	"github.com/ianling/giu"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2tbl"
)

type widget struct {
	id   string
	dict *d2tbl.TextDictionary
}

func Create(id string, dict *d2tbl.TextDictionary) giu.Widget {
	result := &widget{
		id:   id,
		dict: dict,
	}

	return result
}

func (p *widget) Build() {
	numEntries := len(p.dict.Entries)

	// wprobably will remove
	if !(numEntries > 0) {
		giu.Layout{}.Build()
	}

	rows := make([]*giu.RowWidget, numEntries+1)

	columns := []string{"key", "value"}
	columnWidgets := make([]giu.Widget, len(columns))

	for idx := range columns {
		columnWidgets[idx] = giu.Label(columns[idx])
	}

	rows[0] = giu.Row(columnWidgets...)

	keyIdx := 0

	for key := range p.dict.Entries {
		rows[keyIdx+1] = giu.Row(
			giu.Label(key),
			giu.Label(p.dict.Entries[key]),
		)

		keyIdx++
	}

	giu.Layout{
		giu.Child("").Border(false).Layout(giu.Layout{
			giu.FastTable("").Border(true).Rows(rows),
		}),
	}.Build()
}
