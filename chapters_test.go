package tc2rc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unfoldingWord/go-tc2rc/books"
	"github.com/unfoldingWord/go-tc2rc/rc"
	"gopkg.in/yaml.v3"
)

// writeTree writes files (paths relative to dir, subdirectories created) and returns dir.
func writeTree(t *testing.T, dir string, files map[string]string) string {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// A tc-desktop save made before tc_version: no .usfm, the text in <book>/<chapter>.json.
func TestConvertFromChapterJSON(t *testing.T) {
	inDir := writeTree(t, filepath.Join(t.TempDir(), "en_tit"), map[string]string{
		"manifest.json":       `{"generator": {"name": "tc-desktop", "build": ""}, "package_version": 1, "project": {"id": "tit", "name": "Titus"}, "target_language": {"id": "en", "name": "English", "direction": "ltr"}, "license": "CC BY-SA 4.0"}`,
		"tit/manifest.json":   `{"language_id": "en", "subject": "Bible", "resource_id": "targetLanguage"}`,
		"tit/1.json":          `{"2": "with the certain hope", "1": "Paul, a servant of God", "10": "For there are many", "front": "intro"}`,
		"tit/3.json":          `{"1": "Remind them", "2-3": "to slander no one", "4": ""}`,
		"tit/2.json":          `{"1": "But you"}`,
		"tit/notes.txt":       "not a chapter",
		".apps/x/source.usfm": `\id TIT source text that must not be picked`,
	})
	outDir := filepath.Join(t.TempDir(), "rc")
	rep := Convert(context.Background(), inDir, outDir, Options{})
	if !rep.OK {
		t.Fatalf("Convert failed: %s", rep.Error)
	}
	if rep.ChapterDir != "tit" || rep.USFM != "" {
		t.Errorf("ChapterDir = %q, USFM = %q; want tit and empty", rep.ChapterDir, rep.USFM)
	}
	if !strings.Contains(strings.Join(rep.Warnings, "\n"), "unaligned") || !strings.Contains(strings.Join(rep.Warnings, "\n"), "1:front") {
		t.Errorf("warnings = %q, want the unaligned note and the skipped front key", rep.Warnings)
	}

	usfm, err := os.ReadFile(filepath.Join(outDir, "TIT.usfm"))
	if err != nil {
		t.Fatal(err)
	}
	want := `\id TIT
\usfm 3.0
\ide UTF-8
\h Titus
\toc1 Titus
\toc2 Titus
\toc3 Tit
\mt Titus
\c 1
\p
\v 1 Paul, a servant of God
\v 2 with the certain hope
\v 10 For there are many
\c 2
\p
\v 1 But you
\c 3
\p
\v 1 Remind them
\v 2-3 to slander no one
\v 4
`
	if string(usfm) != want {
		t.Errorf("TIT.usfm =\n%s\nwant\n%s", usfm, want)
	}

	buf, err := os.ReadFile(filepath.Join(outDir, "manifest.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var m rc.Manifest
	if err := yaml.Unmarshal(buf, &m); err != nil {
		t.Fatal(err)
	}
	if m.DublinCore.Subject != "Bible" || m.DublinCore.Format != "text/usfm3" {
		t.Errorf("subject/format = %q/%q, want Bible/text/usfm3", m.DublinCore.Subject, m.DublinCore.Format)
	}
	if len(m.Projects) != 1 || m.Projects[0].Identifier != "tit" || m.Projects[0].Path != "./TIT.usfm" || m.Projects[0].Title != "Titus" {
		t.Errorf("projects = %+v", m.Projects)
	}
}

// An exported .usfm wins over chapter JSON, so tc repos with both keep converting as before.
func TestConvertPrefersUSFMOverChapterJSON(t *testing.T) {
	inDir := writeTree(t, filepath.Join(t.TempDir(), "en_tit"), map[string]string{
		"manifest.json": `{"tc_version": 8, "project": {"id": "tit"}, "type": {"id": "text"}, "target_language": {"id": "en"}}`,
		"en_tit.usfm":   "\\id TIT\n\\usfm 3.0\n\\mt Titus\n\\c 1\n\\p\n\\v 1 aligned text\n",
		"tit/1.json":    `{"1": "draft text"}`,
	})
	rep := Convert(context.Background(), inDir, filepath.Join(t.TempDir(), "rc"), Options{})
	if !rep.OK {
		t.Fatalf("Convert failed: %s", rep.Error)
	}
	if rep.USFM != "en_tit.usfm" || rep.ChapterDir != "" {
		t.Errorf("USFM = %q, ChapterDir = %q; want en_tit.usfm and empty", rep.USFM, rep.ChapterDir)
	}
}

func TestBuildChapterUSFMNoVerses(t *testing.T) {
	inDir := writeTree(t, t.TempDir(), map[string]string{"tit/1.json": `{"front": "intro only"}`})
	b, _ := books.Lookup("tit")
	if _, err := buildChapterUSFM(inDir, b, "Titus"); err == nil || !strings.Contains(err.Error(), "no verses") {
		t.Errorf("err = %v, want a no-verses error", err)
	}
}
