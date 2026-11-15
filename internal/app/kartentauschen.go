package app

// Karten umordnen — #628.
//
// **Getauscht wird mit dem Nachbarn**, und das ist MSEs Elementarschritt: es
// ordnet durch Ziehen um und löst je überfahrenem Element einen Tausch aus.
// Warum die Referenz ihn nicht ausführt, steht bei `storage.KartenTauschen`.

import (
	"context"
)

// Kartentauscher vertauscht zwei Stellen — die Naht zum Speicher.
type Kartentauscher interface {
	KartenTauschen(ctx context.Context, setID, eineID, andereID int64) error
}

// KartenTauschen vertauscht die Stellen zweier Karten.
//
// **Zwei Kennungen und keine Richtung.** Welche der beiden „oben" ist, weiß
// die Liste; diese Schicht trifft keine fachliche Entscheidung (§2).
func (a *App) KartenTauschen(
	ctx context.Context, setID, eineID, andereID string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	eine, befund := kennungAus(eineID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	andere, befund := kennungAus(andereID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{
		Fehler: Uebersetze(ctx, a.tauscher.KartenTauschen(ctx, setkennung, eine, andere)),
	}
}

// MitKartentauscher setzt die Naht — für Tests und den Composition Root.
func (a *App) MitKartentauscher(tauscher Kartentauscher) { a.tauscher = tauscher }
