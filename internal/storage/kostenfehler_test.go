package storage_test

// **Die Fehlerwege des Kostenjournals** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Warum eine echte Datenbank, die kaputtgeht
//
// **Weil das ehrlicher ist als ein Treiber, der Nein sagt.** Eine Tabelle, die
// es nicht mehr gibt, und eine Spalte, in die eine Zahl nicht passt, sind
// Zustände, die eine echte Datei annehmen kann — nach einer halben Migration,
// nach einem Zurückspielen auf einen älteren Stand.
//
// **Und weil eine Buchführung, die still schweigt, die gefährlichste ist**:
// wer eine leere Antwort für „nichts gebucht" hält, bucht ein zweites Mal.

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// ohneTafeln nimmt dem Journal seine Tabellen.
func ohneTafeln(t *testing.T, db *sql.DB, tafeln ...string) {
	t.Helper()

	for _, tafel := range tafeln {
		if _, fehler := db.ExecContext(t.Context(), "DROP TABLE "+tafel); fehler != nil {
			t.Fatalf("%s wegnehmen: %v", tafel, fehler)
		}
	}
}

// **Jede Abfrage meldet ihren Fehler** — und keine gibt eine leere Antwort.
func TestJedeKostenabfrageMeldetIhrenFehler(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 100), domain.ZeitraumTag, probezeit); fehler != nil {
		t.Fatalf("reservieren: %v", fehler)
	}

	ohneTafeln(t, db, "kostenposten")

	// **Ohne die Postentafel scheitert jeder Weg, der sie liest** — und zwar
	// laut.
	if _, fehler := journal.Stand(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("der Stand kam ohne Postentafel durch")
	}

	if _, fehler := journal.Posten(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit, 5); fehler == nil {
		t.Error("die Posten kamen ohne Tafel durch")
	}

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("zwei", 100), domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("reserviert wurde ohne Postentafel")
	}

	if _, fehler := journal.Abschliessen(t.Context(), 1, "eins",
		euro(50), probezeit); fehler == nil {
		t.Error("abgeschlossen wurde ohne Postentafel")
	}

	// **Und ohne die Budgettafel ebenso.**
	ohneTafeln(t, db, "kostenbudgets")

	if _, fehler := journal.Budget(t.Context(), 1, "anbieter",
		domain.ZeitraumTag); fehler == nil {
		t.Error("das Budget kam ohne Tafel durch")
	}

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler == nil {
		t.Error("gesetzt wurde ohne Budgettafel")
	}
}

// **Die Summe scheitert für sich** — die Grenze stand schon fest.
//
// **Zwei Abfragen und nicht eine**: `passtInsBudget` liest erst die Grenze und
// zählt dann zusammen. Wer nur die erste prüft, hat den zweiten Weg nie
// betreten.
func TestDieSummeScheitertFuerSich(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	// **Die Postentafel bekommt eine Gestalt, die nicht passt.** Sie gibt es
	// noch — das Lesen eines Postens glückt —, aber die Summe scheitert am
	// Scan: `hoechstbetrag` trägt Text.
	ohneTafeln(t, db, "kostenposten")

	if _, fehler := db.ExecContext(t.Context(), `
		CREATE TABLE kostenposten (
			kennung TEXT PRIMARY KEY, profil_id INTEGER, konto TEXT, aufgabe TEXT,
			hoechstbetrag TEXT, betrag TEXT, waehrung TEXT, stand TEXT,
			grund TEXT, begonnen TEXT, beendet TEXT
		)`); fehler != nil {
		t.Fatalf("die krumme Tafel anlegen: %v", fehler)
	}

	if _, fehler := db.ExecContext(t.Context(), `
		INSERT INTO kostenposten VALUES
			('krumm', 1, 'anbieter', 'karte.text', 'keine Zahl', 'auch nicht',
			 'EUR', 'reserviert', '', '2026-09-15T00:00:00.000Z', '')`); fehler != nil {
		t.Fatalf("die krumme Zeile legen: %v", fehler)
	}

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("neu", 100), domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("die krumme Summe kam durch")
	}

	// **Dieselbe Zeile bricht auch die drei Leser** — jeder an seiner Stelle.
	if _, fehler := journal.Stand(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("der Stand las die krumme Zeile")
	}

	if _, fehler := journal.Posten(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit, 5); fehler == nil {
		t.Error("die Posten lasen die krumme Zeile")
	}

	if _, fehler := journal.Abschliessen(t.Context(), 1, "krumm",
		euro(50), probezeit); fehler == nil {
		t.Error("abgeschlossen wurde gegen die krumme Zeile")
	}
}

