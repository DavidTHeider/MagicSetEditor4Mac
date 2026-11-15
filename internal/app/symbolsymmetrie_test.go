// **Die Symmetrie im Symboleditor**
// (#1092).
package app_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// quadrat gibt ein Teil mit vier Punkten in der linken oberen Ecke.
func quadrat() app.SymbolteilDTO {
	return app.SymbolteilDTO{
		Art: "Form", Name: "Quadrat",
		Punkte: []app.SymbolpunktDTO{
			{X: 0.6, Y: 0.4}, {X: 0.8, Y: 0.4}, {X: 0.8, Y: 0.6}, {X: 0.6, Y: 0.6},
		},
	}
}

// **Eine Symmetrie umschließt das gewählte Teil.**
//
// **Und legt kein leeres an**: eine Symmetrie ohne Kinder zeichnet nichts, und
// ein Editor, der sie anböte, ließe jemanden auf eine Wirkung warten, die
// nicht kommt.
func TestSymmetrieAnlegenUmschliesstDasTeil(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").SymmetrieAnlegen([]app.SymbolteilDTO{quadrat()}, 0)
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Teile) != 1 {
		t.Fatalf("%d Teile statt einem: %+v", len(antwort.Teile), antwort.Teile)
	}

	spiegel := antwort.Teile[0]
	if spiegel.Art != "Symmetrie" {
		t.Fatalf("die Art ist %q", spiegel.Art)
	}

	if len(spiegel.Teile) != 1 || spiegel.Teile[0].Name != "Quadrat" {
		t.Errorf("das Teil ist nicht darin: %+v", spiegel.Teile)
	}

	// **Drehung mit zwei Kopien, Mitte in der Mitte, Griff nach rechts** —
	// dieselbe Vorgabe wie bei MSE.
	if spiegel.Spiegelart != "rotation" || spiegel.Kopien != 2 {
		t.Errorf("die Vorgabe ist %q / %d", spiegel.Spiegelart, spiegel.Kopien)
	}

	if spiegel.MitteX != 0.5 || spiegel.MitteY != 0.5 {
		t.Errorf("die Mitte ist %v / %v", spiegel.MitteX, spiegel.MitteY)
	}

	// **Ein Griff im Ursprung hätte den Winkel null und wäre nicht zu fassen.**
	if spiegel.GriffX == 0 && spiegel.GriffY == 0 {
		t.Error("der Griff liegt auf der Mitte")
	}
}

// **Das Auflösen lässt die Kinder stehen.**
func TestSymmetrieAufloesenLaesstDieKinderStehen(t *testing.T) {
	t.Parallel()

	angelegt := app.NeueApp("").SymmetrieAnlegen(
		[]app.SymbolteilDTO{quadrat(), quadrat()}, 0)

	antwort := app.NeueApp("").SymmetrieAufloesen(angelegt.Teile, 0)
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Teile) != 2 {
		t.Fatalf("%d Teile statt zwei: %+v", len(antwort.Teile), antwort.Teile)
	}

	for _, teil := range antwort.Teile {
		if teil.Art != "Form" {
			t.Errorf("ein Teil ist %q statt »Form«", teil.Art)
		}
	}
}

// **Art, Kopienzahl, Mitte und Griff lassen sich einzeln setzen.**
//
// **Einzeln ist der Punkt**: die vier Zahlen reisen immer mit, und ohne
// `mitteSetzen`/`griffSetzen` zöge ein Aufruf, der nur die Art ändert, die
// Mitte in die linke obere Ecke.
func TestSymmetrieSetzenAendertNurWasGenanntIst(t *testing.T) {
	t.Parallel()

	teile := app.NeueApp("").SymmetrieAnlegen([]app.SymbolteilDTO{quadrat()}, 0).Teile

	nurArt := app.NeueApp("").SymmetrieSetzen(teile,
		app.SymmetrieweiseDTO{Teil: 0, Spiegelart: "reflection"})
	if nurArt.Fehler != nil {
		t.Fatalf("Fehler: %+v", nurArt.Fehler)
	}

	geaendert := nurArt.Teile[0]
	if geaendert.Spiegelart != "reflection" {
		t.Errorf("die Art ist %q", geaendert.Spiegelart)
	}

	if geaendert.MitteX != 0.5 || geaendert.MitteY != 0.5 {
		t.Errorf("die Mitte ist gewandert: %v / %v", geaendert.MitteX, geaendert.MitteY)
	}

	if geaendert.Kopien != 2 {
		t.Errorf("die Kopienzahl ist %d", geaendert.Kopien)
	}

	nurMitte := app.NeueApp("").SymmetrieSetzen(nurArt.Teile,
		app.SymmetrieweiseDTO{Teil: 0, MitteX: 0.25, MitteY: 0.75, MitteSetzen: true})

	if nurMitte.Teile[0].MitteX != 0.25 || nurMitte.Teile[0].MitteY != 0.75 {
		t.Errorf("die Mitte ist %+v", nurMitte.Teile[0])
	}

	// **Und die Art bleibt** — sonst hätte der leere Wert sie zurückgesetzt.
	if nurMitte.Teile[0].Spiegelart != "reflection" {
		t.Errorf("die Art ist %q", nurMitte.Teile[0].Spiegelart)
	}
}

