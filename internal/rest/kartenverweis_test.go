// **Die Abweisungen der Verweiswege** — Werkstück 2 von Phase 3.6
// (#1167).
//
// **Neben den vier Klassen und nicht statt ihnen.** Jene prüfen, dass der Weg
// bis zum Anwendungsfall trägt; hier stehen die Fälle, die nur diese Wege
// haben.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Jeder der vier Wege weist eine unbrauchbare Kennung ab.**
func TestDieVerweiswegeWeisenKennungenAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Verweis lesen", methode: http.MethodGet, weg: "/karten/x/verweis"},
		{
			name: "Verweis setzen", methode: http.MethodPut, weg: "/karten/x/verweis",
			rumpf: `{"printing":"abc"}`,
		},
		{name: "Verweis lösen", methode: http.MethodDelete, weg: "/karten/x/verweis"},
		{name: "Setverweise lesen", methode: http.MethodGet, weg: "/sets/x/verweise"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Ein unlesbarer Rumpf wird abgewiesen** — und ein leerer Druck auch.
//
// **Zwei Fälle und nicht einer**: `{` scheitert am JSON, `{}` an der Regel.
// Wer nur den ersten prüfte, ließe einen Verweis auf den leeren Druck durch —
// und der ginge als Anfrage an Scryfall hinaus.
func TestDasVerweissetzenWeistUnbrauchbareRuempfeAb(t *testing.T) {
	t.Parallel()

	for name, rumpf := range map[string]string{
		"kein JSON":    `{`,
		"ohne Druck":   `{}`,
		"leerer Druck": `{"printing":"   "}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				name: "Verweis setzen", methode: http.MethodPut,
				weg: "/karten/2/verweis", rumpf: rumpf,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Und eine leere Setliste kommt als `[]` und nicht als `null`.**
//
// Genau die Naht, vor der §3 warnt: ein `nil`-Slice wird zu `null`, die Seite
// prüft auf `.length` und bleibt leer — und keine der beiden Testebenen sieht
// es, weil keine durch die Serialisierung geht.
func TestDieLeereSetverweislisteKommtAlsListe(t *testing.T) {
	t.Parallel()

	// **Der zählende Doppelgänger gibt `nil`** — genau der Fall, den
	// `ohneNilVerweise` abfängt.
	antwort := ruf(t, dienstMit(t, &zaehlender{}), endpunkt{
		name: "Setverweise lesen", methode: http.MethodGet, weg: "/sets/1/verweise",
	}, gueltig)

	if !strings.Contains(antwort.Body.String(), `"verweise":[]`) {
		t.Errorf("Rumpf: %s", antwort.Body.String())
	}
}
