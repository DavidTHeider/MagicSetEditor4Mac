package msepaket

// **Die Paketwurzeln, gemerkt und je Ruf geprüft** — PERF-02
// (#2169).
//
// Die Flächen ohne Fenster fragten [Paketwurzeln] bei jedem Ruf, und im vollen
// Pack heißt das: Ordner mit rund tausend Einträgen lesen, zweimal. Gemessen
// am 2026-10-07: ein warmer Ruf `lesen.skript` mit `1 + 1` über `mse-mcp`
// brauchte rund 13 ms (`docs/LEISTUNGSSCHNITTE.md`). Der Vertrag hier: **was
// sich unter einer Wurzel ändert, gilt beim nächsten Ruf** — ohne Frist, denn
// die Änderungszeit eines Ordners wechselt, sobald darin etwas hinzukommt,
// verschwindet oder umbenannt wird.

import (
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"
)

// paketOrdner legt ein leeres Paketverzeichnis an.
func paketOrdner(t *testing.T, pfad string) {
	t.Helper()

	if fehler := os.MkdirAll(pfad, 0o750); fehler != nil {
		t.Fatal(fehler)
	}
}

// spaeter wartet, bis die Uhr der Platte sicher weiter ist — sonst trüge
// eine Änderung im selben Augenblick dieselbe Zeit.
func spaeter() { time.Sleep(5 * time.Millisecond) }

func TestDieWurzelsucheMerktUndPrueft(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	paketOrdner(t, filepath.Join(wurzel, "pack", "data", "eins.mse-style"))

	suche := NeueWurzelsuche()
	daten := filepath.Join(wurzel, "pack", "data")

	if gefunden := suche.Paketwurzeln(wurzel); !slices.Equal(gefunden, []string{daten}) {
		t.Fatalf("zuerst: %v", gefunden)
	}

	t.Run("ohne Änderung kein zweiter Lauf", func(t *testing.T) {
		vorher := suche.laeufe
		if gefunden := suche.Paketwurzeln(wurzel); !slices.Equal(gefunden, []string{daten}) {
			t.Fatalf("warm: %v", gefunden)
		}
		if suche.laeufe != vorher {
			t.Errorf("%d Läufe, erwartet %d", suche.laeufe, vorher)
		}
	})

	t.Run("eine neue Wurzel in der Tiefe gilt sofort", func(t *testing.T) {
		spaeter()
		paketOrdner(t, filepath.Join(wurzel, "zweites", "inhalt", "zwei.mse-game"))
		neu := filepath.Join(wurzel, "zweites", "inhalt")
		if gefunden := suche.Paketwurzeln(wurzel); !slices.Contains(gefunden, neu) {
			t.Errorf("die neue Wurzel fehlt: %v", gefunden)
		}
	})

	t.Run("eine entfernte Wurzel ist sofort fort", func(t *testing.T) {
		spaeter()
		if fehler := os.RemoveAll(filepath.Join(wurzel, "pack")); fehler != nil {
			t.Fatal(fehler)
		}
		if gefunden := suche.Paketwurzeln(wurzel); slices.Contains(gefunden, daten) {
			t.Errorf("die entfernte Wurzel steht noch da: %v", gefunden)
		}
	})
}

// **Eine Wurzel, die es erst später gibt**, wird gefunden, sobald sie da ist.
func TestEineSpaeterAngelegteWurzelGilt(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	wurzel := filepath.Join(ort, "pakete")
	suche := NeueWurzelsuche()

	if gefunden := suche.Paketwurzeln(wurzel); len(gefunden) != 0 {
		t.Fatalf("ohne Wurzel: %v", gefunden)
	}

	paketOrdner(t, filepath.Join(wurzel, "drei.mse-style"))

	if gefunden := suche.Paketwurzeln(wurzel); !slices.Equal(gefunden, []string{wurzel}) {
		t.Errorf("die neue Wurzel fehlt: %v", gefunden)
	}
}

// **Ein Ordner, der sich nicht lesen lässt, wird nicht gemerkt** — wird er
// lesbar, ändert sich seine Änderungszeit nicht, und das Gemerkte bliebe.
func TestEinUnlesbarerOrdnerWirdNichtGemerkt(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	zu := filepath.Join(wurzel, "zu")
	paketOrdner(t, filepath.Join(zu, "vier.mse-style"))

	if fehler := os.Chmod(zu, 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(zu, 0o750) }) //nolint:errcheck,gosec // beim Aufräumen ist nichts mehr zu retten

	suche := NeueWurzelsuche()
	suche.Paketwurzeln(wurzel)

	if fehler := os.Chmod(zu, 0o750); fehler != nil { //nolint:gosec // ein Verzeichnis braucht das Ausführrecht
		t.Fatal(fehler)
	}

	if gefunden := suche.Paketwurzeln(wurzel); !slices.Equal(gefunden, []string{zu}) {
		t.Errorf("nach dem Öffnen: %v", gefunden)
	}
}

// **Zwei erste Rufe zugleich gehen einmal über den Baum.**
func TestZweiErsteRufeLaufenEinmal(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	paketOrdner(t, filepath.Join(wurzel, "fuenf.mse-style"))

	suche := NeueWurzelsuche()

	var warte sync.WaitGroup
	for range 8 {
		warte.Go(func() { suche.Paketwurzeln(wurzel) })
	}

	warte.Wait()

	if suche.laeufe != 1 {
		t.Errorf("%d Läufe, erwartet 1", suche.laeufe)
	}
}

// **Was herausgeht, ist eine Kopie** — wer die Liste ändert, ändert nicht das
// Gemerkte.
func TestDieWurzelsucheGibtEineKopie(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	paketOrdner(t, filepath.Join(wurzel, "sechs.mse-style"))

	suche := NeueWurzelsuche()
	erste := suche.Paketwurzeln(wurzel)
	erste[0] = "verbogen"

	if zweite := suche.Paketwurzeln(wurzel); zweite[0] != wurzel {
		t.Errorf("das Gemerkte wurde verbogen: %v", zweite)
	}
}

// **Eine Datei neben den Paketen** ist kein Ordner und wird nicht gelesen —
// `README.md` liegt in jedem Pack.
func TestEineDateiImBaumStoertNicht(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	paketOrdner(t, filepath.Join(wurzel, "data", "sieben.mse-style"))

	if fehler := os.WriteFile(filepath.Join(wurzel, "README.md"), []byte("Pack"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	suche := NeueWurzelsuche()

	if gefunden := suche.Paketwurzeln(wurzel); !slices.Equal(gefunden, []string{filepath.Join(wurzel, "data")}) {
		t.Errorf("gefunden: %v", gefunden)
	}

	suche.Paketwurzeln(wurzel)

	if suche.laeufe != 1 {
		t.Errorf("%d Läufe, erwartet 1 — die Datei machte den Lauf unsauber", suche.laeufe)
	}
}

// **An der Grenze wird ganz geräumt** — und danach gilt wieder, was gefragt
// wird.
func TestDieWurzelsucheRaeumtAnDerGrenze(t *testing.T) {
	t.Parallel()

	suche := NeueWurzelsuche()
	ort := t.TempDir()

	for stelle := range grenzeWurzelsuche + 1 {
		wurzel := filepath.Join(ort, string(rune('a'+stelle)))
		paketOrdner(t, filepath.Join(wurzel, "acht.mse-style"))
		suche.Paketwurzeln(wurzel)
	}

	if gemerkt := len(suche.gemerkt); gemerkt != 1 {
		t.Errorf("%d gemerkt, erwartet 1 nach dem Räumen", gemerkt)
	}
}
