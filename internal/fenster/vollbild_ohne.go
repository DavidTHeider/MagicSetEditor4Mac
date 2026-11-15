//go:build !darwin

package fenster

// **Ohne macOS kein Vollbild-Eintrag zum Nachführen** — gebaut wird nur für
// darwin (§12); diese Fassung hält die Übersetzung für die anderen Ziele grün.

import "github.com/wailsapp/wails/v3/pkg/application"

func vollbildAnschliessen(*application.WebviewWindow) {}
