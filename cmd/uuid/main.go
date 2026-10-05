package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/PrInStPL/Uuid-Go/uuidgen"
	"github.com/google/uuid"
)

var (
	buildVersion = "dev"
	buildDate    = "unknown"
	buildNumber  = "0"
)

type versionFlag struct {
	value int
	set   bool
}

func (v *versionFlag) Set(s string) error {
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	v.value = parsed
	v.set = true
	return nil
}

func (v *versionFlag) String() string {
	return strconv.Itoa(v.value)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run executes the CLI and returns the process exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	name := filepath.Base(os.Args[0])
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		uuidVersion = &versionFlag{value: 4}
		format      = fs.String("format", "human", "Output format: human, hex, base64, base64url, int")
		letterCase  = fs.String("case", "lower", "Output letter case for human and hex formats: lower or upper")
		count       = fs.Int("n", 1, "Number of UUIDs to generate")
		showInfo    bool
		quickV4     bool
		quickV7     bool
		validateStr string
		pure        bool
	)

	fs.Var(uuidVersion, "uuid", "UUID version to generate (4 or 7)")
	fs.BoolVar(&showInfo, "help", false, "Display build information and usage (exclusive)")
	fs.BoolVar(&showInfo, "h", false, "Display build information and usage (exclusive shorthand)")
	fs.BoolVar(&quickV4, "4", false, "Shortcut for UUID version 4")
	fs.BoolVar(&quickV7, "7", false, "Shortcut for UUID version 7")
	fs.StringVar(&validateStr, "validate", "", "Validate the provided UUID value (any output format); \"-\" reads values from stdin")
	fs.BoolVar(&pure, "pure", false, "With <value>: print only the converted value without a trailing newline, or nothing at all when only validating")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintf(out, "UUID generator\n\n")
		fmt.Fprintf(out, "Usage: %s [options]\n", name)
		fmt.Fprintf(out, "       %s [--format=<format>] [--pure] [options] <value>\n", name)
		fmt.Fprintf(out, "       %s --validate=<id> [<id>...]\n\n", name)
		fmt.Fprintf(out, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(out, "\nExamples:\n")
		fmt.Fprintf(out, "  %s                     # UUIDv4 human readable lowercase (default)\n", name)
		fmt.Fprintf(out, "  %s --uuid=7            # UUIDv7 human readable lowercase\n", name)
		fmt.Fprintf(out, "  %s -7 -n 5             # five UUIDv7, one per line\n", name)
		fmt.Fprintf(out, "  %s --format=base64     # UUIDv4 encoded in base64\n", name)
		fmt.Fprintf(out, "  %s --case=upper        # UUIDv4 human readable uppercase\n", name)
		fmt.Fprintf(out, "  %s <id>                # Validate provided UUID (auto-detect format)\n", name)
		fmt.Fprintf(out, "  %s --format=hex <id>   # Validate and convert UUID to another format\n", name)
		fmt.Fprintf(out, "  %s --pure <id>         # Validate silently, result in exit code only\n", name)
		fmt.Fprintf(out, "  %s --validate=<id>     # Validate provided UUID (auto-detect format)\n", name)
		fmt.Fprintf(out, "  %s --validate=- < ids  # Validate one UUID per line from stdin\n", name)
		fmt.Fprintf(out, "  %s --help              # Display build information and usage\n", name)
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if showInfo {
		fs.SetOutput(stdout)
		fs.Usage()
		printBuildInfo(stdout)
		return 0
	}

	fail := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "error: "+format+"\n", a...)
		return 1
	}

	if quickV4 && quickV7 {
		return fail("conflicting UUID shortcuts: choose only one of -4 or -7")
	}

	selectedVersion := uuidVersion.value
	if quickV4 {
		selectedVersion = 4
	}
	if quickV7 {
		selectedVersion = 7
	}

	uuidExplicit := quickV4 || quickV7 || uuidVersion.set

	if pure && validateStr != "" {
		return fail("--pure cannot be used with --validate")
	}

	if validateStr != "" {
		for _, flagName := range []string{"format", "case", "n"} {
			if isFlagSet(fs, flagName) {
				return fail("-%s cannot be used with --validate", flagName)
			}
		}
		forced := 0
		if uuidExplicit {
			forced = selectedVersion
		}

		if validateStr != "-" && fs.NArg() == 0 {
			id, err := uuidgen.Validate(validateStr, forced)
			if err != nil {
				return fail("validate uuid: %v", err)
			}
			fmt.Fprintln(stdout, describe(id))
			return 0
		}

		values, err := validateInputs(validateStr, fs.Args(), stdin)
		if err != nil {
			return fail("%v", err)
		}
		code := 0
		for _, value := range values {
			id, err := uuidgen.Validate(value, forced)
			if err != nil {
				fmt.Fprintf(stdout, "%s: invalid: %v\n", value, err)
				code = 1
				continue
			}
			fmt.Fprintf(stdout, "%s: %s\n", value, describe(id))
		}
		return code
	}

	if fs.NArg() > 0 {
		forced := 0
		if uuidExplicit {
			forced = selectedVersion
		}
		return convert(fs, forced, pure, stdout, stderr)
	}
	if pure {
		return fail("--pure requires a value to validate or convert")
	}
	if *count < 1 {
		return fail("-n must be at least 1")
	}
	if err := uuidgen.CheckFormat(*format, *letterCase); err != nil {
		return fail("format uuid: %v", err)
	}

	w := bufio.NewWriter(stdout)
	defer w.Flush()
	for i := 0; i < *count; i++ {
		uuidValue, err := uuidgen.Generate(selectedVersion)
		if err != nil {
			w.Flush()
			return fail("generate uuid: %v", err)
		}
		output, err := uuidgen.Format(uuidValue, *format, *letterCase)
		if err != nil {
			w.Flush()
			return fail("format uuid: %v", err)
		}
		fmt.Fprintln(w, output)
	}
	return 0
}

