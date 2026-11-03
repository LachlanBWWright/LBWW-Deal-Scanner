package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"dealscanner/internal/valuation"
)

const maxInputBytes = 16 * 1024 * 1024

func run(input io.Reader, output io.Writer) error {
	data, err := io.ReadAll(io.LimitReader(input, maxInputBytes+1))
	if err != nil {
		return fmt.Errorf("read trade-up input: %w", err)
	}
	if len(data) > maxInputBytes {
		return fmt.Errorf("trade-up input exceeds %d bytes", maxInputBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var study valuation.TradeStudy
	if err := decoder.Decode(&study); err != nil {
		return fmt.Errorf("decode trade-up input: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("input must contain exactly one JSON document")
	}
	report, err := valuation.EvaluateTradeStudy(study)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write trade-up report: %w", err)
	}
	return nil
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
