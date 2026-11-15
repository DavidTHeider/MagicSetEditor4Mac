package fenster

// **Das Fenster passt auf den Bildschirm, auf dem es aufgeht** — UX-14.AK-04
// (#1686):
// *„Gespeicherte Positionen werden bei Monitorwechsel plausibilisiert."*
//
// # Was gemerkt wird
//
// **Die Größe, keine Position** (`app.FenstergroesseMerken`): Wails öffnet
// das Fenster zentriert. Wer es auf einem großen Bildschirm groß gezogen hat
// und ihn dann abzieht, bekäme beim nächsten Start ein Fenster, das über den
// Rand des kleinen hinausragt — mit dem Rand, an dem man es fasst, außer
// Reichweite. `domain.PruefeFenstermass` lässt bis 8.192 Punkte zu; welcher
// Bildschirm gerade dran ist, weiß es nicht.
//
// # Warum erst nach dem Start
//
// **Weil Wails die Bildschirme erst dann kennt** — nachgelesen in
// v3.0.0-beta.23: `ScreenManager.GetPrimary` ist `nil`, bis
// `processAndCacheScreens` beim Start der Anwendung gelaufen ist. Die Größe
// wird deshalb im Startereignis geklemmt (`wails.go`), und die Entscheidung
// steht hier, wo sie sich messen lässt.

// Das Mindestmaß des Fensters — darunter trägt die Oberfläche nicht.
const (
	mindestbreite = 1024
	mindesthoehe  = 700
)

// passend klemmt eine Fenstergröße auf die Arbeitsfläche eines Bildschirms
// und sagt, ob sich etwas geändert hat.
//
// **Eine Fläche ohne Maß ist eine unbekannte** und keine von null Punkten:
// dann bleibt die Größe, wie sie ist. **Und nie unter das Mindestmaß**: Wails
// senkt es mit `SetSize` still mit, und das Fenster ließe sich danach kleiner
// ziehen, als die Oberfläche trägt (Zweitblick über Phase 6.3, G7).
func passend(breite, hoehe, flaecheBreite, flaecheHoehe int) (neueBreite, neueHoehe int, geaendert bool) {
	if flaecheBreite <= 0 || flaecheHoehe <= 0 {
		return breite, hoehe, false
	}

	neueBreite = max(min(breite, flaecheBreite), mindestbreite)
	neueHoehe = max(min(hoehe, flaecheHoehe), mindesthoehe)

	return neueBreite, neueHoehe, neueBreite != breite || neueHoehe != hoehe
}

// inPunkten rechnet eine Arbeitsfläche aus Bildpunkten in Punkte um.
//
// **Wails liefert sie für das Fenster in Bildpunkten** — nachgelesen in
// v3.0.0-beta.23 (`pkg/application/screen_darwin.go`, `cScreenToScreen`):
// `visibleFrame` mal `backingScaleFactor`, und `getScreenForWindow` teilt
// nicht zurück. `Size` und `SetSize` rechnen in Punkten; ohne diese
// Umrechnung war die Fläche auf Retina doppelt so groß und die Klemmung griff
// nie (Zweitblick über Phase 6.3, S2). **Ein Faktor ohne Wert ist keiner.**
func inPunkten(breite, hoehe int, faktor float32) (breiteInPunkten, hoeheInPunkten int) {
	if faktor <= 0 {
		return breite, hoehe
	}

	return int(float32(breite) / faktor), int(float32(hoehe) / faktor)
}
