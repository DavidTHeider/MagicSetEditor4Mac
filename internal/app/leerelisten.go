package app

import "encoding/json"

// **Eine leere Liste bleibt eine leere Liste** —
// #1288.
//
// # Warum diese Datei
//
// **`nil` und `[]` verhalten sich in Go gleich und in JSON nicht.** Ein nicht
// gesetztes Slice geht als `null` hinaus; der erzeugte TypeScript-Typ sagt
// `Array<T>`, **nicht** optional, und die Seite ruft `antwort.befunde.map(…)`.
// Der Wächter über `omitempty`
// (`internal/werkzeuge/leerelisten_test.go`) sieht das nicht — er liest
// Marken, und hier fehlt keine Marke, sondern ein Wert.
//
// **Am 2026-08-29 an der laufenden Brücke gemessen**, danach an allen 142
// Bindungen: **132 Fundstellen in 112 Bindungen**. 112 davon lagen an
// `Fehler.Werte`; die übrigen sind Antwortfelder, die eine **Eingabe**
// zurückgeben — und eine Eingabe, die niemand gesetzt hat, ist `nil`.
//
// Was hier steht, sind die beiden Antworten darauf: eine an der Grenze
// ([Fehler.MarshalJSON]) und eine an der Stelle, an der eine Liste entsteht
// ([nichtNil]).

// nichtNil gibt eine leere Liste statt einer fehlenden.
//
// **Warum keine Methode je Antworttyp.** Weil diese Felder je zu einem
// eigenen Verbund gehören: eine `MarshalJSON` je Verbund wäre dieselbe Zeile
// fünfmal, und sie stünde weit weg von der Stelle, an der die Liste entsteht.
// Hier steht sie dort, wo jemand sie liest.
//
// **Was sie nicht kann**: dafür sorgen, dass jemand sie ruft. Das tut
// `TestKeineBindungGibtEineListeAlsNull`.
func nichtNil[T any](liste []T) []T {
	if liste == nil {
		return []T{}
	}

	return liste
}

// MarshalJSON schreibt eine fehlende Werteliste als `{}` und nicht als `null`.
//
// # Warum an der Grenze und nicht in jedem Konstruktor
//
// **Weil es sieben Konstruktoren und sieben Verbundliterale sind**, und jeder
// neue Fehler wäre ein achter. Dieselbe Überlegung wie eine Zeile weiter oben
// im Kommentar zu `Uebersetze`: *die Grenze normalisiert, nicht der Konstruktor
// allein.* Hier führt kein Weg vorbei — jede Antwort, die dieses Feld ins
// Fenster trägt, geht durch diese Methode.
//
// **Was sie nicht tut**: die Tabelle füllen. Ein Fehler ohne Platzhalter hat
// keine Werte, und `{}` ist die richtige Antwort darauf.
func (f Fehler) MarshalJSON() ([]byte, error) {
	// **Ein eigener Typ, sonst riefe `json.Marshal` diese Methode wieder.**
	type ohneMethode Fehler

	nackt := ohneMethode(f)
	if nackt.Werte == nil {
		nackt.Werte = map[string]string{}
	}

	// **Ohne eigenen Fehlerzweig.** `ohneMethode` trägt nur Zeichenketten und
	// eine Tabelle davon; `json.Marshal` kann daran nicht scheitern. Ein
	// `if err != nil` wäre eine Zeile, die kein Test je betreten kann — und
	// eine unbetretbare Zeile ist keine Absicherung, sondern eine Lücke in der
	// Abdeckung mit einer Ausrede davor (§3).
	return json.Marshal(nackt) //nolint:wrapcheck // ein Fehler ist hier nicht erreichbar
}
