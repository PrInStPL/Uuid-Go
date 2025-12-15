package main

import (
	"flag"
	"fmt"
	"log"
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
	return fmt.Sprintf("%d", v.value)
}

func main() {
	var (
		uuidVersion = &versionFlag{value: 4}
		format      = flag.String("format", "human", "Output format: human, base64, int")
		letterCase  = flag.String("case", "lower", "Output letter case for human format: lower or upper")
		showInfo    bool
		quickV4     bool
		quickV7     bool
		validateStr string
	)

	flag.Var(uuidVersion, "uuid", "UUID version to generate (4 or 7)")
	flag.BoolVar(&showInfo, "help", false, "Display build information and usage (exclusive)")
	flag.BoolVar(&showInfo, "h", false, "Display build information and usage (exclusive shorthand)")
	flag.BoolVar(&quickV4, "4", false, "Shortcut for UUID version 4")
	flag.BoolVar(&quickV7, "7", false, "Shortcut for UUID version 7")
	flag.StringVar(&validateStr, "validate", "", "Validate the provided UUID value (human, base64, or int)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "UUID generator\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s                   # UUIDv4 human readable lowercase (default)\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s --uuid=7          # UUIDv7 human readable lowercase\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -7                # UUIDv7 quick flag\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s --format=base64   # UUIDv4 encoded in base64\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s --case=upper      # UUIDv4 human readable uppercase\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s --validate=<id>   # Validate provided UUID (auto-detect format)\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s --help            # Display build information and usage\n", os.Args[0])
	}

	flag.Parse()

	if showInfo {
		flag.Usage()
		fmt.Printf("\nversion: %s\n", buildVersion)
		fmt.Printf("build date: %s\n", buildDate)
		fmt.Printf("build number: %s\n", buildNumber)
		return
	}

	if quickV4 && quickV7 {
		log.Fatalf("conflicting UUID shortcuts: choose only one of -4 or -7")
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
		forced := ""
		if uuidExplicit {
			forced = fmt.Sprintf("%d", selectedVersion)
		}
		version, err := uuidgen.Validate(validateStr, forced)
		if err != nil {
			log.Fatalf("validate uuid: %v", err)
		}
		fmt.Printf("valid uuid version %d\n", version)
		return
	}

	uuidValue, err := uuidgen.Generate(fmt.Sprintf("%d", selectedVersion))
	if err != nil {
		log.Fatalf("generate uuid: %v", err)
	}

	output, err := uuidgen.Format(uuidValue, *format, *letterCase)
	if err != nil {
		log.Fatalf("format uuid: %v", err)
	}

	fmt.Println(output)
}
