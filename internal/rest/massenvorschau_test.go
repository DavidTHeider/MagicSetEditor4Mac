package rest_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau einer Massenänderung über REST** — Werkstück U6 von Phase
// 5.1 (#1417): `POST`, weil sie einen Rumpf braucht, und trotzdem lesend —
// sie schreibt nichts.
func TestDieVorschauAntwortetMitZeilen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMit(t, &zaehlender{}), endpunkt{
		methode: http.MethodPost, weg: "/sets/1/felder/vorschau",
		rumpf: `{"karteIDs":[7,8],"feld":"artist","wert":"X"}`,
	}, gueltig)
	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	var gelesen struct {
		Feld   string `json:"feld"`
		Wert   string `json:"wert"`
		Zeilen []struct {
			KarteID    string `json:"karteID"`
			Alt        string `json:"alt"`
			GabAlt     bool   `json:"gabAlt"`
			Neu        string `json:"neu"`
			Ausschluss string `json:"ausschluss"`
		} `json:"zeilen"`
	}
	if err := json.Unmarshal(antwort.Body.Bytes(), &gelesen); err != nil {
		t.Fatalf("Rumpf %q: %v", antwort.Body.String(), err)
	}

	if gelesen.Feld != "artist" || gelesen.Wert != "X" || len(gelesen.Zeilen) != 2 {
		t.Fatalf("Antwort: %+v", gelesen)
	}

	if gelesen.Zeilen[0].KarteID != "7" || gelesen.Zeilen[0].Alt != "Alt" || !gelesen.Zeilen[0].GabAlt ||
		gelesen.Zeilen[0].Neu != "X" || gelesen.Zeilen[0].Ausschluss != "" {
		t.Errorf("erste Zeile: %+v", gelesen.Zeilen[0])
	}

	if gelesen.Zeilen[1].Ausschluss != domain.AusschlussGleich {
		t.Errorf("zweite Zeile: %+v", gelesen.Zeilen[1])
	}
}

// **`POST /sets/{id}/felder` nimmt eine Erwartung an** und antwortet dann mit
// dem Vorgang des vergleichenden Wegs.
func TestFelderSetzenMitErwartungUeberRest(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMit(t, &zaehlender{}), endpunkt{
		methode: http.MethodPost, weg: "/sets/1/felder",
		rumpf: `{"karteIDs":[7,8],"feld":"artist","wert":"X",` +
			`"erwartet":[{"karteID":7,"wert":"Alt","gabEs":true},{"karteID":8,"gabEs":false}]}`,
	}, gueltig)
	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if !strings.Contains(antwort.Body.String(), `"vorgang8"`) {
		t.Errorf("Rumpf %q nennt den Vorgang des vergleichenden Wegs nicht", antwort.Body.String())
	}
}

// **Eine Erwartung außerhalb der Auswahl ist ein Bedienfehler** (400) — und
// nichts wird geschrieben.
func TestEineErwartungAusserhalbDerAuswahlIstEinBedienfehler(t *testing.T) {
	t.Parallel()

	zeuge := &zaehlender{}

	antwort := ruf(t, dienstMit(t, zeuge), endpunkt{
		methode: http.MethodPost, weg: "/sets/1/felder",
		rumpf: `{"karteIDs":[7],"feld":"artist","wert":"X","erwartet":[{"karteID":9,"gabEs":false}]}`,
	}, gueltig)

	pruefeAbweisung(t, antwort, http.StatusBadRequest, "validierung")

	if zeuge.rufe != 0 {
		t.Errorf("geschrieben wurde trotzdem: %d Rufe", zeuge.rufe)
	}
}
