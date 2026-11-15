package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Vorgabezustand einer Stiloption reicht bis ins Fenster**
// (#764).
//
// Zwei Angaben, und beide braucht die Oberfläche für **einen** Eintrag:
// `Vorgabe` sagt, **ob** es ihn gibt und welche Farbe er zeigt — MSE zeichnet
// den Tupfer aus `default_color` und blendet den Eintrag ohne `default_script`
// ganz aus (`src/gui/value/color.cpp`, Zeilen 38 und 107–112, Fassung f9d9356
// vom 2021-07-06). `Gesetzt` sagt, ob er **gewählt** ist.
func TestDieStiloptionReichtIhrenVorgabezustandDurch(t *testing.T) {
	t.Parallel()

	waehler := &stilwaehler{optionen: []service.Stiloption{{
		StyleField: domain.StyleField{Name: "randfarbe", Kind: domain.KindColor},
		Wert:       "black",
		Vorgabe:    "black",
		Gesetzt:    false,
	}, {
		StyleField: domain.StyleField{Name: "hintergrund", Kind: domain.KindChoice},
		Wert:       "dunkel",
		Gesetzt:    true,
	}}}

	antwort := mitStilen(waehler).Stiloptionen(probekontext(t), "1")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Optionen) != 2 {
		t.Fatalf("%d Optionen", len(antwort.Optionen))
	}

	if antwort.Optionen[0].Vorgabe != "black" || antwort.Optionen[0].Gesetzt {
		t.Errorf("die Farboption: %+v", antwort.Optionen[0])
	}

	if antwort.Optionen[1].Vorgabe != "" || !antwort.Optionen[1].Gesetzt {
		t.Errorf("die Auswahl: %+v", antwort.Optionen[1])
	}
}

// **Und der Weg zurück auf die Vorgabe steht als eigene Bindung.**
//
// **Nicht als `StiloptionSetzen` mit leerem Wert**: leer ist bei einer
// Textoption ein gültiger Wert, und zwei Bedeutungen an einem Aufruf wären
// später nicht mehr auseinanderzuhalten.
func TestEineStiloptionLaesstSichUeberDieBindungZuruecksetzen(t *testing.T) {
	t.Parallel()

	waehler := &stilwaehler{}

	if befund := mitStilen(waehler).
		StiloptionZuruecksetzen(probekontext(t), "1", "randfarbe").Fehler; befund != nil {
		t.Fatalf("Befund: %+v", befund)
	}

	if waehler.zurueckgesetzt != "randfarbe" {
		t.Errorf("zurückgesetzt wurde %q", waehler.zurueckgesetzt)
	}
}

// **Eine unlesbare Kennung wird abgewiesen** — wie an jeder anderen Bindung.
func TestZuruecksetzenWeistEineUnlesbareKennungAb(t *testing.T) {
	t.Parallel()

	if mitStilen(&stilwaehler{}).
		StiloptionZuruecksetzen(probekontext(t), "keine Zahl", "randfarbe").Fehler == nil {
		t.Error("die Kennung wurde übergangen")
	}
}
