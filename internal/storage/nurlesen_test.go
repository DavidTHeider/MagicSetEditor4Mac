// **Ein Bestand, in den niemand schreiben kann** — Werkstück 16 von Phase 4.12
// (#1351).
//
// # Warum der Schutz hier steht und nicht an den Bindungen
//
// **Weil eine Zusage, die an 116 Bindungen einzeln hängt, beim ersten
// vergessenen Ruf still bricht** — **am 2026-09-06 gezählt** — das steht so in
// [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md) und
// war der Grund, `leser` bis hierher gar nicht erst hineinzulassen. Der
// Treiber kennt keine Ausnahme.
package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	sqlite "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Gelesen wird, geschrieben nicht — und zwar auf jedem Weg.**
//
// **Die vier Wege sind vier verschiedene Stellen im Treiber**: eine Anweisung
// unmittelbar, eine in einer Transaktion, eine Schemaänderung und ein Löschen.
// Wer nur die erste prüfte, hätte drei ungeprüfte Klassen.
func TestEinNurLesenderBestandNimmtKeinenSchreibvorgang(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	bestandMit(t, ort, "Feuerball")

	db, err := storage.OeffnenNurLesend(probekontext(t), ort)
	if err != nil {
		t.Fatalf("nur lesend öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

	// **Erst der Beleg, dass überhaupt etwas geht.** Ein Bestand, der auch
	// nicht liest, erfüllte diesen Lauf, ohne die Fähigkeit zu haben.
	var wert string
	if fehler := db.QueryRowContext(probekontext(t),
		`SELECT wert FROM probe`).Scan(&wert); fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if wert != "Feuerball" {
		t.Fatalf("gelesen wurde %q", wert)
	}

	// **Der Treiber selbst weist ab** — vier Stellen, und sie sind
	// verschieden: eine Anweisung unmittelbar, ein Löschen, eine
	// Schemaänderung und der Beginn einer Transaktion. Mit `_txlock=immediate`
	// scheitert schon das Beginnen, und das ist die früheste Stelle, an der
	// es scheitern kann.
	for name, anweisung := range map[string]string{
		"eine Zeile anlegen": `INSERT INTO probe (wert) VALUES ('neu')`,
		"eine Zeile löschen": `DELETE FROM probe`,
		"das Schema ändern":  `CREATE TABLE zweit (a)`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, fehler := db.ExecContext(
				probekontext(t), anweisung); fehler == nil {
				t.Fatal("der Schreibvorgang glückte")
			}
		})
	}

	t.Run("eine Transaktion beginnen", func(t *testing.T) {
		tx, fehler := db.BeginTx(probekontext(t), nil)
		if fehler == nil {
			_ = tx.Rollback() //nolint:errcheck // Test

			t.Fatal("die schreibende Transaktion begann")
		}
	})
}

// **Und über ein Repository kommt die Meldung der Domäne an, nicht die von
// SQLite.**
//
// # Warum das ein zweiter Lauf ist und keine Zeile im ersten
//
// **Weil es eine andere Naht ist, und der erste Anlauf hat sie verfehlt.** Er
// rief `ExecContext` unmittelbar und erwartete `ErrNurLesend` — die Zuordnung
// sitzt aber in `einordnen`, und dorthin kommt nur, was durch ein Repository
// geht. Der Lauf war rot, und zwar zu Recht: er prüfte den Treiber und
// behauptete etwas über die Übersetzung.
//
// **Was hier ankommt, liest ein Mensch** (§8): „attempt to write a readonly
// database" sagt niemandem, dass es an einer **Freigabe** liegt und nicht an
// der Datei.
func TestUeberEinRepositoryKommtDieMeldungDerDomaeneAn(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	// **Ein migrierter Bestand**, denn der Schreibweg geht über echte
	// Tabellen.
	frisch, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), frisch, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if fehler := frisch.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	db, err := storage.OeffnenNurLesend(probekontext(t), ort)
	if err != nil {
		t.Fatalf("nur lesend öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

	_, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(
		probekontext(t), probeSet(), nil)

	if !errors.Is(fehler, domain.ErrNurLesend) {
		t.Fatalf("Fehler %v, erwartet war ErrNurLesend", fehler)
	}

	// **Und die rohe Ursache reist mit** — fürs Protokoll, nicht für die
	// Oberfläche. Dieselbe Zusage wie beim Konflikt nebenan.
	var ursache sqlite.ExtendedErrorCode
	if !errors.As(fehler, &ursache) {
		t.Error("die Ursache ging verloren")
	}
}

// **Und die Gegenprobe: derselbe Bestand nimmt den Schreibvorgang, wenn er
// gewöhnlich geöffnet wird.**
//
// **Ohne sie sagte der Lauf darüber nichts** (§10): eine Datei, die aus einem
// anderen Grund nichts annimmt, erfüllte ihn genauso.
func TestDerselbeBestandNimmtIhnGewoehnlichGeoeffnet(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	bestandMit(t, ort, "Feuerball")

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO probe (wert) VALUES ('neu')`); fehler != nil {
		t.Errorf("der gewöhnlich geöffnete Bestand nahm nichts an: %v", fehler)
	}
}

// **Und ein verschlüsselter Bestand ohne Schlüssel geht auch nur-lesend nicht
// auf — mit derselben Auskunft.**
//
// # Warum das eine eigene Zeile wert ist
//
// **Weil es zwei Wege in denselben Bestand sind** (§10: zwei grüne Seiten
// einer Naht sind kein Beleg für die Naht). `Oeffnen` fragt nach dem
// Schlüssel, und `OeffnenNurLesend` muss es auch — sonst käme ein `leser` an
// einer Prüfung vorbei, die für den Schreiber gilt.
func TestAuchNurLesendVerlangtDerBestandSeinenSchluessel(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "chiffre.db")

	// **Eine Datei, die keine Klartextmarke trägt** — für `IstVerschluesselt`
	// ist das eine Chiffre, und dann wird der Schlüsselbund gefragt.
	if err := os.WriteFile(ort, []byte("kein SQLite, sondern Text"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	db, fehler := storage.OeffnenNurLesend(probekontext(t), ort)
	if fehler == nil {
		_ = db.Close() //nolint:errcheck // Test

		t.Fatal("der Bestand ging ohne Schlüssel auf")
	}

	if !errors.Is(fehler, domain.ErrSchluesselFehlt) {
		t.Errorf("Fehler %v, erwartet war ErrSchluesselFehlt", fehler)
	}
}
