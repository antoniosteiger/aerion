package settings

import (
	"reflect"
	"testing"
)

func TestSpellcheckLanguages(t *testing.T) {
	t.Setenv("LANG", "de_DE.UTF-8")
	tests := []struct {
		name  string
		codes []string
		ui    string
		want  []string
	}{
		{"explicit selection", []string{"en", "de"}, "fr", []string{"en_US", "de_DE"}},
		{"uk english stays uk", []string{"en-gb"}, "", []string{"en_GB"}},
		{"unknown codes dropped", []string{"vi", "nl"}, "", []string{"nl_NL"}},
		{"empty uses ui language plus english", nil, "fr", []string{"fr_FR", "en_US"}},
		{"empty with regional ui language", nil, "en-GB", []string{"en_GB", "en_US"}},
		{"empty without dictionary for ui language", nil, "zh-TW", []string{"en_US"}},
		{"empty and unset ui language uses LANG", nil, "", []string{"de_DE", "en_US"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SpellcheckLanguages(tt.codes, tt.ui); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SpellcheckLanguages(%v, %q) = %v, want %v", tt.codes, tt.ui, got, tt.want)
			}
		})
	}
}
