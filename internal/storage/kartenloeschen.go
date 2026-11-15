// **Mehrere Karten in einem Zug löschen** — Befund 4 des Zweitblicks über
// Phase 3.7
// (#1168).
//
// # Warum es diesen Weg neben [Kartenschreiber.KarteLoeschen] gibt
//
// **Weil ein Schnitt über mehrere Karten ein Vorgang ist und keine Folge von
// Vorgängen.** `Ausschneiden` löschte Karte für Karte, jede in ihrer eigenen
// Transaktion; scheiterte die dritte, waren die ersten beiden fort — und die
// Ablage trug alle drei. Das ist der Zustand, den niemand mehr auflöst.
//
// **MSE2 bündelt ebenso**: `AddCardAction` nimmt beim Einfügen alle Karten
// zugleich (`src/gui/control/card_list.cpp`, Zeile 170, Fassung `f9d9356` vom
// 2021-07-06), und `doDelete` reicht die ganze Auswahl weiter (ebenda,
// Zeilen 191–199).
//
// **Und der Verlauf wird ein Schritt.** Drei Karten auszuschneiden und dann
// dreimal ⌘Z zu drücken wäre eine Zumutung; dieselbe Gestalt wie beim
// Ersetzen über ein Set (`ersetzungschreiber.go`).

package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KartenLoeschen entfernt mehrere Karten in **einer** Transaktion.
//
// **Ganz oder gar nicht.** Scheitert eine, bleibt keine gelöscht — der
// Aufrufer bekommt den Fehler und ein Set, das noch so aussieht wie vorher.
func (k *Kartenschreiber) KartenLoeschen(ctx context.Context, karteIDs []int64) error {
	if len(karteIDs) == 0 {
		return nil
	}

	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "%d Karten löschen", len(karteIDs))
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	// **Alle oder keine — und nur Karten des gezeigten Profils** (#1815).
	if fehler := kartenImProfil(ctx, tx, karteIDs); fehler != nil {
		return fehler
	}

	// **Eine Gruppe für den ganzen Zug** — ein ⌘Z holt alle zurück.
	gruppe := neueGruppe()

	for _, karteID := range karteIDs {
		stand, lesefehler := kartenstandLesen(ctx, tx, karteID)
		if lesefehler != nil {
			return lesefehler
		}

		if fehler := schreibeKartenverlaufInGruppe(ctx, tx, karteID, stand,
			domain.Kartenstand{}, k.urheber, gruppe); fehler != nil {
			return fehler
		}

		if _, fehler := tx.ExecContext(ctx,
			`DELETE FROM cards WHERE id = ?`, karteID); fehler != nil {
			return uebersetze(fehler, "die Karte %d löschen", karteID)
		}
	}

	if fehler := tx.Commit(); fehler != nil {
		return uebersetze(fehler, "%d Karten löschen", len(karteIDs))
	}

	return nil
}
