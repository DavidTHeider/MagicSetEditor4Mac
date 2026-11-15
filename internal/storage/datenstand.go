package storage

import (
	"context"
	"database/sql"
)

// Standleser liest den Datenstand einer Datenbank.
//
// **Als Typ und nicht als freie Funktion**, damit `app` ihn als Naht halten
// kann: dort darf `storage` nur der Composition Root kennen (§2).
//
// # Warum er keine Verbindung mehr festhält
//
// **`data_version` ist eine Zahl je Verbindung.** Bis zum 2026-09-09 lieh
// sich der Leser deshalb eine aus dem Vorrat und behielt sie — damit zwei
// Fragen dieselbe Zahl vergleichen. Der Preis stand nicht daneben: jedes
// Schreiben des Fensters lief auf einer **anderen** Verbindung desselben
// Vorrats und hob die Zahl der festgehaltenen. Das Fenster hielt sich für
// einen fremden Prozess und lud nach jedem Schreiben die Seite neu
// (#1485).
//
// Seitdem hat jede Datenbank **eine** Verbindung (`oeffnenMit`), und die
// Frage geht über den Vorrat wie das Schreiben: eigene Commits bewegen die
// Zahl nicht, die eines anderen Prozesses schon. Tauscht der Vorrat die
// Verbindung aus — nach einem Fehler, sonst nie, denn eine Lebensdauer ist
// nicht gesetzt —, beginnt die Zahl neu, und das Fenster lädt einmal zu
// viel: genau nach einem Fehler, bei dem ohnehin etwas nachzusehen ist.
type Standleser struct {
	db *sql.DB
}

// NeuerStandleser baut den Leser über einer Datenbank.
func NeuerStandleser(db *sql.DB) *Standleser { return &Standleser{db: db} }

// Datenstand gibt den Stand dieser Datenbank.
func (s *Standleser) Datenstand(ctx context.Context) (int64, error) {
	return standAus(ctx, s.db)
}

// frager ist alles, was eine Zeile hergibt — Vorrat wie einzelne Verbindung.
type frager interface {
	QueryRowContext(ctx context.Context, abfrage string, werte ...any) *sql.Row
}

// standAus liest `data_version` über einen Frager.
func standAus(ctx context.Context, wo frager) (int64, error) {
	var stand int64

	if err := wo.QueryRowContext(ctx, `PRAGMA data_version`).Scan(&stand); err != nil {
		return 0, uebersetze(err, "den Datenstand lesen")
	}

	return stand, nil
}

// Datenstand gibt SQLites `data_version` — die Zahl, die steigt, sobald eine
// **andere** Verbindung bestätigt hat.
//
// # Wofür sie da ist
//
// Der MCP-Server ist ein eigener Prozess auf derselben Datei
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). Ohne
// diese Zahl erfährt das offene Fenster nichts von seinen Änderungen
// (#156).
//
// # Was sie kostet
//
// **Nichts, was auf die Platte ginge.** `PRAGMA data_version` liest einen
// Zähler aus dem Kopf der Datei, den SQLite ohnehin führt; es ist die
// billigste Frage, die man einer Datenbank stellen kann. Deshalb darf sie oft
// gestellt werden — der Vermerk
// ([0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md))
// sagt, wie oft und warum nicht öfter.
//
// # Was sie nicht sagt
//
// **Was sich geändert hat.** Nur dass. Wer wissen will, welche Karte, liest
// nach — und genau das ist der Preis, den der Vermerk benennt.
func Datenstand(ctx context.Context, db *sql.DB) (int64, error) {
	return standAus(ctx, db)
}
