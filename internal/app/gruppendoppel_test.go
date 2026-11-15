package app_test

// **Was die Doppel über Gruppen sagen** — abgetrennt am 2026-09-13, weil
// `kivorschlag_test.go` mit ihnen über die Zeilengrenze aus §2 ging (A13.1b, #1452).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die Doppel für den **einzelnen** Weg; hier die Antwort auf die
// eine Frage, die A13.1b hinzugefügt hat: was geschieht, wenn jemand eine
// **Gruppe** schreiben will. **Sie scheitern laut** (§3) — ein Doppel, das
// hier still gutginge, ließe einen Lauf über Gruppen grün werden, der gar
// keine geschrieben hat.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// GruppeVonAussenMitHerkunft gibt es an diesem Doppel **nur der
// Schnittstelle wegen** (A13.1b, #1452).
//
// **Sie scheitert laut.** Ein Doppel, das hier still gutginge, ließe einen
// Lauf über Gruppen grün werden, der gar keine geschrieben hat — §3: eine
// Vorrichtung, die großzügiger ist als der Betrieb.
// **Das Kartendoppel schreibt sie sehr wohl** — es ist die Ablage der
// KI-Welt, und die Gruppe geht durch **denselben** Weg wie der einzelne
// Vorschlag. Ein Doppel, das hier scheiterte, machte den geglückten Weg
// unprüfbar.
//
// **Es fährt die Schritte in einer Schleife**, und das ist hier richtig: was
// „eine Transaktion" heißt, entscheidet die Ablage, und die steht in
// `internal/storage`. Geprüft wird dort, gegen eine echte Datenbank.
func (k *kartendoppel) GruppeVonAussenMitHerkunft(
	ctx context.Context, schritte []domain.Kartenschritt,
) error {
	for _, schritt := range schritte {
		if _, _, fehler := k.FelderVonAussenMitHerkunft(ctx, schritt.SetID,
			schritt.KarteID, schritt.Werte, schritt.Herkunft,
			schritt.Herkuenfte, schritt.Angaben, ""); fehler != nil {
			return fehler
		}
	}

	return nil
}
