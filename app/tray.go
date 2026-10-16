package app

import (
	"time"

	"github.com/hkdb/aerion/internal/logging"
	"github.com/hkdb/aerion/internal/tray"
)

// trayPollInterval is how often the tray re-reads the unread count.
const trayPollInterval = 3 * time.Second

// startTray shows the tray icon and keeps its notification dot in sync with
// the unified inbox. It polls the cached folder counts (one cheap SQL sum),
// which every sync and flag change already maintain, instead of hooking each
// of the places that change them.
func (a *App) startTray() {
	tray.Start(tray.Actions{
		Show:    a.ShowWindow,
		Compose: func() { a.handleExternalCompose(ComposeRequest{}) },
		Quit:    a.QuitApp,
	})

	a.trayDone = make(chan struct{})
	go func() {
		defer recoverPanic("app.tray", "unread poll")
		log := logging.WithComponent("app.tray")
		ticker := time.NewTicker(trayPollInterval)
		defer ticker.Stop()
		for {
			count, err := a.folderStore.InboxUnreadTotal()
			if err != nil {
				log.Debug().Err(err).Msg("Failed to read unread count")
			} else {
				tray.SetUnread(count > 0)
			}
			select {
			case <-a.trayDone:
				return
			case <-ticker.C:
			}
		}
	}()
}

// stopTray stops the unread poll and removes the tray icon.
func (a *App) stopTray() {
	if a.trayDone != nil {
		close(a.trayDone)
		a.trayDone = nil
	}
	tray.Stop()
}
