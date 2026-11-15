package msescript

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

// **`to_html` — ein Kartentext als HTML** —
// #723.
//
// # Wer es ruft
//
// Zwei der neunzehn Exportvorlagen von `M15-Magic-Pack`, und zwar durchgehend:
// `magic-planesculptors` 14-mal, `magic-spoiler` 17-mal. Keine der beiden
// definiert es selbst, und ohne es läuft keine von beiden.
//
// # Wo es steht
//
// In der Bibliothek und nicht im Ausfuhrlauf. Bei MSE gehört es zwar zu den
// „export functions", aber `init_script_export_functions` wird aus
// `init_script_functions` gerufen (`src/script/functions/functions.hpp`,
// Zeilen 31–41, Fassung f9d9356 vom 2021-07-06) — es steht in **jedem**
// Kontext. Das Orakel hat es in einer gewöhnlichen Sitzung ausgewertet.
//
// # Die Naht zu den Symbolbildern
//
// `to_html` mit einer Symbolschrift schreibt für jeden `<sym>`-Block Bilder
// neben die Ausgabe — das kann nur der Ausfuhrlauf, denn nur er kennt einen
// Ort dafür. Statt einer zweiten Umsetzung in `service` (§10, „zwei
// Wahrheiten") reicht diese hier einen **Rückruf** durch: die Bibliothek
// trägt keinen bei, der Ausfuhrlauf setzt `to_html` mit seinem.
func init() {
	eintragen("to_html", HTMLFunktion(nil))
}

// Symbolwandler macht aus einem Symbolzug die Marken seiner Bilder.
//
// `zug` ist der rohe Inhalt eines `<sym>`-Blocks — `"2WW"` —, `schrift` der
// Name der Symbolschrift, `groesse` ihre Schriftgröße.
type Symbolwandler func(ctx context.Context, zug, schrift string, groesse float64) (string, error)

// HTMLVorgabegroesse ist die Schriftgröße, wenn keine genannt ist.
//
//	if (symbol_font_size <= 0) symbol_font_size = 12; // a default
//
// `src/script/functions/export.cpp`, Zeile 282, Fassung f9d9356 vom
// 2021-07-06.
const HTMLVorgabegroesse = 12

// ErrKeinSymbolwandler steht für MSEs `guard_export_info`.
//
//	void guard_export_info(const String& function_name) {
//	  if (!export_info()) throw ScriptError(…);
//	}
//
// `src/script/functions/export.cpp`, Zeilen 30–34, Fassung f9d9356 vom
// 2021-07-06: außerhalb einer Ausfuhr gibt es keinen Ort für die Bilder, und
// MSE bricht ab, statt sie wegzulassen.
var ErrKeinSymbolwandler = errors.New(
	"symbols_to_html steht außerhalb einer Ausfuhr nicht zur Verfügung; " +
		"ohne `symbol_font:` bleibt der Zug Text")

// HTMLFunktion baut `to_html` über einem Wandler für Symbolzüge.
//
// **Nil heißt: es gibt keinen Ort für Bilder.** Dann gilt MSEs Weg für
// `symbol_font == nullptr` — der Inhalt eines `<sym>`-Blocks steht als Text im
// `<span>`. Ist dagegen eine Symbolschrift genannt und fällt wirklich ein
// Symbolzug an, ist das ein Fehler; MSE bricht an derselben Stelle ab
// (`guard_export_info`, Zeile 184).
func HTMLFunktion(wandler Symbolwandler) Eingebaut {
	return func(ctx context.Context, ruf Ruf) (Wert, error) {
		if err := ruf.eineEingabe("to_html"); err != nil {
			return Nichts, err
		}

		schrift := ruf.ausRufOderUmgebung("symbol_font").AlsText()

		// **Ein unlesbarer Wert ist dasselbe wie keiner.** MSEs
		// `SCRIPT_OPTIONAL_PARAM_` fängt die fehlende Variable ab und
		// vergleicht danach gegen null; eine Angabe, die keine Zahl ist,
		// landet auf demselben Weg bei der Vorgabe.
		groesse, unlesbar := ruf.ausRufOderUmgebung("symbol_font_size").AlsZahl()
		if unlesbar != nil || groesse <= 0 {
			groesse = HTMLVorgabegroesse
		}

		gebaut, err := NachHTML(ctx, ruf.Stellung[0].AlsText(), schrift, groesse, wandler)

		// **Die Wache bekommt die Stelle, der Wandlerfehler behält seine.**
		// Ein Fehler aus dem Rückruf nennt schon, woran er lag; ihn hier zu
		// überschreiben verlöre die Ursache.
		if errors.Is(err, ErrKeinSymbolwandler) {
			return Nichts, ruf.fehlerAn(err.Error())
		}

		if err != nil {
			return Nichts, err
		}

		return Text(gebaut), nil
	}
}

