package storage_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// frischeDatenbank gibt eine frisch migrierte Datenbank im Arbeitsspeicher
// (AGENTS.md §3).
//
// Im Arbeitsspeicher und nicht in einer Datei: eine Datei bliebe zwischen den
// Läufen liegen, und ein Test, der auf den Rest des vorigen trifft, ist
// keiner.
//
// # **Nicht nebenläufig**, und der Grund steht in der Datenquelle
//
// Hier stand bis zum 2026-08-19 „je Test eine eigene" und „im Arbeitsspeicher,
// damit `t.Parallel()` nichts teilt". Beides ist falsch, und zwar aus
// demselben Grund: [storage.ImArbeitsspeicher] trägt `cache=shared`, sonst
// bekäme jede Verbindung des Pools ihre eigene leere Datenbank. Geteilt wird
// damit über den **ganzen Prozess** — es ist eine Datenbank, nicht eine je
// Test, und sie überlebt nur deshalb nicht, weil `Close` sie beim Schließen
// der letzten Verbindung fallen lässt.
//
// Ein `t.Parallel()` daneben ergibt darum
//
//	die Migration schlug fehl: table card_suche already exists
//
// — gesehen, nicht vermutet. Wer hier nebenläufig prüfen will, braucht eine
// Datenquelle je Test; das wäre eine Änderung an `Oeffnen` und keine am Test.
func frischeDatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db, err := storage.Oeffnen(t.Context(), storage.ImArbeitsspeicher)
	if err != nil {
		t.Fatalf("die Datenbank ließ sich nicht öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("die Datenbank ließ sich nicht schließen: %v", fehler)
		}
	})

	if err := storage.Migriere(t.Context(), db, nil); err != nil {
		t.Fatalf("die Migration schlug fehl: %v", err)
	}

	return db
}

// angelegtAm liest den Zeitstempel, den die erste Migration eingesetzt hat.
func angelegtAm(t *testing.T, db *sql.DB) string {
	t.Helper()

	var wert string
	if err := db.QueryRowContext(t.Context(),
		`SELECT wert FROM bestand WHERE schluessel = 'angelegt_am'`).Scan(&wert); err != nil {
		t.Fatalf("angelegt_am nicht lesbar: %v", err)
	}

	return wert
}

// frischeDateidatenbank gibt eine frisch migrierte Datenbank **in einer Datei**.
//
// # Warum es beide Vorrichtungen braucht
//
// `frischeDatenbank` liegt im Arbeitsspeicher, und dort hält `Oeffnen` den
// Vorrat bei **einer** Verbindung — sonst wäre die Datenbank mit der ersten
// geschlossenen fort. Genau das hat einen Fehler drei Phasen lang verdeckt: die
// Pragmas galten nur auf der einen Verbindung, und kein Test hatte eine zweite
// (#157).
//
// **Eine Zusage, die nur auf einer Verbindung gilt, sieht kein Test, der eine
// benutzt.** Wer etwas über mehrere prüfen will, nimmt diese hier.
//
// Sie liegt in `t.TempDir()`, also je Test in einem eigenen Verzeichnis — der
// Grund, aus dem `frischeDatenbank` keine Datei nimmt, gilt damit weiter.
func frischeDateidatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db, err := storage.Oeffnen(t.Context(), "file:"+filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatalf("die Datenbank ließ sich nicht öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("die Datenbank ließ sich nicht schließen: %v", fehler)
		}
	})

	if err := storage.Migriere(t.Context(), db, nil); err != nil {
		t.Fatalf("die Migration schlug fehl: %v", err)
	}

	return db
}

// probekontext gibt einen Kontext, der auf das **erste** Profil zeigt.
//
// # Warum jeder Speichertest ihn braucht
//
// **Seit Werkstück 2 von Phase 4.12 klemmt jede Wurzelabfrage auf ein Profil**
// (#1351).
// Ein Aufruf ohne eines ist ein Fehler und keine leere Antwort — das ist
// Absicht (§8): eine Abfrage mit `profil_id = 0` fände nichts, und das sähe
// aus wie ein leerer Bestand statt wie ein Fehler im Aufruf.
//
// **Die Eins ist keine Annahme, sondern die Wanderung.** `0034` legt einen
// Nutzer mit einem Profil an, und beide bekommen die erste Kennung.
func probekontext(t *testing.T) context.Context {
	t.Helper()

	return storage.MitProfil(t.Context(), 1)
}
