package domain

// Die Kennzahlen eines Sets.
//
// **Zahlen, keine Karten.** Wer eine Statistik ansieht, will nicht warten, bis
// 3.000 Karten geladen sind — und wer sie im Frontend ausrechnete, lüde genau
// die.

// FieldCount sagt, wie viele Karten ein Feld belegt haben.
type FieldCount struct {
	Field string
	Cards int
}

// ValueCount sagt, wie viele Karten denselben Wert tragen.
type ValueCount struct {
	Value string
	Cards int

	// CardIDs sind die Karten dieses Werts, aufsteigend — UXR-F6
	// (#2013).
	//
	// **Eine Karte trägt ein Feld einmal**, also stehen hier genau `Cards`
	// Kennungen. Sie gehören zur Verteilung und nicht zu einer eigenen
	// Abfrage: der Balken und die Liste dahinter kommen aus einem Lauf, und
	// zwei Läufe könnten verschiedene Bestände sehen (§7, Regel 2).
	CardIDs []int64
}

// SetStats sind die Kennzahlen eines Sets.
//
// **Ohne die Belegung.** Sie kommt aus einer eigenen Abfrage und deshalb aus
// einer eigenen Methode: eine Methode mit zwei Abfragen hat Fehlerpfade, die
// nur über die erste zu erreichen sind — und ein Zweig, den kein Test betreten
// kann, ist eine Behauptung darüber, was im Ernstfall geschieht.
type SetStats struct {
	Cards int

	// Fields ist die Zahl **verschiedener** Feldnamen im Set.
	Fields int

	// Empty ist die Zahl der Karten ohne einen einzigen nichtleeren Wert.
	//
	// **Sie steht hier, weil sie die unangenehme Zahl ist.** Ein Set mit 300
	// Karten, von denen 40 leer sind, sieht in jeder anderen Zahl gesund aus.
	Empty int
}
