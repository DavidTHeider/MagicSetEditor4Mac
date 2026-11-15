package msepaket

import (
	"strings"
	"unicode"
)

// bom ist die Byte-Order-Mark, mit der MSE-Dateien beginnen.
//
// Belegt in `docs/FORMATE.md` an `magic-m15.mse-style/style`. Wer sie stehen
// lässt, bekommt einen Schlüssel, der auf nichts passt — und sucht den Fehler
// danach im Vergleich statt im Einlesen.
//
// Als Fluchtfolge und nicht als Zeichen: ein echtes BOM mitten in einer
// Go-Datei lehnt der Übersetzer ab ("illegal byte order mark"), und in einem
// Zeichenkettenliteral wäre es zudem unsichtbar.
const bom = "\uFEFF"

// Zeilenart sagt, was eine Zeile war.
//
// Leerzeilen und Kommentare werden **mitgeführt**, nicht verworfen. Der Grund
// ist die Rundlaufzusage aus `docs/entscheidungen/0002`: was gelesen wurde, muss
// unverändert zurückgeschrieben werden können, und eine weggeworfene Leerzeile
// ist eine Änderung, die niemand angeordnet hat.
type Zeilenart int

const (
	// ArtEintrag ist eine Zeile `schlüssel: wert`.
	ArtEintrag Zeilenart = iota
	// ArtLeer ist eine leere Zeile.
	ArtLeer
	// ArtKommentar ist eine Zeile, die mit `#` beginnt.
	ArtKommentar
	// ArtFortsetzung ist eine eingerückte Zeile ohne Doppelpunkt.
	//
	// **Sie gehört zum Wert des Blocks über ihr, nicht zu seinen Einträgen.**
	// Ein `init script:` trägt Skripttext, und der ist kein `schlüssel: wert`
	// (#170).
	ArtFortsetzung
)

