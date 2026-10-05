package uuidgen

import (
	"strings"
	"testing"
)

var allFormats = []string{"human", "hex", "base64", "base64url", "int"}

// convertFixtures holds UUIDs with every format spelled out (lowercase).
var convertFixtures = []struct {
	version int
	formats map[string]string
}{
	{4, map[string]string{
		"human":     "3f2a8b6c-1d4e-4f5a-9b7c-0d1e2f3a4b5c",
		"hex":       "3f2a8b6c1d4e4f5a9b7c0d1e2f3a4b5c",
		"base64":    "PyqLbB1OT1qbfA0eLzpLXA==",
		"base64url": "PyqLbB1OT1qbfA0eLzpLXA",
		"int":       "83962268023154358113187420214925085532",
	}},
	{7, map[string]string{
		"human":     "0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0b1",
		"hex":       "0192a0b1c2d37e4f8a5b6c7d8e9fa0b1",
		"base64":    "AZKgscLTfk+KW2x9jp+gsQ==",
		"base64url": "AZKgscLTfk-KW2x9jp-gsQ",
		"int":       "2090562606348121737443780115287744689",
	}},
}

// inputVariants returns every accepted spelling of the fixture, keyed by a description.
func inputVariants(formats map[string]string) map[string]string {
	inputs := map[string]string{
		"human upper":  strings.ToUpper(formats["human"]),
		"hex upper":    strings.ToUpper(formats["hex"]),
		"human braces": "{" + formats["human"] + "}",
		"human urn":    "urn:uuid:" + formats["human"],
		"int zero-pad": "000" + formats["int"],
	}
	for _, format := range allFormats {
		inputs[format] = formats[format]
	}
	return inputs
}

func expectedOutput(formats map[string]string, format, letterCase string) string {
	if letterCase == "upper" && (format == "human" || format == "hex") {
		return strings.ToUpper(formats[format])
	}
	return formats[format]
}

func TestConvertAllCombinations(t *testing.T) {
	for _, f := range convertFixtures {
		for inputName, input := range inputVariants(f.formats) {
			for _, format := range allFormats {
				for _, letterCase := range []string{"lower", "upper"} {
					for _, constraint := range []int{0, f.version} {
						got, err := Convert(input, constraint, format, letterCase)
						want := expectedOutput(f.formats, format, letterCase)
						if err != nil || got != want {
							t.Errorf("v%d %s -> %s/%s (constraint %d): got %q, %v; want %q",
								f.version, inputName, format, letterCase, constraint, got, err, want)
						}
					}
				}
			}
		}
	}
}

func TestConvertRejects(t *testing.T) {
	v7 := convertFixtures[1].formats
	tests := []struct {
		name, input, format, letterCase string
		constraint                      int
		wantErr                         string
	}{
		{"unrecognized", "not-a-uuid", "hex", "lower", 0, "unsupported format or parse failure"},
		{"wrong variant", "00000000-0000-4000-0000-000000000000", "hex", "lower", 0, "unsupported variant"},
		{"version 1", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", "hex", "lower", 0, "unsupported UUID version 1"},
		{"version mismatch", v7["human"], "hex", "lower", 4, "does not match constraint"},
		{"bad constraint", v7["human"], "hex", "lower", 1, "invalid version constraint"},
		{"32-digit integer", "11111111111141111111111111111111", "human", "lower", 0, "unsupported variant"},
		{"negative integer", "-" + v7["int"], "human", "lower", 0, "unsupported format"},
		{"integer above 128 bits", "340282366920938463463374607431768211456", "human", "lower", 0, "128-bit range"},
		{"unknown target format", v7["human"], "octal", "lower", 0, "unsupported format"},
		{"unknown case", v7["human"], "human", "mixed", 0, "unsupported case"},
		// Steps run in order: an invalid input is reported before an invalid target format.
		{"invalid input and target", "not-a-uuid", "octal", "lower", 0, "unsupported format or parse failure"},
		{"wrong version and target", v7["human"], "octal", "lower", 4, "does not match constraint"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.input, tt.constraint, tt.format, tt.letterCase)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) || got != "" {
				t.Fatalf("got %q, %v; want error containing %q", got, err, tt.wantErr)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	for _, f := range convertFixtures {
		for _, format := range allFormats {
			if got, err := Detect(f.formats[format]); err != nil || got != format {
				t.Errorf("Detect(%q) = %q, %v; want %q", f.formats[format], got, err, format)
			}
		}
		for _, name := range []string{"human upper", "human braces", "human urn"} {
			input := inputVariants(f.formats)[name]
			if got, err := Detect(input); err != nil || got != "human" {
				t.Errorf("Detect(%q) = %q, %v; want human", input, got, err)
			}
		}
	}
	tests := map[string]string{
		"11111111111141111111111111111111":              "int", // digits only: integer, not hex
		"-KGyw9TlT2CKe4ydDh8qOw":                        "base64url",
		"URN:UUID:3f2a8b6c-1d4e-4f5a-9b7c-0d1e2f3a4b5c": "human",
	}
	for input, want := range tests {
		if got, err := Detect(input); err != nil || got != want {
			t.Errorf("Detect(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{
		"", " 3f2a8b6c-1d4e-4f5a-9b7c-0d1e2f3a4b5c", "3f2a8b6c-1d4e-4f5a-9b7c-0d1e2f3a4b5", "{3f2a8b6c-1d4e-4f5a-9b7c-0d1e2f3a4b5c",
		"3f2a8b6c1d4e4f5a9b7c0d1e2f3a4b5g", "-123", "AZKgscLTfk+KW2x9jp+gsQ", "AZKgscLTfk-KW2x9jp-gsQ==", "AZKgscLTfk+KW2x9jp+gs==",
	} {
		if got, err := Detect(input); err == nil {
			t.Errorf("Detect(%q) = %q, want error", input, got)
		}
	}
}
