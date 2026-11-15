package app

import "context"

// **Die gebundenen Wege des Ersetzens** — Werkstück 2 von Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, weil `bindungenkarten.go` mit ihnen über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort stehen die Wege zu **einer**
// Karte, hier die über **alle** Karten eines Sets — Ersetzen und die Gestalt
// der Liste.

// **Suchen und Ersetzen** — Werkstück 2 von Phase 3.7
// (#1168).
//
// MSE2 hat `ID_EDIT_FIND`, `ID_EDIT_FIND_NEXT` und `ID_EDIT_REPLACE`; hier gab
// es die Volltextsuche und die Massenänderung, aber **kein Ersetzen im Text**.

// ErsetzungVorschau sagt, was sich ändern würde — und schreibt nichts.
func (b *Bindungen) ErsetzungVorschau(
	ctx context.Context,
	setID string, frage ErsetzungsfrageDTO,
) ErsetzungsAntwort {
	return b.app(ctx).ErsetzungVorschau(b.ruf(ctx), setID, frage)
}

// ErsetzungAusfuehren ersetzt und gibt dieselbe Übersicht zurück.
func (b *Bindungen) ErsetzungAusfuehren(
	ctx context.Context,
	setID string, frage ErsetzungsfrageDTO,
) ErsetzungsAntwort {
	return b.app(ctx).ErsetzungAusfuehren(b.ruf(ctx), setID, frage)
}

// **Die Spalten der Kartenliste und ihre Ordnung** — Werkstück 3 von
// Phase 3.7
// (#1168).

// Kartenlistenaufbau gibt Spalten, Zeilen und die wirksame Ordnung.
func (b *Bindungen) Kartenlistenaufbau(ctx context.Context, setID string) KartenlisteAntwort {
	return b.app(ctx).Kartenlistenaufbau(b.ruf(ctx), setID)
}

// ListenspaltenSetzen schreibt die Spaltenwahl fort — **die ganze Wahl**.
func (b *Bindungen) ListenspaltenSetzen(
	ctx context.Context,
	setID string, wahl []ListenspalteDTO,
) SchreibAntwort {
	return b.app(ctx).ListenspaltenSetzen(b.ruf(ctx), setID, wahl)
}

// ListensortierungSetzen schreibt sie fort — leeres Feld heißt: gar nicht.
func (b *Bindungen) ListensortierungSetzen(
	ctx context.Context,
	setID, feld string, aufsteigend bool,
) SchreibAntwort {
	return b.app(ctx).ListensortierungSetzen(b.ruf(ctx), setID, feld, aufsteigend)
}
