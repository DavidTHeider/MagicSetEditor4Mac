package domain

// **Wogegen ein Anmeldeanbieter geprüft ist** — ID-05.AK-01
// (#1704),
// Vermerk [0136](../../docs/entscheidungen/0136-anbieter-maschinen-und-zustimmung.md).
//
// **Neben „eingerichtet", nicht statt seiner**: ob eine `client_id` im
// Bestand steht, weiß die Anwendung zur Laufzeit; **wogegen der Weg geprüft
// ist**, steht hier und ändert sich nur mit einem Lauf. Für die Datendienste
// sagt es der Dienstkatalog je Fähigkeit mit Datum und Fassung (0082).

import "slices"

// Pruefstand sagt, wogegen ein Weg geprüft ist.
type Pruefstand string

// Die Prüfstände.
const (
	// PruefstandSimuliert heißt: gegen eine Nachbildung des Anbieters — ein
	// eigener Server im Test, der den OAuth-Lauf nach der Spezifikation
	// beantwortet. **Für jeden Anbieter ein durchlaufener Weg**:
	// `TestJederAnbieterLaeuftGegenEineNachbildung` in `internal/anmeldung`
	// geht über [Anbieter] und wird rot, wenn einem die Nachbildung fehlt.
	PruefstandSimuliert Pruefstand = "simuliert"

	// PruefstandReal heißt: gegen den echten Anbieter, mit Datum im Beleg.
	PruefstandReal Pruefstand = "realGeprueft"
)

// anbieterlaeufe sind die Anbieter, gegen die ein echter Lauf belegt ist.
//
// **Leer, und das ist die Auskunft**: ein Lauf gegen GitHub, Google,
// Microsoft oder X braucht ein eigenes Klientenkonto, und dieses Projekt
// liefert keines mit — jede Kennung trägt, wer sich anmelden will, selbst
// ein. Wer einen Lauf macht, trägt ihn hier mit seinem Beleg ein.
var anbieterlaeufe = []string{}

// AnbieterPruefstand sagt es für einen Anmeldeanbieter — leer für einen, den
// es nicht gibt.
func AnbieterPruefstand(anbieter string) Pruefstand {
	switch {
	case !slices.Contains(Anbieter(), anbieter):
		return ""
	case slices.Contains(anbieterlaeufe, anbieter):
		return PruefstandReal
	default:
		return PruefstandSimuliert
	}
}
