package domain_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gleichAuf sagt, ob zwei Maße bis auf einen zehntel Millimeter gleich sind.
func gleichAuf(a, b float64) bool { return math.Abs(a-b) < 0.1 }

func TestBlattZuKenntDieZweiFormate(t *testing.T) {
	t.Parallel()

	a4, gibtEs := domain.BlattZu(domain.BlattA4)
	if !gibtEs || !gleichAuf(a4.Breite, 210) || !gleichAuf(a4.Hoehe, 297) {
		t.Errorf("A4 ist %v (%t)", a4, gibtEs)
	}

	letter, gibtEs := domain.BlattZu(domain.BlattLetter)
	if !gibtEs || !gleichAuf(letter.Breite, 215.9) || !gleichAuf(letter.Hoehe, 279.4) {
		t.Errorf("Letter ist %v (%t)", letter, gibtEs)
	}

	// **Ein unbekanntes Format ist kein Blatt** — und keine stillschweigende
	// Vorgabe: wer „A3" tippt, bekäme sonst A4 und merkte es am Ausdruck.
	if _, gibtEs := domain.BlattZu("A3"); gibtEs {
		t.Error("A3 gibt es nicht, wird aber geführt")
	}
}

func TestPlatzAufSeiteLegtMittigUndVerhaeltnistreu(t *testing.T) {
	t.Parallel()

	blatt := domain.Blatt{Breite: 210, Hoehe: 297}

	// **Ein breites Bild** — es stößt an die Breite und bleibt in der Höhe
	// frei.
	platz := domain.PlatzAufSeite(2000, 1000, blatt, 10)

	if !gleichAuf(platz.Breite, 190) {
		t.Errorf("Breite %v, erwartet 190", platz.Breite)
	}

	if !gleichAuf(platz.Hoehe, 95) {
		t.Errorf("Höhe %v, erwartet 95", platz.Hoehe)
	}

	if !gleichAuf(platz.Links, 10) {
		t.Errorf("links %v, erwartet 10", platz.Links)
	}

	// **Mittig in der Höhe**: (297 − 95) / 2.
	if !gleichAuf(platz.Unten, 101) {
		t.Errorf("unten %v, erwartet 101", platz.Unten)
	}
}

// **Ein hohes Bild stößt an die Höhe** — dieselbe Rechnung, andere Kante.
func TestPlatzAufSeiteStoesstAnDieHoehe(t *testing.T) {
	t.Parallel()

	blatt := domain.Blatt{Breite: 210, Hoehe: 297}

	platz := domain.PlatzAufSeite(1000, 2000, blatt, 10)

	if !gleichAuf(platz.Hoehe, 277) {
		t.Errorf("Höhe %v, erwartet 277", platz.Hoehe)
	}

	if !gleichAuf(platz.Breite, 138.5) {
		t.Errorf("Breite %v, erwartet 138,5", platz.Breite)
	}
}

// **Ein Blatt ohne Platz gibt keinen** — der Rand kann größer sein als das
// Papier, und ein negatives Rechteck wäre ein PDF, das niemand öffnet.
func TestPlatzAufSeiteOhneRaum(t *testing.T) {
	t.Parallel()

	platz := domain.PlatzAufSeite(100, 100, domain.Blatt{Breite: 20, Hoehe: 20}, 10)

	if platz.Breite != 0 || platz.Hoehe != 0 {
		t.Errorf("Platz %v, erwartet keinen", platz)
	}
}

// **Und ein Bild ohne Fläche ebenso** — es entstünde bei einem leeren
// Ausschnitt, und eine Division durch null gäbe keinen Platz, sondern NaN.
func TestPlatzAufSeiteOhneBild(t *testing.T) {
	t.Parallel()

	platz := domain.PlatzAufSeite(0, 100, domain.Blatt{Breite: 210, Hoehe: 297}, 10)

	if platz.Breite != 0 || platz.Hoehe != 0 {
		t.Errorf("Platz %v, erwartet keinen", platz)
	}
}

// **Das Blatt dreht sich nach dem Bild** — ein breiter Tisch auf ein
// stehendes A4 gelegt füllte ein Drittel der Seite.
func TestBlattGedrehtFolgtDemBild(t *testing.T) {
	t.Parallel()

	hoch := domain.Blatt{Breite: 210, Hoehe: 297}

	quer := domain.BlattFuer(2000, 1000, hoch)
	if !gleichAuf(quer.Breite, 297) || !gleichAuf(quer.Hoehe, 210) {
		t.Errorf("quer ist %v", quer)
	}

	stehend := domain.BlattFuer(1000, 2000, hoch)
	if !gleichAuf(stehend.Breite, 210) || !gleichAuf(stehend.Hoehe, 297) {
		t.Errorf("stehend ist %v", stehend)
	}

	// **Ein Bild ohne Fläche dreht nichts** — es gibt kein Verhältnis.
	if unveraendert := domain.BlattFuer(0, 0, hoch); unveraendert != hoch {
		t.Errorf("ohne Bild ist das Blatt %v", unveraendert)
	}
}
