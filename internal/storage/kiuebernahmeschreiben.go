package storage

// **Die Übernahme aus einem KI-Vorschlag** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Eine eigene Datei
//
// `feldaenderung.go` stand mit ihr an der Zeilengrenze aus §2, und der
// Schnitt folgt der Frage: dort steht, **wie ein Feld sich ändert** — hier,
// **was eine Übernahme zusätzlich schreibt**.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FelderSetzenMitHerkunft schreibt Werte **und ihre Herkunft in einem Zug**.
//
// # Die Zusage
//
// Werkstück A11 von Phase 5.3
// (#1450):
// „Scheitert Provenienz/Verlauf in derselben Transaktion, stehen keine halben
// Inhaltsänderungen."
//
// **Deshalb hier und nicht in einer eigenen Ablage**: eine zweite Ablage
// hätte eine zweite Verbindung und damit eine zweite Transaktion — und
// zwischen ihnen einen Stand, in dem die Karte einen KI-Wert trägt, der
// nirgends als solcher steht.
//
// **Ohne Herkunft ist es keine Übernahme.** Wer ohne schreiben will, nimmt
// [Kartenschreiber.FelderSetzenVonAussen]; eine Zusage, die man weglassen
// kann, ist keine.
//
// **`angaben` sind die Bildangaben je Feld** (A12, #1451) — leer bei jeder
// Textübernahme. **Als Argument und nicht als zweiter Weg**: ein zweiter Weg
// hätte einen zweiten Regelsatz, und der liefe beim ersten Sonderfall
// auseinander, ohne dass es jemandem auffiele.
// **`erwartet` ist der Stand, auf dem die Karte stehen muss** — leer heißt:
// ohne Bedingung. Geprüft wird er **in** der Transaktion und nicht davor
// ([standHalten], Befund 2 des Zweitblicks über Phase 5.3).
//
// **Ein Argument und kein zweiter Name.** Ein Name daneben hätte bedeutet,
// dass es zwei Schreibwege gibt; es gibt einen, und wer ihn ruft, sagt
// jedes Mal, ob er eine Bedingung hat. Das `""` bei einem Aufrufer ohne
// Bedingung ist die Frage, die man stellen soll.
func (k *Kartenschreiber) FelderSetzenMitHerkunft(
	ctx context.Context, karteID int64, werte []domain.Value,
	herkunft string, herkuenfte []domain.KIFeldherkunft,
	angaben map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	if len(herkuenfte) == 0 {
		return "", false, fmt.Errorf(
			"%w: eine Übernahme trägt ihre Herkunft — ohne sie ist sie keine",
			domain.ErrValidation)
	}

	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return "", false, fehler
	}

	// **Die Marke am Feld bleibt** ([Kartenschreiber.Von], #888): sie steht im
	// Verlauf und sagt „nicht getippt". Die Zeilen daneben sagen, **wer** und
	// **worauf** — das eine ersetzt das andere nicht.
	return k.Von(herkunft).imEinenZugAufStand(
		ctx, karteID, werte, erwartet, func(tx *sql.Tx) error {
			if fehler := herkuenfteSchreiben(
				ctx, tx, profilID, karteID, herkuenfte); fehler != nil {
				return fehler
			}

			// **Die Bildangaben nach der Feldherkunft** (A12, #1451): ihr
			// Fremdschlüssel zeigt auf sie, und eine Angabe ohne ihre Herkunft
			// wäre eine Angabe über nichts.
			//
			// **Und in derselben Transaktion.** Ein Bild, dessen Güte in einem
			// zweiten Zug geschrieben würde, hätte dazwischen einen Stand ohne
			// sie — genau den, den A11 ausgeschlossen hat.
			for feld, je := range angaben {
				if fehler := bildangabenSetzen(
					ctx, tx, profilID, karteID, feld, je); fehler != nil {
					return fehler
				}
			}

			return nil
		})
}
