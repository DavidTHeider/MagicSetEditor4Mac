// **Punkt und Griff über die Naht** —
// #890,
// Werkstück 4 von Phase X.
//
// **Eine eigene Datei**, weil `symbol_test.go` über die Zeilengrenze wuchs,
// als der Nachlauf des Zweitblicks den Griffzug auf Stellen umstellte (§2).
// Der Schnitt folgt der Frage: dort **die Zeichnung als Ganzes**, hier **der
// einzelne Punkt und sein Griff**.
package app_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Punkt lässt sich über die Naht einfügen und löschen** —
// #890.
//
// **`service.PunktEinfuegen` stand gerechnet und geprüft da, ohne dass ein
// Aufrufer sie erreichte** — und die Tastaturhilfe der Seite nannte beide
// Tasten trotzdem. Ein Paket ohne Aufrufer ist keine Anwendungsfunktion (§10).
func TestEinPunktLaesstSichUeberDieNahtAendern(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}

	teile := []app.SymbolteilDTO{{
		Art: "Form", Name: "Dreieck",
		Punkte: []app.SymbolpunktDTO{
			{X: 0, Y: 0, Sperre: "free"},
			{X: 1, Y: 0, Sperre: "free"},
			{X: 1, Y: 1, Sperre: "free"},
		},
	}}

	t.Run("einfügen setzt in die Mitte", func(t *testing.T) {
		t.Parallel()

		antwort := anwendung.PunktEinfuegen(teile, app.SymbolpunktweiseDTO{})
		if antwort.Fehler != nil {
			t.Fatalf("Fehler: %+v", antwort.Fehler)
		}

		if len(antwort.Teile[0].Punkte) != 4 {
			t.Fatalf("%d Punkte", len(antwort.Teile[0].Punkte))
		}

		// **Zwischen (0,0) und (1,0)** — die Mitte der ersten Strecke.
		if hat := antwort.Teile[0].Punkte[1]; hat.X != 0.5 || hat.Y != 0 {
			t.Errorf("der neue Punkt ist %+v", hat)
		}
	})

	t.Run("löschen nimmt einen weg", func(t *testing.T) {
		t.Parallel()

		antwort := anwendung.PunktLoeschen(teile, app.SymbolpunktweiseDTO{Punkt: 1})
		if antwort.Fehler != nil {
			t.Fatalf("Fehler: %+v", antwort.Fehler)
		}

		if len(antwort.Teile[0].Punkte) != 2 {
			t.Errorf("%d Punkte", len(antwort.Teile[0].Punkte))
		}
	})

	t.Run("unter zwei Punkten wird nicht gelöscht", func(t *testing.T) {
		t.Parallel()

		zwei := []app.SymbolteilDTO{{Art: "Form", Punkte: teile[0].Punkte[:2]}}

		antwort := anwendung.PunktLoeschen(zwei, app.SymbolpunktweiseDTO{})
		if antwort.Fehler == nil {
			t.Error("das Löschen kam durch")
		}

		// **Und die Zeichnung kommt unverändert zurück** (§7, Regel 3).
		if len(antwort.Teile) != 1 || len(antwort.Teile[0].Punkte) != 2 {
			t.Errorf("die Zeichnung ist %+v", antwort.Teile)
		}
	})

	t.Run("ein Teil, das es nicht gibt", func(t *testing.T) {
		t.Parallel()

		for name, stelle := range map[string]int{"davor": -1, "dahinter": 9} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				antwort := anwendung.PunktEinfuegen(teile,
					app.SymbolpunktweiseDTO{Teil: stelle})
				if antwort.Fehler == nil {
					t.Errorf("das Teil %d kam durch", stelle)
				}
			})
		}
	})
}

