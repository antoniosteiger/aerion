package settings

import (
	"database/sql"
	"encoding/json"
	"os"
	"strings"
)

// spellcheckTags maps Aerion's spellcheck dictionary codes (the settings
// toggles, see frontend/src/lib/spellcheck/locales.ts) to the Enchant /
// Hunspell language tags the system spellchecker uses.
var spellcheckTags = map[string]string{
	"en":    "en_US",
	"en-gb": "en_GB",
	"cs":    "cs_CZ",
	"de":    "de_DE",
	"fr":    "fr_FR",
	"it":    "it_IT",
	"nb":    "nb_NO",
	"nl":    "nl_NL",
}

// SpellcheckLanguages returns the Enchant language tags to check for the
// enabled dictionary codes. An empty selection falls back to the UI language
// (if it has a dictionary) plus English, matching what the settings dialog
// shows as checked. uiLanguage is the saved language setting; when unset, the
// LANG environment variable stands in for the UI's system-locale detection.
func SpellcheckLanguages(codes []string, uiLanguage string) []string {
	if len(codes) == 0 {
		if uiLanguage == "" {
			uiLanguage = os.Getenv("LANG")
		}
		codes = []string{spellcheckCode(uiLanguage), "en"}
	}
	var tags []string
	seen := map[string]bool{}
	for _, code := range codes {
		tag, ok := spellcheckTags[strings.ToLower(code)]
		if !ok || seen[tag] {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	return tags
}

// spellcheckCode turns a locale ("de", "en-GB", "de_DE.UTF-8") into a
// dictionary code: the full code when there is a dictionary for it (en-gb),
// otherwise the base language.
func spellcheckCode(locale string) string {
	locale = strings.ToLower(strings.ReplaceAll(locale, "_", "-"))
	if i := strings.IndexAny(locale, ".@"); i >= 0 {
		locale = locale[:i]
	}
	if _, ok := spellcheckTags[locale]; ok {
		return locale
	}
	base, _, _ := strings.Cut(locale, "-")
	return base
}

// Spellcheck returns whether spellcheck is on and the Enchant language tags.
func (s *Store) Spellcheck() (bool, []string) {
	enabled, _ := s.GetSpellcheckEnabled()
	codes, _ := s.GetSpellcheckLanguages()
	uiLanguage, _ := s.GetLanguage()
	return enabled, SpellcheckLanguages(codes, uiLanguage)
}

// ReadSpellcheck reads the spellcheck settings directly from the database,
// before the app and its Store exist (like ReadNativeTitleBar). Defaults to
// enabled when the database can't be read.
func ReadSpellcheck(dbPath string) (bool, []string) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return true, SpellcheckLanguages(nil, "")
	}
	defer db.Close()

	read := func(key string) string {
		var value string
		_ = db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
		return value
	}
	var codes []string
	_ = json.Unmarshal([]byte(read(KeySpellcheckLanguages)), &codes)
	return read(KeySpellcheckEnabled) != "false", SpellcheckLanguages(codes, read(KeyLanguage))
}
