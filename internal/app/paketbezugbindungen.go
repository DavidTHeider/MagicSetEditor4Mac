package app

import "context"

// PaketBezuege nennt die bezogenen Bäume samt ihren Ständen — G04 (#2251).
// Gefragt wird die Quelle nur mit `fragen`.
func (b *Bindungen) PaketBezuege(ctx context.Context, fragen bool) PaketbezuegeAntwort {
	return b.app(ctx).PaketBezuege(b.ruf(ctx), fragen)
}

// PaketBezugVorschau sagt, was eine Aktualisierung auf einen Commit täte.
func (b *Bindungen) PaketBezugVorschau(ctx context.Context, name, stand string) BezugsvorschauAntwort {
	return b.app(ctx).PaketBezugVorschau(b.ruf(ctx), name, stand)
}

// PaketBezugAktualisieren bezieht einen Commit an die Stelle eines Baums und
// hebt den bisherigen auf.
func (b *Bindungen) PaketBezugAktualisieren(ctx context.Context, name, stand string, bestaetigt bool) BezugAntwort {
	return b.app(ctx).PaketBezugAktualisieren(b.ruf(ctx), name, stand, bestaetigt)
}

// PaketBezugZurueck tauscht einen Baum mit seinem aufgehobenen Stand.
func (b *Bindungen) PaketBezugZurueck(ctx context.Context, name string) BezugAntwort {
	return b.app(ctx).PaketBezugZurueck(b.ruf(ctx), name)
}
