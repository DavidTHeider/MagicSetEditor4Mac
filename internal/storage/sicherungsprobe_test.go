package storage_test

// **Eine Sicherung wird geprüft, bevor sie zählt** — F05
// (#2344),
// Befund A-21.
//
// Zweimal: **nach dem Schreiben**, bevor sie ihren Namen bekommt, und **vor
// dem Zurückspielen**, bevor sie den Bestand ersetzt. Beide Male mit
// `PRAGMA quick_check` — gelesen wird dabei jede Seite, nicht nur die, die
// das Öffnen und die Buchführung berühren. Bis hierher ging eine Sicherung
// mit einer zerschriebenen Datenseite durch beide Türen.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// datenseiteVerbiegen kehrt die Seite um, auf der eine Zeile der
// Ballasttabelle steht — **ohne die Datei zu öffnen**: `verbiegen` öffnet
// vorher und hebt sie damit auf WAL.
//
// **Gesucht und nicht geschätzt**: die vorletzte Seite, die `verbiegen`
// nimmt, gehört in einer Sicherung zum **Schema** — `VACUUM INTO` ordnet die
// Seiten neu —, und dann scheitert schon das Öffnen. Das prüfte die Tür
// davor, nicht `quick_check` (gemessen am 2026-10-14).
func datenseiteVerbiegen(t *testing.T, pfad string) {
	t.Helper()

	roh, err := os.ReadFile(pfad) // #nosec G304 -- der Pfad kommt aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	stelle := bytes.Index(roh, []byte("Text Nummer 2000"))
	if stelle < dieSeitengroesse {
		t.Fatalf("keine Zeile des Ballasts hinter der ersten Seite: %d", stelle)
	}

	anfang := stelle / dieSeitengroesse * dieSeitengroesse
	for i := anfang; i < anfang+dieSeitengroesse; i++ {
		roh[i] ^= 0xFF
	}

	// #nosec G703 -- der Pfad kommt aus t.TempDir(), nicht von außen
	if err := os.WriteFile(pfad, roh, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// bestandMitBallast ist ein Bestand mit genug Seiten, dass die vorletzte
// Tabellendaten trägt — dieselbe Vorrichtung wie bei der Integrität.
func bestandMitBallast(t *testing.T) string {
	t.Helper()

	return kaputterOrt(t, func(*testing.T, string) {})
}

func TestEineSicherungDieIhrePruefungNichtBestehtWirdVerworfen(t *testing.T) {
	db := geoeffnet(t, bestandMitBallast(t))
	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "sicherung.db")

	// **Die Naht ist für alle da** — darum nur Dateien dieses Ordners.
	storage.SicherungsschrittSetzen(func(schritt, pfad string) {
		if schritt == "vor dem Prüfen" && filepath.Dir(pfad) == ordner {
			datenseiteVerbiegen(t, pfad)
		}
	})
	t.Cleanup(func() { storage.SicherungsschrittSetzen(func(string, string) {}) })

	fehler := storage.Sichern(probekontext(t), db, ziel)
	if !errors.Is(fehler, domain.ErrSicherungVerworfen) {
		t.Fatalf("die Sicherung ging durch: %v", fehler)
	}

	if eintraege, _ := os.ReadDir(ordner); len(eintraege) != 0 { //nolint:errcheck // ein leerer Ordner ist die Antwort
		t.Errorf("im Ordner blieb etwas liegen: %v", eintraege)
	}

	var ergebnis string
	if err := db.QueryRowContext(probekontext(t), "PRAGMA quick_check").Scan(&ergebnis); err != nil || ergebnis != "ok" {
		t.Errorf("der Bestand selbst ist nicht mehr heil: %q, %v", ergebnis, err)
	}
}

func TestEineBeschaedigteSicherungWirdNichtAngenommen(t *testing.T) {
	t.Parallel()

	db := geoeffnet(t, bestandMitBallast(t))
	sicherung := filepath.Join(t.TempDir(), "sicherung.db")

	if err := storage.Sichern(probekontext(t), db, sicherung); err != nil {
		t.Fatalf("sichern: %v", err)
	}

	datenseiteVerbiegen(t, sicherung)

	if _, err := storage.PruefeSicherung(probekontext(t), sicherung); !errors.Is(err, domain.ErrSicherung) {
		t.Errorf("die beschädigte Sicherung ging als brauchbar durch: %v", err)
	}
}

func TestEineHeileSicherungBestehtIhrePruefung(t *testing.T) {
	t.Parallel()

	db := geoeffnet(t, bestandMitBallast(t))
	sicherung := filepath.Join(t.TempDir(), "sicherung.db")

	if err := storage.Sichern(probekontext(t), db, sicherung); err != nil {
		t.Fatalf("sichern: %v", err)
	}

	if _, err := storage.PruefeSicherung(probekontext(t), sicherung); err != nil {
		t.Errorf("eine heile Sicherung wurde abgewiesen: %v", err)
	}
}
