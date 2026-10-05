package main

import (
	"bytes"
	"runtime/debug"
	"testing"
)

func TestWriteBuildInfo(t *testing.T) {
	const pseudo = "v0.0.0-20261005090906-4e4593dd695e"
	vcs := []debug.BuildSetting{
		{Key: "vcs.revision", Value: "4e4593dd695e1234567890abcdef1234567890ab"},
		{Key: "vcs.time", Value: "2026-10-05T11:09:06+02:00"},
		{Key: "vcs.modified", Value: "false"},
	}
	tests := []struct {
		name                  string
		info                  *debug.BuildInfo
		version, date, number string
		want                  string
	}{
		{"no build info", nil, "dev", "unknown", "0",
			"\nversion: dev\nbuild date: unknown\nbuild number: 0\n"},
		{"go install pseudo-version", &debug.BuildInfo{Main: debug.Module{Version: pseudo}}, "dev", "unknown", "0",
			"\nversion: " + pseudo + "\nbuild date: 2026-10-05T09:09:06Z\nbuild number: 0\nrevision: 4e4593dd695e\n"},
		{"pseudo-version after a tag", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.4-0.20261005090906-4e4593dd695e"}}, "dev", "unknown", "0",
			"\nversion: v1.2.4-0.20261005090906-4e4593dd695e\nbuild date: 2026-10-05T09:09:06Z\nbuild number: 0\nrevision: 4e4593dd695e\n"},
		{"pre-release pseudo-version with +dirty", &debug.BuildInfo{Main: debug.Module{Version: "v1.3.0-rc.1.0.20261005090906-4e4593dd695e+dirty"}}, "dev", "unknown", "0",
			"\nversion: v1.3.0-rc.1.0.20261005090906-4e4593dd695e+dirty\nbuild date: 2026-10-05T09:09:06Z\nbuild number: 0\nrevision: 4e4593dd695e\n"},
		{"pseudo-version for a new major version", &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0-20261005090906-4e4593dd695e"}}, "dev", "unknown", "0",
			"\nversion: v2.0.0-20261005090906-4e4593dd695e\nbuild date: 2026-10-05T09:09:06Z\nbuild number: 0\nrevision: 4e4593dd695e\n"},
		// Ordinary tags that only end like a pseudo-version must not get a date or revision.
		{"tagged pre-release ending in timestamp and hash", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3-rc.20261005090906-4e4593dd695e"}}, "dev", "unknown", "0",
			"\nversion: v1.2.3-rc.20261005090906-4e4593dd695e\nbuild date: unknown\nbuild number: 0\n"},
		{"tag with timestamp but no 0. or 0.0 base", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.0-20261005090906-4e4593dd695e"}}, "dev", "unknown", "0",
			"\nversion: v1.2.0-20261005090906-4e4593dd695e\nbuild date: unknown\nbuild number: 0\n"},
		{"tagged release has no date", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, "dev", "unknown", "0",
			"\nversion: v1.2.3\nbuild date: unknown\nbuild number: 0\n"},
		{"local build", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev", "unknown", "0",
			"\nversion: dev\nbuild date: unknown\nbuild number: 0\n"},
		{"VCS settings win over pseudo-version", &debug.BuildInfo{Main: debug.Module{Version: pseudo}, Settings: vcs}, "dev", "unknown", "0",
			"\nversion: " + pseudo + "\nbuild date: 2026-10-05T09:09:06Z\nbuild number: 0\nrevision: 4e4593dd695e1234567890abcdef1234567890ab\n"},
		{"ldflags win over everything", &debug.BuildInfo{Main: debug.Module{Version: pseudo}, Settings: vcs}, "1.0.0", "2026-01-01", "42",
			"\nversion: 1.0.0\nbuild date: 2026-01-01\nbuild number: 42\nrevision: 4e4593dd695e1234567890abcdef1234567890ab\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			writeBuildInfo(&out, tt.info, tt.version, tt.date, tt.number)
			if out.String() != tt.want {
				t.Fatalf("got:\n%q\nwant:\n%q", out.String(), tt.want)
			}
		})
	}
}
