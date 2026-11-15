package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Ein Feld mit `symbol font:` zeichnet Bilder, keinen Text.
//
// # Der Befund
//
// `magic-m15.mse-style` legt die Manakosten mit **beiden** Blöcken an:
//
//	casting cost:
//	    font:        { name: MPlantin,        size: 15 }
//	    symbol font: { name: magic-mana-large, size: 15 }
//
// Der zweite gewinnt: `2R` ist keine Zeichenfolge, sondern zwei Bilder. Bis
// #240 zeichnete
// dieses Programm den Text `2R` in MPlantin — lesbar, und keine Manakosten
// (#173).
//
// # Gemessen wird nicht hier
//
// Wie groß ein Symbol wird, hängt an seinem **Bild**: MSE rechnet die
// tatsächliche Bildgröße über `font_size / image font size` um und legt
// `horizontal space` dazu. Bilder öffnet der Planbauer nicht (§2), und deshalb
// steht hier eine **einzige** Ebene mit den Angaben statt *n* Ebenen mit
// ausgerechneten Quadraten
// ([0022](../../docs/entscheidungen/0022-der-rasterer-misst-die-symbole.md)).

// symbolebenen baut die Ebene eines Felds mit Symbolschrift.
func symbolebenen(
	ctx context.Context, feld *msepaket.Feldstil, name, wert string,
	rahmen render.Rechteck, vorlage Vorlage, umgebung *msescript.Umgebung,
) ([]render.Ebene, []domain.Befund) {
	schrift, da := vorlage.Symbolschriften[name]
	if !da {
		return nil, []domain.Befund{{
			Ort: feld.Feld,
			Meldung: "die Symbolschrift " + name +
				" liegt nicht vor; das Feld wird nicht gezeichnet",
		}}
	}

	// **Ohne Prüfung auf unbekannte Zeichen**, und das ist kein Weglassen:
	// [ganzAusSymbolen] hat sie schon gestellt und diesen Weg nur freigegeben,
	// weil jedes Zeichen einen Code hat. Eine zweite Prüfung wäre ein Zweig,
	// den kein Aufruf betreten kann (§3) — und zwei Stellen, die dieselbe Frage
	// verschieden beantworten könnten.
	codes, _ := zerlegeInSymbole(wert, schrift)

	bilder := symbolbilder(ctx, codes, name, umgebung)

	befunde := faerbeSymboltexte(ctx, bilder, codes, feld.Feld, umgebung)

	return []render.Ebene{{
		Art:    render.ArtSymbolzug,
		Rahmen: rahmen,
		Abschnitte: []render.Abschnitt{{
			Text:        wert,
			Symbole:     bilder,
			Symbolmasse: symbolmasse(ctx, feld, rahmen, schrift, umgebung),
		}},
		// **Nur die Ausrichtung.** Ein Symbolzug hat keine Schrift und keine
		// Schriftgröße; was der Rasterer von diesem Satz braucht, ist die Frage,
		// wo der Zug in seinem Feld sitzt.
		Satz:    render.Satz{Ausrichtung: ausrichtungAus(feld.Ausrichtung)},
		Deckung: 1,
	}}, befunde
}

// symbolbilder macht aus den getroffenen Symbolen die des Plans.
//
// **Aus den Treffern und nicht aus den Codes.** Ein Code darf ein Ausdruck
// sein; ein zweiter Blick in die Schrift fände dann nichts, und ein Code, den
// die Schrift mehrfach führt, fände womöglich die andere Fassung (#293).
//
// **Und das Bild ist ein Baum, kein Name.** `SymbolInFont` hält es bei MSE2
// als `ScriptableImage` (`src/data/symbol_font.cpp`, Zeile 93, Fassung f9d9356
// vom 2021-07-06); über `Full-Magic-Pack` sind 2.435 von 8.110 `image:`-Zeilen
// der Symbolschriften Skripte, gezählt an der Zeile (§10). Bis
// #422 wurde
// die Angabe an den Paketpfad geklebt, und `recolor_image(…)` wurde zu einem
// Dateinamen, den es nicht gibt.
//
// **Befunde verfallen hier**, aus demselben Grund wie bei [zahlVorSatz]: diese
// Zerlegung läuft zweimal — einmal für den Plan, einmal für die Messung des
// Inhalts —, und ein Bild, das sich nicht ausrechnen lässt, meldet der
// Rasterer ohnehin als fehlendes Symbolbild.
func symbolbilder(
	ctx context.Context, treffer []Symboltreffer, name string,
	umgebung *msescript.Umgebung,
) []render.Symbol {
	paket := name + "." + string(msepaket.ArtSymbolschrift)
	symbole := make([]render.Symbol, 0, len(treffer))

	for i := range treffer {
		einer := &treffer[i]

		baum, _ := bildangabe(ctx, name, "symbol "+einer.Symbol.Code,
			einer.Symbol.Bild, umgebung)

		symbole = append(symbole, render.Symbol{
			Bild:        alsZeichenbaum(baum, paket, nil),
			Bildgroesse: einer.Symbol.Bildgroesse,
			Text:        einer.Text,
			Textsatz:    textsatzAus(einer.Symbol),
		})
	}

	return symbole
}

