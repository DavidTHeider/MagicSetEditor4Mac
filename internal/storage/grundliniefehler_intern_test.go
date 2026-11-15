package storage

// **Jeder Fehlerweg der Grundlinie wird betreten** (§3) — J02 (#2266).

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// probebasis ist eine Grundlinie aus Anweisungen, die nichts anlegen.
func probebasis(ohne string) fstest.MapFS {
	basis := fstest.MapFS{
		"grundlinie/tabellen-01.sql": {Data: []byte("SELECT 1;")},
		"grundlinie/saat.sql":        {Data: []byte("SELECT 1;")},
		"grundlinie/rest-01.sql":     {Data: []byte("SELECT 1;")},
		"grundlinie/umfang.txt":      {Data: []byte("# Kopf\n0001_probe.sql abc\n")},
	}
	delete(basis, ohne)

	return basis
}

func probetransaktion(t *testing.T) (tx *sql.Tx, fertig func()) {
	t.Helper()

	db := leererBestand(t)

	tx, fehler := db.BeginTx(t.Context(), nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return tx, func() { _ = tx.Rollback() } //nolint:errcheck // Aufräumen
}

func TestDieGrundlinieMeldetJedenFehler(t *testing.T) {
	for name, basis := range map[string]fstest.MapFS{
		"eine fehlende Saat":            probebasis("grundlinie/saat.sql"),
		"eine Anweisung, die scheitert": {"grundlinie/saat.sql": {Data: []byte("KEIN SQL;")}},
		"eine fehlende Buchführung":     probebasis(""),
	} {
		t.Run(name, func(t *testing.T) {
			tx, fertig := probetransaktion(t)
			defer fertig()

			if _, fehler := grundlinieAnwenden(t.Context(), tx, basis,
				[]string{"0001_probe.sql"}, map[string]string{"0001_probe.sql": "abc"}); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// ohneDateien ist eine Quelle ohne eine einzige Wanderung als Datei — wie die
// eingebettete, solange keine nach der Grundlinie dazukommt.
var ohneDateien = fstest.MapFS{"migrationen": {Mode: fs.ModeDir}}

func TestEinUnlesbarerUmfangWirdAbgewiesen(t *testing.T) {
	for name, basis := range map[string]fstest.MapFS{
		"ohne Umfang":                probebasis("grundlinie/umfang.txt"),
		"mit einer Zeile ohne Summe": {"grundlinie/umfang.txt": {Data: []byte("0001_probe.sql\n")}},
		"mit einem falschen Namen":   {"grundlinie/umfang.txt": {Data: []byte("probe abc\n")}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, fehler := grundlinienumfang(basis); !errors.Is(fehler, ErrMigrationen) {
				t.Errorf("der Fehler ist %v", fehler)
			}

			// **Und das Wandern endet, bevor es etwas anfasst.**
			db := leererBestand(t)
			if fehler := migriereAusMitVorbereitung(t.Context(), db, ohneDateien, nil, nil, basis); !errors.Is(fehler, ErrMigrationen) {
				t.Errorf("das Wandern meldete %v", fehler)
			}
		})
	}
}

// **Eine leere Summe gilt** — sie stammt aus einem Bestand von vor der
// Prüfsumme und wird übergangen, wie bei jeder Wanderung (`summeVergleichen`).
func TestEineLeereSummeImBestandGilt(t *testing.T) {
	if fehler := grundlinieGefuehrt([]string{"0001_probe.sql"},
		map[string]string{"0001_probe.sql": "abc"}, map[string]string{"0001_probe.sql": ""}); fehler != nil {
		t.Errorf("eine leere Summe wurde abgewiesen: %v", fehler)
	}
}

// **Ein abgebrochener Lauf ist kein leerer Bestand.**
func TestEinAbgebrochenerLaufFragtNichtsAnDenBestand(t *testing.T) {
	tx, fertig := probetransaktion(t)
	defer fertig()

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := grundlinieWennLeer(abgebrochen, tx, probebasis(""), nil, nil); fehler == nil {
		t.Error("ein abgebrochener Lauf meldete keinen Fehler")
	}
}

// **Ein Fehler der Grundlinie bricht das ganze Wandern ab** — und lässt
// nichts Halbes stehen: die erste Tafel steht schon, als die Saat scheitert.
func TestEinFehlerDerGrundlinieBrichtDasWandernAb(t *testing.T) {
	db := leererBestand(t)

	// **Mit Buchführungstabelle**, wie `migriereAusMitVorbereitung` sie
	// vorher anlegt — sonst scheiterte schon das Lesen unter der Sperre.
	if fehler := buchfuehrungAnlegen(t.Context(), db); fehler != nil {
		t.Fatal(fehler)
	}

	halb := probebasis("")
	halb["grundlinie/tabellen-01.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE halb (x);")}
	halb["grundlinie/saat.sql"] = &fstest.MapFile{Data: []byte("KEIN SQL;")}

	if fehler := anwendenAlle(t.Context(), db, fstest.MapFS{}, nil, PruefeBestand,
		halb, []string{"0001_probe.sql"}, map[string]string{"0001_probe.sql": "abc"}); fehler == nil {
		t.Error("eine scheiternde Saat meldete keinen Fehler")
	}

	if leer, _ := bestandIstLeerAus(t, db); !leer { //nolint:errcheck // gefragt ist die Leere
		t.Error("nach dem Fehler steht etwas Halbes da")
	}
}

// **Unter der Sperre wird noch einmal gefragt**, ob der Bestand leer ist —
// draußen ist er es schon; dieser Weg betrifft, wer dazwischen fremde
// Tafeln anlegt.
func TestUnterDerSperreWeistEineFremdeDateiAb(t *testing.T) {
	tx, fertig := probetransaktion(t)
	defer fertig()

	if _, fehler := tx.ExecContext(t.Context(), "CREATE TABLE fremd (x)"); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := grundlinieWennLeer(t.Context(), tx, probebasis(""), nil, nil); !errors.Is(fehler, ErrMigrationen) {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Scheitert das Anlegen der Buchführung, wird nicht gewandert** — etwa an
// einer Verbindung, die nur lesen darf.
func TestOhneAnlegbareBuchfuehrungWirdNichtGewandert(t *testing.T) {
	db, fehler := FuerLaeufe("file:" + filepath.Join(t.TempDir(), "leer.db") + "?_pragma=query_only(1)")
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if fehler := Migriere(t.Context(), db, nil); fehler == nil {
		t.Error("eine Verbindung, die nur lesen darf, wurde migriert")
	}
}

func bestandIstLeerAus(t *testing.T, db *sql.DB) (bool, error) {
	t.Helper()

	tx, fehler := db.BeginTx(t.Context(), nil)
	if fehler != nil {
		return false, fehler
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nur gelesen

	return bestandIstLeer(t.Context(), tx)
}

// **Tabellen ohne Buchführung sind kein Bestand dieser Anwendung** — und
// werden nicht still als einer genommen. Bis 0.7.0 scheiterte ein solcher an
// der ersten Wanderung, die eine seiner Tabellen noch einmal anlegen wollte;
// mit der Grundlinie bliebe er sonst einfach liegen, und der erste fachliche
// Aufruf fände seine Tabellen nicht.
func TestTabellenOhneBuchfuehrungWerdenAbgewiesen(t *testing.T) {
	db := leererBestand(t)

	if _, fehler := db.ExecContext(t.Context(), `CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL);
		INSERT INTO bestand VALUES ('angelegt_am', '2026-08-23T00:00:00.000Z')`); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := Migriere(t.Context(), db, nil); !errors.Is(fehler, ErrMigrationen) {
		t.Errorf("der Fehler ist %v, erwartet ErrMigrationen", fehler)
	}
}
