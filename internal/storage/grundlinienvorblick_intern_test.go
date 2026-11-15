package storage

// **Was vor der Sperre gelesen wird, gehört zu einem Stand** — gefunden von
// der Vorprüfung vor der Freigabe 0.8.0, als
// `TestZweiProzesseLegenEinenBestandZugleichAn` unter Last rot wurde, einmal
// mit „kein Bestand dieser Anwendung", einmal mit „verschlüsselt".
//
// Beide Male las der zweite Prozess den Bestand, **während** der erste ihn
// anlegte: einmal zwei Blicke, zwischen denen der erste festschrieb — die
// Buchführung leer, die Tabellen schon da —, einmal den Dateikopf, bevor der
// erste ihn geschrieben hatte. Die beiden Fälle hier stellen den Augenblick,
// statt auf ihn zu hoffen.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die zwei Blicke vor der Sperre sehen denselben Stand.** Schreibt ein
// anderer zwischen ihnen fest, sah der erste die Buchführung leer und der
// zweite die Tabellen — und der Bestand galt als fremd.
func TestZwischenDenBlickenVorDerSperreSchreibtNiemand(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	zweiter, fehler := Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = zweiter.Close() }() //nolint:errcheck // Aufräumen

	erster, fehler := Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = erster.Close() }() //nolint:errcheck // Aufräumen

	// **Der erste legt genau zwischen den beiden Blicken an** — ohne die
	// Sperre über die Wanderung, wie eine Fassung vor M02 (#2346), die sie
	// nicht kennt. Mit ihr wartete er auf den zweiten.
	zwischenDenBlicken = func() {
		zwischenDenBlicken = nil

		if fehler := wandernImGriff(t.Context(), erster, migrationen, PruefeBestand, nil, grundlinie); fehler != nil {
			t.Errorf("der erste: %v", fehler)
		}
	}

	t.Cleanup(func() { zwischenDenBlicken = nil })

	if fehler := Migriere(t.Context(), zweiter, nil); fehler != nil {
		t.Errorf("der zweite: %v", fehler)
	}
}

// **Der Kopf wird unter der Öffnungssperre gelesen.** Hält ein anderer sie,
// weil er die Datei gerade anlegt, steht dort womöglich noch nichts — und
// ein Kopf aus Nullen ist nicht `SQLite format 3`, also galt die Datei als
// verschlüsselt.
func TestDerKopfWirdUnterDerSperreGelesen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	zurueck, fehler := Alleinanspruch(ort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Halb angelegt**: die Datei steht da, ihr Kopf noch nicht.
	if halb := os.WriteFile(ort, make([]byte, 4096), 0o600); halb != nil {
		t.Fatal(halb)
	}

	geoeffnet := make(chan error, 1)

	go func() {
		db, oeffnungsfehler := Oeffnen(t.Context(), ort)
		if oeffnungsfehler == nil {
			oeffnungsfehler = db.Close()
		}

		geoeffnet <- oeffnungsfehler
	}()

	select {
	case zu := <-geoeffnet:
		t.Fatalf("geöffnet, während ein anderer anlegt: %v", zu)
	case <-time.After(200 * time.Millisecond):
	}

	// **Der andere ist fertig**: die Datei ist ein Bestand.
	fertig := filepath.Join(t.TempDir(), "fertig.db")

	db, fehler := Oeffnen(t.Context(), fertig)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := errors.Join(db.Close(), os.Rename(fertig, ort)); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := zurueck(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := <-geoeffnet; errors.Is(fehler, domain.ErrSchluesselFehlt) || fehler != nil {
		t.Errorf("öffnen nach dem Anlegen: %v", fehler)
	}
}
