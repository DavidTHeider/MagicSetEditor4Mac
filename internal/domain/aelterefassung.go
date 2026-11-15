package domain

import (
	"fmt"
	"strings"
)

// AeltereFassung ist die Abweisung eines Bestands oder einer Sicherung, die
// **vor** der Grundlinie stehen geblieben ist (J02,
// #2266,
// Vermerk 0189).
//
// **Seit 0.8.0 gibt es nur die Grundlinie**: die Wanderungen bis zu ihr sind
// fort, und ein Bestand, dem eine davon fehlt, lässt sich hier nicht mehr
// wandern. 0.7.0 kennt sie alle — einmal dort geöffnet, steht er auf der
// Grundlinie. Das Gegenstück zu [NeuereFassung], in denselben Teilen.
type AeltereFassung struct {
	// Fehlend sind die Schritte der Grundlinie, die er nicht führt —
	// sortiert, damit die Meldung bei jedem Lauf dieselbe ist.
	Fehlend []string

	// Grundlinie ist der letzte Schritt, den sie enthält.
	Grundlinie string
}

// Error gibt den Satz, der ohne Katalog gelesen wird.
//
// **Bis drei Schritte einzeln, darüber die Spanne**: einem Bestand aus 0.1.0
// fehlen 96, und eine Meldung mit 96 Dateinamen liest niemand.
func (a *AeltereFassung) Error() string {
	fehlt := "fehlt " + strings.Join(a.Fehlend, ", ")
	if len(a.Fehlend) > 3 {
		fehlt = fmt.Sprintf("fehlen %d Wanderungen, von %s bis %s",
			len(a.Fehlend), a.Fehlend[0], a.Fehlend[len(a.Fehlend)-1])
	}

	return fmt.Sprintf("diesem Bestand %s, und diese Fassung wandert "+
		"nicht mehr bis %s — er stammt aus einer Fassung vor 0.7.0; einmal mit "+
		"MagicSetEditor4Mac 0.7.0 geöffnet, steigt er dort auf und öffnet sich hier",
		fehlt, a.Grundlinie)
}

// Unwrap ordnet sie den Sicherungsfehlern zu — wie [NeuereFassung].
func (a *AeltereFassung) Unwrap() error { return ErrSicherung }
