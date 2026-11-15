// **Wie eine Kartenliste an drei Flächen aussieht** —
// #1169.
//
// **Eine eigene Datei**, weil `kartenleben_test.go` mit den Lesern über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, **was**
// eine Karte durchläuft, hier, **wie** die drei Flächen darüber Auskunft
// geben.
//
// # Die Regel dahinter
//
// **Die Flächen dürfen verschieden aussehen, nicht verschieden antworten.**
// Ein Leser je Gestalt ist deshalb kein Zugeständnis, sondern die Stelle, an
// der die Gestalt endet und die Fachlichkeit anfängt.

package flaechen_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// kartenkennungen zieht die Kartenkennungen **in ihrer Reihenfolge**.
//
// **Der Reihe nach und nicht als Menge**: der Tausch ist genau die Aussage
// über die Reihenfolge, und eine Menge sagte darüber nichts.
//
// # Drei Gestalten, eine Auskunft
//
// `mse karten --json` und der Dienst geben `{"karten":[{"id":"7"}]}`,
// `lesen.karten` eine nackte Liste mit `ID` als Zahl. **Das dürfen sie** —
// gleich sein muss die Fachlichkeit, nicht die Ausgabe (so steht es schon an
// [ersteKarteMCP]). Hier wird deshalb beides gelesen und nicht eines von
// beiden erzwungen.
func kartenkennungen(t *testing.T, roh string) []string {
	t.Helper()

	var gehuellt struct {
		Karten []struct {
			ID kennung `json:"id"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(roh), &gehuellt); err == nil && len(gehuellt.Karten) > 0 {
		kennungen := make([]string, 0, len(gehuellt.Karten))
		for _, karte := range gehuellt.Karten {
			kennungen = append(kennungen, string(karte.ID))
		}

		return kennungen
	}

	var nackt []struct {
		ID kennung `json:"id"`
	}

	if err := json.Unmarshal([]byte(roh), &nackt); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if len(nackt) == 0 {
		t.Fatalf("in der Antwort steht keine Kartenkennung:\n%s", roh)
	}

	kennungen := make([]string, 0, len(nackt))
	for _, karte := range nackt {
		kennungen = append(kennungen, string(karte.ID))
	}

	return kennungen
}

// kennung nimmt eine Kartenkennung als Zeichenkette **oder** als Zahl.
type kennung string

func (k *kennung) UnmarshalJSON(roh []byte) error {
	*k = kennung(strings.Trim(string(roh), `"`))

	return nil
}

// kartenname gibt den Namen einer Karte aus einer Kartenliste.
//
// **Über drei Gestalten**, wie [kartenkennungen] — und aus demselben Grund:
// die Flächen dürfen verschieden aussehen, nicht verschieden antworten.
func kartenname(t *testing.T, roh, karteID string) string {
	t.Helper()

	for kennung, name := range namenNachKennung(t, roh) {
		if kennung == karteID {
			return name
		}
	}

	t.Fatalf("die Karte %s steht nicht in der Liste:\n%s", karteID, roh)

	return ""
}

// namenGezaehlt zählt, wie oft jeder Kartenname vorkommt.
func namenGezaehlt(t *testing.T, roh string) map[string]int {
	t.Helper()

	gezaehlt := map[string]int{}
	for _, name := range namenNachKennung(t, roh) {
		gezaehlt[name]++
	}

	return gezaehlt
}

// namenNachKennung liest Kennung und Name je Karte.
//
// # Drei Gestalten, und keine ist die richtige
//
// | Fläche | Wie der Name dasteht |
// | --- | --- |
// | `mse karten --json` | `{"karten":[{"id":"1","name":"…","werte":{…}}]}` |
// | `GET /sets/{id}/karten` | dieselbe Hülle |
// | `lesen.karten` | eine nackte Liste mit `ID` als Zahl und `Values` als Liste |
// | das Fenster (`Karten`, wie Wails es serialisiert) | dieselbe Hülle, `werte` als Liste aus `feld` und `wert` (C04) |
//
// **Das dürfen sie** — gleich ist die Fachlichkeit, nicht die Ausgabe (so
// steht es schon an [kartenkennungen]). Gelesen wird deshalb beides.
func namenNachKennung(t *testing.T, roh string) map[string]string {
	t.Helper()

	type karte struct {
		ID   kennung `json:"id"`
		Name string  `json:"name"`

		// Values ist die Gestalt von `lesen.karten` — Go-Feldnamen, weil dort
		// die Entität selbst geschrieben wird.
		Values []struct {
			Field string `json:"Field"`
			Text  string `json:"Text"`
		} `json:"Values"`

		// Werte ist an Kommandozeile und Dienst ein Objekt, am Fenster eine
		// Liste — roh gelesen, damit die eine Gestalt die andere nicht bricht.
		Werte json.RawMessage `json:"werte"`
	}

	var gehuellt struct {
		Karten []karte `json:"karten"`
	}

	karten := []karte{}

	if err := json.Unmarshal([]byte(roh), &gehuellt); err == nil && len(gehuellt.Karten) > 0 {
		karten = gehuellt.Karten
	} else if lesefehler := json.Unmarshal([]byte(roh), &karten); lesefehler != nil {
		t.Fatalf("kein JSON: %v\n%s", lesefehler, roh)
	}

	if len(karten) == 0 {
		t.Fatalf("in der Antwort steht keine Karte:\n%s", roh)
	}

	namen := map[string]string{}

	for _, eine := range karten {
		if eine.Name != "" {
			namen[string(eine.ID)] = eine.Name
		}

		for _, wert := range eine.Values {
			if wert.Field == "name" {
				namen[string(eine.ID)] = wert.Text
			}
		}

		var liste []struct {
			Feld string `json:"feld"`
			Wert string `json:"wert"`
		}

		if json.Unmarshal(eine.Werte, &liste) == nil {
			for _, wert := range liste {
				if wert.Feld == "name" {
					namen[string(eine.ID)] = wert.Wert
				}
			}
		}
	}

	if len(namen) == 0 {
		t.Fatalf("keine Karte trägt einen Namen:\n%s", roh)
	}

	return namen
}
