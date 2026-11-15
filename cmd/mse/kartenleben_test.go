// `mse karte` und `mse suche` — der Lebenszyklus und der Volltext auf der
// Kommandozeile
// (#1107).
//
// **Fünf Fähigkeiten, die der MCP-Server seit #591 und #886 kann und die
// Kommandozeile nicht.** Die fachliche Naht ist dieselbe (`app.Schreibnaht`),
// nur hat sie niemand von hier aus gerufen — genau die leere Zelle, die #1107
// zählt.
package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// **`karte anlegen` legt eine Karte an und nennt ihre Kennung.**
func TestKarteAnlegenNenntDieNeueKennung(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "karte", setID, "anlegen", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		KarteID string `json:"karteID"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if _, err := strconv.ParseInt(antwort.KarteID, 10, 64); err != nil {
		t.Fatalf("keine Kennung: %q", antwort.KarteID)
	}

	// **Und sie steht danach in der Liste** — der Beleg, dass geschrieben
	// wurde und nicht nur geantwortet.
	_, liste, _ := fahreAn(t, ort, "karten", setID)
	if !strings.Contains(liste, antwort.KarteID) {
		t.Fatalf("die neue Karte %s fehlt in der Liste:\n%s", antwort.KarteID, liste)
	}
}

// **`karte kopieren` nimmt die Werte mit.**
func TestKarteKopierenUebernimmtDieWerte(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "karte", setID, "kopieren", karteID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		KarteID string `json:"karteID"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if antwort.KarteID == karteID {
		t.Fatalf("die Kopie trägt die Kennung des Originals: %s", antwort.KarteID)
	}

	_, liste, _ := fahreAn(t, ort, "karten", setID)
	if strings.Count(liste, "Feuerball") != 2 {
		t.Fatalf("der Name wurde nicht mitgenommen:\n%s", liste)
	}
}

// **`karte loeschen` nimmt sie aus der Liste.**
func TestKarteLoeschenEntferntSieAusDerListe(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, _, fehler := fahreAn(t, ort, "karte", setID, "loeschen", karteID)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	_, liste, _ := fahreAn(t, ort, "karten", setID)
	if strings.Contains(liste, "Feuerball") {
		t.Fatalf("die Karte steht noch da:\n%s", liste)
	}
}

// **`karte tauschen` dreht die Reihenfolge um.**
func TestKarteTauschenDrehtDieReihenfolge(t *testing.T) {
	ort, setID, ersteID := mitKarte(t)

	_, aus, _ := fahreAn(t, ort, "karte", setID, "anlegen", "--json")

	var angelegt struct {
		KarteID string `json:"karteID"`
	}

	if err := json.Unmarshal([]byte(aus), &angelegt); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	vorher := nummerVon(t, ort, setID, ersteID)

	code, _, fehler := fahreAn(t, ort, "karte", setID, "tauschen", ersteID, angelegt.KarteID)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if nachher := nummerVon(t, ort, setID, ersteID); nachher == vorher {
		t.Fatalf("die Nummer steht unverändert bei %d", nachher)
	}
}

// nummerVon liest die laufende Nummer einer Karte über `mse karten`.
func nummerVon(t *testing.T, ort, setID, karteID string) int {
	t.Helper()

	_, aus, _ := fahreAn(t, ort, "karten", setID, "--json")

	var antwort struct {
		Karten []struct {
			ID     string `json:"id"`
			Nummer int    `json:"nummer"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	for _, karte := range antwort.Karten {
		if karte.ID == karteID {
			return karte.Nummer
		}
	}

	t.Fatalf("die Karte %s steht nicht in der Liste:\n%s", karteID, aus)

	return 0
}

// **`suche` findet über den Volltext.**
func TestSucheFindetUeberDenVolltext(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "suche", "Feuerball", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		Treffer []struct {
			KarteID string `json:"karteID"`
			Feld    string `json:"feld"`
		} `json:"treffer"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if len(antwort.Treffer) == 0 {
		t.Fatalf("kein Treffer:\n%s", aus)
	}

	// **Die Gegenprobe**: ein Wort, das nirgends steht, findet nichts. Ohne
	// sie sagte der Test nur, dass die Ausgabe nicht leer ist.
	_, nichts, _ := fahreAn(t, ort, "suche", "Zaubermeisterin", "--json")
	if err := json.Unmarshal([]byte(nichts), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, nichts)
	}

	if len(antwort.Treffer) != 0 {
		t.Fatalf("ein Wort, das nicht vorkommt, hat getroffen:\n%s", nichts)
	}

	// **Und mit Set eingeschränkt geht es auch.**
	code, _, fehler = fahreAn(t, ort, "suche", "Feuerball", "--set="+setID)
	if code != CodeGut {
		t.Fatalf("mit Set: Rückgabewert %d (%s)", code, fehler)
	}
}

// **`felder` nennt die Feldnamen eines Sets.**
func TestFelderNenntDieFeldnamen(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "felder", setID)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "name") {
		t.Fatalf("das Feld »name« fehlt:\n%s", aus)
	}
}
