package app

// **Der Start räumt alte eigene Reste — nur, wenn er allein ist** — OPS04
// (#2350).

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// gestartetUndBeendet startet eine Anwendung über `bestand` und beendet sie.
func gestartetUndBeendet(t *testing.T, bestand string) {
	t.Helper()

	anwendung := NeueApp(bestand)

	bilder := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return bilder, nil })

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	if err := anwendung.Beenden(); err != nil {
		t.Fatalf("beenden: %v", err)
	}
}

// legeAlt legt eine Datei an und stellt ihre Zeit zwei Stunden zurück.
func legeAlt(t *testing.T, pfad string) {
	t.Helper()

	legeAn(t, pfad, "x")

	damals := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(pfad, damals, damals); err != nil {
		t.Fatal(err)
	}
}

func TestDerStartRaeumtAlteResteNurAllein(t *testing.T) { //nolint:paralleltest // der Start greift auf Paketvariablen
	t.Setenv("HOME", t.TempDir())

	// **Am Vorgabeort** — dort gehören auch die Pakete daneben dieser Anwendung.
	bestand, err := Datenbankpfad()
	if err != nil {
		t.Fatal(err)
	}

	gestartetUndBeendet(t, bestand)

	ordner := filepath.Dir(bestand)
	rest := filepath.Join(ordner, ".rueckspiel-laeuft.json.12345")
	tiefer := filepath.Join(ordner, "beiwerk", "set-1", ".bild.png.9")
	paket := filepath.Join(ordner, "pakete", ".m15-magic-2718")
	fremd := filepath.Join(ordner, ".DS_Store")
	aehnlich := filepath.Join(ordner, ".notizen.2026")

	for _, pfad := range []string{rest, tiefer, paket, fremd, aehnlich} {
		legeAlt(t, pfad)
	}

	// **Ein anderer hält den Bestand** — dann bleibt alles liegen.
	halter, err := storage.Oeffnen(context.Background(), bestand)
	if err != nil {
		t.Fatal(err)
	}

	gestartetUndBeendet(t, bestand)

	if _, err := os.Stat(rest); err != nil {
		t.Errorf("geräumt, obwohl ein anderer den Bestand hielt: %v", err)
	}

	if err := halter.Close(); err != nil {
		t.Fatal(err)
	}

	gestartetUndBeendet(t, bestand)

	for _, pfad := range []string{rest, tiefer, paket} {
		if _, err := os.Stat(pfad); !os.IsNotExist(err) {
			t.Errorf("%s steht noch (%v)", pfad, err)
		}
	}

	for _, pfad := range []string{fremd, aehnlich} {
		if _, err := os.Stat(pfad); err != nil {
			t.Errorf("die fremde Datei %s ist fort: %v", pfad, err)
		}
	}
}

// **Neben einem Bestand an einem anderen Ort ist `pakete` fremd** (Zweitblick
// über Phase 4, Befund 1) — wer `MSE_BESTAND` nach `~/Dokumente` legt, hat
// dort vielleicht einen eigenen Ordner dieses Namens.
func TestNebenEinemFremdenOrtBleibtPaketeUnberuehrt(t *testing.T) { //nolint:paralleltest // der Start greift auf Paketvariablen
	t.Setenv("HOME", t.TempDir())

	bestand := filepath.Join(t.TempDir(), "Dokumente", "bestand.db")
	gestartetUndBeendet(t, bestand)

	// **Und `beiwerk` ebenso** (Nachlauf über Phase 4, Befund 1): neben einem
	// Bestand an einem anderen Ort belegt nichts, dass der Ordner von hier
	// stammt.
	fremde := []string{
		filepath.Join(filepath.Dir(bestand), "pakete", ".entwurf-2026"),
		filepath.Join(filepath.Dir(bestand), "beiwerk", ".notizen.2026"),
	}
	for _, fremd := range fremde {
		legeAlt(t, fremd)
	}

	gestartetUndBeendet(t, bestand)

	for _, fremd := range fremde {
		if _, err := os.Stat(fremd); err != nil {
			t.Errorf("neben einem fremden Ort geräumt: %v", err)
		}
	}
}

// **Ein Ort, der sich nicht lesen lässt, hält den Start nicht auf** — er wird
// gemeldet, und die Anwendung öffnet trotzdem.
func TestEinUnlesbarerOrtHaeltDenStartNichtAuf(t *testing.T) { //nolint:paralleltest // der Start greift auf Paketvariablen
	// **Am Vorgabeort** — nur dort gehört `beiwerk` dieser Anwendung und wird
	// gepflegt (Nachlauf über Phase 4, Befund 1).
	t.Setenv("HOME", t.TempDir())

	bestand, err := Datenbankpfad()
	if err != nil {
		t.Fatal(err)
	}

	gestartetUndBeendet(t, bestand)

	// **`beiwerk` als Datei** — ein Ordner, den es so nicht gibt.
	legeAn(t, filepath.Join(filepath.Dir(bestand), "beiwerk"), "keine Bilder")

	gestartetUndBeendet(t, bestand)
}