// symbolmasse sammelt die Maße eines Symbolzugs aus seinen zwei Quellen.
//
// Größe und Ausrichtung stehen im `symbol font:`-Block des **Feldstils**,
// Bildgröße und Abstände im `.mse-symbol-font`-**Paket**.
func symbolmasse(
	ctx context.Context, feld *msepaket.Feldstil, rahmen render.Rechteck,
	schrift msepaket.Symbolschrift, umgebung *msescript.Umgebung,
) render.Symbolmasse {
	return render.Symbolmasse{
		Groesse:           symbolgroesse(ctx, feld, rahmen, umgebung),
		Bildgroesse:       schrift.Bildgroesse,
		AbstandWaagerecht: schrift.AbstandWaagerecht,
		AbstandSenkrecht:  schrift.AbstandSenkrecht,
		Ausrichtung:       symbolausrichtung(feld.Symbolschrift.Ausrichtung),
	}
}

// symbolausrichtung liest die `alignment:` eines Symbolschriftblocks.
//
// **Ohne Angabe mittig, nicht links oben.** MSE gibt `SymbolFontRef` die
// Vorgabe `ALIGN_MIDDLE_CENTER` mit (`src/data/symbol_font.cpp`, Zeile 556,
// Fassung f9d9356 vom 2021-07-06); steht dagegen etwas da, wird es wie jede
// andere Ausrichtung gelesen — dann gilt für die fehlende Richtung wieder
// links beziehungsweise oben.
func symbolausrichtung(angabe string) render.Ausrichtung {
	if strings.TrimSpace(angabe) == "" {
		return render.MitteMitte
	}

	return ausrichtungAus(angabe)
}

// zerlegeInSymbole zerteilt einen Text in die Codes einer Symbolschrift.
//
// **In der Reihenfolge, in der die Schrift sie führt** — nicht nach Länge.
// `SymbolFont::split` nimmt den ersten Eintrag, dessen Code am Anfang des
// Restes steht (Zeilen 191–230, Fassung f9d9356 vom 2021-07-06); die
// Dateireihenfolge ist die Entscheidung des Schriftautors.
//
// Der zweite Rückgabewert nennt das erste Zeichen, für das kein Code passte.
// **Nicht übersprungen**: eine Manakostenangabe, der ein Symbol fehlt, sieht
// aus wie eine billigere Karte (§8).
func zerlegeInSymbole(
	text string, schrift msepaket.Symbolschrift,
) (treffer []Symboltreffer, unbekannt string) {
	zeichen := []rune(text)

	for stelle := 0; stelle < len(zeichen); {
		rest := string(zeichen[stelle:])

		symbol, gruppen, laenge := ersterTreffer(rest, schrift)
		if laenge == 0 {
			if unbekannt == "" {
				unbekannt = string(zeichen[stelle])
			}

			stelle++

			continue
		}

		treffer = append(treffer, Symboltreffer{Symbol: symbol, Text: gezeichneterText(symbol, gruppen)})
		stelle += len([]rune(rest[:laenge]))
	}

	return treffer, unbekannt
}

