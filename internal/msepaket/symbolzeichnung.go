package msepaket

import (
	"fmt"
	"strconv"
	"strings"
)

// **Eine `.mse-symbol` ist eine Vektorzeichnung.**
//
// Dieselbe Syntax wie jedes andere MSE-Paket — deshalb steht hier kein
// Zerteiler, sondern nur die **Bedeutung** der Blöcke. Sie ist an MSE2
// nachgelesen (`src/data/symbol.cpp`, Zeilen 17–25, 126–166, 153–166 und
// 225–285, Fassung f9d9356 vom 2021-07-06); `docs/FORMATE.md` führt sie
// ausführlich (#323).
//
// # Was hier gelesen wird und was nicht
//
// Gelesen wird die **Gestalt**: Formen, ihre Punkte, ihre Symmetrien. Das
// **Rastern** steht daneben — Bézier, Füllregel, Symmetrien ausführen,
// `combine:` verrechnen — und gehört nach `internal/render`, denn `msepaket`
// zeichnet nichts (§2).

// Vektor ist ein Punkt oder eine Verschiebung in der Zeichnung.
//
// **Auf [0,1] normiert.** Die Zielgröße kommt vom Aufrufer, nicht aus der
// Datei; ein Symbol wird in den Kasten gerechnet, den der Stil ihm gibt.
type Vektor struct{ X, Y float64 }

// Symbolteilart sagt, was ein `part:`-Block ist.
type Symbolteilart string

const (
	// TeilForm ist `type: shape` — ein **geschlossener** Zug aus Punkten.
	//
	// **Auch ohne `type:`.** MSE liest einen leeren Typ als `shape`
	// (`read_new<SymbolPart>`, Zeile 139).
	TeilForm Symbolteilart = "Form"

	// TeilSymmetrie ist `type: symmetry` — es vervielfältigt seine Kinder.
	TeilSymmetrie Symbolteilart = "Symmetrie"

	// TeilGruppe ist `type: group` — es fasst Kinder zusammen, ohne sie zu
	// ändern.
	TeilGruppe Symbolteilart = "Gruppe"
)

// Kombinationsart sagt, wie eine Form mit dem verrechnet wird, was schon
// dasteht — `combine:`.
//
// **Der Nullwert ist `overlap`**, so wie bei MSE (`SymbolShapeCombine`, Zeilen
// 153–160). Was die einzelnen Arten beim Rastern bedeuten, ist hier
// ausdrücklich **nicht** entschieden; gelesen wird der Name.
type Kombinationsart string

const (
	KombinationUeberdecken Kombinationsart = "overlap"
	KombinationVereinen    Kombinationsart = "merge"
	KombinationAbziehen    Kombinationsart = "subtract"
	KombinationSchneiden   Kombinationsart = "intersection"
	KombinationUnterschied Kombinationsart = "difference"
	KombinationUmranden    Kombinationsart = "border"
)

// Symmetrieart ist `kind:` einer Symmetrie.
type Symmetrieart string

const (
	SpiegelungDrehung Symmetrieart = "rotation"
	SpiegelungAchse   Symmetrieart = "reflection"
)

// Kontrollpunkt ist ein `point:`-Block.
//
// **Die Griffe sind relativ zur Stelle**, nicht absolut — `delta_before` und
// `delta_after` bei MSE (`ControlPoint`, Zeilen 96–100).
type Kontrollpunkt struct {
	Stelle      Vektor
	GriffDavor  Vektor
	GriffDanach Vektor

	// KurveDanach sagt, dass die Strecke **zum nächsten** Punkt eine Bézierkurve
	// ist — `line after: curve`.
	KurveDanach bool

	// KurveDavor gilt für die Strecke vom **vorigen** Punkt.
	//
	// **Sie steht nicht in der Datei.** MSE leitet sie ab:
	// `p2->segment_before = p1->segment_after` (`enforceConstraints`, Zeilen
	// 205–212). Eine Strecke hat eine Art, und die schreibt der Punkt davor.
	KurveDavor bool

	// Sperre ist `lock:` — sie bindet die beiden Griffe aneinander, während
	// jemand am Symbol **zeichnet**. Für das Rastern ist sie ohne Belang und
	// steht hier nur, damit sie nicht verloren geht.
	Sperre string
}

