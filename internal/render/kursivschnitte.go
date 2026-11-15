package render

// Kursivschnitte sucht den kursiven und den fett-kursiven Schnitt einer
// Schrift — für Rasterer, SVG und die Messung im Planbauer an **einer**
// Stelle.
//
// # Mit `italic name:` unter diesem Namen
//
// Fett dann über das Gewicht, wie beim aufrechten Schnitt.
//
// # Ohne ihn der Schnitt „Italic" der Familie
//
// MSE reicht dann Familienname und `wxFONTSTYLE_ITALIC` an das
// Betriebssystem (`Font::toWxFont`, `src/data/font.cpp`, Zeilen 83–108,
// Fassung f9d9356 vom 2021-07-06), und das nimmt den kursiven Schnitt der
// Familie. Die Namenstabelle der Suche führt jede Schrift unter Familie und
// Schnitt (`internal/schriften`), also unter „Familie Italic" und „Familie
// Bold Italic". Hier blieb ein solches Feld aufrecht (Befund 3 des
// Zweitblicks über E/F).
//
// **Hat die Familie keinen kursiven Schnitt, bleibt es aufrecht.** Ob wx auf
// dem Mac dann eine Schräglage rechnet, ist **ungeklärt** — die einzige
// Familie im Pack, die das trifft, ist `Matrix` in `magic-scifi`.
func Kursivschnitte(quelle Schriftquelle, name, kursivname string) (kursiv, kursivFett Schrift) {
	if kursivname != "" {
		return gefunden(quelle, kursivname, false), gefunden(quelle, kursivname, true)
	}

	if name == "" {
		return nil, nil
	}

	return gefunden(quelle, name+" Italic", false), gefunden(quelle, name+" Bold Italic", false)
}

// gefunden gibt die Schrift oder nichts — eine fehlende ist Bestenfalls-Arbeit.
func gefunden(quelle Schriftquelle, name string, fett bool) Schrift {
	schrift, err := quelle.Suche(name, fett)
	if err != nil {
		return nil
	}

	return schrift
}
