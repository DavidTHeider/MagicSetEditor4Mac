package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestMigriereLegtDieGrundlageAn(t *testing.T) {
	db := frischeDatenbank(t)

	var anzahl int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'bestand'`,
	).Scan(&anzahl); err != nil {
		t.Fatalf("sqlite_master nicht lesbar: %v", err)
	}

	if anzahl != 1 {
		t.Errorf("die Tabelle bestand fehlt nach der Migration")
	}
}

// Eine Migration muss sich zweimal anwenden lassen, ohne etwas zu tun.
//
// Der Normalfall ist nicht die frische Datenbank, sondern die vorhandene: jeder
// Start migriert, und ein zweiter Lauf, der die Tabelle noch einmal anlegt,
// bräche die Anwendung genau bei denen, die sie schon benutzen.
func TestMigriereIstIdempotent(t *testing.T) {
	db := frischeDatenbank(t)

	vorher := angelegtAm(t, db)

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("die zweite Migration schlug fehl: %v", err)
	}

	if nachher := angelegtAm(t, db); nachher != vorher {
		t.Errorf("der zweite Lauf hat den Bestand geändert: %q → %q", vorher, nachher)
	}
}

func TestMigriereFuehrtBuchUeberDasAngewendete(t *testing.T) {
	db := frischeDatenbank(t)

	angewendet, err := storage.AngewendeteMigrationen(probekontext(t), db)
	if err != nil {
		t.Fatalf("die Buchführung ist nicht lesbar: %v", err)
	}

	if len(angewendet) == 0 {
		t.Fatal("nach der Migration ist nichts verbucht")
	}
	if angewendet[0] != "0001_grundlage.sql" {
		t.Errorf("verbucht ist %q, erwartet war %q", angewendet[0], "0001_grundlage.sql")
	}
}

// Was eine Migration schreibt, muss die eigene Prüfung annehmen.
//
// Der Fall ist im Schwesterprojekt teuer bezahlt worden: eine Migration setzte
// einen Zeitstempel mit `strftime('%f')` — drei Nachkommastellen — ein, während
// die Anwendung neun schrieb und eine Prüfung genau diese Form verlangte. Die
// Anwendung lief; unbrauchbar war erst jede Sicherung, die daraus entstand.
//
// Dieser Wächter kostet einen Durchlauf und deckt jede künftige Migration ab,
// die Daten schreibt.
func TestEineFrischMigrierteDatenbankBestehtIhreEigenePruefung(t *testing.T) {
	db := frischeDatenbank(t)

	if err := storage.PruefeBestand(probekontext(t), db); err != nil {
		t.Errorf("eine frisch migrierte Datenbank besteht ihre eigene Prüfung nicht: %v", err)
	}
}

// Und die Gegenprobe: die Prüfung muss eine falsche Form auch wirklich abweisen.
//
// Ohne sie wäre der Wächter darüber grün, selbst wenn PruefeBestand nichts täte.
func TestPruefeBestandWeistEineFalscheZeitformAb(t *testing.T) {
	db := frischeDatenbank(t)

	// Am CHECK vorbei: die Prüfung soll den Bestand lesen, nicht dem Schema
	// vertrauen. Ein Wert kann auch aus einer Sicherung oder einer fremden
	// Fassung stammen.
	if _, err := db.ExecContext(probekontext(t),
		`PRAGMA ignore_check_constraints = ON`); err != nil {
		t.Fatalf("Prüfungen ließen sich nicht abschalten: %v", err)
	}
	if _, err := db.ExecContext(probekontext(t),
		`UPDATE bestand SET wert = '2026-08-08T10:00:00.123456789Z' WHERE schluessel = 'angelegt_am'`,
	); err != nil {
		t.Fatalf("der Bestand ließ sich nicht ändern: %v", err)
	}

	if err := storage.PruefeBestand(probekontext(t), db); err == nil {
		t.Error("neun Nachkommastellen kamen durch — die Prüfung greift nicht")
	}
}
