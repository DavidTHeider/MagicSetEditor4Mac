package app

import (
	"context"
)

// **Marken an Karten** — Werkstück 3 von Phase 3.5
// (#1208).
//
// # Was hier nicht entschieden wird
//
// **Was eine gültige Marke ist** — das steht in [domain.PruefeMarke], und der
// Weg dorthin in `service.Markenpflege`. Diese Fassade übersetzt Kennungen und
// Fehler und sonst nichts (§2).

// Markenpflege ist, was die Fassade von den Marken braucht — die Naht zu
// `service`.
//
// # Warum nur die beiden schreibenden Wege
//
// **Weil das Fenster die Marken schon hat.** Sie reisen mit den Zeilen der
// Kartenliste ([ListenzeileDTO.Marken]); eine Bindung, die sie ein zweites Mal
// holte, wäre ein zweiter Stand derselben Sache (§7, Regel 2) — nach dem
// Setzen einer Marke wären beide für einen Augenblick verschieden.
//
// **Die drei Flächen ohne Fenster lesen sie sehr wohl** (`lesen.marken`,
// `mse marken`, `GET /sets/{id}/marken`): dort gibt es keine Kartenliste, aus
// der sie mitkämen. Die Naht dafür ist [Fremdmarken].
type Markenpflege interface {
	Setzen(ctx context.Context, setID, karteID int64, marke string) error
	Entfernen(ctx context.Context, setID, karteID int64, marke string) error
}

// MarkeSetzen hängt eine Marke an eine Karte.
func (a *App) MarkeSetzen(ctx context.Context, setID, karteID, marke string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, kartenkennung, befund := kartenstelle(setID, karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{
		Fehler: Uebersetze(ctx, a.marken.Setzen(ctx, setkennung, kartenkennung, marke)),
	}
}

// MarkeEntfernen nimmt eine Marke wieder ab.
func (a *App) MarkeEntfernen(ctx context.Context, setID, karteID, marke string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, kartenkennung, befund := kartenstelle(setID, karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{
		Fehler: Uebersetze(ctx, a.marken.Entfernen(ctx, setkennung, kartenkennung, marke)),
	}
}

// kartenstelle liest Set- und Kartenkennung — beide, oder keine.
func kartenstelle(setID, karteID string) (set, karte int64, befund *Fehler) {
	set, befund = benannteKennungAus("setID", setID)
	if befund != nil {
		return 0, 0, befund
	}

	karte, befund = benannteKennungAus("karteID", karteID)
	if befund != nil {
		return 0, 0, befund
	}

	return set, karte, nil
}

// MitMarken setzt die Naht zur Markenpflege — für Tests und den Composition
// Root.
func (a *App) MitMarken(pflege Markenpflege) { a.marken = pflege }
