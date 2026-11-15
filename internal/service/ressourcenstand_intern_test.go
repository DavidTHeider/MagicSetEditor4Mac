package service

// **Woraus das Zeichnen außer dem Bestand liest, steht im Abdruck** — Befund
// 2 des Zweitblicks über Phase 3 von 0.9.0 (PERF02, #2352): eine bearbeitete
// Datei in einem vorhandenen Paket ändert die Zeit seiner Wurzel nicht, und
// eine Prüfung blieb über sie hinweg „aktuell".

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDerRessourcenstandSiehtPaketdateienUndBilder(t *testing.T) {
	wurzel, speicher := stempelplatz(t)
	bilder := t.TempDir()
	speicher.MitBildablage(NeueBildablage(bilder))

	vorher := speicher.Ressourcenstand(t.Context(), 7)
	if vorher == "" || speicher.Ressourcenstand(t.Context(), 7) != vorher {
		t.Fatalf("ohne Änderung %q, dann %q", vorher, speicher.Ressourcenstand(t.Context(), 7))
	}

	zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"), "card width: 375", "card width: 376")

	nachDemStil := speicher.Ressourcenstand(t.Context(), 7)
	if nachDemStil == vorher {
		t.Error("eine geänderte Datei im Stilpaket blieb ungesehen")
	}

	ordner := NeueBildablage(bilder).Verzeichnis(7)
	if err := os.MkdirAll(ordner, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(ordner, "bild1"), []byte("ein Bild"), 0o600); err != nil {
		t.Fatal(err)
	}

	if speicher.Ressourcenstand(t.Context(), 7) == nachDemStil {
		t.Error("ein Bild neben dem Set blieb ungesehen")
	}
}

// **Ein Set ohne Stil hat einen Stand und keinen Fehler** — die Befunde der
// Prüfung nennen den fehlenden Stil; entsteht er, ändert sich die Wurzel.
func TestEinSetOhneStilHatEinenRessourcenstand(t *testing.T) {
	_, speicher := stempelplatz(t)

	if stand := speicher.Ressourcenstand(t.Context(), 99); stand == "" {
		t.Error("kein Stand für ein Set ohne Stil")
	}
}
