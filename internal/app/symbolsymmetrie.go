// **Die Symmetrie im Symboleditor**
// (#1092).
//
// # Was fehlte
//
// `service.SymmetrieAbbilden` und `service.SymmetriekopieN` waren portiert,
// geprüft und hatten **keinen produktiven Aufrufer**. Ein Paket ohne Aufrufer
// ist keine Anwendungsfunktion — §10 nennt das zuerst.
//
// Der Editor bot weder Art noch Kopienzahl an und zeichnete keine Kopien: wer
// ein Symbol mit Symmetrie öffnete, sah die Teile darin und konnte die
// Symmetrie selbst nicht anfassen.
//
// # Warum die Kopien von hier kommen und nicht aus dem Fenster
//
// Weil die Rechnung dahinter drei Fallen hat, die man ohne den C++-Quelltext
// verfehlt: die Multiplikationsrichtung, die Spiegelachse bei `(a + b)/2` und
// das Abrunden der Kopienzahl auf eine gerade. Sie steht **einmal** in
// `service`; eine zweite Fassung in TypeScript liefe auseinander, und dann
// zeichnete der Editor etwas anderes als der Rasterer.

package app

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SymmetrieweiseDTO sagt, welche Symmetrie geändert werden soll und wie.
type SymmetrieweiseDTO struct {
	// Teil ist die Stelle in der Teilliste.
	Teil int `json:"teil"`

	// Spiegelart ist `rotation` oder `reflection` — leer lässt sie stehen.
	Spiegelart string `json:"spiegelart"`

	// Kopien ist die gewünschte Zahl. **0 lässt sie stehen**, damit ein
	// Aufruf, der nur die Mitte verschiebt, sie nicht auf null setzt.
	Kopien int `json:"kopien"`

	MitteX float64 `json:"mitteX"`
	MitteY float64 `json:"mitteY"`
	GriffX float64 `json:"griffX"`
	GriffY float64 `json:"griffY"`

	// MitteSetzen und GriffSetzen sagen, ob die vier Zahlen darüber gelten.
	//
	// **Zwei Wahrheitswerte statt Zeigern**: die 0 ist eine gültige Stelle —
	// die linke obere Ecke —, und ein Aufruf, der nur die Art ändert, dürfte
	// die Mitte nicht dorthin ziehen.
	MitteSetzen bool `json:"mitteSetzen"`
	GriffSetzen bool `json:"griffSetzen"`
}

// SymmetriekopienAntwort trägt die Punkte jeder Kopie.
//
// **Als fertige Punktlisten und nicht als Anweisung**, wie sie zu rechnen
// wäre: die Zeichenfläche malt, was ankommt.
type SymmetriekopienAntwort struct {
	// Kopien sind die Züge — je Kopie eine Liste aus Punkten je Kindteil.
	//
	// **Die erste Kopie ist das Original.** MSE zeichnet sie mit, weil die
	// Abbildung für `kopie == 0` die Identität ist; wer sie ausließe, hätte
	// bei zwei Kopien eine.
	Kopien [][]SymbolzugDTO `json:"kopien"`

	Fehler *Fehler `json:"fehler"`
}

// SymbolzugDTO ist ein geschlossener Zug aus Punkten — eine Kopie eines
// Kindteils.
type SymbolzugDTO struct {
	Punkte []SymbolpunktDTO `json:"punkte"`
}

// HoechsteKopien ist die Obergrenze, die der Editor anbietet.
//
// **Zehn, wie MSE** (`SymmetryEditor`, `src/gui/symbol/symmetry_editor.cpp`,
// Fassung f9d9356 vom 2021-07-06): sein Drehfeld läuft von 2 bis 10. Mehr
// wäre auf einem Symbol von 500 Punkten Kantenlänge nicht zu unterscheiden.
const HoechsteKopien = 10

// NiedrigsteKopien ist die Untergrenze — **zwei**.
//
// Eine Symmetrie mit einer Kopie ist keine: sie bildet auf sich selbst ab.
const NiedrigsteKopien = 2