// convert validates the single positional value and, when --format is given,
// prints it in that format; without --format it only reports the validation result.
// Each step (format detection, validation, conversion) stops at the first error.
func convert(fs *flag.FlagSet, forced int, pure bool, stdout, stderr io.Writer) int {
	fail := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "error: "+format+"\n", a...)
		return 1
	}
	if fs.NArg() > 1 {
		return fail("expected a single value, got %d: %v (options must come before the value)", fs.NArg(), fs.Args())
	}
	if isFlagSet(fs, "n") {
		return fail("-n cannot be used with a value")
	}
	value := fs.Arg(0)
	format := fs.Lookup("format").Value.String()
	letterCase := fs.Lookup("case").Value.String()

	if !isFlagSet(fs, "format") {
		if isFlagSet(fs, "case") {
			return fail("-case requires --format")
		}
		id, err := uuidgen.Validate(value, forced)
		if pure {
			if err != nil {
				return 1
			}
			return 0
		}
		if err != nil {
			return fail("validate uuid: %v", err)
		}
		fmt.Fprintln(stdout, describe(id))
		return 0
	}

	output, err := uuidgen.Convert(value, forced, format, letterCase)
	if err != nil {
		return fail("convert uuid: %v", err)
	}
	if pure {
		fmt.Fprint(stdout, output)
	} else {
		fmt.Fprintln(stdout, output)
	}
	return 0
}

// validateInputs collects the values to validate: positional arguments after the
// --validate value, or one value per non-empty stdin line when --validate is "-".
func validateInputs(first string, rest []string, stdin io.Reader) ([]string, error) {
	if first != "-" {
		return append([]string{first}, rest...), nil
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("unexpected arguments with --validate=-: %v", rest)
	}
	var values []string
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			values = append(values, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stdin: %v", err)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("no values to validate on stdin")
	}
	return values, nil
}

func describe(id uuid.UUID) string {
	msg := fmt.Sprintf("valid uuid version %d", id.Version())
	if ts, ok := uuidgen.Timestamp(id); ok {
		msg += fmt.Sprintf(" (timestamp %s)", ts.Format("2006-01-02T15:04:05.000Z07:00"))
	}
	return msg
}

func isFlagSet(fs *flag.FlagSet, name string) bool {
	set := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

// printBuildInfo prints build metadata, falling back to the information embedded by
// the Go toolchain (e.g. for `go install`) when it was not set via -ldflags.
func printBuildInfo(w io.Writer) {
	info, _ := debug.ReadBuildInfo()
	writeBuildInfo(w, info, buildVersion, buildDate, buildNumber)
}

// pseudoVersion matches a complete Go pseudo-version, capturing its commit time and
// revision. It follows golang.org/x/mod/module's grammar, which allows three forms:
// vX.0.0-yyyymmddhhmmss-rev, vX.Y.Z-pre.0.yyyymmddhhmmss-rev and
// vX.Y.(Z+1)-0.yyyymmddhhmmss-rev, each optionally followed by +build metadata.
// Ordinary tags that merely end like one (e.g. v1.2.3-rc.20261005090906-4e4593dd695e) do not match.
var pseudoVersion = regexp.MustCompile(`^v[0-9]+\.(?:0\.0-|[0-9]+\.[0-9]+-(?:[^+]*\.)?0\.)([0-9]{14})-([A-Za-z0-9]+)(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

// writeBuildInfo prints version, build date, build number and revision. Values set via
// -ldflags take precedence, then VCS settings, then the module pseudo-version.
func writeBuildInfo(w io.Writer, info *debug.BuildInfo, version, date, number string) {
	revision, modified := "", false
	if info != nil {
		if version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.time":
				if date == "unknown" {
					if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
						date = t.UTC().Format(time.RFC3339)
					}
				}
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
		// Modules built by `go install pkg@version` carry no VCS settings, but a
		// pseudo-version encodes the commit time (UTC) and an abbreviated revision.
		if m := pseudoVersion.FindStringSubmatch(info.Main.Version); m != nil {
			if date == "unknown" {
				if t, err := time.Parse("20060102150405", m[1]); err == nil {
					date = t.UTC().Format(time.RFC3339)
				}
			}
			if revision == "" {
				revision = m[2]
			}
		}
	}
	if modified && revision != "" {
		revision += "-dirty"
	}
	fmt.Fprintf(w, "\nversion: %s\n", version)
	fmt.Fprintf(w, "build date: %s\n", date)
	fmt.Fprintf(w, "build number: %s\n", number)
	if revision != "" {
		fmt.Fprintf(w, "revision: %s\n", revision)
	}
}
