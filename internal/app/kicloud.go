package app

// **Was über die Auslagerung der lokalen Laufzeit feststeht** — Werkstück A7
// von Phase 5.3
// (#1446),
// ausgelagert für A8
// (#1447).
//
// **Eine eigene Datei**, weil `kilaufzeit.go` mit der Anbieterliste über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, wie die
// **lokale** Laufzeit gefragt wird, hier, was sie über ihre eigene Cloud sagt.

import (
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// MitKIZuhause setzt das Zuhause, in dem die Serverdatei liegt.
//
// **Damit ein Lauf nicht den Stand dieses Rechners liest** (§3): ohne das
// Feld läse jeder Test `~/.ollama/server.json` der Maschine, auf der er
// läuft — und wäre auf der nächsten ein anderer.
func (a *App) MitKIZuhause(zuhause string) *App {
	a.kizuhause = zuhause

	return a
}

// cloudstand liest, was über die Auslagerung feststeht.
//
// **Ohne Fehlerrückgabe**: jeder Ausgang ist eine Auskunft, und „ungewiss"
// ist eine davon ([ollama.LiesCloudstand]).
//
// **Und er gilt allein der lokalen Laufzeit.** Anthropic hat keine
// Auslagerung, die man abschalten könnte — seine Modelle rechnen dort, wo sie
// stehen, und das sagt ihr Ort.
func (a *App) cloudstand() KICloudDTO {
	zuhause := a.kizuhause

	if zuhause == "" {
		eigenes, fehler := os.UserHomeDir()
		if fehler != nil {
			return KICloudDTO{
				Stand: string(ollama.CloudUngewiss),
				Beleg: "Das Zuhause dieses Nutzers ist nicht zu finden: " +
					fehler.Error(),
			}
		}

		zuhause = eigenes
	}

	befund := ollama.LiesCloudstand(zuhause)

	return KICloudDTO{Stand: string(befund.Stand), Beleg: befund.Beleg}
}