// Zerteile liest eine Paketdatei in einen Baum.
//
// `datei` geht nur in die Fehlermeldungen ein; gelesen wird `inhalt`. Diese
// Trennung ist Absicht: so lässt sich der Zerteiler gegen Zeichenketten fahren,
// ohne dass für jeden Fall eine Datei entstehen muss — und die Prüfungen gegen
// **echte** Dateien aus `testdaten/` bleiben trotzdem die Regel (AGENTS.md §3).
func Zerteile(datei string, inhalt []byte) (*Baum, error) {
	baum := &Baum{Datei: datei, MitBOM: strings.HasPrefix(string(inhalt), bom)}

	// Der Stapel der offenen Ebenen. `ebenen[0]` ist die Wurzel; `ebenen[n]`
	// nimmt auf, was mit n Tabulatoren eingerückt ist.
	ebenen := []*[]*Knoten{&baum.Wurzel}

	// **Die Klammerbilanz hält einen mehrzeiligen Ausdruck zusammen.**
	//
	// Ein Skript darf über mehrere Zeilen laufen (`docs/FORMATE.md`), und seine
	// Zeilen sind kein `schlüssel: wert` — manche tragen trotzdem einen
	// Doppelpunkt:
	//
	//	mana_sort := {
	//		if styling.unsorted_casting_cost then input
	//		else sort_text@(order: "XYZI[0123456789]HSCA(WUBRG)")()
	//	}
	//
	// Wer sie zerteilt, bekommt einen Schlüssel `sort_text@(order` und einen
	// Einrückungssprung. Solange eine Klammer offen steht, ist deshalb **jede**
	// Zeile Fortsetzung (#170).
	//
	// **Aber die Einrückung schließt sie wieder** — und das ist MSE2s Regel:
	//
	//	} else if (value.empty()) {
	//	  // a multiline string
	//	  readLine(true);
	//	  while (indent >= expected_indent && !input.Eof()) {
	//	    previous_value += line.substr(expected_indent);
	//
	// `src/util/io/reader.cpp`, Zeilen 258–272, Fassung f9d9356 vom
	// 2021-07-06. **Ein mehrzeiliger Wert endet an der Einrückung und an
	// nichts sonst**; MSE2 zählt dabei keine einzige Klammer.
	//
	// Ohne diese Grenze kostete eine einzige Klammer 1.980 Karten: `Blue`
	// führt `Magical Hacker`, deren Flavortext Leetspeak ist —
	// `1|= y()u (4|\| r3@d 7#][5 …` — und die offene Klammer verschluckte
	// den ganzen Rest des Sets, still und ohne Meldung
	// (#539).
	offeneKlammern := 0

	// klammertiefe ist die Einrückung der Zeile, die die Bilanz eröffnet hat,
	// und vonEintrag, ob sie ein Eintrag war.
	klammertiefe, vonEintrag := 0, false

	if fehler := kodierungPruefen(datei, inhalt); fehler != nil {
		return nil, fehler
	}

	for nummer, zeile := range zerlegen(string(inhalt)) {
		if spalte := ungueltigAb(zeile.text); spalte > 0 {
			return nil, &Zerteilfehler{
				Stelle: Stelle{Datei: datei, Zeile: nummer + 1, Spalte: spalte},
				Meldung: "diese Zeile ist kein gültiges UTF-8 — MSE2 liest dort nicht weiter, " +
					"und dieses Programm rät keine Kodierung",
				Grund: ErrKodierung,
			}
		}

		knoten, tiefe, err := zeileLesen(datei, nummer+1, zeile)
		if err != nil {
			return nil, err
		}

		// **Eine flachere Zeile schließt, was die Klammer offen ließ.**
		//
		// Leerzeilen und Kommentare zählen nicht mit: ihre Einrückung sagt
		// nichts, und MSE2 überspringt sie ausdrücklich
		// (`reader.cpp`, Zeilen 265–269).
		//
		// **Und ein Eintrag derselben Tiefe schließt, was ein Eintrag offen
		// ließ** (#2293):
		// MSE2 liest einen Wert bis zum Zeilenende und einen Block nur aus
		// tieferen Zeilen (`reader.cpp`, Zeilen 207–223 und 258–272). Hinter
		// `match: [^(.,”":;\n—]+` verschwand sonst `optional: false` — 16-mal
		// in `magic.mse-game`, gemessen am 2026-10-11. In einem Skriptblock
		// ändert es nichts: `Skripttext` sammelt jeden Nachfahren.
		gleichtief := vonEintrag && tiefe == klammertiefe && knoten.Zeilenart == ArtEintrag

		if offeneKlammern > 0 && (tiefe < klammertiefe || gleichtief) &&
			knoten.Zeilenart != ArtLeer && knoten.Zeilenart != ArtKommentar {
			offeneKlammern = 0
		}

		if offeneKlammern > 0 {
			knoten.Zeilenart = ArtFortsetzung
			knoten.Schluessel = ""
			knoten.Wert = strings.TrimLeft(zeile.text, "\t")
		}

		// **Die Summe wird geklemmt, nicht die einzelne Zeile.** Eine Zeile, die
		// nur schließt (`\t]`), gibt −1 und gehört auch so gezählt; wer sie
		// schon dort auf 0 höbe, bekäme eine Bilanz, die nie wieder zurückkommt.
		// Unter null darf die **Summe** trotzdem nicht: eine überzählige
		// schließende Klammer höbe sonst eine spätere öffnende auf.
		// **Bei einem Eintrag zählen nur die geschweiften Klammern.**
		//
		// In `magic-genevensis-20-battle.mse-style` steht
		// `choice: left doesn't work :(` — ein Smiley, dessen Klammer nie
		// geschlossen wird. Als offener Ausdruck gelesen verschluckt er die
		// restlichen 224 Zeilen samt `card style:`. Ein Eintrag öffnet einen
		// Ausdruck nur mit `{`; eine **Fortsetzung** darf jede Klammer öffnen,
		// denn sie ist schon Ausdruckstext (`replace@(`).
		vorher := offeneKlammern

		offeneKlammern = max(0, offeneKlammern+
			klammerbilanz(zeile.text, knoten.Zeilenart == ArtEintrag))

		if vorher == 0 && offeneKlammern > 0 {
			klammertiefe, vonEintrag = tiefe, knoten.Zeilenart == ArtEintrag
		}

		// Leerzeilen und Kommentare rühren die Ebenen nicht an.
		//
		// Ihre Einrückung sagt nichts: eine Leerzeile hat keine, und in echten
		// Paketen steht sie **zwischen** zwei eingerückten Einträgen. Sie als
		// Ebene zu nehmen schlösse jeden offenen Block, und alles danach hinge
		// an der falschen Stelle. Sie kommen deshalb in die gerade offene
		// Ebene und öffnen selbst keine.
		if knoten.Zeilenart != ArtEintrag {
			offen := ebenen[len(ebenen)-1]
			*offen = append(*offen, knoten)

			continue
		}

		// **Eine zu tiefe Einrückung ist kein Fehler, sondern überflüssig.**
		//
		// `magic-blends.mse-include/new-blends` rückt seine Einträge um einen
		// Tabulator ein, ohne dass ein Block offen wäre — hinter einem
		// Kommentar, auf der Wurzelebene. Wer das abweist, kann die Datei nicht
		// lesen; wer sie in die tiefste offene Ebene hängt, verliert nichts:
		// die Rohzeile bleibt, und die Byte-Zusage aus #48 hält
		// (#170).
		if tiefe >= len(ebenen) {
			tiefe = len(ebenen) - 1
		}

		ebene := ebenen[tiefe]
		*ebene = append(*ebene, knoten)

		// Was jetzt folgt, kann Kind dieses Knotens sein.
		ebenen = append(ebenen[:tiefe+1], &knoten.Kinder)
	}

	return baum, nil
}

