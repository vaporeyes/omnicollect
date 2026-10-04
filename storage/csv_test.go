// ABOUTME: CSV exports neutralize spreadsheet formulas in headers and all text fields.
// ABOUTME: Tests parse actual CSV to distinguish quoting from formula neutralization.
package storage

import (
	"encoding/csv"
	"strings"
	"testing"
)

func TestCSVFormulaCells(t *testing.T) {
	for _, value := range []string{"=1+1", "+cmd", "-1+2", "@SUM(A1)", " \t=1", "\ufeff=1", "\ttext", "\rtext", "\ntext", "＝1", "＝\"a,b\""} {
		rows, err := csv.NewReader(strings.NewReader(csvRow([]string{value}))).ReadAll()
		if err != nil || len(rows) != 1 || rows[0][0] != "'"+value {
			t.Fatalf("%q: %#v %v", value, rows, err)
		}
	}
	for _, value := range []string{"plain", "already 'quoted'", "a,b", "quoted \"text\""} {
		row, e := csv.NewReader(strings.NewReader(csvRow([]string{value}))).Read()
		if e != nil || row[0] != value {
			t.Fatalf("plain value changed: %q %v", row, e)
		}
	}
	output, err := buildCSV([]Item{{ID: "=id", Title: "+title", ModuleID: "m", Tags: []string{"@tag"}, Attributes: map[string]any{"=header": "-value"}}}, []ModuleSchema{{ID: "m", DisplayName: "=module"}})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0][7] != "'=header" || rows[1][0] != "'=id" || rows[1][1] != "'+title" || rows[1][2] != "'=module" || rows[1][4] != "'@tag" || rows[1][7] != "'-value" {
		t.Fatalf("unprotected cells: %#v", rows)
	}
}
