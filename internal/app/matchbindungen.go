package app

import "context"

// **Das Turniermatch am Fenster** — Werkstück G7 von Phase 5.5
// (#1464).

// TurnierMatch gibt die Bindung einer Paarung samt Meldungen.
func (b *Bindungen) TurnierMatch(ctx context.Context, paarungID string) MatchAntwort {
	return b.app(ctx).TurnierMatch(b.ruf(ctx), paarungID)
}

// TurnierMatchBinden sagt, dass diese Partie zu dieser Paarung gehört.
func (b *Bindungen) TurnierMatchBinden(
	ctx context.Context,
	turnierID, paarungID, partieID string, zuordnung []MatchsitzDTO,
) MatchAntwort {
	return b.app(ctx).TurnierMatchBinden(
		b.ruf(ctx), turnierID, paarungID, partieID, zuordnung)
}

// TurnierMatchMelden trägt das Ergebnis **einer** Seite ein.
//
// **Und trägt damit nichts ins Turnier ein.** Was aus zwei Meldungen wird,
// entscheidet `TurnierMatchUebernehmen` — auf Zuruf der Turnierleitung.
func (b *Bindungen) TurnierMatchMelden(
	ctx context.Context,
	paarungID, von, kette string, stand ErgebnisDTO,
) MatchAntwort {
	return b.app(ctx).TurnierMatchMelden(
		b.ruf(ctx), paarungID, von, kette, stand)
}

// TurnierMatchAusDerPartie meldet, was die Partie festgestellt hat.
//
// **Niemand tippt einen Spielstand und niemand schreibt eine Kette ab** —
// beides kommt aus der beendeten Partie (M17, #1618).
func (b *Bindungen) TurnierMatchAusDerPartie(
	ctx context.Context,
	paarungID, von string,
) MatchAntwort {
	return b.app(ctx).TurnierMatchAusDerPartie(b.ruf(ctx), paarungID, von)
}

// TurnierMatchUebernehmen trägt das Ergebnis ins Turnier ein — **einmal**.
func (b *Bindungen) TurnierMatchUebernehmen(
	ctx context.Context,
	turnierID, paarungID string,
) MatchAntwort {
	return b.app(ctx).TurnierMatchUebernehmen(
		b.ruf(ctx), turnierID, paarungID)
}
