package service

// **Was ankommt, bevor es gilt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum eine Nachricht nicht gleich wirkt
//
// **Weil eine Leitung wiederholt.** Eine zweite Zustellung ist der Normalfall;
// sie darf nur nicht zweimal wirken. Das Issue nennt den Schaden beim Namen:
// „erzeugen keinen doppelten Draw" — und ein doppelter Draw ist genau das, was
// eine zweimal angewandte Nachricht anrichtet.
//
// # Und warum ein Urteil und kein Fehler
//
// **Weil „schon da" kein Fehler ist.** Ein Fehler verlangt eine Reaktion; eine
// zweite Zustellung verlangt, dass nichts geschieht. Wer beides gleich
// behandelte, füllte das Protokoll mit Meldungen über den Normalbetrieb — und
// darin ginge die Lücke unter, die wirklich eine ist (§8).

import (
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Empfangsausgang ist, was aus einer Zustellung wird.
type Empfangsausgang struct {
	Nachricht domain.Partienachricht

	// Urteil sagt, ob sie gilt.
	//
	// **Nur bei [domain.FolgeNeu] wirkt sie.** Alles andere ist eine
	// Auskunft: schon da, Lücke, alte Epoche.
	Urteil domain.Folgeurteil
}

// Partieempfang nimmt Nachrichten entgegen und hält den Stand.
type Partieempfang struct {
	stand *domain.Empfaengerstand

	// kette ist das zuletzt angenommene Glied — der Beleg für die
	// Wiederaufnahme.
	kette string
}

// NeuerPartieempfang fängt in einer Epoche an.
func NeuerPartieempfang(epoche uint64) *Partieempfang {
	return &Partieempfang{stand: domain.NeuerEmpfaengerstand(epoche)}
}

// Nimm packt eine Zustellung aus und urteilt über sie.
func (e *Partieempfang) Nimm(roh []byte) (Empfangsausgang, error) {
	var nachricht domain.Partienachricht

	if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
		return Empfangsausgang{}, fmt.Errorf(
			"%w: die Zustellung ist keine Partienachricht: %s",
			domain.ErrValidation, fehler.Error())
	}

	if fehler := nachricht.Pruefe(); fehler != nil {
		return Empfangsausgang{}, fehler
	}

	// **Flüchtiges geht an der Reihenfolge vorbei.** Es hat keine Folge — und
	// wer es durch den Stand schickte, bekäme bei jedem Mauszeiger ein
	// „schon da".
	if !nachricht.Art.WirdAufbewahrt() {
		return Empfangsausgang{Nachricht: nachricht, Urteil: domain.FolgeNeu}, nil
	}

	urteil := e.stand.Nimm(nachricht.Epoche, nachricht.Folge)
	if urteil != domain.FolgeNeu {
		return Empfangsausgang{Nachricht: nachricht, Urteil: urteil}, nil
	}

	if nachricht.Art.GehtInDieKette() {
		e.kette = nachricht.Kette
	}

	return Empfangsausgang{Nachricht: nachricht, Urteil: domain.FolgeNeu}, nil
}

// Wunsch gibt den Stand als Wiederaufnahmewunsch.
//
// **Damit fragt ein Gast nach seiner Rückkehr**, und `domain.Wiederaufnahme`
// entscheidet, was er bekommt.
func (e *Partieempfang) Wunsch(partie string) domain.Wiederaufnahmewunsch {
	return domain.Wiederaufnahmewunsch{
		Partie: partie,
		Epoche: e.stand.Epoche(),
		Folge:  e.stand.Folge(),
		Kette:  e.kette,
	}
}