// **Ein Griff lässt sich über die Naht ziehen** —
// #890.
//
// **`service.GriffZiehen` stand gerechnet und geprüft da, ohne Aufrufer** —
// und die Zeichenfläche zeichnete keinen einzigen Griff. Angestrichen vom
// Zweitblick über Phase X.
func TestEinGriffLaesstSichUeberDieNahtZiehen(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}

	teile := []app.SymbolteilDTO{{
		Art: "Form",
		Punkte: []app.SymbolpunktDTO{
			{
				X: 0.5, Y: 0.5,
				GriffDavorX: -0.2, GriffDanachX: 0.2,
				KurveDanach: true, Sperre: "size",
			},
			{X: 1, Y: 1, Sperre: "size"},
		},
	}}

	t.Run("die Sperre zieht den anderen nach", func(t *testing.T) {
		t.Parallel()

		// **Die Spitze steht auf (0,7 | 0,5)** — der Punkt (0,5 | 0,5) plus
		// sein Griff (0,2 | 0). Sie auf (0,7 | 0,8) zu setzen heißt: der Griff
		// wird (0,2 | 0,3).
		antwort := anwendung.GriffZiehen(teile,
			app.SymbolgriffDTO{X: 0.7, Y: 0.8})
		if antwort.Fehler != nil {
			t.Fatalf("Fehler: %+v", antwort.Fehler)
		}

		punkt := antwort.Teile[0].Punkte[0]

		// Der gezogene Griff ist (0,2 | 0,3); `size` macht den anderen zu
		// seinem Gegenteil.
		//
		// **Verglichen wird mit Spielraum, und der hat einen Grund.** Aus
		// einer Stelle wird hier eine Verschiebung — `0,7 − 0,5 − 0,2` ist in
		// Gleitkomma nicht null, sondern −2,8·10⁻¹⁷. Ein Vergleich auf
		// Gleichheit prüfte die Rechenart der Maschine und nicht die Regel.
		if !nahebei(punkt.GriffDanachX, 0.2) ||
			!nahebei(punkt.GriffDanachY, 0.3) {
			t.Errorf("der gezogene Griff ist %+v", punkt)
		}

		if !nahebei(punkt.GriffDavorX, -0.2) || !nahebei(punkt.GriffDavorY, -0.3) {
			t.Errorf("der andere Griff ist (%v, %v)", punkt.GriffDavorX, punkt.GriffDavorY)
		}
	})

	// **Und der vordere Griff geht denselben Weg.**
	//
	// Er hat eine eigene Zeile in der Naht — welcher der beiden Griffe heute
	// dasteht, entscheidet, wovon die Verschiebung ausgeht. Ohne diesen Lauf
	// wäre sie eine Behauptung (§3).
	t.Run("der vordere Griff ebenso", func(t *testing.T) {
		t.Parallel()

		// **Die vordere Spitze steht auf (0,3 | 0,5)** — der Punkt (0,5 | 0,5)
		// plus sein Griff (−0,2 | 0). Auf (0,3 | 0,2) gesetzt gibt das den
		// Griff (−0,2 | −0,3).
		antwort := anwendung.GriffZiehen(teile,
			app.SymbolgriffDTO{Davor: true, X: 0.3, Y: 0.2})
		if antwort.Fehler != nil {
			t.Fatalf("Fehler: %+v", antwort.Fehler)
		}

		punkt := antwort.Teile[0].Punkte[0]

		if !nahebei(punkt.GriffDavorX, -0.2) || !nahebei(punkt.GriffDavorY, -0.3) {
			t.Errorf("der gezogene Griff ist (%v, %v)", punkt.GriffDavorX, punkt.GriffDavorY)
		}

		// `size` macht den anderen zum Gegenteil.
		if !nahebei(punkt.GriffDanachX, 0.2) || !nahebei(punkt.GriffDanachY, 0.3) {
			t.Errorf("der andere Griff ist (%v, %v)", punkt.GriffDanachX, punkt.GriffDanachY)
		}
	})

	t.Run("eine Stelle, die es nicht gibt", func(t *testing.T) {
		t.Parallel()

		for name, griff := range map[string]app.SymbolgriffDTO{
			"kein Teil":  {Teil: 9},
			"kein Punkt": {Punkt: 9},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				antwort := anwendung.GriffZiehen(teile, griff)
				if antwort.Fehler == nil {
					t.Errorf("%+v kam durch", griff)
				}
			})
		}
	})
}

// nahebei vergleicht zwei Gleitkommazahlen mit dem Spielraum, den die
// Umrechnung von Stelle auf Verschiebung braucht.
func nahebei(hat, soll float64) bool {
	return math.Abs(hat-soll) < 1e-9
}
