package domain

// Fensterzahlen sind die Zahlen zu einem Ausschnitt der Sammlung —
// Werkstück #1500
// und U15 (#2160).
//
// **Ein Verbund und keine drei Rückgabewerte**: die Zahlen reisen gemeinsam
// durch Dienst, Fenster, REST, MCP und Kommandozeile, und wer eine dazulegt,
// soll nicht jede Signatur auf dem Weg anfassen.
type Fensterzahlen struct {
	// Gesamt ist, wie viele Zeilen die **Sicht** hat.
	Gesamt int

	// Insgesamt ist die Zahl der Exemplarzeilen **ohne** Sicht.
	Insgesamt int

	// Stueck ist die Summe der Mengen **ohne** Sicht — eine Zeile mit vier
	// Stück zählt vier. Dieselben Zeilen wie `Insgesamt`, also mit Wünschen,
	// so wie die Startseite sie bis U15 aus der ganzen Liste summierte.
	Stueck int
}
