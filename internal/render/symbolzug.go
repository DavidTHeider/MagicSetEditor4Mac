package render

import (
	"image"
	"slices"
	"strings"
)

// **Ein ganzes Feld aus Symbolen — die Manakosten.**
//
// # Warum das eine eigene Ebenenart ist und keine Textebene
//
// Bei MSE ist es keine: ein Feld mit `always symbol: true` geht durch denselben
// Textsatz und wird dort in `SymbolTextElement`e zerlegt
// (`src/render/text/element.cpp`, Zeilen 244–262, Fassung f9d9356 vom
// 2021-07-06). Hier steht es daneben, weil der Textweg eine **Schrift** braucht
// und ein Symbolzug keine hat: ein Feld, dessen Textschrift auf diesem Rechner
// fehlt, verlöre sonst auch seine Manakosten (§8).
//
// # Und warum es keine vorberechneten Rechtecke mehr sind
//
// Bis zum 2026-08-17 zerlegte `internal/service` die Manakosten in *n*
// Bildebenen mit ausgerechneten Quadraten. Wie groß ein Symbol wird, hängt aber
// an seinem Bild — und Bilder öffnet der Planbauer nicht
// (`docs/entscheidungen/0022-der-rasterer-misst-die-symbole.md`).

// symbolzug setzt einen Symbolzug in seinem Rahmen und zeichnet ihn.
func (r *Rasterer) symbolzug(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, vergroesserung float64,
	guete Guete,
) []string {
	// **Ohne Prüfung auf einen leeren Zug**, und das ist kein Weglassen:
	// [ebenenmangel] hat sie schon gestellt, und eine Ebene ohne Symbole
	// erreicht diese Stelle nicht. Eine zweite Prüfung wäre ein Zweig, den kein
	// Aufruf betreten kann (§3) — und zwei Stellen, die dieselbe Frage
	// verschieden beantworten könnten.
	symbole, masse := zugAus(ebene.Abschnitte)

	breite, hoehe := Symbolzugmass(symbole, masse, r.quelle)

	// **Der Zug wird im Feld ausgerichtet, das Bild in seinem Kasten.** Zwei
	// Ausrichtungen, zwei Quellen: `alignment:` des Feldstils und `alignment:`
	// des Symbolschriftblocks. `magic-m15` schreibt für die Manakosten beide
	// hin, und sie sind zufällig gleich — verwechseln ließen sie sich trotzdem.
	links := ebene.Rahmen.Links +
		waagerechterVersatz(ebene.Satz.Ausrichtung, ebene.Rahmen.Breite, breite)
	oben := ebene.Rahmen.Oben +
		senkrechterVersatz(ebene.Satz.Ausrichtung, ebene.Rahmen.Hoehe, hoehe)

	bilder, schriften := r.zeichneSymbole(ziel, ebene, maske, skaliert(ebene.Rahmen, vergroesserung),
		symbole, masse, links, oben, hoehe, vergroesserung,
		gestauchtVon(ebene.Rahmen.Links, 1), guete)

	return symbolmaengel(bilder, schriften)
}

// symbolmaengel gibt die Mängel eines Symbolzugs — **fehlende Bilder je Art in
// einer Meldung, jede fehlende Schrift in ihrer eigenen**, damit jede ihre
// eigene Art bekommt (QA-03, #1656). Ein Bild, das in keinem Paket liegt
// (Vorlage), nähme sonst eines mit, dessen Verknüpfung hier fehlt
// (Portierung). Dieselbe Schrift steht einmal da, auch wenn zehn Symbole sie
// vermissen.
func symbolmaengel(bilder, schriften []string) []string {
	var (
		maengel []string
		arten   []Befundart
	)

	jeArt := map[Befundart][]string{}

	for _, bild := range bilder {
		art := ArtAus(bild)
		if _, gesehen := jeArt[art]; !gesehen {
			arten = append(arten, art)
		}

		jeArt[art] = append(jeArt[art], bild)
	}

	for _, art := range arten {
		maengel = append(maengel, "die Symbolbilder fehlen: "+strings.Join(jeArt[art], " "))
	}

	for _, meldung := range schriften {
		if !slices.Contains(maengel, meldung) {
			maengel = append(maengel, meldung)
		}
	}

	return maengel
}

