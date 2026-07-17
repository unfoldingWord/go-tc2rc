package tc2rc

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findUSFM picks the USFM file to convert from a tc repo. Selection priority:
//
//  1. <repoName>.usfm (case-insensitive)
//  2. <projectID>.usfm (case-insensitive), e.g. amo.usfm for Amos
//  3. the first .usfm file in the repo root (alphabetical)
func findUSFM(inDir, repoName, projectID string) (string, error) {
	entries, err := os.ReadDir(inDir)
	if err != nil {
		return "", err
	}
	var candidates []string
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".usfm") {
			continue
		}
		candidates = append(candidates, e.Name())
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no .usfm file found in %s", inDir)
	}
	for _, want := range []string{repoName + ".usfm", projectID + ".usfm"} {
		for _, c := range candidates {
			if strings.EqualFold(c, want) {
				return c, nil
			}
		}
	}
	return candidates[0], nil // os.ReadDir returns entries sorted by filename
}

// usfmHeader holds the identification markers read from the top of a USFM file.
type usfmHeader struct {
	IDCode string // first token after \id, e.g. "JON"
	H      string // \h
	MT     string // \mt
	MT1    string // \mt1
	USFM3  bool   // file declares \usfm 3.x
}

// marker returns (text, true) when line starts with the exact USFM marker
// (e.g. `\mt` matches `\mt Jonah` but not `\mt1 Jonah`).
func marker(line, tag string) (string, bool) {
	if line == tag {
		return "", true
	}
	if strings.HasPrefix(line, tag+" ") {
		return strings.TrimSpace(line[len(tag)+1:]), true
	}
	return "", false
}

// readUSFMHeader scans the first headerLines lines of a USFM file for the
// \id, \h, \mt, \mt1 and \usfm markers.
func readUSFMHeader(path string) (usfmHeader, error) {
	const headerLines = 50
	var h usfmHeader
	f, err := os.Open(path)
	if err != nil {
		return h, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for i := 0; scanner.Scan() && i < headerLines; i++ {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if !strings.HasPrefix(line, `\`) {
			continue
		}
		if text, ok := marker(line, `\id`); ok {
			if fields := strings.Fields(text); len(fields) > 0 {
				h.IDCode = strings.ToUpper(fields[0])
			}
		} else if text, ok := marker(line, `\h`); ok && h.H == "" {
			h.H = text
		} else if text, ok := marker(line, `\mt`); ok && h.MT == "" {
			h.MT = text
		} else if text, ok := marker(line, `\mt1`); ok && h.MT1 == "" {
			h.MT1 = text
		} else if text, ok := marker(line, `\usfm`); ok {
			if strings.HasPrefix(strings.TrimSpace(text), "3") {
				h.USFM3 = true
			}
		}
	}
	return h, scanner.Err()
}

// BookTitle returns the localized book name using the priority
// \mt, then \mt1, then \h.
func (h usfmHeader) BookTitle() string {
	for _, t := range []string{h.MT, h.MT1, h.H} {
		if t != "" {
			return t
		}
	}
	return ""
}
