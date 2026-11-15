// **Die Bindungen des Adressbuchs** — ID-04
// (#1703).

package app

import "context"

// KontakteEinlesen fragt nach einer vCard-Datei und zeigt ihre Karten.
func (b *Bindungen) KontakteEinlesen(ctx context.Context) KontaktvorschauAntwort {
	return b.app(ctx).KontakteEinlesen(b.ruf(ctx))
}

// KontakteUebernehmen schreibt die gewählten Kontakte.
func (b *Bindungen) KontakteUebernehmen(ctx context.Context, auswahl []KontaktDTO) KontaktuebernahmeAntwort {
	return b.app(ctx).KontakteUebernehmen(b.ruf(ctx), auswahl)
}

// Kontakte gibt die Kontakte des Profils.
func (b *Bindungen) Kontakte(ctx context.Context) KontakteAntwort {
	return b.app(ctx).Kontakte(b.ruf(ctx))
}

// KontaktLoeschen nimmt einen Kontakt heraus.
func (b *Bindungen) KontaktLoeschen(ctx context.Context, id, fassung int64) SchreibAntwort {
	return b.app(ctx).KontaktLoeschen(b.ruf(ctx), id, fassung)
}
