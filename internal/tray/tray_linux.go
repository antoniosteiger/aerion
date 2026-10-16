//go:build linux

package tray

import (
	_ "embed"
	"sync"

	"fyne.io/systray"

	"github.com/hkdb/aerion/internal/logging"
)

var (
	//go:embed icons/mail.png
	iconRead []byte
	//go:embed icons/mail-unread.png
	iconUnread []byte
)

var (
	mu      sync.Mutex
	started bool
	unread  bool
	stop    func()
)

// Start registers the tray icon on the session bus (StatusNotifierItem).
// Safe to call once; panels without a StatusNotifierWatcher just don't show it.
func Start(actions Actions) {
	mu.Lock()
	defer mu.Unlock()
	if started {
		return
	}
	started = true

	start, end := systray.RunWithExternalLoop(func() {
		systray.SetTitle("Aerion")
		systray.SetTooltip("Aerion")
		systray.SetOnTapped(actions.Show)

		show := systray.AddMenuItem("Open Aerion", "")
		compose := systray.AddMenuItem("Compose", "")
		systray.AddSeparator()
		quit := systray.AddMenuItem("Quit", "")

		go func() {
			for {
				select {
				case <-show.ClickedCh:
					actions.Show()
				case <-compose.ClickedCh:
					actions.Compose()
				case <-quit.ClickedCh:
					actions.Quit()
				}
			}
		}()
	}, nil)
	start()
	stop = end
	setIcon(unread)

	log := logging.WithComponent("tray")
	log.Info().Msg("Tray icon registered")
}

// SetUnread switches between the plain and the notification-dot icon.
func SetUnread(hasUnread bool) {
	mu.Lock()
	defer mu.Unlock()
	if hasUnread == unread {
		return
	}
	unread = hasUnread
	if started {
		setIcon(unread)
	}
}

// Stop removes the tray icon.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if stop != nil {
		stop()
		stop = nil
	}
}

func setIcon(hasUnread bool) {
	if hasUnread {
		systray.SetIcon(iconUnread)
		return
	}
	systray.SetIcon(iconRead)
}
