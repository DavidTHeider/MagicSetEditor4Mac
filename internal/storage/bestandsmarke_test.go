// **Die Marke eines Bestands** —
// #967.
package storage_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// migrierterBestand legt einen Bestand an einem **benannten** Pfad an.
//
// **Nicht `frischeDateidatenbank`**: die wählt ihren Pfad selbst, und dieser
// Lauf braucht ihn — er liest dieselbe Datei ein zweites Mal über ihren Namen.
func migrierterBestand(t *testing.T, pfad string) *sql.DB {
	t.Helper()

	db := frischeDatei(t, pfad)

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("die Migration schlug fehl: %v", err)
	}

	return db
}

// frischeDatei öffnet eine leere Datenbankdatei, ohne zu migrieren.
func frischeDatei(t *testing.T, pfad string) *sql.DB {
	t.Helper()

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("die Datenbank ließ sich nicht öffnen: %v", err)
	}

	return db
}

// **Derselbe Bestand gibt dieselbe Marke.**
//
// Ohne das wäre jede Sicherung von einem anderen Stand, und die
// Wiederverwendung fiele weg — mit ihr die Zusage, dass eine Migration nicht
// bei jedem Versuch eine neue Kopie schreibt.
func TestDerselbeBestandGibtDieselbeMarke(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db := migrierterBestand(t, pfad)

	erste, err := storage.BestandsmarkeVon(probekontext(t), db)
	if err != nil {
		t.Fatalf("die erste Marke: %v", err)
	}

	zweite, err := storage.BestandsmarkeVon(probekontext(t), db)
	if err != nil {
		t.Fatalf("die zweite Marke: %v", err)
	}

	if erste != zweite {
		t.Errorf("zweimal derselbe Bestand gab %q und %q", erste, zweite)
	}

	if len(erste) != 12 {
		t.Errorf("die Marke ist %q — zwölf Zeichen waren gemeint", erste)
	}
}

// **Eine Änderung bewegt die Marke.**
//
// # Die Reichweite dieser Messung
//
// Geprüft wird der Fall, für den es die Marke gibt: **Kartenarbeit** (§10 —
// die Reichweite einer Messung ist die Reichweite ihrer Aussage). Eine reine
// Änderung an einer bestehenden Zeile bewegt sie **nicht**, und das steht am
// Quelltext. Wer die Marke für etwas anderes benutzt, liest das zuerst.
func TestEineAenderungBewegtDieMarke(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db := migrierterBestand(t, pfad)

	vorher, err := storage.BestandsmarkeVon(probekontext(t), db)
	if err != nil {
		t.Fatalf("die Marke vorher: %v", err)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id)
		 VALUES ('Neu', 'magic', '2026-08-25T00:00:00Z', 1)`,
	); fehler != nil {
		t.Fatalf("ein Set anlegen: %v", fehler)
	}

	nachher, err := storage.BestandsmarkeVon(probekontext(t), db)
	if err != nil {
		t.Fatalf("die Marke nachher: %v", err)
	}

	if vorher == nachher {
		t.Errorf("beide Marken sind %q — ein angelegtes Set hat sie nicht bewegt", vorher)
	}
}

// **Und sie lässt sich über den Pfad einer Datei bestimmen**, ohne sie zu
// verändern.
//
// Das ist der Weg, den `sicherungVorMigration` geht: die Sicherung liegt als
// Datei da und darf nicht angefasst werden.
func TestDieMarkeEinerDateiLiestOhneZuSchreiben(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")
	db := migrierterBestand(t, pfad)

	ausDerVerbindung, err := storage.BestandsmarkeVon(probekontext(t), db)
	if err != nil {
		t.Fatalf("die Marke aus der Verbindung: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	ausDerDatei, err := storage.Bestandsmarke(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("die Marke aus der Datei: %v", err)
	}

	if ausDerDatei != ausDerVerbindung {
		t.Errorf("die Datei gab %q, die Verbindung %q", ausDerDatei, ausDerVerbindung)
	}
}

// **Eine Datei ohne Tabellen hat keine Marke.**
//
// Sie ist kein Bestand, und eine Marke über nichts wäre für jede solche Datei
// dieselbe — zwei leere Dateien sähen aus wie derselbe Stand.
func TestEineDateiOhneTabellenHatKeineMarke(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "leer.db")

	db := frischeDatei(t, pfad)
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	_, err := storage.Bestandsmarke(probekontext(t), pfad)
	if !errors.Is(err, domain.ErrSicherung) {
		t.Fatalf("Fehler %v, erwartet war ein Sicherungsfehler", err)
	}

	if !strings.Contains(err.Error(), "keine einzige Tabelle") {
		t.Errorf("die Meldung sagt nicht, was fehlt: %v", err)
	}
}

// **Eine Datei, die es nicht gibt, hat keine Marke.**
func TestEineFehlendeDateiHatKeineMarke(t *testing.T) {
	t.Parallel()

	_, err := storage.Bestandsmarke(probekontext(t), filepath.Join(t.TempDir(), "gibt-es-nicht.db"))
	if !errors.Is(err, domain.ErrSicherung) {
		t.Errorf("Fehler %v, erwartet war ein Sicherungsfehler", err)
	}
}

// **Eine geschlossene Verbindung gibt keine Marke.**
//
// # Beide Abfragen, und zwar einzeln
//
// Die Marke geht über **zwei** Abfragen: erst die Tabellen, dann die Zahlen.
// Ein Lauf, der nur die erste scheitern lässt, sagt über die zweite nichts —
// und der Fehlerweg der zweiten wäre eine Behauptung (§3).
//
// **Die zweite wird deshalb an einer Datenbank gefahren, deren Tabellen schon
// gelesen sind**: die Liste kommt aus einer offenen Verbindung, gezählt wird
// über eine geschlossene.
func TestEineGeschlosseneVerbindungGibtKeineMarke(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")
	db := migrierterBestand(t, pfad)

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	_, err := storage.BestandsmarkeVon(probekontext(t), db)
	if !errors.Is(err, domain.ErrSicherung) {
		t.Errorf("Fehler %v, erwartet war ein Sicherungsfehler", err)
	}
}
