// **Eine Massenänderung gezielt zurücknehmen — über jede Fläche**
// (#1393).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Die Zusage, um die es geht
//
// **Die Rücknahme trifft, was gemeint war.** `zuruecknehmen` ohne Vorgang nimmt
// den jüngsten noch geltenden Schritt des Sets — am Fenster genau richtig, denn
// ⌘Z heißt „das da eben". Ohne Fenster schreibt womöglich ein zweiter Prozess
// dazwischen, und dann trifft es dessen Änderung.
//
// Deshalb gibt die Massenänderung an jeder Fläche einen **Vorgang** zurück, und
// die Rücknahme nimmt ihn entgegen.
//
// # Der Fall ist hergestellt, nicht beschrieben
//
// Zwischen Massenänderung und Rücknahme schreibt dieser Lauf **wirklich** in
// dasselbe Set. Ohne die Vorgangskennung ginge diese Zwischenänderung zurück
// und die Massenänderung bliebe stehen — beides wird geprüft.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// vorgangsruecknahme ist, was eine Fläche für diesen Lauf können muss.
//
// **Drei Schritte und kein Verbund je Fläche**: die Massenänderung gibt den
// Vorgang, dazwischen wird geschrieben, dann wird zurückgenommen. Wer die drei
// zusammenfasste, könnte nicht mehr sagen, welcher davon eine Fläche nicht
// kann.
type vorgangsruecknahme struct {
	aendern    func(t *testing.T, umgebung []string, setID string, karteIDs []string) string
	dazwischen func(t *testing.T, umgebung []string, setID, karteID string)
	zurueck    func(t *testing.T, umgebung []string, setID, vorgang string)
}

func TestEineMassenaenderungGehtGezieltZurueckUeberJedeFlaeche(t *testing.T) {
	for name, flaeche := range map[string]vorgangsruecknahme{
		"Kommandozeile": {aendern: kzAendern, dazwischen: kzDazwischen, zurueck: kzZurueck},
		"MCP":           {aendern: mcpAendern, dazwischen: mcpDazwischen, zurueck: mcpZurueck},
		"REST":          {aendern: restAendern, dazwischen: restDazwischen, zurueck: restZurueck},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung, setID, karteIDs := setMitKarten(t)
			zwei := karteIDs[:2]

			vorgang := flaeche.aendern(t, umgebung, setID, zwei)
			if vorgang == "" {
				t.Fatal("die Massenänderung nennt ihren Vorgang nicht")
			}

			// **Ein anderer schreibt dazwischen** — das ist der ganze Befund.
			flaeche.dazwischen(t, umgebung, setID, zwei[0])

			flaeche.zurueck(t, umgebung, setID, vorgang)

			// **Die Massenänderung ist fort** …
			for _, karteID := range zwei {
				if wert := feldwert(t, umgebung, setID, karteID, "artist"); wert != "" {
					t.Errorf("Karte %s trägt noch %q", karteID, wert)
				}
			}

			// … **und die Zwischenänderung steht.** Das ist die Hälfte, die
			// eine Rücknahme des jüngsten Schritts falsch macht.
			if wert := feldwert(t, umgebung, setID, zwei[0], "rarity"); wert != "dazwischen" {
				t.Errorf("die Zwischenänderung trägt %q, erwartet war „dazwischen\"", wert)
			}
		})
	}
}

func kzAendern(t *testing.T, umgebung []string, setID string, karteIDs []string) string {
	t.Helper()

	argumente := append(
		[]string{"felder-setzen", setID, "artist", "Aus der Fläche", "--json"}, karteIDs...)

	var bericht struct {
		Vorgang string `json:"vorgang"`
	}

	if err := json.Unmarshal(
		[]byte(mussKommandozeile(t, umgebung, argumente...)), &bericht); err != nil {
		t.Fatalf("kein JSON: %v", err)
	}

	return bericht.Vorgang
}

func kzDazwischen(t *testing.T, umgebung []string, setID, karteID string) {
	t.Helper()

	mussKommandozeile(t, umgebung, "feld", setID, karteID, "rarity", "dazwischen")
}

func kzZurueck(t *testing.T, umgebung []string, setID, vorgang string) {
	t.Helper()

	if aus := mussKommandozeile(
		t, umgebung, "zuruecknehmen", setID, vorgang); !strings.Contains(aus, "Bewegt") {
		t.Errorf("die Rücknahme meldet %q", aus)
	}
}

func mcpAendern(t *testing.T, umgebung []string, setID string, karteIDs []string) string {
	t.Helper()

	text, istFehler := mcpRuf(t, umgebung, "schreiben.felder", map[string]any{
		"setID": zahl(t, setID), "karteIDs": zahlen(t, karteIDs),
		"feld": "artist", "wert": "Aus der Fläche",
	})
	if istFehler {
		t.Fatalf("schreiben.felder: %s", text)
	}

	var bericht struct {
		Vorgang string `json:"Vorgang"`
	}

	if err := json.Unmarshal([]byte(text), &bericht); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, text)
	}

	return bericht.Vorgang
}

func mcpDazwischen(t *testing.T, umgebung []string, setID, karteID string) {
	t.Helper()

	if text, istFehler := mcpRuf(t, umgebung, "schreiben.feld", map[string]any{
		"setID": zahl(t, setID), "karteID": zahl(t, karteID),
		"feld": "rarity", "wert": "dazwischen",
	}); istFehler {
		t.Fatalf("schreiben.feld: %s", text)
	}
}

func mcpZurueck(t *testing.T, umgebung []string, setID, vorgang string) {
	t.Helper()

	text, istFehler := mcpRuf(t, umgebung, "schreiben.zuruecknehmen", map[string]any{
		"setID": zahl(t, setID), "vorgang": vorgang,
	})
	if istFehler {
		t.Fatalf("schreiben.zuruecknehmen: %s", text)
	}

	// **Der MCP-Server gibt eingerücktes JSON** — ein Vergleich ohne
	// Leerzeichen wäre grün, solange niemand die Einrückung anfasst.
	if !strings.Contains(text, `"bewegt": true`) {
		t.Errorf("die Rücknahme meldet %q", text)
	}
}

func restAendern(t *testing.T, umgebung []string, setID string, karteIDs []string) string {
	t.Helper()

	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/felder",
		`{"karteIDs":`+alsJSONZahlen(t, karteIDs)+
			`,"feld":"artist","wert":"Aus der Fläche"}`)
	if status != http.StatusOK {
		t.Fatalf("POST felder: %d (%s)", status, rumpf)
	}

	var bericht struct {
		Vorgang string `json:"vorgang"`
	}

	if err := json.Unmarshal([]byte(rumpf), &bericht); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, rumpf)
	}

	return bericht.Vorgang
}

func restDazwischen(t *testing.T, umgebung []string, setID, karteID string) {
	t.Helper()

	an := starteDienst(t, umgebung)

	if status, rumpf := an.ruf(t, http.MethodPut,
		"/sets/"+setID+"/karten/"+karteID+"/felder/rarity",
		`{"wert":"dazwischen"}`); status != http.StatusOK {
		t.Fatalf("PUT feld: %d (%s)", status, rumpf)
	}
}

func restZurueck(t *testing.T, umgebung []string, setID, vorgang string) {
	t.Helper()

	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodPost,
		"/sets/"+setID+"/zuruecknehmen?vorgang="+vorgang, "")
	if status != http.StatusOK {
		t.Fatalf("POST zurueck: %d (%s)", status, rumpf)
	}

	if !strings.Contains(rumpf, `"bewegt":true`) {
		t.Errorf("die Rücknahme meldet %q", rumpf)
	}
}