// **Ein unbekannter Stand ist ein Befund und kein stiller Nullwert** (§8).
//
// Er käme aus einer neueren Fassung — und ihn wegzulassen machte aus
// gebundenem Geld verfügbares.
func TestEinUnbekannterStandIstEinBefund(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	ohneTafeln(t, db, "kostenposten")

	if _, fehler := db.ExecContext(t.Context(), `
		CREATE TABLE kostenposten (
			kennung TEXT PRIMARY KEY, profil_id INTEGER, konto TEXT, aufgabe TEXT,
			hoechstbetrag INTEGER, betrag INTEGER, waehrung TEXT, stand TEXT,
			grund TEXT, begonnen TEXT, beendet TEXT
		)`); fehler != nil {
		t.Fatalf("die offene Tafel anlegen: %v", fehler)
	}

	if _, fehler := db.ExecContext(t.Context(), `
		INSERT INTO kostenposten VALUES
			('neu', 1, 'anbieter', 'karte.text', 300, 0, 'EUR', 'verhandelt',
			 '', '2026-09-15T00:00:00.000Z', '')`); fehler != nil {
		t.Fatalf("den unbekannten Stand legen: %v", fehler)
	}

	if _, fehler := journal.Stand(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("der unbekannte Stand ging als null durch")
	}
}

// **Eine Grenze in einer anderen Währung sperrt** — statt umzurechnen.
func TestEineGrenzeInAndererWaehrungSperrt(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, domain.Betrag{Wert: 500, Waehrung: "USD"},
		probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	// **Der Auftrag rechnet in Euro, die Grenze in Dollar.**
	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 100), domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("zwei Währungen wurden verrechnet")
	}

	// **Und der Stand ebenso** — er zieht die Summe von der Grenze ab.
	if _, fehler := db.ExecContext(t.Context(), `
		INSERT INTO kostenposten (
			kennung, profil_id, konto, aufgabe, hoechstbetrag, betrag,
			waehrung, stand, begonnen
		) VALUES ('eins', 1, 'anbieter', 'karte.text', 100, 0, 'EUR',
			'reserviert', '2026-09-15T00:00:00.000Z')`); fehler != nil {
		t.Fatalf("den Posten legen: %v", fehler)
	}

	if _, fehler := journal.Stand(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("der Stand rechnete zwei Währungen gegeneinander")
	}
}

// **Ein geschlossener Bestand nimmt keine Buchung an** — und sagt es beim
// Beginnen, nicht beim Schreiben.
func TestEinGeschlossenerBestandNimmtKeineBuchung(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler == nil {
		t.Error("gesetzt wurde auf einem geschlossenen Bestand")
	}
}

// **Die Grenze fehlt, während der Posten steht** — zwei Abfragen, zwei Wege.
func TestOhneBudgettafelScheitertDieReservierungAnIhr(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	// **Nur die Budgettafel geht weg** — die Suche nach der Identität glückt
	// also, und der Weg kommt bis zur Grenze.
	ohneTafeln(t, db, "kostenbudgets")

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 100), domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("reserviert wurde ohne Budgettafel")
	}

	// **Und der Stand ebenso** — er liest sie zuerst.
	if _, fehler := journal.Stand(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("der Stand kam ohne Budgettafel durch")
	}
}

// **Ein Auftrag, den die Tafel nicht annimmt, wird gemeldet** — das Lesen
// glückte, das Schreiben nicht.
func TestEineTafelDieNichtsAnnimmtWirdGemeldet(t *testing.T) {
	db := frischeDatenbank(t)
	journal := storage.NeuesKostenjournal(db)

	if fehler := journal.BudgetSetzen(t.Context(), 1, "anbieter",
		domain.ZeitraumTag, euro(500), probezeit); fehler != nil {
		t.Fatalf("das Budget setzen: %v", fehler)
	}

	ohneTafeln(t, db, "kostenposten")

	// **Eine Bedingung, die für jede Zeile falsch ist**: gelesen wird, wie
	// zuvor — geschrieben wird nichts mehr.
	if _, fehler := db.ExecContext(t.Context(), `
		CREATE TABLE kostenposten (
			kennung TEXT PRIMARY KEY, profil_id INTEGER, konto TEXT, aufgabe TEXT,
			hoechstbetrag INTEGER, betrag INTEGER, waehrung TEXT, stand TEXT,
			grund TEXT, begonnen TEXT, beendet TEXT,
			CHECK (kennung <> kennung)
		)`); fehler != nil {
		t.Fatalf("die sperrige Tafel anlegen: %v", fehler)
	}

	if _, fehler := journal.Reservieren(t.Context(),
		auftrag("eins", 100), domain.ZeitraumTag, probezeit); fehler == nil {
		t.Error("gebucht wurde in eine Tafel, die nichts annimmt")
	}
}
