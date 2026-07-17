package tc2rc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeRepo creates a temp dir named repoName holding the given files.
func makeRepo(t *testing.T, repoName string, files map[string]string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), repoName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFindUSFM(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		project  string
		files    map[string]string
		want     string
	}{
		{
			name:     "exact repo name wins",
			repoName: "as_irv_1th_book",
			project:  "1th",
			files:    map[string]string{"as_irv_1th_book.usfm": "x", "1th.usfm": strings.Repeat("y", 100)},
			want:     "as_irv_1th_book.usfm",
		},
		{
			name:     "renamed repo falls back to book file",
			repoName: "acz_ssve_1pe_book",
			project:  "1pe",
			files:    map[string]string{"acz_ssv_1pe_book.usfm": "x", "1pe.usfm": "y"},
			want:     "1pe.usfm",
		},
		{
			name:     "date-prefixed repo falls back to book file",
			repoName: "200303_en_ult_jon_book",
			project:  "jon",
			files: map[string]string{
				"en_ult_jon_book.usfm": "x",
				"32-JON.usfm":          "y",
				"jon.usfm":             "z",
			},
			want: "jon.usfm",
		},
		{
			name:     "book file in unconventional repo",
			repoName: "AmosAlignment-MessarraJan8",
			project:  "amo",
			files:    map[string]string{"amo.usfm": "x"},
			want:     "amo.usfm",
		},
		{
			name:     "first alphabetical as last resort",
			repoName: "myrepo",
			project:  "jon",
			files:    map[string]string{"b.usfm": strings.Repeat("y", 100), "a.usfm": "x"},
			want:     "a.usfm",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := makeRepo(t, tt.repoName, tt.files)
			got, err := findUSFM(dir, tt.repoName, tt.project)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("findUSFM = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindUSFMNone(t *testing.T) {
	dir := makeRepo(t, "empty_repo", map[string]string{"manifest.json": "{}"})
	if _, err := findUSFM(dir, "empty_repo", "jon"); err == nil {
		t.Fatal("expected error when no .usfm file exists")
	}
}

func TestReadUSFMHeader(t *testing.T) {
	dir := makeRepo(t, "r", map[string]string{"b.usfm": `\id 1TH EN_IRV as_Assamese_ltr tc
\usfm 3.0
\ide UTF-8
\h Header Name
\toc1 Toc One
\mt MT Name
\c 1
\v 1 text
`})
	h, err := readUSFMHeader(filepath.Join(dir, "b.usfm"))
	if err != nil {
		t.Fatal(err)
	}
	if h.IDCode != "1TH" {
		t.Errorf("IDCode = %q, want 1TH", h.IDCode)
	}
	if !h.USFM3 {
		t.Error("USFM3 = false, want true")
	}
	if h.MT != "MT Name" || h.H != "Header Name" || h.MT1 != "" {
		t.Errorf("markers = mt:%q mt1:%q h:%q", h.MT, h.MT1, h.H)
	}
	if got := h.BookTitle(); got != "MT Name" {
		t.Errorf("BookTitle = %q, want MT Name (\\mt has priority)", got)
	}
}

func TestBookTitlePriority(t *testing.T) {
	tests := []struct {
		h    usfmHeader
		want string
	}{
		{usfmHeader{MT: "mt", MT1: "mt1", H: "h"}, "mt"},
		{usfmHeader{MT1: "mt1", H: "h"}, "mt1"},
		{usfmHeader{H: "h"}, "h"},
		{usfmHeader{}, ""},
	}
	for _, tt := range tests {
		if got := tt.h.BookTitle(); got != tt.want {
			t.Errorf("BookTitle(%+v) = %q, want %q", tt.h, got, tt.want)
		}
	}
}

// \mt must not swallow \mt1 (marker token matching is exact).
func TestReadUSFMHeaderMT1Only(t *testing.T) {
	dir := makeRepo(t, "r", map[string]string{"b.usfm": `\id 2CH
\h 2 Chronicles
\toc1 The Second Book of the Chronicles
\mt1 2 Chronicles Main
`})
	h, err := readUSFMHeader(filepath.Join(dir, "b.usfm"))
	if err != nil {
		t.Fatal(err)
	}
	if h.MT != "" {
		t.Errorf("MT = %q, want empty", h.MT)
	}
	if h.MT1 != "2 Chronicles Main" {
		t.Errorf("MT1 = %q", h.MT1)
	}
	if got := h.BookTitle(); got != "2 Chronicles Main" {
		t.Errorf("BookTitle = %q", got)
	}
}
