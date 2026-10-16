// Package tray provides the system tray (status notifier) icon: a mail icon
// that grows a notification dot while there are unread messages.
package tray

// Actions are the callbacks behind the tray icon and its menu.
type Actions struct {
	Show    func() // left click and "Open Aerion"
	Compose func() // "Compose"
	Quit    func() // "Quit"
}
