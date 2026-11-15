package app

import "context"

// **Die Bindung der Regelfreigabe** — Werkstück G4 von Phase 5.5
// (#1461).

// Regelfreigabe sagt, ob eine Partie so anfangen darf.
//
// **Modus und Abwahl kamen mit PLAY-02** (#1710): der Modus sagt, wie viel
// Hilfe gewünscht ist, die Abwahl, welche Bereiche die Runde selbst
// übernimmt. **Leer heißt Anfänger und nichts abgewählt** — die vorsichtige
// Vorgabe.
func (b *Bindungen) Regelfreigabe(
	ctx context.Context,
	spiel, strenge string, sitze int, modus string, abgewaehlt []string,
) RegelfreigabeAntwort {
	return b.app(ctx).Regelfreigabe(
		b.ruf(ctx), spiel, strenge, sitze, modus, abgewaehlt)
}
