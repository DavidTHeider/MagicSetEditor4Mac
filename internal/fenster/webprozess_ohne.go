//go:build !darwin

package fenster

// **Ohne macOS kein Webprozess von WebKit** — gebaut wird nur für darwin
// (§12); diese Fassung hält die Übersetzung für die anderen Ziele grün.

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func webprozessAnschliessen(context.Context, *application.WebviewWindow, func(context.Context)) {}
