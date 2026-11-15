package vcard

// **Die Zeilen einer vCard** — Faltung, Kopf und Maskierung nach RFC 6350
// §3.2 bis §3.4 (nachgelesen am 2026-09-27).

import "strings"

// entfalten macht aus physischen Zeilen logische.
//
// **Eine Zeile, die mit einem Leerraum beginnt, setzt die vorige fort**, und
// genau dieser eine Leerraum fällt weg (§3.2). Neben CRLF wird auch ein
// blankes LF angenommen — Dateien, die über andere Werkzeuge gingen, haben
// es. Leere Zeilen zählen mit, stehen aber nicht in der Liste.
func entfalten(text string) []zeile {
	var zeilen []zeile

	for nummer, roh := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		roh = strings.TrimSuffix(roh, "\r")

		fortsetzung := strings.HasPrefix(roh, " ") || strings.HasPrefix(roh, "\t")
		if fortsetzung && len(zeilen) > 0 {
			zeilen[len(zeilen)-1].text += roh[1:]

			continue
		}

		if strings.TrimSpace(roh) != "" {
			zeilen = append(zeilen, zeile{nummer: nummer + 1, text: roh})
		}
	}

	return zeilen
}

// teilen trennt Kopf und Wert am ersten Doppelpunkt **außerhalb von
// Anführungszeichen** — ein Parameterwert darf einen enthalten.
func teilen(text string) (kopf, wert string, gefunden bool) {
	zitiert := false

	for stelle, zeichen := range text {
		switch {
		case zeichen == '"':
			zitiert = !zitiert
		case zeichen == ':' && !zitiert:
			return text[:stelle], text[stelle+1:], true
		}
	}

	return text, "", false
}

// eigenschaft zerlegt den Kopf in Namen und Parameter.
//
// **Eine Gruppe fällt weg** (`item1.EMAIL`), Namen und Parameternamen sind
// unabhängig von der Schreibung, und ein Parameter ohne `=` ist ein `TYPE`
// wie in vCard 2.1 (`TEL;CELL`). Werte in Anführungszeichen dürfen eine
// Kommaliste tragen.
func eigenschaft(kopf string) (name string, parameter map[string][]string) {
	teile := maskiertTeilen(kopf, ';')

	name = strings.ToUpper(strings.TrimSpace(teile[0]))
	if _, ohneGruppe, gruppiert := strings.Cut(name, "."); gruppiert {
		name = ohneGruppe
	}

	parameter = map[string][]string{}

	for _, eintrag := range teile[1:] {
		schluessel, werte, mitWert := strings.Cut(eintrag, "=")
		if !mitWert {
			schluessel, werte = "TYPE", eintrag
		}

		schluessel = strings.ToUpper(strings.TrimSpace(schluessel))

		for _, einer := range strings.Split(strings.Trim(werte, `"`), ",") {
			if einer = strings.TrimSpace(einer); einer != "" {
				parameter[schluessel] = append(parameter[schluessel], einer)
			}
		}
	}

	return name, parameter
}

// maskiertTeilen teilt an einem Trennzeichen, das nicht maskiert ist.
func maskiertTeilen(text string, trenner rune) []string {
	var (
		teile    []string
		anfang   int
		maskiert bool
	)

	for stelle, zeichen := range text {
		switch {
		case maskiert:
			maskiert = false
		case zeichen == '\\':
			maskiert = true
		case zeichen == trenner:
			teile = append(teile, text[anfang:stelle])
			anfang = stelle + 1
		}
	}

	return append(teile, text[anfang:])
}

// demaskieren löst `\n`, `\,`, `\;` und `\\` auf (§3.4); ein anderer
// Backslash bleibt stehen.
func demaskieren(text string) string {
	if !strings.Contains(text, `\`) {
		return text
	}

	var ergebnis strings.Builder

	maskiert := false

	for _, zeichen := range text {
		switch {
		case maskiert && (zeichen == 'n' || zeichen == 'N'):
			ergebnis.WriteRune('\n')
		case maskiert && (zeichen == ',' || zeichen == ';' || zeichen == '\\'):
			ergebnis.WriteRune(zeichen)
		case maskiert:
			ergebnis.WriteRune('\\')
			ergebnis.WriteRune(zeichen)
		case zeichen == '\\':
			maskiert = true

			continue
		default:
			ergebnis.WriteRune(zeichen)
		}

		maskiert = false
	}

	if maskiert {
		ergebnis.WriteRune('\\')
	}

	return ergebnis.String()
}
