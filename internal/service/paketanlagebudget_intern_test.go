package service

// **Die `game`-Datei unter einem neuen Stil wird mit Budget gelesen** —
// Befund 4 des Zweitblicks über Phase 6.1 (FMT-06.AK-01).
//
// `spieldatei` las mit `os.ReadFile`, und ein Spielpaket kommt von außen. Eine
// Datei über der Grenze einer Paketdatei belegte ganz den Speicher, bevor
// irgendetwas geprüft war — und kam danach als „liegt in keiner Paketwurzel"
// an, weil jeder Lesefehler dort als „nicht gefunden" zählte.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinUebergrossesSpielWirdNichtGanzGelesen(t *testing.T) {
	ort := t.TempDir()
	spielMit(t, ort, "gross", "")

	// **Eine Datei mit Loch** — größer als die Grenze, ohne die Platte zu füllen.
	if err := os.Truncate(filepath.Join(ort, "gross.mse-game", "game"),
		msepaket.GrenzePaketdatei+1); err != nil {
		t.Fatal(err)
	}

	var vorher, nachher runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&vorher)

	_, fehler := NeuerPaketverwalter(&anlageablage{}, ort, ort).Anlegen(t.Context(), Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "S", Spiel: "gross",
	})

	runtime.ReadMemStats(&nachher)

	if !errors.Is(fehler, msepaket.ErrZuGross) {
		t.Errorf("Fehler %v, erwartet ErrZuGross", fehler)
	}

	if belegt := nachher.TotalAlloc - vorher.TotalAlloc; belegt > 64<<20 {
		t.Errorf("%d MiB zugeteilt — die Datei wurde ganz gelesen", belegt>>20)
	}
}
