package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Symmetrie vervielfältigt ihre Kinder** —
// #890,
// Werkstück 3 von Phase X.
//
//	p' = ((p - d) * rot + d) * m + o
//
// `src/render/symbol/viewer.cpp`, Zeile 174, Fassung f9d9356 vom 2021-07-06.
// `d` ist die Mitte der Symmetrie — gedreht wird **um sie**, nicht um den
// Ursprung.

// **Die erste Kopie ist das Original.**
func TestDieErsteKopieIstDasOriginal(t *testing.T) {
	t.Parallel()

	spiegel := msepaket.Symbolteil{
		Art: msepaket.TeilSymmetrie, Spiegelart: msepaket.SpiegelungDrehung,
		Kopien: 4, Mitte: msepaket.Vektor{X: 0.5, Y: 0.5},
		Griff: msepaket.Vektor{X: 0, Y: -0.5},
	}

	punkt := msepaket.Vektor{X: 0.9, Y: 0.5}

	if hat := service.SymmetrieAbbilden(spiegel, 0, punkt); !nahe(hat, punkt) {
		t.Errorf("die erste Kopie bildet %+v auf %+v ab", punkt, hat)
	}
}

// **Vier Drehkopien liegen auf den Vierteln.**
func TestVierDrehkopienLiegenAufDenVierteln(t *testing.T) {
	t.Parallel()

	mitte := msepaket.Vektor{X: 0.5, Y: 0.5}

	spiegel := msepaket.Symbolteil{
		Art: msepaket.TeilSymmetrie, Spiegelart: msepaket.SpiegelungDrehung,
		Kopien: 4, Mitte: mitte, Griff: msepaket.Vektor{X: 0, Y: -0.5},
	}

	// Ein Punkt rechts von der Mitte.
	punkt := msepaket.Vektor{X: 0.9, Y: 0.5}

	// **Gegen den Uhrzeigersinn**, so wie MSEs Drehmatrix es tut.
	will := []msepaket.Vektor{
		{X: 0.9, Y: 0.5}, {X: 0.5, Y: 0.9}, {X: 0.1, Y: 0.5}, {X: 0.5, Y: 0.1},
	}

	for kopie := range 4 {
		if hat := service.SymmetrieAbbilden(spiegel, kopie, punkt); !nahe(hat, will[kopie]) {
			t.Errorf("Kopie %d ist %+v, will %+v", kopie, hat, will[kopie])
		}
	}
}

// **Eine Spiegelung spiegelt an der Geraden durch den Griff.**
//
// Der Griff zeigt in die Richtung des Originals; gespiegelt wird an der
// Geraden durch die Mitte in dieser Richtung. Die Matrix dafür ist
//
//	Matrix2D rot(cos(a+b),sin(a+b), sin(a+b),-cos(a+b));
//
// ebd., Zeile 182, mit `b = 2 * handle.angle()`.
func TestEineSpiegelungSpiegeltAmGriff(t *testing.T) {
	t.Parallel()

	mitte := msepaket.Vektor{X: 0.5, Y: 0.5}

	punkt := msepaket.Vektor{X: 0.7, Y: 0.8}

	// **Der Griff sagt, wo das Original liegt** — und gespiegelt wird so, dass
	// es dort bleibt.
	//
	// **Der erste Anlauf dieses Tests hatte es umgekehrt** und wäre um ein
	// Haar als Fehler der Umsetzung gebucht worden: „Griff nach rechts, also
	// waagerechte Achse". Nachgerechnet ist es die **senkrechte** — bei
	// zwei Kopien ist `a = π`, und die Achse liegt bei `(a + b)/2`, also um
	// ein Viertel gedreht zum Griff. Das Original bleibt damit auf seiner
	// Seite, und genau das soll der Griff bedeuten.
	for name, fall := range map[string]struct {
		griff, will msepaket.Vektor
	}{
		"Griff nach rechts spiegelt an der Senkrechten": {
			griff: msepaket.Vektor{X: 0.5, Y: 0},
			will:  msepaket.Vektor{X: 0.3, Y: 0.8},
		},
		"Griff nach unten spiegelt an der Waagerechten": {
			griff: msepaket.Vektor{X: 0, Y: 0.5},
			will:  msepaket.Vektor{X: 0.7, Y: 0.2},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			spiegel := msepaket.Symbolteil{
				Art: msepaket.TeilSymmetrie, Spiegelart: msepaket.SpiegelungAchse,
				Kopien: 2, Mitte: mitte, Griff: fall.griff,
			}

			if hat := service.SymmetrieAbbilden(spiegel, 1, punkt); !nahe(hat, fall.will) {
				t.Errorf("die Spiegelkopie ist %+v, will %+v", hat, fall.will)
			}
		})
	}
}

