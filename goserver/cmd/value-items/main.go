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
		return fmt.Errorf("read valuation input: %w", err)
	}
	if len(data) > maxInputBytes {
		return fmt.Errorf("valuation input exceeds %d bytes", maxInputBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var dataset valuation.Dataset
	if err := decoder.Decode(&dataset); err != nil {
		return fmt.Errorf("decode valuation input: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("input must contain exactly one JSON document")
	}
	reports, err := valuation.Evaluate(dataset)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(reports); err != nil {
		return fmt.Errorf("write valuation reports: %w", err)
	}
	return nil
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
