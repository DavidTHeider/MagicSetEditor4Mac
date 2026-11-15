package app_test

// **Ein fremder Dienst, der nicht antwortet, heißt so — mit Rat** — D02
// (#2233).
//
// Bis dahin kam jeder der sechs Fälle als `unbekannt` an, und das Fenster
// zeigte „Es ist ein unerwarteter Fehler aufgetreten" — der Satz, der ins
// Protokoll schickt, für einen Fehler, dessen Rat klar ist. Ein 429 ebenso.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinFremderDienstHeisstSo(t *testing.T) {
	t.Parallel()

	grund := errors.New("dial tcp: lookup api.scryfall.com: no such host")

	for _, fall := range []struct {
		name       string
		fehler     error
		schluessel domain.Schluessel
		werte      map[string]string
	}{
		{
			"ein Name, der sich nicht auflöst",
			domain.NeuerGegenstellenfehler(domain.ErrNameUnaufloesbar, grund, "api.scryfall.com", 0),
			domain.SchluesselNetzName,
			map[string]string{"wirt": "api.scryfall.com"},
		},
		{
			"ein Port ohne Annahme, umschlossen vom Klienten und mit Proxyhinweis",
			fmt.Errorf("Scryfall fragen: %w (über den Proxy http://proxy:3128 aus umgebung)",
				domain.NeuerGegenstellenfehler(domain.ErrNichtErreichbar, grund, "127.0.0.1:9", 0)),
			domain.SchluesselNetzNichtErreichbar,
			map[string]string{"wirt": "127.0.0.1:9"},
		},
		{
			"keine Antwort binnen der Kopffrist",
			domain.NeuerGegenstellenfehler(domain.ErrKeineAntwort, grund, "api.scryfall.com", 30*time.Second),
			domain.SchluesselNetzKeineAntwort,
			map[string]string{"wirt": "api.scryfall.com", "sekunden": "30"},
		},
		{
			"eine Antwort, die stockt",
			domain.NeuerGegenstellenfehler(domain.ErrAntwortStockt, grund, "mtgjson.com", time.Minute),
			domain.SchluesselNetzAntwortStockt,
			map[string]string{"wirt": "mtgjson.com", "sekunden": "60"},
		},
		{
			// **#2487**: bis dahin kam ein Abriss als `unbekannt` an.
			"eine Verbindung, die abreißt",
			domain.NeuerGegenstellenfehler(domain.ErrVerbindungAbgerissen, grund, "api.scryfall.com", 0),
			domain.SchluesselNetzAbgerissen,
			map[string]string{"wirt": "api.scryfall.com"},
		},
		{
			"ein Proxy, der eine Anmeldung verlangt",
			domain.NeuerGegenstellenfehler(domain.ErrProxyAnmeldung, grund, "proxy.firma:3128", 0),
			domain.SchluesselNetzProxyAnmeldung,
			map[string]string{"wirt": "proxy.firma:3128"},
		},
		{
			"ein 429, auch ohne Wirt",
			fmt.Errorf("%w: https://api.scryfall.com/cards — bitte später noch einmal versuchen", domain.ErrGedrosselt),
			domain.SchluesselGedrosselt,
			map[string]string{},
		},
		{
			"ein gedrosseltes Einlösen",
			domain.ErrEinloesenGedrosselt,
			domain.SchluesselGedrosselt,
			map[string]string{},
		},
		{
			// **Mit Ende die Uhrzeit** (R01, #2333): bis zum Zweitblick über
			// Phase 0 übersetzte das Fenster sie weg.
			"eine Drosselung mit Ende",
			domain.NeueDrosselung(grund, "api.github.com", time.Date(2026, 10, 13, 21, 37, 0, 0, time.Local)),
			domain.SchluesselGedrosseltBis,
			map[string]string{"wirt": "api.github.com", "uhrzeit": "21:37"},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			uebersetzt := app.Uebersetze(context.Background(), fall.fehler)
			if uebersetzt.Code != app.CodeNetz || uebersetzt.Schluessel != fall.schluessel {
				t.Fatalf("übersetzt als %s/%s, erwartet %s/%s", uebersetzt.Code, uebersetzt.Schluessel, app.CodeNetz, fall.schluessel)
			}

			if len(uebersetzt.Werte) != len(fall.werte) {
				t.Errorf("Werte %v, erwartet %v", uebersetzt.Werte, fall.werte)
			}

			for name, wert := range fall.werte {
				if uebersetzt.Werte[name] != wert {
					t.Errorf("%s = %q, erwartet %q", name, uebersetzt.Werte[name], wert)
				}
			}

			if strings.Contains(uebersetzt.Meldung, "no such host") {
				t.Errorf("der Grund steht in der Meldung: %q", uebersetzt.Meldung)
			}
		})
	}
}
