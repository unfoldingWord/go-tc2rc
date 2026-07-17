// Package books provides Bible book code data used when building RC manifests
// and USFM/TSV content from translationCore projects.
package books

import "strings"

// BookInfo holds canonical data for a protocanonical book (or OBS).
type BookInfo struct {
	Code     string // lowercase project id, e.g. "gen", "1sa"
	Title    string // canonical English title (fallback when manifest has none)
	Sort     int    // RC manifest sort (gen=1 .. rev=66)
	USFMNum  string // USFM file number prefix (gen=01 .. mal=39, mat=41 .. rev=67)
	Category string // bible-ot | bible-nt
}

// AllBooks is the ordered list of all 66 Bible books.
var AllBooks = []BookInfo{
	{"gen", "Genesis", 1, "01", "bible-ot"},
	{"exo", "Exodus", 2, "02", "bible-ot"},
	{"lev", "Leviticus", 3, "03", "bible-ot"},
	{"num", "Numbers", 4, "04", "bible-ot"},
	{"deu", "Deuteronomy", 5, "05", "bible-ot"},
	{"jos", "Joshua", 6, "06", "bible-ot"},
	{"jdg", "Judges", 7, "07", "bible-ot"},
	{"rut", "Ruth", 8, "08", "bible-ot"},
	{"1sa", "1 Samuel", 9, "09", "bible-ot"},
	{"2sa", "2 Samuel", 10, "10", "bible-ot"},
	{"1ki", "1 Kings", 11, "11", "bible-ot"},
	{"2ki", "2 Kings", 12, "12", "bible-ot"},
	{"1ch", "1 Chronicles", 13, "13", "bible-ot"},
	{"2ch", "2 Chronicles", 14, "14", "bible-ot"},
	{"ezr", "Ezra", 15, "15", "bible-ot"},
	{"neh", "Nehemiah", 16, "16", "bible-ot"},
	{"est", "Esther", 17, "17", "bible-ot"},
	{"job", "Job", 18, "18", "bible-ot"},
	{"psa", "Psalms", 19, "19", "bible-ot"},
	{"pro", "Proverbs", 20, "20", "bible-ot"},
	{"ecc", "Ecclesiastes", 21, "21", "bible-ot"},
	{"sng", "Song of Songs", 22, "22", "bible-ot"},
	{"isa", "Isaiah", 23, "23", "bible-ot"},
	{"jer", "Jeremiah", 24, "24", "bible-ot"},
	{"lam", "Lamentations", 25, "25", "bible-ot"},
	{"ezk", "Ezekiel", 26, "26", "bible-ot"},
	{"dan", "Daniel", 27, "27", "bible-ot"},
	{"hos", "Hosea", 28, "28", "bible-ot"},
	{"jol", "Joel", 29, "29", "bible-ot"},
	{"amo", "Amos", 30, "30", "bible-ot"},
	{"oba", "Obadiah", 31, "31", "bible-ot"},
	{"jon", "Jonah", 32, "32", "bible-ot"},
	{"mic", "Micah", 33, "33", "bible-ot"},
	{"nam", "Nahum", 34, "34", "bible-ot"},
	{"hab", "Habakkuk", 35, "35", "bible-ot"},
	{"zep", "Zephaniah", 36, "36", "bible-ot"},
	{"hag", "Haggai", 37, "37", "bible-ot"},
	{"zec", "Zechariah", 38, "38", "bible-ot"},
	{"mal", "Malachi", 39, "39", "bible-ot"},
	{"mat", "Matthew", 40, "41", "bible-nt"},
	{"mrk", "Mark", 41, "42", "bible-nt"},
	{"luk", "Luke", 42, "43", "bible-nt"},
	{"jhn", "John", 43, "44", "bible-nt"},
	{"act", "Acts", 44, "45", "bible-nt"},
	{"rom", "Romans", 45, "46", "bible-nt"},
	{"1co", "1 Corinthians", 46, "47", "bible-nt"},
	{"2co", "2 Corinthians", 47, "48", "bible-nt"},
	{"gal", "Galatians", 48, "49", "bible-nt"},
	{"eph", "Ephesians", 49, "50", "bible-nt"},
	{"php", "Philippians", 50, "51", "bible-nt"},
	{"col", "Colossians", 51, "52", "bible-nt"},
	{"1th", "1 Thessalonians", 52, "53", "bible-nt"},
	{"2th", "2 Thessalonians", 53, "54", "bible-nt"},
	{"1ti", "1 Timothy", 54, "55", "bible-nt"},
	{"2ti", "2 Timothy", 55, "56", "bible-nt"},
	{"tit", "Titus", 56, "57", "bible-nt"},
	{"phm", "Philemon", 57, "58", "bible-nt"},
	{"heb", "Hebrews", 58, "59", "bible-nt"},
	{"jas", "James", 59, "60", "bible-nt"},
	{"1pe", "1 Peter", 60, "61", "bible-nt"},
	{"2pe", "2 Peter", 61, "62", "bible-nt"},
	{"1jn", "1 John", 62, "63", "bible-nt"},
	{"2jn", "2 John", 63, "64", "bible-nt"},
	{"3jn", "3 John", 64, "65", "bible-nt"},
	{"jud", "Jude", 65, "66", "bible-nt"},
	{"rev", "Revelation", 66, "67", "bible-nt"},
}

var bookMap = func() map[string]BookInfo {
	m := make(map[string]BookInfo, len(AllBooks))
	for _, b := range AllBooks {
		m[b.Code] = b
	}
	return m
}()

// Lookup returns book info for a lowercase or mixed-case project id.
func Lookup(code string) (BookInfo, bool) {
	b, ok := bookMap[strings.ToLower(code)]
	return b, ok
}

// OBS is the synthetic book entry used for Open Bible Stories projects.
var OBS = BookInfo{Code: "obs", Title: "Open Bible Stories", Sort: 0}
