package app

// **Der freie Wert einer Panik bleibt am Rechner** — #2399, Vermerk 0193.
//
// Im Protokoll dieses Rechners steht er ganz: der Verwalter braucht ihn, um
// einen Fehler zu finden (F02, F03), und die Hüllenproben `panik09` und
// `fehlerform09` lesen ihn dort. **Was hinausgehen kann** — die Diagnose,
// die jemand kopiert, und das Diagnosepaket — trägt ihn nur in der Gestalt,
// die das Laufzeitsystem schreibt ([wertFuerDasPaket]); dieselbe Zusage wie
// für die Absturzspur. Bis dahin schwärzte das Protokoll nach Gestalt: ein
// langes Token fiel, ein kurzes Kennwort aus Buchstaben nicht.

import "regexp"

var (
	// panikmerkmal ist das Merkmal, unter dem eine gefangene Panik steht —
	// mit Anführung, wie slog einen Wert mit Leerzeichen schreibt, oder ohne;
	// **und im Format JSON** (`"panik":"…"`, Zweitblick über Phase 5,
	// Befund 1), das die Weiche schreibt, wenn `protokoll.format` es sagt.
	panikmerkmal = regexp.MustCompile(`(\bpanik=|"panik":)(?:"((?:[^"\\]|\\.)*)"|([^\s",}]+))`)

	// paniksatz ist der Wert hinter den Sätzen, mit denen eine Panik in
	// einem anderen Merkmal reist: Wails' `RuntimeError` und Go
	// (`panic: …`), ein Zusammenbruch (`domain.ErrZusammengebrochen`) und
	// eine Panik am Fenster (`panikFehler`) — bis zum Ende des Werts.
	paniksatz = regexp.MustCompile(`((?:panic|zusammengebrochen|eine Panik): )((?:[^"\\\n]|\\.)*)`)
)

// paniktexteSchwaerzen nimmt aus Protokollzeilen den freien Wert gefangener
// Paniken.
func paniktexteSchwaerzen(text string) string {
	text = panikmerkmal.ReplaceAllStringFunc(text, func(treffer string) string {
		teile := panikmerkmal.FindStringSubmatch(treffer)
		if teile[3] != "" {
			return teile[1] + wertFuerDasPaket(teile[3])
		}

		return teile[1] + `"` + wertFuerDasPaket(teile[2]) + `"`
	})

	return paniksatz.ReplaceAllStringFunc(text, func(treffer string) string {
		teile := paniksatz.FindStringSubmatch(treffer)

		return teile[1] + wertFuerDasPaket(teile[2])
	})
}
