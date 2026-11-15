package render

import "slices"

// verkleinertGesetzt sucht den Maßstab, in dem ein zu langer Text gesetzt
// wird — **`scale down to:`, ein Text, der nicht passt, wird kleiner gesetzt**
// (#498).
// Der zweite Rückgabewert sagt, ob überhaupt gesucht wurde.
//
// **Wie MSE2 sucht** (`prepareLinesTryScales`, `src/render/text/viewer.cpp`,
// Zeilen 444–543, Fassung f9d9356 vom 2021-07-06): beim Maßstab 1 beginnen;
// passt es nicht, aus der Unterkante eine untere Schranke schätzen
// (`bound_on_min_scale`), dort setzen und dann halbieren, bis die Schranken
// weniger als einen Schritt auseinanderliegen. Ein Schritt ist
// `1/(4·Schriftgröße)`, mindestens 0,01 (`scaleStep`, `src/render/text/font.cpp`,
// Zeilen 54–56; `CompoundTextElement::scaleStep` nimmt den kleinsten, und
// ein Symbol hat dieselbe Schriftgröße oder eine eigene). **Gefunden wird
// damit nicht der größte passende Maßstab**, sondern einer höchstens einen
// Schritt darunter — so steht der Text bei MSE2 da
// (#2244).
//
// **Ohne Zustand und trotzdem wie MSE2**: MSE beginnt beim Maßstab des
// vorigen Laufs, setzt ihn aber bei jeder Wertänderung auf 1 zurück
// (`TextViewer::reset(false)`, ebd. Zeile 169, aus `onValueChange`). Eine
// Ausgabe zeichnet jede Karte frisch, und eine Vorschau entsteht nach einer
// Änderung — beide beginnen bei 1.
//
// **Es dehnt nicht**: passt der Text bei 1, bleibt er dort ([massstabssuche.finde]).
func verkleinertGesetzt(
	satz Satz, rahmen Rechteck, abschnitte []Abschnitt, schriften Schriftsatz,
	bilder Bildquelle,
) (Satzergebnis, bool) {
	if satz.Kleinstgroesse <= 0 || satz.Kleinstgroesse >= satz.Groesse ||
		rahmen.Hoehe <= 0 {
		return Satzergebnis{}, false
	}

	suche := massstabssuche{
		satz: satz, rahmen: rahmen, abschnitte: abschnitte,
		schriften: schriften, bilder: bilder,
	}

	return suche.finde(), true
}

// massstabssuche hält, was jeder Versuch braucht.
type massstabssuche struct {
	satz       Satz
	rahmen     Rechteck
	abschnitte []Abschnitt
	schriften  Schriftsatz
	bilder     Bildquelle
}

// setze setzt den Zug im Maßstab — Schrift und Symbole zugleich.
func (m massstabssuche) setze(massstab float64) Satzergebnis {
	versuch := m.satz
	versuch.Groesse = m.satz.Groesse * massstab

	return gesetzterSatz(versuch, m.rahmen, symboleMit(m.abschnitte, massstab),
		m.schriften, m.bilder)
}

// passt ist `lines.back().bottom() <= height - padding_bottom` (ebd. Zeile
// 733) — siehe [unterkante].
func (m massstabssuche) passt(gesetzt Satzergebnis) bool {
	return gesetzt.Unterkante <= m.rahmen.Hoehe-m.satz.Polster.Unten
}

// hoechstens ist `bound_on_max_scale` (ebd. Zeilen 429–436): ein Maßstab,
// der passt, lässt so viel Luft, wie seine Unterkante übrig lässt.
func (m massstabssuche) hoechstens(gesetzt Satzergebnis, massstab float64) float64 {
	gesamt := m.rahmen.Hoehe + 1

	hoehe := min(gesamt, gesetzt.Unterkante+m.satz.Polster.Unten)
	if hoehe < 1 {
		return 1
	}

	return massstab * gesamt / hoehe
}

