package msescript

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Zwei Funktionen, die eine Liste liefern beziehungsweise in einer suchen.
//
// Beide sind an `M15-Magic-Pack` ausgezählt
// (#216).
func init() {
	eintragen("break_text", zerlegeText)
	eintragen("position", stelleVon)
}

// zerlegeText gibt die **Treffer** eines Musters als Liste —
// `break_text(card.casting_cost, match:"H(/[WUBRGC])?|…")`.
//
// # Der Unterschied zu `split_text`
//
// `split_text` gibt, was **zwischen** den Treffern steht; `break_text` gibt die
// Treffer selbst. Die Vorlagen zerlegen damit Manakosten in Symbole: das Muster
// beschreibt ein Symbol, und die Liste ist die Folge der Symbole.
//
// `filter_text` mit demselben Muster gäbe dieselben Stücke als **einen** Text.
// Die drei sind damit dieselbe Suche mit drei Ausgaben, und jede steht in den
// Vorlagen.
func zerlegeText(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("match")
	if !benannt {
		return Nichts, ruf.fehlerAn("break_text braucht ein Argument `match:`")
	}

	if err := ruf.eineEingabe("break_text"); err != nil {
		return Nichts, err
	}

	muster, err := UebersetzeMuster(roh.AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("break_text bekam " + err.Error())
	}

	muster, err = mitKontext(ruf, muster, "break_text")
	if err != nil {
		return Nichts, err
	}

	treffer, err := muster.AlleTreffer(ruf.Stellung[0].AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("break_text: " + err.Error())
	}

	eintraege := make([]Eintrag, 0, len(treffer))
	for stelle, stueck := range treffer {
		eintraege = append(eintraege, Eintrag{Name: strconv.Itoa(stelle), Wert: Text(stueck)})
	}

	return Wert{Art: ArtAbbildung, Eintraege: eintraege}, nil
}

// stelleVon sucht einen Wert in einer Liste — `position(of: x, in: liste)`.
//
// # Was sie zurückgibt
//
// Die **Stelle**, von null an. Was nicht dasteht, ergibt `-1`.
//
// **Nicht null und nicht ein Fehler.** Null wäre die erste Stelle und damit ein
// Treffer, den es nicht gab; ein Fehler bräche die ganze Vorlage ab, obwohl
// „kommt nicht vor" eine gewöhnliche Antwort ist — `magic-xml` baut daraus
// einen Dateinamen und rechnet mit beidem.
//
// # Verglichen wird über die Textform
//
// `position(of: [card.color_category], in: indexes)` sucht mit einer
// einelementigen Liste in einer Liste von Listen. Ein Vergleich über die
// Struktur unterschiede dabei zwischen `["a"]` und `"a"`; die Vorlagen meinen
// dasselbe. Die Textform ist die Form, in der beide gleich sind — und sie ist
// die, mit der die Vorlage danach weiterrechnet.
func stelleVon(ctx context.Context, ruf Ruf) (Wert, error) {
	gesucht, hat := ruf.Benannter("of")
	if !hat {
		return Nichts, ruf.fehlerAn("position braucht ein Argument `of:`")
	}

	liste, hat := ruf.Benannter("in")
	if !hat {
		return Nichts, ruf.fehlerAn("position braucht ein Argument `in:`")
	}

	// **Mit `order_by:` oder `filter:` zählt die Karte im geordneten Set**
	// (#2006) — MSE prüft das nach dem Textfall und vor dem ungeordneten.
	//
	//	if (of_t == SCRIPT_STRING || in_t == SCRIPT_STRING) { … }
	//	else if (order_by || filter) { … positionOfCard(…) }
	//
	// `src/script/functions/basic.cpp`, Zeilen 449–473, Fassung f9d9356 vom
	// 2021-07-06.
	if gesucht.Art == ArtTextwert || liste.Art == ArtTextwert {
		return stelleImText(ruf, gesucht, liste)
	}

	ordnung, filter := ordnungsangaben(ruf)
	if ordnung != nil || filter != nil {
		return stelleImSet(ctx, ruf, gesucht, liste, ordnung, filter)
	}

	nadel := gesucht.AlsText()

	for stelle := range liste.Eintraege {
		if liste.Eintraege[stelle].Wert.AlsText() == nadel {
			return Zahl(float64(stelle)), nil
		}
	}

	return Zahl(-1), nil
}

// stelleImText ist MSEs Textsuche — **`in` in `of`**, nicht umgekehrt
// (#2006):
//
//	return (int)of->toString().find(in->toString()); // (int)npos == -1
//
// `src/script/functions/basic.cpp`, Zeilen 456–459, Fassung f9d9356 vom
// 2021-07-06, gefragt am Orakel. Sie greift, sobald **eine** Seite Text ist;
// eine Liste auf der anderen hat keine Textform, und MSE bricht ab („Can't
// convert from collection to string"). Hier stand bis dahin die umgekehrte
// Suche, erschlossen aus den Namen der Argumente.
func stelleImText(ruf Ruf, gesucht, liste Wert) (Wert, error) {
	if gesucht.Art == ArtAbbildung || liste.Art == ArtAbbildung {
		return Nichts, ruf.fehlerAn("position kann eine Liste nicht als Text lesen")
	}

	return Zahl(float64(textstelle(gesucht.AlsText(), liste.AlsText()))), nil
}

// textstelle ist `String::find` in Zeichen und nicht in Bytes — wxString zählt
// Zeichen, und eine Stelle in Bytes stimmte nach dem ersten Umlaut nicht mehr.
func textstelle(text, teil string) int {
	bei := strings.Index(text, teil)
	if bei < 0 {
		return -1
	}

	return utf8.RuneCountInString(text[:bei])
}

// ordnungsangaben liest `order_by:` und `filter:` aus dem Ruf.
//
// **`nil` als Filter heißt: keiner** — `if (filter == script_nil) filter =
// ScriptValueP();`, `basic.cpp` Zeile 528. `set_filter()` aus
// `magic.mse-game` gibt das für jede Karte ohne eigenen Zweig.
func ordnungsangaben(ruf Ruf) (ordnung, filter *Wert) {
	if wert, da := ruf.Benannter("order_by"); da {
		ordnung = &wert
	}

	if wert, da := ruf.Benannter("filter"); da && wert.Art != ArtNichts {
		filter = &wert
	}

	return ordnung, filter
}