// rohzeile ist eine Zeile samt ihrem Zeilenende.
type rohzeile struct {
	text string
	ende string
}

// zerlegen trennt den Inhalt in Zeilen und merkt sich je Zeile ihr Ende.
//
// Nicht `strings.Split(inhalt, "\n")`: das wirft die Unterscheidung zwischen
// `\n` und `\r\n` weg, und echte Pakete tragen **beide in derselben Datei** —
// belegt in `docs/FORMATE.md`. Wer normalisiert, kann nicht mehr
// zurückschreiben, was er gelesen hat.
//
// **Und ein einzelnes `\r` beendet eine Zeile auch** — so bricht MSE2:
// `read_utf8_line` endet an `\n`, an `\r\n` und an einem `\r`, dem kein
// `\n` folgt (`src/util/io/reader.cpp`, Zeilen 159–186, Fassung f9d9356 vom
// 2021-07-06). Hier brach der Zerteiler nur an `\n`, und eine Datei mit den
// Zeilenenden des alten Mac OS war eine einzige Zeile (FMT-06, `ported`).
// Im vollen Pack steht kein einzelnes `\r`, nachgesehen am 2026-09-23 über
// 807 Textdateien; die Regel gilt für Dateien, die jemand anderswo schrieb.
func zerlegen(inhalt string) []rohzeile {
	inhalt = strings.TrimPrefix(inhalt, bom)

	var zeilen []rohzeile

	for inhalt != "" {
		bruch := strings.IndexAny(inhalt, "\r\n")
		if bruch < 0 {
			// Die letzte Zeile ohne Zeilenende.
			zeilen = append(zeilen, rohzeile{text: inhalt})

			break
		}

		ende := inhalt[bruch : bruch+1]
		if strings.HasPrefix(inhalt[bruch:], "\r\n") {
			ende = "\r\n"
		}

		zeilen = append(zeilen, rohzeile{text: inhalt[:bruch], ende: ende})
		inhalt = inhalt[bruch+len(ende):]
	}

	return zeilen
}

// klammerbilanz zählt, wie viele Klammern eine Zeile offen lässt.
//
// **Runde Klammern zählen mit.** `magic-DayNight.mse-style` bricht einen Aufruf
// über vier Zeilen um:
//
//	card_name_replace :=
//		replace@(
//			match: "…"
//			replace: "…")
//
// Nur `{` und `[` zu zählen ließe die Bilanz auf null, und `match:` würde zum
// Eintrag — zwei Ebenen tiefer als sein Elternteil.
//
// **Zeichenketten zählen nicht mit.** `"XYZI[0123456789]HSCA(WUBRG)"` steht so
// in `magic-m15.mse-style` — wer die Klammern darin zählte, hätte eine Bilanz,
// die nie wieder auf null kommt, und läse den Rest der Datei als Fortsetzung.
func klammerbilanz(zeile string, ohneRunde bool) int {
	bilanz := 0
	inZeichenkette := false

	for stelle, zeichen := range zeile {
		switch {
		case zeichen == '"' && (stelle == 0 || zeile[stelle-1] != '\\'):
			inZeichenkette = !inZeichenkette

		case inZeichenkette:

		// **Ab dem Rautenzeichen zählt nichts mehr.** In
		// `magic-m15-dungeons.mse-style` steht `## for loop doesn't work or
		// images :(` — ein Smiley mit einer öffnenden Klammer. Ohne diese Zeile
		// bleibt die Bilanz für den Rest der Datei offen, und der Stil verliert
		// **alle** seine Feldstile: 471 Zeilen weiter unten steht `card style:`,
		// und es wurde zur Fortsetzung.
		case zeichen == '#':
			return bilanz

		case zeichen == '{' || zeichen == '[':
			bilanz++

		case zeichen == '}' || zeichen == ']':
			bilanz--

		// **Runde Klammern zählen nur in einer Fortsetzung.** Ein Eintrag
		// `choice: left doesn't work :(` trägt einen Smiley und keinen offenen
		// Ausdruck; eine Fortsetzung wie `replace@(` ist schon Ausdruckstext.
		case ohneRunde:

		case zeichen == '(':
			bilanz++

		case zeichen == ')':
			bilanz--
		}
	}

	return bilanz
}

