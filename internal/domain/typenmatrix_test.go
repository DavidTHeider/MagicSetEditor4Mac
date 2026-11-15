package domain_test

// **Was das Grundprofil seit M2 begleitet** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Warum `begleitet` und nicht `automatisch`
//
// **Weil das Programm die Typzeile liest und nichts durchsetzt.** Es sagt,
// welches Wort CR 205.2a nicht kennt — das ist eine **Warnung**, und dafür
// steht [domain.UnterstuetzungBegleitet]. Dass nur Länder gespielt werden
// dürfen, wann ein Zauber eine Kreatur wird und welcher Untertyp zu welchem
// Typ gehört, prüft niemand; wer `automatisch` einträge, behauptete es.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Typzeile wird begleitet, nicht durchgesetzt.**
func TestDieTypzeileWirdBegleitet(t *testing.T) {
	t.Parallel()

	matrix := domain.Grundprofil("magic").Matrix

	if grad := matrix.Grad(domain.BereichTypen); grad != domain.UnterstuetzungBegleitet {
		t.Fatalf("der Grad ist %q", grad)
	}
}

// **Und sie nennt Quelle und Grenzen** — ab `begleitet` ist beides Pflicht.
func TestDieTypzeileNenntQuelleUndGrenzen(t *testing.T) {
	t.Parallel()

	for _, eintrag := range domain.Grundprofil("magic").Matrix.Eintraege {
		if eintrag.Bereich != domain.BereichTypen {
			continue
		}

		switch {
		case len(eintrag.Quelle.Nummern) == 0:
			t.Error("die Quelle nennt keine Nummer")
		case eintrag.Quelle.Fassung != domain.Grundprofilquelle:
			t.Errorf("die Fassung ist %q", eintrag.Quelle.Fassung)
		case len(eintrag.Grenzen) == 0:
			t.Error("der Eintrag nennt keine Grenze")
		}

		for _, grenze := range eintrag.Grenzen {
			if !strings.Contains(grenze, "CR ") {
				t.Errorf("die Grenze %q nennt keine Regel", grenze)
			}
		}

		return
	}

	t.Fatal("die Matrix führt den Bereich nicht")
}
