package domain_test

// **Die Fehlerwege der Handlungen mit Schlüsselwort** — Werkstück M7 der
// Regelmaschine
// (#1609).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was eine Handlung tut**; hier steht, **woran sie
// scheitert** — und `schluesselhandlungen_test.go` stand mit beidem bei 459
// Zeilen (§2).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg, auf dem eine Handlung nicht geht, sagt warum** — §8.
func TestJederFehlerwegDerHandlungenSagtWarum(t *testing.T) {
	t.Parallel()

	partie := einHandlungstisch(t)

	faelle := map[string]struct {
		tun  func(p domain.Partie) (domain.Partie, error)
		wort string
	}{
		"zerstören von der Hand": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Zerstoeren("sitz-1/hand/0")
			},
			wort: "CR 701.8a",
		},
		"zerstören, was es nicht gibt": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Zerstoeren("gibtsnicht")
			},
			wort: "gibt es in dieser Partie nicht",
		},
		"opfern von der Hand": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Opfern("sitz-1", "sitz-1/hand/0")
			},
			wort: "CR 701.21a",
		},
		"exilieren, was es nicht gibt": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Exilieren("gibtsnicht")
			},
			wort: "gibt es in dieser Partie nicht",
		},
		"neutralisieren, was es nicht gibt": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Neutralisieren("gibtsnicht")
			},
			wort: "gibt es in dieser Partie nicht",
		},
		"minus eins mahlen": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Mahlen("sitz-1", -1)
			},
			wort: "CR 701.17a",
		},
		"für einen Sitz mahlen, den es nicht gibt": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Mahlen("gibtsnicht", 1)
			},
			wort: "gibt es in dieser Partie nicht",
		},
		"eine Karte auf der Hand kämpfen lassen": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Kaempfen("sitz-1/hand/0", "sitz-1/baer/0")
			},
			wort: "CR 701.14a",
		},
		"ein Objekt ohne Karte kämpfen lassen": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Kaempfen("sitz-1/stein/0", "sitz-1/baer/0")
			},
			wort: "ist keine Kreatur",
		},
		"der zweite Kämpfer liegt falsch": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Kaempfen("sitz-1/baer/0", "sitz-1/hand/0")
			},
			wort: "CR 701.14a",
		},
		"minus eins hellsehen": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Hellsehen("sitz-1", -1, nil)
			},
			wort: "ansehen geht nicht",
		},
		"für einen Sitz hellsehen, den es nicht gibt": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Hellsehen("gibtsnicht", 1, nil)
			},
			wort: "gibt es in dieser Partie nicht",
		},
		"überwachen, was nicht obenauf liegt": {
			tun: func(p domain.Partie) (domain.Partie, error) {
				return p.Ueberwachen("sitz-1", 1, []string{"sitz-1/bib/d"})
			},
			wort: "CR 701.22a",
		},
	}

	// **Ein Spielstein ohne Karte** — er hat keine Typzeile (CR 111).
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/stein/0", Art: domain.ObjektSpielstein,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch,
	})

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := fall.tun(partie)

			if fehler == nil {
				t.Fatal("es kam kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %q", fehler)
			}
		})
	}
}

// **Aus einer leeren Bibliothek sieht niemand etwas** — CR 701.22a.
//
// **Und das ist kein Fehler**: es gibt schlicht nichts obenauf. Ob daraus ein
// Verlust wird, entscheidet CR 704.5b beim nächsten Ziehen.
func TestAusEinerLeerenBibliothekSiehtNiemandEtwas(t *testing.T) {
	t.Parallel()

	partie := einHandlungstisch(t)

	behalten := make([]domain.Spielobjekt, 0, len(partie.Objekte))

	for _, objekt := range partie.Objekte {
		if objekt.Zone != domain.ZoneBibliothek {
			behalten = append(behalten, objekt)
		}
	}

	partie.Objekte = behalten

	if oben := partie.Obenauf("sitz-1", 3); len(oben) != 0 {
		t.Errorf("obenauf liegen %v", oben)
	}

	// **Und hellsehen geht trotzdem** — es sortiert nichts um.
	if _, fehler := partie.Hellsehen("sitz-1", 3, nil); fehler != nil {
		t.Errorf("hellsehen auf einer leeren Bibliothek: %v", fehler)
	}
}
