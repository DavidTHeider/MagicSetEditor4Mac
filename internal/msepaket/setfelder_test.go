package msepaket_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Spiel erklärt auch seine Setfelder, nicht nur seine Kartenfelder.**
//
// # Warum sie gebraucht werden
//
// `magic.mse-game/set_fields` führt **89 `set field:`-Blöcke**. Die Skripte der
// Stile lesen daraus: `set.symbol` 420-mal, `set.border_color` 345-mal, über
// alle Namen 1.400 Zugriffe in 167 Dateien.
//
// Ohne die Erklärung wüsste die Umgebung nur, was ein Set **ausgefüllt** hat.
// Ein leer gelassenes Feld gäbe dann nichts statt einer leeren Zeichenkette —
// und `set.copyright + " x"` bräche ab, wo MSE weiterrechnet. Es ist dieselbe
// Klasse wie bei den Kartenfeldern
// (#278).
func TestEinSpielErklaertSeineSetfelder(t *testing.T) {
	paket, err := msepaket.Oeffne("../../saat/probe.mse-game")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	spiel, modellbefunde := paket.Spielmodell()
	if len(modellbefunde) > 0 {
		t.Fatalf("Modellbefunde: %v", modellbefunde)
	}

	if len(spiel.SetFields) != 2 {
		t.Fatalf("%d Setfelder, erwartet zwei: %+v", len(spiel.SetFields), spiel.SetFields)
	}

	// **Und ihre Vorgabe reist mit** — sonst wäre die Erklärung nur eine Liste
	// von Namen, und die Vorgabe des Spiels ginge auf dem Weg verloren.
	if spiel.SetFields[1].Default != "Probe" {
		t.Errorf("die Vorgabe ist %q, erwartet Probe", spiel.SetFields[1].Default)
	}
}

// **Ein Setfeld ohne `name` ist ein Vorbehalt, kein Abbruch.**
//
// Dieselbe Antwort wie beim Kartenfeld (§8): das Spiel wird gelesen, das eine
// Feld benannt und übergangen. Ein Spielpaket mit 89 Setfeldern, von denen
// eines unvollständig ist, muss die anderen 88 hergeben.
//
// **Und der Befund sagt, welche Sorte** — „ein Setfeld ohne `name`". Stünde
// dort „Kartenfeld", suchte der Leser an der falschen Stelle in einer Datei mit
// 1.400 Zeilen.
func TestEinSetfeldOhneNamenWirdBenannt(t *testing.T) {
	verzeichnis := t.TempDir()
	pfad := filepath.Join(verzeichnis, "ohnenamen.mse-game")

	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Ohne\nset field:\n\ttype: text\n"
	if err := os.WriteFile(filepath.Join(pfad, "game"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	spiel, befunde := paket.Spielmodell()
	if len(spiel.SetFields) != 0 {
		t.Errorf("das Spiel führt %d Setfelder, erwartet keines", len(spiel.SetFields))
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einen: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "Setfeld") {
		t.Errorf("der Befund lautet %q und nennt die Sorte nicht", befunde[0].Meldung)
	}
}
