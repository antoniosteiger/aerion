package app

import (
	"context"
	"runtime"
	"strings"

	"github.com/hkdb/aerion/internal/logging"
	"github.com/hkdb/aerion/internal/platform"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// File dialogs. Wails opens a plain GTK dialog on Linux, which bypasses the
// XDG FileChooser portal (and so any portal backend such as
// xdg-desktop-portal-termfilechooser). On Linux we ask the portal first and
// fall back to the Wails dialog if the portal is unavailable.

// viaPortal runs fn and reports whether the portal handled the request.
// A cancelled dialog counts as handled (nil paths, ok = true).
func viaPortal(fn func() ([]string, error)) ([]string, bool) {
	if runtime.GOOS != "linux" {
		return nil, false
	}
	paths, err := fn()
	if err != nil {
		log := logging.WithComponent("filedialog")
		log.Warn().Err(err).Msg("Portal file dialog failed, falling back to GTK dialog")
		return nil, false
	}
	return paths, true
}

func portalFilters(filters []wailsRuntime.FileFilter) []platform.PortalFilter {
	out := make([]platform.PortalFilter, len(filters))
	for i, f := range filters {
		out[i] = platform.PortalFilter{Name: f.DisplayName, Patterns: strings.Split(f.Pattern, ";")}
	}
	return out
}

func openFileDialog(ctx context.Context, opts wailsRuntime.OpenDialogOptions) (string, error) {
	paths, ok := viaPortal(func() ([]string, error) {
		return platform.PortalOpenFiles(opts.Title, opts.DefaultDirectory, false, false, portalFilters(opts.Filters))
	})
	if !ok {
		return wailsRuntime.OpenFileDialog(ctx, opts)
	}
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], nil
}

func openMultipleFilesDialog(ctx context.Context, opts wailsRuntime.OpenDialogOptions) ([]string, error) {
	paths, ok := viaPortal(func() ([]string, error) {
		return platform.PortalOpenFiles(opts.Title, opts.DefaultDirectory, true, false, portalFilters(opts.Filters))
	})
	if !ok {
		return wailsRuntime.OpenMultipleFilesDialog(ctx, opts)
	}
	return paths, nil
}

func openDirectoryDialog(ctx context.Context, opts wailsRuntime.OpenDialogOptions) (string, error) {
	paths, ok := viaPortal(func() ([]string, error) {
		return platform.PortalOpenFiles(opts.Title, opts.DefaultDirectory, false, true, nil)
	})
	if !ok {
		return wailsRuntime.OpenDirectoryDialog(ctx, opts)
	}
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], nil
}

func saveFileDialog(ctx context.Context, opts wailsRuntime.SaveDialogOptions) (string, error) {
	paths, ok := viaPortal(func() ([]string, error) {
		path, err := platform.PortalSaveFile(opts.Title, opts.DefaultFilename, opts.DefaultDirectory)
		if err != nil || path == "" {
			return nil, err
		}
		return []string{path}, nil
	})
	if !ok {
		return wailsRuntime.SaveFileDialog(ctx, opts)
	}
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], nil
}
