package main

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Fixed UUIDs with every output format spelled out, shared by the regression
// and conversion tests.
type fixture struct {
	version int
	formats map[string]string // format -> lowercase value
}

var (
	fixtureV4 = fixture{4, map[string]string{
		"human":     "3f2a8b6c-1d4e-4f5a-9b7c-0d1e2f3a4b5c",
		"hex":       "3f2a8b6c1d4e4f5a9b7c0d1e2f3a4b5c",
		"base64":    "PyqLbB1OT1qbfA0eLzpLXA==",
		"base64url": "PyqLbB1OT1qbfA0eLzpLXA",
		"int":       "83962268023154358113187420214925085532",
	}}
	fixtureV7 = fixture{7, map[string]string{
		"human":     "0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0b1",
		"hex":       "0192a0b1c2d37e4f8a5b6c7d8e9fa0b1",
		"base64":    "AZKgscLTfk+KW2x9jp+gsQ==",
		"base64url": "AZKgscLTfk-KW2x9jp-gsQ",
		"int":       "2090562606348121737443780115287744689",
	}}
	fixtures   = []fixture{fixtureV4, fixtureV7}
	allFormats = []string{"human", "hex", "base64", "base64url", "int"}
)

func (f fixture) validMessage() string {
	if f.version == 7 {
		return "valid uuid version 7 (timestamp 2024-10-18T17:34:17.299Z)\n"
	}
	return "valid uuid version 4\n"
}

// TestRegressionGolden pins the exact output of behaviour that existed before
// format conversion was added.
func TestRegressionGolden(t *testing.T) {
	v4, v7 := fixtureV4.formats["human"], fixtureV7.formats["human"]
	tests := []struct {
		name           string
		args           []string
		stdin          string
		stdout, stderr string
		code           int
	}{
		{"version mismatch", []string{"--validate=" + v7, "-4"}, "",
			"", "error: validate uuid: UUID version 7 does not match constraint 4\n", 1},
		{"multiple invalid values", []string{"--validate=a", "b"}, "",
			"a: invalid: invalid UUID: unsupported format or parse failure\n" +
				"b: invalid: invalid UUID: unsupported format or parse failure\n", "", 1},
		{"stdin report", []string{"--validate=-", "-7"}, v7 + "\n\nbad\n" + v4 + "\n",
			v7 + ": valid uuid version 7 (timestamp 2024-10-18T17:34:17.299Z)\n" +
				"bad: invalid: invalid UUID: unsupported format or parse failure\n" +
				v4 + ": invalid: UUID version 4 does not match constraint 7\n", "", 1},
		{"unsupported version", []string{"--uuid=5"}, "",
			"", "error: generate uuid: unsupported UUID version: use 4 or 7\n", 1},
		{"unsupported format", []string{"--format=octal"}, "",
			"", "error: format uuid: unsupported format: use human, hex, base64, base64url, or int\n", 1},
		{"conflicting shortcuts", []string{"-4", "-7"}, "",
			"", "error: conflicting UUID shortcuts: choose only one of -4 or -7\n", 1},
		{"zero count", []string{"-n", "0"}, "",
			"", "error: -n must be at least 1\n", 1},
		{"validate with format", []string{"--validate=x", "--format=int"}, "",
			"", "error: -format cannot be used with --validate\n", 1},
		{"32-digit integer is not hex", []string{"--validate=11111111111141111111111111111111"}, "",
			"", "error: validate uuid: invalid UUID: unsupported variant Reserved\n", 1},
		{"non-RFC variant", []string{"--validate=00000000-0000-4000-0000-000000000000"}, "",
			"", "error: validate uuid: invalid UUID: unsupported variant Reserved\n", 1},
		{"version 1 constraint", []string{"--validate=6ba7b810-9dad-11d1-80b4-00c04fd430c8", "--uuid=1"}, "",
			"", "error: validate uuid: invalid version constraint: use 4 or 7\n", 1},
		{"negative integer", []string{"--validate=-" + fixtureV7.formats["int"]}, "",
			"", "error: validate uuid: invalid UUID: unsupported format or parse failure\n", 1},
		{"integer above 128 bits", []string{"--validate=340282366920938463463374607431768211456"}, "",
			"", "error: validate uuid: invalid UUID: integer out of 128-bit range\n", 1},
	}
	for _, f := range fixtures {
		for _, format := range allFormats {
			tests = append(tests, struct {
				name           string
				args           []string
				stdin          string
				stdout, stderr string
				code           int
			}{"validate v" + string(rune('0'+f.version)) + " " + format,
				[]string{"--validate=" + f.formats[format]}, "", f.validMessage(), "", 0})
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runWithStdin(t, tt.stdin, tt.args...)
			if stdout != tt.stdout || stderr != tt.stderr || code != tt.code {
				t.Fatalf("run(%q)\n got: stdout=%q stderr=%q code=%d\nwant: stdout=%q stderr=%q code=%d",
					tt.args, stdout, stderr, code, tt.stdout, tt.stderr, tt.code)
			}
		})
	}
}

func TestRegressionHelpLayout(t *testing.T) {
	stdout, _, _ := runCommand(t, "--help")
	last := -1
	for _, section := range []string{"UUID generator\n", "Usage: ", "Options:\n", "Examples:\n", "version: ", "build date: ", "build number: "} {
		idx := strings.Index(stdout, section)
		if idx <= last {
			t.Fatalf("section %q missing or out of order in help output:\n%s", section, stdout)
		}
		last = idx
	}
}

// TestRegressionGeneratedValuesValidate checks every generated format and case
// is accepted back by --validate with the right version.
func TestRegressionGeneratedValuesValidate(t *testing.T) {
	for _, version := range []string{"-4", "-7"} {
		for _, format := range allFormats {
			for _, letterCase := range []string{"lower", "upper"} {
				args := []string{version, "-n", "3", "--format=" + format, "--case=" + letterCase}
				stdout, stderr, code := runCommand(t, args...)
				if code != 0 {
					t.Fatalf("%v: code %d, stderr %q", args, code, stderr)
				}
				for _, value := range strings.Fields(stdout) {
					out, stderr, code := runCommand(t, "--validate="+value, version)
					if code != 0 || !strings.HasPrefix(out, "valid uuid version "+version[1:]) {
						t.Fatalf("%v: generated %q not accepted: %q %q", args, value, out, stderr)
					}
					if format == "human" {
						if _, err := uuid.Parse(value); err != nil {
							t.Fatalf("%v: %q is not a canonical UUID", args, value)
						}
					}
				}
			}
		}
	}
}