// SymbolSymmetrieAnlegen wandelt ein Teil in eine Symmetrie um.
//
// **Es umschließt das gewählte Teil, statt ein leeres anzulegen.** MSE macht
// es ebenso: eine Symmetrie ohne Kinder zeichnet nichts, und ein Editor, der
// sie anböte, ließe jemanden auf eine Wirkung warten, die nicht kommt.
func SymbolSymmetrieAnlegen(teile []SymbolteilDTO, stelle int) SymbolAntwort {
	if stelle < 0 || stelle >= len(teile) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinSymbolteil(stelle)}
	}

	umgesetzt := teileAusDTO(teile)
	kind := umgesetzt[stelle]

	umgesetzt[stelle] = msepaket.Symbolteil{
		Art:  msepaket.TeilSymmetrie,
		Name: kind.Name,

		// **Drehung mit zwei Kopien** — dieselbe Vorgabe wie bei MSE
		// (`SymbolSymmetry::SymbolSymmetry`, `src/data/symbol.cpp`, Fassung
		// f9d9356 vom 2021-07-06).
		Spiegelart: msepaket.SpiegelungDrehung,
		Kopien:     NiedrigsteKopien,

		// **Die Mitte in der Mitte, der Griff nach rechts.** Ein Griff im
		// Ursprung hätte den Winkel null und wäre nicht zu fassen.
		Mitte: msepaket.Vektor{X: 0.5, Y: 0.5},
		Griff: msepaket.Vektor{X: 0.25, Y: 0},

		Teile: []msepaket.Symbolteil{kind},
	}

	return SymbolAntwort{Teile: teileAlsDTO(umgesetzt)}
}

// SymbolSymmetrieAufloesen macht aus einer Symmetrie ihre Kinder.
//
// **Die Kinder bleiben stehen.** Eine Symmetrie zu löschen und ihre Formen
// mitzunehmen wäre ein Datenverlust, den niemand angeordnet hat; wer die
// Formen loswerden will, löscht sie danach einzeln.
func SymbolSymmetrieAufloesen(teile []SymbolteilDTO, stelle int) SymbolAntwort {
	if stelle < 0 || stelle >= len(teile) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinSymbolteil(stelle)}
	}

	umgesetzt := teileAusDTO(teile)
	if umgesetzt[stelle].Art != msepaket.TeilSymmetrie {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keineSymmetrie(stelle)}
	}

	kinder := umgesetzt[stelle].Teile

	ohne := make([]msepaket.Symbolteil, 0, len(umgesetzt)+len(kinder))
	ohne = append(ohne, umgesetzt[:stelle]...)
	ohne = append(ohne, kinder...)
	ohne = append(ohne, umgesetzt[stelle+1:]...)

	return SymbolAntwort{Teile: teileAlsDTO(ohne)}
}

// SymbolSymmetrieSetzen ändert Art, Kopienzahl, Mitte oder Griff.
func SymbolSymmetrieSetzen(
	teile []SymbolteilDTO, weise SymmetrieweiseDTO,
) SymbolAntwort {
	if weise.Teil < 0 || weise.Teil >= len(teile) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinSymbolteil(weise.Teil)}
	}

	umgesetzt := teileAusDTO(teile)

	spiegel := &umgesetzt[weise.Teil]
	if spiegel.Art != msepaket.TeilSymmetrie {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keineSymmetrie(weise.Teil)}
	}

	if befund := symmetrieAendern(spiegel, weise); befund != nil {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: befund}
	}

	return SymbolAntwort{Teile: teileAlsDTO(umgesetzt)}
}

// symmetrieAendern setzt, was der Aufruf nennt.
func symmetrieAendern(spiegel *msepaket.Symbolteil, weise SymmetrieweiseDTO) *Fehler {
	if weise.Spiegelart != "" {
		art := msepaket.Symmetrieart(weise.Spiegelart)
		if art != msepaket.SpiegelungDrehung && art != msepaket.SpiegelungAchse {
			return Nutzerfehler(fmt.Errorf(
				"%w: %q ist keine Symmetrieart — es gibt »rotation« und »reflection«",
				domain.ErrValidation, weise.Spiegelart))
		}

		spiegel.Spiegelart = art
	}

	if weise.Kopien != 0 {
		if weise.Kopien < NiedrigsteKopien || weise.Kopien > HoechsteKopien {
			return Nutzerfehler(fmt.Errorf(
				"%w: %d Kopien gehen nicht — es sind %d bis %d",
				domain.ErrValidation, weise.Kopien, NiedrigsteKopien, HoechsteKopien))
		}

		spiegel.Kopien = weise.Kopien
	}

	if weise.MitteSetzen {
		spiegel.Mitte = msepaket.Vektor{X: weise.MitteX, Y: weise.MitteY}
	}

	if weise.GriffSetzen {
		spiegel.Griff = msepaket.Vektor{X: weise.GriffX, Y: weise.GriffY}
	}

	return nil
}