// ersterTreffer sucht das erste Symbol, dessen Code hier steht.
//
// Die Länge ist in **Bytes** des Restes, denn ein Ausdruck misst so; der
// Aufrufer rechnet sie in Zeichen um.
//
// **Ein Muster, das sich nicht lesen lässt, trifft nicht.** Es wird
// übersprungen wie ein Code, der nicht passt — die Meldung dazu steht in
// `symbolmuster.go`, und ein Befund je Karte und Muster stünde in einer Liste,
// die niemand zu Ende liest.
func ersterTreffer(
	rest string, schrift msepaket.Symbolschrift,
) (getroffen msepaket.Symbol, gruppen []string, laenge int) {
	// **Ohne Prüfung auf einen leeren Code**, den MSE ausdrücklich stellt
	// (`!sym->code.empty()`): [Paket.Symbolmodell] lässt ein Symbol ohne `code`
	// gar nicht erst ins Modell, sondern meldet es. Hier wäre es ein Zweig, den
	// kein Aufruf betreten kann (§3).
	for i := range schrift.Symbole {
		symbol := schrift.Symbole[i]

		if !symbol.Ausdruck {
			if strings.HasPrefix(rest, symbol.Code) {
				// **Ohne Muster gibt es nur den ganzen Treffer** — MSE
				// schreibt dann den Code selbst (`sym->draw_text >= 0 ?
				// sym->code : ""`, Zeile 221).
				return symbol, []string{symbol.Code}, len(symbol.Code)
			}

			continue
		}

		gelesen, err := muster(symbol.Code)
		if err != nil {
			continue
		}

		if gedeckt, gruppen, passt := gelesen.treffer(rest); passt {
			return symbol, gruppen, gedeckt
		}
	}

	return msepaket.Symbol{}, nil, 0
}

// Symboltreffer ist ein Symbol samt dem Text, der auf sein Bild kommt.
type Symboltreffer struct {
	Symbol msepaket.Symbol
	Text   string
}

// gezeichneterText liest die Gruppe, die `draw text:` nennt.
//
//	if (sym->draw_text >= 0 && sym->draw_text < (int)results.size()) {
//	  out.push_back(DrawableSymbol(results.str(), results.str(sym->draw_text), *sym));
//	}
//
// `SymbolFont::split`, `src/data/symbol_font.cpp`, Zeilen 205–221, Fassung
// f9d9356 vom 2021-07-06.
//
// **Eine Nummer außerhalb der Gruppen zeichnet nichts** — MSE prüft sie gegen
// `results.size()` und lässt den Text sonst weg. Eine Vorlage mit
// `draw text: 3` und zwei Gruppen ist falsch geschrieben; ein erfundener Text
// verdeckte das.
func gezeichneterText(symbol msepaket.Symbol, gruppen []string) string {
	if symbol.Textgruppe < 0 || symbol.Textgruppe >= len(gruppen) {
		return ""
	}

	return gruppen[symbol.Textgruppe]
}

// symbolgroesse gibt die Kantenlänge eines Symbols.
//
// **Die Angabe des Feldstils, sonst die Höhe des Kastens.** `size:` fehlt in
// einigen Stilen; dann ist die Kastenhöhe die beste vorhandene Auskunft — eine
// erfundene Zahl wäre schlechter, und null zeichnete nichts.
func symbolgroesse(
	ctx context.Context, feld *msepaket.Feldstil, rahmen render.Rechteck,
	umgebung *msescript.Umgebung,
) float64 {
	if zahl := zahlVorSatz(ctx, feld.Feld, "symbolgröße",
		feld.Symbolschrift.Groesse, Symbolgroessenvorgabe, umgebung); zahl > 0 {
		return zahl
	}

	return rahmen.Hoehe
}

// symbolschriftname gibt den Namen der Symbolschrift eines Felds.
//
// **Er darf ein Skript sein.** `text:` in `magic-m15.mse-style` schreibt
// `symbol font: { styling.text_box_mana_symbols }` — welche Symbolschrift gilt,
// entscheidet eine Stilwahl. Wer den rohen Text nähme, suchte ein Paket namens
// `{ styling.text_box_mana_symbols }.mse-symbol-font`.
func symbolschriftname(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) string {
	roh := strings.TrimSpace(feld.Symbolschrift.Name)

	rumpf, geklammert := skriptrumpf(roh)
	if !geklammert {
		return roh
	}

	wert, err := msescript.AuswertenQuelle(ctx, rumpf, umgebung)
	if err != nil {
		// **Kein Befund hier.** Ob das eine Lücke ist, entscheidet sich weiter
		// unten: ohne Namen wird das Feld als Text gezeichnet, und das ist für
		// jedes Feld mit beiden Blöcken die richtige Antwort.
		return ""
	}

	return ohnePaketendung(strings.TrimSpace(wert.AlsText()))
}

