//go:build darwin

package fenster

// **Ein Webprozess, der endet, wird gezählt** — OPS02
// (#2348).
//
// WebKit beendet den Prozess einer Webansicht, wenn er abstürzt oder macOS
// ihn bei Speicherdruck wegnimmt — auffällig nach dem Ruhezustand. **Wails
// lädt die Seite danach selbst neu** (`webViewWebContentProcessDidTerminate:`
// in `webview_window_darwin.m`, v3.0.0-beta.27) und meldet es als
// `events.Mac.WebViewWebContentProcessDidTerminate`; bis hierhin hörte
// niemand zu, und ein offenes Formular verschwand ohne Satz.
//
// **Ohne Lauf, und das steht in `.covignore`**: es braucht ein echtes
// Fenster und einen echten Webprozess. Belegt mit der Hüllenprobe (Drehbuch
// `ops02`), die den Webprozess beendet.

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// webprozessAnschliessen zählt jedes Ende des Webprozesses.
func webprozessAnschliessen(ctx context.Context, fenster *application.WebviewWindow, gezaehlt func(context.Context)) {
	fenster.OnWindowEvent(events.Mac.WebViewWebContentProcessDidTerminate, gesichert("das Ende des Webprozesses", func(*application.WindowEvent) {
		gezaehlt(ctx)
	}))
}
