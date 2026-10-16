//go:build !linux

package tray

// The tray icon is Linux-only for now; these are no-ops elsewhere.

func Start(Actions)  {}
func SetUnread(bool) {}
func Stop()          {}
