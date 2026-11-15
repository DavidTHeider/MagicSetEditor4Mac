package msescript

import (
	"context"
	"math"
	"strconv"
	"strings"
)

// `to_int` — die Ganzzahl zu einem Wert.
//
// # Erschlossen, nicht erfunden
//
// Sie steht in den Stilen von `M15-Magic-Pack` dort, wo aus einer gerechneten
// Größe eine Stelle wird — `to_int(…)` neben `to_real`, das die Bibliothek
// schon führt (#214).
func init() {
	eintragen("to_int", zuGanzzahl)

	eintragen("to_number", zuNummer)
}

// zuGanzzahl schneidet den Nachkommateil ab.
//
// **Abschneiden und nicht runden.** `to_int` steht neben `to_real`, und beide
// erscheinen in Ausdrücken über Bildpunkte: eine Ebene, die bei 10,6 auf 11
// gerundet würde, säße einen Punkt tiefer als die daneben, die 10,4 rechnet.
// MSE-Stile rechnen Ränder als Differenz — dort ist Abschneiden das
// Verhalten, das die Summe erhält.
func zuGanzzahl(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_int"); err != nil {
		return Nichts, err
	}

	wert := ruf.Stellung[0]
	if wert.Art == ArtZahlwert {
		return Zahl(math.Trunc(wert.Zahl)), nil
	}

	// **Ein Wahrheitswert wird zu eins oder null.**
	//
	//	if (t == SCRIPT_BOOL) {
	//	  result = input->toBool() ? 1 : 0;
	//	}
	//
	// Zeilen 111–112. **Nur hier, nicht überall**: `ScriptBool` überschreibt
	// `toInt` nicht, und der Kommentar dort lautet „bools don't autoconvert to
	// int". Die Umwandlung steht in der Funktion, nicht am Typ — und die
	// Zeichenkette `"true"` bleibt deshalb ein Fehlerwert.
	//
	// **226 Befunde** über die 190 Stile von `M15-Magic-Pack`
	// (#345).
	if wert.Art == ArtWahrheitswert {
		if wert.Wahrheit {
			return Zahl(1), nil
		}

		return Zahl(0), nil
	}

	// **Eine Farbe ist das Mittel ihrer drei Kanäle**, ganzzahlig und ohne
	// die Deckung (Zeilen 113–115) — siehe [kanalsumme].
	if summe, ist := kanalsumme(wert); ist {
		return Zahl(float64(summe / 3)), nil
	}

	// **Ein leerer Text ist null** (Zeile 121), alles andere ein **verzögerter**
	// Fehler:
	//
	//	} else if (str.empty()) {
	//	  result = 0;
	//	} else {
	//	  return delay_error(ScriptErrorConversion(str, …, _TYPE_("integer")));
	//	}
	//
	// `src/script/functions/basic.cpp`, Zeilen 116–124, Fassung f9d9356 vom
	// 2021-07-06.
	//
	// **Verzögert und nicht sofort**: die Vorlagen schreiben
	// `to_int(x) or else 0`, und `or else` fängt genau diesen Wert
	// (#345).
	roh := strings.TrimSpace(wert.AlsText())
	if roh == "" {
		return Zahl(0), nil
	}

	// **Und eine Kommazahl als Zeichenkette wird abgeschnitten, wo MSE
	// wirft.** `to_int("1.5")` gibt dort *Can’t convert "1.5" … to integer
	// number*, hier eine Eins — am Orakel gefragt, nicht erschlossen.
	//
	// **Die Abweichung bleibt stehen, gemessen folgenlos**: über die 376
	// Stile und 8.915 Feldstile des Packs bekommt `to_int` **kein einziges
	// Mal** eine Zeichenkette, die eine Kommazahl ist
	// (#271,
	// `docs/FORMATE.md`).
	//
	// **Und `5 / 2` weicht nicht ab** — beide geben 2,5. Das Issue führte es
	// bis zum 2026-08-20 als zweiten Fall; er ist widerlegt.
	zahl, lesbar := gelesenerZahlwert(roh)
	if !lesbar {
		return Fehlerwert("to_int bekam " + strconv.Quote(wert.AlsText()) +
			" — das ist keine Zahl"), nil
	}

	return Zahl(math.Trunc(zahl)), nil
}