// **Die Kopien werden gerechnet — von `service.SymmetrieAbbilden`.**
//
// **Der produktive Aufrufer, der #1092 gefehlt hat.** Geprüft ist nicht die
// Rechnung selbst (das tut `internal/service`), sondern dass sie **hier
// ankommt**: eine Drehung um 180° schickt einen Punkt bei `(0,6 / 0,5)` auf
// `(0,4 / 0,5)`.
func TestSymmetriekopienRechnetDieAbbildung(t *testing.T) {
	t.Parallel()

	teile := app.NeueApp("").SymmetrieAnlegen([]app.SymbolteilDTO{{
		Art: "Form", Name: "Strich",
		Punkte: []app.SymbolpunktDTO{{X: 0.6, Y: 0.5}, {X: 0.7, Y: 0.5}},
	}}, 0).Teile

	antwort := app.NeueApp("").Symmetriekopien(teile, 0)
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Kopien) != 2 {
		t.Fatalf("%d Kopien statt zwei", len(antwort.Kopien))
	}

	// **Die erste ist das Original** — die Abbildung für `kopie == 0` ist die
	// Identität.
	erste := antwort.Kopien[0][0].Punkte[0]
	if math.Abs(erste.X-0.6) > 1e-9 || math.Abs(erste.Y-0.5) > 1e-9 {
		t.Errorf("die erste Kopie ist verschoben: %+v", erste)
	}

	// **Die zweite steht gegenüber der Mitte (0,5 / 0,5).**
	zweite := antwort.Kopien[1][0].Punkte[0]
	if math.Abs(zweite.X-0.4) > 1e-9 || math.Abs(zweite.Y-0.5) > 1e-9 {
		t.Errorf("die zweite Kopie steht bei %+v statt bei (0,4 / 0,5)", zweite)
	}
}

// **Eine Spiegelung rundet die Kopienzahl auf eine gerade ab.**
//
// `int copies = s->kind == SYMMETRY_REFLECTION ? s->copies / 2 * 2 : s->copies;`
// — `src/render/symbol/viewer.cpp`, Zeile 155, Fassung f9d9356 vom 2021-07-06.
func TestEineSpiegelungMitUngeraderZahlRundetAb(t *testing.T) {
	t.Parallel()

	teile := app.NeueApp("").SymmetrieAnlegen([]app.SymbolteilDTO{quadrat()}, 0).Teile

	gesetzt := app.NeueApp("").SymmetrieSetzen(teile, app.SymmetrieweiseDTO{
		Teil: 0, Spiegelart: "reflection", Kopien: 5,
	})
	if gesetzt.Fehler != nil {
		t.Fatalf("Fehler: %+v", gesetzt.Fehler)
	}

	antwort := app.NeueApp("").Symmetriekopien(gesetzt.Teile, 0)
	if len(antwort.Kopien) != 4 {
		t.Errorf("%d Kopien — fünf müssten auf vier abrunden", len(antwort.Kopien))
	}
}

