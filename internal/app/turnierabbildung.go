package app

// **Die Abbildung der Turniere auf ihre DTOs** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turnierbindungen.go` unter der Zeilengrenze
// aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, **wie eine Antwort
// entsteht**, hier, **wie eine Entität zur Gestalt wird**.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// alsTurnier macht aus einem Turnier sein DTO.
func alsTurnier(turnier domain.Turnier) TurnierDTO {
	dto := TurnierDTO{
		ID:          kennungtext(turnier.ID),
		Fassung:     kennungtext(turnier.Fassung),
		Name:        turnier.Name,
		Spiel:       turnier.Spiel,
		Notiz:       turnier.Notiz,
		Einzelspiel: turnier.Einzelspiel,
		Teilnehmer:  make([]TeilnehmerDTO, 0, len(turnier.Teilnehmer)),
		Runden:      make([]RundeDTO, 0, len(turnier.Runden)),

		// **Die Empfehlung geht mit dem Turnier** und nicht über einen
		// eigenen Ruf: sie folgt allein aus der Teilnehmerzahl, und ein
		// zweiter Weg dafür wäre eine zweite Gelegenheit, sie zu vergessen.
		EmpfohleneRunden: domain.RundenvorschlagFuer(turnier),
	}

	// **Das Datum des Abschlusses geht mit** — DOM-04.AK-04 (#1699).
	dto.Abgeschlossen = turnier.Abgeschlossen

	if turnier.FormatID != 0 {
		dto.FormatID = kennungtext(turnier.FormatID)
	}

	for _, teilnehmer := range turnier.Teilnehmer {
		dto.Teilnehmer = append(dto.Teilnehmer, alsTeilnehmer(teilnehmer))
	}

	for _, runde := range turnier.Runden {
		dto.Runden = append(dto.Runden, alsRunde(runde))
	}

	return dto
}

// alsTeilnehmer macht aus einem Teilnehmer sein DTO.
//
// **Eine fehlende Deckkennung ist leer und nicht „0".** Die Null wäre eine
// Kennung, und die Oberfläche versuchte, sie aufzuschlagen.
func alsTeilnehmer(teilnehmer domain.Teilnehmer) TeilnehmerDTO {
	dto := TeilnehmerDTO{
		ID:           kennungtext(teilnehmer.ID),
		Name:         teilnehmer.Name,
		Ausgestiegen: teilnehmer.Ausgestiegen,
		Stelle:       teilnehmer.Stelle,
		Deckabzug:    make([]AbzugsplatzDTO, 0, len(teilnehmer.Deckabzug)),
	}

	if teilnehmer.DeckID != 0 {
		dto.DeckID = kennungtext(teilnehmer.DeckID)
	}

	for _, platz := range teilnehmer.Deckabzug {
		dto.Deckabzug = append(dto.Deckabzug, AbzugsplatzDTO{
			ID:          kennungtext(platz.ID),
			Bezeichnung: platz.Bezeichnung,
			Printing:    platz.Printing,
			DruckID:     kennungtext(platz.DruckID),
			Menge:       platz.Menge,
			Bereich:     string(platz.Bereich),
		})
	}

	return dto
}

// alsRunde macht aus einer Runde ihr DTO.
func alsRunde(runde domain.Runde) RundeDTO {
	return RundeDTO{
		ID:        kennungtext(runde.ID),
		Nummer:    runde.Nummer,
		Paarungen: alsPaarungen(runde.Paarungen),
	}
}

// alsPaarungen macht aus Paarungen ihre DTOs.
func alsPaarungen(paarungen []domain.Paarung) []PaarungDTO {
	alle := make([]PaarungDTO, 0, len(paarungen))

	for _, paarung := range paarungen {
		eine := PaarungDTO{
			ID: kennungtext(paarung.ID),
			A:  kennungtext(paarung.A),
		}

		// **Ein Bye trägt kein B** — und nicht die Null. Der Unterschied ist
		// der ganze Punkt: eine „0" sähe wie ein Gegner aus.
		if !paarung.IstBye() {
			eine.B = kennungtext(paarung.B)
		}

		if paarung.Ergebnis != nil {
			eine.Ergebnis = &ErgebnisDTO{
				SpieleA:       paarung.Ergebnis.SpieleA,
				SpieleB:       paarung.Ergebnis.SpieleB,
				Unentschieden: paarung.Ergebnis.Unentschieden,
			}
		}

		alle = append(alle, eine)
	}

	return alle
}

// alsRang macht aus einem Rang sein DTO — **mit gerundeten Quoten**.
func alsRang(rang domain.Rang) RangDTO {
	return RangDTO{
		TeilnehmerID:  kennungtext(rang.TeilnehmerID),
		Name:          rang.Name,
		Platz:         rang.Platz,
		Siege:         rang.Bilanz.Siege,
		Niederlagen:   rang.Bilanz.Niederlagen,
		Unentschieden: rang.Bilanz.Unentschieden,
		Byes:          rang.Bilanz.Byes,
		Matchpunkte:   rang.Matchpunkte,

		Matchquote:       gerundet(rang.Matchquote),
		Gegnermatchquote: gerundet(rang.Gegnermatchquote),
		Spielquote:       gerundet(rang.Spielquote),
		Gegnerspielquote: gerundet(rang.Gegnerspielquote),

		Ausgestiegen: rang.Ausgestiegen,
	}
}
