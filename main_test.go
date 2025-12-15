package main

import (
	"encoding/base64"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func runCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestHelpFlagIsExclusive(t *testing.T) {
	output, err := runCommand(t, "--help", "--uuid=7", "--format=base64")
	if err != nil {
		t.Fatalf("help flag should exit successfully: %v\noutput: %s", err, output)
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
	output, err := runCommand(t, "-4")
	if err != nil {
		t.Fatalf("-4 flag failed: %v\noutput: %s", err, output)
	}
	id, parseErr := uuid.Parse(strings.TrimSpace(output))
	if parseErr != nil {
		t.Fatalf("failed to parse uuid output: %v", parseErr)
	}
	if id.Version() != 4 {
		t.Fatalf("expected version 4, got %d", id.Version())
	}

	output, err = runCommand(t, "--uuid=7")
	if err != nil {
		t.Fatalf("--uuid flag failed: %v\noutput: %s", err, output)
	}
	id, parseErr = uuid.Parse(strings.TrimSpace(output))
	if parseErr != nil {
		t.Fatalf("failed to parse uuid output: %v", parseErr)
	}
	if id.Version() != 7 {
		t.Fatalf("expected version 7, got %d", id.Version())
	}
}

func TestQuickFlagConflicts(t *testing.T) {
	output, err := runCommand(t, "-4", "-7")
	if err == nil {
		t.Fatalf("expected conflict error when using -4 and -7 together")
	}
	if !strings.Contains(output, "conflicting") {
		t.Fatalf("expected conflict message, got: %s", output)
	}
}

func TestValidateAutoDetect(t *testing.T) {
	id := uuid.New()
	output, err := runCommand(t, "--validate="+id.String())
	if err != nil {
		t.Fatalf("validation should succeed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "valid uuid version 4") {
		t.Fatalf("expected validation confirmation for version 4, got %s", output)
	}
}

func TestValidateWithConstraintMismatch(t *testing.T) {
	id, _ := uuid.NewV7()
	output, err := runCommand(t, "--validate="+id.String(), "--uuid=4")
	if err == nil {
		t.Fatalf("expected validation to fail for version mismatch")
	}
	if !strings.Contains(output, "does not match constraint") {
		t.Fatalf("expected constraint mismatch message, got: %s", output)
	}
}

func TestValidateBase64VersionSelection(t *testing.T) {
	id, _ := uuid.NewV7()
	encoded := base64.StdEncoding.EncodeToString(id[:])
	output, err := runCommand(t, "--validate="+encoded, "-7")
	if err != nil {
		t.Fatalf("validation should succeed for base64 input: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "valid uuid version 7") {
		t.Fatalf("expected base64 validation confirmation, got %s", output)
	}
}
