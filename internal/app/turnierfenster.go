package app

import "context"

// **Die Bindungen für Turniere** — Werkstück 5 von Phase 4.7
// (#1294).
//
// Sie hängen den Kontext des Fensters an und reichen weiter; entschieden wird
// eine Schicht tiefer (§2).

func (b *Bindungen) Turniere(ctx context.Context, spiel string) TurniereAntwort {
	return b.app(ctx).Turniere(b.ruf(ctx), spiel)
}

func (b *Bindungen) Turnier(ctx context.Context, turnierID string) TurnierAntwort {
	return b.app(ctx).Turnier(b.ruf(ctx), turnierID)
}

func (b *Bindungen) TurnierAnlegen(
	ctx context.Context,
	name, spiel, notiz string, einzelspiel bool, formatID string,
) TurnierAntwort {
	return b.app(ctx).TurnierAnlegen(b.ruf(ctx), name, spiel, notiz, einzelspiel, formatID)
}

// TurnierAbschliessen schließt ein Turnier ab (DOM-04.AK-04, #1699).
func (b *Bindungen) TurnierAbschliessen(ctx context.Context, turnierID, fassung string) SchreibAntwort {
	return b.app(ctx).TurnierAbschliessen(b.ruf(ctx), turnierID, fassung)
}

// TurnierWiedereroeffnen nimmt den Abschluss zurück.
func (b *Bindungen) TurnierWiedereroeffnen(ctx context.Context, turnierID, fassung string) SchreibAntwort {
	return b.app(ctx).TurnierWiedereroeffnen(b.ruf(ctx), turnierID, fassung)
}

// TurnierProtokoll gibt, was an einem Turnier geschah.
func (b *Bindungen) TurnierProtokoll(ctx context.Context, turnierID string) ProtokollAntwort {
	return b.app(ctx).TurnierProtokoll(b.ruf(ctx), turnierID)
}

// TurnierKorrekturfolgen sagt, was eine Berichtigung verschöbe (DOM-04.AK-02).
func (b *Bindungen) TurnierKorrekturfolgen(
	ctx context.Context,
	turnierID, paarungID string, stand *ErgebnisDTO,
) KorrekturfolgenAntwort {
	return b.app(ctx).TurnierKorrekturfolgen(b.ruf(ctx), turnierID, paarungID, stand)
}

// TurnierKopieren legt ein Turnier nach einer Vorlage an (DOM-04, #1699).
func (b *Bindungen) TurnierKopieren(
	ctx context.Context,
	turnierID, name string, mitTeilnehmenden bool,
) TurnierAntwort {
	return b.app(ctx).TurnierKopieren(b.ruf(ctx), turnierID, name, mitTeilnehmenden)
}

// TurnierLoeschen nimmt ein Turnier samt seinem Verlauf heraus.
//
// **Ohne Rückfrage hier** — die gehört ins Fenster, wo jemand sie sieht.
func (b *Bindungen) TurnierLoeschen(ctx context.Context, turnierID string) SchreibAntwort {
	return b.app(ctx).TurnierLoeschen(b.ruf(ctx), turnierID)
}

func (b *Bindungen) TurnierAnmelden(ctx context.Context, turnierID, name, deckID string) TeilnehmerAntwort {
	return b.app(ctx).TurnierAnmelden(b.ruf(ctx), turnierID, name, deckID)
}

func (b *Bindungen) TurnierAussteigen(
	ctx context.Context,
	turnierID, teilnehmerID string, ausgestiegen bool,
) SchreibAntwort {
	return b.app(ctx).TurnierAussteigen(
		b.ruf(ctx), turnierID, teilnehmerID, ausgestiegen)
}

// TurnierPaarungsvorschlag zeigt die nächste Runde, ohne sie zu legen.
func (b *Bindungen) TurnierPaarungsvorschlag(ctx context.Context, turnierID string) PaarungenAntwort {
	return b.app(ctx).TurnierPaarungsvorschlag(b.ruf(ctx), turnierID)
}

func (b *Bindungen) TurnierRundeLegen(ctx context.Context, turnierID string) RundeAntwort {
	return b.app(ctx).TurnierRundeLegen(b.ruf(ctx), turnierID)
}

// TurnierErgebnisSetzen trägt ein Ergebnis ein.
//
// **Ein Zeiger, weil `nil` etwas bedeutet**: doch nicht gespielt. Ein
// Verbund mit drei Nullen hieße 0–0–0, und das ist ein anderer Satz.
func (b *Bindungen) TurnierErgebnisSetzen(
	ctx context.Context,
	turnierID, fassung, paarungID string, stand *ErgebnisDTO,
) SchreibAntwort {
	return b.app(ctx).TurnierErgebnisSetzen(
		b.ruf(ctx), turnierID, fassung, paarungID, stand)
}

func (b *Bindungen) TurnierRangliste(ctx context.Context, turnierID string) RanglisteAntwort {
	return b.app(ctx).TurnierRangliste(b.ruf(ctx), turnierID)
}
