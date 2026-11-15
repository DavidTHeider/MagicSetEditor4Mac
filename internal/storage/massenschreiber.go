package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FelderSetzen schreibt einen Feldwert auf mehrere Karten.
//
// # Alles oder nichts
//
// **In einer Transaktion**, und die Prüfung, ob es alle Karten gibt, steht
// darin. Eine Massenänderung, die zur Hälfte durchgeht, hinterlässt einen
// Bestand, von dem niemand sagen kann, welche Hälfte — und die Rücknahme müsste
// raten.
//
// Gescheitert wird deshalb **vor** dem ersten Schreiben und mit allen fehlenden
// Karten auf einmal: wer zwanzig Karten anstößt, will nicht zwanzigmal
// nacheinander eine erfahren. Der Fremdschlüssel allein könnte das nicht — er
// meldet die erste und bricht ab.
//
// # Was hier kein Test belegt
//
// Dass nach einem Fehler **mitten** in der Schreibfolge nichts im Bestand steht.
// Der Zweitblick über Phase 5 hat es angestrichen, und es stimmt: die Prüfungen
// mit dem Stolpertreiber sehen nur die Rückgabe, denn dieser Treiber hat gar
// keinen Bestand, in den man nachsehen könnte.
//
// Mit einer echten SQLite-Datenbank ist der Fall nicht herzustellen: was die
// Vorprüfung in derselben Transaktion durchlässt, schreibt sich auch. Was den
// Bestand hält, ist deshalb nicht unser Code, sondern die Transaktion — und die
// wird durch das zurückgestellte `Rollback` in **jedem** Rückweg geschlossen.
// Eine Attrappe, die Zeilen führt, nur um diesen Satz zu wiederholen, prüfte
// die Attrappe.
//
// Was ein Test **belegt**: eine gescheiterte Vorprüfung lässt die vorhandenen
// Karten unverändert (`TestFelderSetzenNenntDieFehlendenKarten`).
//
// # Der Vorgang kommt mit heraus
//
// Der erste Rückgabewert ist die **Gruppenkennung** dieses Schrittes
// (#1393).
// Sie stand längst im Verlauf und wurde nur nie zurückgegeben; mit ihr nimmt
// [Verlaufsleser.VorgangZuruecknehmen] **diesen** Vorgang zurück und nicht den
// jüngsten des Sets. Am Fenster ist der jüngste richtig — über MCP,
// Kommandozeile und REST schreibt womöglich ein zweiter Prozess dazwischen.
//
// **Bei leerer Auswahl ist sie leer.** Es hat kein Vorgang stattgefunden, und
// eine Kennung dafür wäre eine, die nichts benennt.
func (k *Kartenschreiber) FelderSetzen(
	ctx context.Context, karten []int64, feld, wert string,
) (string, error) {
	if len(karten) == 0 {
		return "", nil
	}

	// Dieselbe Regel wie auf dem Einzelweg (#89) — und vor dem ersten
	// Schreiben, wie jede Vorprüfung dieser Transaktion.
	if befund := domain.PruefeFeldname(feld); befund != nil {
		return "", befund
	}

	transaktion, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return "", uebersetze(err, "Massenänderung beginnen")
	}

	defer func() { _ = transaktion.Rollback() }()

	if err := fehlendeMelden(ctx, transaktion, karten); err != nil {
		return "", err
	}

	// **Eine Gruppe für alle** (#52): die Massenänderung ist **ein** Schritt im
	// Verlauf. Ohne sie nähme ⌘Z danach die vorige Änderung zurück, während die
	// geänderten Karten stehen blieben — und mit einem Schritt je Karte müsste
	// jemand dreihundertmal drücken.
	gruppe := neueGruppe()

	for _, karte := range karten {
		if err := k.massenzeile(ctx, transaktion, karte, feld, wert, gruppe); err != nil {
			return "", err
		}
	}

	if err := transaktion.Commit(); err != nil {
		return "", uebersetze(err, "Massenänderung abschließen")
	}

	return gruppe, nil
}

// massenzeile ändert **eine** Karte und schreibt ihren Verlaufseintrag.
//
// **Der Vorwert wird vorher gelesen**, in derselben Transaktion: ohne ihn gäbe
// es keinen Weg zurück, und ein Verlauf ohne Rückweg ist ein Protokoll.
//
// Er wird je Karte gelesen und nicht in einer Abfrage für alle: die Karten
// stehen auf verschiedenen Werten, und ein gemeinsamer Vorwert wäre eine
// Behauptung, die für die meisten falsch ist.
func (k *Kartenschreiber) massenzeile(
	ctx context.Context, tx *sql.Tx, karte int64, feld, wert, gruppe string,
) error {
	var vorher sql.NullString

	err := tx.QueryRowContext(ctx,
		`SELECT wert FROM card_values WHERE card_id = ? AND feld = ?`, karte, feld).Scan(&vorher)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &domain.BulkError{
			CardIDs: []int64{karte},
			Err:     uebersetze(err, "den vorigen Wert von %q lesen", feld),
		}
	}

	if _, err := tx.ExecContext(ctx, feldSetzen, karte, feld, wert, karte); err != nil {
		return &domain.BulkError{
			CardIDs: []int64{karte},
			Err:     uebersetze(err, "Feld %q schreiben", feld),
		}
	}

	// **Der neue Wert steht fest** — anders als bei der Einzeländerung, die auch
	// entfernen kann. Ihn zurückzulesen wäre eine Abfrage je Karte für eine
	// Antwort, die schon dasteht.
	fehler := schreibeVerlauf(ctx, tx, karte, feld,
		feldweg{Feld: feld, Wert: vorher.String, GabEs: vorher.Valid},
		feldweg{Feld: feld, Wert: wert, GabEs: true},
		k.urheber, gruppe, k.herkunft)
	if fehler != nil {
		return &domain.BulkError{CardIDs: []int64{karte}, Err: fehler}
	}

	return nil
}

// fehlendeMelden gibt einen BulkError über alle Karten, die es nicht gibt.
func fehlendeMelden(ctx context.Context, transaktion *sql.Tx, karten []int64) error {
	// **Eine Karte eines fremden Profils fehlt hier** (#1815) — sie wird
	// genannt wie eine, die es nicht gibt, und nichts wird geschrieben.
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	zeilen, err := transaktion.QueryContext(ctx,
		`SELECT cards.id FROM cards JOIN sets ON sets.id = cards.set_id
		  WHERE sets.profil_id = ? AND cards.id IN (SELECT value FROM json_each(?))`,
		profilID, alsJSONFeld(karten))
	if err != nil {
		return uebersetze(err, "die Karten prüfen")
	}

	defer func() { _ = zeilen.Close() }()

	vorhanden := map[int64]bool{}

	for zeilen.Next() {
		var kennung int64
		if err := zeilen.Scan(&kennung); err != nil {
			return uebersetze(err, "eine Kartenkennung lesen")
		}

		vorhanden[kennung] = true
	}

	if err := zeilen.Err(); err != nil {
		return uebersetze(err, "die Karten prüfen")
	}

	var fehlend []int64

	for _, karte := range karten {
		if !vorhanden[karte] && !slices.Contains(fehlend, karte) {
			fehlend = append(fehlend, karte)
		}
	}

	if len(fehlend) == 0 {
		return nil
	}

	return &domain.BulkError{
		CardIDs: fehlend,
		Err:     fmt.Errorf("%w: eine Karte der Auswahl", domain.ErrNotFound),
	}
}
