// **Die Fehlerwege der Scryfall-Endpunkte** — Bild, Auskünfte, Satz
// (#1107).
//
// **Eine eigene Datei**, weil `fehlerwege_test.go` mit den Setangaben über die
// Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage: dort die Fehlerwege
// des **Bestands**, hier die der **fremden Dienste**.
package rest_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Die drei Abweisungen des Bildendpunkts** (#1005).
//
// Der Weg über die vier Klassen prüft den **Rahmen**; hier stehen die Fälle,
// die nur dieser Endpunkt hat: eine unlesbare Kennung, ein unlesbarer Rumpf
// und die Vorgabe der Bildart.
func TestScryfallBildWeistAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		weg    string
		rumpf  string
		status int
	}{
		"unlesbare Setkennung": {
			"/sets/keine/karten/2/scryfall/bild",
			`{"feld":"image","name":"Prüfstein","art":"normal"}`,
			http.StatusBadRequest,
		},
		"unlesbare Kartenkennung": {
			"/sets/1/karten/keine/scryfall/bild",
			`{"feld":"image","name":"Prüfstein","art":"normal"}`,
			http.StatusBadRequest,
		},
		"unlesbarer Rumpf": {
			"/sets/1/karten/2/scryfall/bild", `{"feld":`, http.StatusBadRequest,
		},

		// **Und dieselben zwei am Anlegeweg** (#1032). Er hat **keine**
		// Kartenkennung — die Karte gibt es noch nicht —, also sind es zwei
		// statt drei.
		"Anlegen mit unlesbarer Setkennung": {
			"/sets/keine/karten/scryfall", `{"name":"Prüfstein"}`, http.StatusBadRequest,
		},
		"Anlegen mit unlesbarem Rumpf": {
			"/sets/1/karten/scryfall", `{"name":`, http.StatusBadRequest,
		},

		// **Und am Anlegeweg eines Sets** (#891). Er hat **gar keine**
		// Kennung im Weg — das Set gibt es noch nicht —, also bleibt allein
		// der Rumpf.
		"Ein neues Set mit unlesbarem Rumpf": {
			"/sets/neu", `{"name":`, http.StatusBadRequest,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				name: name, methode: http.MethodPost, weg: fall.weg, rumpf: fall.rumpf,
				erfolg: http.StatusOK,
			}, gueltig)

			if antwort.Code != fall.status {
				t.Errorf("Status = %d, erwartet %d — %s",
					antwort.Code, fall.status, antwort.Body.String())
			}
		})
	}
}

// **Ohne Bildart wird die ganze Karte geholt.**
//
// Sie trägt Künstlername und Copyright im Bild selbst; die Bedingung von
// Scryfall ist damit ohne Zutun erfüllt. Dieselbe Vorgabe wie an der
// Kommandozeile.
func TestScryfallBildOhneArtNimmtDieGanzeKarte(t *testing.T) {
	t.Parallel()

	gemerkt := &artmerker{}

	dienst := rest.NeuerDienst(probemarke, func() time.Time {
		return time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	},
		rest.Naehte{
			Leser:         williger{},
			Schreiber:     williger{},
			Ausfuhr:       williger{},
			Abgleich:      williger{},
			Bilder:        gemerkt,
			Auskunft:      nil,
			Sammeldaten:   williger{},
			Anlage:        williger{},
			Saetze:        williger{},
			Konsole:       williger{},
			Werkstatt:     williger{},
			Wortwahl:      nil,
			Woerter:       nil,
			Pack:          williger{},
			Angaben:       williger{},
			Stile:         williger{},
			Zusatz:        williger{},
			Pakete:        williger{},
			Kartenbilder:  williger{},
			Vorlagen:      williger{},
			Probelauf:     williger{},
			Zahlen:        williger{},
			Verlauf:       williger{},
			Einstellungen: williger{},
			Fassungen:     williger{},
			Anleger:       nil,
		})

	antwort := ruf(t, dienst, endpunkt{
		name: "ohne Art", methode: http.MethodPost,
		weg: "/sets/1/karten/2/scryfall/bild", rumpf: `{"feld":"image","name":"Prüfstein"}`,
		erfolg: http.StatusOK,
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status = %d: %s", antwort.Code, antwort.Body.String())
	}

	if gemerkt.art != "normal" {
		t.Errorf("Art = %q, erwartet »normal«", gemerkt.art)
	}
}

// artmerker hält fest, mit welcher Bildart gefragt wurde.
type artmerker struct {
	williger

	art string
}

func (a *artmerker) Einsetzen(
	ctx context.Context, setID, karteID int64, feld, name, art string,
) (rest.Bildeinsatz, error) {
	a.art = art

	return a.williger.Einsetzen(ctx, setID, karteID, feld, name, art)
}

// **Die beiden lesenden Auskunftsendpunkte weisen fehlende Argumente ab**
// (#888).
//
// Sie hängen an keinem Set und haben deshalb keine Kennung, an der eine Wache
// greifen könnte — ihre einzige ist der Parameter an der Adresse.
func TestDieAuskuenfteWeisenFehlendeArgumenteAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"Regeln ohne Namen":          "/scryfall/regeln",
		"Manakosten ohne Kosten":     "/scryfall/manakosten",
		"Regeln mit leerem Namen":    "/scryfall/regeln?name=",
		"Manakosten mit leerem Wert": "/scryfall/manakosten?kosten=",
	}

	for name, weg := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				name: name, methode: http.MethodGet, weg: weg, erfolg: http.StatusOK,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, erwartet 400 — %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Der Satzweg weist einen unlesbaren Rumpf ab** (#1033).
//
// Er hat **keine** Kennung im Weg — das Set entsteht erst —, also ist der
// Rumpf der einzige Weg, an dem er scheitern kann.
func TestScryfallSatzWeistEinenUnlesbarenRumpfAb(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		name: "Satz", methode: http.MethodPost, weg: "/sets/scryfall",
		rumpf: `{"code":`, erfolg: http.StatusCreated,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet 400", antwort.Code)
	}
}

// **Ein Satz, der schon liegt, gibt 200 und nicht 201** (#886).
//
// Ein 201 behauptete eine Anlage, die es nicht gab.
func TestEinSchonOffenerSatzGibt200(t *testing.T) {
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
			Anlage:        williger{},
			Saetze:        schonOffen{},
			Konsole:       schonOffen{},
			Werkstatt:     schonOffen{},
			Wortwahl:      nil,
			Woerter:       nil,
			Pack:          williger{},
			Angaben:       williger{},
			Stile:         williger{},
			Zusatz:        williger{},
			Pakete:        williger{},
			Kartenbilder:  williger{},
			Zahlen:        williger{},
			Verlauf:       williger{},
			Einstellungen: williger{},
			Fassungen:     williger{},
			Anleger:       nil,
		})

	antwort := ruf(t, dienst, endpunkt{
		name: "Satz", methode: http.MethodPost, weg: "/sets/scryfall",
		rumpf: `{"code":"lea","spiel":"magic"}`, erfolg: http.StatusOK,
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Errorf("Status %d, erwartet 200 — es ist nichts entstanden", antwort.Code)
	}
}

// schonOffen meldet einen Satz, der schon im Bestand liegt.
type schonOffen struct{ williger }

func (schonOffen) Einlesen(context.Context, string, string) (rest.Satz, error) {
	return rest.Satz{SetID: 7, Karten: 103, Befunde: []string{}, SchonOffen: true}, nil
}
