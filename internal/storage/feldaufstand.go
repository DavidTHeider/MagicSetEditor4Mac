package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FeldSetzenAufStand schreibt einen Feldwert nur, wenn die Karte noch auf dem
// erwarteten Stand steht ([domain.CardRow.Stand]) — geprüft in derselben
// Transaktion, sonst `ErrConflict` — und gibt den vorigen Wert.
//
// # Warum es das gibt
//
// **Für einen Wert, der aus dem gelesenen gebaut wird** (Befund 1 des
// Nachlaufs über E/F, #2298). Die Zusatzwerte einer Karte stehen zusammen im
// Feld `extra data`; wer eines setzt, schreibt den ganzen Block. Ohne
// Bedingung schrieb ein zweiter Aufruf den Block von vorhin zurück, und die
// Änderung dazwischen ging verloren, während beide Erfolg meldeten.
//
// **Neben [Kartenschreiber.FeldSetzenMitVorwert] und nicht statt seiner**: ein
// getippter Wert hängt nicht am Rest der Karte. Und neben der Übernahme mit
// Herkunft, die denselben Stand prüft, aber eine Herkunft verlangt — eine
// eigene Eingabe hat keine.
func (k *Kartenschreiber) FeldSetzenAufStand(
	ctx context.Context, karteID int64, feld, wert, erwartet string,
) (vorher string, hatte bool, fehler error) {
	return k.imEinenZugAufStand(ctx, karteID,
		[]domain.Value{{Field: feld, Text: wert}}, erwartet, nil)
}