// Symbolzugmass misst einen Symbolzug, ohne ihn zu zeichnen.
//
// **Für die erste Stufe des Satzes.** `card_style.<feld>.content_width` fragt,
// wie breit der gesetzte Inhalt eines Feldes ist, und der Planbauer rechnet
// damit, bevor irgendetwas gezeichnet ist. Gemessen wird hier und nicht dort:
// wie groß ein Symbol wird, hängt an seinem Bild (Vermerk 0022), und zwei
// Rechnungen über dieselbe Sache laufen auseinander
// (#291).
//
// Die Höhe ist die des größten Kastens — dieselbe Regel wie für eine Zeile.
func Symbolzugmass(
	symbole []Symbol, masse Symbolmasse, bilder Bildquelle,
) (breite, hoehe float64) {
	for i := range symbole {
		symbol := &symbole[i]

		kasten := masse.kasten(symbol, bilder)
		breite += kasten.Breite
		hoehe = max(hoehe, kasten.Hoehe)
	}

	return breite, hoehe
}

// zugAus fasst die Symbole aller Abschnitte einer Ebene zusammen.
//
// **Die Maße des ersten symbolischen Abschnitts gelten für alle.** Ein
// Symbolzug ist ein Feld mit **einer** Symbolschrift; mehrere Abschnitte
// entstehen dort nur durch Auszeichnungen, und die ändern keine Symbolgröße.
func zugAus(abschnitte []Abschnitt) ([]Symbol, Symbolmasse) {
	var (
		symbole []Symbol
		masse   Symbolmasse
	)

	for _, abschnitt := range abschnitte {
		if !abschnitt.istSymbolisch() {
			continue
		}

		if len(symbole) == 0 {
			masse = abschnitt.Symbolmasse
		}

		symbole = append(symbole, abschnitt.Symbole...)
	}

	return symbole, masse
}

// zeichneSymbole malt Symbole nebeneinander und nennt die Bilder, die fehlen.
//
// **Eine Stelle für beide Wege** — das Feld aus Symbolen und das `<sym>` mitten
// im Regeltext. Zwei Rechnungen über dieselbe Sache laufen auseinander, sobald
// eine von ihnen etwas lernt.
//
// `zeilenhoehe` ist die Höhe des Kastens, in dem ein Symbol **senkrecht**
// ausgerichtet wird: im Text die Zeile, im Symbolzug der Zug. Bei MSE ist es
// beide Male dasselbe — `split_left` erhält die Höhe des Zeilenrechtecks
// (`symbol_font.cpp`, Zeilen 279–294).
func (r *Rasterer) zeichneSymbole(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, rahmen image.Rectangle,
	symbole []Symbol, masse Symbolmasse,
	links, oben, zeilenhoehe, vergroesserung float64, gestaucht stauchflaeche,
	guete Guete,
) (fehlend, schriften []string) {
	lauf := links

	for i := range symbole {
		symbol := &symbole[i]

		kasten := masse.kasten(symbol, r.quelle)

		quelle, befund := r.ausBaum(symbol.Bild,
			image.Point{X: int(kasten.Bildbreite), Y: int(kasten.Bildhoehe)})
		if befund != "" {
			// **Der Grund gehört dazu.** Ein Symbolbild kann fehlen, weil die
			// Datei nicht daliegt — oder weil die Verknüpfung darüber nicht
			// ausführbar ist. Zwei Abhilfen, und der bloße Pfad nennt keine
			// (#422).
			fehlend = append(fehlend, befund)
			lauf += kasten.Breite

			continue
		}

		bildLinks, bildOben := kasten.bildstelle(masse.Ausrichtung, lauf, oben, zeilenhoehe)

		// **Auch ein Symbol wird gestaucht.** MSE staucht die ganze
		// Zeichenfläche, und ein Bild darauf geht mit (#319).
		gestaucheteKante := gestaucht.stelle(bildLinks)

		feld := image.Rect(
			int(gestaucheteKante*vergroesserung), int(bildOben*vergroesserung),
			int((gestaucheteKante+gestaucht.breite(kasten.Bildbreite))*vergroesserung),
			int((bildOben+kasten.Bildhoehe)*vergroesserung),
		)

		for y := feld.Min.Y; y < feld.Max.Y; y++ {
			for x := feld.Min.X; x < feld.Max.X; x++ {
				if !(image.Point{X: x, Y: y}).In(ziel.Bounds()) {
					continue
				}

				r.punkt(ziel, x, y, abtasten(quelle, feld, x, y, guete),
					ebene.Deckung*maske.anteil(rahmen, x, y), nil)
			}
		}

		// **Und der Text darauf** — die Ziffer im Manakreis (#336). Was ihm
		// fehlt, ist eine Schrift und kein Bild.
		schriften = append(schriften, r.symboltext(ziel, ebene, maske, rahmen,
			symbol, kasten, masse.Groesse,
			gestaucheteKante, bildOben, vergroesserung, gestaucht)...)

		lauf += kasten.Breite
	}

	return fehlend, schriften
}
