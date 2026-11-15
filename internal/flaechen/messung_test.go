package flaechen_test

// **Die Kennung von außen und die Zahlen des Dienstes** — OPS-05.AK-01/02
// (#1722),
// am **gebauten** Dienst: die Kennung entsteht in seiner Wurzel, und die
// Messung hängt dort.

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"
)

// kennungsruf fragt mit einer mitgebrachten Kennung und gibt die zurückgegebene.
func (d dienst) kennungsruf(t *testing.T, weg, kennung string) (status int, zurueck string) {
	t.Helper()

	anfrage, err := http.NewRequestWithContext(t.Context(), http.MethodGet, d.adresse+weg, http.NoBody)
	if err != nil {
		t.Fatalf("Anfrage bauen: %v", err)
	}

	anfrage.Header.Set("Authorization", "Bearer "+d.marke)
	anfrage.Header.Set(vorgangskopf, kennung)

	kam, fehler := d.klient.Do(anfrage)
	if fehler != nil {
		t.Fatalf("%s: %v", weg, fehler)
	}

	if schliessfehler := kam.Body.Close(); schliessfehler != nil {
		t.Errorf("den Rumpf schließen: %v", schliessfehler)
	}

	return kam.StatusCode, kam.Header.Get(vorgangskopf)
}

// **Wer eine Kennung mitbringt, findet sie im Kopf der Antwort wieder** —
// und damit in jeder Protokollzeile dieses Rufs. Eine fremde Gestalt wird
// nicht übernommen.
func TestDerDienstUebernimmtEineMitgebrachteKennung(t *testing.T) {
	t.Parallel()

	an := starteDienst(t, deckumgebung(t))

	if code, zurueck := an.kennungsruf(t, "/gesundheit", "0123456789abcdef"); code != http.StatusOK ||
		zurueck != "0123456789abcdef" {
		t.Errorf("%d, %q", code, zurueck)
	}

	if _, zurueck := an.kennungsruf(t, "/gesundheit", "NICHT-HEX"); !achtHexziffern.MatchString(zurueck) {
		t.Errorf("eine fremde Gestalt wurde übernommen: %q", zurueck)
	}

	// **Und die Rufe stehen in der Verteilung** — gemessen im Dienst selbst.
	code, rumpf := an.ruf(t, http.MethodGet, "/metriken", "")
	if code != http.StatusOK {
		t.Fatalf("/metriken gab %d: %s", code, rumpf)
	}

	var zahlen struct {
		Laufzeiten []struct {
			Flaeche string `json:"flaeche"`
			Anzahl  int    `json:"anzahl"`
		} `json:"laufzeiten"`
		Ungemessen []string `json:"ungemessen"`
		Datenbank  *struct {
			Offen int `json:"offen"`
		} `json:"datenbank"`
	}

	if fehler := json.Unmarshal([]byte(rumpf), &zahlen); fehler != nil {
		t.Fatal(fehler)
	}

	if len(zahlen.Laufzeiten) == 0 || zahlen.Laufzeiten[0].Flaeche != "dienst" || zahlen.Laufzeiten[0].Anzahl < 2 ||
		!slices.Contains(zahlen.Ungemessen, "fenster") || zahlen.Datenbank == nil || zahlen.Datenbank.Offen < 1 {
		t.Errorf("%s", rumpf)
	}
}
