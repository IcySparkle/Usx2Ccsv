package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"usxtocsv/convert"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var inputs stringSlice
	output := flag.String("output", "", "Output folder (optional)")
	help := flag.Bool("help", false, "Show help")
	quiet := flag.Bool("quiet", false, "Suppress progress output")
	jsonOut := flag.Bool("json", false, "Output JSON summary to stdout")
	flag.Var(&inputs, "input", "Input file/folder/wildcard path (repeatable)")
	flag.Parse()

	if *help || len(inputs) == 0 {
		showUsage()
		return
	}

	inputItems, err := convert.ResolveInputItems(inputs)
	if err != nil {
		fail(err.Error(), *jsonOut)
	}

	files, err := convert.CollectFiles(inputItems)
	if err != nil {
		fail(err.Error(), *jsonOut)
	}

	if len(files) == 0 {
		fail("No .usx, .usfm, or .sfm files found.", *jsonOut)
	}

	summary, err := convert.ConvertFiles(files, *output, convert.Options{Quiet: *quiet})
	if err != nil {
		fail(err.Error(), *jsonOut)
	}

	if *jsonOut {
		writeJSONSummary(summary)
		return
	}

	fmt.Println("All conversions completed.")
}

func showUsage() {
	fmt.Println("usxtocsv (Go) - Convert USX/USFM/SFM to CSV")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  usxtocsv -input <file|folder|wildcard> [-output <folder>]")
	fmt.Println("  usxtocsv -input <path1> -input <path2>")
	fmt.Println("  usxtocsv -quiet -json")
	fmt.Println("  usxtocsv -help")
}

func writeJSONSummary(runSummary convert.Summary) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(runSummary)
}

func fail(message string, jsonOut bool) {
	if jsonOut {
		out := map[string]string{"error": message}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
