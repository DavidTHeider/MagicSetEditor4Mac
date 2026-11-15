package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// mitSymbolen löst die `<sym>`-Abschnitte eines Textzugs in Bilder auf.
//
// # Der Befund
//
// Bis zum 2026-08-16 war ein Feld **entweder** ganz Text **oder** ganz Symbole
// ([ganzAusSymbolen]). Das trifft die Manakosten und lässt den Regeltext Text
// sein — und ließ damit jedes `<sym>T</sym>` mitten im Regeltext als
// **Buchstaben** stehen.
//
// MSE macht es anders: es zerlegt den Zug in Elemente, und ein Element
// innerhalb von `<sym>` wird ein `SymbolTextElement`
// (`src/render/text/element.cpp`, Zeilen 82–83 und 230–232, Fassung f9d9356 vom
// 2021-07-06).
//
// # Was hier entschieden wird und was nicht
//
// Hier wird **zugeordnet**: aus dem Code wird ein Bildpfad, aus der Angabe
// `symbol font: size:` eine Kantenlänge. Wo das Symbol landet, weiß erst der
// Umbruch, und der steht in `internal/render` — deshalb reist der Pfad mit dem
// Abschnitt und nicht als eigene Ebene
// (#312).
func mitSymbolen(
	ctx context.Context, feld *msepaket.Feldstil, rahmen render.Rechteck,
	roh []rohabschnitt, umgebung *msescript.Umgebung, vorlage Vorlage,
) ([]render.Abschnitt, []domain.Befund) {
	name := symbolschriftname(ctx, feld, umgebung)

	abschnitte := make([]render.Abschnitt, 0, len(roh))

	var befunde []domain.Befund

	for _, teil := range roh {
		abschnitt := render.Abschnitt{
			Text: teil.Text, Kursiv: teil.Kursiv, Fett: teil.Fett,
			Blockende:   teil.Blockende,
			Weich:       teil.Weich,
			WeicheZeile: teil.WeicheZeile,
			Margen:      teil.Margen,
			ImRand:      teil.ImRand,
		}

		if teil.Symbolisch {
			symbole, masse, befund := symbolstueck(
				ctx, feld, name, teil.Text, rahmen, vorlage, umgebung)
			if befund != nil {
				befunde = append(befunde, *befund)
			}

			abschnitt.Symbole = symbole
			abschnitt.Symbolmasse = masse
		}

		abschnitte = append(abschnitte, abschnitt)
	}

	return abschnitte, befunde
}

// symbolstueck macht aus einem Codezug die Symbole eines Abschnitts samt Maßen.
//
// **Ein Befund statt eines halben Zugs.** Fehlt die Schrift oder ein Code, wird
// der Abschnitt zu Text — sichtbar falsch statt unsichtbar fehlend. Eine Karte,
// der ein Manasymbol fehlt, sieht aus wie eine billigere (§8).
//
// **Die Maße reisen mit den Symbolen**, nicht daneben: gemessen wird erst beim
// Rastern, und ein Abschnitt, dessen Bilder aus einer Schrift stammen und
// dessen Maße aus einer anderen, wäre nicht mehr zusammenzubringen (0022).
func symbolstueck(
	ctx context.Context, feld *msepaket.Feldstil, name, text string,
	rahmen render.Rechteck, vorlage Vorlage, umgebung *msescript.Umgebung,
) ([]render.Symbol, render.Symbolmasse, *domain.Befund) {
	if name == "" {
		return nil, render.Symbolmasse{}, &domain.Befund{
			Ort: feld.Feld,
			Meldung: "der Text enthält Symbole, aber für dieses Feld ist keine " +
				"Symbolschrift eingestellt; sie bleiben als Text stehen",
		}
	}

	schrift, da := vorlage.Symbolschriften[name]
	if !da {
		return nil, render.Symbolmasse{}, &domain.Befund{
			Ort: feld.Feld,
			Meldung: "die Symbolschrift " + name + " liegt nicht vor; " +
				lesbar(text) + " bleibt als Text stehen",
		}
	}

	codes, unbekannt := zerlegeInSymbole(text, schrift)

	// **Ein unbekanntes Zeichen wird übersprungen, nicht verworfen.**
	//
	//	// unknown code, draw single character as text
	//	//out.push_back(DrawableSymbol(text.substr(pos, 1), _(""), defaultSymbol()));
	//	pos += 1;
	//
	// `SymbolFont::split`, `src/data/symbol_font.cpp`, Zeilen 192–231,
	// Fassung f9d9356 vom 2021-07-06 — **der `push_back` ist
	// auskommentiert**. MSE zeichnet die übrigen Symbole und lässt das
	// Zeichen weg.
	//
	// **Hier wurde bis zum 2026-08-19 der ganze Zug zu Text**, und aus
	// Manasymbolen wurden Buchstaben. Über alle 23.168 kanonischen Karten
	// waren das drei Befunde
	// (#517).
	//
	// **Der Befund bleibt trotzdem**, und er ist von der Art `Vorlage`: die
	// Vorlage hat ein Zeichen in einen Symbolblock gesetzt, das dort nichts zu
	// suchen hat. MSE schweigt dazu; hier steht es, ohne das Bild zu ändern.
	var befund *domain.Befund

	if unbekannt != "" {
		befund = &domain.Befund{
			Ort: feld.Feld,
			Meldung: "die Symbolschrift " + name + " führt kein Symbol für " +
				lesbar(unbekannt) + "; es wird übersprungen",
			Art: domain.ArtVorlage,
		}
	}

	// **Ohne ein einziges Symbol bleibt es Text.** Ein Symbolzug ohne Symbole
	// zeichnete nichts, und ein leerer Kasten ist schlechter als der Text, den
	// die Vorlage hineingeschrieben hat (§8).
	if len(codes) == 0 {
		return nil, render.Symbolmasse{}, befund
	}

	return symbolbilder(ctx, codes, name, umgebung),
		symbolmasse(ctx, feld, rahmen, schrift, umgebung), befund
}
