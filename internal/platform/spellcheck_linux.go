//go:build linux && webkit2_41

package platform

/*
#cgo pkg-config: webkit2gtk-4.1
#include <stdlib.h>
#include <webkit2/webkit2.h>

typedef struct {
	gboolean enabled;
	gchar **languages; // NULL-terminated, owned
} spellcheck_request;

static gboolean apply_spellcheck_cb(gpointer data) {
	spellcheck_request *req = data;
	WebKitWebContext *ctx = webkit_web_context_get_default();
	webkit_web_context_set_spell_checking_enabled(ctx, req->enabled);
	if (req->enabled && req->languages != NULL && req->languages[0] != NULL) {
		webkit_web_context_set_spell_checking_languages(ctx, (const gchar * const *)req->languages);
	}
	g_strfreev(req->languages);
	g_free(req);
	return G_SOURCE_REMOVE;
}

static gchar **new_language_list(int n) {
	return g_new0(gchar *, n + 1);
}

static void set_language(gchar **list, int i, char *language) {
	list[i] = g_strdup(language);
}

// Runs on the GTK main thread: immediately when called there before the main
// loop owns the context (startup), otherwise dispatched to the main loop.
static void schedule_spellcheck(gboolean enabled, gchar **languages) {
	spellcheck_request *req = g_new0(spellcheck_request, 1);
	req->enabled = enabled;
	req->languages = languages;
	g_main_context_invoke(NULL, apply_spellcheck_cb, req);
}
*/
import "C"

import "unsafe"

// SetSpellcheck configures WebKitGTK's built-in spellchecker (Enchant, with
// the system's Hunspell dictionaries) for every webview of this process.
// languages are Enchant tags such as "en_US"; tags without an installed
// dictionary are skipped by WebKit. Safe to call from any goroutine.
func SetSpellcheck(enabled bool, languages []string) {
	list := C.new_language_list(C.int(len(languages)))
	for i, lang := range languages {
		cs := C.CString(lang)
		C.set_language(list, C.int(i), cs)
		C.free(unsafe.Pointer(cs))
	}
	var on C.gboolean
	if enabled {
		on = 1
	}
	C.schedule_spellcheck(on, list)
}
