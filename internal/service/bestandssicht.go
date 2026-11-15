package service

// **Für welchen Bestand eine Sicht gerechnet wird** — P9
// (#2377),
// Vermerk 0198.
//
// **Im Kontext und nicht als Argument**, wie die Sitzbindung daneben: der
// Wirt rechnet jedem Gast seine Sicht über dieselbe Naht wie dem Fenster, und
// ein Argument mehr an ihr hieße, jeden ihrer Nehmer zu ändern. Leer heißt:
// der Bestand, in dem die Partie liegt.

import "context"

type bestandssichtSchluessel struct{}

// MitBestandssicht sagt, dass für den Bestand des Peers `peer` gerechnet wird.
func MitBestandssicht(ctx context.Context, peer string) context.Context {
	return context.WithValue(ctx, bestandssichtSchluessel{}, peer)
}

// BestandssichtAus gibt den Bestand, für den gerechnet wird — leer: diesen.
func BestandssichtAus(ctx context.Context) string {
	peer, _ := ctx.Value(bestandssichtSchluessel{}).(string) //nolint:errcheck // fehlt er, ist es dieser

	return peer
}
