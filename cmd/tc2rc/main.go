// Command tc2rc converts a translationCore repository to a Resource Container (RC).
//
// Usage:
//
//	tc2rc [flags] <tcRepoDir> <rcOutDir>
//
// Flags:
//
//	--repo-name <name>  DCS repo name to use when locating the exported USFM
//	                    file (<name>.usfm). Defaults to the input directory
//	                    basename.
//
// A conversion report is written to <rcOutDir>.report.json. Exit code is 0 when
// the conversion produced a usable RC repo, 1 otherwise.
//
// Pipeline with go-rc2sb (tc → RC → SB):
//
//	tc2rc <tcDir> <rcDir>
//	rc2sb <rcDir> <sbDir>
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tc2rc "github.com/unfoldingWord/go-tc2rc"
)

func main() {
	repoName := flag.String("repo-name", "", "DCS repo name used to locate the exported USFM file (defaults to input directory basename)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `tc2rc — convert a translationCore repo to a Resource Container repo

Usage:
  tc2rc [flags] <tcRepoDir> <rcOutDir>

A conversion report is written to <rcOutDir>.report.json. Exit code is 0 when
the conversion produced a usable RC repo, 1 otherwise.

Pipeline (tc → RC → SB):
  tc2rc <tcDir> <rcDir>
  rc2sb <rcDir> <sbDir>

Flags:
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	inDir, outDir := flag.Arg(0), flag.Arg(1)

	rep := tc2rc.Convert(context.Background(), inDir, outDir, tc2rc.Options{RepoName: *repoName})

	reportPath := filepath.Clean(outDir) + ".report.json"
	if err := tc2rc.WriteReport(rep, reportPath); err != nil {
		fmt.Fprintf(os.Stderr, "writing report: %v\n", err)
	}

	if !rep.OK {
		fmt.Fprintf(os.Stderr, "FAILED %s: %s\n", inDir, rep.Error)
		os.Exit(1)
	}
	fmt.Printf("OK %s -> %s (book=%s, tc_version=%d, usfm=%s, warnings=%d)\n",
		inDir, outDir, rep.Book, rep.TCVersion, rep.USFM, len(rep.Warnings))
}