// AusRufOderUmgebung liest ein benanntes Argument, sonst die Variable daneben.
//
// **Ausgeführt, weil `symbols_to_html` in `internal/service` steht** und
// dieselben zwei Angaben liest wie `to_html` (#658). Zwei Umsetzungen derselben
// Regel liefen auseinander (§10).
func (r Ruf) AusRufOderUmgebung(name string) Wert { return r.ausRufOderUmgebung(name) }

// ausRufOderUmgebung liest ein benanntes Argument, sonst die Variable daneben.
//
//	#define SCRIPT_OPTIONAL_PARAM_N(Type, str, name) …
//	    Type name = from_script<Type>(ctx.getVariable(str), str)
//
// `src/script/functions/util.hpp`, Zeilen 98–101, Fassung f9d9356 vom
// 2021-07-06: MSE liest aus dem **Kontext**, nicht aus der Argumentliste. Ein
// benanntes Argument bindet dort eine Variable — eine anderswo gebundene tut
// es genauso.
//
// **Die Vorlagen leben davon.** `magic-planesculptors` setzt `symbol_font` und
// `symbol_font_size` in Zeile 27–28 als Variablen und ruft danach
// `to_html(card.name)` ohne benanntes Argument.
func (r Ruf) ausRufOderUmgebung(name string) Wert {
	if wert, benannt := r.Benannter(name); benannt {
		return wert
	}

	if r.Umgebung != nil {
		if wert, da := r.Umgebung.Suche(name); da {
			return wert
		}
	}

	return Nichts
}

// NachHTML wandelt einen ausgezeichneten Kartentext in HTML.
//
//	String to_html(const String& str_in, const SymbolFontP& symbol_font, double symbol_size)
//
// `src/script/functions/export.cpp`, Zeilen 208–269, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der Vergleich ist eine Vorsilbe, kein Markenname.** `is_substr(str, i,
// "i")` trifft alles, was mit `i` anfängt — `<i-flavor>` wird deshalb zu
// `<i>`. Was gar nicht trifft, fällt weg; sein Inhalt bleibt stehen.
func NachHTML(
	ctx context.Context, text, schrift string, groesse float64, wandler Symbolwandler,
) (string, error) {
	zeichen := []rune(ohneMarkeninhalt(text, "<sep-soft"))

	fett := htmlmarke{auf: "<b>", zu: "</b>"}
	kursiv := htmlmarke{auf: "<i>", zu: "</i>"}
	symbol := htmlmarke{auf: `<span class="symbol">`, zu: "</span>"}

	var (
		gebaut    strings.Builder
		stapel    markenstapel
		symbolzug strings.Builder
	)

	// zugSchreiben gibt einen gesammelten Symbolzug an den Wandler.
	zugSchreiben := func() error {
		if symbolzug.Len() == 0 {
			return nil
		}

		stapel.schreibeVorgemerkte(&gebaut)

		if wandler == nil {
			return ErrKeinSymbolwandler
		}

		marken, err := wandler(ctx, symbolzug.String(), schrift, groesse)
		if err != nil {
			return err
		}

		gebaut.WriteString(marken)
		symbolzug.Reset()

		return nil
	}

	for stelle := 0; stelle < len(zeichen); {
		if zeichen[stelle] != '<' {
			stapel.schreibeVorgemerkte(&gebaut)

			// **Erst schreiben, dann fragen.** `symbol.opened` wird beim
			// Schreiben der Vormerkung hochgezählt; die Abfrage davor sähe
			// das erste Zeichen eines Symbolzugs noch als Text.
			//
			// **Und in den Symbolzug geht das Zeichen roh.**
			//
			//	if (symbol.opened > 0 && symbol_font) {
			//	  symbols += c; // write as symbols instead
			//	} else {
			//	  c = untag_char(c);
			//
			// `export.cpp`, Zeilen 244–247: `untag_char` steht im **anderen**
			// Zweig. Diese Umsetzung hatte es zunächst in beiden — gefunden
			// hat es die Gegenprobe, nicht das Lesen (§10).
			if symbol.offen > 0 && schrift != "" {
				symbolzug.WriteRune(zeichen[stelle])
			} else {
				schreibeZeichen(&gebaut, zeichen[stelle])
			}

			stelle++

			continue
		}

		stelle++

		if err := markeDeuten(zeichen, stelle,
			&fett, &kursiv, &symbol, &stapel, zugSchreiben); err != nil {
			return "", err
		}

		naechste, weiter := markenende(zeichen, stelle-1)
		if !weiter {
			break
		}

		stelle = naechste
	}

	if err := zugSchreiben(); err != nil {
		return "", err
	}

	stapel.schliesseAlles(&gebaut)

	return gebaut.String(), nil
}

