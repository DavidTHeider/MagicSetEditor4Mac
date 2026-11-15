package storage

// **Mehrere Karten in einem Zug** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Warum das nicht [Kartenschreiber.FelderSetzenMitHerkunft] in einer
// Schleife ist
//
// **Weil eine Schleife N Transaktionen sind.** Zwischen der dritten und der
// vierten stünde ein Bestand, in dem drei Karten den Vorschlag tragen und
// zwei nicht — genau der Teilbestand, den das Issue ausschließt („entweder
// vollständig … oder klarer Konflikt ohne Teilbestand").
//
// # Und warum die Herkunft je Karte geschrieben wird
//
// **Weil sie an der Karte hängt** ([herkuenfteSchreiben] nimmt eine
// Kartenkennung). Die Zusage ist nicht „eine Herkunftszeile für alle",
// sondern „jede Karte trägt ihre, und entweder alle oder keine".

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Schritt selbst steht in `domain`** ([domain.Kartenschritt]): ihn
// nennen drei Schichten, und `storage` darf `service` nicht kennen (§2).
//
// **Den Vorwert gibt die Gruppe nicht zurück.** Er ist die Auskunft für eine
// Rücknahme am Fenster, und eine Gruppe wird als Gruppe zurückgenommen oder
// gar nicht.

// GruppeSetzenMitHerkunft schreibt **alle** Schritte in einer Transaktion.
//
// **Eine leere Gruppe ändert nichts und ist kein Fehler** — dieselbe Regel wie
// bei den Werten. Ob eine leere überhaupt vorkommen darf, entscheidet
// `service`.
func (k *Kartenschreiber) GruppeSetzenMitHerkunft(
	ctx context.Context, schritte []domain.Kartenschritt,
) (fehler error) {
	if len(schritte) == 0 {
		return nil
	}

	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "%d Karten ändern", len(schritte))
	}

	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	// **Und ein Schritt im Verlauf** (#2396): wer die Gruppe zurücknimmt,
	// nimmt sie ganz zurück — ein ⌘Z je Feld ließe den Teilbestand stehen,
	// den diese Transaktion ausschließt.
	gruppe := k.inEinemSchritt()

	for _, schritt := range schritte {
		if fehler := gruppe.einSchritt(ctx, tx, profilID, schritt); fehler != nil {
			return fehler
		}
	}

	return bestaetige(tx, "%d Karten ändern", len(schritte))
}

// einSchritt schreibt eine Karte **in die laufende Transaktion**.
//
// **Ohne eigenes `BeginTx`** — das ist der ganze Unterschied zu
// [Kartenschreiber.FelderSetzenMitHerkunft], und er ist der Grund, warum es
// diese Datei gibt.
func (k *Kartenschreiber) einSchritt(
	ctx context.Context, tx *sql.Tx, profilID int64, schritt domain.Kartenschritt,
) error {
	if len(schritt.Herkuenfte) == 0 {
		return fmt.Errorf(
			"%w: eine Übernahme trägt ihre Herkunft — ohne sie ist sie keine",
			domain.ErrValidation)
	}

	// **Nur an einer Karte des gezeigten Profils** (#1815). Das Profil stand
	// schon im Zug — für die Herkunftszeilen, nicht für die Karte: die Gruppe
	// schrieb den Namen einer fremden Karte um und trug die Herkunft im
	// eigenen Profil ein.
	if fehler := karteImProfil(ctx, tx, schritt.KarteID); fehler != nil {
		return fehler
	}

	// **Der Stand wird hier geprüft, in der laufenden Transaktion** — warum
	// nicht im Aufrufer, steht bei [standHalten].
	if fehler := standHalten(ctx, tx, schritt.KarteID, schritt.Erwartet); fehler != nil {
		return fehler
	}

	// **Die Marke am Feld bleibt** ([Kartenschreiber.Von], #888): sie steht im
	// Verlauf und sagt „nicht getippt". Die Zeilen daneben sagen, **wer** und
	// **worauf**.
	von := k.Von(schritt.Herkunft)

	for i := range schritt.Werte {
		wert := schritt.Werte[i]

		// **Nachgezogenes ohne Bedingung**, wie beim einzelnen Feld.
		lauf := ctx
		if wert.Nachgezogen {
			lauf = ohneBekannte(ctx)
		}

		if _, _, _, fehler := von.imLauf(lauf, tx, schritt.KarteID, wert.Field,
			func(tx *sql.Tx) error {
				_, fehler := tx.ExecContext(ctx, feldSetzen,
					schritt.KarteID, wert.Field, wert.Text, schritt.KarteID)

				return fehler
			}); fehler != nil {
			return fehler
		}
	}

	if fehler := herkuenfteSchreiben(
		ctx, tx, profilID, schritt.KarteID, schritt.Herkuenfte); fehler != nil {
		return fehler
	}

	// **Die Bildangaben nach der Feldherkunft** (A12, #1451): ihr
	// Fremdschlüssel zeigt auf sie, und eine Angabe ohne ihre Herkunft wäre
	// eine Angabe über nichts.
	for feld, je := range schritt.Angaben {
		if fehler := bildangabenSetzen(
			ctx, tx, profilID, schritt.KarteID, feld, je); fehler != nil {
			return fehler
		}
	}

	return nil
}
