package mtgjsonpreise_test

// **Die Preise, die MTGJSON führt** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Gegen selbst geschriebene Beispiele
//
// Die echten Dateien wiegen ein Gigabyte und liegen auf einem fremden Server
// (§16): geprüft wird gegen **eingecheckte, selbst geschriebene** Gestalten in
// derselben Form, die die Dokumentation nennt. Ein Lauf gegen den Dienst
// stünde hinter einem Build-Tag und zählte nicht zur Abdeckung.

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjsonpreise"
)

// dieGestalt ist die Form, die MTGJSON dokumentiert.
//
// **Nachgelesen am 2026-09-11** unter <https://mtgjson.com/data-models/price/>
// samt `price-list` und `price-points`: je Karte zwei Märkte, je Markt mehrere
// Anbieter, je Anbieter eine Währung und zwei Richtungen, je Richtung drei
// Ausführungen, je Ausführung ein **Datum** und ein Preis.
const dieGestalt = `{
  "meta": {"date": "2026-09-11", "version": "5.2.2"},
  "data": {
    "aaaa1111-2222-3333-4444-555555555555": {
      "paper": {
        "cardkingdom": {
          "currency": "USD",
          "retail":  {"normal": {"2026-09-10": 4.17}, "foil": {"2026-09-10": 12.00}},
          "buylist": {"normal": {"2026-09-10": 2.10}}
        },
        "cardmarket": {
          "currency": "EUR",
          "retail": {"normal": {"2026-09-09": 3.99, "2026-09-10": 4.05}}
        }
      },
      "mtgo": {
        "cardhoarder": {
          "currency": "USD",
          "retail": {"normal": {"2026-09-10": 0.03}}
        }
      }
    },
    "bbbb1111-2222-3333-4444-555555555555": {
      "paper": {
        "tcgplayer": {"currency": "USD", "retail": {"normal": {"2026-09-10": 1.00}}}
      }
    }
  }
}`

// gepackt macht aus der Probegestalt einen gzip-Strom.
func gepackt(t *testing.T) []byte {
	text := dieGestalt

	t.Helper()

	var puffer bytes.Buffer

	schreiber := gzip.NewWriter(&puffer)

	if _, fehler := schreiber.Write([]byte(text)); fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return puffer.Bytes()
}

// gewollt ist der Crosswalk: MTGJSON-Kennung → Scryfall-Alias.
func gewollt() map[string]string {
	return map[string]string{
		"aaaa1111-2222-3333-4444-555555555555": "scry-aaaa",
	}
}

// **Aus einer Karte werden fünf Reihen — und die zweite Karte bleibt draußen.**
func TestAusEinerKarteWerdenIhreReihen(t *testing.T) {
	t.Parallel()

	staende, befunde, fehler := mtgjsonpreise.Lies(
		strings.NewReader(dieGestalt), gewollt(), mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %+v", befunde)
	}

	// **Fünf**: cardkingdom retail normal + foil, cardkingdom buylist normal,
	// cardmarket retail normal an **zwei** Tagen, cardhoarder retail normal.
	if len(staende) != 6 {
		t.Fatalf("%d Stände: %+v", len(staende), staende)
	}

	for _, stand := range staende {
		if stand.Printing != "scry-aaaa" {
			t.Errorf("der Stand gehört zu %q", stand.Printing)
		}

		if stand.Lage != domain.PreisGenannt {
			t.Errorf("die Lage ist %q", string(stand.Lage))
		}
	}
}

// **Jede Angabe der Reihe kommt aus der Datei.**
func TestJedeAngabeKommtAusDerDatei(t *testing.T) {
	t.Parallel()

	staende, _, fehler := mtgjsonpreise.Lies(
		strings.NewReader(dieGestalt), gewollt(), mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	gesucht := domain.Preisstand{
		Printing:   "scry-aaaa",
		Quelle:     "mtgjson:cardkingdom",
		Markt:      domain.MarktPapier,
		Art:        domain.PreisartAnkauf,
		Waehrung:   "USD",
		Finish:     domain.FinishNonfoil,
		Lage:       domain.PreisGenannt,
		Wert:       210,
		Beobachtet: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
	}

	for _, stand := range staende {
		if stand.Reihe() == gesucht.Reihe() {
			if stand.Wert != gesucht.Wert || !stand.Beobachtet.Equal(gesucht.Beobachtet) {
				t.Fatalf("der Ankaufspreis ist %+v", stand)
			}

			return
		}
	}

	t.Fatalf("der Ankaufspreis bei cardkingdom fehlt: %+v", staende)
}

// **Der Onlinemarkt bleibt der Onlinemarkt.**
func TestDerOnlinemarktBleibtGetrennt(t *testing.T) {
	t.Parallel()

	staende, _, fehler := mtgjsonpreise.Lies(
		strings.NewReader(dieGestalt), gewollt(), mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	var online int

	for _, stand := range staende {
		if stand.Markt == domain.MarktOnline {
			online++

			if stand.Quelle != "mtgjson:cardhoarder" {
				t.Errorf("der Onlinepreis kommt von %q", string(stand.Quelle))
			}
		}
	}

	if online != 1 {
		t.Errorf("%d Onlinepreise, erwartet einer", online)
	}
}

// **Ein gepackter Strom wird ausgepackt.**
func TestEinGepackterStromWirdGelesen(t *testing.T) {
	t.Parallel()

	staende, _, fehler := mtgjsonpreise.LiesGepackt(
		bytes.NewReader(gepackt(t)), gewollt(), mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("LiesGepackt: %v", fehler)
	}

	if len(staende) != 6 {
		t.Errorf("%d Stände", len(staende))
	}
}
