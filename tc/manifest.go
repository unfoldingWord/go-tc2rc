// Package tc parses and normalizes translationCore manifest.json files.
package tc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Manifest is the normalized view of a translationCore manifest.json.
// It is parsed from a map so that lenient value typings (int/string/junk)
// across tc_version 7/8 and tS-generated hybrids normalize without error.
type Manifest struct {
	TCVersion      int    // tc_version
	TCEditVersion  string // tc_edit_version (e.g. "3.6.10")
	TypeID         string // type.id, normally "text"
	ProjectID      string // book code, e.g. "jon", "1th"
	ProjectName    string // English book name, e.g. "Jonah"
	ResourceID     string // resource.id lowercased, e.g. "ult", "irv", "reg"
	ResourceName   string // resource.name, e.g. "ULT - Jonah" (often empty)
	GeneratorName  string
	GeneratorBuild string
	License        string // e.g. "CC BY-SA 4.0"
	TimeCreated    string // ISO timestamp, e.g. "2019-06-13T15:08:30.339Z"
	Language       Language
	BookName       string // target_language.book.name (localized, may be junk)
	Relation       []string
	Sources        []Source
	Translators    []string
	Checkers       []string
}

// Language is the target language of a tc project.
type Language struct {
	ID        string
	Name      string
	Direction string
}

// Source is one entry from source_translations.
type Source struct {
	Identifier string // resource id, e.g. ult, ulb
	Language   string
	Version    string
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

func asInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(t))
		return n
	case json.Number:
		n, _ := t.Int64()
		return int(n)
	default:
		return 0
	}
}

func asBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(strings.TrimSpace(t), "true")
	default:
		return false
	}
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	return map[string]any{}
}

func asStringSlice(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, e := range items {
		switch t := e.(type) {
		case string:
			if s := strings.TrimSpace(t); s != "" {
				out = append(out, s)
			}
		case map[string]any:
			if s := strings.TrimSpace(asString(t["name"])); s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

// LoadManifest reads and normalizes <repoDir>/manifest.json.
func LoadManifest(repoDir string) (*Manifest, error) {
	buf, err := os.ReadFile(filepath.Join(repoDir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("reading manifest.json: %w", err)
	}
	buf = []byte(strings.TrimPrefix(string(buf), "\uFEFF"))
	var raw map[string]any
	if err := json.Unmarshal(buf, &raw); err != nil {
		return nil, fmt.Errorf("parsing manifest.json: %w", err)
	}

	m := &Manifest{}
	m.TCVersion = asInt(raw["tc_version"])
	// Saves made before tc_version existed still name tc-desktop as their generator.
	if m.TCVersion == 0 && !asBool(raw["tcInitialized"]) && asString(getMap(raw, "generator")["name"]) != "tc-desktop" {
		return nil, fmt.Errorf("manifest.json has no tc_version, tcInitialized is not true and the generator is not tc-desktop; not a translationCore repo")
	}
	m.TCEditVersion = strings.TrimSpace(asString(raw["tc_edit_version"]))

	proj := getMap(raw, "project")
	m.ProjectID = strings.ToLower(strings.TrimSpace(asString(proj["id"])))
	m.ProjectName = strings.TrimSpace(asString(proj["name"]))
	if m.ProjectID == "" {
		tsProj := getMap(raw, "ts_project")
		m.ProjectID = strings.ToLower(strings.TrimSpace(asString(tsProj["id"])))
		if m.ProjectName == "" {
			m.ProjectName = strings.TrimSpace(asString(tsProj["name"]))
		}
	}

	m.TypeID = strings.ToLower(strings.TrimSpace(asString(getMap(raw, "type")["id"])))
	if m.TypeID == "" {
		m.TypeID = "text"
	}

	res := getMap(raw, "resource")
	m.ResourceID = strings.ToLower(strings.TrimSpace(asString(res["id"])))
	m.ResourceName = strings.TrimSpace(asString(res["name"]))

	gen := getMap(raw, "generator")
	m.GeneratorName = asString(gen["name"])
	m.GeneratorBuild = asString(gen["build"])

	m.License = strings.TrimSpace(asString(raw["license"]))
	m.TimeCreated = strings.TrimSpace(asString(raw["time_created"]))

	tl := getMap(raw, "target_language")
	m.Language = Language{
		ID:        strings.TrimSpace(asString(tl["id"])),
		Name:      strings.TrimSpace(asString(tl["name"])),
		Direction: strings.TrimSpace(asString(tl["direction"])),
	}
	if m.Language.Direction == "" {
		m.Language.Direction = "ltr"
	}
	m.BookName = strings.TrimSpace(asString(getMap(tl, "book")["name"]))

	m.Relation = asStringSlice(raw["tsv_relation"])

	if st, ok := raw["source_translations"].([]any); ok {
		for _, e := range st {
			em, ok := e.(map[string]any)
			if !ok {
				continue
			}
			src := Source{
				Identifier: strings.ToLower(strings.TrimSpace(asString(em["resource_id"]))),
				Language:   strings.TrimSpace(asString(em["language_id"])),
				Version:    strings.TrimSpace(asString(em["version"])),
			}
			if src.Identifier == "" && src.Language == "" {
				continue
			}
			m.Sources = append(m.Sources, src)
		}
	}

	m.Translators = asStringSlice(raw["translators"])
	m.Checkers = asStringSlice(raw["checkers"])

	return m, nil
}
