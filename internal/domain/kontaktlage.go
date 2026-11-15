package domain

// **Wie weit ich einem Gegenüber traue — in den Worten des Issues** (NET-01,
// #1705).
//
// # Warum eine Lage neben dem Stand
//
// **Weil der Stand eine Speicherfrage beantwortet und die Lage eine
// Anzeigefrage.** Gespeichert sind drei Stände; NET-01.AK-03 verlangt, dass
// „unbekannt, bestätigt, widerrufen und Schlüssel geändert" **unterschiedliche
// Zustände** sind. „Unbekannt" hat keine Zeile, „Schlüssel geändert" ist ein
// unbestätigter Kontakt mit einem Vorgänger — beides folgt aus dem, was da
// ist, und ein vierter gespeicherter Stand müsste mit dem Vorgänger
// übereinstimmen, statt aus ihm zu folgen.

// Kontaktlage ist der Zustand eines Gegenübers, wie ein Mensch ihn sieht.
type Kontaktlage string

// Die fünf Lagen.
const (
	// LageUnbekannt heißt: es gibt keinen Kontakt.
	LageUnbekannt Kontaktlage = "unbekannt"

	// LageUnbestaetigt heißt: aufgenommen, nicht verglichen.
	LageUnbestaetigt Kontaktlage = "unbestaetigt"

	// LageSchluesselGeaendert heißt: jemand, den ich kannte, kommt mit einem
	// neuen Schlüssel — und **nichts** von vorher gilt für ihn.
	LageSchluesselGeaendert Kontaktlage = "schluesselGeaendert"

	// LageBestaetigt heißt: ein Mensch hat verglichen.
	LageBestaetigt Kontaktlage = "bestaetigt"

	// LageWiderrufen heißt: gesperrt.
	LageWiderrufen Kontaktlage = "widerrufen"
)

// DieKontaktlagen ist die Aufzählung in fester Reihenfolge.
var DieKontaktlagen = []Kontaktlage{
	LageUnbekannt, LageUnbestaetigt, LageSchluesselGeaendert, LageBestaetigt,
	LageWiderrufen,
}

// Lage gibt die Lage eines vorhandenen Kontakts.
func (k Peerkontakt) Lage() Kontaktlage {
	switch {
	case k.Stand == KontaktGesperrt:
		return LageWiderrufen
	case k.Stand == KontaktBestaetigt:
		return LageBestaetigt
	case k.Vorgaenger != "":
		return LageSchluesselGeaendert
	default:
		return LageUnbestaetigt
	}
}
