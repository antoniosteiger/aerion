//go:build !(linux && webkit2_41)

package platform

// SetSpellcheck is a no-op off WebKitGTK; see spellcheck_linux.go.
func SetSpellcheck(enabled bool, languages []string) {}
