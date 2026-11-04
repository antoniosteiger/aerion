package platform

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/hkdb/aerion/internal/logging"
)

// PortalFilter is a named set of glob patterns for the Open dialog.
type PortalFilter struct {
	Name     string
	Patterns []string
}

// Wire types for the portal's filters option: a(sa(us)).
type portalGlob struct {
	Kind    uint32 // 0 = glob pattern
	Pattern string
}

type portalFilterWire struct {
	Name  string
	Globs []portalGlob
}

// portalFileChooser calls an org.freedesktop.portal.FileChooser method and
// waits for the Response signal. Returns the chosen paths, or (nil, nil) if
// the user cancelled.
func portalFileChooser(method, title string, options map[string]dbus.Variant) ([]string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	// Don't close conn — it's the shared session bus used by GTK/WebKit

	// Check if the portal service is actually running
	var hasOwner bool
	if err := conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, "org.freedesktop.portal.Desktop").Store(&hasOwner); err != nil || !hasOwner {
		return nil, fmt.Errorf("portal service not available")
	}

	handleToken := fmt.Sprintf("aerion_%d", time.Now().UnixNano())
	options["handle_token"] = dbus.MakeVariant(handleToken)

	// Compute expected request path from our unique bus name and handle token
	sender := conn.Names()[0]
	senderPath := strings.ReplaceAll(sender[1:], ".", "_")
	requestPath := dbus.ObjectPath(fmt.Sprintf(
		"/org/freedesktop/portal/desktop/request/%s/%s", senderPath, handleToken,
	))

	// Subscribe to Response signal BEFORE calling the method to avoid races
	matchRule := fmt.Sprintf(
		"type='signal',interface='org.freedesktop.portal.Request',member='Response',path='%s'",
		requestPath,
	)
	if err := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule).Err; err != nil {
		return nil, fmt.Errorf("failed to subscribe to portal response: %w", err)
	}
	defer conn.BusObject().Call("org.freedesktop.DBus.RemoveMatch", 0, matchRule)

	signals := make(chan *dbus.Signal, 1)
	conn.Signal(signals)
	defer conn.RemoveSignal(signals)

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	if call := obj.Call("org.freedesktop.portal.FileChooser."+method, 0, "", title, options); call.Err != nil {
		return nil, fmt.Errorf("%s portal call failed: %w", method, call.Err)
	}

	// The user may take a while in the dialog, so the timeout is generous.
	timeout := time.After(30 * time.Minute)
	for {
		select {
		case signal := <-signals:
			if signal == nil || signal.Path != requestPath || signal.Name != "org.freedesktop.portal.Request.Response" {
				continue
			}
			if len(signal.Body) < 2 {
				return nil, fmt.Errorf("incomplete response from FileChooser portal")
			}
			response, ok := signal.Body[0].(uint32)
			if !ok {
				return nil, fmt.Errorf("unexpected response type from FileChooser portal")
			}
			// 1 = user cancelled
			if response == 1 {
				return nil, nil
			}
			if response != 0 {
				return nil, fmt.Errorf("%s dialog failed (response: %d)", method, response)
			}

			results, ok := signal.Body[1].(map[string]dbus.Variant)
			if !ok {
				return nil, fmt.Errorf("unexpected results type from FileChooser portal")
			}
			uris, ok := results["uris"].Value().([]string)
			if !ok || len(uris) == 0 {
				return nil, fmt.Errorf("no URIs in FileChooser portal response")
			}

			paths := make([]string, len(uris))
			for i, uri := range uris {
				p, err := uriToPath(uri)
				if err != nil {
					return nil, err
				}
				paths[i] = p
			}
			return paths, nil

		case <-timeout:
			return nil, fmt.Errorf("FileChooser portal request timed out")
		}
	}
}

// currentFolder encodes a directory as the portal's null-terminated "ay" type.
func currentFolder(directory string) dbus.Variant {
	return dbus.MakeVariant(append([]byte(directory), 0))
}

// PortalOpenFiles shows the XDG FileChooser portal Open dialog. With
// selectDirectory it picks folders instead of files. Returns the chosen
// paths, or (nil, nil) if the user cancelled.
func PortalOpenFiles(title, directory string, multiple, selectDirectory bool, filters []PortalFilter) ([]string, error) {
	options := map[string]dbus.Variant{
		"multiple":  dbus.MakeVariant(multiple),
		"directory": dbus.MakeVariant(selectDirectory),
	}
	if directory != "" {
		options["current_folder"] = currentFolder(directory)
	}
	if len(filters) > 0 {
		wire := make([]portalFilterWire, len(filters))
		for i, f := range filters {
			wire[i].Name = f.Name
			for _, pat := range f.Patterns {
				wire[i].Globs = append(wire[i].Globs, portalGlob{Pattern: pat})
			}
		}
		options["filters"] = dbus.MakeVariant(wire)
	}
	return portalFileChooser("OpenFile", title, options)
}

