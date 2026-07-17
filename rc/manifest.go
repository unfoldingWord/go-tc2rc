// Package rc provides Resource Container 0.2 manifest types and helpers.
// Output is intentionally compatible with github.com/unfoldingWord/go-rc2sb.
package rc

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Manifest is an RC 0.2 manifest.yaml document.
type Manifest struct {
	DublinCore DublinCore `yaml:"dublin_core"`
	Checking   Checking   `yaml:"checking"`
	Projects   []Project  `yaml:"projects"`
}

// DublinCore holds the dublin_core section of an RC manifest.
type DublinCore struct {
	Conformsto  string   `yaml:"conformsto"`
	Contributor []string `yaml:"contributor"`
	Creator     string   `yaml:"creator"`
	Description string   `yaml:"description"`
	Format      string   `yaml:"format"`
	Identifier  string   `yaml:"identifier"`
	Issued      string   `yaml:"issued"`
	Language    Language `yaml:"language"`
	Modified    string   `yaml:"modified"`
	Publisher   string   `yaml:"publisher"`
	Relation    []string `yaml:"relation"`
	Rights      string   `yaml:"rights"`
	Source      []Source `yaml:"source"`
	Subject     string   `yaml:"subject"`
	Title       string   `yaml:"title"`
	Type        string   `yaml:"type"`
	Version     string   `yaml:"version"`
}

// Language is the RC language block.
type Language struct {
	Direction  string `yaml:"direction"`
	Identifier string `yaml:"identifier"`
	Title      string `yaml:"title"`
}

// Source is one dublin_core.source entry.
type Source struct {
	Identifier string `yaml:"identifier"`
	Language   string `yaml:"language"`
	Version    string `yaml:"version"`
}

// Checking is the RC checking section.
type Checking struct {
	CheckingEntity []string `yaml:"checking_entity"`
	CheckingLevel  string   `yaml:"checking_level"`
}

// Project is one RC projects[] entry.
type Project struct {
	Title         string   `yaml:"title"`
	Versification string   `yaml:"versification"`
	Identifier    string   `yaml:"identifier"`
	Sort          int      `yaml:"sort"`
	Path          string   `yaml:"path"`
	Categories    []string `yaml:"categories"`
}

// WriteManifest serializes the manifest to <outDir>/manifest.yaml.
func WriteManifest(outDir string, m *Manifest) error {
	buf, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	head := []byte("---\n")
	return os.WriteFile(filepath.Join(outDir, "manifest.yaml"), append(head, buf...), 0o644)
}

// RepoDate returns the last commit date (YYYY-MM-DD) of a git repo,
// falling back to today when git is unavailable.
func RepoDate(repoDir string) string {
	out, err := exec.Command("git", "-C", repoDir, "log", "-1", "--format=%cs").Output()
	if err == nil {
		if s := strings.TrimSpace(string(out)); len(s) == 10 {
			return s
		}
	}
	return time.Now().UTC().Format("2006-01-02")
}

// DetectRights greps LICENSE.md for a recognizable license; defaults to the
// translationCore standard CC BY-SA 4.0. fromFile is false when defaulted.
func DetectRights(repoDir string) (rights string, fromFile bool) {
	for _, name := range []string{"LICENSE.md", "LICENSE", "LICENSE.txt"} {
		buf, err := os.ReadFile(filepath.Join(repoDir, name))
		if err != nil {
			continue
		}
		s := string(buf)
		if strings.Contains(s, "Attribution-ShareAlike 4.0") || strings.Contains(s, "CC BY-SA 4.0") {
			return "CC BY-SA 4.0", true
		}
		if strings.Contains(s, "Attribution 4.0") {
			return "CC BY 4.0", true
		}
	}
	return "CC BY-SA 4.0", false
}