// Symbolteil ist ein `part:`-Block.
type Symbolteil struct {
	Art  Symbolteilart
	Name string

	// Kombination und Punkte gelten für [TeilForm].
	Kombination Kombinationsart
	Punkte      []Kontrollpunkt

	// Spiegelart, Kopien, Mitte und Griff gelten für [TeilSymmetrie].
	Spiegelart Symmetrieart
	Kopien     int
	Mitte      Vektor
	Griff      Vektor

	// Teile sind die Kinder einer Symmetrie oder einer Gruppe.
	Teile []Symbolteil
}

// Symbolzeichnung ist der Inhalt einer `.mse-symbol`.
type Symbolzeichnung struct{ Teile []Symbolteil }

// LiesSymbolzeichnung liest eine `.mse-symbol` von der Platte.
//
// **Eine Datei und kein Paket.** Anders als `.mse-style` oder `.mse-game` ist
// eine `.mse-symbol` eine einzelne Datei, die in einem Stilpaket liegt —
// [Oeffne] und sein Dateibaum passen darauf nicht.
func LiesSymbolzeichnung(pfad string) (Symbolzeichnung, []Befund, error) {
	// **Mit der Grenze eines Symbols** und nicht der eines Pakets — DOM-02.AK-04.
	roh, err := leseSymboldatei(pfad)
	if err != nil {
		return Symbolzeichnung{}, nil, fmt.Errorf("die Symbolzeichnung %s: %w", pfad, err)
	}

	return SymbolzeichnungAus(pfad, roh)
}

// Symbolzeichnungsmodell deutet einen gelesenen Baum als Zeichnung.
//
// **Getrennt vom Lesen**, weil die Deutung prüfbar ist und der Dateizugriff
// nicht: derselbe Schnitt wie zwischen [Zerteile] und den übrigen Modellen.
func Symbolzeichnungsmodell(baum *Baum) (Symbolzeichnung, []Befund) {
	var (
		zeichnung Symbolzeichnung
		befunde   []Befund
	)

	for _, knoten := range baum.Alle("part") {
		teil, befund := symbolteilAus(knoten)
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		zeichnung.Teile = append(zeichnung.Teile, teil)
	}

	return zeichnung, befunde
}

// symbolteilAus deutet einen `part:`-Block.
func symbolteilAus(knoten *Knoten) (Symbolteil, *Befund) {
	art, bekannt := teilarten[strings.TrimSpace(wertOderLeer(knoten.Ersten("type")))]
	if !bekannt {
		return Symbolteil{}, &Befund{
			Stelle: knoten.Stelle,
			Meldung: "ein Symbolteil vom Typ " +
				wertOderLeer(knoten.Ersten("type")) + " — den kennt dieses Programm nicht",
		}
	}

	// **Der Name auch als Block** (C02, #2227, `FuzzSymbolzeichnung`): ein
	// Name, der eine Klammer offen lässt, steht geschrieben als Block.
	teil := Symbolteil{Art: art, Name: WertOderBlock(knoten.Ersten("name"))}

	switch art {
	case TeilForm:
		teil.Kombination = kombinationAus(knoten)
		teil.Punkte = punkteAus(knoten)

	case TeilSymmetrie:
		teil.Spiegelart = Symmetrieart(strings.TrimSpace(wertOderLeer(knoten.Ersten("kind"))))
		teil.Kopien = ganzzahlOderNull(knoten.Ersten("copies"))
		teil.Mitte = vektorAus(knoten.Ersten("center"))
		teil.Griff = vektorAus(knoten.Ersten("handle"))
		teil.Teile = kinderAus(knoten)

	case TeilGruppe:
		teil.Teile = kinderAus(knoten)
	}

	return teil, nil
}

// teilarten ordnet `type:` einer Teilart zu.
//
// **Der leere Typ ist eine Form** (`read_new<SymbolPart>`, Zeile 139).
var teilarten = map[string]Symbolteilart{
	"":         TeilForm,
	"shape":    TeilForm,
	"symmetry": TeilSymmetrie,
	"group":    TeilGruppe,
}

