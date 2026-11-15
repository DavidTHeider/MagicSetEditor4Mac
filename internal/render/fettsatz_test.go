package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Fett ist ein eigener Schnitt, und er wird gemessen wie einer.**
//
// Kursiv nennt der Stil unter `italic name:`; fett ist bei MSE ein **Gewicht**,
// das an das Betriebssystem geht. Diese Suche macht daraus wieder einen Namen —
// `<Familie> Bold`
// ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
//
// Geprüft wird an der **Breite**: der fette Schnitt ist hier doppelt so breit,
// und ein Abschnitt, der ihn nicht bekäme, wäre so breit wie der aufrechte.
func TestEinFetterAbschnittWirdMitDemFettenSchnittGemessen(t *testing.T) {
	satz := render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben}

	gesetzt := render.SetzeAbschnitte(satz, render.Rechteck{Breite: 500, Hoehe: 30},
		[]render.Abschnitt{
			{Text: "ab"},
			{Text: "cd", Fett: true},
		},
		render.Schriftsatz{
			Aufrecht: gleichbreit{breite: 0.5},
			Fett:     gleichbreit{breite: 1.0},
		}, nil)

	if len(gesetzt.Zeilen) != 1 || len(gesetzt.Zeilen[0].Stuecke) != 2 {
		t.Fatalf("gesetzt wurde %+v", gesetzt.Zeilen)
	}

	stuecke := gesetzt.Zeilen[0].Stuecke

	if stuecke[0].Breite != 10 {
		t.Errorf("der aufrechte Abschnitt ist %g breit, erwartet 10", stuecke[0].Breite)
	}

	if stuecke[1].Breite != 20 {
		t.Errorf("der fette Abschnitt ist %g breit, erwartet 20 — "+
			"gemessen wurde mit dem aufrechten Schnitt", stuecke[1].Breite)
	}
}

// **Ohne fetten Schnitt wird der aufrechte genommen — und nicht nichts.**
//
// Dieselbe Bestenfalls-Arbeit wie bei der Kursiven: ein aufrechter Text ist
// eine Karte, ein fehlender ist keine (§8).
func TestOhneFettenSchnittWirdDerAufrechteGenommen(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 500, Hoehe: 30},
		[]render.Abschnitt{{Text: "ab", Fett: true}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if len(gesetzt.Zeilen) != 1 || len(gesetzt.Zeilen[0].Stuecke) != 1 {
		t.Fatalf("gesetzt wurde %+v", gesetzt.Zeilen)
	}

	if breite := gesetzt.Zeilen[0].Stuecke[0].Breite; breite != 10 {
		t.Errorf("%g breit, erwartet 10 — der aufrechte Schnitt wurde nicht genommen", breite)
	}
}

// **Kursiv und fett zusammen sind ein dritter Schnitt.**
//
// MSE nimmt dafür die Schrift aus `italic name:` **und** das fette Gewicht
// (`src/data/font.cpp`, Zeilen 100–101, Fassung f9d9356 vom 2021-07-06) — bei
// uns also `<Kursivname> Bold`.
func TestKursivUndFettSindEinDritterSchnitt(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 500, Hoehe: 30},
		[]render.Abschnitt{{Text: "ab", Kursiv: true, Fett: true}},
		render.Schriftsatz{
			Aufrecht:   gleichbreit{breite: 0.5},
			Kursiv:     gleichbreit{breite: 1.0},
			Fett:       gleichbreit{breite: 2.0},
			KursivFett: gleichbreit{breite: 4.0},
		}, nil)

	if breite := gesetzt.Zeilen[0].Stuecke[0].Breite; breite != 80 {
		t.Errorf("%g breit, erwartet 80 — der kursiv-fette Schnitt wurde nicht genommen", breite)
	}
}
