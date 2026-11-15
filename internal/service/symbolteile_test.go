// **Gruppen und Reihenfolge in der Teilliste** — Werkstück 5 von Phase 3.7
// (#1168,
// Befund Z5-02).
//
// # Woher die Regeln kommen
//
// Aus `src/data/action/symbol.cpp`, Fassung `f9d9356` vom 2021-07-06:
// `GroupSymbolPartsAction` (Zeilen 508–529), `UngroupSymbolPartsAction`
// (Zeilen 534–550) und `ReorderSymbolPartsAction` (Zeilen 453–473). Die
// Bedingungen, unter denen MSE die Einträge überhaupt anbietet, stehen
// woanders — in `src/gui/symbol/select_editor.cpp`, Zeilen 155–164, Fassung
// `f9d9356` vom 2021-07-06 — und sie
// gehören dazu: eine Regel ohne ihre Vorbedingung ist eine halbe Regel.
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// formen baut eine Liste benannter Formen — je drei Punkte, damit sie zählen.
func formen(bezeichnungen ...string) []msepaket.Symbolteil {
	gebaut := make([]msepaket.Symbolteil, 0, len(bezeichnungen))

	for _, name := range bezeichnungen {
		gebaut = append(gebaut, msepaket.Symbolteil{
			Art:  msepaket.TeilForm,
			Name: name,
			Punkte: []msepaket.Kontrollpunkt{
				{Stelle: msepaket.Vektor{X: 0, Y: 0}},
				{Stelle: msepaket.Vektor{X: 1, Y: 0}},
				{Stelle: msepaket.Vektor{X: 0, Y: 1}},
			},
		})
	}

	return gebaut
}

// teilnamen liest die Namen einer Liste — für die Zusicherungen.
func teilnamen(teile []msepaket.Symbolteil) []string {
	gelesen := make([]string, 0, len(teile))

	for _, teil := range teile {
		gelesen = append(gelesen, teil.Name)
	}

	return gelesen
}

// gleich vergleicht zwei Namenslisten.
func gleich(t *testing.T, bekommen, erwartet []string) {
	t.Helper()

	if len(bekommen) != len(erwartet) {
		t.Fatalf("Liste = %v, erwartet %v", bekommen, erwartet)
	}

	for i := range erwartet {
		if bekommen[i] != erwartet[i] {
			t.Fatalf("Liste = %v, erwartet %v", bekommen, erwartet)
		}
	}
}

// **Die Gruppe steht dort, wo das erste gewählte Teil stand.**
//
// Das ist die Aussage von `done` in `GroupSymbolPartsAction`
// (Zeilen 519–522): die Gruppe wird **einmal** eingefügt, und zwar beim
// ersten Treffer der Schleife über die Liste. Alles Ungewählte behält seine
// Reihenfolge.
func TestEineGruppeStehtBeimErstenGewaehltenTeil(t *testing.T) {
	t.Parallel()

	teile, err := service.TeileGruppieren(formen("A", "B", "C", "D"), []int{1, 3})
	if err != nil {
		t.Fatalf("Gruppieren: %v", err)
	}

	gleich(t, teilnamen(teile), []string{"A", "Gruppe", "C"})

	if got := teilnamen(teile[1].Teile); len(got) != 2 || got[0] != "B" || got[1] != "D" {
		t.Fatalf("Kinder = %v, erwartet [B D]", got)
	}

	if teile[1].Art != msepaket.TeilGruppe {
		t.Fatalf("Art = %v, erwartet TeilGruppe", teile[1].Art)
	}
}

// **Die Kinder stehen in Listenreihenfolge, nicht in Wahlreihenfolge.**
//
// `GroupSymbolPartsAction` läuft über `root.parts` und fragt bei jedem, ob es
// gewählt ist — die Wahl ist ein `set`, und ein `set` hat keine Reihenfolge.
// Wer zuerst D und dann B anklickt, bekommt trotzdem B vor D.
func TestDieKinderStehenInListenreihenfolge(t *testing.T) {
	t.Parallel()

	teile, err := service.TeileGruppieren(formen("A", "B", "C", "D"), []int{3, 1})
	if err != nil {
		t.Fatalf("Gruppieren: %v", err)
	}

	gleich(t, teilnamen(teile[1].Teile), []string{"B", "D"})
}

