package discord

import (
	"strings"
	"testing"

	"dealscanner/internal/db/query"
)

func TestParseAgentChanges(t *testing.T) {
	good := `{"format":"dealscanner.query-changes","version":1,"changes":[{"operation":"create","query":{"type":"ebay","url":"https://www.ebay.com.au/sch/i.html?_nkw=switch","maxPrice":25,"dmOnly":false}}]}`
	tests := []struct {
		name, input string
		valid       bool
	}{
		{"valid", good, true},
		{"code fence", "```json\n" + good + "\n```", true},
		{"wrong version", strings.Replace(good, `"version":1`, `"version":2`, 1), false},
		{"unknown field", strings.Replace(good, `"maxPrice":25`, `"surprise":true,"maxPrice":25`, 1), false},
		{"missing dmOnly", strings.Replace(good, `,"dmOnly":false`, "", 1), false},
		{"external host", strings.Replace(good, "www.ebay.com.au", "localhost", 1), false},
		{"missing update revision", strings.Replace(good, `"create"`, `"update"`, 1), false},
		{"negative price", strings.Replace(good, `25`, `-1`, 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAgentChanges(tt.input)
			if (err == nil) != tt.valid {
				t.Fatalf("error = %v, valid = %v", err, tt.valid)
			}
		})
	}
}
func TestQueryRevisionChangesWithQuery(t *testing.T) {
	a := query.QueryItem{QueryId: "id", Type: "ebay", DmOnly: false}
	b := a
	b.DmOnly = true
	first, err := queryRevision(a)
	if err != nil {
		t.Fatal(err)
	}
	second, err := queryRevision(b)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("revision did not change")
	}
}