// markeDeuten ordnet eine Marke einer der drei HTML-Marken zu.
//
// **Die Reihenfolge ist die des Originals** und sie trägt: `sym` fängt nicht
// mit `i` an, und `/sym` wird erst nach `/i` geprüft — beides trifft sich
// nicht.
func markeDeuten(
	zeichen []rune, stelle int,
	fett, kursiv, symbol *htmlmarke, stapel *markenstapel, zugSchreiben func() error,
) error {
	switch {
	case vorsilbe(zeichen, stelle, "b"):
		stapel.oeffne(fett)

	case vorsilbe(zeichen, stelle, "/b"):
		stapel.schliesse(fett)

	case vorsilbe(zeichen, stelle, "i"):
		stapel.oeffne(kursiv)

	case vorsilbe(zeichen, stelle, "/i"):
		stapel.schliesse(kursiv)

	case vorsilbe(zeichen, stelle, "sym"):
		stapel.oeffne(symbol)

	case vorsilbe(zeichen, stelle, "/sym"):
		if err := zugSchreiben(); err != nil {
			return err
		}

		stapel.schliesse(symbol)
	}

	return nil
}

// schreibeZeichen maskiert ein einzelnes Zeichen.
//
//	if (c == _('<')) ret += _("&lt;");
//	else if (c == _('&')) ret += _("&amp;");
//	else if (c >= 0x80) ret += String(_("&#")) << (int)c << _(';');
//	else if (c == _('\n')) ret += _("<br>\n");
//
// `export.cpp`, Zeilen 247–258, Fassung f9d9356 vom 2021-07-06.
//
// **`>`, `"` und `'` bleiben roh.** Das ist der Unterschied zu `html_escape`,
// das MSE für Attributwerte benutzt — nachgemessen am Orakel: `a>b"c'd` kommt
// unverändert zurück.
func schreibeZeichen(gebaut *strings.Builder, roh rune) {
	zeichen := entschuetztesZeichen(roh)

	switch {
	case zeichen == '<':
		gebaut.WriteString("&lt;")

	case zeichen == '&':
		gebaut.WriteString("&amp;")

	case zeichen >= 0x80:
		gebaut.WriteString("&#" + strconv.Itoa(int(zeichen)) + ";")

	case zeichen == '\n':
		gebaut.WriteString("<br>\n")

	default:
		gebaut.WriteRune(zeichen)
	}
}

// entschuetztesZeichen ist `untag_char`.
//
//	if (c == ESCAPED_LANGLE) return _('<');
//	else if (c == CONNECTION_SPACE) return _(' ');
//
// `src/util/tagged_string.cpp`, Zeilen 14–18, Fassung f9d9356 vom 2021-07-06.
func entschuetztesZeichen(zeichen rune) rune {
	switch zeichen {
	case '\u0001':
		return '<'

	case '\ueb00':
		return ' '

	default:
		return zeichen
	}
}

// vorsilbe ist `is_substr` — steht `gesucht` ab dieser Stelle?
func vorsilbe(zeichen []rune, stelle int, gesucht string) bool {
	for _, eine := range gesucht {
		if stelle >= len(zeichen) || zeichen[stelle] != eine {
			return false
		}

		stelle++
	}

	return true
}

// markenende gibt die Stelle hinter der nächsten schließenden Klammer.
//
//	size_t end = str.find_first_of(_('>'), start);
//	return end == String::npos ? String::npos : end + 1;
//
// `skip_tag`, `src/util/tagged_string.cpp`, Zeilen 229–233, Fassung f9d9356
// vom 2021-07-06.
//
// **Ohne `>` ist die Eingabe zu Ende** — und zwar mitsamt allem, was noch
// dahinter stand. `to_html("x & y < z")` gibt `"x &amp; y "`; das `z` ist weg.
// Gemessen, nicht erschlossen.
func markenende(zeichen []rune, ab int) (stelle int, weiter bool) {
	for stelle := ab; stelle < len(zeichen); stelle++ {
		if zeichen[stelle] == '>' {
			return stelle + 1, true
		}
	}

	return 0, false
}
