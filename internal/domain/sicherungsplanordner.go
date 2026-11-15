package domain

// **Der Ordner des Plans** — OPS01
// (#2347).
// Leer heißt: kein Plan. **Freiwillig**, und erst nach einer Ordnerwahl —
// das Issue nennt es ausdrücklich als Risiko, und eine Vorgabe, die von selbst
// in einen Ordner schriebe, wäre genau das.

// SicherungsplanOrdner ist der Schlüssel der Einstellung.
const SicherungsplanOrdner = "sicherung.planordner"

// BereichSicherung ist das Gebiet „Sicherung" der Einstellungsseite.
//
// **Ein eigener Bereich, weil die Einstellung dort steht, wo gesichert wird**
// — neben „Jetzt sichern" und dem Alter, nicht als Textfeld unter den
// Vorgaben. Wie bei [BereichBildausfuhr] zeichnet ihn nicht die allgemeine
// Liste, sondern sein eigenes Gebiet.
const BereichSicherung = "sicherung"

// sicherungsplaneinstellungen ist die Zeile für die Reihe der Einstellungen.
//
// **Mit Regel und ohne Werteliste** wie die Ordner der Dialoge: absolut oder
// mit `~/` vorn; ob es den Ordner gibt, fragt `service` die Platte beim
// Setzen ([IstOrdnerangabe]) und der Plan bei jedem Lauf — ein Stick, der
// fehlt, ist kein Grund, unter `/Volumes` einen Ordner anzulegen.
func sicherungsplaneinstellungen() []einstellung {
	return []einstellung{{
		Schluessel: SicherungsplanOrdner,
		Wirkung:    WirkungSofort,
		Bereich:    BereichSicherung,
		Pruefung:   PruefeOrdnerangabe,
	}}
}
