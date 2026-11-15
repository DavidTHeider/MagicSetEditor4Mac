// Package laufbeleg belegt, dass eine Messung stattgefunden hat — QA-01
// (#1733).
//
// # Wofür
//
// **Ein grüner Lauf und ein Lauf, der nichts gemessen hat, sahen gleich aus.**
// Vom 2026-09-14 bis zum 2026-09-20 übersprangen drei Kanonläufe, weil ihr
// Verzeichnis im eigenen Zuhause der Läufe nicht lag — und `make fremdpakete`
// meldete grün
// (#1656).
//
// Dieses Paket liest den Strom von `go test -json` und sagt je erwartetem
// Fall **zwei** Dinge getrennt (AK-03): ob die Messung stattfand — die
// **Ausführung** — und was sie ergab — das **Ergebnis**. Ein bewusst roter
// Trendlauf liefert damit Messdaten, und ein abgebrochener Prozess gilt
// trotzdem nicht als gelaufen.
//
// # Fünf Stände, und keiner steht für einen anderen
//
// PASS, FAIL, BLOCKED, SKIPPED und NOT_RUN (AK-02). Ein Sprung ist kein
// Bestehen, ein Treffer aus dem Zwischenspeicher keine Messung, und ein
// Absturz kein roter Fall, sondern ein gescheiterter Lauf.
//
// # Was es nicht weiß
//
// **Ob eine Messung das Richtige misst.** Es belegt, dass sie lief, an welchem
// Stand und mit welchen Eingaben — nicht, dass ihr Maß stimmt. Dafür steht
// `docs/MESSMETHODIK.md`.
//
// **Und es ist ein Werkzeug und kein Teil der Anwendung.** Es kennt nichts aus
// diesem Projekt, und kein ausgeliefertes Programm kennt es.
package laufbeleg

// Status ist einer der fünf Stände aus QA-01.AK-02 — die Werte stehen so im
// Issue und bleiben deshalb, wie sie dort heißen.
type Status string

// Die fünf Stände.
const (
	Bestanden     Status = "PASS"
	Gescheitert   Status = "FAIL"
	Blockiert     Status = "BLOCKED"
	Uebersprungen Status = "SKIPPED"
	NichtGelaufen Status = "NOT_RUN"
)

// rang ordnet die Stände nach Schwere — der schwerste bestimmt das Ganze.
//
// **FAIL vor BLOCKED**: ein Lauf, der scheitert und dem zudem etwas fehlt, ist
// zuerst gescheitert. **NOT_RUN vor SKIPPED**: ein Fall, von dem gar nichts
// kam, sagt weniger als einer, der sich mit Grund abgemeldet hat.
func (s Status) rang() int {
	switch s {
	case Gescheitert:
		return 4
	case Blockiert:
		return 3
	case NichtGelaufen:
		return 2
	case Uebersprungen:
		return 1
	default:
		return 0
	}
}
