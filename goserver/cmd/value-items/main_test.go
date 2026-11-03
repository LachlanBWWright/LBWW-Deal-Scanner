package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"dealscanner/internal/valuation"
)

func TestExampleRuns(t *testing.T) {
	data, err := os.ReadFile("../../examples/valuation/cs2-floats.synthetic.json")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(bytes.NewReader(data), &output); err != nil {
		t.Fatal(err)
	}
	var reports []valuation.Report
	if err := json.Unmarshal(output.Bytes(), &reports); err != nil {
		t.Fatal(err)
	}
	if len(reports) != 4 || reports[0].Estimate == nil || reports[0].Estimate.GrossDiscountMinor != 450 ||
		reports[1].Estimate == nil || reports[1].Estimate.FloatPremiumMinor != 0 ||
		reports[2].State != "insufficient_evidence" || reports[3].Estimate == nil || reports[3].Estimate.Basis != valuation.Ask {
		t.Fatalf("unexpected demo result: %s", output.String())
	}
}

func TestRejectsMalformedUnknownAndTrailingJSON(t *testing.T) {
	for _, input := range []string{"{", `{"unexpected":true}`, `{} {}`, "{} rubbish", strings.Repeat(" ", maxInputBytes+1)} {
		var output bytes.Buffer
		if err := run(strings.NewReader(input), &output); err == nil {
			t.Fatal("accepted invalid input")
		}
		if output.Len() != 0 {
			t.Fatal("wrote partial report for invalid input")
		}
	}
}
