package storage

// **Woher ein übernommener Feldwert kommt** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Warum das Schreiben nicht hier steht
//
// **Weil es in der Transaktion des Feldschreibers liegt.** Das Issue:
// „Scheitert Provenienz/Verlauf in derselben Transaktion, stehen keine halben
// Inhaltsänderungen." Eine eigene Ablage mit eigener Verbindung hätte eine
// eigene Transaktion — und genau das wäre der halbe Stand.
//
// Was hier steht, ist das **Lesen** und der Satz SQL, den der Feldschreiber
// in seiner Transaktion ausführt ([herkunftSchreiben]).

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Feldherkunft liest, woher die Werte einer Karte stammen.
type Feldherkunft struct{ db *sql.DB }

// NeueFeldherkunft baut sie über einer Datenbank.
func NeueFeldherkunft(db *sql.DB) *Feldherkunft { return &Feldherkunft{db: db} }

// herkunftSchreiben legt eine Zeile an — und ersetzt die vorige.
//
// **Ersetzen und nicht anhängen**: am Feld steht ein Wert und nicht zwei, und
// zwei Herkünfte nebeneinander ließen offen, welche gilt.
const herkunftSchreiben = `
	INSERT INTO ki_feldherkunft
		(profil_id, karte_id, feld, anbieter, modell, aufgabe, auftrag,
		 zeitpunkt, bezug)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT (profil_id, karte_id, feld) DO UPDATE SET
		anbieter = excluded.anbieter, modell = excluded.modell,
		aufgabe = excluded.aufgabe, auftrag = excluded.auftrag,
		zeitpunkt = excluded.zeitpunkt, bezug = excluded.bezug`

// herkunftLesen gibt die Herkünfte einer Karte.
const herkunftLesen = `
	SELECT feld, anbieter, modell, aufgabe, auftrag, zeitpunkt, bezug
	FROM ki_feldherkunft
	WHERE profil_id = ? AND karte_id = ?
	ORDER BY feld`

// inTransaktion schreibt die Herkünfte in eine **bestehende** Transaktion.
//
// **Ohne eigene Verbindung**, und das ist der ganze Punkt: der Aufrufer hält
// die Transaktion, in der auch die Werte stehen.
//
// **Ein doppelt genanntes Feld ist ein Fehler und keine Wiederholung.** Es
// hieße, dass eine Übernahme demselben Feld zwei Werte gäbe — und welcher
// gälte, entschiede die Reihenfolge.
func herkuenfteSchreiben(
	ctx context.Context, tx *sql.Tx, profilID, karteID int64,
	herkuenfte []domain.KIFeldherkunft,
) error {
	gesehen := make(map[string]bool, len(herkuenfte))

	for _, herkunft := range herkuenfte {
		if befunde := herkunft.Validate(); len(befunde) > 0 {
			return fmt.Errorf("%w: die Herkunft von %q ist unvollständig: %s",
				domain.ErrValidation, herkunft.Feld, befunde[0].Error())
		}

		if gesehen[herkunft.Feld] {
			return fmt.Errorf(
				"%w: das Feld %q steht zweimal in der Herkunft — eine "+
					"Übernahme gibt ihm einen Wert und nicht zwei",
				domain.ErrValidation, herkunft.Feld)
		}

		gesehen[herkunft.Feld] = true

		if _, fehler := tx.ExecContext(ctx, herkunftSchreiben,
			profilID, karteID, herkunft.Feld, herkunft.Anbieter,
			herkunft.Modell, herkunft.Aufgabe, herkunft.Auftrag,
			herkunft.Zeitpunkt.UTC().Format(time.RFC3339Nano),
			herkunft.Bezug); fehler != nil {
			return uebersetze(fehler, "die Herkunft von %q schreiben",
				herkunft.Feld)
		}
	}

	return nil
}

// Lesen gibt die Herkünfte einer Karte — sortiert nach Feldname.
func (h *Feldherkunft) Lesen(ctx context.Context, karteID int64) (
	[]domain.KIFeldherkunft, error,
) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, fehler := h.db.QueryContext(ctx, herkunftLesen, profilID, karteID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Herkunft von Karte %d lesen", karteID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird oben

	gelesen := make([]domain.KIFeldherkunft, 0)

	for zeilen.Next() {
		herkunft, fehler := eineHerkunftAus(zeilen)
		if fehler != nil {
			return nil, fehler
		}

		gelesen = append(gelesen, herkunft)
	}

	// **Der Iterationsfehler steht im Rückgabewert und nicht in einem `if`**
	// — dieselbe Gestalt wie in `anbieterkonten.go`. Ein Zweig, den kein
	// Lauf betreten kann, sähe aus wie eine geprüfte Zusage und wäre keine
	// (§10).
	return gelesen, uebersetze(zeilen.Err(),
		"die Herkunft von Karte %d lesen", karteID)
}

// eineHerkunftAus liest eine Zeile.
func eineHerkunftAus(zeilen *sql.Rows) (domain.KIFeldherkunft, error) {
	var (
		herkunft  domain.KIFeldherkunft
		zeitpunkt string
	)

	if fehler := zeilen.Scan(&herkunft.Feld, &herkunft.Anbieter,
		&herkunft.Modell, &herkunft.Aufgabe, &herkunft.Auftrag,
		&zeitpunkt, &herkunft.Bezug); fehler != nil {
		return domain.KIFeldherkunft{}, uebersetze(fehler,
			"eine Zeile der Herkunft lesen")
	}

	// **Als Zeitpunkt und nicht als Zeichenkette.** Wer sie roh weiterreichte,
	// ließe jede Fläche selbst deuten — und eine von ihnen anders.
	gelesen, fehler := time.Parse(time.RFC3339Nano, zeitpunkt)
	if fehler != nil {
		return domain.KIFeldherkunft{}, domain.MitGrund(domain.ErrValidation,
			fehler, "den Zeitpunkt der Herkunft von %q lesen", herkunft.Feld)
	}

	herkunft.Zeitpunkt = gelesen

	return herkunft, nil
}
