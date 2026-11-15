// **Die beiden Sicherungswerkzeuge** —
// #1169.
//
// # Was hier geprüft wird
//
//   - Beide kommen durch — mit ihren Argumenten, in ihrer Gestalt.
//   - Beide weisen eine leere Angabe ab, **bevor** sie die Naht rufen.
//   - Ein Fehler der Naht kommt an und wird nicht verschluckt.
//   - **Und es gibt sie nicht ohne Naht**: `nil` heißt kein Werkzeug, so wie
//     bei jeder anderen Fläche, die eine Naht nicht einrichtet.
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// sicherungsprobe ist eine Naht, die mitschreibt.
type sicherungsprobe struct {
	gesichert string
	geschaut  string

	// klartext ist, was der letzte Ruf dazu sagte (OPS-06).
	klartext bool

	// verschluesselt ist, was der Stand melden soll.
	verschluesselt bool

	fehler error
}

func (s *sicherungsprobe) Verschluesselungsstand() (bool, error) {
	return s.verschluesselt, s.fehler
}

func (s *sicherungsprobe) Sicherungsstand(context.Context) (any, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return map[string]any{"stufe": "alt", "tage": 16}, nil
}

func (s *sicherungsprobe) BestandSichern(_ context.Context, ziel string, klartext bool) error {
	s.gesichert, s.klartext = ziel, klartext

	return s.fehler
}

func (s *sicherungsprobe) SicherungSchauen(_ context.Context, pfad string) (
	mcp.Sicherungsangaben, error,
) {
	s.geschaut = pfad

	if s.fehler != nil {
		return mcp.Sicherungsangaben{}, s.fehler
	}

	return mcp.Sicherungsangaben{
		AngelegtAm: "2026-08-28T00:00:00.000Z", Schemastand: "0021_listenspalten.sql",
		Sets: 1, Karten: 3, Bilder: "geleert",
	}, nil
}

// rufeSicherung sucht ein Werkzeug und ruft es.
func rufeSicherung(
	t *testing.T, naht mcp.Bestandssicherung, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Sicherung: naht}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q steht nicht in der Liste", name)

	return "", nil
}

func TestDieSicherungswerkzeugeKommenDurch(t *testing.T) {
	naht := &sicherungsprobe{}

	aus, err := rufeSicherung(t, naht, "schreiben.bestand.sichern",
		`{"ziel":"/tmp/x.msesicherung"}`)
	if err != nil {
		t.Fatalf("sichern: %v", err)
	}

	if naht.gesichert != "/tmp/x.msesicherung" || naht.klartext {
		t.Errorf("gesichert wurde nach %q, im Klartext %v", naht.gesichert, naht.klartext)
	}

	// **Der Klartext ist ausdrücklich** — und kommt an (OPS-06.AK-02).
	if _, klarfehler := rufeSicherung(t, naht, "schreiben.bestand.sichern",
		`{"ziel":"/tmp/y.msesicherung","klartext":true}`); klarfehler != nil || !naht.klartext {
		t.Errorf("der Klartextwunsch kam nicht an: %v, %v", klarfehler, naht.klartext)
	}

	if !strings.Contains(aus, "/tmp/x.msesicherung") {
		t.Errorf("die Antwort nennt das Ziel nicht: %s", aus)
	}

	aus, err = rufeSicherung(t, naht, "lesen.sicherung", `{"pfad":"/tmp/x.msesicherung"}`)
	if err != nil {
		t.Fatalf("schauen: %v", err)
	}

	if naht.geschaut != "/tmp/x.msesicherung" {
		t.Errorf("geschaut wurde in %q", naht.geschaut)
	}

	// **Eingerückt und nicht am Stück** — `alsJSON` schreibt lesbar, weil die
	// Antwort an ein Sprachmodell geht.
	for _, erwartet := range []string{`"sets": 1`, `"karten": 3`, `"bilder": "geleert"`} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("%s fehlt in %s", erwartet, aus)
		}
	}
}

// **Eine leere Angabe wird abgewiesen, bevor die Naht gerufen wird.**
//
// Die zweite Hälfte ist die, die zählt: ein Werkzeug, das erst ruft und dann
// prüft, hätte den Pfad schon angefasst.
func TestEineLeereAngabeErreichtDieSicherungsnahtNicht(t *testing.T) {
	for name, fall := range map[string]struct {
		werkzeug, argumente string
	}{
		"sichern ohne Ziel": {"schreiben.bestand.sichern", `{"ziel":""}`},
		"schauen ohne Pfad": {"lesen.sicherung", `{"pfad":""}`},
	} {
		t.Run(name, func(t *testing.T) {
			naht := &sicherungsprobe{}

			if _, err := rufeSicherung(t, naht, fall.werkzeug, fall.argumente); err == nil {
				t.Fatal("die leere Angabe kam durch")
			} else if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("der Fehler ist keine Validierung: %v", err)
			}

			if naht.gesichert != "" || naht.geschaut != "" {
				t.Error("die Naht wurde trotzdem gerufen")
			}
		})
	}
}

