package app

import "context"

// **Die Bindungen der Marken an jedem Fachobjekt** — UX-12
// (#1684).

// ObjektmarkeSetzen hängt eine Marke an ein Objekt.
func (b *Bindungen) ObjektmarkeSetzen(ctx context.Context, art, kennung, marke string) SchreibAntwort {
	return b.app(ctx).ObjektmarkeSetzen(b.ruf(ctx), art, kennung, marke)
}

// ObjektmarkeEntfernen nimmt eine Marke von einem Objekt.
func (b *Bindungen) ObjektmarkeEntfernen(ctx context.Context, art, kennung, marke string) SchreibAntwort {
	return b.app(ctx).ObjektmarkeEntfernen(b.ruf(ctx), art, kennung, marke)
}

// MarkeVergeben hängt eine Marke an mehrere Objekte.
func (b *Bindungen) MarkeVergeben(ctx context.Context, art string, kennungen []string, marke string) SchreibAntwort {
	return b.app(ctx).MarkeVergeben(b.ruf(ctx), art, kennungen, marke)
}

// MarkenJeObjekt nennt die Marken jedes Objekts einer Art.
func (b *Bindungen) MarkenJeObjekt(ctx context.Context, art string) MarkenJeObjektAntwort {
	return b.app(ctx).MarkenJeObjekt(b.ruf(ctx), art)
}

// MarkenFiltern gibt die Objekte einer Art, die durch den Filter gehen.
func (b *Bindungen) MarkenFiltern(ctx context.Context, art string, marken []string, verknuepfung string) MarkenfilterAntwort {
	return b.app(ctx).MarkenFiltern(b.ruf(ctx), art, marken, verknuepfung)
}

// Markenkatalog zählt die Marken des Profils.
func (b *Bindungen) Markenkatalog(ctx context.Context) MarkenkatalogAntwort {
	return b.app(ctx).Markenkatalog(b.ruf(ctx))
}

// MarkeUmbenennen gibt einer Marke einen neuen Namen.
func (b *Bindungen) MarkeUmbenennen(ctx context.Context, alt, neu string, zusammenfuehren bool) MarkenzahlAntwort {
	return b.app(ctx).MarkeUmbenennen(b.ruf(ctx), alt, neu, zusammenfuehren)
}

// MarkeUeberallEntfernen nimmt eine Marke von jedem Objekt des Profils.
func (b *Bindungen) MarkeUeberallEntfernen(ctx context.Context, marke string) MarkenzahlAntwort {
	return b.app(ctx).MarkeUeberallEntfernen(b.ruf(ctx), marke)
}
