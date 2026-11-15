package domain_test

// **Gepinnte Drahtvektoren** — D05
// (#2236).
//
// Zwei Fassungen dieses Programms sprechen über eine Leitung miteinander, und
// keine weiß, ob die andere ein Feld umbenannt hat. Die Fuzzziele sagen, dass
// Lesen und Schreiben einander umkehren — aber nicht, dass das Geschriebene
// **heute dasselbe ist wie gestern**. Hier steht je Nachrichtenfamilie das
// Byte für Byte, was diese Fassung sendet, in `testdata/draht/`:
//
//  1. Ein kanonischer Wert schreibt sich **genau** zu seinem Vektor.
//  2. Der Vektor liest sich zurück, ohne Befund.
//  3. Jeder negative Vektor unter `testdata/draht/negativ/` wird mit seiner
//     Klasse abgewiesen — andere Fassung, fremde Art, Überlänge, fehlendes
//     Pflichtfeld.
//
// **Ein geänderter Vektor ist ein Bruch des Drahtformats.** Wer ihn neu
// schreibt (`DRAHT_SCHREIBEN=1`), hebt die Fassung der Familie und trägt sie
// in `docs/FASSUNGEN.md` ein; der Vertrag beider Enden steht in
// `docs/DRAHTFORMAT.md`.
//
// **Was er nicht hält**: wie eine Nachricht über die Leitung kommt (Rahmen,
// Fristen, Zuteilung — `internal/leitung`) und ob eine Wiederholung wirkt
// (`Empfaengerstand`, Kennungen der Spielbefehle, Briefkasten des Handels).

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// drahtfamilie ist eine Nachrichtenart mit ihrem kanonischen Wert.
type drahtfamilie struct {
	name    string
	schreib func() ([]byte, error)
	lies    func([]byte) error
}

func drahtfamilien() []drahtfamilie {
	return []drahtfamilie{
		{
			name: "einloesung",
			schreib: func() ([]byte, error) {
				return domain.Einloesung{Nonce: eineNonce(), Adresse: "192.0.2.10:41234", Name: "Ada"}.Kodiere()
			},
			lies: func(roh []byte) error {
				gelesen, fehler := domain.ZerlegeEinloesung(roh)
				if fehler == nil && (gelesen.Nonce != eineNonce() || gelesen.Name != "Ada") {
					return errors.New("anders gelesen als geschrieben")
				}

				return fehler
			},
		},
		{
			name: "einloeseantwort",
			schreib: func() ([]byte, error) {
				return domain.AntwortAuf(domain.ErrEinladungVerbraucht).Kodiere(), nil
			},
			lies: func(roh []byte) error {
				gelesen, fehler := domain.ZerlegeEinloeseantwort(roh)
				if fehler == nil && gelesen.Ergebnis != domain.EinloesungVerbraucht {
					return errors.New("anders gelesen als geschrieben")
				}

				return fehler
			},
		},
		{
			name:    "einladung",
			schreib: func() ([]byte, error) { return []byte(domain.NeueEinladung("192.0.2.10:41234")), nil },
			lies: func(roh []byte) error {
				adresse, fehler := domain.ZerlegeEinladung(domain.Einladung(roh))
				if fehler == nil && adresse != "192.0.2.10:41234" {
					return errors.New("anders gelesen als geschrieben")
				}

				return fehler
			},
		},
		{
			name: "kontrolle-rueckkehr",
			schreib: func() ([]byte, error) {
				return domain.VerpackeKontrolle(domain.KontrolleRueckkehr, map[string]any{"epoche": 2, "folge": 7})
			},
			lies: func(roh []byte) error {
				_, _, fehler := domain.LiesKontrolle(roh, "gast", domain.KontrolleRueckkehr)

				return fehler
			},
		},
		{
			name: "partienachricht",
			schreib: func() ([]byte, error) {
				return json.Marshal(domain.Partienachricht{
					Art: domain.ArtEreignis, Partie: "p-2236", Epoche: 1, Folge: 3, Von: "bada",
					Inhalt: json.RawMessage(`{"art":"muenzwurf"}`), Kette: strings.Repeat("ab", 32),
				})
			},
			lies: func(roh []byte) error {
				var gelesen domain.Partienachricht
				if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
					return fehler
				}

				return gelesen.Pruefe()
			},
		},
		{
			// **Die Form, die unterschrieben wird**, und nicht nur der
			// Umschlag: eine Änderung an ihr bräche jede Unterschrift der
			// Gegenseite, auch bei unverändertem JSON.
			name: "handelsumschlag-unterschriftsform",
			schreib: func() ([]byte, error) {
				return handelsumschlag().Unterschriftsform(), nil
			},
			lies: func(roh []byte) error {
				if !bytes.Equal(roh, handelsumschlag().Unterschriftsform()) {
					return errors.New("die Unterschriftsform ist eine andere")
				}

				return nil
			},
		},
		{
			name: "handelsumschlag",
			schreib: func() ([]byte, error) {
				return json.Marshal(handelsumschlag())
			},
			lies: func(roh []byte) error {
				var gelesen domain.Handelsumschlag
				if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
					return fehler
				}

				if befunde := gelesen.Validate(); len(befunde) > 0 {
					return befunde[0]
				}

				return nil
			},
		},
	}
}

