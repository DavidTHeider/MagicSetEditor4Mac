// **Hält ein zweiter Verbraucher die Datei?**
// (#961)
package storage_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Datei, die niemand hält, gehört uns allein.**
func TestAlleinAmBestandOhneZweitenVerbraucher(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := storage.AlleinAmBestand(probekontext(t), pfad); fehler != nil {
		t.Errorf("die geschlossene Datei galt als belegt: %v", fehler)
	}
}

// **Eine Datei, die ein zweiter offen hält, gehört uns nicht.**
//
// Das ist der Fall aus #961: `mse-mcp` läuft, und das Fenster will tauschen.
// Die zweite Verbindung steht hier für den zweiten **Prozess** — für SQLite
// ist beides dasselbe, es sperrt über die Datei.
func TestAlleinAmBestandMitZweitemVerbraucher(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	// **Ein offener Leser genügt.** Er blockiert in WAL keinen Schreiber —
	// aber er verhindert, dass jemand die Datei allein belegt, und genau das
	// ist die Frage.
	//nolint:rowserrcheck // die Zeilen sind die Vorrichtung, nicht das Ergebnis:
	// sie bleiben offen, damit die Datei belegt ist.
	zeilen, err := db.QueryContext(probekontext(t), `SELECT name FROM migrationen`)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	t.Cleanup(func() { _ = zeilen.Close() }) //nolint:errcheck // nur zum Aufräumen

	fehler := storage.AlleinAmBestand(probekontext(t), pfad)
	if fehler == nil {
		t.Fatal("die belegte Datei galt als frei")
	}

	// **Als Konflikt und nicht als irgendein Fehler**: die Oberfläche soll
	// sagen können, dass jemand anderes daran ist — nicht „unerwarteter
	// Fehler" (§8).
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", fehler)
	}
}

// **Und die Prüfung hinterlässt nichts.**
//
// Sie nimmt eine exklusive Sperre und eine Schreibtransaktion; beides muss
// wieder verschwinden, sonst prüft der nächste Aufruf sich selbst.
func TestAlleinAmBestandHinterlaesstNichts(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	for i := range 3 {
		if fehler := storage.AlleinAmBestand(probekontext(t), pfad); fehler != nil {
			t.Fatalf("Lauf %d: %v", i+1, fehler)
		}
	}

	// Danach lässt sie sich wie gewohnt öffnen und lesen.
	wieder, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("wieder öffnen: %v", err)
	}

	namen, err := storage.AngewendeteMigrationen(probekontext(t), wieder)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(namen) == 0 {
		t.Error("nach der Prüfung steht keine Migration mehr in der Buchführung")
	}

	if fehler := wieder.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}
}

// **Eine Datei, die kein Bestand ist, hält keine Sperre** (#962).
//
// # Warum das die teuerste Fehlauskunft der Prüfung war
//
// `BeginTx` scheitert auch, wenn die Datei gar keine Datenbank ist — und bis
// zum 2026-08-25 meldete diese Prüfung dafür „von einem anderen Programm
// geöffnet". Der Weg, der sie durchläuft, ist das **Zurückspielen einer
// Sicherung**: wer eine unbrauchbare Bestandsdatei hatte, bekam sie deshalb
// nicht ersetzt. Die Prüfung, die den Bestand schützen soll, verhinderte
// seine Rettung.
//
// Gefunden hat es die laufende Anwendung, kein Test.
func TestEineDateiDieKeinBestandIstGiltAlsFrei(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "kein-bestand.db")

	if err := os.WriteFile(pfad, []byte("das ist kein Bestand, sondern Text\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := storage.AlleinAmBestand(probekontext(t), pfad); err != nil {
		t.Errorf("eine Textdatei galt als belegt: %v", err)
	}
}

// **Ein Fehler, der nicht von SQLite kommt, bleibt ein Konflikt** (#962).
//
// `keinBestand` fragt den Fehlercode ab; kommt der Fehler von woanders — ein
// abgebrochener Kontext beim Beenden —, gibt es keinen. Ohne diesen Lauf wäre
// der Zweig eine Behauptung (§3).
func TestEinAbgebrochenerKontextBleibtEinKonflikt(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "kein-bestand.db")

	if err := os.WriteFile(pfad, []byte("kein Bestand\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	abgebrochen, abbrechen := context.WithCancel(probekontext(t))
	abbrechen()

	if err := storage.AlleinAmBestand(abgebrochen, pfad); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ein Konflikt", err)
	}
}