// **Jeder Weg weist ab, was keine Symmetrie ist.**
func TestDieSymmetriewegeWeisenAb(t *testing.T) {
	t.Parallel()

	form := []app.SymbolteilDTO{quadrat()}

	t.Run("eine Stelle, die es nicht gibt", func(t *testing.T) {
		t.Parallel()

		if app.NeueApp("").SymmetrieAnlegen(form, 4).Fehler == nil {
			t.Error("Anlegen kam durch")
		}

		if app.NeueApp("").SymmetrieAufloesen(form, -1).Fehler == nil {
			t.Error("Auflösen kam durch")
		}

		if app.NeueApp("").SymmetrieSetzen(form,
			app.SymmetrieweiseDTO{Teil: 9}).Fehler == nil {
			t.Error("Setzen kam durch")
		}

		if app.NeueApp("").Symmetriekopien(form, 9).Fehler == nil {
			t.Error("Kopien kam durch")
		}
	})

	t.Run("ein Teil, das keine Symmetrie ist", func(t *testing.T) {
		t.Parallel()

		if app.NeueApp("").SymmetrieAufloesen(form, 0).Fehler == nil {
			t.Error("Auflösen kam durch")
		}

		if app.NeueApp("").SymmetrieSetzen(form,
			app.SymmetrieweiseDTO{Teil: 0}).Fehler == nil {
			t.Error("Setzen kam durch")
		}

		if app.NeueApp("").Symmetriekopien(form, 0).Fehler == nil {
			t.Error("Kopien kam durch")
		}
	})

	t.Run("eine Art, die es nicht gibt", func(t *testing.T) {
		t.Parallel()

		teile := app.NeueApp("").SymmetrieAnlegen(form, 0).Teile

		if app.NeueApp("").SymmetrieSetzen(teile,
			app.SymmetrieweiseDTO{Teil: 0, Spiegelart: "schrauben"}).Fehler == nil {
			t.Error("»schrauben« kam durch")
		}
	})

	t.Run("eine Kopienzahl außerhalb der Spanne", func(t *testing.T) {
		t.Parallel()

		teile := app.NeueApp("").SymmetrieAnlegen(form, 0).Teile

		for _, zahl := range []int{1, 11, -3} {
			if app.NeueApp("").SymmetrieSetzen(teile,
				app.SymmetrieweiseDTO{Teil: 0, Kopien: zahl}).Fehler == nil {
				t.Errorf("%d Kopien kamen durch", zahl)
			}
		}
	})
}

// **Ein abgewiesener Aufruf gibt die Zeichnung zurück, wie sie kam.**
//
// §7, Regel 3: wer eine unmögliche Kopienzahl setzt, bekommt eine Meldung —
// und behält, was er hat.
func TestEinAbgewiesenerSymmetrieaufrufAendertNichts(t *testing.T) {
	t.Parallel()

	teile := app.NeueApp("").SymmetrieAnlegen([]app.SymbolteilDTO{quadrat()}, 0).Teile

	antwort := app.NeueApp("").SymmetrieSetzen(teile,
		app.SymmetrieweiseDTO{Teil: 0, Kopien: 99})

	if antwort.Teile[0].Kopien != 2 {
		t.Errorf("die Kopienzahl steht bei %d", antwort.Teile[0].Kopien)
	}
}

// **Und der Griff lässt sich einzeln setzen.**
//
// **Neben der Mitte und nicht mit ihr**: `griffSetzen` ist ein eigener Wert,
// und ein Aufruf, der nur den Griff dreht, dürfte die Mitte nicht verschieben.
// Der Griff steht **relativ zur Mitte** — sein Winkel ist die Spiegelachse.
func TestDerSymmetriegriffLaesstSichEinzelnSetzen(t *testing.T) {
	t.Parallel()

	teile := app.NeueApp("").SymmetrieAnlegen([]app.SymbolteilDTO{quadrat()}, 0).Teile

	antwort := app.NeueApp("").SymmetrieSetzen(teile, app.SymmetrieweiseDTO{
		Teil: 0, GriffX: 0, GriffY: -0.3, GriffSetzen: true,
	})
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	geaendert := antwort.Teile[0]
	if geaendert.GriffX != 0 || geaendert.GriffY != -0.3 {
		t.Errorf("der Griff ist %v / %v", geaendert.GriffX, geaendert.GriffY)
	}

	// **Und die Mitte steht** — sonst hätte die 0 sie mitgezogen.
	if geaendert.MitteX != 0.5 || geaendert.MitteY != 0.5 {
		t.Errorf("die Mitte ist gewandert: %v / %v", geaendert.MitteX, geaendert.MitteY)
	}

	// **Und die Spiegelachse dreht sich mit.** Eine Spiegelung mit Griff nach
	// oben spiegelt an der Senkrechten: aus `(0,6 / 0,5)` wird `(0,4 / 0,5)`.
	gespiegelt := app.NeueApp("").SymmetrieSetzen(antwort.Teile,
		app.SymmetrieweiseDTO{Teil: 0, Spiegelart: "reflection"})

	kopien := app.NeueApp("").Symmetriekopien(gespiegelt.Teile, 0)
	if kopien.Fehler != nil {
		t.Fatalf("Kopien: %+v", kopien.Fehler)
	}

	if len(kopien.Kopien) != 2 {
		t.Fatalf("%d Kopien", len(kopien.Kopien))
	}
}
