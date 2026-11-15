package app

// **Was der Beitritt dem Fenster sagt und von ihm bekommt** — herausgelöst
// aus `partiebeitritt.go`, als P9c es über die Zeilengrenze aus §2 brachte.
// Die Sendeseite bleibt dort, neben der Empfangsschleife: beide bauen den
// `Partiegast`, und `nahtverdrahtung_test.go` liest sie als **eine** Wurzel.

// PartiebeitrittAntwort sagt, an welchem Tisch dieses Programm sitzt.
type PartiebeitrittAntwort struct {
	Laeuft bool   `json:"laeuft"`
	Partie string `json:"partie"`

	// Gegenueber ist der Fingerabdruck des Wirts — die Gestalt, die ein
	// Mensch vergleichen kann.
	Gegenueber string `json:"gegenueber"`

	// Pausiert heißt: die Leitung riss, ohne dass jemand aufgestanden ist.
	//
	// **Und nicht: die Partie ist vorbei.** Wem der Wirt wegbricht, dem
	// gehört sein Platz weiter; was mit der Partie geschieht, sagt
	// [App.tischVerloren] — sie **pausiert**, und fortsetzen kann sie nur
	// derselbe Wirt. Das Fenster sagt es, statt still das Formular
	// zurückzubringen (§6).
	Pausiert bool `json:"pausiert"`

	// Absage ist der Grund, mit dem der Wirt den Gast nicht setzte (P9c).
	Absage string `json:"absage"`

	Fehler *Fehler `json:"fehler"`
}

// PartiemitbringDTO ist, was ein eingeladener Gast an den Tisch mitbringt
// (P9b, #2377): ein eigenes Deck **oder** eine Leihe dieses Wirts — beides
// leer, wenn der Sitz schon ein Deck hat.
type PartiemitbringDTO struct {
	DeckID  string `json:"deckID"`
	LeiheID string `json:"leiheID"`
}
