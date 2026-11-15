//go:build server

package fenster

import (
	"io/fs"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// schutzbau lässt die Assetleitung im Servermodus, wie sie ist — **die
// Schutzrichtlinie gilt der WebView** (SEC-17, #2164). Hier liefert ein Port
// dieselbe Seite an einen Browser; wer überhaupt fragen darf, entscheidet die
// Wache in [assetweiche].
func schutzbau(innen application.Middleware, _ fs.FS, _ *slog.Logger) application.Middleware {
	return innen
}