// handelsumschlag ist ein vollständiger, gültiger Umschlag.
func handelsumschlag() domain.Handelsumschlag {
	return domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung, Art: domain.NachrichtVorschlag,
		NachrichtID: "n-2236", Vorgang: "v-2236", Basisrevision: 2, Abdruck: "abdruck-2236",
		Folge: 5, Absender: "bada", Unterschrift: "unterschrift",
		Gesendet: time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC),
	}
}

func TestJedeNachrichtSchreibtSichZuIhremGepinntenVektor(t *testing.T) {
	t.Parallel()

	for _, familie := range drahtfamilien() {
		t.Run(familie.name, func(t *testing.T) {
			t.Parallel()

			geschrieben, fehler := familie.schreib()
			if fehler != nil {
				t.Fatalf("schreiben: %v", fehler)
			}

			pfad := filepath.Join("testdata", "draht", familie.name+".vektor")
			if os.Getenv("DRAHT_SCHREIBEN") == "1" {
				if schreibfehler := os.WriteFile(pfad, geschrieben, 0o600); schreibfehler != nil {
					t.Fatalf("den Vektor schreiben: %v", schreibfehler)
				}
			}

			gepinnt, lesefehler := os.ReadFile(pfad) //nolint:gosec // ein Vektor dieses Pakets
			if lesefehler != nil {
				t.Fatalf("der Vektor fehlt (%v) — neu nur mit DRAHT_SCHREIBEN=1 und gehobener Fassung", lesefehler)
			}

			if !bytes.Equal(geschrieben, gepinnt) {
				t.Errorf("das Drahtformat hat sich geändert:\ngeschrieben %q\ngepinnt     %q", geschrieben, gepinnt)
			}

			if fehler := familie.lies(gepinnt); fehler != nil {
				t.Errorf("der gepinnte Vektor liest sich nicht: %v", fehler)
			}
		})
	}
}

func TestJederNegativeVektorWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	familien := map[string]drahtfamilie{}
	for _, familie := range drahtfamilien() {
		familien[familie.name] = familie
	}

	eintraege, fehler := os.ReadDir(filepath.Join("testdata", "draht", "negativ"))
	if fehler != nil {
		t.Fatalf("die negativen Vektoren: %v", fehler)
	}

	if len(eintraege) == 0 {
		t.Fatal("kein negativer Vektor — der Lauf prüfte nichts")
	}

	for _, eintrag := range eintraege {
		// **Name ist Familie, zwei Bindestriche, Fall**: `einloesung--andere-fassung.vektor`.
		name, fall, gut := strings.Cut(strings.TrimSuffix(eintrag.Name(), ".vektor"), "--")
		familie, bekannt := familien[name]

		if !gut || !bekannt {
			t.Errorf("%s: keine Familie dieses Laufs", eintrag.Name())

			continue
		}

		t.Run(name+"/"+fall, func(t *testing.T) {
			roh, lesefehler := os.ReadFile(filepath.Join("testdata", "draht", "negativ", eintrag.Name())) //nolint:gosec // ein Vektor dieses Pakets
			if lesefehler != nil {
				t.Fatalf("lesen: %v", lesefehler)
			}

			if fehler := familie.lies(roh); fehler == nil {
				t.Errorf("angenommen: %q", roh)
			}
		})
	}
}
