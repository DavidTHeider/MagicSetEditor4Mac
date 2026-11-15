//go:build server

package fenster

// **Die Weiche vor Wails' Assetleitung im Servermodus** — QA-09
// (#1740),
// Befund P1 in `docs/PENTEST.md`.

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// assetweiche ist, was vor Wails' Assetleitung steht — **die Wache vor der
// Bildweiche**: im Servermodus lauscht ein Port, und Wails' HTTP-Weg ruft
// jede Bindung, gleich woher die Anfrage kommt. Warum und wogegen, steht in
// `internal/vorschauweg/rechnerwache.go`.
func assetweiche(bildweg *vorschauweg.Weg) application.Middleware {
	return vorschauweg.NurVonDiesemRechner(bildweg.Vor)
}