// PortalSaveFile shows the XDG FileChooser portal Save dialog for a single file.
// Returns the chosen path, or ("", nil) if the user cancelled.
func PortalSaveFile(title, suggestedName, directory string) (string, error) {
	options := map[string]dbus.Variant{"current_name": dbus.MakeVariant(suggestedName)}
	if directory != "" {
		options["current_folder"] = currentFolder(directory)
	}
	paths, err := portalFileChooser("SaveFile", title, options)
	if err != nil || len(paths) == 0 {
		return "", err
	}
	return paths[0], nil
}

// PortalSaveFiles shows the XDG FileChooser portal SaveFiles dialog for multiple files.
// Returns the list of chosen paths, or (nil, nil) if the user cancelled.
func PortalSaveFiles(title string, filenames []string, directory string) ([]string, error) {
	// files is aay (array of null-terminated byte arrays)
	files := make([][]byte, len(filenames))
	for i, name := range filenames {
		files[i] = append([]byte(name), 0)
	}
	options := map[string]dbus.Variant{"files": dbus.MakeVariant(files)}
	if directory != "" {
		options["current_folder"] = currentFolder(directory)
	}
	return portalFileChooser("SaveFiles", title, options)
}

// PortalOpenDirectory opens the folder containing a file using the OpenURI portal.
// This resolves sandboxed document portal paths to the real host location.
func PortalOpenDirectory(filePath string) error {
	log := logging.WithComponent("filechooser")

	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	// Check if the portal service is actually running
	var hasOwner bool
	if err := conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, "org.freedesktop.portal.Desktop").Store(&hasOwner); err != nil || !hasOwner {
		return fmt.Errorf("portal service not available")
	}

	// Open the file to get a file descriptor — the portal uses the fd
	// to resolve the real host path even from inside the sandbox
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer f.Close()

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	options := map[string]dbus.Variant{}

	call := obj.Call("org.freedesktop.portal.OpenURI.OpenDirectory", 0, "", dbus.UnixFD(f.Fd()), options)
	if call.Err != nil {
		return fmt.Errorf("OpenDirectory portal call failed: %w", call.Err)
	}

	log.Debug().Str("path", filePath).Msg("Opened directory via portal")
	return nil
}

// IsDocPortalPath checks if a path is on the document portal FUSE mount.
func IsDocPortalPath(path string) bool {
	return strings.HasPrefix(path, "/run/user/") && strings.Contains(path, "/doc/")
}

// PortalOpenFile opens a file with the default application using the OpenURI portal.
func PortalOpenFile(filePath string) error {
	log := logging.WithComponent("filechooser")

	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	var hasOwner bool
	if err := conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, "org.freedesktop.portal.Desktop").Store(&hasOwner); err != nil || !hasOwner {
		return fmt.Errorf("portal service not available")
	}

	f, err := os.OpenFile(filePath, os.O_RDWR, 0)
	if err != nil {
		f, err = os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %q: %w", filePath, err)
		}
	}
	defer f.Close()

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	options := map[string]dbus.Variant{
		"writable": dbus.MakeVariant(true),
	}
	call := obj.Call("org.freedesktop.portal.OpenURI.OpenFile", 0, "", dbus.UnixFD(f.Fd()), options)
	if call.Err != nil {
		return fmt.Errorf("OpenFile portal call failed: %w", call.Err)
	}

	log.Debug().Str("path", filePath).Msg("Opened file via portal")
	return nil
}

// PortalOpenURI opens a URI with the user's default handler via the XDG
// OpenURI portal. Works inside the Flatpak sandbox (unlike xdg-open) and
// triggers the host's URL-handler notification on Wayland DEs.
// Callers should fall back to wailsRuntime.BrowserOpenURL on error.
func PortalOpenURI(uri string) error {
	log := logging.WithComponent("filechooser")

	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	var hasOwner bool
	if err := conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, "org.freedesktop.portal.Desktop").Store(&hasOwner); err != nil || !hasOwner {
		return fmt.Errorf("portal service not available")
	}

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	options := map[string]dbus.Variant{}
	call := obj.Call("org.freedesktop.portal.OpenURI.OpenURI", 0, "", uri, options)
	if call.Err != nil {
		return fmt.Errorf("OpenURI portal call failed: %w", call.Err)
	}

	log.Debug().Str("uri", uri).Msg("Opened URI via portal")
	return nil
}

// uriToPath converts a file:// URI to a filesystem path.
func uriToPath(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse URI %q: %w", uri, err)
	}
	if parsed.Scheme != "file" {
		return "", fmt.Errorf("unexpected URI scheme %q (expected file://)", parsed.Scheme)
	}
	return parsed.Path, nil
}
