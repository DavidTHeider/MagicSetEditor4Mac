package app

import (
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// uebersichtsnaehte sind die Nähte der Startseite — Werkstück U11 von Phase
// 5.1 (#1422).
//
// **Eingebettet und nicht benannt**, damit `a.weiterarbeiten` bleibt, was es
// ist: `anwendung.go` stand mit ihnen bei 409 Zeilen, und die Grenze aus §2
// gilt auch für den Composition Root.
type uebersichtsnaehte struct {
	// **Und die Kommandozentrale** (U12, #1423) — aus demselben Grund hier.
	befehlsnaehte

	// benutzung merkt sich, welches Set zuletzt aufgeschlagen wurde (#891) —
	// die Liste „Zuletzt benutzt", die in der Übersicht aufgegangen ist.
	benutzung Setbenutzung

	// weiterarbeiten führt die Übersicht.
	weiterarbeiten Weiterarbeitsnaht

	// auftraege ist das Auftragsbuch der Hintergrundaufgaben — beim ersten Ruf
	// angelegt, hinter einem Riegel, weil drei Wege es zugleich erreichen können.
	auftraege      *service.Auftragsbuch
	auftragsriegel sync.Mutex

	// skriptlaeufe hält den Abbruch je Karte (#1424).
	skriptlaeufe

	// probleme ist die Problemprüfung der Sets über dem Zeichenwerk des
	// Fensters — für die Kacheln der Übersicht (AP-8, #2365).
	probleme *Problemweg
}
