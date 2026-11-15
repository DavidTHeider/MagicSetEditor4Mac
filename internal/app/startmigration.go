package app

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Fassungswache gibt die Freigabe, die ein Dienst oder Werkzeugserver vor
// jedem Ruf holt: kennt dieser Prozess den Bestand noch, und wandert
// niemand, solange er ruft (`fassungswache.go`, M02, #2346)? **Hier und nicht dort**, weil hier
// der Root steht, der `storage` kennen darf (§2).
func Fassungswache(db *sql.DB, beenden func()) func(context.Context) (func(), error) {
	return fassungswacheMit(func(ctx context.Context) (func(), error) { return storage.HalteSchemafreigabe(ctx, db) }, beenden)
}

// migriere ist die Wanderung beider Wege, auf denen die Anwendung einen
// Bestand öffnet — beim Start und nach dem Zurückspielen.
//
// **Eine Naht für den Lauf eines abbrechenden Upgrades**
// (`abbrechendesupgrade_intern_test.go`): seit es nur die Grundlinie gibt
// (Vermerk 0189), steht im Betrieb keine Wanderung aus, und der Lauf bringt
// eine mit, die scheitert. Im Betrieb ist sie [storage.Migriere].
var migriere = storage.Migriere

// aufDenHeutigenStand migriert den Bestand und liest danach das Profil.
//
// # Der Rückweg gehört dazu
//
// **Ab hier gehört jeder Rückweg über `abbrechen`.** Der Nachlauf des
// Zweitblicks hat es gefunden: `a.db` war gesetzt, und ein Fehler beim
// Migrieren kehrte unmittelbar zurück. In `main.lauf` wird das schließende
// `defer` erst **nach** einem geglückten Start eingerichtet — also blieb genau
// auf den neu eingeführten Fehlerpfaden eine offene Datenbank zurück. Eine
// Korrektur, die einen Weg baut, baut auch seinen Rückweg.
//
// # Und vorher wird gesichert
//
// (#939) — aber nur, wenn etwas aussteht.
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// sagt „vorwärts immer, rückwärts nie" zu, und der zweite Satz ist nur
// vertretbar, wenn es einen Weg zurück gibt.
//
// # Die Gegenprobe steht in [storage.Migriere]
//
// Sie lief bis zum 2026-08-23 hier ein zweites Mal. Sie gehört dorthin, weil
// sie dort **innerhalb** der Transaktion läuft: weist sie ab, ist die Migration
// zurückgenommen statt festgeschrieben
// (#938). Und
// sie läuft dort bei **jedem** Start, auch wenn nichts zu migrieren war — damit
// wäre ein Aufruf hier ein Zweig, den kein Aufruf mehr betreten kann (§3).
func (a *App) aufDenHeutigenStand(ctx context.Context, db *sql.DB) error {
	// **Mit dem Wärter des Fensters** — er ruft `launchctl` über den
	// Ausführer, den ein Lauf einsetzen kann.
	sichern := mitDienstanhalt(sicherungVorMigrationAllein(db, filepath.Dir(a.datenbankpfad), Fassung),
		func(ctx context.Context) error { return a.dienstwaerter().anhalten(ctx, a.datenbankpfad) })

	if fehler := migriere(ctx, db, sichern); fehler != nil {
		return a.abbrechen(fmt.Errorf("migrieren: %w", fehler))
	}

	// **Erst nach der Migration** — `profile` entsteht in `0034`.
	a.profilLaden(ctx)

	return nil
}

// SicherungVorMigration gibt die Vorbereitung für `storage.Migriere` — **den
// Dienst anhalten, dann sichern**.
//
// **Exportiert, weil es mehrere Composition Roots gibt** (§2). Das Fenster
// hatte sie seit #939,
// `cmd/mse-mcp` nicht — es rief `storage.Migriere(ctx, db, nil)`, und der
// Migrator überspringt bei `nil` ausdrücklich die Vorbereitung.
//
// **Wer zuerst startet, migriert.** Ein Rechner, auf dem der MCP-Server vor
// dem Fenster läuft, migrierte damit denselben produktiven Bestand **ohne**
// Sicherung; das Fenster fand danach nichts mehr zu tun und holte sie nicht
// nach. Angestrichen vom Zweitblick über Phase S. **Und wer migriert, hält
// den eingerichteten Dienst vorher an** — an jeder Wurzel, nicht nur am
// Fenster (Zweitblick über Phase 2 von 0.9.0, Befund 5).
//
// `ort` ist das Verzeichnis neben dem Bestand, `fassung` die Fassung, auf die
// migriert wird.
func SicherungVorMigration(db *sql.DB, ort, fassung string) func(ctx context.Context) error {
	waerter := dienstwaerter{befehl: befehlAusfuehren}

	return mitDienstanhalt(sicherungVorMigrationAllein(db, ort, fassung), func(ctx context.Context) error {
		pfad, fehler := storage.Bestandsdatei(ctx, db)
		if fehler != nil {
			return fehler
		}

		return waerter.anhalten(ctx, pfad)
	})
}

// sicherungVorMigrationAllein ist die Sicherung ohne den Wärter davor.
func sicherungVorMigrationAllein(db *sql.DB, ort, fassung string) func(ctx context.Context) error {
	return sicherungVorMigration(
		sichernUeber(db),
		sicherungTaugt,
		func(ctx context.Context, _ string) (string, error) {
			return storage.BestandsmarkeVon(ctx, db)
		},
		storage.Bestandsmarke, ort, fassung)
}
