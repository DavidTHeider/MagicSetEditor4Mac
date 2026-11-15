package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"

// symbollaeufe teilt einen Text in **Symbolläufe und Textläufe**.
//
//	if (style.always_symbol && style.symbol_font.valid()) {
//	  // mixed symbols/text, autodetected by symbol font
//	  size_t text_pos = 0;
//	  size_t pos = 0;
//	  while (pos < end-start) {
//	    if (size_t n = style.symbol_font.font->recognizePrefix(content,pos)) {
//	      // at 'pos' there are n symbol font characters
//	      if (text_pos < pos) { … }   // das Stück davor wird Text
//	      pos += n;
//	      text_pos = pos;
//	    } else {
//	      pos += 1;                   // ein Zeichen weiter, es bleibt Text
//	    }
//	  }
//
// `src/render/text/element.cpp`, Zeilen 244–262, und `recognizePrefix`,
// `src/data/symbol_font.cpp`, Zeilen 234–261. Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum das eine eigene Frage ist
//
// **Bis zum 2026-08-18 galt „alles oder nichts"**: entweder bestand der ganze
// Wert aus Symbolcodes — dann wurde er ein Symbolzug —, oder er wurde
// vollständig zu Text, mit einem Befund daneben. Aus Manasymbolen wurden
// Buchstaben, sobald ein einziges Zeichen dazwischenstand. Über alle 23.168
// kanonischen Karten waren das vier Befunde
// (#508).
//
// **Der Symbolzug bleibt**, wo alles Symbol ist: er misst und setzt anders als
// ein Textzug, und die gewöhnlichen Manakosten gehen weiter dort durch. Diese
// Aufteilung greift erst, wenn es **nicht** aufgeht.
//
// # Was sie nicht tut
//
// **Sie erzeugt keine Auszeichnungen.** Kursiv, fett und die Blockgrenzen des
// umgebenden Abschnitts reisen unverändert an jedem Lauf mit; ein Symbol
// mitten in kursivem Text bleibt in dessen Abschnitt.
func symbollaeufe(teil rohabschnitt, schrift msepaket.Symbolschrift) []rohabschnitt {
	zeichen := []rune(teil.Text)

	var (
		laeufe   []rohabschnitt
		textseit int
	)

	// anhaengen schließt den laufenden Textlauf ab.
	anhaengen := func(bis int, symbolisch bool) {
		if bis <= textseit {
			return
		}

		lauf := teil
		lauf.Text = string(zeichen[textseit:bis])
		lauf.Symbolisch = symbolisch

		laeufe = append(laeufe, lauf)
	}

	stelle := 0

	for stelle < len(zeichen) {
		laenge := erkannteVorsilbe(string(zeichen[stelle:]), schrift)
		if laenge == 0 {
			stelle++

			continue
		}

		// **Das Stück davor wird Text** — `if (text_pos < pos)`.
		anhaengen(stelle, false)

		textseit = stelle
		stelle += laenge

		anhaengen(stelle, true)

		textseit = stelle
	}

	anhaengen(len(zeichen), false)

	// **Kein Zweig für „nichts herausgekommen".** Bei nicht leerem Text hängt
	// der Aufruf oben immer mindestens einen Lauf an, und den leeren Text
	// filtert [aufgeteilt] vorher — ein Weg, den kein Aufruf betreten kann,
	// ist ein Zweig zu viel (§3).
	return laeufe
}

// erkannteVorsilbe ist `recognizePrefix`: wie viele **Zeichen** ab hier eine
// ununterbrochene Folge von Symbolcodes bilden.
//
//	size_t SymbolFont::recognizePrefix(const String& text, size_t start) const {
//	  size_t pos;
//	  for (pos = start ; pos < text.size() ; ) {
//	    … if (is_substr(text, pos, sym->code)) { pos += sym->code.size(); goto next_symbol; }
//	    break;
//	  next_symbol:;
//	  }
//	  return pos - start;
//	}
//
// `src/data/symbol_font.cpp`, Zeilen 234–261, Fassung f9d9356 vom
// 2021-07-06 — **es bricht beim ersten unerkannten Zeichen ab**, es
// überspringt nicht.
func erkannteVorsilbe(rest string, schrift msepaket.Symbolschrift) int {
	var zeichen int

	for rest != "" {
		_, _, laenge := ersterTreffer(rest, schrift)
		if laenge == 0 {
			return zeichen
		}

		zeichen += len([]rune(rest[:laenge]))
		rest = rest[laenge:]
	}

	return zeichen
}

// aufgeteilt wendet [symbollaeufe] auf jeden noch nicht symbolischen Abschnitt
// an.
//
// **Ein `<sym>`-Abschnitt bleibt, wie er ist.** Er sagt schon, dass er Symbole
// trägt; ihn ein zweites Mal zu befragen hieße, der Auszeichnung nicht zu
// glauben.
func aufgeteilt(roh []rohabschnitt, schrift msepaket.Symbolschrift) []rohabschnitt {
	aufgeteilt := make([]rohabschnitt, 0, len(roh))

	for _, teil := range roh {
		if teil.Symbolisch || teil.Text == "" {
			aufgeteilt = append(aufgeteilt, teil)

			continue
		}

		aufgeteilt = append(aufgeteilt, symbollaeufe(teil, schrift)...)
	}

	return aufgeteilt
}
