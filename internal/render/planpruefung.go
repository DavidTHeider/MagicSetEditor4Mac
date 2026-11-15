package render

import (
	"errors"
	"fmt"
)

// ErrPlan meldet einen Plan, der so nicht zu zeichnen ist.
var ErrPlan = errors.New("der Zeichenplan ist nicht brauchbar")

// Befund ist ein benannter Vorbehalt, kein Fehler.
//
// Ein Plan mit zwölf Ebenen, von denen eine ein fehlendes Bild nennt, wird mit
// elf Ebenen und einem Befund gezeichnet — nicht abgewiesen (AGENTS.md §8). Die
// Stelle ist die Ebenennummer: ohne sie sucht jemand in einer Vorlage mit 60
// Ebenen von Hand.
type Befund struct {
	Ebene int

	// Feld ist der Name des Kartenfeldes, aus dem die Ebene entstand — leer,
	// wo keines dahintersteht.
	//
	// **Leer ist ein Fall, kein Mangel**: der weiße Grund, den der Planbauer
	// selbst hinzufügt, gehört zu keinem Feld. Dann steht nur die Nummer da.
	//
	// „Ebene 11: das Bild … ist nicht lesbar" schickte jemanden auf die Suche —
	// bei rund vierzig Ebenen je Karte
	// (#246).
	Feld string

	Meldung string

	// Art sagt, **wem** der Befund gehört — dieselbe Frage wie
	// `domain.Befundart`, und derselbe Nullwert: „Portierung"
	// (#513).
	//
	// **Als eigener Typ und nicht als `domain.Befundart`**: `render` kennt
	// `domain` nicht (§2). `internal/app` bildet die beiden aufeinander ab —
	// dieselbe Naht wie bei jeder anderen Übersetzung.
	Art Befundart
}

// Befundart sagt, wem ein Befund gehört. Die Begründung steht bei
// `domain.Befundart`; hier steht sie zum zweiten Mal, weil `render` nicht
// dorthin sehen darf.
type Befundart string

const (
	// ArtPortierung ist die Vorgabe: dieses Programm kann etwas nicht.
	ArtPortierung Befundart = "Portierung"

	// ArtVorlage heißt: die Vorlage rechnet ins Leere, und MSE täte es
	// genauso.
	ArtVorlage Befundart = "Vorlage"

	// ArtUmgebung heißt: es hängt an diesem Rechner.
	ArtUmgebung Befundart = "Umgebung"
)

func (b Befund) String() string {
	// **Die Nummer bleibt neben dem Namen.** Sie ist die einzige Angabe, die
	// eine Ebene eindeutig macht: ein Feld kann mehrere ergeben — `magic-m15`
	// zeichnet für `card color` vier.
	if b.Feld != "" {
		return fmt.Sprintf("Ebene %d (%s): %s", b.Ebene, b.Feld, b.Meldung)
	}

	return fmt.Sprintf("Ebene %d: %s", b.Ebene, b.Meldung)
}

// Pruefe sagt, was an einem Plan nicht stimmt.
//
// # Zwei Arten von Mangel, und sie werden verschieden behandelt
//
// Ein **Fehler** macht den Plan unzeichenbar: eine Kartengröße von null hat
// keine sinnvolle Auslegung. Ein **Befund** betrifft eine Ebene: sie fällt weg,
// der Rest wird gezeichnet.
//
// Die Trennung ist nicht kosmetisch. Wer beides als Fehler behandelt, zeigt bei
// einem fehlenden Bild eine leere Karte; wer beides als Befund behandelt,
// zeichnet auf eine Fläche der Größe null und wundert sich über ein leeres
// Bild ohne Meldung.
func Pruefe(plan *Plan) ([]Befund, error) {
	if plan == nil {
		return nil, fmt.Errorf("%w: es gibt keinen", ErrPlan)
	}

	if plan.Breite <= 0 || plan.Hoehe <= 0 {
		return nil, fmt.Errorf("%w: die Karte misst %g×%g", ErrPlan, plan.Breite, plan.Hoehe)
	}

	var befunde []Befund

	for nummer := range plan.Ebenen {
		if meldung := ebenenmangel(plan.Ebenen[nummer]); meldung != "" {
			befunde = append(befunde, Befund{
				Ebene: nummer + 1, Feld: plan.Ebenen[nummer].Feld, Meldung: meldung,
			})
		}
	}

	return befunde, nil
}

