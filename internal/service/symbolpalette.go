package service

// Die **Palette** eines Feldes — welche Symbole es anzubieten hat
// (#564).
//
// # Wofür
//
// Im Eingabefeld für die Manakosten steht `2RR`. Auf der Karte werden daraus
// Symbole; im Feld bleiben es Buchstaben, und wer nicht weiß, dass `T` das
// Tap-Symbol ist, findet es nicht. MSE bietet dafür ein Menü an, und es steht
// **im Paket**:
//
//	insert symbol menu:
//		item: T
//		item:
//			type: line
//
// `magic-mana-large.mse-symbol-font/symbol-font`, ab Zeile 841. Über
// `full-magic-pack` haben **76 von 76** Symbolschriften eines, zusammen 5.219
// Codes, 448 Trenner, 62 eigene und 433 Untermenüs.
//
// # Was hier entsteht und was nicht
//
// Beantwortet werden die drei Fragen, die eine Umgebung brauchen: **welche**
// Symbolschrift ein Feld benutzt — der Name darf ein Skript sein —, **welche
// Fassung** eines Symbols gerade gilt (`enabled:`), und wie ein Symbol
// aussieht. Gerastert wird nicht: jeder Eintrag trägt einen Plan (§2).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Palettengroesse ist die Kantenlänge, mit der ein Palettensymbol geplant wird.
//
//	item->SetBitmap(symbol->getBitmap(font, wxSize(16,16)));
//
// `InsertSymbolMenu::makeMenuItem`, `src/data/symbol_font.cpp`, Zeile 515,
// Fassung f9d9356 vom 2021-07-06. MSE nimmt 16 Bildpunkte für ein Menü; hier
// steht dieselbe Zahl in **Vorlagenpunkten**, und wer eine schärfere Kachel
// will, rastert mit Vergrößerung.
const Palettengroesse = 16

// Paletteneintrag ist ein Eintrag der Palette.
type Paletteneintrag struct {
	Art msepaket.Menueart

	// Code ist, was eingefügt wird — leer bei Trenner und Untermenü.
	Code string

	// Beschriftung ist die Aufschrift des Eintrags.
	//
	// **Der großgeschriebene Name, solange kein Sprachpaket gelesen wird.**
	//
	//	if (m.label.empty())
	//	  m.label.default_ = tr(*symbol_font_for_reading(), "menu_item", m.name, capitalize);
	//
	// `after_reading(InsertSymbolMenu&)`, `src/data/symbol_font.cpp`, Zeilen
	// 545–549, Fassung f9d9356 vom 2021-07-06. Die `.mse-locale`-Pakete sind
	// #561.
	Beschriftung string

	// Rueckfrage ist der Satz über dem Eingabefeld eines [msepaket.MenueEigen].
	Rueckfrage string

	// Plan zeichnet dieses eine Symbol — `nil` bei Trenner, Untermenü und
	// jedem Code, den die Schrift nicht kennt.
	Plan *render.Plan

	Kinder []Paletteneintrag
}

// Symbolpalette gibt die Einträge, die ein Feld zum Einfügen anbietet.
//
// **`nil` ohne Symbolschrift**, und ohne Befund: über `magic-m15.mse-style`
// führen die allerwenigsten Feldstile einen `symbol font:`-Block, und eine
// Meldung je Feld ohne einen wäre eine Liste, die niemand liest (§8).
func (p *Planbauer) Symbolpalette(
	ctx context.Context, karte *domain.CardRow, feld string, vorlage Vorlage,
) ([]Paletteneintrag, []domain.Befund, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	stil := feldstil(vorlage.Stil, feld)
	if stil == nil {
		return nil, nil, nil
	}

	umgebung, _, befunde := umgebungAus(ctx, karte, vorlage, p.schriften, p.bilder)

	// **Welche Fassung eines Symbols gilt, steht erst nach dem Initskript
	// fest** — derselbe Grund wie in [Planbauer.Stilplan]: die Bedingungen
	// fragen nach Funktionen, die es anlegt.
	schriften, fassungsbefunde := aktiveSymbolschriften(ctx, vorlage.Symbolschriften, umgebung)
	befunde = append(befunde, fassungsbefunde...)

	name := symbolschriftname(ctx, stil, umgebung)
	if name == "" {
		return nil, nil, nil
	}

	schrift, da := schriften[name]
	if !da {
		return nil, append(befunde, domain.Befund{
			Ort:     feld,
			Meldung: "die Symbolschrift " + name + " liegt nicht vor",
		}), nil
	}

	if schrift.Einfuegemenue == nil {
		return nil, befunde, nil
	}

	return paletteneintraege(ctx, schrift.Einfuegemenue, stil, schrift, name, umgebung),
		befunde, nil
}

// Kartenpalette sucht den Stil der Karte und gibt die Palette eines Feldes.
//
// **Dieselbe Auflösung wie [Planbauer.Kartenplan]**, und aus demselben Grund
// getrennt: die Vorschau und die Palette sprechen über dieselbe Karte, und
// zwei Wege zum Stil liefen beim ersten Sonderfall auseinander.
//
// **Ohne Behelf.** Fällt der Stil aus, gibt es keine Symbolschrift und damit
// nichts anzubieten; der Behelfsplan zeichnet Feldwerte als Zeilen und kennt
// keine Symbole.
func (p *Planbauer) Kartenpalette(ctx context.Context, karte *domain.CardRow, feld string) (
	[]Paletteneintrag, bool, []domain.Befund, error,
) {
	if p.stile == nil {
		return nil, false, nil, nil
	}

	vorlage, befunde, da := p.stile.StilFuerKarte(ctx, karte)
	if !da {
		return nil, false, befunde, nil
	}

	eintraege, weitere, err := p.Symbolpalette(ctx, karte, feld, vorlage)

	return eintraege, FuegtRohEin(vorlage.Stil, feld),
		append(befunde, weitere...), err
}

// FuegtRohEin sagt, ob ein eingefügter Code ohne `<sym>…</sym>` auskommt.
//
//	String code = style().symbol_font.font->insertSymbolCode(id);
//	if (!style().always_symbol) {
//	  code = _("<sym>") + code + _("</sym>");
//	}
//
// `TextValueEditor::onCommand`, `src/gui/value/text.cpp`, Zeilen 589–593,
// Fassung f9d9356 vom 2021-07-06. Bei den Manakosten ist der ganze Wert
// Symbole; im Regeltext steht `<sym>T</sym>` mitten im Fließtext.
//
// **Ausgeführt und nicht am Eintrag**, denn es ist eine Aussage über das
// **Feld** und nicht über das Symbol: an einem Eintrag stünde dieselbe Antwort
// sechzigmal.
func FuegtRohEin(stil msepaket.Stil, feld string) bool {
	gefunden := feldstil(stil, feld)

	return gefunden != nil && gefunden.ImmerSymbole
}

// feldstil sucht den Feldstil eines Feldes — Kartenfelder wie Zusatzfelder.
func feldstil(stil msepaket.Stil, feld string) *msepaket.Feldstil {
	for _, kandidat := range alleFeldstile(stil.Felder, stil.Zusatzstile) {
		if kandidat.stil.Feld == feld {
			return kandidat.stil
		}
	}

	return nil
}
