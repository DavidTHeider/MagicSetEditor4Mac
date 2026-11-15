package mcp

// **Geschrieben ist geschrieben — an den Werkzeugen** (§7, Regel 3;
// #1960,
// #1970).
//
// **Ein Feld neben dem Stand und kein Werkzeugfehler**: ein Fehler hieße für
// ein Modell „nicht geschehen", und es versuchte es noch einmal — gegen den
// geschriebenen Stand. Ein leerer Stand ohne dieses Feld hielte es dagegen
// für die Wahrheit.

// Nachladefehler ist der Grund, aus dem der Stand nach einer geglückten
// Änderung nicht zu lesen war — dieselbe Gestalt wie im Dienst.
type Nachladefehler struct {
	Code    string `json:"code"`
	Meldung string `json:"meldung"`
}

// Nachgeladen ist **kein Fehler**, sondern ein geglückter Ruf, nach dem nur
// das Nachlesen scheiterte — für Werkzeuge, deren Ergebnis eine **Liste** ist
// und kein Feld daneben hat.
//
// **Als Fehlerwert, weil [Werkzeug.Ruf] nur Text und Fehler kennt.** Der
// Server erkennt ihn und antwortet mit einem geglückten Ergebnis aus zwei
// Teilen: der Liste, wie sie war, und einem Satz dahinter. Die Liste in ein
// Objekt zu packen wäre brechend (0033); ein zweiter Teil ist es nicht.
type Nachgeladen struct {
	// Ergebnis ist der Text, der ohne Nachladefehler stünde.
	Ergebnis string

	Grund Nachladefehler
}

// Error ist der Satz, der als zweiter Teil hinausgeht.
func (n *Nachgeladen) Error() string {
	return "geschrieben — nur das Nachlesen danach scheiterte: " + n.Grund.Meldung
}

// nachgeladen gibt das Ergebnis eines Werkzeugs, dessen Ausgabe eine Liste
// ist — und wo nur das Nachlesen scheiterte, ein [Nachgeladen] statt des
// Textes.
func nachgeladen(text string, grund *Nachladefehler) (string, error) {
	if grund == nil {
		return text, nil
	}

	return "", &Nachgeladen{Ergebnis: text, Grund: *grund}
}
