package msescript

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Die Funktionen, die in den gesichteten Vorlagen auf Zeichenketten und Zahlen
// angewendet werden. Jede nennt ihre Fundstelle.
func init() {
	eintragen("to_lower", zuKlein)
	eintragen("to_real", zuZahl)
	eintragen("to_text", zuText)
	eintragen("length", laenge)
	eintragen("max", groesster)
	eintragen("min", kleinster)
	eintragen("number_of_items", zahlDerEintraege)
}

// zuKlein — `to_lower(…)` aus `magic.mse-game/game_script` Zeile 62.
//
// Dort wird ein Stilname kleingeschrieben, bevor er mit `contains` gegen feste
// Muster geprüft wird.
func zuKlein(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_lower"); err != nil {
		return Nichts, err
	}

	return Text(strings.ToLower(ruf.Stellung[0].AlsText())), nil
}

// zuText — `to_text(card.casting_cost)` aus jeder der 19 Exportvorlagen von
// `M15-Magic-Pack`.
//
// **Sie ist die eine Funktion, ohne die keine einzige Vorlage läuft**
// (#51). Die
// Textform steht an [Wert.AlsText]: eine Zahl ohne überflüssige Nullen, eine
// Wahrheit als `true`/`false`.
//
// **Und sie nimmt die Auszeichnung heraus** — ported
// (#2006):
//
//	SCRIPT_RETURN(untag_hide_sep(input));
//
// `src/script/functions/export.cpp`, Zeilen 361–364, Fassung f9d9356 vom
// 2021-07-06. Bis dahin gab sie den Text, wie er war; `magic-mws` schrieb
// deshalb `<i-flavor>` in die Ausfuhr, wo MSE2 den bloßen Satz schreibt.
func zuText(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_text"); err != nil {
		return Nichts, err
	}

	return Text(ohneAuszeichnung(ruf.Stellung[0].AlsText())), nil
}

// ohneAuszeichnung ist MSEs `untag_hide_sep`: erst fällt jeder weiche Trenner
// samt Inhalt weg, dann jede Marke.
//
//	return untag(remove_tag_contents(str,_("<sep-soft")));
//
// `src/util/tagged_string.cpp`, Zeilen 33–47 und 65–67, Fassung f9d9356 vom
// 2021-07-06. **Anders als `remove_tags` hier**: `untag` kennt keine offene
// Klammer, die stehen bleibt — ab einem `<` ohne `>` fällt der Rest weg.
// Das geschützte Kleinerzeichen wird dabei wieder eines ([entschuetztesZeichen]).
func ohneAuszeichnung(text string) string {
	var gebaut strings.Builder

	inMarke := false

	for _, zeichen := range ohneMarkeninhalt(text, "<sep-soft") {
		switch {
		case zeichen == '<':
			inMarke = true
		case !inMarke:
			gebaut.WriteRune(entschuetztesZeichen(zeichen))
		case zeichen == '>':
			inMarke = false
		}
	}

	return gebaut.String()
}

// zuZahl ist `to_real` — `to_real(styling.imageheight)` aus
// `kamen-decade.mse-style/style`.
//
//	if (t == SCRIPT_BOOL)        result = input->toBool() ? 1.0 : 0.0;
//	else if (t == SCRIPT_COLOR)  result = (c.Red() + c.Blue() + c.Green()) / 3.0;
//	else if (t == SCRIPT_STRING) { leer → 0.0, unlesbar → delay_error(…) }
//	else                         result = input->toDouble();
//
// `src/script/functions/basic.cpp`, Zeilen 135–159, Fassung f9d9356 vom
// 2021-07-06. **Bis F02 stand hier eine eigene Regel**
// (#2244):
// ein unlesbarer Wert brach mit Stelle ab, und `FORMATE.md` führte als offen,
// was MSE tut. Nachgelesen und am Orakel gefragt, ist es ein **verzögerter**
// Fehler — `to_real("x") or else 7` ist 7 —, und `true`, `""`, `nil` und eine
// Farbe sind Zahlen.
//
// **Leerraum vorn ja, hinten nicht**: `wxString::ToDouble` liest über
// `strtod`, das führenden Leerraum überspringt, und verlangt, dass die ganze
// Zeichenkette verbraucht ist — `to_real(" 2.5")` ist 2,5, `to_real("2.5 ")`
// ein Fehler (Orakel, am 2026-10-11).
func zuZahl(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_real"); err != nil {
		return Nichts, err
	}

	wert := ruf.Stellung[0]

	switch wert.Art {
	case ArtZahlwert:
		return wert, nil

	case ArtWahrheitswert:
		if wert.Wahrheit {
			return Zahl(1), nil
		}

		return Zahl(0), nil

	case ArtFarbe:
		summe, _ := kanalsumme(wert)

		return Zahl(float64(summe) / 3), nil

	case ArtNichts:
		return Zahl(0), nil

	case ArtTextwert, ArtAbbildung, ArtFunktion:
	}

	roh := wert.AlsText()
	if roh == "" {
		return Zahl(0), nil
	}

	if zahl, lesbar := gelesenerZahlwert(strings.TrimLeft(roh, " \t\n\v\f\r")); lesbar {
		return Zahl(zahl), nil
	}

	return Fehlerwert("to_real bekam " + strconv.Quote(roh) + " — das ist keine Zahl"), nil
}

