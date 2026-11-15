package storage

// **Mehrere Karten in einem Zug** — Werkstück B von Phase 4.11
// (#1337).
//
// # Was nachgelesen wurde (§15)
//
// `AddCardsScript::perform` (`src/data/add_cards_script.cpp`, Zeilen 37–46,
// Fassung f9d9356 vom 2021-07-06) hängt die Karten in **einer** Handlung ans
// Set:
//
//	if (!cards.empty()) {
//	  set.actions.addAction(make_unique<AddCardAction>(ADD, set, cards));
//	}
//
// **Ein ⌘Z holt sie zusammen zurück.** Zwölf Grundlande als zwölf Schritte
// wären zwölfmal ⌘Z für einen Handgriff — und §15 sagt: Semantik konservativ.
//
// # Warum es hier steht und nicht in `KarteAnlegen`
//
// Weil die Gruppe die einzige Zutat ist, die fehlt. `karteEinreihen` legt seit
// je eine Karte samt Feldern und Verlaufsschritt an; was eine Reihe daraus
// macht, ist die gemeinsame Marke — dieselbe Bauform wie [Kartenschreiber.KartenLoeschen]
// (Befund 4 des Zweitblicks über Phase 3.7).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KartenAnlegen legt mehrere Karten in **einer** Transaktion und einem
// Verlaufsschritt an.
//
// **Ganz oder gar nicht.** Scheitert die siebte, ist keine angelegt — sonst
// stünden sechs Karten da, die niemand bestellt hat, und der Rest fehlte.
func (k *Kartenschreiber) KartenAnlegen(
	ctx context.Context, setID int64, karten [][]domain.Feldstand,
) ([]int64, error) {
	if len(karten) == 0 {
		return []int64{}, nil
	}

	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, uebersetze(err, "%d Karten im Set %d anlegen", len(karten), setID)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	// **Nur in ein Set des gezeigten Profils** (#1815).
	if fehler := setImProfil(ctx, tx, setID); fehler != nil {
		return nil, fehler
	}

	// **Eine Gruppe für den ganzen Zug** — ein ⌘Z holt alle zurück.
	gruppe := neueGruppe()

	kennungen := make([]int64, 0, len(karten))

	for _, felder := range karten {
		kennung, fehler := k.karteEinreihenInGruppe(ctx, tx, setID, felder, gruppe)
		if fehler != nil {
			return nil, fehler
		}

		kennungen = append(kennungen, kennung)
	}

	if fehler := tx.Commit(); fehler != nil {
		return nil, uebersetze(fehler, "%d Karten im Set %d anlegen", len(karten), setID)
	}

	return kennungen, nil
}
