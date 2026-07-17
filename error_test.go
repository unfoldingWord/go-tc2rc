package tc2rc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertMissingManifest(t *testing.T) {
	rep := Convert(context.Background(), t.TempDir(), filepath.Join(t.TempDir(), "out"), Options{})
	if rep.OK {
		t.Fatal("expected failure for missing manifest.json")
	}
}

func TestConvertNonTextType(t *testing.T) {
	inDir := makeRepo(t, "repo", map[string]string{
		"manifest.json": `{"tc_version": 8, "project": {"id": "jon"}, "type": {"id": "tn"}, "target_language": {"id": "en"}}`,
	})
	rep := Convert(context.Background(), inDir, filepath.Join(t.TempDir(), "out"), Options{})
	if rep.OK {
		t.Fatal("expected failure for non-text tc type")
	}
}

func TestConvertBadBookCode(t *testing.T) {
	inDir := makeRepo(t, "repo", map[string]string{
		"manifest.json": `{"tc_version": 8, "project": {"id": "xyz"}, "type": {"id": "text"}, "target_language": {"id": "en"}}`,
	})
	rep := Convert(context.Background(), inDir, filepath.Join(t.TempDir(), "out"), Options{})
	if rep.OK {
		t.Fatal("expected failure for invalid book code")
	}
}

// A repo without a .usfm file fails, even when tc chapter JSON data exists.
func TestConvertNoUSFM(t *testing.T) {
	inDir := makeRepo(t, "repo", map[string]string{
		"manifest.json": `{"tc_version": 8, "project": {"id": "jon"}, "type": {"id": "text"}, "target_language": {"id": "en"}}`,
	})
	if err := os.MkdirAll(filepath.Join(inDir, "jon"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "jon", "1.json"), []byte(`{"1": "In the beginning"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(t.TempDir(), "out")
	rep := Convert(context.Background(), inDir, outDir, Options{})
	if rep.OK {
		t.Fatal("expected failure when no .usfm file exists")
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Error("outDir should not exist after failed conversion")
	}
}

func TestConvertCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	outDir := filepath.Join(t.TempDir(), "out")
	rep := Convert(ctx, filepath.Join("testdata", "aa_ult_jon_book"), outDir, Options{})
	if rep.OK {
		t.Fatal("expected failure for cancelled context")
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Error("outDir should not exist after cancelled conversion")
	}
}
