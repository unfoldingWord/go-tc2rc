// Package tc2rc converts translationCore (tc) repositories to Resource Container (RC) format.
package tc2rc

import "fmt"

// Options configures a translationCore to RC conversion.
type Options struct {
	// RepoName overrides the repo name used to locate the exported USFM file
	// (<RepoName>.usfm). If empty, the basename of the input directory is used.
	RepoName string
}

// Report is the fail-safe conversion record for one tc → RC conversion.
// Check OK to determine whether a usable RC repo was produced.
type Report struct {
	// Source is the input translationCore directory.
	Source string `json:"source"`

	// Output is the output RC directory.
	Output string `json:"output"`

	// Book is the lowercase book code converted (e.g. "jon").
	Book string `json:"book"`

	// TCVersion is the manifest tc_version.
	TCVersion int `json:"tc_version"`

	// USFM is the source .usfm filename selected from the tc repo.
	USFM string `json:"usfm"`

	// ChapterDir is set when the repo had no .usfm file: the book directory
	// whose <chapter>.json files the USFM was built from. That text is unaligned.
	ChapterDir string `json:"chapter_dir,omitempty"`

	// OK is true when a usable RC repository was written.
	OK bool `json:"ok"`

	// Error is set when OK is false.
	Error string `json:"error,omitempty"`

	// Err is the error behind Error, for callers that need to tell a failure
	// caused by the input (e.g. a manifest it can't read) from one caused by
	// the system (e.g. a full disk) with errors.Is and errors.As.
	Err error `json:"-"`

	// Warnings are non-fatal issues encountered during conversion.
	Warnings []string `json:"warnings,omitempty"`
}

// Warnf appends a formatted warning to the report.
func (r *Report) Warnf(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

// Result holds summary information about a completed conversion.
// It mirrors go-rc2sb's Result shape for library consumers.
type Result struct {
	// Book is the lowercase book code converted.
	Book string

	// TCVersion is the manifest tc_version.
	TCVersion int

	// InDir is the input translationCore directory.
	InDir string

	// OutDir is the output RC directory.
	OutDir string

	// ChapterDir is set when the USFM was built from chapter JSON (see Report.ChapterDir).
	ChapterDir string

	// Warnings is the list of non-fatal conversion warnings.
	Warnings []string
}
