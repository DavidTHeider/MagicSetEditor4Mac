package service_test

// **Ein Abschluss gilt im Profil dessen, der fragt** — Befund 4 des Nachlaufs
// über Phase 6.1.
//
// Die Kennung eines Postens gibt der Aufrufer vor, und in zwei Profilen kann
// es dieselbe sein (`probe/<anbieter>/<alias>`). Die Ablage schließt deshalb
// im Profil ab — und welches das ist, sagt nicht die Kennung, sondern der
// Akteur, genau wie beim Reservieren.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinAbschlussGiltImProfilDesAkteurs(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{posten: domain.Kostenposten{Kennung: "probe/a/b"}}
	weg := kostenwache(doppel, wache())

	for name, ruf := range map[string]func() error{
		"abschließen": func() error {
			_, fehler := weg.Abschliessen(t.Context(), "probe/a/b", domain.Betrag{Wert: 1, Waehrung: "EUR"})

			return fehler
		},
		"freigeben": func() error { _, fehler := weg.Freigeben(t.Context(), "probe/a/b"); return fehler },
		"unklar": func() error {
			_, fehler := weg.AlsUnklarMerken(t.Context(), "probe/a/b", "die Frist lief ab")

			return fehler
		},
	} {
		doppel.abschlussprofil = 0

		if fehler := ruf(); fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		if doppel.abschlussprofil != 3 {
			t.Errorf("%s ging an Profil %d, erwartet das des Akteurs (3)", name, doppel.abschlussprofil)
		}
	}
}

// **Ohne Akteur wird nichts abgeschlossen.**
func TestOhneAkteurWirdNichtsAbgeschlossen(t *testing.T) {
	t.Parallel()

	unbekannt := errors.New("wer arbeitet, ist nicht zu sagen")
	doppel := &kostendoppel{}
	wacht := wache()
	wacht.fehler = unbekannt
	weg := kostenwache(doppel, wacht)

	_, abschluss := weg.Abschliessen(t.Context(), "k", domain.Betrag{Wert: 1, Waehrung: "EUR"})
	_, frei := weg.Freigeben(t.Context(), "k")
	_, unklar := weg.AlsUnklarMerken(t.Context(), "k", "g")

	for name, fehler := range map[string]error{"abschließen": abschluss, "freigeben": frei, "unklar": unklar} {
		if !errors.Is(fehler, unbekannt) {
			t.Errorf("%s: %v, erwartet den Fehler der Wache", name, fehler)
		}
	}

	if doppel.abschlussprofil != 0 {
		t.Errorf("es wurde abgeschlossen, in Profil %d", doppel.abschlussprofil)
	}
}
