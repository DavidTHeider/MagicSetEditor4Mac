package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FeldSetzenMitVorwert schreibt einen Feldwert und gibt den vorigen zurück.
//
// # Warum in einer Transaktion und nicht in zwei Aufrufen
//
// **Weil zwei Aufrufer schreiben** ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
// Der MCP-Server las den Vorwert und schrieb danach; dazwischen konnte das
// Fenster denselben Wert ändern, und die zugesagte Rücknahme führte auf einen
// Stand, den es nie gab. Der Zweitblick über Phase 9 hat es gefunden.
//
// Der zweite Rückgabewert unterscheidet **„war leer"** von **„gab es nicht"** —
// dieselbe Unterscheidung, die `domain.Card.Value` trifft, und aus demselben
// Grund: wer sie zusammenwirft, kann eine Rücknahme nicht mehr ausführen.
func (k *Kartenschreiber) FeldSetzenMitVorwert(
	ctx context.Context, karteID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	vorher, hatte, _, fehler = k.inTransaktion(ctx, karteID, feld, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, feldSetzen, karteID, feld, wert, karteID); err != nil {
			return err
		}

		return nil
	})

	return vorher, hatte, fehler
}

// FelderSetzenMitVorwert schreibt **mehrere** Felder einer Karte in einer
// Transaktion und gibt den vorigen Wert des **ersten**.
//
// # Warum es das gibt
//
// Ein Feld mit `save value: false` geht in die Felder, aus denen es entsteht
// ([service.Feldweg], #696). Bis hierher schrieb der Weg Teil für Teil, und
// jeder Aufruf eröffnete seine **eigene** Transaktion: scheiterte Teil zwei,
// war Teil eins schon bestätigt. Bei `text` einer M15-Karte heißt das, der
// Regeltext steht neu da und der Flavortext alt — eine halb geänderte Karte,
// und niemand kann sagen, welche Hälfte. Angestrichen vom Zweitblick durch
// Codex am 2026-08-21
// (#714).
//
// # Was sie zusagt
//
// **Alles oder nichts** — und **ein** Verlaufsschritt für alle
// ([Kartenschreiber.inEinemSchritt], #2396): bis dahin war es je Feld einer,
// und ein ⌘Z nahm nur das letzte zurück. Der Vorwert des ersten geht an den
// Aufrufer, weil dessen Naht einen erwartet; die übrigen stehen im Verlauf.
//
// **Eine leere Reihe ist kein Fehler**: sie ändert nichts, und ein Aufrufer,
// der nichts zu schreiben hat, soll dafür keine Meldung bekommen.
func (k *Kartenschreiber) FelderSetzenMitVorwert(
	ctx context.Context, karteID int64, werte []domain.Value,
) (vorher string, hatte bool, fehler error) {
	return k.imEinenZug(ctx, karteID, werte, nil)
}

// imEinenZug schreibt die Werte und **danach, in derselben Transaktion**, was
// der Nachtrag hinzufügt — nil heißt: nichts.
//
// # Warum es den Nachtrag gibt
//
// **Weil eine Übernahme aus einem KI-Vorschlag mehr schreibt als Werte**
// (A11, #1450): sie schreibt auch, woher sie kommen, und beides gehört
// zusammen. Ein zweiter Ruf danach hätte eine eigene Transaktion — und
// zwischen ihnen einen Stand, in dem die Karte einen KI-Wert trägt, der
// nirgends als solcher steht.
//
// **Nach den Werten und nicht davor**: die Herkunft spricht über das, was
// geschrieben wurde; scheitert das Schreiben, gibt es nichts zu belegen.
func (k *Kartenschreiber) imEinenZug(
	ctx context.Context, karteID int64, werte []domain.Value,
	nachtrag func(*sql.Tx) error,
) (vorher string, hatte bool, fehler error) {
	return k.imEinenZugAufStand(ctx, karteID, werte, "", nachtrag)
}

