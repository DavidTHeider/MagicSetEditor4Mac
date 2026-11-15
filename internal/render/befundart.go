package render

import "strings"

// ArtAus ordnet eine Meldung ihrer [Befundart] zu.
//
// # Warum die Zuordnung an der Meldung hängt und nicht an der Stelle
//
// **Weil die Stelle nichts darüber sagt.** Dieselbe Zeile in `textzeichnen.go`
// meldet „die Schrift X ist nicht zu finden" — das hängt daran, was auf diesem
// Rechner installiert ist — und „in der Schrift X fehlen Zeichen", was daran
// hängt, was in der Datei steht. Beides ist [ArtUmgebung], und beides entsteht
// dort, wo auch echte Portierungsmängel entstehen.
//
// **Der Preis gehört ausgesprochen**: eine Meldung, die jemand umformuliert,
// fällt hier heraus und zählt wieder als Portierung. Das ist die richtige
// Richtung — sie fällt in die **strengere** Art, nicht in die mildere — und
// `TestJedeArtHatIhreMeldung` hält die Zuordnung fest.
//
// **Was hier ausdrücklich nicht geschieht**: eine Meldung wird nicht
// unterdrückt. Sie steht vollständig da und wird nur getrennt gezählt
// (#513).
func ArtAus(meldung string) Befundart {
	switch {
	// **Die Schriftdatei liegt nicht da, oder ihr fehlt ein Zeichen.** Auf
	// einem Rechner mit den Paketschriften steht die Zahl anders; über die
	// 23.168 kanonischen Karten waren es 159 von 174 Befunden
	// (#446).
	case strings.Contains(meldung, "fehlen Zeichen"),
		strings.Contains(meldung, "ist nicht zu finden"):
		return ArtUmgebung

	// **Die Vorlage rechnet einen Pfad aus, den es nicht gibt.** Am Orakel
	// belegt: `"a" + true` gibt auch bei MSE `"atrue"`, und MSE findet dort
	// ebenso wenig ein Bild — es schweigt nur dazu
	// (#497).
	case strings.Contains(meldung, "ist in keinem Paket zu finden"),

		// **Die Vorlage hat ein Zeichen in einen Symbolblock gesetzt**, das
		// dort nichts zu suchen hat — ein Zeilenumbruch etwa. MSE überspringt
		// es wortlos; hier steht es, ohne das Bild zu ändern
		// (#517).
		strings.Contains(meldung, "es wird übersprungen"):
		return ArtVorlage
	}

	return ArtPortierung
}

// String gibt die Art als Text.
func (a Befundart) String() string { return string(a) }
