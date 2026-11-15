package app

// **Zwei Fenster in das Paket** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Warum sie hier stehen
//
// **Weil zwei Wege der Bestenfalls-Arbeit von außen nicht erreichbar sind**:
// [App.mitBisherigem] antwortet auch ohne Bestand, und das ist die Zusage —
// ein Vorschlag ohne Vergleich ist weniger wert als einer mit, gar keiner ist
// schlechter als beides. Von außen käme man nur über [App.KIVorschlagen]
// dorthin, und der scheitert vorher an derselben fehlenden Ablage.
//
// **Nur für Läufe**: diese Datei endet auf `_test.go` und steht in keiner
// Auslieferung.

import "context"

// MitBisherigemZumPruefen ruft [App.mitBisherigem].
func MitBisherigemZumPruefen(
	a *App, ctx context.Context, karteID int64, felder []KIFeldwertDTO,
) []KIFeldwertDTO {
	return a.mitBisherigem(ctx, karteID, felder)
}
