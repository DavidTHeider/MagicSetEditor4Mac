package domain

// **Die zwei Einstellungen der Ausgabe aus AP-10**
// (#2367) —
// hier und nicht in `einstellungsliste.go`, weil die Liste an der Zeilengrenze
// aus §2 steht. Eingehängt werden sie aus dem `init` in
// `einstellungsreihe.go`, an der Stelle, an der sie in der Reihe stehen.

// bildordnereinstellung ist der zuletzt gewählte Ordner der Bildausgabe.
//
// **Ein gemerkter Ordner und keine Vorgabe**: die Bildausgabe schreibt ihn bei
// jeder Wahl fort, wie MSE2. Dieselbe Regel wie die Ordner der Dialoge —
// absolut oder mit `~/` vorn.
func bildordnereinstellung() einstellung {
	return einstellung{
		Schluessel: EinstellungBildordner,
		Wirkung:    WirkungSofort,
		Bereich:    BereichBildausfuhr,
		Pruefung:   PruefeOrdnerangabe,
	}
}

// ausgabezieleinstellung ist das zuletzt gezeigte Ziel je Set.
//
// **Ohne Werteliste, mit einer Regel**, wie die gespeicherten Filter: eine
// Liste in JSON lässt sich prüfen, aber nicht aufzählen.
func ausgabezieleinstellung() einstellung {
	return einstellung{
		Schluessel: EinstellungAusgabeziele,
		Wirkung:    WirkungSofort,
		Bereich:    BereichArbeitsplatz,
		Pruefung:   PruefeAusgabeziele,
	}
}
