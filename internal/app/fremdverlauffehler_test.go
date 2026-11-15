package app_test

// **Ein Fremdverlauf, der scheitert, sagt warum** — gefunden von I05
// (#2263)
// am nativen Fenster: Scrydex und CardTrader antworteten über einen
// geschlossenen Proxy mit „Es ist ein unerwarteter Fehler aufgetreten",
// Scryfall daneben mit „der fremde Dienst ist nicht erreichbar". Ein Fehler,
// der draußen liegt, ist ein Netzfehler mit Rat, kein unbekannter.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinUnerreichbarerScrydexIstEinNetzfehler(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	adresse := server.URL
	server.Close()

	db := scrydexbestand(t, adresse)

	antwort := app.AnbieterAnwendungUeber(db).MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "", nil)

	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNetz {
		t.Fatalf("der Fehler ist %+v — erwartet %s", antwort.Fehler, app.CodeNetz)
	}
}

// **Ein ausgeschalteter Dienst sagt, wo er sich einschalten lässt** — der
// Grund aus I05: `domain.ErrDienstAus` stand in keiner Übersetzung, und
// „unter Einstellungen lässt er sich einschalten" kam als „Es ist ein
// unerwarteter Fehler aufgetreten" an. Der Druck liegt da, der Dienst ist aus
// — erst dann erreicht der Weg die Frage, ob er fragen darf.
func TestEinAusgeschalteterDienstSagtWoErSichEinschaltenLaesst(t *testing.T) {
	db := frischerBestand(t)
	if fehler := storage.NeueDrucke(db).Sichern(probekontext(t), &domain.Druck{
		Name: "Probe", Setcode: "ptkdf", Nummer: "1", Sprache: "en",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
			Wert: "11111111-1111-4111-8111-111111111111",
		}},
	}); fehler != nil {
		t.Fatalf("den Druck sichern: %v", fehler)
	}

	anwendung := app.AnbieterAnwendungUeber(db).MitMarktbund(marktbund{marke: "geheim"})

	for _, fall := range []struct {
		dienst  string
		antwort func() *app.Fehler
	}{
		{"scrydex", func() *app.Fehler {
			return anwendung.Fremdpreisverlauf(probekontext(t), "scrydex",
				"11111111-1111-4111-8111-111111111111", "", nil).Fehler
		}},
		{"cardtrader", func() *app.Fehler {
			return anwendung.Marktangebote(probekontext(t), "cardtrader",
				"11111111-1111-4111-8111-111111111111", "", "", nil).Fehler
		}},
	} {
		t.Run(fall.dienst, func(t *testing.T) {
			fehler := fall.antwort()
			if fehler == nil || fehler.Code != app.CodeAbgeschaltet ||
				fehler.Schluessel != domain.SchluesselDienstAus ||
				fehler.Werte["dienst"] != fall.dienst {
				t.Fatalf("der Fehler ist %+v — erwartet %s/%s mit dem Dienst %s",
					fehler, app.CodeAbgeschaltet, domain.SchluesselDienstAus, fall.dienst)
			}
		})
	}
}
