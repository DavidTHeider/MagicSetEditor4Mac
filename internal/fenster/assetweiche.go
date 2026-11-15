//go:build !server

package fenster

// **Die Weiche vor Wails' Assetleitung im Fensterbau** — QA-09
// (#1740).
//
// **Ohne Wache, und das ist eine Aussage über den Bau**: das Fenster spricht
// über das eigene Schema `wails://localhost` mit seiner Assetleitung, und
// dafür lauscht kein Port. Eine Seite im Browser erreicht diesen Weg nicht.
// Den Lauscher gibt es nur im Servermodus (`assetweiche_server.go`) — und
// dort steht die Wache davor.

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// assetweiche ist, was vor Wails' Assetleitung steht.
func assetweiche(bildweg *vorschauweg.Weg) application.Middleware {
	return bildweg.Vor
}