// **Eine Spiegelung mit ungerader Zahl rundet ab.**
//
//	int copies = s->kind == SYMMETRY_REFLECTION ? s->copies / 2 * 2 : s->copies;
//
// Ebd., Zeile 155. Eine Spiegelung braucht **Paare** — Original und Bild;
// eine ungerade Zahl ließe eines allein.
func TestEineSpiegelungMitUngeraderZahlRundetAb(t *testing.T) {
	t.Parallel()

	// **Eine negative Zahl gibt null.**
	//
	//	copies = new wxSpinCtrl(tb, ID_COPIES, _("2"), …, 2, 10, 2);
	//
	// `src/gui/symbol/symmetry_editor.cpp`, Zeile 57, Fassung f9d9356 vom
	// 2021-07-06: MSEs Zähler läuft von **2 bis 10**, eine negative Zahl
	// bekommt er nie.
	//
	// Hier kommt `Kopien` aus einer **gelesenen Datei**, und `copies: -3` ist
	// ein Wert, den eine fremde Datei tragen kann (§8). `-3 / 2 * 2` gäbe in
	// Go **-2** — eine Zahl, die als „Anzahl Kopien" durch jede Schleife und
	// jede Winkelrechnung dahinter liefe.
	for _, fall := range []struct{ kopien, will int }{
		{-3, 0}, {-1, 0}, {0, 0}, {1, 0}, {2, 2}, {3, 2}, {4, 4}, {5, 4}, {7, 6},
	} {
		spiegel := msepaket.Symbolteil{
			Art: msepaket.TeilSymmetrie, Spiegelart: msepaket.SpiegelungAchse,
			Kopien: fall.kopien,
		}

		if hat := service.SymmetriekopieN(spiegel); hat != fall.will {
			t.Errorf("%d Kopien geben %d, will %d", fall.kopien, hat, fall.will)
		}
	}

	// **Eine Drehung rundet nicht.**
	drehung := msepaket.Symbolteil{
		Art: msepaket.TeilSymmetrie, Spiegelart: msepaket.SpiegelungDrehung, Kopien: 3,
	}

	if hat := service.SymmetriekopieN(drehung); hat != 3 {
		t.Errorf("eine Drehung mit drei Kopien gibt %d", hat)
	}
}

// **Eine Kopie außerhalb bildet nicht ab.**
func TestEineKopieAusserhalbBildetNichtAb(t *testing.T) {
	t.Parallel()

	spiegel := msepaket.Symbolteil{
		Art: msepaket.TeilSymmetrie, Spiegelart: msepaket.SpiegelungDrehung,
		Kopien: 2, Mitte: msepaket.Vektor{X: 0.5, Y: 0.5},
	}

	punkt := msepaket.Vektor{X: 0.9, Y: 0.5}

	for _, kopie := range []int{-1, 2, 99} {
		if hat := service.SymmetrieAbbilden(spiegel, kopie, punkt); !nahe(hat, punkt) {
			t.Errorf("Kopie %d bildete auf %+v ab", kopie, hat)
		}
	}

	// **Und null Kopien ebenso wenig** — sonst teilte die Winkelrechnung
	// durch null.
	leer := msepaket.Symbolteil{Art: msepaket.TeilSymmetrie, Kopien: 0}
	if hat := service.SymmetrieAbbilden(leer, 0, punkt); !nahe(hat, punkt) {
		t.Errorf("ohne Kopien bildete es auf %+v ab", hat)
	}
}
