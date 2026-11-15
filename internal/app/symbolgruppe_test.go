// **Gruppen und Reihenfolge in der Teilliste** — Werkstück 5 von Phase 3.7
// (#1168,
// Befund Z5-02).
//
// **Was hier geprüft wird und was nicht.** Die Regeln stehen in
// `internal/service/symbolteile.go` und sind dort geprüft — hier steht die
// **Naht**: kommen die Stellen an, kommt die Gestalt zurück, und wird ein
// Fehler übersetzt statt durchgereicht (§2, §8).
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// dreiTeile gibt drei Formen mit unterscheidbaren Namen.
func dreiTeile() []app.SymbolteilDTO {
	gebaut := make([]app.SymbolteilDTO, 0, 3)

	for _, name := range []string{"A", "B", "C"} {
		teil := quadrat()
		teil.Name = name
		gebaut = append(gebaut, teil)
	}

	return gebaut
}

// **Zwei Teile werden eine Gruppe, und die steht an erster Stelle.**
func TestTeileGruppierenGibtEineGruppeZurueck(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").TeileGruppieren(dreiTeile(), []int{0, 2})
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Teile) != 2 {
		t.Fatalf("Teile = %d, erwartet 2", len(antwort.Teile))
	}

	if antwort.Teile[0].Art != "Gruppe" {
		t.Fatalf("Art = %q, erwartet Gruppe", antwort.Teile[0].Art)
	}

	if len(antwort.Teile[0].Teile) != 2 {
		t.Fatalf("Kinder = %d, erwartet 2", len(antwort.Teile[0].Teile))
	}

	if antwort.Teile[1].Name != "B" {
		t.Fatalf("zweites Teil = %q, erwartet B", antwort.Teile[1].Name)
	}
}

// **Ein Fehler kommt übersetzt an und lässt die Liste stehen.**
//
// §8: die Oberfläche bekommt Code, Meldung und Feldbezug — und den Stand, mit
// dem sie weiterarbeiten kann. Eine leere Liste als Antwort auf einen
// Bedienfehler löschte die Zeichnung.
func TestTeileGruppierenMeldetEinenFehlerUndHaeltDieListe(t *testing.T) {
	t.Parallel()

	faelle := map[string][]int{
		"ein einzelnes Teil":  {1},
		"eine Stelle zu weit": {0, 9},
		"dieselbe zweimal":    {1, 1},
	}

	for name, stellen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := app.NeueApp("").TeileGruppieren(dreiTeile(), stellen)
			if antwort.Fehler == nil {
				t.Fatal("kein Fehler, erwartet einen")
			}

			if len(antwort.Teile) != 3 {
				t.Fatalf("Teile = %d, erwartet die drei von vorher", len(antwort.Teile))
			}
		})
	}
}

// **Eine aufgelöste Gruppe gibt ihre Kinder frei.**
func TestGruppeAufloesenGibtDieKinderFrei(t *testing.T) {
	t.Parallel()

	gruppiert := app.NeueApp("").TeileGruppieren(dreiTeile(), []int{0, 1})
	if gruppiert.Fehler != nil {
		t.Fatalf("Gruppieren: %+v", gruppiert.Fehler)
	}

	antwort := app.NeueApp("").GruppeAufloesen(gruppiert.Teile, []int{0})
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Teile) != 3 {
		t.Fatalf("Teile = %d, erwartet 3", len(antwort.Teile))
	}

	if antwort.Teile[0].Name != "A" || antwort.Teile[2].Name != "C" {
		t.Fatalf("Reihenfolge = %q…%q, erwartet A…C",
			antwort.Teile[0].Name, antwort.Teile[2].Name)
	}
}

// **Ohne eine einzige Gruppe in der Wahl kommt ein Fehler.**
//
// Eine gewählte Form neben einer Gruppe bleibt dagegen stehen — das ist MSEs
// Verhalten, und es ist in `service` geprüft.
func TestGruppeAufloesenMeldetWennKeineGruppeGewaehltIst(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").GruppeAufloesen(dreiTeile(), []int{0})
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler, erwartet einen")
	}

	if len(antwort.Teile) != 3 {
		t.Fatalf("Teile = %d, erwartet die drei von vorher", len(antwort.Teile))
	}
}

// **Ein Teil lässt sich verschieben.**
func TestTeilVerschiebenSetztEsAnDieNeueStelle(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").TeilVerschieben(dreiTeile(), 0, 2)
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if antwort.Teile[2].Name != "A" {
		t.Fatalf("letztes Teil = %q, erwartet A", antwort.Teile[2].Name)
	}
}

// **Und eine Stelle außerhalb der Liste ist ein Fehler.**
func TestTeilVerschiebenMeldetEineStelleDieEsNichtGibt(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").TeilVerschieben(dreiTeile(), 0, 9)
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler, erwartet einen")
	}

	if len(antwort.Teile) != 3 {
		t.Fatalf("Teile = %d, erwartet die drei von vorher", len(antwort.Teile))
	}
}