// imEinenZugAufStand tut dasselbe und hält davor den Stand der Karte gegen
// `erwartet` — leer heißt: ohne Bedingung ([standHalten]).
func (k *Kartenschreiber) imEinenZugAufStand(
	ctx context.Context, karteID int64, werte []domain.Value,
	erwartet string, nachtrag func(*sql.Tx) error,
) (vorher string, hatte bool, fehler error) {
	if len(werte) == 0 {
		return "", false, nil
	}

	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, uebersetze(err, "%d Felder ändern", len(werte))
	}

	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	// **Die Karte gehört dem gezeigten Profil**, bevor irgendetwas an ihr
	// gelesen wird (#1815).
	if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
		return "", false, fehler
	}

	if fehler := standHalten(ctx, tx, karteID, erwartet); fehler != nil {
		return "", false, fehler
	}

	schritt := k.inEinemSchritt()

	for i := range werte {
		// **Was nachgezogen wird, steht nicht unter der Bedingung** —
		// Gesamtreview über 0.9.0, Befund 1.
		lauf := ctx
		if werte[i].Nachgezogen {
			lauf = ohneBekannte(ctx)
		}

		voriger, hatteVor, _, err := schritt.imLauf(lauf, tx, karteID, werte[i].Field,
			func(tx *sql.Tx) error {
				_, fehler := tx.ExecContext(ctx, feldSetzen,
					karteID, werte[i].Field, werte[i].Text, karteID)

				return fehler
			})
		if err != nil {
			return "", false, err
		}

		if i == 0 {
			vorher, hatte = voriger, hatteVor
		}
	}

	if nachtrag != nil {
		if err := nachtrag(tx); err != nil {
			return "", false, err
		}
	}

	if err := bestaetige(tx, "%d Felder ändern", len(werte)); err != nil {
		return "", false, err
	}

	return vorher, hatte, nil
}

// FeldSetzenAnStelle schreibt einen Feldwert **an eine gegebene Stelle**.
//
// # Wofür
//
// Damit ein Gegenüber eine Entfernung vollständig zurücknehmen kann
// (#160):
// `FeldEntfernen` nennt die Stelle, und dieser Weg setzt das Feld wieder
// dorthin. Ohne ihn käme es ans Ende, und aus einer zurückgenommenen Änderung
// würde eine bleibende — sichtbar erst beim nächsten Speichern.
//
// **Der Verlauf benutzt ihn nicht**; der geht seinen eigenen Weg über
// `wegGehen` und trägt die Stelle im Eintrag mit.
func (k *Kartenschreiber) FeldSetzenAnStelle(
	ctx context.Context, karteID int64, feld, wert string, stelle int64,
) (vorher string, hatte bool, fehler error) {
	vorher, hatte, _, fehler = k.inTransaktion(ctx, karteID, feld, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, feldSetzenAnStelle, karteID, feld, wert, stelle)

		return err
	})

	return vorher, hatte, fehler
}

