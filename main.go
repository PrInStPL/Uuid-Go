package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"uuid-go/uuidgen"
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
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	name := os.Args[0]
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		uuidVersion = &versionFlag{value: 4}
		format      = fs.String("format", "human", "Output format: human, base64, int")
		letterCase  = fs.String("case", "lower", "Output letter case for human format: lower or upper")
		showInfo    bool
		quickV4     bool
		quickV7     bool
		validateStr string
	)

	fs.Var(uuidVersion, "uuid", "UUID version to generate (4 or 7)")
	fs.BoolVar(&showInfo, "help", false, "Display build information and usage (exclusive)")
	fs.BoolVar(&showInfo, "h", false, "Display build information and usage (exclusive shorthand)")
	fs.BoolVar(&quickV4, "4", false, "Shortcut for UUID version 4")
	fs.BoolVar(&quickV7, "7", false, "Shortcut for UUID version 7")
	fs.StringVar(&validateStr, "validate", "", "Validate the provided UUID value (human, base64, or int)")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintf(out, "UUID generator\n\n")
		fmt.Fprintf(out, "Usage: %s [options]\n\n", name)
		fmt.Fprintf(out, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(out, "\nExamples:\n")
		fmt.Fprintf(out, "  %s                   # UUIDv4 human readable lowercase (default)\n", name)
		fmt.Fprintf(out, "  %s --uuid=7          # UUIDv7 human readable lowercase\n", name)
		fmt.Fprintf(out, "  %s -7                # UUIDv7 quick flag\n", name)
		fmt.Fprintf(out, "  %s --format=base64   # UUIDv4 encoded in base64\n", name)
		fmt.Fprintf(out, "  %s --case=upper      # UUIDv4 human readable uppercase\n", name)
		fmt.Fprintf(out, "  %s --validate=<id>   # Validate provided UUID (auto-detect format)\n", name)
		fmt.Fprintf(out, "  %s --help            # Display build information and usage\n", name)
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if showInfo {
		fs.SetOutput(stdout)
		fs.Usage()
		fmt.Fprintf(stdout, "\nversion: %s\n", buildVersion)
		fmt.Fprintf(stdout, "build date: %s\n", buildDate)
		fmt.Fprintf(stdout, "build number: %s\n", buildNumber)
		return 0
	}

	fail := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "error: "+format+"\n", a...)
		return 1
	}

	if fs.NArg() > 0 {
		return fail("unexpected arguments: %v", fs.Args())
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

	if validateStr != "" {
		forced := 0
		if uuidExplicit {
			forced = selectedVersion
		}
		version, err := uuidgen.Validate(validateStr, forced)
		if err != nil {
			return fail("validate uuid: %v", err)
		}
		fmt.Fprintf(stdout, "valid uuid version %d\n", version)
		return 0
	}

	if err := uuidgen.CheckFormat(*format, *letterCase); err != nil {
		return fail("format uuid: %v", err)
	}

	uuidValue, err := uuidgen.Generate(selectedVersion)
	if err != nil {
		return fail("generate uuid: %v", err)
	}

	output, err := uuidgen.Format(uuidValue, *format, *letterCase)
	if err != nil {
		return fail("format uuid: %v", err)
	}

	fmt.Fprintln(stdout, output)
	return 0
}
