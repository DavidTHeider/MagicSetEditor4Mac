package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Was schon geschrieben ist, wird auch im Fehlerfall genannt** —
// #1041.
//
// # Warum das mehr ist als Höflichkeit
//
// Die Werkzeugerklärung sagt: „Gibt die Kennung der angelegten Karte zurück;
// zurückgenommen wird über »schreiben.karte.loeschen«." **Genau im Fehlerfall
// kam sie nicht** — der halbe Stand war über diese Fläche weder auffindbar
// noch gezielt zurücknehmbar.
//
// Angestrichen vom Nachlauf über Phase V.
type halbeAnlage struct{ williger }

func (halbeAnlage) Anlegen(context.Context, int64, string) (rest.Anlage, error) {
	return rest.Anlage{
			KarteID: 42, Felder: []string{"name"}, Befunde: []string{},
		},
		errors.New("das Feld nimmt den Wert nicht an: " + domain.ErrValidation.Error())
}

func TestEinHalberStandKommtMit(t *testing.T) {
	t.Parallel()

	dienst := rest.NeuerDienst(probemarke, func() time.Time {
		return time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)
	},
		rest.Naehte{
			Leser:         williger{},
			Schreiber:     williger{},
			Ausfuhr:       williger{},
			Abgleich:      williger{},
			Bilder:        williger{},
			Auskunft:      williger{},
			Sammeldaten:   williger{},
			Anlage:        halbeAnlage{},
			Saetze:        williger{},
			Konsole:       williger{},
			Wortwahl:      nil,
			Woerter:       nil,
			Pack:          williger{},
			Angaben:       williger{},
			Stile:         williger{},
			Zusatz:        williger{},
			Pakete:        williger{},
			Zahlen:        williger{},
			Verlauf:       williger{},
			Einstellungen: williger{},
			Fassungen:     williger{},
			Anleger:       nil,
		})

	antwort := ruf(t, dienst, endpunkt{
		name: "Karte", methode: http.MethodPost, weg: "/sets/1/karten/scryfall",
		rumpf: `{"name":"Prüfstein"}`, erfolg: http.StatusCreated,
	}, gueltig)

	if antwort.Code == http.StatusCreated {
		t.Fatalf("Status %d — der Aufruf ist gescheitert", antwort.Code)
	}

	var gelesen struct {
		Code      string `json:"code"`
		Meldung   string `json:"meldung"`
		Teilstand *struct {
			KarteID int64    `json:"karteID"`
			Felder  []string `json:"felder"`
		} `json:"teilstand"`
	}

	if err := json.Unmarshal(antwort.Body.Bytes(), &gelesen); err != nil {
		t.Fatalf("Antwort unlesbar: %v — %s", err, antwort.Body.String())
	}

	if gelesen.Teilstand == nil {
		t.Fatalf("kein Teilstand in der Fehlerantwort: %s", antwort.Body.String())
	}

	if gelesen.Teilstand.KarteID != 42 {
		t.Errorf("KarteID = %d — ohne sie ist der halbe Stand nicht zurücknehmbar",
			gelesen.Teilstand.KarteID)
	}

	if len(gelesen.Teilstand.Felder) != 1 {
		t.Errorf("Felder = %v — was schon geschrieben ist, gehört genannt",
			gelesen.Teilstand.Felder)
	}
}

// **Ein Fehler ohne Teilstand trägt keinen.**
//
// Ein leeres Feld in jeder Fehlerantwort wäre eine Frage, die niemand stellt —
// und ein `teilstand: null` ließe jede Seite zweimal prüfen.
func TestEinFehlerOhneTeilstandTraegtKeinen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		name: "Karte", methode: http.MethodPost, weg: "/sets/keine/karten/scryfall",
		rumpf: `{"name":"Prüfstein"}`, erfolg: http.StatusCreated,
	}, gueltig)

	if antwort.Body.Len() > 0 &&
		json.Valid(antwort.Body.Bytes()) &&
		containsSchluessel(antwort.Body.Bytes(), "teilstand") {
		t.Errorf("die Fehlerantwort trägt einen leeren Teilstand: %s", antwort.Body.String())
	}
}

// containsSchluessel sagt, ob ein JSON-Objekt einen Schlüssel führt.
func containsSchluessel(rumpf []byte, schluessel string) bool {
	var gelesen map[string]json.RawMessage
	if err := json.Unmarshal(rumpf, &gelesen); err != nil {
		return false
	}

	_, da := gelesen[schluessel]

	return da
}
