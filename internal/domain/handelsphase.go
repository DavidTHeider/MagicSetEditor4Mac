package domain

// **Vier Phasen für dreizehn Zustände** — TRADE-02.AK-01
// (#1716).
//
// # Warum die Phase hier steht und nicht im Fenster
//
// **Weil sie eine Lesart des Zustands ist**, und jede Fläche dieselbe
// braucht: das Fenster zeigt sie als Schritt des Assistenten, die
// Kommandozeile nennt sie in einer Zeile. Eine zweite Zuordnung im Fenster
// liefe beim vierzehnten Zustand von dieser weg.
//
// # Was sie nicht ist
//
// **Kein neuer Zustand.** Das Zustandsmodell bleibt, wie es ist (das Issue
// verlangt es ausdrücklich); die Phase fasst nur zusammen, was ein Mensch
// als **einen** Abschnitt erlebt — das Artboard *Handeln* nennt sie „Mit
// wem", „Was tauscht ihr", „Zusagen" und „Verbuchen".

// Handelsphase ist ein Abschnitt des Handels, wie ein Mensch ihn erlebt.
type Handelsphase string

// Die Phasen, in der Reihenfolge des Weges.
const (
	// HandelsphaseGegenueber ist die Wahl des Gegenübers — vor jedem Zustand.
	HandelsphaseGegenueber Handelsphase = "gegenueber"

	// HandelsphaseAngebot heißt: was hin und her geht, steht noch zur Verhandlung.
	HandelsphaseAngebot Handelsphase = "angebot"

	// HandelsphaseZusagen heißt: man ist sich einig und bereitet die Übergabe vor.
	HandelsphaseZusagen Handelsphase = "zusagen"

	// HandelsphaseVerbuchen heißt: übergeben, verbuchen, abgleichen — oder klären.
	HandelsphaseVerbuchen Handelsphase = "verbuchen"

	// HandelsphaseEnde heißt: abgelehnt, abgebrochen oder abgelaufen.
	HandelsphaseEnde Handelsphase = "ende"
)

// DieHandelsphasen ist die Aufzählung in der Reihenfolge des Weges.
var DieHandelsphasen = []Handelsphase{
	HandelsphaseGegenueber, HandelsphaseAngebot, HandelsphaseZusagen, HandelsphaseVerbuchen, HandelsphaseEnde,
}

// Phase gibt den Abschnitt, in dem dieser Zustand liegt.
//
// **Eine Klärung gehört zum Verbuchen** — sie entsteht nur dort
// (`handelsmatrix.go`), und wer sie ansieht, sucht die Übergabe.
// **Ein unbekannter Zustand ist ein Ende**: er erlaubt keinen Schritt, und
// ihn einer tätigen Phase zuzuordnen, zeigte Knöpfe, die nichts tun.
func (z Handelszustand) Phase() Handelsphase {
	switch z {
	case HandelEntwurf, HandelAngeboten, HandelVerhandlung:
		return HandelsphaseAngebot
	case HandelAngenommen, HandelVorbereitet:
		return HandelsphaseZusagen
	case HandelUebergabeOffen, HandelLokalVerbucht, HandelAbgleichOffen,
		HandelBeidseitigVerbucht, HandelKlaerung:
		return HandelsphaseVerbuchen
	default:
		return HandelsphaseEnde
	}
}
