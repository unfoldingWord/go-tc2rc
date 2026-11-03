package tc2rc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

// A repo with neither a .usfm file nor chapter JSON fails and leaves no output.
func TestConvertNoUSFM(t *testing.T) {
	inDir := makeRepo(t, "repo", map[string]string{
		"manifest.json": `{"tc_version": 8, "project": {"id": "jon"}, "type": {"id": "text"}, "target_language": {"id": "en"}}`,
	})
	if err := os.MkdirAll(filepath.Join(inDir, "jon"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "jon", "manifest.json"), []byte(`{"subject": "Bible"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(t.TempDir(), "out")
	rep := Convert(context.Background(), inDir, outDir, Options{})
	if rep.OK {
		t.Fatal("expected failure when there is no .usfm file and no chapter JSON")
	}
	if !strings.Contains(rep.Error, "no .usfm file") || !strings.Contains(rep.Error, "no chapter JSON files") {
		t.Errorf("Error = %q, want it to name both missing inputs", rep.Error)
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
