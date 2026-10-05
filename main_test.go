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
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
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
		{"invalid format", []string{"--format=hex"}, "unsupported format"},
		{"positional args", []string{"extra"}, "unexpected arguments"},
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
