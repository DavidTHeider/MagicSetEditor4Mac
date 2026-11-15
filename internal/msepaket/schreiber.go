package msepaket

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Schreibe gibt den Baum als Dateiinhalt zurück.
//
// # Die Rundlaufzusage
//
// Was gelesen wurde, kommt **Byte für Byte** wieder heraus
// ([0002](../../docs/entscheidungen/0002-sqlite-als-arbeitsformat.md)). Deshalb
// wird nicht aus Schlüssel und Wert zusammengesetzt, was als Rohzeile dasteht:
// `key:  wert` und `key: wert` sind verschiedene Bytes, gemischte Zeilenenden
// kommen in derselben Datei vor, und die letzte Zeile hat oft gar keins.
//
// **Zusammengesetzt wird nur, was niemand gelesen hat** — ein Knoten ohne
// Rohzeile, wie er entsteht, wenn wir eine Datei erzeugen statt sie
// zurückzuschreiben.
func Schreibe(baum *Baum) []byte {
	var gebaut strings.Builder

	if baum.MitBOM {
		gebaut.WriteString(bom)
	}

	schreibeKnoten(&gebaut, baum.Wurzel, 0)

	return []byte(gebaut.String())
}

// schreibeKnoten gibt eine Ebene aus und steigt in ihre Kinder.
//
// **Tiefe zuerst und in Reihenfolge**: ein Kind gehört unter seine Zeile, und
// wiederholte Schlüssel stehen so, wie sie kamen. Eine Abbildung an dieser
// Stelle hätte beides verloren.
func schreibeKnoten(gebaut *strings.Builder, knoten []*Knoten, tiefe int) {
	for _, eintrag := range knoten {
		gebaut.WriteString(zeileVon(eintrag, tiefe))
		gebaut.WriteString(zeilenendeVon(eintrag))

		schreibeKnoten(gebaut, eintrag.Kinder, tiefe+1)
	}
}

// zeileVon gibt die Zeile eines Knotens.
func zeileVon(knoten *Knoten, tiefe int) string {
	if knoten.Roh != "" {
		return knoten.Roh
	}

	// Eine leere Rohzeile ist entweder eine gelesene Leerzeile — dann ist sie
	// leer und richtig — oder ein selbst gebauter Knoten.
	if knoten.Schluessel == "" {
		return knoten.Wert
	}

	einzug := strings.Repeat("\t", tiefe)

	// **Ein Wert mit Umbrüchen steht unter seinem Schlüssel, nicht dahinter**
	// (#704).
	//
	//	} else if (value.empty()) {
	//	  // a multiline string
	//	  readLine(true);
	//	  while (indent >= expected_indent && !input.Eof()) {
	//	    previous_value += line.substr(expected_indent);
	//
	// `src/util/io/reader.cpp`, Zeilen 258–272, Fassung f9d9356 vom
	// 2021-07-06. Hinter dem Doppelpunkt steht **nichts**, und der Wert folgt
	// eine Ebene tiefer eingerückt.
	//
	// **Was es gekostet hat, ohne diesen Zweig:** eine Karte des kanonischen
	// `Vanguard` trägt ein `notes` über drei Zeilen. Geschrieben stand die
	// zweite Zeile ohne Einrückung — und eine Zeile ohne Einrückung beendet
	// den `card:`-Block. Von 18 Feldwerten kamen nach „sichern, wieder
	// öffnen" **zwei** zurück; die Datei sah aus wie eine Datei und war eine
	// andere.
	//
	// **Und einer mit führendem Leerraum ebenso** (FMT-06.AK-02,
	// #1695).
	// Hinter dem Doppelpunkt liest MSE2 mit `trim_left`
	// (`src/util/io/reader.cpp`, Zeile 223) — und dieser Zerteiler genauso —,
	// der Leerraum ginge beim Wiederlesen verloren. MSE2s Schreiber setzt
	// einen solchen Wert deshalb als Block (`src/util/io/writer.cpp`, Zeile 74,
	// Fassung f9d9356 vom 2021-07-06); im Block schneidet das Lesen nur den
	// Einzug ab.
	//
	// **Und einer, der eine Klammer offen lässt** (C02, #2227, gefunden von
	// `FuzzSymbolzeichnung`). MSE2 liest einen Wert bis zum Zeilenende; dieser
	// Zerteiler zählt Klammern ([klammerbilanz]) und läse jede folgende Zeile
	// derselben Tiefe als Fortsetzung — eine Karte namens `[` verlöre jedes
	// Feld nach ihrem Namen. Im Block schließt die flachere Folgezeile, was
	// offen blieb; dieselbe Bilanz wie beim Lesen, sonst liefen beide
	// auseinander.
	zeile := einzug + knoten.Schluessel + ": " + knoten.Wert

	if strings.Contains(knoten.Wert, "\n") || beginntMitLeerraum(knoten.Wert) || klammerbilanz(zeile, true) > 0 {
		gebaut := einzug + knoten.Schluessel + ":"

		for _, teil := range strings.Split(knoten.Wert, "\n") {
			gebaut += "\n" + einzug + "\t" + teil
		}

		return gebaut
	}

	if knoten.Wert == "" {
		return einzug + knoten.Schluessel + ":"
	}

	return zeile
}

// zeilenendeVon gibt das Zeilenende eines Knotens.
//
// **Ein selbst gebauter Knoten bekommt `\n`.** Ein leeres Zeilenende gibt es
// nur einmal je Datei — an der letzten Zeile, wenn sie keines hatte —, und wer
// einen Baum erzeugt, meint keine Datei ohne Umbrüche.
func zeilenendeVon(knoten *Knoten) string {
	if knoten.Zeilenende == "" && knoten.Roh == "" && knoten.Schluessel != "" {
		return "\n"
	}

	return knoten.Zeilenende
}

// beginntMitLeerraum sagt, ob ein Wert mit dem Leerraum beginnt, den der
// Zerteiler hinter dem Doppelpunkt abschneidet — **dasselbe Prädikat**
// (`unicode.IsSpace`), sonst liefen Schreiben und Lesen auseinander.
func beginntMitLeerraum(wert string) bool {
	zeichen, _ := utf8.DecodeRuneInString(wert)

	return unicode.IsSpace(zeichen)
}
