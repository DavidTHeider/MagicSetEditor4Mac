package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Zwei Typen für dieselbe Gestalt, und das ist die Schichtenregel** (§2):
// `internal/render` kennt nichts aus diesem Projekt, also führt es seine
// eigenen Formen; `msepaket` liest die Datei und kennt `render` ebenso wenig.
// Diese Datei ist die Naht — und sie liegt in `service`, wo beide zusammenlaufen
// (#323).

// alsZeichenteile übersetzt die gelesene Gestalt in die des Rasterers.
//
// **Zwei Typen für dieselbe Gestalt, und das ist die Schichtenregel** (§2):
// `internal/render` kennt nichts aus diesem Projekt, also führt es seine
// eigenen Formen. Diese Stelle ist die Naht.
func alsZeichenteile(teile []msepaket.Symbolteil) []render.Symbolteil {
	umgesetzt := make([]render.Symbolteil, 0, len(teile))

	for i := range teile {
		teil := &teile[i]

		umgesetzt = append(umgesetzt, render.Symbolteil{
			Art:         zeichenarten[teil.Art],
			Kombination: zeichenkombinationen[teil.Kombination],
			Punkte:      alsZeichenpunkte(teil.Punkte),
			Spiegelart:  zeichenspiegelungen[teil.Spiegelart],
			Kopien:      teil.Kopien,
			Mitte:       render.Punkt{X: teil.Mitte.X, Y: teil.Mitte.Y},
			Griff:       render.Punkt{X: teil.Griff.X, Y: teil.Griff.Y},
			Teile:       alsZeichenteile(teil.Teile),
		})
	}

	return umgesetzt
}

// alsZeichenpunkte übersetzt die Kontrollpunkte einer Form.
func alsZeichenpunkte(punkte []msepaket.Kontrollpunkt) []render.Symbolpunkt {
	umgesetzt := make([]render.Symbolpunkt, 0, len(punkte))

	for _, punkt := range punkte {
		umgesetzt = append(umgesetzt, render.Symbolpunkt{
			Stelle:      render.Punkt{X: punkt.Stelle.X, Y: punkt.Stelle.Y},
			GriffDavor:  render.Punkt{X: punkt.GriffDavor.X, Y: punkt.GriffDavor.Y},
			GriffDanach: render.Punkt{X: punkt.GriffDanach.X, Y: punkt.GriffDanach.Y},
			KurveDanach: punkt.KurveDanach,
			KurveDavor:  punkt.KurveDavor,
		})
	}

	return umgesetzt
}

var (
	zeichenarten = map[msepaket.Symbolteilart]render.Symbolteilart{
		msepaket.TeilForm: render.TeilForm, msepaket.TeilSymmetrie: render.TeilSymmetrie,
		msepaket.TeilGruppe: render.TeilGruppe,
	}

	zeichenkombinationen = map[msepaket.Kombinationsart]render.Kombinationsart{
		msepaket.KombinationUeberdecken: render.KombinationUeberdecken,
		msepaket.KombinationVereinen:    render.KombinationVereinen,
		msepaket.KombinationAbziehen:    render.KombinationAbziehen,
		msepaket.KombinationSchneiden:   render.KombinationSchneiden,
		msepaket.KombinationUnterschied: render.KombinationUnterschied,
		msepaket.KombinationUmranden:    render.KombinationUmranden,
	}

	zeichenspiegelungen = map[msepaket.Symmetrieart]render.Symmetrieart{
		msepaket.SpiegelungDrehung: render.SpiegelungDrehung,
		msepaket.SpiegelungAchse:   render.SpiegelungAchse,
	}
)
