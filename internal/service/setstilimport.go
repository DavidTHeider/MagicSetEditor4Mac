package service

// **Der `styling:`-Block eines Sets** — die setweiten Stiloptionen
// (#557).
//
// Eigene Datei, weil `import.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was** aus einer Paketdatei in den
// Bestand kommt, hier, wie ein einzelner Block dafür flach gemacht wird.

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// setstilwertAus macht aus dem `styling:`-Block **einen** Setwert.
//
// **Flach, eine Zeile je Option**, `<paket>/<name>: <wert>` — derselbe Weg wie
// bei der Karte, deren `styling_data:` als gewöhnlicher Feldwert liegt
// ([Setstilfeld]). Eine eigene Tabelle wäre eine zweite Antwort auf dieselbe
// Frage.
//
// **Der Rundlauf hängt nicht daran**: geschrieben wird aus der mitgeführten
// Quelldatei ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)),
// nicht aus den Rohblöcken.
func setstilwertAus(knoten *msepaket.Knoten) domain.Value {
	var gebaut strings.Builder

	for _, paket := range knoten.Kinder {
		if paket.Zeilenart != msepaket.ArtEintrag {
			continue
		}

		for _, option := range paket.Kinder {
			if option.Zeilenart != msepaket.ArtEintrag {
				continue
			}

			gebaut.WriteString(paket.Schluessel)
			gebaut.WriteString("/")
			gebaut.WriteString(option.Schluessel)
			gebaut.WriteString(": ")
			gebaut.WriteString(option.Wert)
			gebaut.WriteString("\n")
		}
	}

	return domain.Value{Field: Setstilfeld, Text: gebaut.String()}
}
