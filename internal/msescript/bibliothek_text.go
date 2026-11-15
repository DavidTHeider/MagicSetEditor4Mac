package msescript

import (
	"context"
	"strings"
)

// Die Funktionen, die auf Kartentext arbeiten.
func init() {
	eintragen("remove_tags", ohneTags)
	eintragen("contains", enthaelt)
	eintragen("replace", ersetze)

	// `replace_text` ist dasselbe unter einem zweiten Namen: `magic.mse-game`
	// schreibt `replace_text@(match: "(?i)(Legendary|Basic|…)[- ]?",
	// replace:"")` — gleiche Argumente, gleiche Wirkung
	// (#216).
	eintragen("replace_text", ersetze)
	eintragen("filter_text", filtereText)
}

// Benannter sucht ein benanntes Argument.
func (r Ruf) Benannter(name string) (Wert, bool) {
	for i := range r.Benannt {
		if r.Benannt[i].Name == name {
			return r.Benannt[i].Wert, true
		}
	}

	return Nichts, false
}

// ohneTags — `remove_tags(card.type_1)` aus `yugioh-series10` Zeile 732.
//
// Kartentext trägt Auszeichnungen in spitzen Klammern: `<i-flavor>…</i-flavor>`
// steht in `magic-m15` Zeile 290. Für die Breitenrechnung dort zählt der Text,
// nicht die Auszeichnung.
//
// **Eine offene Klammer ohne schließende bleibt stehen.** Sie ist dann keine
// Auszeichnung, sondern ein Kleinerzeichen, das jemand getippt hat — und ein
// Kartentext, der ab dem ersten `<` verschwindet, wäre schlimmer als einer mit
// einem Zeichen zu viel.
func ohneTags(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("remove_tags"); err != nil {
		return Nichts, err
	}

	quelle := ruf.Stellung[0].AlsText()

	var gebaut strings.Builder

	gebaut.Grow(len(quelle))

	for {
		auf := strings.IndexByte(quelle, '<')
		if auf < 0 {
			break
		}

		zu := strings.IndexByte(quelle[auf:], '>')
		if zu < 0 {
			break
		}

		gebaut.WriteString(quelle[:auf])
		quelle = quelle[auf+zu+1:]
	}

	gebaut.WriteString(quelle)

	return Text(gebaut.String()), nil
}

// enthaelt — `contains(style, match:"walker")` aus `magic.mse-game/game_script`.
//
// # `match:` ist eine feste Zeichenkette, kein Muster
//
// Das ist der Befund dieses Werkstücks, und er kommt aus `auto_replace`: dort
// steht `match: (C)` mit `replace: ©` und `match: **` mit `replace: AUTOREM>`.
// Als reguläre Ausdrücke wäre das erste eine Gruppe um ein `C` und das zweite
// gar nicht übersetzbar. In 1541 Zeilen kommt **kein einziger** Rückstrich und
// kein `[^`, `.*` oder `\d` vor.
//
// Daraus folgt die dritte Zusage von #25 von selbst: ein bösartiger regulärer
// Ausdruck kann die Anwendung nicht anhalten, **weil es keinen gibt**. Eine
// Zeitschranke um eine Mustererkennung, die nicht stattfindet, wäre eine
// Vorkehrung gegen eine erfundene Gefahr.
func enthaelt(_ context.Context, ruf Ruf) (Wert, error) {
	muster, benannt := ruf.Benannter("match")
	if !benannt {
		return Nichts, ruf.fehlerAn("contains braucht ein Argument `match:`")
	}

	if err := ruf.eineEingabe("contains"); err != nil {
		return Nichts, err
	}

	return Wahrheit(strings.Contains(ruf.Stellung[0].AlsText(), muster.AlsText())), nil
}

