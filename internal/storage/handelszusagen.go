package storage

// **Die Zusagen eines Handels** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum es zwei Fragen an dieselbe Tabelle gibt
//
// **Weil zwei verschiedene Leute sie stellen.** Der Schutz des Bestands fragt
// „was hängt an **dieser Zeile**" — er will wissen, ob eine Änderung etwas
// wegnähme, das zugesagt ist. Die Verbuchung fragt „was habe ich für **diesen
// Handel** festgehalten" — sie will wissen, woher die Stücke kommen.
//
// Wer die zweite aus der ersten baute, liefe über den ganzen Bestand: der
// Preis wüchse mit der Sammlung statt mit dem Handel, und beide Indizes der
// Tabelle stünden ungenutzt daneben.
//
// # Und warum der Ableser geteilt wird
//
// **Weil zwei Fassungen auseinanderlaufen** — und die, die niemand ändert,
// ist die im fremden Weg. Derselbe Satz steht schon an `bestandsfrager`.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zusagenabfrage ist die Spaltenliste beider Wege.
const zusagenabfrage = `SELECT vorgang, exemplar_id, revision, menge, fassung,
                               frist, uebergeben
                          FROM reservierungen`

// FuerVorgang gibt die Zusagen eines Handels — über alle Besitzzeilen.
func (a *Reservierungsablage) FuerVorgang(
	ctx context.Context, vorgang string,
) ([]domain.Reservierung, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		zusagenabfrage+` WHERE profil_id = ? AND vorgang = ? ORDER BY id ASC`,
		profilID, vorgang)
	if fehler != nil {
		return nil, fmt.Errorf("die Zusagen ließen sich nicht lesen: %w", fehler)
	}

	return zusagenLesen(zeilen)
}

// zusagenLesen liest einen Zeilensatz aus [zusagenabfrage].
//
// **Es schließt den Zeilensatz selbst.** Wer ihn dem Aufrufer überließe,
// hätte zwei Stellen, an denen es zu vergessen ist.
func zusagenLesen(zeilen *sql.Rows) ([]domain.Reservierung, error) {
	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler steht unten

	// **Eine leere Liste und kein `nil`** — „nichts festgehalten" ist eine
	// Antwort, und sie sieht anders aus als keine.
	zusagen := []domain.Reservierung{}

	for zeilen.Next() {
		var (
			zusage     domain.Reservierung
			frist      string
			uebergeben int
		)

		if fehler := zeilen.Scan(&zusage.Vorgang, &zusage.ExemplarID,
			&zusage.Revision, &zusage.Menge, &zusage.Fassung, &frist,
			&uebergeben,
		); fehler != nil {
			return nil, fmt.Errorf("eine Zusage ließ sich nicht lesen: %w",
				fehler)
		}

		gelesen, fehler := ausFrist(frist)
		if fehler != nil {
			return nil, fehler
		}

		zusage.Frist, zusage.Uebergeben = gelesen, uebergeben == 1
		zusagen = append(zusagen, zusage)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, fmt.Errorf("die Zusagen brachen ab: %w", fehler)
	}

	return zusagen, nil
}

// Uebergeben vermerkt, dass die Zusagen eines Vorgangs über den Tisch
// gegangen sind.
//
// # Warum je Vorgang und nicht je Zusage
//
// **Weil eine Übergabe ein Zeitpunkt des Handels ist und keiner einer
// Besitzzeile.** Wer drei Karten aus drei Zeilen hergibt, gibt sie in einem
// Zug her; ein Vermerk je Zeile ließe einen halben Zustand entstehen, den
// die Verbuchung als „es fehlen noch zwei" läse.
//
// # Und warum keine Zahl zurückkommt
//
// **Weil sie niemand läse** (#1643).
// Ein Vorgang ohne festgehaltene Exemplare ist kein Fehler — es gibt Handel,
// bei denen diese Seite nichts weggibt —, und wer wissen will, was
// festgehalten ist, fragt [Reservierungsablage.FuerVorgang]. Eine Zahl, die
// kein Aufrufer benutzt, ist ein Zweig, den kein Lauf betritt (§10).
func (a *Reservierungsablage) Uebergeben(
	ctx context.Context, vorgang string,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	return uebergebenInTx(ctx, a.db, profilID, vorgang)
}

// zusagenwirkungInTx führt aus, was ein Zustandswechsel bedeutet — in der
// Transaktion des Aufrufers.
//
// **Deshalb nimmt sie einen `*sql.Tx` und keinen Bestand**: Vermerk, neue
// Revision und Wirkung stehen zusammen oder gar nicht (Befund 3 des
// Nachlaufs zum Zweitblick über Phase 5.6).
func zusagenwirkungInTx(
	ctx context.Context, zug *sql.Tx, profilID int64, vorgang string,
	wirkung domain.Zusagenwirkungsart,
) error {
	switch wirkung {
	case domain.ZusagenUebergeben:
		return uebergebenInTx(ctx, zug, profilID, vorgang)

	case domain.ZusagenFreigeben:
		return freigebenInTx(ctx, zug, profilID, vorgang)
	}

	return nil
}

// uebergebenInTx vermerkt die Übergabe über einen beliebigen Ableser.
//
// **Eine Anweisung und nicht zwei**: dieselbe Zeile gilt für den Ruf von
// außen und für die Transaktion einer Zustellung — zwei Abschriften hätten
// zwei Fehlerzweige, von denen einer nie zu betreten wäre (§10).
func uebergebenInTx(
	ctx context.Context, wer bestandsfrager, profilID int64, vorgang string,
) error {
	if _, fehler := wer.ExecContext(ctx,
		`UPDATE reservierungen SET uebergeben = 1
		  WHERE profil_id = ? AND vorgang = ?`,
		profilID, vorgang); fehler != nil {
		return fmt.Errorf(
			"die Übergabe ließ sich nicht vermerken: %w", fehler)
	}

	return nil
}