// zeileLesen macht aus einer Zeile einen Knoten und gibt ihre Einrückungstiefe.
func zeileLesen(
	datei string, nummer int, zeile rohzeile,
) (*Knoten, int, error) {
	tiefe := 0
	for tiefe < len(zeile.text) && zeile.text[tiefe] == '\t' {
		tiefe++
	}

	rest := zeile.text[tiefe:]
	stelle := Stelle{Datei: datei, Zeile: nummer, Spalte: tiefe + 1}

	knoten := &Knoten{Stelle: stelle, Roh: zeile.text, Zeilenende: zeile.ende}

	switch {
	case strings.TrimSpace(rest) == "":
		knoten.Zeilenart = ArtLeer
		knoten.Wert = rest

		return knoten, tiefe, nil

	case strings.HasPrefix(rest, "#"):
		knoten.Zeilenart = ArtKommentar
		knoten.Wert = rest

		return knoten, tiefe, nil
	}

	schluessel, wert, getrennt := strings.Cut(rest, ":")
	if !getrennt {
		// **Eine eingerückte Zeile ohne Doppelpunkt ist Fortsetzung.**
		//
		// Von siebzehn Stilen eines bezogenen Packs ließen sich fünf öffnen,
		// bevor diese Regel stand: die anderen zwölf scheiterten an Zeilen wie
		// `	]` — dem Ende einer mehrzeiligen Skriptklammer (#170). Ein
		// Skriptblock ist ein mehrzeiliger Wert, und seine Zeilen sind keine
		// Einträge.
		//
		// **Auf der Wurzelebene gilt dasselbe** —
		// [0014](../../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md).
		// Hier stand bis #172 ein Fehler, mit der Begründung, es gebe dort
		// keinen Block, zu dem etwas fortgesetzt werden könnte.
		//
		// Die Begründung setzt voraus, dass die Datei ein Baum ist, und das
		// tut sie nicht immer: von 555 Textdateien in `M15-Magic-Pack`
		// scheiterten zwölf an dieser Zeile, und **keine davon war eine
		// Hauptdatei**. Es sind Skriptrümpfe (`magic.mse-game/script`, dessen
		// `{` unter `crop_safe :=` auf der Wurzelebene steht), Anweisungen
		// (`unit_tests`), JavaScript und CSS.
		//
		// **Was das kostet, steht im Vermerk**: der Zerteiler kann eine
		// Hauptdatei mit verlorenem Doppelpunkt nicht mehr abweisen.
		knoten.Zeilenart = ArtFortsetzung
		knoten.Wert = rest

		return knoten, tiefe, nil
	}

	knoten.Zeilenart = ArtEintrag

	// **Der Schlüssel beidseitig, der Wert nur links** — und das ist nicht
	// dieselbe Zeile zweimal:
	//
	//	key   = canonical_name_form(trim(key));
	//	value = trim_left(substr(line, pos+1));
	//
	// `Reader::readLine`, `src/util/io/reader.cpp`, Zeilen 217 und 224,
	// Fassung f9d9356 vom 2021-07-06. Zwei verschiedene Aufrufe in zwei
	// aufeinanderfolgenden Zeilen.
	//
	// **Wo das zählt** (#565):
	// über 1.102 Textdateien in `full-magic-pack/data` tragen **1.432** Werte
	// Leerraum am Ende. Der sichtbarste Fall ist `auto_replace` — `replace: •_`
	// meint den Punkt **samt** folgendem Abstand; wer ihn wegschnitte, tippte
	// beim Nutzer einen Aufzählungspunkt ohne Abstand.
	//
	// **`unicode.IsSpace` und nicht `" \t"`**: MSEs `isSpace` nimmt 0x09 bis
	// 0x0D und 0x20, darüber `iswspace` (`src/util/string.hpp`, Zeilen
	// 116–126). Das ist Gos Prädikat, nicht die kurze Liste.
	knoten.Schluessel = strings.TrimSpace(schluessel)
	knoten.Wert = strings.TrimLeftFunc(wert, unicode.IsSpace)

	if knoten.Schluessel == "" {
		return nil, 0, &Zerteilfehler{Stelle: stelle, Meldung: "der Schlüssel ist leer"}
	}

	return knoten, tiefe, nil
}
