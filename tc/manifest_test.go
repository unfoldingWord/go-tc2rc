package tc

import (
	"os"
	"path/filepath"
	"testing"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

const tcDesktopManifest = `{
  "project": { "id": "1th", "name": "1 Thessalonians" },
  "resource": { "id": "IRV", "name": "" },
  "generator": { "name": "tc-desktop", "build": "" },
  "target_language": {
    "id": "as", "name": "Assamese", "direction": "ltr",
    "book": { "name": "১ থিচল" }
  },
  "ts_project": { "id": "1th", "name": "1 Thessalonians" },
  "type": { "id": "text", "name": "Text" },
  "source_translations": [
    { "language_id": "en", "resource_id": "ULT", "checking_level": "", "version": "5" }
  ],
  "translators": ["Alice"],
  "checkers": [{ "name": "Bob" }],
  "time_created": "2020-12-24T10:20:25.034Z",
  "tcInitialized": true,
  "tc_version": 7,
  "license": "CC BY-SA 4.0",
  "tc_edit_version": "3.0.0",
  "tsv_relation": [ "en/ult", "hbo/uhb?v=2.1.14" ]
}`

func TestLoadManifestTCDesktop(t *testing.T) {
	dir := writeManifest(t, tcDesktopManifest)
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.TCVersion != 7 {
		t.Errorf("TCVersion = %d, want 7", m.TCVersion)
	}
	if m.TypeID != "text" {
		t.Errorf("TypeID = %q, want text", m.TypeID)
	}
	if m.ProjectID != "1th" || m.ProjectName != "1 Thessalonians" {
		t.Errorf("project = %q/%q", m.ProjectID, m.ProjectName)
	}
	if m.ResourceID != "irv" {
		t.Errorf("ResourceID = %q, want irv (lowercased)", m.ResourceID)
	}
	if m.Language.ID != "as" || m.Language.Name != "Assamese" || m.Language.Direction != "ltr" {
		t.Errorf("language = %+v", m.Language)
	}
	if m.BookName != "১ থিচল" {
		t.Errorf("BookName = %q", m.BookName)
	}
	if m.License != "CC BY-SA 4.0" {
		t.Errorf("License = %q", m.License)
	}
	if m.TimeCreated != "2020-12-24T10:20:25.034Z" {
		t.Errorf("TimeCreated = %q", m.TimeCreated)
	}
	if len(m.Relation) != 2 || m.Relation[0] != "en/ult" || m.Relation[1] != "hbo/uhb?v=2.1.14" {
		t.Errorf("Relation = %v", m.Relation)
	}
	if len(m.Sources) != 1 || m.Sources[0].Identifier != "ult" || m.Sources[0].Language != "en" || m.Sources[0].Version != "5" {
		t.Errorf("Sources = %+v", m.Sources)
	}
	if len(m.Translators) != 1 || m.Translators[0] != "Alice" {
		t.Errorf("Translators = %v", m.Translators)
	}
	if len(m.Checkers) != 1 || m.Checkers[0] != "Bob" {
		t.Errorf("Checkers = %v", m.Checkers)
	}
}

// ts-android generated repos can carry tc_version/tcInitialized alongside
// tS-style fields; they must still parse.
const tsAndroidHybridManifest = `{
  "generator": { "name": "ts-android", "build": 177 },
  "target_language": { "name": "Afaraf", "direction": "ltr", "id": "aa", "book": { "name": "" } },
  "project": { "id": "gal", "name": "Galatians" },
  "type": { "id": "text", "name": "Text" },
  "source_translations": [
    { "language_id": "en", "resource_id": "ulb", "checking_level": "3", "date_modified": 20170329, "version": "9" }
  ],
  "translators": ["jesse_test"],
  "checkers": [],
  "time_created": "2019-05-28T19:23:20.951Z",
  "tcInitialized": true,
  "package_version": 7,
  "format": "usfm",
  "resource": { "id": "reg", "name": "" },
  "tc_version": 7,
  "license": "CC BY-SA 4.0"
}`

func TestLoadManifestTSAndroidHybrid(t *testing.T) {
	dir := writeManifest(t, tsAndroidHybridManifest)
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.ProjectID != "gal" || m.ResourceID != "reg" {
		t.Errorf("project/resource = %q/%q", m.ProjectID, m.ResourceID)
	}
	if m.GeneratorName != "ts-android" || m.GeneratorBuild != "177" {
		t.Errorf("generator = %q %q", m.GeneratorName, m.GeneratorBuild)
	}
	if len(m.Relation) != 0 {
		t.Errorf("Relation = %v, want empty", m.Relation)
	}
}

func TestLoadManifestNotTC(t *testing.T) {
	dir := writeManifest(t, `{"package_version": 7, "project": {"id": "jon"}}`)
	if _, err := LoadManifest(dir); err == nil {
		t.Fatal("expected error for manifest without tc_version/tcInitialized")
	}
}

func TestLoadManifestMissing(t *testing.T) {
	if _, err := LoadManifest(t.TempDir()); err == nil {
		t.Fatal("expected error for missing manifest.json")
	}
}

func TestLoadManifestInvalidJSON(t *testing.T) {
	dir := writeManifest(t, `{not json`)
	if _, err := LoadManifest(dir); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadManifestDirectionDefault(t *testing.T) {
	dir := writeManifest(t, `{"tc_version": 8, "project": {"id": "jon"}, "type": {"id": "text"}, "target_language": {"id": "en", "name": "English"}}`)
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Language.Direction != "ltr" {
		t.Errorf("Direction = %q, want ltr default", m.Language.Direction)
	}
}

// tc-desktop saves made before tc_version and tcInitialized still load.
func TestLoadManifestTCDesktopWithoutVersion(t *testing.T) {
	dir := writeManifest(t, `{"generator": {"name": "tc-desktop", "build": ""}, "package_version": 1, "project": {"id": "tit", "name": "Titus"}, "target_language": {"id": "en"}}`)
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.TCVersion != 0 || m.ProjectID != "tit" || m.TypeID != "text" {
		t.Errorf("manifest = %+v", m)
	}
	if _, err := LoadManifest(writeManifest(t, `{"generator": {"name": "ts-android"}, "package_version": 7, "project": {"id": "tit"}}`)); err == nil {
		t.Error("a tS manifest with no tc markers loaded as translationCore")
	}
}
