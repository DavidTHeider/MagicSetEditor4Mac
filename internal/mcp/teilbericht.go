// **Ein Fehler nennt, was schon geschrieben ist** — Befund 3 des Zweitblicks
// über Phase 4.5
// (#1292).
//
// # Warum es das braucht
//
// **MCP trägt entweder einen Text oder einen Fehler, nicht beides.** Ein
// Werkzeug, das bei Zeile 217 von 300 abbricht, gäbe damit nur den Fehler —
// und der Rufende wüsste nicht, dass 216 Zeilen stehen. Sein zweiter Versuch
// schriebe sie noch einmal.
//
// §7, Regel 3 sagt es andersherum: **geschrieben ist geschrieben.** Was
// geschrieben wurde, gehört in jede Antwort — auch in eine, die scheitert.

package mcp

import "fmt"

// mitBericht hängt an einen Fehler, was vor dem Abbruch geschehen ist.
//
// **Nichts anhängen, wenn nichts geschehen ist.** Ein „(0 angelegt, 0
// geändert)" hinter jedem Fehler wäre Rauschen — und der Satz soll gelesen
// werden, wenn er dasteht.
func mitBericht(fehler error, angelegt []int64, geaendert int) error {
	if len(angelegt) == 0 && geaendert == 0 {
		return fehler
	}

	return fmt.Errorf("%w — vor dem Abbruch geschrieben: %d angelegt, %d geändert",
		fehler, len(angelegt), geaendert)
}