// zuNummer ist `to_number` — und es ist **nicht** `to_int`.
//
// # Was MSE tut
//
//	if (str.ToLong(&l))        SCRIPT_RETURN((int)l);
//	else if (str.ToDouble(&d)) SCRIPT_RETURN((double)d);
//	else if (str.empty())      SCRIPT_RETURN(0);
//	else                       return delay_error(…);
//
// `src/script/functions/basic.cpp`, Zeilen 160–190, Fassung f9d9356 vom
// 2021-07-06. **Erst ganz, dann gebrochen** — und die ganze Zahl wird auf
// C++-`int` verengt.
//
// # Warum es zwei Funktionen sind
//
// **Der Unterschied ist gemessen, nicht erfunden.** `to_real` und `to_int`
// stehen in den Stilen hinter einer Prüfung auf leer
// (`if styling.imagetop != "" then to_real(…) else 420`) — dort wäre eine
// erfundene Null ein Bild an der oberen Kante, und niemand sähe warum.
//
// `to_number` steht in den Exportvorlagen **ohne** eine solche Prüfung:
// `magic-field-test` und `magic-lackeybot` rufen es auf ein `substring`, das
// bei einer Karte ohne diesen Wert leer ist. Ein Fehler bräche dort die ganze
// Ausfuhr an einer Karte ab, der schlicht ein Feld fehlt.
//
// # Bis zum 2026-08-22 reichte es an `to_int` durch
//
// Und schnitt damit ab: `to_number("1.5")` gab **1**, MSE gibt **1,5**. Der
// Nachlauf zum Zweitblick über Phase P hat es angestrichen
// (#859).
//
// **Aus der Zeichenkette wird hier exakt verengt**, denn dort steht eine
// ganze Zahl und kein `float64`: `to_number("9223372036854775807")` gibt −1
// wie MSE. `format` kann das nicht — dort kommt die Zahl schon als `float64`
// an, und die unteren Bits fehlen bereits.
func zuNummer(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_number"); err != nil {
		return Nichts, err
	}

	wert := ruf.Stellung[0]

	switch wert.Art {
	// **Eine Zahl bleibt sie selbst** — MSEs `SCRIPT_DOUBLE`-Zweig gibt sie
	// unverändert zurück, und eine ganze läuft über `toString` und `ToLong`
	// auf denselben Wert.
	case ArtZahlwert:
		return wert, nil

	case ArtWahrheitswert:
		if wert.Wahrheit {
			return Zahl(1), nil
		}

		return Zahl(0), nil

	case ArtNichts:
		return Zahl(0), nil

	// `SCRIPT_RETURN( (c.Red() + c.Blue() + c.Green()) / 3 );` — ganzzahlig
	// wie bei `to_int` (Zeilen 167–169).
	case ArtFarbe:
		summe, _ := kanalsumme(wert)

		return Zahl(float64(summe / 3)), nil

	case ArtTextwert, ArtAbbildung, ArtFunktion:
	}

	return ausText(strings.TrimSpace(wert.AlsText()))
}

// ausText liest eine Zahl aus einer Zeichenkette — erst ganz, dann gebrochen.
func ausText(roh string) (Wert, error) {
	// **`ParseInt` und nicht `ParseFloat` zuerst**: nur so bleibt die ganze
	// Zahl exakt, bis sie verengt wird. Über `float64` wären die unteren Bits
	// jenseits von 2⁵³ schon weg.
	if ganz, err := strconv.ParseInt(roh, 10, 64); err == nil {
		return Zahl(float64(int32(ganz))), nil //#nosec G115 -- die Verengung auf C-`int` ist der Zweck
	}

	if zahl, lesbar := gelesenerZahlwert(roh); lesbar {
		return Zahl(zahl), nil
	}

	if roh == "" {
		return Zahl(0), nil
	}

	return Fehlerwert("to_number bekam " + strconv.Quote(roh) +
		" — das ist keine Zahl"), nil
}

// gelesenerZahlwert liest eine Zahl und sagt, ob es eine war.
//
// **Ein eigener Name statt eines `err`**, weil das Ergebnis hier kein Fehler
// ist, sondern ein **Wert**: `to_int` gibt einen verzögerten Fehler zurück, und
// ein `if err != nil { return …, nil }` sähe aus wie ein verschluckter Fehler.
func gelesenerZahlwert(roh string) (float64, bool) {
	zahl, err := strconv.ParseFloat(roh, 64)

	return zahl, err == nil
}

// kanalsumme gibt Rot plus Grün plus Blau einer Farbe — **ohne** die Deckung.
// `to_int`, `to_real` und `to_number` teilen sie durch drei:
//
//	} else if (t == SCRIPT_COLOR) {
//	  Color c = input->toColor();
//	  result = (c.Red() + c.Blue() + c.Green()) / 3;
//
// `src/script/functions/basic.cpp`, Zeilen 113–115, 142–144 und 167–169,
// Fassung f9d9356 vom 2021-07-06. **Bis F02 fehlte der Zweig**
// (#2244):
// `magic-nouveau` fragt `to_int(card.border_color) < 96`, und die
// Copyright-Zeile bekam einen Fehler statt einer Farbe. Am Orakel gefragt:
// `to_int(rgba(10,20,30,0))` ist 20.
func kanalsumme(wert Wert) (int, bool) {
	farbe, ist := wert.AlsFarbe()
	if !ist {
		return 0, false
	}

	return int(farbe.R) + int(farbe.G) + int(farbe.B), true
}