// SymbolSymmetriekopien gibt die Züge, die eine Symmetrie zeichnet.
//
// **Der produktive Aufrufer von [service.SymmetrieAbbilden]** — der, der
// #1092 gefehlt hat.
func SymbolSymmetriekopien(
	teile []SymbolteilDTO, stelle int,
) SymmetriekopienAntwort {
	leer := SymmetriekopienAntwort{Kopien: [][]SymbolzugDTO{}}

	if stelle < 0 || stelle >= len(teile) {
		leer.Fehler = keinSymbolteil(stelle)

		return leer
	}

	spiegel := teileAusDTO(teile)[stelle]
	if spiegel.Art != msepaket.TeilSymmetrie {
		leer.Fehler = keineSymmetrie(stelle)

		return leer
	}

	kopien := service.SymmetriekopieN(spiegel)
	leer.Kopien = make([][]SymbolzugDTO, 0, kopien)

	for kopie := range kopien {
		leer.Kopien = append(leer.Kopien, abgebildeteZuege(spiegel, kopie))
	}

	return leer
}

// abgebildeteZuege bildet jedes Kindteil auf eine Kopie ab.
//
// **Über die Punkte und nicht über die Griffe.** Eine Kurve dreht sich mit
// ihren Griffen; MSE bildet sie im Zeichenweg mit ab (`SymbolViewer::draw`,
// `src/render/symbol/viewer.cpp`, Zeilen 150–190, Fassung f9d9356 vom
// 2021-07-06). Hier steht der **Umriss** für die Zeichenfläche, und für ihn
// zählen die Stellen.
func abgebildeteZuege(spiegel msepaket.Symbolteil, kopie int) []SymbolzugDTO {
	zuege := make([]SymbolzugDTO, 0, len(spiegel.Teile))

	for _, kind := range spiegel.Teile {
		punkte := make([]SymbolpunktDTO, 0, len(kind.Punkte))

		for _, punkt := range kind.Punkte {
			abgebildet := service.SymmetrieAbbilden(spiegel, kopie, punkt.Stelle)

			punkte = append(punkte, SymbolpunktDTO{
				X: abgebildet.X, Y: abgebildet.Y,
				KurveDanach: punkt.KurveDanach,
			})
		}

		zuege = append(zuege, SymbolzugDTO{Punkte: punkte})
	}

	return zuege
}

// keineSymmetrie meldet ein Teil, das keine ist.
func keineSymmetrie(stelle int) *Fehler {
	return Nutzerfehler(fmt.Errorf("%w: das Teil an Stelle %d ist keine Symmetrie",
		domain.ErrValidation, stelle))
}

// SymmetrieAnlegen wandelt ein Teil in eine Symmetrie um.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolSymmetrieAnlegen]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) SymmetrieAnlegen(teile []SymbolteilDTO, stelle int) SymbolAntwort {
	return SymbolSymmetrieAnlegen(teile, stelle)
}

// SymmetrieSetzen ändert Art, Kopienzahl, Mitte oder Griff einer Symmetrie.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolSymmetrieSetzen]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) SymmetrieSetzen(teile []SymbolteilDTO, weise SymmetrieweiseDTO) SymbolAntwort {
	return SymbolSymmetrieSetzen(teile, weise)
}

// SymmetrieAufloesen macht aus einer Symmetrie ihre Kinder.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolSymmetrieAufloesen]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) SymmetrieAufloesen(teile []SymbolteilDTO, stelle int) SymbolAntwort {
	return SymbolSymmetrieAufloesen(teile, stelle)
}

// Symmetriekopien gibt die gespiegelten Züge einer Symmetrie.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolSymmetriekopien]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) Symmetriekopien(teile []SymbolteilDTO, stelle int) SymmetriekopienAntwort {
	return SymbolSymmetriekopien(teile, stelle)
}
