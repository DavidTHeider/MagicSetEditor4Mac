// **Der Griffzug der Symbolnaht** —
// #890.
//
// **Eine eigene Datei**, weil `symbol.go` über die Zeilengrenze wuchs, als der
// Nachlauf des Zweitblicks den Zug von Verschiebungen auf Stellen umstellte
// (§2). Der Schnitt folgt der Frage: dort **die Zeichnung**, hier **der
// einzelne Griff**.

package app

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SymbolgriffDTO nennt, welcher Griff wohin gezogen wird.
type SymbolgriffDTO struct {
	Teil  int `json:"teil"`
	Punkt int `json:"punkt"`

	// Davor sagt, welcher der beiden Griffe gemeint ist.
	Davor bool `json:"davor"`

	// X und Y sind die **neue Stelle der Griffspitze** im Einheitsquadrat,
	// nicht eine Verschiebung.
	//
	// **Und das ist der Unterschied zwischen einem Zug und einem Fehler.**
	// Hier stand bis zum 2026-08-25 eine Verschiebung, und die musste der
	// Aufrufer aus einem Punktwert ausrechnen, den er sich beim Aufsetzen
	// gemerkt hatte. Beim zweiten `pointermove` war der Wert alt: die
	// Differenz wurde gegen den Anfang gerechnet und auf den **schon
	// gezogenen** Griff addiert. Von 0,5 über 0,6 nach 0,7 kam 0,8 heraus.
	//
	// Eine Stelle veraltet nicht. Wo der Zeiger steht, steht der Griff —
	// und die Umrechnung geschieht dort, wo der Punkt in der Hand liegt.
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// SymbolGriffZiehen bewegt einen Bézierhandgriff und zieht den anderen nach.
//
// **Was die Sperre daraus macht, steht in `service`** — `free` lässt den
// anderen stehen, `size` macht ihn zum Gegenteil, `direction` dreht ihn und
// behält seine Länge (`ControlPoint::onUpdateHandle`, `src/data/symbol.cpp`,
// Zeilen 55–62, Fassung f9d9356 vom 2021-07-06).
func SymbolGriffZiehen(teile []SymbolteilDTO, griff SymbolgriffDTO) SymbolAntwort {
	if griff.Teil < 0 || griff.Teil >= len(teile) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinSymbolteil(griff.Teil)}
	}

	umgesetzt := teileAusDTO(teile)

	punkte := umgesetzt[griff.Teil].Punkte
	if griff.Punkt < 0 || griff.Punkt >= len(punkte) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinPunkt(griff.Punkt)}
	}

	// **Die Griffspitze ist der Punkt plus sein Griff.** Was ankommt, ist die
	// Stelle, an der sie stehen soll; die Verschiebung dahin ist die Differenz
	// zu dem, was heute dasteht.
	stelle := punkte[griff.Punkt]
	heute := stelle.GriffDanach
	if griff.Davor {
		heute = stelle.GriffDavor
	}

	punkte[griff.Punkt] = service.GriffZiehen(stelle, griff.Davor,
		msepaket.Vektor{
			X: griff.X - stelle.Stelle.X - heute.X,
			Y: griff.Y - stelle.Stelle.Y - heute.Y,
		})

	return SymbolAntwort{Teile: teileAlsDTO(umgesetzt)}
}

// GriffZiehen bewegt einen Bézierhandgriff.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolGriffZiehen]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) GriffZiehen(teile []SymbolteilDTO, griff SymbolgriffDTO) SymbolAntwort {
	return SymbolGriffZiehen(teile, griff)
}

// SymbolsperreDTO nennt, welcher Punkt welche Art bekommt — #2436.
type SymbolsperreDTO struct {
	Teil  int `json:"teil"`
	Punkt int `json:"punkt"`

	// Sperre ist `free` (frei), `direction` (glatt) oder `size`
	// (symmetrisch) — MSEs drei Knöpfe am Punkteditor
	// (`src/gui/symbol/point_editor.cpp`, Zeilen 160–162, Fassung f9d9356
	// vom 2021-07-06).
	Sperre string `json:"sperre"`
}

// sperren sind die drei Arten eines Punktes.
var sperren = []string{"free", "direction", "size"}

// SymbolSperreSetzen gibt einem Punkt seine Art und mittelt dabei die beiden
// Griffe, wie MSE es tut ([service.SperreSetzen],
// `ControlPoint::onUpdateLock`, `src/data/symbol.cpp`, Zeilen 63–75, Fassung
// f9d9356 vom 2021-07-06).
//
// **Am Fenster gibt es dafür noch keinen Knopf** — 0.9.0 baut eine Fläche
// nach ihrem Artboard (§6), und keines zeigt die Punktart. Die Flächen ohne
// Fenster haben sie über `lesen.symbol.sperre`; `docs/PARITAET.md` führt den
// Knopf als vertagt.
func SymbolSperreSetzen(teile []SymbolteilDTO, wahl SymbolsperreDTO) SymbolAntwort {
	if wahl.Teil < 0 || wahl.Teil >= len(teile) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinSymbolteil(wahl.Teil)}
	}

	umgesetzt := teileAusDTO(teile)

	punkte := umgesetzt[wahl.Teil].Punkte
	if wahl.Punkt < 0 || wahl.Punkt >= len(punkte) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinPunkt(wahl.Punkt)}
	}

	punkte[wahl.Punkt] = service.SperreSetzen(punkte[wahl.Punkt], wahl.Sperre)

	return SymbolAntwort{Teile: teileAlsDTO(umgesetzt)}
}