// mindestens ist `bound_on_min_scale` (ebd. Zeilen 437–444).
func (m massstabssuche) mindestens(gesetzt Satzergebnis, massstab float64) float64 {
	hoehe := gesetzt.Unterkante + m.satz.Polster.Unten
	if hoehe < 1 {
		return 0
	}

	return massstab * m.rahmen.Hoehe / hoehe
}

// schritt ist der kleinste `scaleStep` der Elemente, mindestens 0,01.
func (m massstabssuche) schritt() float64 {
	groesste := m.satz.Groesse
	for i := range m.abschnitte {
		if len(m.abschnitte[i].Symbole) > 0 {
			groesste = max(groesste, m.abschnitte[i].Symbolmasse.Groesse)
		}
	}

	return max(0.01, 1/max(groesste*4, 1))
}

// finde ist der Ablauf aus `prepareLinesTryScales`. Die Namen folgen ihm:
// unten `min_scale`, oben `max_scale`, bester `best_scale`.
//
// **Passt der Text bei 1, bleibt es bei 1.** MSE versucht danach einen
// Schritt darüber — aber nur, wenn `1 + Schritt` unter `max_scale` liegt, und
// das ist `min(1 + Schritt, …)` (ebd. Zeilen 474–480). Beginnt die Suche bei
// 1, wird der Versuch nie gemacht; er gilt einem Lauf, der beim Maßstab des
// vorigen beginnt.
func (m massstabssuche) finde() Satzergebnis {
	gesetzt := m.setze(1)
	if m.passt(gesetzt) {
		return gesetzt
	}

	schritt := m.schritt()

	// **Die erste Schätzung gilt, auch wenn sie nicht passt** — „or we force
	// it anyway" (ebd. Zeile 511): MSE prüft sie nicht, bevor es sucht.
	unten := max(m.satz.Kleinstgroesse/m.satz.Groesse, m.mindestens(gesetzt, 1))
	bester := unten
	gesetzt = m.setze(unten)
	oben := min(1, m.hoechstens(gesetzt, unten))

	for unten+schritt < oben {
		mitte := (unten + oben) / 2

		versuch := m.setze(mitte)
		if m.passt(versuch) {
			unten, bester, gesetzt = mitte, mitte, versuch
			oben = min(oben, m.hoechstens(versuch, mitte))

			continue
		}

		oben = mitte
		unten = max(unten, m.mindestens(versuch, mitte))
	}

	if bester != unten {
		gesetzt = m.setze(unten)
	}

	return gesetzt
}

// gesetzterSatz ruft den gewöhnlichen Weg — **ohne** die Verkleinerung, sonst
// riefe sie sich selbst.
func gesetzterSatz(
	satz Satz, rahmen Rechteck, abschnitte []Abschnitt, schriften Schriftsatz,
	bilder Bildquelle,
) Satzergebnis {
	ohne := satz
	ohne.Kleinstgroesse = 0

	return SetzeAbschnitte(ohne, rahmen, abschnitte, schriften, bilder)
}

// symboleMit verkleinert die Symbole eines Zugs um denselben Faktor wie seine
// Schrift:
//
//	font.font->getCharInfo(dc, ctx, font.size * scale, …, out);
//
// `SymbolTextElement::getCharInfo`, `src/render/text/symbol.cpp`, Zeilen
// 22–26, Fassung f9d9356 vom 2021-07-06 — `scale` ist der Faktor, den
// `prepareLinesTryScales` sucht, für Schrift und Symbole derselbe.
//
// **Was hier nicht steht**: `scale down to:` einer Symbolschrift. MSE nimmt
// die engere der Schranken (`CompoundTextElement::minScale`, `compound.cpp`,
// Zeilen 39–45); ohne Angabe ist sie 1 Punkt (`SymbolFontRef`,
// `symbol_font.cpp`, Zeile 555) und bindet nie. Gesucht wird hier nur bis
// zur Schranke der Schrift.
func symboleMit(abschnitte []Abschnitt, faktor float64) []Abschnitt {
	verkleinert := slices.Clone(abschnitte)
	for i := range verkleinert {
		verkleinert[i].Symbolmasse.Groesse *= faktor
	}

	return verkleinert
}
