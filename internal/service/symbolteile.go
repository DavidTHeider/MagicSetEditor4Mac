// **Gruppen und Reihenfolge in der Teilliste eines Symbols** — Werkstück 5 von
// Phase 3.7
// (#1168,
// Befund Z5-02).
//
// # Was hier steht und was nicht
//
// Drei Umformungen einer Teilliste, portiert aus
// `src/data/action/symbol.cpp`, Fassung `f9d9356` vom 2021-07-06. Sie stehen
// in `service` und nicht in `app`, aus demselben Grund wie [Grundform] und
// [PunktEinfuegen]: das ist die **Regel**, und `app` übersetzt nur DTOs (§2).
//
// **Die Vorbedingungen gehören dazu.** MSE graut seine Menüeinträge aus
// (`src/gui/symbol/select_editor.cpp`, Zeilen 155–164, Fassung `f9d9356` vom
// 2021-07-06); eine Bindung hat keine
// Menüleiste, die sich ausgrauen ließe, und meldet deshalb einen Fehler.

package service

import (
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Gruppenname ist der Name, den eine neu gebildete Gruppe bekommt.
//
// **Deutsch, wie die Grundformen** — MSE nimmt ihn aus seinem eigenen
// Übersetzungskatalog (`SymbolGroup::SymbolGroup`, `src/data/symbol.cpp`,
// Zeile 290, Fassung `f9d9356` vom 2021-07-06: `name =
// capitalize(_TYPE_("group"))`), nicht fest aus dem Quelltext. Dieselbe
// Begründung wie bei [kreisname].
const Gruppenname = "Gruppe"

// TeileGruppieren fasst die genannten Stellen zu einer Gruppe zusammen.
//
// **Die Gruppe steht, wo das erste gewählte Teil stand**, und ihre Kinder
// stehen in **Listenreihenfolge** — nicht in der Reihenfolge des Anklickens.
// Beides ist die Aussage von `GroupSymbolPartsAction` (ebd., Zeilen 512–527):
// die Schleife läuft über die Liste und fragt bei jedem Teil, ob es gewählt
// ist; die Wahl selbst ist ein `set` und hat keine Reihenfolge.
//
// **Unter zwei Teilen gibt es nichts zu gruppieren** — `select_editor.cpp`,
// Zeile 156.
func TeileGruppieren(
	teile []msepaket.Symbolteil, stellen []int,
) ([]msepaket.Symbolteil, error) {
	gewaehlt, err := gewaehlteStellen(teile, stellen)
	if err != nil {
		return nil, err
	}

	if len(gewaehlt) < 2 {
		return nil, fmt.Errorf(
			"%w: eine Gruppe braucht mindestens zwei Teile, gewählt sind %d",
			domain.ErrValidation, len(gewaehlt))
	}

	gruppe := msepaket.Symbolteil{Art: msepaket.TeilGruppe, Name: Gruppenname}
	gebaut := make([]msepaket.Symbolteil, 0, len(teile))

	// **Die Stelle wird beim ersten Treffer freigehalten.** MSE führt dafür
	// ein `done`; ohne es stünde die Gruppe so oft in der Liste, wie Teile
	// gewählt sind. Eingetragen wird sie erst danach — fertig ist sie erst,
	// wenn die Schleife ihre Kinder eingesammelt hat.
	stelleDerGruppe := -1

	for stelle, teil := range teile {
		if !gewaehlt[stelle] {
			gebaut = append(gebaut, teil)

			continue
		}

		gruppe.Teile = append(gruppe.Teile, teil)

		if stelleDerGruppe < 0 {
			stelleDerGruppe = len(gebaut)

			gebaut = append(gebaut, msepaket.Symbolteil{})
		}
	}

	gebaut[stelleDerGruppe] = gruppe

	return gebaut, nil
}

// GruppeAufloesen setzt die Kinder der genannten Gruppen an deren Stelle.
//
// Portiert aus `UngroupSymbolPartsAction` (ebd., Zeilen 538–549): die Schleife
// läuft über die Liste, und wo eine gewählte Gruppe steht, kommen ihre Kinder
// in Reihenfolge an ihre Stelle.
//
// **Was gewählt ist und keine Gruppe, bleibt stehen** — auch das ist MSE
// (`else { old_part_list.push_back(p); }`, Zeile 546). Der Fehler kommt
// deshalb erst, wenn **gar keine** Gruppe gewählt ist: MSE graut den Eintrag
// genau dann aus (`select_editor.cpp`, Zeilen 157–164, Fassung `f9d9356` vom
// 2021-07-06).
//
// **Eine Symmetrie wird hier nie aufgelöst, und das ist eine Abweichung**
// (§15, Regel 7). In MSE **ist** sie eine Gruppe
// (`class SymbolSymmetry : public SymbolGroup`, `src/data/symbol.hpp`,
// Zeile 261, Fassung `f9d9356` vom 2021-07-06), und die Handlung bräche sie
// mit auf — der Editor versucht das gerade zu verhindern:
// `p->isSymbolGroup() && !p->isSymbolSymmetry()` schaltet den Eintrag nur für
// **andere** Gruppen frei. Wer eine Symmetrie **und** eine Gruppe wählt,
// verliert bei MSE trotzdem beide, weil die Handlung die Bedingung des Menüs
// nicht kennt.
//
// Hier sind es zwei Arten und keine Vererbung; die Absicht des Menüs ist
// deshalb umsetzbar, und sie wird umgesetzt. Aufgelöst wird eine Symmetrie
// über [App.SymmetrieAufloesen] — dort, wo sie auch angelegt wird.
func GruppeAufloesen(
	teile []msepaket.Symbolteil, stellen []int,
) ([]msepaket.Symbolteil, error) {
	gewaehlt, err := gewaehlteStellen(teile, stellen)
	if err != nil {
		return nil, err
	}

	if !slices.ContainsFunc(stellen, func(stelle int) bool {
		return teile[stelle].Art == msepaket.TeilGruppe
	}) {
		return nil, fmt.Errorf(
			"%w: unter den gewählten Teilen ist keine Gruppe",
			domain.ErrValidation)
	}

	gebaut := make([]msepaket.Symbolteil, 0, len(teile))

	for stelle, teil := range teile {
		if gewaehlt[stelle] && teil.Art == msepaket.TeilGruppe {
			gebaut = append(gebaut, teil.Teile...)

			continue
		}

		gebaut = append(gebaut, teil)
	}

	return gebaut, nil
}

// TeilVerschieben nimmt ein Teil heraus und setzt es an anderer Stelle ein.
//
// Portiert aus `ReorderSymbolPartsAction::perform` (ebd., Zeilen 462–469).
//
// **Die Zielstelle zählt ohne das bewegte Teil.** MSE rechnet sie eigens um:
// `if (drag_parent == drop_parent && drag_position < drop_position)
// drop_position -= 1;` (`src/gui/symbol/part_list.cpp`, Zeile 158, Fassung
// `f9d9356` vom 2021-07-06). Wer das
// erste Teil an Stelle 2 schiebt, bekommt es deshalb hinter das dritte und
// nicht davor.
func TeilVerschieben(
	teile []msepaket.Symbolteil, von, nach int,
) ([]msepaket.Symbolteil, error) {
	if von < 0 || von >= len(teile) {
		return nil, keineTeilstelle(von)
	}

	if nach < 0 || nach >= len(teile) {
		return nil, keineTeilstelle(nach)
	}

	gebaut := slices.Clone(teile)
	bewegt := gebaut[von]
	gebaut = slices.Delete(gebaut, von, von+1)

	return slices.Insert(gebaut, nach, bewegt), nil
}

// gewaehlteStellen prüft die Stellen und gibt sie als Menge.
//
// **Eine Menge und keine Liste**, weil MSE eine hat (`const
// set<SymbolPartP>&`): dieselbe Stelle zweimal zu nennen ist ein Fehler des
// Aufrufers, kein Teil, das zweimal in die Gruppe wandert.
func gewaehlteStellen(teile []msepaket.Symbolteil, stellen []int) (map[int]bool, error) {
	gewaehlt := make(map[int]bool, len(stellen))

	for _, stelle := range stellen {
		if stelle < 0 || stelle >= len(teile) {
			return nil, keineTeilstelle(stelle)
		}

		if gewaehlt[stelle] {
			return nil, fmt.Errorf(
				"%w: die Stelle %d steht zweimal in der Wahl",
				domain.ErrValidation, stelle)
		}

		gewaehlt[stelle] = true
	}

	return gewaehlt, nil
}

// keineTeilstelle meldet eine Stelle, die es in der Liste nicht gibt.
func keineTeilstelle(stelle int) error {
	return fmt.Errorf("%w: an Stelle %d steht kein Teil",
		domain.ErrValidation, stelle)
}
