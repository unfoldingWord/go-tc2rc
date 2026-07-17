package tc2rc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/unfoldingWord/go-tc2rc/rc"
	"gopkg.in/yaml.v3"
)

func TestConvertFixture(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "rc")
	rep := Convert(context.Background(), filepath.Join("testdata", "aa_ult_jon_book"), outDir, Options{})
	if !rep.OK {
		t.Fatalf("Convert failed: %s", rep.Error)
	}
	if rep.Book != "jon" {
		t.Errorf("Book = %q, want jon", rep.Book)
	}
	if rep.USFM != "aa_ult_jon_book.usfm" {
		t.Errorf("USFM = %q, want aa_ult_jon_book.usfm (not the jon.usfm stub)", rep.USFM)
	}

	// The RC repo must hold the USFM (renamed), manifest, license, and README.
	for _, name := range []string{"JON.usfm", "manifest.yaml", "LICENSE.md", "README.md"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("missing output file %s: %v", name, err)
		}
	}

	buf, err := os.ReadFile(filepath.Join(outDir, "manifest.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var m rc.Manifest
	if err := yaml.Unmarshal(buf, &m); err != nil {
		t.Fatal(err)
	}

	dc := m.DublinCore
	if dc.Conformsto != "rc0.2" {
		t.Errorf("conformsto = %q", dc.Conformsto)
	}
	if dc.Subject != "Aligned Bible" {
		t.Errorf("subject = %q, want Aligned Bible", dc.Subject)
	}
	if dc.Identifier != "ult" {
		t.Errorf("identifier = %q, want ult", dc.Identifier)
	}
	if dc.Format != "text/usfm3" {
		t.Errorf("format = %q, want text/usfm3", dc.Format)
	}
	if dc.Type != "bundle" {
		t.Errorf("type = %q, want bundle", dc.Type)
	}
	if dc.Language.Identifier != "aa" || dc.Language.Title != "Afaraf" || dc.Language.Direction != "ltr" {
		t.Errorf("language = %+v", dc.Language)
	}
	if dc.Rights != "CC BY-SA 4.0" {
		t.Errorf("rights = %q", dc.Rights)
	}
	if dc.Issued != "2020-12-24" {
		t.Errorf("issued = %q, want 2020-12-24 (from time_created)", dc.Issued)
	}
	if len(dc.Relation) != 2 || dc.Relation[0] != "en/ult" {
		t.Errorf("relation = %v", dc.Relation)
	}
	if len(dc.Source) != 1 || dc.Source[0].Identifier != "ult" || dc.Source[0].Version != "5" {
		t.Errorf("source = %+v", dc.Source)
	}
	if len(dc.Contributor) != 2 || dc.Contributor[0] != "Test Translator" || dc.Contributor[1] != "Test Checker" {
		t.Errorf("contributor = %v", dc.Contributor)
	}
	// Fixture resource.name is empty, so the title falls back to the \mt book title.
	if dc.Title != "Yunus Kitab" {
		t.Errorf("title = %q, want Yunus Kitab (from \\mt)", dc.Title)
	}

	if len(m.Checking.CheckingEntity) != 1 || m.Checking.CheckingEntity[0] != "Test Checker" {
		t.Errorf("checking_entity = %v", m.Checking.CheckingEntity)
	}

	if len(m.Projects) != 1 {
		t.Fatalf("projects = %d, want 1", len(m.Projects))
	}
	p := m.Projects[0]
	if p.Identifier != "jon" || p.Sort != 32 || p.Path != "./JON.usfm" || p.Versification != "ufw" {
		t.Errorf("project = %+v", p)
	}
	if p.Title != "Yunus Kitab" {
		t.Errorf("project title = %q, want Yunus Kitab (\\mt beats \\h and book.name)", p.Title)
	}
	if len(p.Categories) != 1 || p.Categories[0] != "bible-ot" {
		t.Errorf("categories = %v", p.Categories)
	}

	// Content must be byte-identical to the selected source USFM.
	src, err := os.ReadFile(filepath.Join("testdata", "aa_ult_jon_book", "aa_ult_jon_book.usfm"))
	if err != nil {
		t.Fatal(err)
	}
	dst, err := os.ReadFile(filepath.Join(outDir, "JON.usfm"))
	if err != nil {
		t.Fatal(err)
	}
	if string(src) != string(dst) {
		t.Error("JON.usfm differs from source aa_ult_jon_book.usfm")
	}

	// Alignment data dir and settings.json must not be copied.
	// (jon.usfm is omitted: it collides with JON.usfm on case-insensitive filesystems.)
	for _, name := range []string{"jon", "settings.json", "aa_ult_jon_book.usfm"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err == nil {
			t.Errorf("unexpected %s in RC output", name)
		}
	}
}

func TestConvertRepoNameOverride(t *testing.T) {
	// Simulate a checkout dir whose basename differs from the DCS repo name.
	inDir := makeRepo(t, "checkout-xyz", map[string]string{
		"manifest.json": `{"tc_version": 8, "project": {"id": "jon", "name": "Jonah"}, "type": {"id": "text"}, "resource": {"id": "ult"}, "target_language": {"id": "en", "name": "English", "direction": "ltr"}}`,
		"aa_ult_jon_book.usfm": `\id JON
\h Jonah
\c 1
\v 1 Hi.
`,
		"jon.usfm": "\\id JON stub\n\\c 1\n\\v 1 Stub stub stub stub stub.\n",
	})
	outDir := filepath.Join(t.TempDir(), "rc")
	rep := Convert(context.Background(), inDir, outDir, Options{RepoName: "aa_ult_jon_book"})
	if !rep.OK {
		t.Fatalf("Convert failed: %s", rep.Error)
	}
	if rep.USFM != "aa_ult_jon_book.usfm" {
		t.Errorf("USFM = %q, want aa_ult_jon_book.usfm via RepoName override", rep.USFM)
	}
}
