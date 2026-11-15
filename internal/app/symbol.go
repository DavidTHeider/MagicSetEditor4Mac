// **Ein Symbol zeichnen und ins Set legen** —
// #890,
// Werkstück 4 von Phase X.
//
// # Was hier passiert und was nicht
//
// Die Rechenseite steht in `internal/service` — Grundformen, Punkte, Ziehen,
// Symmetrie. Hier wird **übersetzt**: Kennungen aus Text, DTOs, Fehler. Ein
// `if` über eine Symbolgestalt gehört nicht hierher (§2).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// keineSymbolablage meldet die fehlende Naht.
func keineSymbolablage() *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: die Symbolablage ist nicht eingerichtet", domain.ErrValidation))
}

// SymbolpunktDTO ist ein Kontrollpunkt, wie die Oberfläche ihn kennt.
//
// **Die Griffe stehen relativ zur Stelle**, so wie in der Datei — wer sie
// absolut reichte, müsste sie bei jedem Ziehen umrechnen, und die Umrechnung
// stünde dann an zwei Stellen.
type SymbolpunktDTO struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`

	GriffDavorX  float64 `json:"griffDavorX"`
	GriffDavorY  float64 `json:"griffDavorY"`
	GriffDanachX float64 `json:"griffDanachX"`
	GriffDanachY float64 `json:"griffDanachY"`

	KurveDanach bool `json:"kurveDanach"`

	// Sperre ist `free`, `direction` oder `size`.
	Sperre string `json:"sperre"`
}

// SymbolteilDTO ist ein Teil einer Zeichnung.
type SymbolteilDTO struct {
	// Art ist `Form`, `Symmetrie` oder `Gruppe`.
	Art  string `json:"art"`
	Name string `json:"name"`

	Kombination string           `json:"kombination"`
	Punkte      []SymbolpunktDTO `json:"punkte"`

	Spiegelart string  `json:"spiegelart"`
	Kopien     int     `json:"kopien"`
	MitteX     float64 `json:"mitteX"`
	MitteY     float64 `json:"mitteY"`
	GriffX     float64 `json:"griffX"`
	GriffY     float64 `json:"griffY"`

	Teile []SymbolteilDTO `json:"teile"`
}

// SymbolAntwort trägt eine Zeichnung oder einen Fehler.
type SymbolAntwort struct {
	Teile []SymbolteilDTO `json:"teile"`

	// Name ist der Dateiname im Set — leer, solange keiner abgelegt ist.
	Name string `json:"name"`

	Fehler *Fehler `json:"fehler"`
}

// Symbolablage ist die Naht zu den gezeichneten Symbolen.
type Symbolablage interface {
	Ablegen(setID int64, zeichnung msepaket.Symbolzeichnung, unter string) (string, error)
	Gestalt(setID int64, name string) (msepaket.Symbolzeichnung, error)
}

// MitSymbolablage setzt die Naht.
func (a *App) MitSymbolablage(ablage Symbolablage) { a.symbole = ablage }

// SymbolSetzen legt eine gezeichnete Zeichnung ins Set und gibt ihren Namen.
//
// **Ohne Kontext.** Es gibt nichts abzuwarten — die Ablage schreibt eine
// Datei und kommt zurück. Ein `ctx`, den niemand liest, wäre eine Zusage über
// Abbrechbarkeit, die es nicht gibt; `Beispielsets` steht aus demselben Grund
// ohne.
//
// **Sie wird nicht gleich als Setsymbol eingetragen.** Das ist ein Feldwert,
// und `SetwertSetzen` schreibt ihn — eine zweite Stelle, die dasselbe Feld
// setzt, liefe mit der ersten auseinander (§7, Regel 4).
// **`unter` nennt die Datei, die überschrieben werden soll.**
//
// Leer heißt: eine neue anlegen. Wer sein Symbol zum zweiten Mal ins Set legt,
// reicht den Namen von vorhin — sonst entstünde `symbol2.mse-symbol`, und die
// erste bliebe als verwaistes Beiwerk im Paket liegen (#890, Befund 1 des
// Nachlaufs).
func (a *App) SymbolSetzen(ctx context.Context, setID string, teile []SymbolteilDTO, unter string) SymbolAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: befund}
	}

	if a.symbole == nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: keineSymbolablage()}
	}

	if fehler := a.symbolsetDa(ctx, kennung); fehler != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	name, err := a.symbole.Ablegen(kennung,
		msepaket.Symbolzeichnung{Teile: teileAusDTO(teile)}, unter)
	if err != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return SymbolAntwort{Teile: nichtNil(teile), Name: name}
}

// Symbol gibt die abgelegte Zeichnung eines Sets.
func (a *App) Symbol(ctx context.Context, setID, name string) SymbolAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: befund}
	}

	if a.symbole == nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: keineSymbolablage()}
	}

	if fehler := a.symbolsetDa(ctx, kennung); fehler != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	zeichnung, err := a.symbole.Gestalt(kennung, name)
	if err != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return SymbolAntwort{Teile: teileAlsDTO(zeichnung.Teile), Name: name}
}

// teileAusDTO wandelt die Oberflächengestalt ins Modell.
func teileAusDTO(teile []SymbolteilDTO) []msepaket.Symbolteil {
	umgesetzt := make([]msepaket.Symbolteil, 0, len(teile))

	for _, teil := range teile {
		umgesetzt = append(umgesetzt, msepaket.Symbolteil{
			Art:         msepaket.Symbolteilart(teil.Art),
			Name:        teil.Name,
			Kombination: msepaket.Kombinationsart(teil.Kombination),
			Punkte:      punkteAusDTO(teil.Punkte),
			Spiegelart:  msepaket.Symmetrieart(teil.Spiegelart),
			Kopien:      teil.Kopien,
			Mitte:       msepaket.Vektor{X: teil.MitteX, Y: teil.MitteY},
			Griff:       msepaket.Vektor{X: teil.GriffX, Y: teil.GriffY},
			Teile:       teileAusDTO(teil.Teile),
		})
	}

	return umgesetzt
}

// punkteAusDTO wandelt die Punkte und leitet `KurveDavor` ab.
//
// **`KurveDavor` steht nicht in der Datei und nicht im DTO.** MSE leitet es ab:
// `p2->segment_before = p1->segment_after` (`enforceConstraints`,
// `src/data/symbol.cpp`, Zeilen 205–212, Fassung f9d9356 vom 2021-07-06). Eine
// Strecke hat **eine** Art, und die schreibt der Punkt davor — wer beide
// Seiten reichen ließe, könnte sie widersprüchlich setzen.
func punkteAusDTO(punkte []SymbolpunktDTO) []msepaket.Kontrollpunkt {
	umgesetzt := make([]msepaket.Kontrollpunkt, 0, len(punkte))

	for _, punkt := range punkte {
		umgesetzt = append(umgesetzt, msepaket.Kontrollpunkt{
			Stelle:      msepaket.Vektor{X: punkt.X, Y: punkt.Y},
			GriffDavor:  msepaket.Vektor{X: punkt.GriffDavorX, Y: punkt.GriffDavorY},
			GriffDanach: msepaket.Vektor{X: punkt.GriffDanachX, Y: punkt.GriffDanachY},
			KurveDanach: punkt.KurveDanach,
			Sperre:      punkt.Sperre,
		})
	}

	for stelle := range umgesetzt {
		davor := (stelle + len(umgesetzt) - 1) % len(umgesetzt)
		umgesetzt[stelle].KurveDavor = umgesetzt[davor].KurveDanach
	}

	return umgesetzt
}

// teileAlsDTO wandelt das Modell in die Oberflächengestalt.
func teileAlsDTO(teile []msepaket.Symbolteil) []SymbolteilDTO {
	umgesetzt := make([]SymbolteilDTO, 0, len(teile))

	for _, teil := range teile {
		umgesetzt = append(umgesetzt, SymbolteilDTO{
			Art:         string(teil.Art),
			Name:        teil.Name,
			Kombination: string(teil.Kombination),
			Punkte:      punkteAlsDTO(teil.Punkte),
			Spiegelart:  string(teil.Spiegelart),
			Kopien:      teil.Kopien,
			MitteX:      teil.Mitte.X, MitteY: teil.Mitte.Y,
			GriffX: teil.Griff.X, GriffY: teil.Griff.Y,
			Teile: teileAlsDTO(teil.Teile),
		})
	}

	return umgesetzt
}

// punkteAlsDTO wandelt die Punkte zurück.
func punkteAlsDTO(punkte []msepaket.Kontrollpunkt) []SymbolpunktDTO {
	umgesetzt := make([]SymbolpunktDTO, 0, len(punkte))

	for _, punkt := range punkte {
		umgesetzt = append(umgesetzt, SymbolpunktDTO{
			X: punkt.Stelle.X, Y: punkt.Stelle.Y,
			GriffDavorX: punkt.GriffDavor.X, GriffDavorY: punkt.GriffDavor.Y,
			GriffDanachX: punkt.GriffDanach.X, GriffDanachY: punkt.GriffDanach.Y,
			KurveDanach: punkt.KurveDanach,
			Sperre:      punkt.Sperre,
		})
	}

	return umgesetzt
}

// SymbolgrundformDTO nennt, welche Grundform gewünscht ist.
type SymbolgrundformDTO struct {
	// Art ist `Kreis`, `Rechteck`, `Vieleck` oder `Stern`.
	Art string `json:"art"`

	// Ecken gilt für Vieleck und Stern.
	Ecken int `json:"ecken"`

	// Eng heißt gleichseitig — Kreis statt Ellipse, Quadrat statt Rechteck.
	Eng bool `json:"eng"`
}

// SymbolGrundform gibt eine der vier Grundformen, mittig und in voller Größe.
//
// **Gerechnet wird in `service`**, portiert aus
// `SymbolBasicShapeEditor::makeCenteredShape` (`src/gui/symbol/basic_shape_editor.cpp`,
// Zeilen 188–297, Fassung f9d9356 vom 2021-07-06). Die Oberfläche sagt nur,
// **welche** — eine Kreisformel im Frontend wäre eine zweite Antwort auf
// dieselbe Frage, und die beiden liefen auseinander.
//
// **Mitte und Größe stehen fest.** MSEs Editor nimmt sie aus dem Mausziehen;
// hier entsteht die Form erst und wird danach an ihren Punkten verändert. Eine
// Form, die 80 % der Fläche einnimmt, ist ein Anfang, den man sieht.
func SymbolGrundform(form SymbolgrundformDTO) SymbolAntwort {
	teil := service.Grundform(service.Formart(form.Art),
		msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.4, Y: 0.4},
		form.Ecken, form.Eng)

	if len(teil.Punkte) == 0 {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: keineGrundform(form.Art)}
	}

	return SymbolAntwort{Teile: teileAlsDTO([]msepaket.Symbolteil{teil})}
}

// keineGrundform meldet eine Art, die es nicht gibt — oder zu wenige Ecken.
func keineGrundform(art string) *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: %q ist keine Grundform, die sich mit dieser Eckenzahl zeichnen lässt",
		domain.ErrValidation, art))
}

// SymbolpunktweiseDTO nennt, was mit einem Punkt geschehen soll.
type SymbolpunktweiseDTO struct {
	// Teil ist die Stelle des Teils in der Zeichnung.
	Teil int `json:"teil"`

	// Punkt ist die Stelle des Punktes in diesem Teil.
	Punkt int `json:"punkt"`
}

// SymbolPunktEinfuegen setzt einen Punkt hinter den genannten, auf die Mitte der
// Strecke zum nächsten.
//
// **Die Mitte und nicht eine Stelle aus dem Aufruf.** MSEs Editor nimmt sie
// aus dem Mausklick auf die Strecke (`ControlPointAddAction`,
// `src/data/action/symbol_part.cpp`, Zeile 253, Fassung f9d9356 vom
// 2021-07-06); über die Tastatur gibt es keinen Klick, und die Mitte ist die
// eine Stelle, die niemand erst wählen muss.
func SymbolPunktEinfuegen(teile []SymbolteilDTO, weise SymbolpunktweiseDTO) SymbolAntwort {
	return punktweise(teile, weise, func(teil msepaket.Symbolteil) (msepaket.Symbolteil, bool) {
		return service.PunktEinfuegen(teil, weise.Punkt, 0.5)
	})
}

// SymbolPunktLoeschen entfernt den genannten Punkt.
//
// **Unter drei Punkten geht es nicht** — die Grenze steht in
// [service.PunktLoeschen] und kommt von MSE.
func SymbolPunktLoeschen(teile []SymbolteilDTO, weise SymbolpunktweiseDTO) SymbolAntwort {
	return punktweise(teile, weise, func(teil msepaket.Symbolteil) (msepaket.Symbolteil, bool) {
		return service.PunktLoeschen(teil, weise.Punkt)
	})
}

// punktweise wendet eine Punktänderung auf **ein** Teil an.
//
// **Die Zeichnung geht hin und zurück, statt im Fenster zu liegen.** Sie ist
// klein — ein Symbol hat Dutzende Punkte —, und die Rechnung gehört nach
// `service` (§2). Ein Zustand hier wäre ein zweiter neben dem der Seite, und
// zwei Zustände laufen auseinander.
func punktweise(
	teile []SymbolteilDTO, weise SymbolpunktweiseDTO,
	aendere func(msepaket.Symbolteil) (msepaket.Symbolteil, bool),
) SymbolAntwort {
	if weise.Teil < 0 || weise.Teil >= len(teile) {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinSymbolteil(weise.Teil)}
	}

	umgesetzt := teileAusDTO(teile)

	geaendert, gut := aendere(umgesetzt[weise.Teil])
	if !gut {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: keinPunkt(weise.Punkt)}
	}

	umgesetzt[weise.Teil] = geaendert

	return SymbolAntwort{Teile: teileAlsDTO(umgesetzt)}
}

// keinSymbolteil meldet eine Teilstelle, die es nicht gibt.
func keinSymbolteil(stelle int) *Fehler {
	return Nutzerfehler(fmt.Errorf("%w: die Zeichnung hat kein Teil an Stelle %d",
		domain.ErrValidation, stelle))
}

// keinPunkt meldet einen Punkt, der sich nicht ändern ließ.
//
// **Beides in einer Meldung**, weil der Aufrufer beides gleich behandelt: eine
// Stelle außerhalb und ein Löschen, das zu wenige Punkte übrig ließe, enden
// gleichermaßen darin, dass sich nichts ändert.
func keinPunkt(stelle int) *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: der Punkt %d lässt sich so nicht ändern — eine Form braucht mindestens zwei",
		domain.ErrValidation, stelle))
}

// Grundform gibt eine der vier Grundformen, mittig und in voller Größe.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolGrundform]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) Grundform(form SymbolgrundformDTO) SymbolAntwort {
	return SymbolGrundform(form)
}

// PunktEinfuegen setzt einen Punkt in die Mitte der Strecke dahinter.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolPunktEinfuegen]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) PunktEinfuegen(teile []SymbolteilDTO, weise SymbolpunktweiseDTO) SymbolAntwort {
	return SymbolPunktEinfuegen(teile, weise)
}

// PunktLoeschen entfernt einen Punkt.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolPunktLoeschen]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) PunktLoeschen(teile []SymbolteilDTO, weise SymbolpunktweiseDTO) SymbolAntwort {
	return SymbolPunktLoeschen(teile, weise)
}
