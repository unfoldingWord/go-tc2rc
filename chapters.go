package tc2rc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/unfoldingWord/go-tc2rc/books"
)

var (
	chapterFileRe = regexp.MustCompile(`^([0-9]+)\.json$`)
	verseKeyRe    = regexp.MustCompile(`^([0-9]+)(?:-([0-9]+))?$`)
)

// chapterUSFM is a USFM file built from a tc project's <book>/<chapter>.json files.
type chapterUSFM struct {
	Dir      string // book directory the chapters were read from, relative to the repo
	Content  string
	Chapters int
	Verses   int
	Skipped  []string // non-verse keys left out, as "<chapter>:<key>"
}

// findBookDir returns the name of the repo's <book> directory, matched
// case-insensitively, or "" when there is none.
func findBookDir(inDir, bookCode string) string {
	entries, err := os.ReadDir(inDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() && strings.EqualFold(e.Name(), bookCode) {
			return e.Name()
		}
	}
	return ""
}

type verseText struct {
	key        string
	start, end int
	text       string
}

// buildChapterUSFM builds an unaligned USFM file from <inDir>/<book>/<chapter>.json,
// where each chapter file maps verse numbers (or ranges like "1-2") to text.
// translationCore saves made before tc_version keep their text this way, with
// no exported USFM. It returns an error when the repo has no chapter files.
func buildChapterUSFM(inDir string, book books.BookInfo, title string) (*chapterUSFM, error) {
	dir := findBookDir(inDir, book.Code)
	if dir == "" {
		return nil, fmt.Errorf("no %s/ directory of chapter JSON files", book.Code)
	}
	entries, err := os.ReadDir(filepath.Join(inDir, dir))
	if err != nil {
		return nil, err
	}
	type chapterFile struct {
		num  int
		name string
	}
	var files []chapterFile
	for _, e := range entries {
		m := chapterFileRe.FindStringSubmatch(e.Name())
		if e.IsDir() || m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		if n > 0 {
			files = append(files, chapterFile{n, e.Name()})
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no chapter JSON files (<chapter>.json) in %s/", dir)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].num < files[j].num })

	code := strings.ToUpper(book.Code)
	out := &chapterUSFM{Dir: dir}
	var b strings.Builder
	fmt.Fprintf(&b, "\\id %s\n\\usfm 3.0\n\\ide UTF-8\n", code)
	fmt.Fprintf(&b, "\\h %s\n\\toc1 %s\n\\toc2 %s\n\\toc3 %s\n\\mt %s\n", title, title, title, strings.ToUpper(book.Code[:1])+book.Code[1:], title)

	for _, f := range files {
		buf, err := os.ReadFile(filepath.Join(inDir, dir, f.name))
		if err != nil {
			return nil, err
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(strings.TrimPrefix(string(buf), "\uFEFF")), &raw); err != nil {
			return nil, fmt.Errorf("parsing %s/%s: %w", dir, f.name, err)
		}
		var verses []verseText
		for key, v := range raw {
			m := verseKeyRe.FindStringSubmatch(key)
			if m == nil {
				out.Skipped = append(out.Skipped, fmt.Sprintf("%d:%s", f.num, key))
				continue
			}
			start, _ := strconv.Atoi(m[1])
			end := start
			if m[2] != "" {
				end, _ = strconv.Atoi(m[2])
			}
			text, _ := v.(string)
			verses = append(verses, verseText{key, start, end, strings.TrimSpace(text)})
		}
		if len(verses) == 0 {
			continue
		}
		sort.Slice(verses, func(i, j int) bool {
			if verses[i].start != verses[j].start {
				return verses[i].start < verses[j].start
			}
			return verses[i].end < verses[j].end
		})
		fmt.Fprintf(&b, "\\c %d\n\\p\n", f.num)
		for _, v := range verses {
			if v.text == "" {
				fmt.Fprintf(&b, "\\v %s\n", v.key)
			} else {
				fmt.Fprintf(&b, "\\v %s %s\n", v.key, v.text)
			}
		}
		out.Chapters++
		out.Verses += len(verses)
	}
	if out.Chapters == 0 {
		return nil, fmt.Errorf("the chapter JSON files in %s/ have no verses", dir)
	}
	sort.Strings(out.Skipped)
	out.Content = b.String()
	return out, nil
}