// laenge — `length(remove_tags(card.type_1))` aus `yugioh-series10` Zeile 732.
//
// **In Zeichen, nicht in Bytes.** Das Ergebnis wird dort mit 17 malgenommen und
// als Feldbreite verwendet; ein japanischer Kartentyp bekäme in Bytes die
// dreifache Breite.
func laenge(ctx context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("length"); err != nil {
		return Nichts, err
	}

	wert := ruf.Stellung[0]
	if wert.Kennung == "set" {
		return zahlImSet(ctx, ruf, wert)
	}

	if wert.Art == ArtAbbildung {
		return Zahl(float64(len(wert.Eintraege))), nil
	}

	return Zahl(float64(utf8.RuneCountInString(wert.AlsText()))), nil
}

// zahlDerEintraege — `number_of_items(in: colors)` aus
// `magic-blends.mse-include/new-blends`.
//
// # Dieselbe Rechnung wie [laenge], ein anderes Argument
//
// Herkunft: `src/script/functions/basic.cpp:568` von MSE2 gibt sie an
// `script_length_of` weiter — genau wie `length`. Der einzige Unterschied ist
// das Argument: sie liest **`in:`**, nicht `input`.
//
// MSE führt sie mit dem Kommentar `// deprecated`. Im Pack steht sie trotzdem,
// und ein Paket von 2007 kümmert das nicht.
//
// **Woran es hing:** `color_background` zählt damit die Farben einer Karte
// (`color_count := … number_of_items(in: colors)`). Ohne sie fiel `card color →
// image` aus — das Feld, das den Rahmen zeichnet
// (#240).
//
// Fassung f9d9356 vom 2021-07-06.
func zahlDerEintraege(ctx context.Context, ruf Ruf) (Wert, error) {
	wert, benannt := ruf.Benannter("in")
	if !benannt {
		return Nichts, ruf.fehlerAn("number_of_items braucht ein Argument `in:`")
	}

	return laenge(ctx, Ruf{
		Stellung: []Wert{wert}, Benannt: ruf.Benannt, Umgebung: ruf.Umgebung,
		Zeile: ruf.Zeile, Spalte: ruf.Spalte,
	})
}

// groesster und kleinster — `max(22, card_style.rarity.content_width)` und
// `min(14, font_size())` aus `magic-m15`.
//
// # Form und Herkunft
//
// `src/script/parser.cpp:632` von MSE2 behandelt beide gemeinsam: **mit
// Klammer vorangestellt und beliebig vielen Argumenten**, paarweise
// zusammengefaltet. Eine Infix-Form gibt es nicht, auch wenn `cost min 5` im
// Pack danach aussieht — das ist ein `min(cost, 5)` über zwei Zeilen.
//
// # Sie rechnen wie jedes Rechenzeichen
//
// Über [Wert.AlsZahl], nicht über eine eigene Prüfung auf `ArtZahlwert`: MSE
// faltet sie mit demselben `OPERATOR_FUN_DI`, das auch `-` und `*` benutzen
// (`context.cpp:533`). Eine eigene, engere Prüfung wiese `max(0,
// move_typebox())` ab, wo MSE null rechnet.
//
// Fassung f9d9356 vom 2021-07-06.
func groesster(_ context.Context, ruf Ruf) (Wert, error) {
	return gefaltet(ruf, "max", func(bisher, naechste float64) bool { return naechste > bisher })
}

func kleinster(_ context.Context, ruf Ruf) (Wert, error) {
	return gefaltet(ruf, "min", func(bisher, naechste float64) bool { return naechste < bisher })
}

func gefaltet(ruf Ruf, name string, besser func(bisher, naechste float64) bool) (Wert, error) {
	if len(ruf.Stellung) == 0 {
		return Nichts, ruf.fehlerAn(name + " erwartet mindestens ein Argument, bekam keins")
	}

	gewaehlt, err := ruf.Stellung[0].AlsZahl()
	if err != nil {
		return Nichts, ruf.fehlerAn(name + " rechnet mit Zahlen, " + err.Error())
	}

	for _, wert := range ruf.Stellung[1:] {
		zahl, err := wert.AlsZahl()
		if err != nil {
			return Nichts, ruf.fehlerAn(name + " rechnet mit Zahlen, " + err.Error())
		}

		if besser(gewaehlt, zahl) {
			gewaehlt = zahl
		}
	}

	return Zahl(gewaehlt), nil
}
