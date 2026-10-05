package main

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func runCommand(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	return runWithStdin(t, "", args...)
}

func runWithStdin(t *testing.T, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(args, strings.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), code
}

func TestHelpFlagIsExclusive(t *testing.T) {
	output, stderr, code := runCommand(t, "--help", "--uuid=7", "--format=base64")
	if code != 0 {
		t.Fatalf("help flag should exit successfully, got %d\nstderr: %s", code, stderr)
	}
	usageIdx := strings.Index(output, "Usage:")
	if usageIdx == -1 {
		t.Fatalf("expected usage text in help output, got %q", output)
	}
	if !strings.Contains(output, "version: "+buildVersion) {
		t.Fatalf("expected build version %q in output, got %q", buildVersion, output)
	}
	versionIdx := strings.Index(output, "version:")
	if versionIdx < usageIdx {
		t.Fatalf("expected usage information to appear before build metadata, got: %s", output)
	}
	if !strings.Contains(output, "build date:") || !strings.Contains(output, "build number:") {
		t.Fatalf("expected build metadata in help output, got %q", output)
	}
	usageLine := output[usageIdx:]
	usageLine = usageLine[:strings.Index(usageLine, "\n")]
	if strings.Contains(usageLine, "/") {
		t.Fatalf("expected program name without directory in usage, got %q", usageLine)
	}
}

func TestQuickVersionFlags(t *testing.T) {
	tests := []struct {
		args    []string
		version uuid.Version
	}{
		{[]string{"-4"}, 4},
		{[]string{"-7"}, 7},
		{[]string{"--uuid=7"}, 7},
		{nil, 4},
	}
	for _, tt := range tests {
		output, stderr, code := runCommand(t, tt.args...)
		if code != 0 {
			t.Fatalf("%v failed with code %d\nstderr: %s", tt.args, code, stderr)
		}
		id, err := uuid.Parse(strings.TrimSpace(output))
		if err != nil {
			t.Fatalf("%v: failed to parse uuid output: %v", tt.args, err)
		}
		if id.Version() != tt.version {
			t.Fatalf("%v: expected version %d, got %d", tt.args, tt.version, id.Version())
		}
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"conflicting shortcuts", []string{"-4", "-7"}, "conflicting"},
		{"unsupported version", []string{"--uuid=5"}, "unsupported UUID version"},
		{"invalid format", []string{"--format=octal"}, "unsupported format"},
		{"positional args", []string{"extra"}, "unexpected arguments"},
		{"zero count", []string{"-n", "0"}, "-n must be at least 1"},
		{"validate with format", []string{"--validate=x", "--format=int"}, "-format cannot be used with --validate"},
		{"validate with case", []string{"--validate=x", "--case=upper"}, "-case cannot be used with --validate"},
		{"validate with count", []string{"--validate=x", "-n", "2"}, "-n cannot be used with --validate"},
		{"stdin with args", []string{"--validate=-", "extra"}, "unexpected arguments"},
		{"empty stdin", []string{"--validate=-"}, "no values to validate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, stderr, code := runCommand(t, tt.args...)
			if code == 0 {
				t.Fatalf("expected failure, got output %q", output)
			}
			if !strings.Contains(stderr, tt.want) {
				t.Fatalf("expected %q in stderr, got: %s", tt.want, stderr)
			}
			if !strings.HasPrefix(stderr, "error: ") {
				t.Fatalf("expected plain error prefix without timestamp, got: %s", stderr)
			}
		})
	}
}

func TestValidateAutoDetect(t *testing.T) {
	id := uuid.New()
	output, stderr, code := runCommand(t, "--validate="+id.String())
	if code != 0 {
		t.Fatalf("validation should succeed, got %d\nstderr: %s", code, stderr)
	}
	if !strings.Contains(output, "valid uuid version 4") {
		t.Fatalf("expected validation confirmation for version 4, got %s", output)
	}
}

func TestValidateWithConstraintMismatch(t *testing.T) {
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	_, stderr, code := runCommand(t, "--validate="+id.String(), "--uuid=4")
	if code == 0 {
		t.Fatalf("expected validation to fail for version mismatch")
	}
	if !strings.Contains(stderr, "does not match constraint") {
		t.Fatalf("expected constraint mismatch message, got: %s", stderr)
	}
}

func TestValidateBase64VersionSelection(t *testing.T) {
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(id[:])
	output, stderr, code := runCommand(t, "--validate="+encoded, "-7")
	if code != 0 {
		t.Fatalf("validation should succeed for base64 input, got %d\nstderr: %s", code, stderr)
	}
	if !strings.Contains(output, "valid uuid version 7") {
		t.Fatalf("expected base64 validation confirmation, got %s", output)
	}
}

func TestGenerateCount(t *testing.T) {
	output, stderr, code := runCommand(t, "-7", "-n", "5", "--format=hex", "--case=upper")
	if code != 0 {
		t.Fatalf("expected success, got %d\nstderr: %s", code, stderr)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d: %q", len(lines), output)
	}
	seen := map[string]bool{}
	for _, line := range lines {
		if len(line) != 32 || line != strings.ToUpper(line) {
			t.Fatalf("expected 32 uppercase hex characters, got %q", line)
		}
		id, err := uuid.Parse(line)
		if err != nil || id.Version() != 7 {
			t.Fatalf("expected UUIDv7, got %q (err %v)", line, err)
		}
		if seen[line] {
			t.Fatalf("duplicate uuid %q", line)
		}
		seen[line] = true
	}
}

func TestValidateV7ShowsTimestamp(t *testing.T) {
	id := uuid.MustParse("0192a0b1-c2d3-7e4f-8a5b-6c7d8e9fa0b1")
	output, stderr, code := runCommand(t, "--validate="+id.String())
	if code != 0 {
		t.Fatalf("expected success, got %d\nstderr: %s", code, stderr)
	}
	want := "valid uuid version 7 (timestamp 2024-10-18T17:34:17.299Z)\n"
	if output != want {
		t.Fatalf("expected %q, got %q", want, output)
	}
}

func TestValidateMultiple(t *testing.T) {
	v4 := uuid.New().String()
	v7 := uuid.Must(uuid.NewV7()).String()

	output, stderr, code := runCommand(t, "--validate="+v4, v7)
	if code != 0 {
		t.Fatalf("expected success, got %d\nstderr: %s\noutput: %s", code, stderr, output)
	}
	if !strings.Contains(output, v4+": valid uuid version 4\n") || !strings.Contains(output, v7+": valid uuid version 7") {
		t.Fatalf("unexpected report: %s", output)
	}

	output, _, code = runWithStdin(t, v4+"\n\n  bad  \n"+v7+"\n", "--validate=-", "-4")
	if code != 1 {
		t.Fatalf("expected exit code 1 when a value is invalid, got %d", code)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 report lines, got %q", output)
	}
	if !strings.HasPrefix(lines[0], v4+": valid") ||
		!strings.HasPrefix(lines[1], "bad: invalid:") ||
		!strings.HasPrefix(lines[2], v7+": invalid:") {
		t.Fatalf("unexpected report: %s", output)
	}
}
