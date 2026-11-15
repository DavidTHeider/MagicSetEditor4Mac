package app

import "context"

// Kartenzusatzfelder gibt die Zusatzfelder des Stils einer Karte, die sich von
// Hand setzen lassen, samt Wert
// (#2298).
func (b *Bindungen) Kartenzusatzfelder(ctx context.Context, setID, karteID string) ZusatzfelderAntwort {
	return b.app(ctx).Kartenzusatzfelder(b.ruf(ctx), setID, karteID)
}

// ZusatzwertSetzen setzt den Wert eines Zusatzfelds an einer Karte.
func (b *Bindungen) ZusatzwertSetzen(ctx context.Context, setID, karteID, feld, wert string) SchreibAntwort {
	return b.app(ctx).ZusatzwertSetzen(b.ruf(ctx), setID, karteID, feld, wert)
}
