package domain

// **Ob gespielt wird** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Aus `partie.go` herausgelöst am 2026-09-17**, als sie über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, **was**
// eine Partie ist, hier, **in welchem Zustand** sie sein kann.
//
// **Und der Stand sagt nicht, warum sie beendet ist** — das steht im
// [Partieausgang] (M17,
// #1618).

type Partiestand string

const (
	// StandLaeuft heißt: Befehle werden angenommen.
	StandLaeuft Partiestand = "laeuft"

	// StandPausiert heißt: der Wirt ist fort, und es geht weiter, wenn er
	// wiederkommt.
	//
	// **Kein automatischer Wirtswechsel** (#1458, Abgrenzung): ein dauerhafter
	// Verlust endet in einem ausdrücklichen Abbruch und nicht in einem
	// erfundenen Ergebnis.
	StandPausiert Partiestand = "pausiert"

	// StandBeendet heißt: nichts kommt mehr.
	StandBeendet Partiestand = "beendet"
)

// Partiestaende gibt alle drei — für Prüfungen und Flächen.
func Partiestaende() []Partiestand {
	return []Partiestand{StandLaeuft, StandPausiert, StandBeendet}
}