// ebenenmangel gibt den Grund, warum eine Ebene nicht zu zeichnen ist.
func ebenenmangel(ebene Ebene) string {
	switch ebene.Art {
	case ArtFlaeche:
		return flaechenmangel(ebene)

	case ArtBild, ArtSymbol:
		if ebene.Bild.Leer() {
			return string(ebene.Art) + " ohne Bild"
		}

	case ArtSymbolzug:
		// **Ein Symbolzug ohne Symbole ist ein Mangel**, dieselbe Zusage wie
		// bei einer Bildebene ohne Bild. Seine *Größe* wird hier nicht geprüft:
		// sie hängt am Bild, und Bilder kennt erst der Rasterer.
		if symbole, _ := zugAus(ebene.Abschnitte); len(symbole) == 0 {
			return "ein Symbolzug ohne Symbole"
		}

	case ArtText:
		if ebene.Satz.Groesse <= 0 {
			return fmt.Sprintf("Schriftgröße %g", ebene.Satz.Groesse)
		}

		// **Ein Rahmen ohne Fläche wird nicht gezeichnet.** Ohne diese Prüfung
		// nahm `absatzUmbrechen` das erste Wort unabhängig von seiner Breite an
		// und setzte jedes weitere in eine eigene, überstehende Zeile — ein
		// Text, der aussieht wie umgebrochen und keiner ist. Angestrichen vom
		// Zweitblick über Phase 4.
		//
		// **Gemeldet wird nur eine negative Größe**, und der Unterschied ist
		// die Absicht: `magic-m15-saga.mse-style` schreibt am Textfeld
		// `width: {if is_saga() then 0 else 153}` — die Vorlage **verbirgt** es
		// so, und MSE malt in einen Kasten ohne Fläche nichts und sagt nichts.
		// Eine negative Größe dagegen kann keine Vorlage meinen; sie entsteht,
		// wenn `right` links von `left` liegt
		// (#270).
		if ebene.Rahmen.Breite < 0 || ebene.Rahmen.Hoehe < 0 {
			return fmt.Sprintf("ein Textrahmen von %g×%g",
				ebene.Rahmen.Breite, ebene.Rahmen.Hoehe)
		}

	default:
		return fmt.Sprintf("unbekannte Ebenenart %q", ebene.Art)
	}

	return ""
}

// flaechenmangel prüft, was eine Fläche zum Zeichnen braucht.
//
// **Ein Verlauf mit einer Stufe ist keiner.** Er sähe aus wie eine einfarbige
// Fläche und wäre einer, den jemand halb geschrieben hat — das wortlos zu
// zeichnen verbirgt genau die Zeile, die er ansehen müsste.
func flaechenmangel(ebene Ebene) string {
	if ebene.Verlauf == nil {
		return ""
	}

	if len(ebene.Verlauf.Stufen) < 2 {
		return fmt.Sprintf("ein Verlauf mit %d Stufen", len(ebene.Verlauf.Stufen))
	}

	for i := 1; i < len(ebene.Verlauf.Stufen); i++ {
		if ebene.Verlauf.Stufen[i].Stelle < ebene.Verlauf.Stufen[i-1].Stelle {
			return "die Stufen des Verlaufs stehen nicht aufsteigend"
		}
	}

	return ""
}

// Zeichenbar gibt den Plan ohne die beanstandeten Ebenen — und die Nummern,
// die diese Ebenen im **ursprünglichen** Plan hatten.
//
// **Die zweite Rückgabe ist der ganze Grund für diese Fassung.** Der erste
// Anlauf gab nur den gekürzten Plan, und ein Laufzeitbefund bekam dann dessen
// Zählung: fiel Ebene 1 bei der Prüfung weg, meldete ein fehlendes Bild in
// Ebene 2 sich als Ebene 1. Wer danach in einer Vorlage mit 60 Ebenen sucht,
// sucht an der falschen. Gefunden vom Nachlauf des Zweitblicks über Phase 3.
//
// Der ursprüngliche Plan bleibt unberührt: er ist die Auskunft darüber, was die
// Vorlage wollte, und die soll ein Befund nicht überschreiben.
func Zeichenbar(plan *Plan, befunde []Befund) (gekuerzt *Plan, nummern []int) {
	nummern = make([]int, 0, len(plan.Ebenen))

	if len(befunde) == 0 {
		for nummer := range plan.Ebenen {
			nummern = append(nummern, nummer+1)
		}

		return plan, nummern
	}

	weg := make(map[int]bool, len(befunde))
	for _, befund := range befunde {
		weg[befund.Ebene] = true
	}

	gekuerzt = &Plan{
		Breite: plan.Breite,
		Hoehe:  plan.Hoehe,
		Ebenen: make([]Ebene, 0, len(plan.Ebenen)),
	}

	for nummer := range plan.Ebenen {
		if !weg[nummer+1] {
			gekuerzt.Ebenen = append(gekuerzt.Ebenen, plan.Ebenen[nummer])
			nummern = append(nummern, nummer+1)
		}
	}

	return gekuerzt, nummern
}
