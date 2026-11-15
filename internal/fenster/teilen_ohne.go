//go:build !darwin

package fenster

// **Ohne macOS kein Teilen-Dialog** — gebaut wird nur für darwin (§12); diese
// Fassung hält die Übersetzung für die anderen Ziele grün.

import (
	"errors"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func systemTeilen(application.Window, string, string) error {
	return errors.New("Teilen gibt es nur unter macOS")
}
