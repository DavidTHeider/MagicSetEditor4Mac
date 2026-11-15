package app

import "context"

// **Die Bindungen des Probelaufs** — UXR-F7
// (#2015).
//
// Der Ausgabeassistent fragt vor dem Schreiben, was herauskäme: für den
// Druckbogen („9 je Bogen · 2 Bogen") und für die Exportvorlage („12 KB für
// 41 Karten"). **Beide lesen nur**: kein Dialog, kein Pfad, kein Eintrag im
// Auftragsbuch, und der Schreibriegel (0078) wird nicht gefragt.

// BogenProbelauf rechnet den Druckbogen durch, ohne ihn zu legen.
//
// **Mit denselben Wahlen wie `BogenLegen`** — und ohne Dialog.
func (b *Bindungen) BogenProbelauf(
	ctx context.Context,
	setID, papier string, anschnitt float64, dpi int, marken bool,
	auswahl []string,
) BogenprobelaufAntwort {
	return b.app(ctx).BogenProbelauf(b.ruf(ctx), setID, papier, anschnitt, dpi, marken, auswahl)
}

// VorlageProbelauf wertet eine Exportvorlage aus, ohne etwas zu schreiben.
//
// **Mit denselben Wahlen wie `VorlageAusfuehren`** — und ohne Ziel: das Fenster
// kennt es erst nach dem Dialog, und der Probelauf sagt den Namen, den der
// Dialog vorschlägt.
func (b *Bindungen) VorlageProbelauf(
	ctx context.Context,
	setID, paket string, wahl map[string]string, auswahl []string,
) VorlagenprobelaufAntwort {
	return b.app(ctx).VorlageProbelauf(b.ruf(ctx), setID, paket, "", wahl, auswahl)
}
