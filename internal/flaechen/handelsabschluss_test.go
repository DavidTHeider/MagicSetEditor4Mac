package flaechen_test

// **Der Abschluss eines Handels auf jeder Maschine** — #2285
// (#2285).
//
// Er reist mit Absicht nicht über den Draht (`docs/ABNAHME-T12.md`, „Wo der
// Draht aufhört"): jeder der drei Schritte verlangt einen Nachweis, den nur
// das Fenster ausstellt. **Hier für beide Läufe mit zwei Prozessen** — über
// TCP (`handelsabnahme_test.go`) und über fremde Netze
// (`fernhandel_netzlauf_test.go`).

import (
	"encoding/json"
	"testing"
)

// schliesseAb fährt die drei Abschlussschritte auf einer Maschine:
// verbuchen, abgleichen und beide Stimmen bestätigen — jede mit ihrem
// eigenen Nachweis (`eineFreigabeFuer`, ausgestellt wie am Fenster).
func schliesseAb(t *testing.T, wer einRechner, parteien ...string) {
	t.Helper()

	mussKommandozeile(t, wer.umgebung, "handel", "verbuchen", derAbnahmevorgang,
		eineFreigabeFuer(t, wer.umgebung, derAbnahmevorgang, "verbuchen", ""))
	warteAufZustand(t, wer, "lokalVerbucht")

	mussKommandozeile(t, wer.umgebung, "handel", "abgleichen", derAbnahmevorgang)
	warteAufZustand(t, wer, "abgleichOffen")

	for _, partei := range parteien {
		mussKommandozeile(t, wer.umgebung, "handel", "bestaetigen", derAbnahmevorgang,
			partei, eineFreigabeFuer(t, wer.umgebung, derAbnahmevorgang, "bestaetigen", partei))
	}

	warteAufZustand(t, wer, "beidseitigVerbucht")
}

// mengenJeDruck liest die Sammlung einer Maschine als Menge je Druck.
func mengenJeDruck(t *testing.T, wer einRechner) map[string]int {
	t.Helper()

	var sammlung struct {
		Exemplare []struct {
			Printing string `json:"printing"`
			Menge    int    `json:"menge"`
		} `json:"exemplare"`
	}

	roh := mussKommandozeile(t, wer.umgebung, "--json", "sammlung")
	if fehler := json.Unmarshal([]byte(roh), &sammlung); fehler != nil {
		t.Fatalf("die Sammlung lesen: %v (%s)", fehler, roh)
	}

	mengen := map[string]int{}
	for _, exemplar := range sammlung.Exemplare {
		mengen[exemplar.Printing] += exemplar.Menge
	}

	return mengen
}
