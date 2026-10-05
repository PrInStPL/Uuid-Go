package main

import (
	"strings"
	"testing"
)

func cliInputs(f fixture) map[string]string {
	inputs := map[string]string{
		"human upper": strings.ToUpper(f.formats["human"]),
		"hex upper":   strings.ToUpper(f.formats["hex"]),
	}
	for _, format := range allFormats {
		inputs[format] = f.formats[format]
	}
	return inputs
}

// TestConvertCLI runs every input format into every output format and case,
// with and without --pure, and checks the exact bytes written.
func TestConvertCLI(t *testing.T) {
	for _, f := range fixtures {
		version := "-" + string(rune('0'+f.version))
		for inputName, input := range cliInputs(f) {
			for _, format := range allFormats {
				for _, letterCase := range []string{"lower", "upper"} {
					want := f.formats[format]
					if letterCase == "upper" && (format == "human" || format == "hex") {
						want = strings.ToUpper(want)
					}
					for _, extra := range [][]string{nil, {version}} {
						base := append([]string{"--format=" + format, "--case=" + letterCase}, extra...)

						args := append(append([]string{}, base...), input)
						stdout, stderr, code := runCommand(t, args...)
						if code != 0 || stderr != "" || stdout != want+"\n" {
							t.Errorf("v%d %s %v: got stdout=%q stderr=%q code=%d; want %q",
								f.version, inputName, args, stdout, stderr, code, want+"\n")
						}

						args = append(append([]string{"--pure"}, base...), input)
						stdout, stderr, code = runCommand(t, args...)
						if code != 0 || stderr != "" || stdout != want {
							t.Errorf("v%d %s %v: got stdout=%q stderr=%q code=%d; want %q",
								f.version, inputName, args, stdout, stderr, code, want)
						}
					}
				}
			}
		}
	}
}

func TestConvertDefaultCase(t *testing.T) {
	stdout, _, code := runCommand(t, "--format=human", strings.ToUpper(fixtureV7.formats["hex"]))
	if code != 0 || stdout != fixtureV7.formats["human"]+"\n" {
		t.Fatalf("expected lowercase output by default, got %q (code %d)", stdout, code)
	}
}

// A base64url value may start with "-"; "--" ends option parsing.
func TestConvertLeadingDash(t *testing.T) {
	stdout, stderr, code := runCommand(t, "--pure", "--format=human", "--", "-KGyw9TlT2CKe4ydDh8qOw")
	if code != 0 || stdout != "f8a1b2c3-d4e5-4f60-8a7b-8c9d0e1f2a3b" {
		t.Fatalf("got stdout=%q stderr=%q code=%d", stdout, stderr, code)
	}
}

// TestConvertRoundTrip converts freshly generated values back to their own
// format, and through every other format, without changing them.
func TestConvertRoundTrip(t *testing.T) {
	for _, version := range []string{"-4", "-7"} {
		for _, format := range allFormats {
			for _, letterCase := range []string{"lower", "upper"} {
				gen := []string{version, "-n", "3", "--format=" + format, "--case=" + letterCase}
				generated, _, code := runCommand(t, gen...)
				if code != 0 {
					t.Fatalf("%v failed", gen)
				}
				for _, value := range strings.Fields(generated) {
					for _, via := range allFormats {
						mid, stderr, code := runCommand(t, "--pure", version, "--format="+via, "--", value)
						if code != 0 {
							t.Fatalf("%q -> %s: %s", value, via, stderr)
						}
						back, stderr, code := runCommand(t, "--pure", version, "--format="+format, "--case="+letterCase, "--", mid)
						if code != 0 || back != value {
							t.Fatalf("%q -> %s %q -> %s: got %q (%s)", value, via, mid, format, back, stderr)
						}
					}
				}
			}
		}
	}
}

func TestValidatePositional(t *testing.T) {
	for _, f := range fixtures {
		for _, format := range allFormats {
			value := f.formats[format]
			stdout, stderr, code := runCommand(t, value)
			if code != 0 || stderr != "" || stdout != f.validMessage() {
				t.Errorf("%q: got stdout=%q stderr=%q code=%d", value, stdout, stderr, code)
			}
			stdout, stderr, code = runCommand(t, "--pure", value)
			if code != 0 || stdout != "" || stderr != "" {
				t.Errorf("--pure %q: got stdout=%q stderr=%q code=%d", value, stdout, stderr, code)
			}
		}
	}
	for _, args := range [][]string{
		{"--pure", "not-a-uuid"},
		{"--pure", "00000000-0000-4000-0000-000000000000"},
		{"--pure", "-4", fixtureV7.formats["human"]},
	} {
		stdout, stderr, code := runCommand(t, args...)
		if code != 1 || stdout != "" || stderr != "" {
			t.Errorf("%v: expected silent failure, got stdout=%q stderr=%q code=%d", args, stdout, stderr, code)
		}
	}
}

func TestConvertErrors(t *testing.T) {
	v7 := fixtureV7.formats["human"]
	tests := []struct {
		name   string
		args   []string
		stderr string
	}{
		{"unrecognized format", []string{"--format=hex", "not-a-uuid"},
			"error: convert uuid: invalid UUID: unsupported format or parse failure\n"},
		{"wrong variant", []string{"--format=hex", "00000000-0000-4000-0000-000000000000"},
			"error: convert uuid: invalid UUID: unsupported variant Reserved\n"},
		{"version 1", []string{"--format=hex", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
			"error: convert uuid: unsupported UUID version 1: expected 4 or 7\n"},
		{"version mismatch", []string{"--format=hex", "-4", v7},
			"error: convert uuid: UUID version 7 does not match constraint 4\n"},
		{"version mismatch with pure", []string{"--pure", "--format=hex", "-4", v7},
			"error: convert uuid: UUID version 7 does not match constraint 4\n"},
		{"unknown target format", []string{"--format=octal", v7},
			"error: convert uuid: unsupported format: use human, hex, base64, base64url, or int\n"},
		{"32-digit integer", []string{"--format=human", "11111111111141111111111111111111"},
			"error: convert uuid: invalid UUID: unsupported variant Reserved\n"},
		{"two values", []string{"--format=hex", v7, v7},
			"error: expected a single value, got 2: [" + v7 + " " + v7 + "] (options must come before the value)\n"},
		{"option after value", []string{v7, "--format=hex"},
			"error: expected a single value, got 2: [" + v7 + " --format=hex] (options must come before the value)\n"},
		{"count with value", []string{"-n", "2", v7}, "error: -n cannot be used with a value\n"},
		{"case without format", []string{"--case=upper", v7}, "error: -case requires --format\n"},
		{"pure without value", []string{"--pure"}, "error: --pure requires a value to validate or convert\n"},
		{"pure with validate", []string{"--pure", "--validate=" + v7}, "error: --pure cannot be used with --validate\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runCommand(t, tt.args...)
			if code != 1 || stdout != "" || stderr != tt.stderr {
				t.Fatalf("got stdout=%q stderr=%q code=%d; want stderr=%q", stdout, stderr, code, tt.stderr)
			}
		})
	}
}
