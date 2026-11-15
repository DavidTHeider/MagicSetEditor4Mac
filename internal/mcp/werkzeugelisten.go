package mcp

// **Die Werkzeuge rund um Listen und die Sammlung** — herausgelöst aus
// `werkzeugliste.go`, als die Decklisteneinfuhr sie über die Zeilengrenze aus
// §2 trug (T8, #1473).
//
// **Der Schnitt folgt der Frage**: dort steht, **welche** Gruppen es gibt und
// in welcher Reihenfolge sie zusammenkommen; hier stehen die vier, die eine
// **Liste** in den Bestand hinein- oder aus ihm herausbringen.

// sammlungslistenwerkzeuge gibt die lesenden und die schreibenden.
func sammlungslistenwerkzeuge(naehte Naehte) (lesend, schreibend []Werkzeug) {
	// **Die Listeneinfuhr** (#1292) — eines lesend, eines schreibend. Ohne
	// sie braucht jedes Exemplar eine UUID von Hand, und eine Sammlung von
	// fünftausend Karten entsteht so nicht.
	einfuhrlesend, einfuhrschreibend := geteilt(listeneinfuhr(naehte.Listeneinfuhr))

	// **Und die Decklisteneinfuhr** (T8, #1473) — eines lesend, eines
	// schreibend. **Neben der Listeneinfuhr und nicht darin**: jene schreibt
	// Besitz, diese Plätze in einem Deck, und das Issue trennt beides
	// ausdrücklich.
	decklistelesend, decklisteschreibend := geteilt(deckeinfuhr(naehte.Deckeinfuhr))

	// **Und die Ausfuhr** (#1292) — eines lesend, eines schreibend. Was
	// hineingeht, muss auch wieder heraus.
	ausfuhrlesend, ausfuhrschreibend := geteilt(sammlungsausfuhr(naehte.Sammlungsausfuhr))

	// **Und die vier Fragen an eine geordnete Sammlung** (#1292, Vermerk
	// 0057) — alle lesend.
	sichten := sammlungssichten(naehte.Sammlungssichten)

	lesend = append(append(append(append(lesend, einfuhrlesend...),
		decklistelesend...), ausfuhrlesend...), sichten...)

	schreibend = append(append(append(schreibend, einfuhrschreibend...),
		decklisteschreibend...), ausfuhrschreibend...)

	return lesend, schreibend
}
