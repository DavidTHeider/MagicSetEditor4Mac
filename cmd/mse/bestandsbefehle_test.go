// `mse statistik`, `mse zuruecknehmen`, `mse wiederherstellen`,
// `mse einstellungen`
// (#1107).
package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// **`statistik` zählt Karten, Felder und leere Karten.**
func TestStatistikZaehltDasSet(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "statistik", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		Statistik struct {
			Karten   int `json:"karten"`
			Felder   int `json:"felder"`
			Belegung []struct {
				Name   string `json:"name"`
				Karten int    `json:"karten"`
			} `json:"belegung"`
		} `json:"statistik"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if antwort.Statistik.Karten != 1 {
		t.Errorf("%d Karten statt einer", antwort.Statistik.Karten)
	}

	if len(antwort.Statistik.Belegung) == 0 {
		t.Errorf("keine Belegung — die Karte trägt einen Namen:\n%s", aus)
	}
}

// **Mit Feld gibt sie die Verteilung seiner Werte.**
func TestStatistikMitFeldGibtDieVerteilung(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "statistik", setID, "name", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Feuerball") {
		t.Errorf("der Wert fehlt in der Verteilung:\n%s", aus)
	}

	// **Die Gegenprobe**: ein Feld, das keine Karte trägt, gibt nichts.
	_, leer, _ := fahreAn(t, ort, "statistik", setID, "gibtsnicht", "--json")
	if strings.Contains(leer, "Feuerball") {
		t.Errorf("ein Feld, das es nicht gibt, hat getroffen:\n%s", leer)
	}
}

// **`zuruecknehmen` nimmt zurück, `wiederherstellen` stellt wieder her.**
func TestZuruecknehmenUndWiederherstellen(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort,
		"feld", setID, karteID, "name", "Blitz"); code != CodeGut {
		t.Fatalf("schreiben: Rückgabewert %d (%s)", code, fehler)
	}

	if name := kartenname(t, ort, setID, karteID); name != "Blitz" {
		t.Fatalf("das Schreiben ist nicht angekommen: %q", name)
	}

	if code, _, fehler := fahreAn(t, ort, "zuruecknehmen", setID); code != CodeGut {
		t.Fatalf("zurücknehmen: Rückgabewert %d (%s)", code, fehler)
	}

	if name := kartenname(t, ort, setID, karteID); name != "Feuerball" {
		t.Fatalf("nach der Rücknahme steht %q statt »Feuerball«", name)
	}

	if code, _, fehler := fahreAn(t, ort, "wiederherstellen", setID); code != CodeGut {
		t.Fatalf("wiederherstellen: Rückgabewert %d (%s)", code, fehler)
	}

	if name := kartenname(t, ort, setID, karteID); name != "Blitz" {
		t.Errorf("nach dem Wiederherstellen steht %q statt »Blitz«", name)
	}
}

// **Und „nichts zu bewegen" ist kein Fehler.**
//
// Ein Skript, das bis zum Ende zurücknimmt, liefe sonst in einen Rückgabewert,
// den es nicht von einem echten Fehlschlag unterscheiden kann.
func TestNichtsZuBewegenIstKeinFehler(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "zuruecknehmen", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		Bewegt bool `json:"bewegt"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if antwort.Bewegt {
		t.Errorf("es wurde etwas bewegt, obwohl nichts geschrieben wurde:\n%s", aus)
	}
}

// kartenname liest den Namen einer Karte über `mse karten`.
func kartenname(t *testing.T, ort, setID, karteID string) string {
	t.Helper()

	_, aus, _ := fahreAn(t, ort, "karten", setID, "--json")

	var antwort struct {
		Karten []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	for _, karte := range antwort.Karten {
		if karte.ID == karteID {
			return karte.Name
		}
	}

	t.Fatalf("die Karte %s steht nicht in der Liste:\n%s", karteID, aus)

	return ""
}

// **`einstellungen` nennt jede Einstellung samt ihrem Bereich.**
func TestEinstellungenLesenUndSetzen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "einstellungen", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		Einstellungen []struct {
			Schluessel string   `json:"schluessel"`
			Wert       string   `json:"wert"`
			Werte      []string `json:"werte"`
		} `json:"einstellungen"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if len(antwort.Einstellungen) == 0 {
		t.Fatalf("keine Einstellung:\n%s", aus)
	}

	erste := antwort.Einstellungen[0]
	if len(erste.Werte) < 2 {
		t.Skipf("die erste Einstellung hat keine Auswahl: %+v", erste)
	}

	// **Ein Wert, der nicht der aktuelle ist** — sonst prüfte der Vergleich
	// danach nichts.
	neu := erste.Werte[0]
	if neu == erste.Wert {
		neu = erste.Werte[1]
	}

	if code, _, fehler := fahreAn(t, ort,
		"einstellungen", "setzen", erste.Schluessel, neu); code != CodeGut {
		t.Fatalf("setzen: Rückgabewert %d (%s)", code, fehler)
	}

	_, nachher, _ := fahreAn(t, ort, "einstellungen", "--json")
	if err := json.Unmarshal([]byte(nachher), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, nachher)
	}

	if antwort.Einstellungen[0].Wert != neu {
		t.Errorf("die Einstellung steht bei %q statt bei %q",
			antwort.Einstellungen[0].Wert, neu)
	}
}

// **Ein unbekannter Schlüssel wird abgewiesen.**
func TestEineUnbekannteEinstellungWirdAbgewiesen(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, _, fehler := fahreAn(t, ort, "einstellungen", "setzen", "gibtsnicht", "x")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d statt %d (%s)", code, CodeFachlich, fehler)
	}
}

// **Falsche Aufrufe sind Bedienfehler.**
func TestDieBestandsbefehleWeisenFalscheAufrufeAb(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	faelle := [][]string{
		{"statistik"},
		{"statistik", "keinezahl"},
		{"statistik", setID, "a", "b"},
		{"zuruecknehmen"},
		{"zuruecknehmen", "keinezahl"},
		{"wiederherstellen"},
		{"einstellungen", "setzen"},
		{"einstellungen", "gibtsnicht", "a", "b"},
	}

	for _, fall := range faelle {
		code, _, fehler := fahreAn(t, ort, fall...)
		if code != CodeBedienung {
			t.Errorf("%v endete mit %d statt %d (%s)",
				fall, code, CodeBedienung, strings.SplitN(fehler, "\n", 2)[0])
		}
	}
}
