// Spellcheck itself is WebKitGTK's system spellchecker (Enchant + Hunspell),
// configured backend-side from the saved settings (internal/settings/
// spellcheck.go). This only mirrors its default for the settings UI.
import { get } from 'svelte/store'
import { locale } from 'svelte-i18n'
import { appLocaleToDict } from './locales'

// The effective default dictionary set when the user hasn't chosen one: the UI
// language (if spellcheckable) plus English. Must match SpellcheckLanguages in
// internal/settings/spellcheck.go.
export function defaultSpellcheckLanguages(): string[] {
  const ui = appLocaleToDict(get(locale) as string)
  return [...new Set([ui, 'en'].filter((c): c is string => !!c))]
}
