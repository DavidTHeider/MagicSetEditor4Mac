package service

// **Zwei Farben für die Läufe** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum sie hier stehen und nicht in einer Testdatei
//
// **Weil zwei Pakete sie brauchen** — `internal/service` prüft damit den
// Ausschnitt, und wer die Naht darüber prüft, will dieselben Farben. Zwei
// Fassungen liefen beim ersten Umbau auseinander, und eine Gegenprobe gegen
// die falsche Farbe ist keine (§3, dieselbe Überlegung wie bei
// [Texturfachprobe]).

import "image/color"

// LinkeFarbe und RechteFarbe malen die zwei Hälften einer Vorlage.
//
// **Volldeckend und weit auseinander**, damit ein Vergleich nicht an einer
// Kantenglättung hängt.
var (
	LinkeFarbe  = color.NRGBA{R: 220, G: 30, B: 30, A: 255}
	RechteFarbe = color.NRGBA{R: 30, G: 60, B: 220, A: 255}
)
