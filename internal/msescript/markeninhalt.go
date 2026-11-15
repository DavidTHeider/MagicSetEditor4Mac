package msescript

import "strings"

// **Die Markenhelfer — eine Stelle für zwei Fragen** —
// #723.
//
// Zwei Wege im Paket fragen dasselbe: [schlussende] beim Spalten eines
// Verbundwerts und [ohneMarkeninhalt] beim Bau von HTML. Zwei Rechnungen über
// dieselbe Sache laufen auseinander, sobald eine von ihnen etwas lernt (§10).
//
// Alle vier stammen aus `src/util/tagged_string.cpp`, Fassung f9d9356 vom
// 2021-07-06.

// ohneMarkeninhalt nimmt eine Marke samt ihrem Inhalt heraus.
//
// `to_html` beginnt damit:
//
//	String str = remove_tag_contents(str_in,_("<sep-soft"));
//
// `src/script/functions/export.cpp`, Zeile 209, Fassung f9d9356 vom
// 2021-07-06. Ein weicher Trenner ist die
// Stelle, an der MSE beim Umbrechen eine Zeile teilen **darf**; in einer
// HTML-Ausgabe hat er nichts verloren.
//
// # Warum kein regulärer Ausdruck
//
// Weil er drei von vier gemessenen Fällen anders beantwortete:
//
//	a<sep-soft>X<sep>Y</sep>Z</sep-soft>b   →  "ab"   (Schachtelung zählt mit)
//	a<sep-soft>offen                       →  ""     (alles fällt weg)
//	a<sep>X</sep>b                         →  "aXb"  (bleibt unberührt)
//
// Der zweite ist der teure: fehlt die schließende Marke, gibt MSE zurück, was
// es **bis dahin** gesammelt hat — und das ist nichts, denn das Stück vor der
// Marke wird erst danach angehängt (`tagged_string.cpp`, Zeilen 585–588). Ein
// regulärer Ausdruck ließe den Text stehen.
func ohneMarkeninhalt(text, marke string) string {
	stelle := strings.Index(text, marke)
	if stelle < 0 {
		return text
	}

	var (
		gebaut strings.Builder
		seit   int
	)

	for stelle >= 0 {
		schluss, gefunden := schlussmarke(text, stelle)
		if !gefunden {
			return gebaut.String()
		}

		gebaut.WriteString(text[seit:stelle])
		seit = hinterMarke(text, schluss)

		if seit >= len(text) {
			return gebaut.String()
		}

		if stelle = strings.Index(text[seit:], marke); stelle >= 0 {
			stelle += seit
		}
	}

	gebaut.WriteString(text[seit:])

	return gebaut.String()
}

// schlussende gibt die Stelle **hinter** der schließenden Marke.
//
// Findet sich kein Schluss, gibt der Weg das Ende der Zeichenkette — bei MSE
// über `npos` und das `min` beim Aufrufer, hier unmittelbar.
func schlussende(wert string, anfang int) int {
	stelle, gefunden := schlussmarke(wert, anfang)
	if !gefunden {
		return len(wert)
	}

	return hinterMarke(wert, stelle)
}

// schlussmarke sucht das `<` der passenden schließenden Marke.
//
//	size_t match_close_tag(const String& str, size_t start) {
//	  String tag  = tag_type_at(str, start);
//	  String ctag = _("/") + tag;
//	  int taglevel = 1;
//	  for (size_t pos = start + tag.size() + 2 ; pos < size ; ++pos) { … }
//	}
//
// `tagged_string.cpp`, Zeilen 235–257. **Geschachtelte Marken zählen mit**,
// und `<sep-soft>` gilt als dieselbe Art wie `<sep>`: [markenart] schneidet am
// ersten `>` **oder `-`** ab.
func schlussmarke(wert string, anfang int) (stelle int, gefunden bool) {
	art := markenart(wert, anfang)
	if art == "" {
		return 0, false
	}

	schluss := "/" + art
	tiefe := 1

	for stelle := anfang + len(art) + 2; stelle < len(wert); stelle++ {
		if wert[stelle] != '<' {
			continue
		}

		switch {
		case folgt(wert, stelle+1, schluss):
			tiefe--

			if tiefe == 0 {
				return stelle, true
			}

			stelle += len(schluss) + 1

		case folgt(wert, stelle+1, art):
			tiefe++
			stelle += len(art) + 1
		}
	}

	return 0, false
}

// markenart liest den Namen einer Marke — bis zum ersten `>` oder `-`.
//
//	size_t end = str.find_first_of(_(">-"), pos);
//	return str.substr(pos + 1, end - pos - 1);
//
// `tagged_string.cpp`, Zeilen 310–314. **Der Bindestrich beendet den Namen**:
// `<sep-soft>` heißt hier `sep`, `<i-flavor>` heißt `i`.
func markenart(wert string, anfang int) string {
	for stelle := anfang + 1; stelle < len(wert); stelle++ {
		if wert[stelle] == '>' || wert[stelle] == '-' {
			return wert[anfang+1 : stelle]
		}
	}

	return ""
}

// hinterMarke gibt die Stelle hinter dem `>` einer Marke — `skip_tag`.
func hinterMarke(wert string, anfang int) int {
	for stelle := anfang; stelle < len(wert); stelle++ {
		if wert[stelle] == '>' {
			return stelle + 1
		}
	}

	return len(wert)
}

// folgt sagt, ob an dieser Stelle dieser Text steht.
func folgt(wert string, stelle int, text string) bool {
	return stelle+len(text) <= len(wert) && wert[stelle:stelle+len(text)] == text
}
