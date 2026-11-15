//go:build !darwin

package app

// **Der Platz auf den Plattformen, die hier nicht gebaut werden** — Werkstück
// O6 von Phase 5.2 (#1430).
//
// **Nicht geraten, sondern gesagt**: gebaut wird für darwin/arm64 (§12), und
// die Übersetzung für Windows und Linux läuft nur mit, damit
// plattformabhängiger Code nicht unbemerkt bricht. Eine erfundene Zahl wäre
// hier schlimmer als keine — sie ließe eine Vorprüfung durchgehen, die nichts
// geprüft hat.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func freierPlatz(string) (int64, error) {
	return 0, fmt.Errorf(
		"%w: der freie Platz lässt sich auf dieser Plattform nicht bestimmen",
		domain.ErrValidation)
}

func platzAuf(string) (frei, gesamt uint64, fehler error) {
	_, fehler = freierPlatz("")

	return 0, 0, fehler
}
