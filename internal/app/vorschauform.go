// **Die Gestalten, in denen eine Vorschau ankommt.**
//
// Eine eigene Datei, weil es eine eigene Frage ist: `karteneditor.go` sagt,
// **wie** gezeichnet wird — hier steht, **was zurückkommt**. Der Schnitt folgt
// derselben Überlegung wie bei `nennung.go` (§2).

package app

// FeldflaecheDTO ist das Rechteck einer Ebene, samt dem Feld, aus dem sie kam.
//
// **In Bildpunkten und nicht in Kartenpunkten.** Das Frontend legt die Flächen
// über das PNG; käme die Maßzahl aus dem Plan, müsste es die Vergrößerung
// selbst nachrechnen — und läge um genau den Faktor daneben, den es vergessen
// hat.
type FeldflaecheDTO struct {
	Feld   string  `json:"feld"`
	Links  float64 `json:"links"`
	Oben   float64 `json:"oben"`
	Breite float64 `json:"breite"`
	Hoehe  float64 `json:"hoehe"`
}

// VorschauAntwort trägt ein Bild oder einen Fehler.
type VorschauAntwort struct {
	Vorschau *VorschauDTO `json:"vorschau"`
	Fehler   *Fehler      `json:"fehler"`
}
