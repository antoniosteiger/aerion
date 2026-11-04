//go:build linux

package main

import "os"

// WebKitGTK opens <input type="file"> pickers (e.g. the composer's attach
// button) via GtkFileChooserNative, which only talks to the XDG FileChooser
// portal inside a sandbox. Outside Flatpak it falls back to a plain GTK
// dialog that ignores portal backends such as xdg-desktop-portal-termfilechooser.
// Must run before GTK initializes, hence init(); the composer subprocess
// runs it too.
func init() {
	if os.Getenv("GTK_USE_PORTAL") == "" {
		_ = os.Setenv("GTK_USE_PORTAL", "1")
	}
}
