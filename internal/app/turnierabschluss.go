package app

// **Abschluss und Protokoll eines Turniers** — DOM-04.AK-04
// (#1699),
// Vermerk 0139. Entschieden wird eine Schicht tiefer (§2); hier wird
// übersetzt.

import "context"

// ProtokollAntwort trägt das Protokoll eines Turniers.
type ProtokollAntwort struct {
	Eintraege []ProtokolleintragDTO `json:"eintraege"`
	Fehler    *Fehler               `json:"fehler"`
}

// ProtokolleintragDTO ist eine Zeile des Protokolls.
//
// **Die Angaben sind sprachfrei** — ein Name, eine Rundennummer, ein
// Spielstand. Den Satz baut die Fläche in ihrer Sprache aus `vorgang`.
type ProtokolleintragDTO struct {
	Zeit    string `json:"zeit"`
	Vorgang string `json:"vorgang"`
	Angaben string `json:"angaben"`

	// PaarungID ist **leer**, wenn der Vorgang keine Paarung betrifft.
	PaarungID string `json:"paarungID"`
}

// TurnierAbschliessen schließt ein Turnier ab — gegen die gelesene Fassung.
func (a *App) TurnierAbschliessen(ctx context.Context, turnierID, fassung string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.abschlussSetzen(ctx, turnierID, fassung, a.turniere.Abschliessen)
}

// TurnierWiedereroeffnen nimmt den Abschluss zurück — gegen die gelesene
// Fassung.
func (a *App) TurnierWiedereroeffnen(ctx context.Context, turnierID, fassung string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.abschlussSetzen(ctx, turnierID, fassung, a.turniere.Wiedereroeffnen)
}

// abschlussSetzen ist der gemeinsame Rumpf beider Richtungen — der Kontext
// trägt das Profil schon.
func (a *App) abschlussSetzen(
	ctx context.Context, turnierID, fassung string,
	weg func(context.Context, int64, int64) error,
) SchreibAntwort {
	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	stand, standbefund := benannteKennungAus("fassung", fassung)
	if standbefund != nil {
		return SchreibAntwort{Fehler: standbefund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, weg(ctx, kennung, stand))}
}

// TurnierProtokoll gibt, was an einem Turnier geschah.
func (a *App) TurnierProtokoll(ctx context.Context, turnierID string) ProtokollAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return ProtokollAntwort{Eintraege: []ProtokolleintragDTO{}, Fehler: befund}
	}

	protokoll, err := a.turniere.Protokoll(ctx, kennung)
	if err != nil {
		return ProtokollAntwort{Eintraege: []ProtokolleintragDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	eintraege := make([]ProtokolleintragDTO, 0, len(protokoll))

	for _, eintrag := range protokoll {
		dto := ProtokolleintragDTO{
			Zeit: eintrag.Zeit, Vorgang: string(eintrag.Vorgang), Angaben: eintrag.Angaben,
		}

		if eintrag.PaarungID != 0 {
			dto.PaarungID = kennungtext(eintrag.PaarungID)
		}

		eintraege = append(eintraege, dto)
	}

	return ProtokollAntwort{Eintraege: eintraege}
}
