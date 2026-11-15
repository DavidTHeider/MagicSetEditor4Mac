package app

import "context"

// **Die Verbuchung am Fenster** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// **Vier Bindungen und nicht eine.** Verbuchen, Abgleichen und Bestätigen
// sind drei Entscheidungen von Menschen; das Journal ist die Auskunft
// darüber, was daraus geworden ist.

// Handelsverbuchung bucht einen übergebenen Handel im eigenen Bestand.
func (b *Bindungen) Handelsverbuchung(
	ctx context.Context,
	vorgang, nachweis string,
) VerbuchungAntwort {
	return b.app(ctx).Handelsverbuchung(b.ruf(ctx), vorgang, nachweis)
}

// Handelsabgleich meldet, dass hier fertig ist und die Gegenseite fehlt.
func (b *Bindungen) Handelsabgleich(ctx context.Context, vorgang string) VerbuchungAntwort {
	return b.app(ctx).Handelsabgleich(b.ruf(ctx), vorgang)
}

// Handelsbestaetigung nimmt eine Stimme für den beidseitigen Abschluss.
func (b *Bindungen) Handelsbestaetigung(
	ctx context.Context,
	vorgang, wer, nachweis string,
) VerbuchungAntwort {
	return b.app(ctx).Handelsbestaetigung(b.ruf(ctx), vorgang, wer, nachweis)
}

// Handelsbeleg gibt Stand und Beleg eines Handels — rein lesend.
func (b *Bindungen) Handelsbeleg(ctx context.Context, vorgang string) VerbuchungAntwort {
	return b.app(ctx).Handelsbeleg(b.ruf(ctx), vorgang)
}

// Handelsjournal gibt die Bewegungen zu einem Vorgang.
func (b *Bindungen) Handelsjournal(ctx context.Context, vorgang string) JournalAntwort {
	return b.app(ctx).Handelsjournal(b.ruf(ctx), vorgang)
}
