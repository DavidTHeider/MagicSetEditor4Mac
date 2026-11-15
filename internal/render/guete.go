package render

// **Wie sorgfältig gerastert wird** — Werkstück Ω.4 aus
// #1029.
//
// # Warum es überhaupt vier gibt
//
// Gemessen am 2026-08-25 über zwanzig Karten aus `Black.mse-set`: eine Karte
// zu lesen kostet 72 µs, ihren Plan zu bauen 35,8 ms, sie zu rastern 28,8 ms.
// **Der Plan ist für jede Güte derselbe** — Layout und Zeilenumbruch bleiben
// gleich, das verlangt das Werkstück —, also kann eine Güte nur die 45 Prozent
// anfassen, die das Rastern trägt.
//
// Davon tragen **Bildebenen 78 Prozent** (22,4 ms von 28,8 ms), und dort
// liegt der einzige Hebel, der groß genug ist: die Abtastung. Der nächste
// Nachbar halbiert die Bildzeit auf 11,8 ms.
//
// Flächen (4,6 ms), Text (0,9 ms) und Symbolzüge (0,4 ms) sind zusammen
// kleiner als der Unterschied zwischen zwei Abtastungen. Eine Güte, die dort
// spart, spart nichts.
//
// # Was eine Güte **nicht** ist
//
// **Keine Vergrößerung.** Die beiden gehen getrennt in den Rasterer, und das
// ist eine Entscheidung mit Grund
// ([0044](../../docs/entscheidungen/0049-die-guete-ist-eine-wahl.md)): wer
// eine Vorschau in Druckauflösung will, soll sie bekommen, und wer beim Tippen
// eine große Ansicht offen hat, soll nicht auf jede Taste warten.

// Guete sagt, wie sorgfältig gerastert wird.
//
// **Der Nullwert ist die Vorgabe.** Wer den Parameter vergisst, bekommt das
// Bild, das dieses Programm ohnehin zeichnete — nicht das schnellste und nicht
// das schönste. Eine Güte, deren Vergessen das Ergebnis verschöbe, wäre eine
// Falle.
//
// # Wie weit „Parität" trägt
//
// `GueteMittel` ist die **nachgemessene Filterpolitik** von
// `wxIMAGE_QUALITY_NORMAL`: bilinear beim Verkleinern unter Faktor zwei, sonst
// das Flächenmittel ([abtasten]). Das ist keine pixelgenaue Übereinstimmung
// mit MSE, und der Unterschied steht dort auch beziffert — die bilineare
// Arithmetik liegt noch **2,305** neben wx, und **oberhalb von Faktor zwei ist
// wx' Rechnung ungeklärt**; das Flächenmittel ist dort die beste bekannte
// Wahl, nicht die belegte.
//
// Die Quelle trägt den Aufruf `image.Rescale(w,h)` ohne Qualitätsangabe; alles
// Weitere ist gemessen — und in einem begrenzten Bereich.
type Guete int

const (
	// GueteMittel folgt MSEs Filterpolitik — die gewöhnliche Anzeige und die
	// Vorgabe. **Wie weit das trägt, steht am Typ**: es ist die nachgemessene
	// Filterwahl, keine pixelgenaue Übereinstimmung.
	GueteMittel Guete = 0

	// GueteSchnell ist für die Vorschau beim Tippen.
	GueteSchnell Guete = 1

	// GueteHoch ist für die Bildausfuhr und den Druckbogen.
	GueteHoch Guete = 2

	// GueteHoechst ist für den Druck in voller Auflösung.
	GueteHoechst Guete = 3
)

// Gueten führt alle vier, von der schnellsten zur sorgfältigsten.
//
// **Die Aufzählung ist der Weg der Flächen** (AGENTS.md §16): Fenster,
// Kommandozeile, MCP und REST bieten an, was hier steht. Eine Güte, die hier
// fehlt, ist über keine von ihnen erreichbar — und ein Wächter zählt nach.
var Gueten = []Guete{GueteSchnell, GueteMittel, GueteHoch, GueteHoechst}

// Name gibt den Bezeichner, unter dem eine Güte über die Flächen läuft.
//
// **Englisch**, weil er über MCP und REST geht und dort neben `stylesheet` und
// `card` steht — dieselbe Regel wie für jeden fremden Begriff
// ([0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md)).
func (g Guete) Name() string {
	switch g {
	case GueteSchnell:
		return "fast"
	case GueteHoch:
		return "high"
	case GueteHoechst:
		return "ultra"
	case GueteMittel:
		return "medium"
	}

	return ""
}

// Hinweis sagt, worin diese Güte von MSE abweicht — leer für die Vorgabe.
//
// **An der Ausgabe, nicht in der Dokumentation**: das Werkstück verlangt, dass
// eine Güte ihre Abweichung dorthin schreibt, wo jemand sie sieht.
//
// **Angehängt wird er in `service.guetenbefund`**, nicht hier und nicht im
// Rasterer.
//
// Der Satz gehört dorthin, wo die Güte **gewählt** wird: an die Ausfuhr.
//
// **Ankommen tut er an allen vier Flächen.** Kommandozeile, MCP und REST
// bekommen die Befunde über `app.Bildschreibung`, seit dem 2026-08-25
// (#1137).
//
// **Auch die bessere Abweichung wird angesagt.** Wer eine Karte über MSE
// hinaus glättet, bekommt ein Bild, das MSE so nicht zeichnet — und darf das
// wissen, bevor er es mit einer MSE-Ausgabe vergleicht.
func (g Guete) Hinweis() string {
	switch g {
	case GueteSchnell:
		return "die Güte fast tastet mit dem nächsten Nachbarn ab: " +
			"Kanten treppen, und das Bild ist nicht mit MSE vergleichbar"
	case GueteHoch:
		return "die Güte high mittelt beim Verkleinern über die Fläche " +
			"statt bilinear wie MSE: feine Linien flimmern nicht mehr"
	case GueteHoechst:
		return "die Güte ultra mittelt beim Verkleinern über die Fläche " +
			"und rechnet beim Vergrößern bikubisch — beides weicht von MSE ab"
	case GueteMittel:
		return ""
	}

	return ""
}

// GueteNach gibt die Güte zu einem Namen — und ob es eine gibt.
//
// **Ohne Vorgabe bei einem unbekannten Namen.** Ein Tippfehler, der still die
// gewöhnliche Güte liefert, sähe wie ein Erfolg aus; die Flächen sollen ihn
// melden können.
func GueteNach(name string) (Guete, bool) {
	for _, guete := range Gueten {
		if guete.Name() == name {
			return guete, true
		}
	}

	return GueteMittel, false
}
