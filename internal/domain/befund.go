package domain

// Befund ist eine Stelle, an der ein Import etwas **nicht deuten** konnte.
//
// **Befunde und Erfolg schließen einander nicht aus** (AGENTS.md §8): ein Set
// mit 300 lesbaren und zwei unlesbaren Karten wird mit 300 Karten und zwei
// benannten Befunden geöffnet, nicht abgewiesen.
//
// **In der Domäne, weil zwei Leser ihn brauchen.** `internal/msepaket` liest
// zeilenweise und hat eine eigene, feinere Stelle mit Zeile und Spalte;
// `internal/mtgjson` liest JSON, wo eine Zeilennummer nichts bedeutet. Der
// gemeinsame Typ steht deshalb dort, wo beide hinreichen — und er beschreibt
// die Stelle als **Text**, statt einem der beiden Formate die Form des anderen
// aufzuzwingen.
type Befund struct {
	// Datei ist die Datei, in der es aufgefallen ist.
	Datei string

	// Ort ist die Stelle innerhalb der Datei, in der Ausdrucksweise ihres
	// Formats: eine Zeilennummer bei zeilenweisen Paketen, ein Pfad bei JSON.
	// Leer ist erlaubt und heißt: die ganze Datei.
	Ort string

	Meldung string

	// Art sagt, wem der Befund gehört — siehe [Befundart]. Der Nullwert ist
	// [ArtPortierung].
	Art Befundart
}

func (b Befund) String() string {
	switch {
	case b.Datei == "" && b.Ort == "":
		return b.Meldung

	// **Ein Ort ohne Datei ist keiner zum Wegwerfen.** Beim Schreiben gibt es
	// noch keine Datei — der Befund entsteht, bevor irgendetwas abgelegt ist —,
	// und der erste Anlauf ließ dabei genau die Angabe fallen, die sagt,
	// **welches Feld** nicht abbildbar war.
	case b.Datei == "":
		return b.Ort + ": " + b.Meldung

	case b.Ort == "":
		return b.Datei + ": " + b.Meldung

	default:
		return b.Datei + ":" + b.Ort + ": " + b.Meldung
	}
}
