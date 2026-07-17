package tc2rc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/unfoldingWord/go-tc2rc/books"
	"github.com/unfoldingWord/go-tc2rc/rc"
	"github.com/unfoldingWord/go-tc2rc/tc"
)

// copyFile copies src to dst, creating or truncating dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// copyFirstOf copies the first existing candidate file from inDir to
// outDir/<destName>, reporting whether one was found.
func copyFirstOf(inDir, outDir, destName string, candidates []string) (bool, error) {
	for _, name := range candidates {
		src := filepath.Join(inDir, name)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyFile(src, filepath.Join(outDir, destName)); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// dedupeStrings returns items with duplicates and blanks removed, order kept.
func dedupeStrings(items []string) []string {
	seen := make(map[string]bool, len(items))
	var out []string
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// issuedDate extracts YYYY-MM-DD from a manifest time_created timestamp,
// falling back to the repo's last commit date.
func issuedDate(timeCreated, inDir string) string {
	if len(timeCreated) >= 10 {
		date := timeCreated[:10]
		if date[4] == '-' && date[7] == '-' {
			return date
		}
	}
	return rc.RepoDate(inDir)
}

// Convert converts one translationCore repository at inDir into an RC
// repository at outDir.
//
// It never panics. Failures are recorded on the returned Report (OK == false).
// Warnings are non-fatal. On hard conversion failure the incomplete outDir is removed.
//
// The resulting RC (subject "Aligned Bible", one book per repo) is suitable
// for conversion to Scripture Burrito via github.com/unfoldingWord/go-rc2sb.
func Convert(ctx context.Context, inDir, outDir string, opts Options) *Report {
	rep := &Report{Source: inDir, Output: outDir}

	fail := func(err error) *Report {
		rep.OK = false
		rep.Error = err.Error()
		return rep
	}

	if err := ctx.Err(); err != nil {
		return fail(fmt.Errorf("context error: %w", err))
	}

	m, err := tc.LoadManifest(inDir)
	if err != nil {
		return fail(err)
	}
	rep.TCVersion = m.TCVersion
	if m.TypeID != "text" {
		return fail(fmt.Errorf("unsupported tc type %q (only \"text\" scripture projects are supported)", m.TypeID))
	}
	book, ok := books.Lookup(m.ProjectID)
	if !ok {
		return fail(fmt.Errorf("project id %q is not a valid book code", m.ProjectID))
	}
	rep.Book = book.Code
	if m.Language.ID == "" {
		return fail(fmt.Errorf("manifest has no target_language.id"))
	}

	code := strings.ToUpper(book.Code)
	repoName := opts.RepoName
	if repoName == "" {
		repoName = filepath.Base(filepath.Clean(inDir))
	}
	usfmName, err := findUSFM(inDir, repoName, book.Code)
	if err != nil {
		return fail(err)
	}
	rep.USFM = usfmName
	if !strings.EqualFold(usfmName, repoName+".usfm") {
		rep.Warnf("no %s.usfm found; selected %s", repoName, usfmName)
	}
	header, err := readUSFMHeader(filepath.Join(inDir, usfmName))
	if err != nil {
		return fail(fmt.Errorf("reading %s: %w", usfmName, err))
	}
	if header.IDCode != "" && header.IDCode != code {
		rep.Warnf("%s has \\id %s but manifest project is %s", usfmName, header.IDCode, code)
	}

	title := header.BookTitle()
	if title == "" {
		title = m.BookName
		if title != "" {
			rep.Warnf("no \\mt, \\mt1 or \\h marker in %s; using target_language.book.name %q", rep.USFM, title)
		}
	}
	if title == "" {
		title = m.ProjectName
	}
	if title == "" {
		title = book.Title
		rep.Warnf("no book title found; using canonical %q", title)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fail(err)
	}
	cleanup := func(e error) *Report {
		os.RemoveAll(outDir) // don't leave a half-written RC repo behind
		return fail(e)
	}

	outUSFM := code + ".usfm"
	if err := copyFile(filepath.Join(inDir, usfmName), filepath.Join(outDir, outUSFM)); err != nil {
		return cleanup(err)
	}

	identifier := m.ResourceID
	if identifier == "" {
		identifier = "reg"
		rep.Warnf("empty resource id; defaulting identifier to %q", identifier)
	}
	rcTitle := m.ResourceName
	if rcTitle == "" {
		rcTitle = title
	}
	langTitle := m.Language.Name
	if langTitle == "" {
		langTitle = m.Language.ID
	}
	contributors := dedupeStrings(append(append([]string{}, m.Translators...), m.Checkers...))
	if len(contributors) == 0 {
		contributors = []string{"Door43 World Missions Community"}
		rep.Warnf("no translators or checkers in manifest; using default contributor")
	}
	checkingEntity := dedupeStrings(m.Checkers)
	if len(checkingEntity) == 0 {
		checkingEntity = []string{"Door43 World Missions Community"}
	}
	rights := m.License
	if rights == "" {
		var fromFile bool
		rights, fromFile = rc.DetectRights(inDir)
		if !fromFile {
			rep.Warnf("no license in manifest or license file; defaulting rights to %q", rights)
		}
	}
	format := "text/usfm"
	if header.USFM3 {
		format = "text/usfm3"
	}
	var sources []rc.Source
	for _, s := range m.Sources {
		sources = append(sources, rc.Source{Identifier: s.Identifier, Language: s.Language, Version: s.Version})
	}
	relation := m.Relation
	if relation == nil {
		relation = []string{}
	}
	generator := strings.TrimSpace(m.GeneratorName + " " + m.GeneratorBuild)
	if generator == "" {
		generator = "translationCore"
	}

	man := &rc.Manifest{
		DublinCore: rc.DublinCore{
			Conformsto:  "rc0.2",
			Contributor: contributors,
			Creator:     "Door43 World Missions Community",
			Description: fmt.Sprintf("Converted from translationCore project (tc_version %d, %s)", m.TCVersion, generator),
			Format:      format,
			Identifier:  identifier,
			Issued:      issuedDate(m.TimeCreated, inDir),
			Language: rc.Language{
				Direction:  m.Language.Direction,
				Identifier: m.Language.ID,
				Title:      langTitle,
			},
			Modified:  rc.RepoDate(inDir),
			Publisher: "Door43",
			Relation:  relation,
			Rights:    rights,
			Source:    sources,
			Subject:   "Aligned Bible",
			Title:     rcTitle,
			Type:      "bundle",
			Version:   "1",
		},
		Checking: rc.Checking{
			CheckingEntity: checkingEntity,
			CheckingLevel:  "1",
		},
		Projects: []rc.Project{{
			Title:         title,
			Versification: "ufw",
			Identifier:    book.Code,
			Sort:          book.Sort,
			Path:          "./" + outUSFM,
			Categories:    []string{book.Category},
		}},
	}
	if err := rc.WriteManifest(outDir, man); err != nil {
		return cleanup(err)
	}

	if err := ctx.Err(); err != nil {
		return cleanup(fmt.Errorf("context error: %w", err))
	}

	copied, err := copyFirstOf(inDir, outDir, "LICENSE.md", []string{"LICENSE.md", "LICENSE", "LICENSE.txt"})
	if err != nil {
		rep.Warnf("copying license: %v", err)
	}
	if !copied {
		if err := writeDefaultLicense(outDir); err != nil {
			rep.Warnf("writing default license: %v", err)
		} else {
			rep.Warnf("no license file in source; wrote default unfoldingWord CC BY-SA 4.0 LICENSE.md")
		}
	}
	if _, err := copyFirstOf(inDir, outDir, "README.md", []string{"README.md", "README", "README.txt", "readme.md"}); err != nil {
		rep.Warnf("copying README: %v", err)
	}
	rep.OK = true
	return rep
}

// ResultFromReport builds a Result summary from a completed Report.
func ResultFromReport(rep *Report) Result {
	if rep == nil {
		return Result{}
	}
	return Result{
		Book:      rep.Book,
		TCVersion: rep.TCVersion,
		InDir:     rep.Source,
		OutDir:    rep.Output,
		Warnings:  append([]string(nil), rep.Warnings...),
	}
}

// WriteReport writes the conversion report as JSON to path.
func WriteReport(rep *Report, path string) error {
	buf, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(buf, '\n'), 0o644)
}
