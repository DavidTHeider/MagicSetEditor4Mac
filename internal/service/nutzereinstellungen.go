// **Die Einstellungen, die dem Menschen gehören** — Werkstück 8 von
// Phase 4.12
// (#1351),
// Punkte 9 und 11 der Kommentierung vom 2026-09-05.
//
// # Was hier entschieden wird
//
// **Was ohne Wahl gilt.** Die Vorgabewerte kommen von hier und nicht aus dem
// Fenster — stünde die Regel dort, hielte sie kein Go-Test, und der
// MCP-Server bekäme eine andere Antwort als das Fenster.
//
// # Was hier **noch nicht** steht
//
// **Die Verschmelzung mit den Profileinstellungen.** Sprache, Erscheinungsbild
// und Schriftgröße hängen seit Wanderung 0035 am Profil; sie zur Vorgabe des
// **Menschen** zu machen und vom Profil übersteuern zu lassen, ist die andere
// Hälfte derselben Sache — und sie gehört zu Werkstück 9, wo die
// Übersteuerung des Profils gebaut wird. **Was hier steht, gibt es nur auf
// dieser Ebene** und braucht deshalb keine Verschmelzung.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nutzereinstellungsablage ist die Naht zum Bestand.
type Nutzereinstellungsablage interface {
	Nutzereinstellungen(ctx context.Context, nutzerID int64) (map[string]string, error)
	NutzereinstellungSetzen(
		ctx context.Context, nutzerID int64, schluessel, wert string) error
}

// Nutzereinstellungen verwaltet, was einem Menschen gehört.
type Nutzereinstellungen struct{ ablage Nutzereinstellungsablage }

// NeueNutzereinstellungen baut sie über einer Ablage.
func NeueNutzereinstellungen(ablage Nutzereinstellungsablage) *Nutzereinstellungen {
	return &Nutzereinstellungen{ablage: ablage}
}

// Alle gibt die Einstellungen eines Nutzers, Ungesetztes als Vorgabewert.
//
// **Vollständig und nicht nur das Gesetzte**: eine Antwort, die einen
// Schlüssel wegließe, zwänge jede Fläche, die Vorgabe noch einmal zu kennen —
// und die zweite Liste liefe zurück.
func (n *Nutzereinstellungen) Alle(ctx context.Context, nutzerID int64) (
	map[string]string, error,
) {
	gesetzt, fehler := n.ablage.Nutzereinstellungen(ctx, nutzerID)
	if fehler != nil {
		return nil, fehler
	}

	werte := make(map[string]string, len(domain.Nutzereinstellungsreihe()))

	for _, schluessel := range domain.Nutzereinstellungsreihe() {
		if wert, da := gesetzt[schluessel]; da {
			werte[schluessel] = wert

			continue
		}

		werte[schluessel] = domain.NutzerVorgabe(schluessel)
	}

	return werte, nil
}

// Setzen schreibt eine Einstellung fort — geprüft.
func (n *Nutzereinstellungen) Setzen(
	ctx context.Context, nutzerID int64, schluessel, wert string,
) error {
	if fehler := domain.PruefeNutzereinstellung(schluessel, wert); fehler != nil {
		return fehler
	}

	return n.ablage.NutzereinstellungSetzen(ctx, nutzerID, schluessel, wert)
}

// Wert gibt eine einzelne Einstellung — die Vorgabe, wenn keine dasteht.
//
// **Für den, der genau eine braucht** und nicht die ganze Abbildung: der
// Prüfer des zweiten Faktors fragt nach `sicherheit.faktorbeimwechsel` und
// sonst nichts.
func (n *Nutzereinstellungen) Wert(
	ctx context.Context, nutzerID int64, schluessel string,
) (string, error) {
	alle, fehler := n.Alle(ctx, nutzerID)
	if fehler != nil {
		return "", fehler
	}

	return alle[schluessel], nil
}