// **Fremde Argumente und ein Fehler der Naht kommen an.**
func TestDieSicherungswerkzeugeReichenIhreFehlerDurch(t *testing.T) {
	naht := &sicherungsprobe{fehler: errors.New("die Platte ist voll")}

	for _, fall := range []struct{ werkzeug, argumente string }{
		{"schreiben.bestand.sichern", `{"ziel":"/tmp/x.msesicherung"}`},
		{"lesen.sicherung", `{"pfad":"/tmp/x.msesicherung"}`},
	} {
		if _, err := rufeSicherung(t, naht, fall.werkzeug, fall.argumente); err == nil {
			t.Errorf("%s verschluckte den Fehler der Naht", fall.werkzeug)
		}
	}

	// **Und kein JSON ist auch ein Fehler.**
	for _, fall := range []struct{ werkzeug, argumente string }{
		{"schreiben.bestand.sichern", `kein JSON`},
		{"lesen.sicherung", `kein JSON`},
	} {
		if _, err := rufeSicherung(t, naht, fall.werkzeug, fall.argumente); err == nil {
			t.Errorf("%s nahm etwas an, das kein JSON ist", fall.werkzeug)
		}
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
//
// Dieselbe Zusage wie bei jeder anderen Fläche, die eine Naht nicht
// einrichtet — und die Gegenprobe dazu ist der Lauf darüber, der sie findet.
func TestOhneSicherungsnahtGibtEsDieWerkzeugeNicht(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "schreiben.bestand.sichern" || werkzeug.Name == "lesen.sicherung" {
			t.Errorf("%s steht in der Liste, obwohl keine Naht eingerichtet ist", werkzeug.Name)
		}
	}
}

// **Der Verschlüsselungsstand kommt durch — und ein Fehler ebenso.**
//
// **Beide Wege**, weil der zweite der ist, an dem ein Mensch etwas tut: kommt
// die Auskunft „der Schlüssel fehlt" hier nicht an, sieht das Modell einen
// leeren Erfolg (§8).
func TestDerVerschluesselungsstandKommtDurch(t *testing.T) {
	for name, fall := range map[string]struct {
		probe    *sicherungsprobe
		erwartet string
		scheitet bool
	}{
		"verschlüsselt": {
			probe: &sicherungsprobe{verschluesselt: true}, erwartet: `"verschluesselt": true`,
		},
		"im Klartext": {
			probe: &sicherungsprobe{}, erwartet: `"verschluesselt": false`,
		},
		"ein Fehler": {
			probe: &sicherungsprobe{fehler: errors.New("der Schlüssel fehlt")}, scheitet: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			aus, err := rufeSicherung(t, fall.probe,
				"lesen.bestand.verschluesselung", `{}`)

			if fall.scheitet {
				if err == nil {
					t.Fatalf("kein Fehler, herausgekommen ist %q", aus)
				}

				return
			}

			if err != nil {
				t.Fatalf("rufen: %v", err)
			}

			if !strings.Contains(aus, fall.erwartet) {
				t.Errorf("die Antwort ist %q, erwartet war %q", aus, fall.erwartet)
			}
		})
	}
}

// **Der Stand der Sicherung kommt durch, wie die Naht ihn gibt** — OPS01
// (#2347): die Gestalt kommt aus `internal/app`, und dieses Paket reicht sie
// weiter, statt eine Abschrift zu führen.
func TestDerSicherungsstandKommtDurch(t *testing.T) {
	aus, err := rufeSicherung(t, &sicherungsprobe{}, "lesen.bestand.sicherungsstand", `{}`)
	if err != nil || !strings.Contains(aus, `"stufe": "alt"`) || !strings.Contains(aus, `"tage": 16`) {
		t.Errorf("die Antwort ist %q (%v)", aus, err)
	}

	if aus, err := rufeSicherung(t, &sicherungsprobe{fehler: domain.ErrBelegt}, "lesen.bestand.sicherungsstand", `{}`); err == nil {
		t.Errorf("kein Fehler, herausgekommen ist %q", aus)
	}
}
