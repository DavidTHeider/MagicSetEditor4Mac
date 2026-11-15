// **Die Bindungen der Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Eine eigene Datei**, damit `nutzerbindungen.go` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, **wer** hier arbeitet,
// hier, **womit sich ein Automat ausweist**.

package app

import "context"

// Token gibt die Zugriffstoken eines Profils — **ohne ihre Geheimnisse**.
func (b *Bindungen) Token(ctx context.Context, profilID string) TokenAntwort {
	return b.app(ctx).Token(b.ruf(ctx), profilID)
}

// TokenAnlegen legt eines an.
//
// **Der Klartext steht in dieser einen Antwort und danach nirgends mehr.** Die
// Seite sagt das, bevor sie fragt: ein Feld, das man später nachschlagen kann,
// kopiert niemand — und dann fehlt es.
func (b *Bindungen) TokenAnlegen(ctx context.Context, angaben TokenangabenDTO) EinTokenAntwort {
	return b.app(ctx).TokenAnlegen(b.ruf(ctx), angaben)
}

// TokenAendern schreibt Bezeichnung, Erlaubnis und Ablauf fort.
//
// **Das Geheimnis bleibt dasselbe** — wer umbenennt, entzieht niemandem den
// Zugang.
func (b *Bindungen) TokenAendern(ctx context.Context, angaben TokenangabenDTO) EinTokenAntwort {
	return b.app(ctx).TokenAendern(b.ruf(ctx), angaben)
}

// TokenLoeschen zieht eines zurück.
func (b *Bindungen) TokenLoeschen(ctx context.Context, tokenID string) SchreibAntwort {
	return b.app(ctx).TokenLoeschen(b.ruf(ctx), tokenID)
}
