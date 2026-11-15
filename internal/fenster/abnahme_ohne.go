//go:build !abnahme || !darwin

package fenster

import "github.com/wailsapp/wails/v3/pkg/application"

// abnahmeAnschliessen tut außerhalb des Abnahmebaus nichts — die Leitung gibt
// es dort nicht, auch nicht verschlossen (`abnahme.go`).
func abnahmeAnschliessen(*application.App, *application.WebviewWindow) {}

// abnahmeDialog hat außerhalb des Abnahmebaus nie eine Antwort bereit — der
// Dialog von macOS geht auf.
func abnahmeDialog() (string, bool) { return "", false }

// abnahmeAblage gibt außerhalb des Abnahmebaus keine Ersatzablage — kopiert
// wird in die Zwischenablage des Rechners.
func abnahmeAblage() *ersatzablage { return nil }

// AbnahmeVorbereiten tut außerhalb des Abnahmebaus nichts — es gibt dort
// weder eine Leitung noch Wails' MCP-Server.
func AbnahmeVorbereiten() error { return nil }

// rufbuchHuelle lässt außerhalb des Abnahmebaus alles, wie es ist — das
// Rufbuch gibt es nur dort (`rufbuch.go`).
func rufbuchHuelle(innen application.Middleware) application.Middleware { return innen }
