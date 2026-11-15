package service

// **Karten ausschneiden** — Werkstück 1 von Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, weil `kartenablage.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **wie Karten in die Ablage
// kommen und zurück**, hier, **was zusätzlich gilt, wenn sie dabei
// verschwinden sollen**.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ausschneiden legt die Karten ab und entfernt sie **danach**.
//
//	case ID_EDIT_CUT: current_panel->doCut(); break;
//
// `src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06 — MSE2 kennt
// `ID_EDIT_CUT` neben `ID_EDIT_COPY` über Karten.
//
// # Abgelegt wird zuerst, gelöscht danach
//
// **Und zwar wirklich abgelegt, nicht nur zusammengebaut** — Befund 2 des
// Nachlaufs über Phase 3.7.
//
// `ablegen` ist der Weg nach draußen, und er muss glücken, bevor eine Karte
// fällt. **So macht es MSE2**: `doCut` ruft das wirkliche `doCopy()` samt
// `wxTheClipboard->SetData` und löscht nur bei dessen Erfolg
// (`src/gui/control/item_list.cpp`, Zeilen 67–72; `card_list.cpp`,
// Zeilen 146–156).
//
// **Scheitert das Löschen**, steht der Inhalt in der Ablage und die Karte
// noch da — das ist wiederholbar. Andersherum wäre die Karte fort und die
// Ablage leer.
func (k *Kartenablage) Ausschneiden(
	ctx context.Context, setID int64, karteIDs []int64, ablegen func(string) error,
) (string, error) {
	if k.loescher == nil {
		return "", fmt.Errorf(
			"%w: das Ausschneiden ist nicht eingerichtet", domain.ErrValidation)
	}

	inhalt, err := k.Kopieren(ctx, setID, karteIDs)
	if err != nil {
		return "", err
	}

	if fehler := ablegen(inhalt); fehler != nil {
		return "", fmt.Errorf("die Karten ablegen: %w", fehler)
	}

	// **Und zwischen Lesen und Löschen kann jemand geschrieben haben.**
	//
	// `ablegen` geht nach draußen — an die Zwischenablage des Systems — und
	// währenddessen läuft das Fenster weiter: ein Feldschreiben nimmt einen
	// **anderen** Vorgangsschlüssel und wird nicht angehalten. Gelöscht wurde
	// danach allein nach Kennung, und der neue Stand war fort, ohne je in der
	// Ablage gestanden zu haben.
	//
	// **Also noch einmal lesen und vergleichen.** Stimmt es nicht mehr, bleibt
	// die Karte stehen und der Inhalt in der Ablage — dasselbe Verhalten wie
	// bei einem gescheiterten Löschen, und wiederholbar. Angestrichen vom
	// Nachlauf über die Nacharbeit
	// (#1282,
	// Befund 7).
	nachher, nachfehler := k.Kopieren(ctx, setID, karteIDs)
	if nachfehler != nil {
		return inhalt, fmt.Errorf("den Stand nachlesen: %w", nachfehler)
	}

	if nachher != inhalt {
		return inhalt, fmt.Errorf(
			"%w: an diesen Karten hat sich etwas geändert, während sie in die "+
				"Zwischenablage gingen — sie stehen noch da, und der abgelegte "+
				"Stand ist der von vorhin",
			domain.ErrConflict)
	}

	// **Ohne Bestätigung** (#1369): wer ausschneidet, will die Karten
	// weiterreichen und nicht einen Deckplatz verlieren. Liegt eine in einem
	// Deck, wird das Ausschneiden abgewiesen — mit den Namen der Decks. Ein
	// „trotzdem" gibt es hier **nicht**, und das ist ausgesprochen: wer die
	// Karte wirklich loswerden will, nimmt sie aus dem Deck oder löscht sie
	// ausdrücklich.
	if fehler := k.loescher.KartenLoeschen(ctx, setID, karteIDs, false); fehler != nil {
		return inhalt, fmt.Errorf("die Karten entfernen: %w", fehler)
	}

	return inhalt, nil
}
