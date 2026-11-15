// **Die Sitzung am Fenster, gebunden** — ID-02
// (#1701).
// Jede Methode reicht durch und tut sonst nichts.

package app

import "context"

// Sitzungsstand sagt, ob gearbeitet, gesperrt oder abgemeldet ist.
func (b *Bindungen) Sitzungsstand(ctx context.Context) SitzungsstandAntwort {
	return b.frei().Sitzungsstand(b.frei().rufkontext(ctx))
}

// Sperren sperrt das Fenster.
func (b *Bindungen) Sperren(ctx context.Context) SchreibAntwort {
	return b.app(ctx).Sperren(b.ruf(ctx))
}

// Entsperren löst die Sperre mit dem Kennwort.
func (b *Bindungen) Entsperren(ctx context.Context, kennwort string) SchreibAntwort {
	return b.frei().Entsperren(b.frei().rufkontext(ctx), kennwort)
}

// EntsperrenMitSensor löst die Sperre über Touch ID.
func (b *Bindungen) EntsperrenMitSensor(ctx context.Context) SchreibAntwort {
	return b.frei().EntsperrenMitSensor(b.frei().rufkontext(ctx))
}

// Abmelden beendet die Sitzung.
func (b *Bindungen) Abmelden(ctx context.Context) SchreibAntwort {
	return b.frei().Abmelden(b.frei().rufkontext(ctx))
}

// LokalAnmelden meldet einen Nutzer dieses Rechners an.
func (b *Bindungen) LokalAnmelden(ctx context.Context, nutzerID, kennwort, nachweis string) SchreibAntwort {
	return b.frei().LokalAnmelden(b.frei().rufkontext(ctx), nutzerID, kennwort, nachweis)
}

// KennwortWechseln setzt ein neues Kennwort für den, der arbeitet.
func (b *Bindungen) KennwortWechseln(ctx context.Context, alt, neu string) SchreibAntwort {
	return b.app(ctx).KennwortWechseln(b.ruf(ctx), alt, neu)
}

// KennwortWiederherstellen setzt ein vergessenes Kennwort über den Zettel.
func (b *Bindungen) KennwortWiederherstellen(ctx context.Context, nutzerID, code, neu string) SchreibAntwort {
	return b.frei().KennwortWiederherstellen(b.frei().rufkontext(ctx), nutzerID, code, neu)
}
