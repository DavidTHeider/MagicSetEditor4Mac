// **Die Abweisungen der Partienwege** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Neben den vier Klassen und nicht statt ihnen.** Jene prüfen, dass der Weg
// bis zum Anwendungsfall trägt; hier stehen die Fälle, die nur diese Wege
// haben — ein fehlendes Pflichtfeld, ein unlesbarer Rumpf, eine Partie mit zu
// wenigen Sitzen.
//
// # Und warum das fehlende Feld **benannt** wird
//
// **Weil ein 400 ohne Feld niemandem sagt, wo er nachsehen soll.** Über eine
// `map` gelesen nennte dieselbe unvollständige Anfrage bei zwei fehlenden
// Feldern mal das eine und mal das andere; die Prüfung geht deshalb über eine
// **Liste**, und dieser Lauf hält sie fest.
package rest_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Jeder Partienweg weist eine unvollständige Anfrage ab — mit dem Feld.**
func TestDiePartienwegeWeisenFehlendeFelderAb(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		weg  endpunkt
		feld string
	}{
		"anlegen ohne Spiel": {
			endpunkt{
				methode: http.MethodPost, weg: "/partien",
				rumpf: `{"sitze":[{"kennung":"a"},{"kennung":"b"}]}`,
			}, "spiel",
		},
		"anlegen mit einem Sitz": {
			endpunkt{
				methode: http.MethodPost, weg: "/partien",
				rumpf: `{"spiel":"magic","sitze":[{"kennung":"a"}]}`,
			}, "sitze",
		},
		"ansehen ohne Sitz": {
			endpunkt{methode: http.MethodGet, weg: "/partien/p1"}, "fuer",
		},
		"Verlauf ohne Sitz": {
			endpunkt{methode: http.MethodGet, weg: "/partien/p1/verlauf"},
			"fuer",
		},
		"befehl ohne Absender": {
			endpunkt{
				methode: http.MethodPost, weg: "/partien/p1/befehle",
				rumpf: `{"kennung":"b1","art":"objekt.tappen"}`,
			}, "absender",
		},
		"befehl ohne Kennung": {
			endpunkt{
				methode: http.MethodPost, weg: "/partien/p1/befehle",
				rumpf: `{"absender":"a","art":"objekt.tappen"}`,
			}, "kennung",
		},
		"befehl ohne Art": {
			endpunkt{
				methode: http.MethodPost, weg: "/partien/p1/befehle",
				rumpf: `{"absender":"a","kennung":"b1"}`,
			}, "art",
		},
		"stand ohne Sitz": {
			endpunkt{
				methode: http.MethodPost, weg: "/partien/p1/stand",
				rumpf: `{"stand":"pausieren"}`,
			}, "fuer",
		},

		// **Die Einfuhr ohne Text** — PLAY-03.AK-05 (#1711): ein leerer
		// Rumpf ist kein leerer Zustand, sondern eine Anfrage ohne Inhalt.
		"Einfuhr ohne Text": {
			endpunkt{
				methode: http.MethodPost, weg: "/partien/einfuhr",
				rumpf: `{}`,
			}, "inhalt",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall.weg, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Fatalf("endete mit %d statt 400: %s",
					antwort.Code, antwort.Body.String())
			}

			// **Das Feld steht in der Antwort** — sonst wäre die Abweisung
			// eine Auskunft über nichts.
			if !strings.Contains(antwort.Body.String(), fall.feld) {
				t.Errorf("die Antwort nennt %q nicht: %s",
					fall.feld, antwort.Body.String())
			}
		})
	}
}

// **Und jeder Weg mit Rumpf weist einen unlesbaren ab.**
func TestDiePartienwegeWeisenUnlesbareRuempfeAb(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]endpunkt{
		"anlegen": {methode: http.MethodPost, weg: "/partien", rumpf: `{`},
		"befehl": {
			methode: http.MethodPost, weg: "/partien/p1/befehle", rumpf: `{`,
		},
		"stand": {
			methode: http.MethodPost, weg: "/partien/p1/stand", rumpf: `{`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), weg, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}

// **Ein leerer Verlauf geht als Liste hinaus und nicht als `null`.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil `nil` und „leer" über JSON dasselbe Wort haben könnten und es nicht
// tun.** Ein Verbraucher, der über die Antwort liefe, bekäme bei `null` einen
// Fehler statt einer leeren Anzeige — und zwar erst bei der ersten Partie
// ohne Zug, also lange nach der Abnahme. Dieselbe Zusage steht an jeder
// anderen Liste dieses Dienstes (G5, #1462).
func TestEinLeererVerlaufGehtAlsListeHinaus(t *testing.T) {
	t.Parallel()

	jetzt := func() time.Time {
		return time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC)
	}

	dienst := mitHaushalt(t, rest.NeuerDienst(probemarke, jetzt,
		rest.Naehte{Partienverlauf: schweigsamerVerlauf{}}),
		domain.Vorgabegrenzen(), jetzt)

	antwort := ruf(t, dienst,
		endpunkt{methode: http.MethodGet, weg: "/partien/p1/verlauf?fuer=a"},
		gueltig)

	switch {
	case antwort.Code != http.StatusOK:
		t.Fatalf("endete mit %d: %s", antwort.Code, antwort.Body.String())
	case strings.TrimSpace(antwort.Body.String()) != "[]":
		t.Errorf("hinaus ging %q, erwartet []", antwort.Body.String())
	}
}

// schweigsamerVerlauf gibt `nil` — das, was ein Speicher ohne Zeilen liefert.
type schweigsamerVerlauf struct{}

func (schweigsamerVerlauf) PartieVerlauf(context.Context, string, string) (
	[]rest.Partieereignis, error,
) {
	return nil, nil
}