// **Unter zwei Teilen gibt es nichts zu gruppieren.**
//
// `select_editor.cpp`, Zeile 156: `ev.Enable(control.selected_parts.size() >=
// 2)`. MSE bietet den Eintrag gar nicht erst an; hier ist es ein Fehler, weil
// eine Bindung keine Menüleiste hat, die sich ausgrauen ließe.
func TestUnterZweiTeilenGibtEsKeineGruppe(t *testing.T) {
	t.Parallel()

	for _, stellen := range [][]int{{}, {0}} {
		if _, err := service.TeileGruppieren(formen("A", "B"), stellen); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("Gruppieren(%v) = %v, erwartet ErrValidation", stellen, err)
		}
	}
}

// **Eine Stelle, die es nicht gibt, ist ein Fehler.**
func TestEineGruppeUeberEineStelleDieEsNichtGibt(t *testing.T) {
	t.Parallel()

	for _, stellen := range [][]int{{0, 2}, {-1, 0}, {0, 0}} {
		if _, err := service.TeileGruppieren(formen("A", "B"), stellen); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("Gruppieren(%v) = %v, erwartet ErrValidation", stellen, err)
		}
	}
}

// **Auflösen setzt die Kinder an die Stelle der Gruppe.**
//
// `UngroupSymbolPartsAction`, Zeilen 538–549: die Schleife läuft über die
// Liste, und wo eine gewählte Gruppe steht, kommen ihre Kinder in Reihenfolge
// an ihre Stelle.
func TestEineAufgeloesteGruppeGibtIhreKinderFrei(t *testing.T) {
	t.Parallel()

	gruppiert, err := service.TeileGruppieren(formen("A", "B", "C", "D"), []int{1, 2})
	if err != nil {
		t.Fatalf("Gruppieren: %v", err)
	}

	teile, err := service.GruppeAufloesen(gruppiert, []int{1})
	if err != nil {
		t.Fatalf("Auflösen: %v", err)
	}

	gleich(t, teilnamen(teile), []string{"A", "B", "C", "D"})
}

// **Eine Symmetrie wird hier nicht aufgelöst — und das ist eine Abweichung.**
//
// In MSE *ist* sie eine Gruppe (`class SymbolSymmetry : public SymbolGroup`,
// `src/data/symbol.hpp`, Zeile 261, Fassung `f9d9356` vom 2021-07-06), und die
// Handlung bräche sie mit auf. Der Editor versucht genau das zu verhindern:
// `p->isSymbolGroup() && !p->isSymbolSymmetry()` (`select_editor.cpp`,
// Zeile 160) schaltet den Eintrag nur für **andere** Gruppen frei — wer beides
// wählt, verliert bei MSE trotzdem beide. Hier sind es zwei Arten und keine
// Vererbung; die Absicht des Menüs ist umsetzbar und wird umgesetzt (§15,
// Regel 7).
func TestEineSymmetrieBleibtBeimAufloesenStehen(t *testing.T) {
	t.Parallel()

	teile := formen("A", "B", "C")
	teile[0].Art = msepaket.TeilSymmetrie
	teile[0].Teile = formen("Kind")

	gruppiert, err := service.TeileGruppieren(teile, []int{1, 2})
	if err != nil {
		t.Fatalf("Gruppieren: %v", err)
	}

	// Stelle 0 ist die Symmetrie, Stelle 1 die Gruppe — beide gewählt.
	aufgeloest, err := service.GruppeAufloesen(gruppiert, []int{0, 1})
	if err != nil {
		t.Fatalf("Auflösen: %v", err)
	}

	gleich(t, teilnamen(aufgeloest), []string{"A", "B", "C"})

	if aufgeloest[0].Art != msepaket.TeilSymmetrie {
		t.Fatalf("Art an Stelle 0 = %v, erwartet TeilSymmetrie", aufgeloest[0].Art)
	}
}