// ersetze — `replace@(match: "<soft-line>\n", replace: "")` aus
// `magic-m15-godzilla-planeswalker.mse-style` und acht Exportvorlagen von
// `M15-Magic-Pack`.
//
// **Sie ist die eine Funktion, die die meisten Vorlagen blockiert**
// (#51):
// acht von neunzehn scheiterten allein an ihrem Namen.
//
// # Welcher Muster-Dialekt
//
// Gos `regexp` — RE2. **Das ist eine Annahme, und sie gehört benannt**: MSE
// selbst liegt wxWidgets bei, und dessen Ausdrücke sind PCRE-nah. Gemeinsam
// sind beiden die Formen, die in den Vorlagen vorkommen — Zeichenklassen,
// Alternativen, `\b`, Quantoren. **Rückwärtsverweise und Vorausschau kennt RE2
// nicht**; ein Muster, das sie benutzt, wird hier zum Fehler statt still
// anders zu wirken, und das ist die bessere von zwei Antworten.
func ersetze(ctx context.Context, ruf Ruf) (Wert, error) {
	if fehler := ruf.eineEingabe("replace/replace_text"); fehler != nil {
		return Nichts, fehler
	}

	muster, ersatz, err := ersetzungsteile(ruf)
	if err != nil {
		return Nichts, err
	}

	rekursiv := false
	if wert, da := ruf.Benannter("recursive"); da {
		rekursiv, _ = wert.AlsWahrheit()
	}

	// **Eine Funktion wird gerufen, ein Text eingesetzt.**
	//
	//	replacer.replacement_function = ctx.getVariable(SCRIPT_VAR_replace);
	//	if (replacer.replacement_function->type() != SCRIPT_FUNCTION) {
	//	  replacer.replacement_string = …->toString();
	//	  replacer.replacement_function = ScriptValueP();
	//	}
	//
	// `src/script/functions/regex.cpp`, Zeilen 125–129, Fassung f9d9356 vom
	// 2021-07-06.
	if ersatz.Art == ArtFunktion && ruf.Umgebung != nil {
		gebaut, fehler := muster.ErsetzeMitFunktion(ctx, ruf.Stellung[0].AlsText(), ruf.Umgebung,
			func(ctx context.Context, umgebung *Umgebung) (Wert, error) {
				return RufeFunktion(ctx, ersatz, umgebung)
			}, rekursiv)
		if fehler != nil {
			return Nichts, ruf.fehlerAn("replace: " + fehler.Error())
		}

		return Text(gebaut), nil
	}

	// **Auch ein Text kann rekursiv sein** — MSE unterscheidet die beiden
	// Wege erst hinter der Rekursion (Zeilen 100–104 gegen 135–140).
	if rekursiv {
		gebaut, fehler := muster.ErsetzeRekursiv(ruf.Stellung[0].AlsText(), ersatz.AlsText())
		if fehler != nil {
			return Nichts, ruf.fehlerAn("replace: " + fehler.Error())
		}

		return Text(gebaut), nil
	}

	// Was `\1` und `&` im Ersatztext bedeuten, steht bei [Muster.Ersetze].
	gebaut, err := muster.Ersetze(ruf.Stellung[0].AlsText(), ersatz.AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("replace: " + err.Error())
	}

	return Text(gebaut), nil
}

// filtereText — `filter_text@(match: "\n")` aus
// `magic-m15-godzilla-planeswalker.mse-style`.
//
// **Sie behält, was das Muster trifft** — das Gegenstück zu [ersetze]. Die
// Vorlage zählt damit Absätze: sie filtert die Umbrüche heraus und misst, wie
// viele übrig bleiben.
func filtereText(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("match")
	if !benannt {
		return Nichts, ruf.fehlerAn("filter_text braucht ein Argument `match:`")
	}

	if err := ruf.eineEingabe("filter_text"); err != nil {
		return Nichts, err
	}

	muster, err := UebersetzeMuster(roh.AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("filter_text bekam " + err.Error())
	}

	muster, err = mitKontext(ruf, muster, "filter_text")
	if err != nil {
		return Nichts, err
	}

	treffer, err := muster.AlleTreffer(ruf.Stellung[0].AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("filter_text: " + err.Error())
	}

	return Text(strings.Join(treffer, "")), nil
}

// ersetzungsteile liest `match:` und `replace:` eines Rufs.
func ersetzungsteile(ruf Ruf) (*Muster, Wert, error) {
	roh, benannt := ruf.Benannter("match")
	if !benannt {
		return nil, Nichts, ruf.fehlerAn("replace braucht ein Argument `match:`")
	}

	// **`replace:` darf fehlen und heißt dann „weg damit".** Die Vorlagen
	// schreiben es meist als `replace: ""`; eine, die es wegließe, meinte
	// dasselbe.
	ersatz, _ := ruf.Benannter("replace")

	muster, err := UebersetzeMuster(roh.AlsText())
	if err != nil {
		return nil, Nichts, ruf.fehlerAn("replace bekam " + err.Error())
	}

	muster, err = mitKontext(ruf, muster, "replace")
	if err != nil {
		return nil, Nichts, err
	}

	return muster, ersatz, nil
}

// mitKontext hängt `in_context:` an ein Muster, wenn der Ruf ihn nennt —
// `replace`, `filter_text` und `break_text` nehmen ihn
// (`src/script/functions/regex.cpp`, Zeilen 117–119, 157 und 189, Fassung
// f9d9356 vom 2021-07-06; #2315). **Bis dahin wurde er hier nicht gelesen**,
// und jede Ersetzung mit Kontext wirkte wie eine ohne.
func mitKontext(ruf Ruf, muster *Muster, wer string) (*Muster, error) {
	roh, benannt := ruf.Benannter("in_context")
	if !benannt {
		return muster, nil
	}

	kontext, err := UebersetzeMuster(roh.AlsText())
	if err != nil {
		return nil, ruf.fehlerAn(wer + " bekam in_context: " + err.Error())
	}

	return muster.MitKontext(kontext), nil
}
