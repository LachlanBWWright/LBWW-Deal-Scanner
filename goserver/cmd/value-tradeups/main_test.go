package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"dealscanner/internal/valuation"
)

func TestTradeExampleRuns(t *testing.T) {
	data, err := os.ReadFile("../../examples/valuation/cs2-tradeup.synthetic.json")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(bytes.NewReader(data), &output); err != nil {
		t.Fatal(err)
	}
	var report valuation.TradeReport
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Candidate.ExpectedProfitMinor != 1050 || report.MarginalExpectedProfitMinor != 800 || report.Candidate.LossProbability != .5 {
		t.Fatalf("unexpected result: %s", output.String())
	}
}

func TestRejectsInvalidTradeJSON(t *testing.T) {
	for _, input := range []string{"{", `{"unexpected":true}`, `{} {}`, "{} rubbish", strings.Repeat(" ", maxInputBytes+1)} {
		var output bytes.Buffer
		if err := run(strings.NewReader(input), &output); err == nil {
			t.Fatal("accepted invalid input")
		}
		if output.Len() != 0 {
			t.Fatal("wrote partial output")
		}
	}
}