// **Eine gewählte Form bleibt stehen, statt einen Fehler zu machen** —
// `else { old_part_list.push_back(p); }` (`symbol.cpp`, Zeile 546).
func TestEineGewaehlteFormBleibtBeimAufloesenStehen(t *testing.T) {
	t.Parallel()

	gruppiert, err := service.TeileGruppieren(formen("A", "B", "C"), []int{0, 1})
	if err != nil {
		t.Fatalf("Gruppieren: %v", err)
	}

	// Stelle 0 ist die Gruppe, Stelle 1 die übrig gebliebene Form C.
	aufgeloest, err := service.GruppeAufloesen(gruppiert, []int{0, 1})
	if err != nil {
		t.Fatalf("Auflösen: %v", err)
	}

	gleich(t, teilnamen(aufgeloest), []string{"A", "B", "C"})
}

// **Ohne eine einzige Gruppe in der Wahl gibt es nichts aufzulösen.**
//
// MSE graut den Eintrag dann aus (`select_editor.cpp`, Zeilen 157–164); eine
// Bindung hat keine Menüleiste und meldet einen Fehler.
func TestAufloesenWeistZurueck(t *testing.T) {
	t.Parallel()

	faelle := map[string][]int{
		"keine Gruppe in der Wahl": {0, 1},
		"eine Stelle zu weit":      {2},
		"dieselbe zweimal":         {0, 0},
	}

	for name, stellen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := service.GruppeAufloesen(formen("A", "B"), stellen); !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("Auflösen(%v) = %v, erwartet ErrValidation", stellen, err)
			}
		})
	}
}

// **Verschieben ist Herausnehmen und Einsetzen** —
// `ReorderSymbolPartsAction::perform`, Zeilen 462–469.
//
// **Und die Zielstelle meint die Liste nach dem Herausnehmen.** Das ist die
// Zeile, die MSE eigens dafür hat: `if (drag_parent == drop_parent &&
// drag_position < drop_position) drop_position -= 1;`
// (`src/gui/symbol/part_list.cpp`, Zeile 158, Fassung `f9d9356` vom
// 2021-07-06). Hier zählt die Zielstelle
// deshalb ohne das bewegte Teil — wer A an Stelle 2 schiebt, bekommt es
// hinter C.
func TestEinTeilLaesstSichVerschieben(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		von, nach int
		erwartet  []string
	}{
		{"nach hinten", 0, 2, []string{"B", "C", "A", "D"}},
		{"nach vorn", 3, 0, []string{"D", "A", "B", "C"}},
		{"ans Ende", 0, 3, []string{"B", "C", "D", "A"}},
		{"auf sich selbst", 1, 1, []string{"A", "B", "C", "D"}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			teile, err := service.TeilVerschieben(formen("A", "B", "C", "D"), fall.von, fall.nach)
			if err != nil {
				t.Fatalf("Verschieben: %v", err)
			}

			gleich(t, teilnamen(teile), fall.erwartet)
		})
	}
}

// **Eine Stelle außerhalb der Liste ist ein Fehler.**
func TestVerschiebenUeberDieListe(t *testing.T) {
	t.Parallel()

	faelle := [][2]int{{-1, 0}, {4, 0}, {0, -1}, {0, 4}}

	for _, fall := range faelle {
		if _, err := service.TeilVerschieben(formen("A", "B", "C", "D"), fall[0], fall[1]); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("Verschieben(%d, %d) = %v, erwartet ErrValidation", fall[0], fall[1], err)
		}
	}
}

// **Keine der drei fasst die übergebene Liste an.**
//
// Sie kommt aus einer Bindung und gehört dem Aufrufer; wer sie an Ort und
// Stelle umbaute, änderte den Stand des Fensters, bevor die Antwort da ist.
func TestDieUebergebeneListeBleibtStehen(t *testing.T) {
	t.Parallel()

	urspruenglich := formen("A", "B", "C")

	if _, err := service.TeileGruppieren(urspruenglich, []int{0, 1}); err != nil {
		t.Fatalf("Gruppieren: %v", err)
	}

	if _, err := service.TeilVerschieben(urspruenglich, 0, 2); err != nil {
		t.Fatalf("Verschieben: %v", err)
	}

	gleich(t, teilnamen(urspruenglich), []string{"A", "B", "C"})
}