// FeldEntfernen nimmt ein Feld von einer Karte und gibt seinen Wert zurück.
//
// **Der einzige Weg, eine Änderung zurückzunehmen, die ein Feld erst angelegt
// hat.** `FeldSetzen` kann es nur leeren, und eine leere Zeile ist nicht
// dasselbe wie keine Zeile.
// **Und die Stelle kommt mit** (#160): ohne sie kann ein Gegenüber die
// Entfernung nur halb zurücknehmen — das Feld käme ans Ende. `nil` heißt „gab
// es nicht", und die 0 ist die erste Stelle, nicht „keine".
func (k *Kartenschreiber) FeldEntfernen(
	ctx context.Context, karteID int64, feld string,
) (vorher string, hatte bool, stelle *int64, fehler error) {
	return k.inTransaktion(ctx, karteID, feld, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM card_values WHERE card_id = ? AND feld = ?`, karteID, feld)

		return err
	})
}

// verlaufen schreibt den Verlaufseintrag einer Feldänderung.
//
// **Der Nachwert wird gelesen, nicht angenommen.** Die Änderung ist schon
// ausgeführt; was jetzt dasteht, ist die Wahrheit — und bei `FeldEntfernen`
// steht dort nichts.
func (k *Kartenschreiber) verlaufen(
	ctx context.Context, tx *sql.Tx, karteID int64, feld, vorher string, hatte bool,
	alteStelle *int64, alteHerkunft string,
) error {
	var (
		neuerWert  sql.NullString
		neueStelle sql.NullInt64
	)

	err := tx.QueryRowContext(ctx,
		`SELECT wert, reihenfolge FROM card_values WHERE card_id = ? AND feld = ?`,
		karteID, feld).Scan(&neuerWert, &neueStelle)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return uebersetze(err, "den neuen Wert von %q lesen", feld)
	}

	return schreibeVerlauf(ctx, tx, karteID, feld,
		feldweg{
			Feld: feld, Wert: vorher, GabEs: hatte,
			Reihenfolge: alteStelle, Herkunft: alteHerkunft,
		},
		feldweg{
			Feld: feld, Wert: neuerWert.String, GabEs: neuerWert.Valid,
			Reihenfolge: stelleAus(neueStelle), Herkunft: k.herkunft,
		},
		k.urheber, k.gruppe, k.herkunft)
}

// stelleAus macht aus einer nullbaren Spalte einen Zeiger.
//
// **Kein `int64` mit 0 als „keine".** Die 0 ist eine gültige Stelle — die
// erste —, und wer sie als „keine" läse, setzte jedes erste Feld einer Karte
// ans Ende zurück.
func stelleAus(spalte sql.NullInt64) *int64 {
	if !spalte.Valid {
		return nil
	}

	stelle := spalte.Int64

	return &stelle
}

// inTransaktion liest den Vorwert und führt die Änderung in einem Zug aus.
func (k *Kartenschreiber) inTransaktion(
	ctx context.Context, karteID int64, feld string, aendern func(*sql.Tx) error,
) (vorher string, hatte bool, stelle *int64, fehler error) {
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, nil, uebersetze(err, "Feld %q ändern", feld)
	}

	// **Zurückrollen, wenn nicht bestätigt wurde.** Nach einem geglückten
	// `Commit` ist das ein `ErrTxDone` und keine Nachricht wert.
	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	// **Die Karte gehört dem gezeigten Profil** (#1815) — in derselben
	// Transaktion wie die Änderung.
	if besitzfehler := karteImProfil(ctx, tx, karteID); besitzfehler != nil {
		return "", false, nil, besitzfehler
	}

	vorher, hatte, stelle, fehler = k.imLauf(ctx, tx, karteID, feld, aendern)
	if fehler != nil {
		return "", false, nil, fehler
	}

	if err := bestaetige(tx, "Feld %q ändern", feld); err != nil {
		return "", false, nil, err
	}

	return vorher, hatte, stelle, nil
}

// bestaetige schließt eine Transaktion und übersetzt, was dabei schiefgeht.
//
// **Eine Stelle für beide Wege** ([Kartenschreiber.inTransaktion] und
// [Kartenschreiber.FelderSetzenMitVorwert]): zwei gleichlautende Zweige wären
// zwei, von denen nur einer je einen Test sähe (§3, #714).
func bestaetige(tx *sql.Tx, was string, werte ...any) error {
	if err := tx.Commit(); err != nil {
		return uebersetze(err, was, werte...)
	}

	return nil
}

// imLauf ist der Rumpf einer Feldänderung **innerhalb** einer Transaktion.
//
// **Getrennt vom Rahmen**, damit mehrere Felder in **einer** Transaktion
// geändert werden können ([Kartenschreiber.FelderSetzenMitVorwert], #714). Der
// Rahmen entscheidet, wann bestätigt wird; hier steht, was an einem Feld
// geschieht.
func (k *Kartenschreiber) imLauf(
	ctx context.Context, tx *sql.Tx, karteID int64, feld string,
	aendern func(*sql.Tx) error,
) (vorher string, hatte bool, stelle *int64, fehler error) {
	// **Auf dem Schreibweg und nicht an der Fassade** (#89): so gilt die Regel
	// für jeden Schreiber, auch für den, den es noch nicht gibt.
	//
	// **Und innerhalb der Transaktion, nicht davor** (#714). Vorgeschaltet
	// wäre sie billiger — und die Reihe bräche dann **vor** dem ersten
	// Schreiben ab, sodass kein Lauf je die Rücknahme beträte. Ein
	// Fehlerzweig, den kein Test erreicht, ist eine Behauptung darüber, was im
	// Ernstfall geschieht (§3).
	if befund := domain.PruefeFeldname(feld); befund != nil {
		return "", false, nil, befund
	}

	// **Eine Abfrage für beides.** Ob es die Karte gibt und was im Feld stand,
	// sind zwei Fragen an denselben Stand — zwei Abfragen wären zwei Stände,
	// und der Unterschied fiele nur unter Last auf.
	//
	// Beide Spalten sind `NULL`-fähig: `karteDa` ist leer, wenn es die Karte
	// nicht gibt, `alterWert`, wenn das Feld fehlt. Genau diese Unterscheidung
	// braucht die Rücknahme.
	//
	// **Die Reihenfolge kommt mit** (#160): eine zurückgenommene Entfernung
	// setzt das Feld sonst ans Ende, und aus einer zurückgenommenen Änderung
	// wird eine bleibende.
	var (
		karteDa    sql.NullInt64
		alterWert  sql.NullString
		alteStelle sql.NullInt64
	)

	err := tx.QueryRowContext(ctx, `
		SELECT (SELECT 1 FROM cards WHERE id = ?),
		       (SELECT wert FROM card_values WHERE card_id = ? AND feld = ?),
		       (SELECT reihenfolge FROM card_values WHERE card_id = ? AND feld = ?)`,
		karteID, karteID, feld, karteID, feld).Scan(&karteDa, &alterWert, &alteStelle)
	if err != nil {
		return "", false, nil, uebersetze(err, "den vorigen Wert von %q lesen", feld)
	}

	if !karteDa.Valid {
		return "", false, nil, fmt.Errorf("%w: die Karte %d liegt nicht im Bestand",
			domain.ErrNotFound, karteID)
	}

	vorher, hatte, stelle = alterWert.String, alterWert.Valid, stelleAus(alteStelle)

	if fremd := fremderWert(ctx, feld, vorher); fremd != nil {
		return "", false, nil, fremd
	}

	// **Die vorige Herkunft wird vor der Änderung gelesen** (#1036). Danach
	// hat `herkunftFuehren` sie überschrieben, und der Verlaufseintrag könnte
	// sie nicht mehr tragen — die Rücknahme stellte dann den Wert wieder her
	// und seine Nennung nicht.
	//
	// **Eine eigene Abfrage und keine Spalte in der Vorprüfung.** Die trägt
	// drei Spalten, und jede Erweiterung zieht die Zeilengestalten des
	// Stolpertreibers mit — derselbe Fehler ist in Phase 6 zweimal passiert.
	alteHerkunft, herkunftfehler := k.herkunftLesen(ctx, tx, karteID, feld)
	if herkunftfehler != nil {
		return "", false, nil, herkunftfehler
	}

	if err := aendern(tx); err != nil {
		return "", false, nil, uebersetze(err, "Feld %q ändern", feld)
	}

	// **Und die Herkunft ebenso** (#1025): sie sagt, woher der Wert stammt, der
	// jetzt dasteht. Eine Zeile, die nach der Änderung geschrieben würde, fehlte
	// genau dann, wenn jemand dazwischen abstürzt — und dann stünde ein Bild von
	// Scryfall ohne seine Nennung da.
	if err := k.herkunftFuehren(ctx, tx, karteID, feld); err != nil {
		return "", false, nil, err
	}

	// **Der Verlauf entsteht in derselben Transaktion** (#52): ein Eintrag, der
	// danach geschrieben wird, fehlt genau dann, wenn jemand dazwischen
	// abstürzt — und dann ist die Änderung da und die Rücknahme fort.
	if err := k.verlaufen(
		ctx, tx, karteID, feld, vorher, hatte, stelle, alteHerkunft); err != nil {
		return "", false, nil, err
	}

	return vorher, hatte, stelle, nil
}

// verwerfen rollt eine Transaktion zurück, die nicht bestätigt wurde.

// FelderSetzenVonAussen schreibt mehrere Felder und hält die Herkunft fest.
//
// **Die Herkunft gehört zur Änderung**, nicht zum Schreiber — dasselbe Fenster
// schreibt getippte Werte und übernommene, im Wechsel (#888). Leer heißt: von
// hier.
//
// **`erwartet` ist der Stand, aus dem die Werte gerechnet sind** — leer heißt
// ohne Bedingung, geprüft in derselben Transaktion ([standHalten]). Seit #2396
// hängt ein nachgezogener Wert am Rest der Karte, und ein Schreiber dazwischen
// ließ ihn zu keinem Stand passen (Gesamtreview über 0.9.0, Befund 2).
func (k *Kartenschreiber) FelderSetzenVonAussen(
	ctx context.Context, karteID int64, werte []domain.Value, herkunft, erwartet string,
) (vorher string, hatte bool, fehler error) {
	return k.Von(herkunft).imEinenZugAufStand(ctx, karteID, werte, erwartet, nil)
}
