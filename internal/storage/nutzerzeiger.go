// **Wer arbeitet, steht im Bestand** — Werkstück 17 von Phase 4.12
// (#1400),
// entschieden in
// [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).
//
// # Warum das eine eigene Datei ist
//
// **Weil es eine eigene Frage ist** (§2): nebenan steht, welche Nutzer und
// Profile es gibt; hier steht, **wer von ihnen gerade arbeitet**. Bis zu
// diesem Werkstück gab es die Frage nicht — sie wurde aus dem gezeigten Profil
// beantwortet, und genau das war der Fehler.

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nutzerzeigerschluessel ist der Eintrag im Bestand, der auf den arbeitenden
// Menschen zeigt.
//
// **Neben [Profilzeigerschluessel] und nicht in ihm.** Zwei Fragen, zwei
// Einträge: *wer* arbeitet, und *worin*. Sie ändern sich unabhängig
// voneinander — das ist der ganze Inhalt von Werkstück 17.
const Nutzerzeigerschluessel = "nutzer_zeiger"

// Nutzerzeiger sagt, wer gerade arbeitet.
//
// # Der Rückfall ist das Verhalten von vorher
//
// **Ohne Eintrag gilt der Eigner des gezeigten Profils.** Jeder Bestand, der
// vor diesem Werkstück entstand, hat keinen — und beim ersten Start soll sich
// nichts ändern, nicht einmal für einen Augenblick.
//
// **Und ein Zeiger auf einen Nutzer, den es nicht mehr gibt, fällt ebenso
// zurück.** Derselbe Grund wie beim Profilzeiger: die Kaskade räumt beim
// Löschen eines Nutzers seine Profile, und ein Zeiger ins Leere ließe jede
// Abfrage ohne Identität dastehen.
//
// # Ein Zug und nicht zwei
//
// **Weil zwei Abfragen zwei Stände gäben**, zwischen denen sich etwas ändern
// kann — derselbe Grund, den `Profilzeiger` nebenan nennt.
func (a *Nutzerablage) Nutzerzeiger(ctx context.Context) (int64, error) {
	var arbeitet int64

	err := a.db.QueryRowContext(ctx,
		`SELECT coalesce(
		          (SELECT n.id FROM nutzer n
		            JOIN bestand b ON b.wert = CAST(n.id AS TEXT)
		           WHERE b.schluessel = ?),
		          (SELECT p.nutzer_id FROM profile p
		            JOIN bestand b ON b.wert = CAST(p.id AS TEXT)
		           WHERE b.schluessel = ?),
		          (SELECT nutzer_id FROM profile ORDER BY id LIMIT 1),
		          0)`,
		Nutzerzeigerschluessel, Profilzeigerschluessel).Scan(&arbeitet)
	if err != nil {
		return 0, uebersetze(err, "den arbeitenden Nutzer lesen")
	}

	return arbeitet, nil
}

// NutzerzeigerSetzen stellt um, wer arbeitet.
//
// **Den Nutzer gibt es, oder es ist ein Nein** — dieselbe Bauform und derselbe
// Grund wie bei [Nutzerablage.ProfilzeigerSetzen].
func (a *Nutzerablage) NutzerzeigerSetzen(ctx context.Context, nutzerID int64) error {
	var da int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM nutzer WHERE id = ?`, nutzerID).Scan(&da); err != nil {
		return uebersetze(err, "den Nutzer suchen")
	}

	if da == 0 {
		return fmt.Errorf(
			"%w: den Nutzer %d führt der Bestand nicht", domain.ErrNotFound, nutzerID)
	}

	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO bestand (schluessel, wert) VALUES (?, ?)
		 ON CONFLICT (schluessel) DO UPDATE SET wert = excluded.wert`,
		Nutzerzeigerschluessel, strconv.FormatInt(nutzerID, 10)); err != nil {
		return uebersetze(err, "den arbeitenden Nutzer setzen")
	}

	return nil
}

// ZeigerSetzen stellt **beide** Zeiger in einer Transaktion um.
//
// # Der Befund
//
// **Zweitblick über die Phasen 4.0 bis 4.13** (2026-09-07): `NutzerWechseln`
// rief `NutzerzeigerSetzen` und danach `ProfilzeigerSetzen`. Beide führen
// eigene Anweisungen aus; scheiterte die zweite, war die erste durch — dann
// **arbeitete ein anderer Mensch, während das Profil des alten gezeigt wurde**,
// und der Ruf meldete dazu einen Fehler.
//
// **Das ist kein Wettlauf, sondern ein halber Stand.** Er bleibt liegen, bis
// jemand erneut wechselt.
//
// # Warum beide Prüfungen mit hineingehören
//
// **Weil ein Nein nach dem halben Schreiben dasselbe hinterließe.** Nutzer und
// Profil werden in derselben Transaktion gesucht wie geschrieben; wer nicht
// da ist, lässt nichts zurück.
func (a *Nutzerablage) ZeigerSetzen(ctx context.Context, nutzerID, profilID int64) error {
	return imGeschaeft(ctx, a.db, "die Zeiger umstellen", func(geschaeft *sql.Tx) error {
		return zeigerSetzenIn(ctx, geschaeft, nutzerID, profilID)
	})
}

// zeigerSetzenIn stellt beide Zeiger in einem laufenden Geschäft um.
func zeigerSetzenIn(ctx context.Context, tx *sql.Tx, nutzerID, profilID int64) error {
	for _, gesucht := range []struct {
		tabelle, was string
		id           int64
	}{
		{"nutzer", "den Nutzer", nutzerID},
		{"profile", "das Profil", profilID},
	} {
		if fehler := gibtEs(ctx, tx, gesucht.tabelle, gesucht.was, gesucht.id); fehler != nil {
			return fehler
		}
	}

	for _, zu := range []struct {
		schluessel string
		wert       int64
	}{
		{Nutzerzeigerschluessel, nutzerID},
		{Profilzeigerschluessel, profilID},
	} {
		if _, fehler := tx.ExecContext(ctx,
			`INSERT INTO bestand (schluessel, wert) VALUES (?, ?)
			 ON CONFLICT (schluessel) DO UPDATE SET wert = excluded.wert`,
			zu.schluessel, strconv.FormatInt(zu.wert, 10)); fehler != nil {
			return uebersetze(fehler, "die Zeiger setzen")
		}
	}

	return nil
}

// gibtEs weist ab, was der Bestand nicht führt.
func gibtEs(ctx context.Context, tx *sql.Tx, tabelle, was string, id int64) error {
	var da int

	// **Der Tabellenname wird eingesetzt, nicht gebunden** — SQLite nimmt
	// dafür keinen Platzhalter. Unbedenklich, weil er aus der festen Liste im
	// Aufrufer kommt und nie aus einer Eingabe.
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM `+tabelle+` WHERE id = ?`, id).Scan(&da); err != nil {
		return uebersetze(err, was+" suchen")
	}

	if da == 0 {
		return fmt.Errorf("%w: %s %d führt der Bestand nicht",
			domain.ErrNotFound, was, id)
	}

	return nil
}