// kinderAus liest die `part:`-Blöcke unter einem Knoten.
//
// **Ein unbrauchbares Kind fällt weg, ohne den Rest mitzunehmen** (§8). Sein
// Befund geht dabei verloren, und das ist die Abwägung: ein Symbol mit einem
// fremden Teil zeichnet fast richtig, eine Befundkette durch jede
// Verschachtelung machte den Leser zum Sammler.
func kinderAus(knoten *Knoten) []Symbolteil {
	var teile []Symbolteil

	for _, kind := range knoten.Alle("part") {
		if teil, befund := symbolteilAus(kind); befund == nil {
			teile = append(teile, teil)
		}
	}

	return teile
}

// kombinationAus liest `combine:`; ohne Angabe gilt `overlap`.
func kombinationAus(knoten *Knoten) Kombinationsart {
	roh := strings.TrimSpace(wertOderLeer(knoten.Ersten("combine")))
	if roh == "" {
		return KombinationUeberdecken
	}

	return Kombinationsart(roh)
}

// punkteAus liest die `point:`-Blöcke einer Form.
//
// **Die Kurve davor wird abgeleitet**, nicht gelesen: sie ist die Kurve danach
// des vorigen Punktes, und beim ersten die des letzten — der Zug ist
// geschlossen (`enforceConstraints`, Zeilen 205–212).
func punkteAus(knoten *Knoten) []Kontrollpunkt {
	kinder := knoten.Alle("point")
	punkte := make([]Kontrollpunkt, 0, len(kinder))

	for _, kind := range kinder {
		punkte = append(punkte, Kontrollpunkt{
			Stelle:      vektorAus(kind.Ersten("position")),
			GriffDavor:  vektorAus(kind.Ersten("handle before")),
			GriffDanach: vektorAus(kind.Ersten("handle after")),
			KurveDanach: strings.TrimSpace(wertOderLeer(kind.Ersten("line after"))) == "curve",
			Sperre:      strings.TrimSpace(wertOderLeer(kind.Ersten("lock"))),
		})
	}

	for stelle := range punkte {
		davor := (stelle + len(punkte) - 1) % len(punkte)
		punkte[stelle].KurveDavor = punkte[davor].KurveDanach
	}

	return punkte
}

// vektorAus liest ein `(x,y)`.
//
// **Was nicht so aussieht, ist der Nullpunkt.** Ein halb gelesener Vektor wäre
// eine Zahl, die aussieht wie eine gemessene.
func vektorAus(knoten *Knoten) Vektor {
	if knoten == nil {
		return Vektor{}
	}

	roh := strings.TrimSpace(knoten.Wert)

	innen, auf := strings.CutPrefix(roh, "(")
	if !auf {
		return Vektor{}
	}

	innen, zu := strings.CutSuffix(innen, ")")
	if !zu {
		return Vektor{}
	}

	teile := strings.SplitN(innen, ",", 2)
	if len(teile) != 2 {
		return Vektor{}
	}

	x, fehlerX := strconv.ParseFloat(strings.TrimSpace(teile[0]), 64)
	y, fehlerY := strconv.ParseFloat(strings.TrimSpace(teile[1]), 64)

	if fehlerX != nil || fehlerY != nil {
		return Vektor{}
	}

	return Vektor{X: x, Y: y}
}

// ganzzahlOderNull liest eine ganze Zahl aus einem Knoten.
//
// **Eine unlesbare ist eine Null und kein Fehler** (§8): fremde Daten sind der
// Normalfall, und ein Paket von 2007 mit `card list column: viele` soll sich
// öffnen lassen.
//
// **Und eine führende Null ist keine Achtelzahl.** Das Pack schreibt
// `card list column: 0400`; `strconv.Atoi` liest zur Basis zehn, und das ist
// hier der Punkt — eine Auflösung zur Basis null läse oktal, und die Spalte
// stünde an einer Stelle, die niemand gemeint hat (#1168).
func ganzzahlOderNull(knoten *Knoten) int {
	if knoten == nil {
		return 0
	}

	zahl, err := strconv.Atoi(strings.TrimSpace(knoten.Wert))
	if err != nil {
		return 0
	}

	return zahl
}