// ohnePaketendung schneidet `.mse-symbol-font` von einem Namen ab.
//
// **Weil beide Formen vorkommen.** `magic-m15.mse-style` schreibt am Feld
// `casting cost` den bloßen Namen und am Feld `text` ein Skript, dessen
// Ergebnis `magic-mana-small.mse-symbol-font` lautet — mit Endung. Die
// Symbolschriften liegen unter dem bloßen Namen, und ohne diesen Schnitt fände
// die zweite Form nie eine
// (#312).
func ohnePaketendung(name string) string {
	return strings.TrimSuffix(name, "."+string(msepaket.ArtSymbolschrift))
}

// ganzAusSymbolen sagt, ob ein Wert vollständig in Symbolen darstellbar ist.
//
// **Die Bedingung für den Symbolzweig, und sie hat zwei Hälften.**
//
// Die erste ist eine **Ansage des Stils**: `always symbol: true`. MSE zerlegt
// nur dann ein ganzes Feld in Symbole (`src/render/text/element.cpp`, Zeile
// 244, Fassung f9d9356 vom 2021-07-06); ohne sie ist der Wert Text, in dem nur
// `<sym>`-Stellen Symbole sind. Im Pack steht sie 133-mal, an den Manakosten
// und nie am Regeltext.
//
// **Die zweite Hälfte allein genügt nicht** — „lässt sich der ganze Wert
// zerlegen?". `magic-mana-large` endet auf den Auffangcode
// `[0-9](?!/[WUBRGCSTQ2])|.`, der **jedes** Zeichen deckt; ohne die Ansage
// würde damit auch ein Regeltext zu einer Reihe von Bildern
// (#293).
//
// **Geht der Wert nicht ganz auf**, wird er kein Symbolzug, sondern ein
// Textzug — und der mischt Symbole und Text in Läufen, wie MSE bei `always
// symbol:` ([symbollaeufe],
// #508).
func ganzAusSymbolen(feld *msepaket.Feldstil, wert, name string, vorlage Vorlage) bool {
	if strings.TrimSpace(wert) == "" || !feld.ImmerSymbole {
		return false
	}

	schrift, da := vorlage.Symbolschriften[name]
	if !da {
		// **Ohne die Schrift zur Hand entscheidet, ob es eine Textschrift gibt.**
		//
		// Gibt es keine, ist der Symbolzweig der einzige Weg, und ein Befund
		// dort sagt, was fehlt. Gibt es eine — wie am Feld `text`, dessen
		// Symbolschrift aus einer Stilwahl kommt —, ist Text die bessere
		// Antwort als gar nichts: der Regeltext steht dann auf der Karte, und
		// nur die eingestreuten Symbole fehlen.
		return feld.Schrift.Name == ""
	}

	_, unbekannt := zerlegeInSymbole(wert, schrift)

	return unbekannt == ""
}

// symboltextausrichtung liest `text alignment:` eines Symbols.
//
// **Leer bleibt leer.** MSEs Vorgabe für den Symboltext ist `middle center`
// (`text_alignment(ALIGN_MIDDLE_CENTER)`, `src/data/symbol_font.cpp`, Zeile
// 106, Fassung f9d9356 vom 2021-07-06), die eines Feldes dagegen links oben.
// Wer hier [ausrichtungAus] nähme, schöbe jede Ziffer in die linke obere Ecke
// ihres Kreises (#336).
func symboltextausrichtung(angabe string) render.Ausrichtung {
	if strings.TrimSpace(angabe) == "" {
		return ""
	}

	return ausrichtungAus(angabe)
}

// SymbolbilderFuerTest gibt die Zerlegung eines Symbolzugs frei.
//
// **Weil die Kette dazwischen lang ist.** Von `draw text:` bis zum Plan liegen
// Zerlegung, Gruppenlesung und Übertragung; ein Test über den ganzen Planbauer
// prüfte alles drei zugleich und sagte bei einem Fehler nicht, welches.
func SymbolbilderFuerTest(
	ctx context.Context, text string, schrift msepaket.Symbolschrift,
	umgebung *msescript.Umgebung,
) []render.Symbol {
	treffer, _ := zerlegeInSymbole(text, schrift)

	return symbolbilder(ctx, treffer, schrift.Name, umgebung)
}
