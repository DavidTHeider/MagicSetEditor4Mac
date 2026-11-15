package app

// **Ein Turnier nach einer Vorlage** — DOM-04.AK-01 und AK-03
// (#1699).
//
// Was mitkommt und was nicht, entscheidet `service.Turnierleitung.Kopieren`;
// hier wird übersetzt (§2).

import "context"

// TurnierKopieren legt ein Turnier nach einer Vorlage an.
func (a *App) TurnierKopieren(
	ctx context.Context, turnierID, name string, mitTeilnehmenden bool,
) TurnierAntwort {
	ctx = a.imProfil(ctx)

	quelle, befund := benannteKennungAus("turnierID", turnierID)
	if befund != nil {
		return TurnierAntwort{Fehler: befund}
	}

	kopie, err := a.turniere.Kopieren(ctx, quelle, name, mitTeilnehmenden)
	if err != nil {
		return TurnierAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelesen := alsTurnier(kopie)

	return TurnierAntwort{Turnier: &gelesen}
}
