package domain

// **Was eine Berichtigung verschöbe** — DOM-04.AK-02
// (#1699).
//
// # Gerechnet über dieselbe Rangliste
//
// **Die Folgen sind der Unterschied zweier Ranglisten** — der heutigen und
// der, die mit dem neuen Ergebnis entstünde. Eine zweite Rechnung daneben
// liefe bei der ersten Änderung an den Tiebreakern auseinander (Vermerk 0062,
// Entscheidung 1: die Rangliste ist abgeleitet).
//
// # Was mit den späteren Runden geschieht
//
// **Nichts.** Sie wurden nach dem alten Stand gepaart und bleiben, wie sie
// sind: ein Turnier ist ein Verlauf, und eine gespielte Runde nachträglich
// neu zu paaren, änderte Begegnungen, die stattgefunden haben. **Dass sie
// nach dem alten Stand gepaart wurden, wird gesagt** — das ist die Auskunft,
// die AK-02 verlangt. Eine eigene Regel dieses Programms; die Paarungen der
// späteren Runden anzutasten, schlägt es nicht vor.

import "fmt"

// Korrekturfolgen sagt, was eine Berichtigung verschöbe.
type Korrekturfolgen struct {
	// Runde ist die Nummer der Runde, zu der die Paarung gehört.
	Runde int

	// SpaetereRunden sind die Runden, die **nach** ihr gelegt wurden — nach
	// einem Stand, den die Berichtigung ändert.
	SpaetereRunden []int

	// Verschoben nennt, wer Platz oder Punkte ändert — **nur** sie.
	Verschoben []Platzwechsel
}

// Platzwechsel ist ein Teilnehmer vor und nach der Berichtigung.
type Platzwechsel struct {
	TeilnehmerID int64
	Name         string

	PlatzVorher, PlatzNachher   int
	PunkteVorher, PunkteNachher int
}

// FolgenDerBerichtigung rechnet, was ein neues Ergebnis an einer Paarung
// verschöbe — `nil` heißt: zurücknehmen.
func FolgenDerBerichtigung(turnier Turnier, paarungID int64, neu *Ergebnis) (Korrekturfolgen, error) {
	folgen := Korrekturfolgen{SpaetereRunden: []int{}, Verschoben: []Platzwechsel{}}

	nachher, runde, gefunden := mitErgebnis(turnier, paarungID, neu)
	if !gefunden {
		return Korrekturfolgen{}, fmt.Errorf(
			"%w: die Paarung %d gehört nicht zu diesem Turnier", ErrNotFound, paarungID)
	}

	folgen.Runde = runde

	for _, spaeter := range turnier.Runden {
		if spaeter.Nummer > runde {
			folgen.SpaetereRunden = append(folgen.SpaetereRunden, spaeter.Nummer)
		}
	}

	vorher := map[int64]Rang{}
	for _, rang := range Rangliste(turnier) {
		vorher[rang.TeilnehmerID] = rang
	}

	for _, rang := range Rangliste(nachher) {
		alt := vorher[rang.TeilnehmerID]

		if alt.Platz == rang.Platz && alt.Matchpunkte == rang.Matchpunkte {
			continue
		}

		folgen.Verschoben = append(folgen.Verschoben, Platzwechsel{
			TeilnehmerID: rang.TeilnehmerID, Name: rang.Name,
			PlatzVorher: alt.Platz, PlatzNachher: rang.Platz,
			PunkteVorher: alt.Matchpunkte, PunkteNachher: rang.Matchpunkte,
		})
	}

	return folgen, nil
}

// mitErgebnis gibt eine Abschrift des Turniers mit dem neuen Ergebnis —
// **eine Abschrift**, denn die Runden des Aufrufers bleiben, wie sie sind.
func mitErgebnis(turnier Turnier, paarungID int64, neu *Ergebnis) (Turnier, int, bool) {
	abschrift := turnier
	abschrift.Runden = make([]Runde, len(turnier.Runden))

	gefunden, runde := false, 0

	for stelle, eine := range turnier.Runden {
		eine.Paarungen = append([]Paarung(nil), eine.Paarungen...)

		for platz := range eine.Paarungen {
			if eine.Paarungen[platz].ID == paarungID {
				eine.Paarungen[platz].Ergebnis = neu
				gefunden, runde = true, eine.Nummer
			}
		}

		abschrift.Runden[stelle] = eine
	}

	return abschrift, runde, gefunden
}
