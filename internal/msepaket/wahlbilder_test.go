package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein `choice images:`-Block, beide Formen.**
//
// `magic-m15.mse-style:163` schreibt den Pfad hin;
// `magic-modules.mse-include/rarities/choice_images` schachtelt ein `script:`.
// Der erste Anlauf las nur die erste Form und ließ genau die 342
// Seltenheitsfelder leer.
func TestEinWahlbildblockKenntBeideFormen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "wahl.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\ncard width: 10\ncard height: 10\n" +
		"card style:\n" +
		"\trarity:\n" +
		"\t\trender style: image\n" +
		"\t\tchoice images:\n" +
		"\t\t\tcommon: haeufig.png\n" +
		"\t\t\trare:\n" +
		"\t\t\t\tscript:\n" +
		"\t\t\t\t\tif x then \"a.png\" else \"b.png\"\n"

	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(stil.Felder) != 1 {
		t.Fatalf("%d Feldstile, erwartet einer", len(stil.Felder))
	}

	bilder := stil.Felder[0].Wahlbilder
	if len(bilder) != 2 {
		t.Fatalf("%d Wahlbilder, erwartet zwei: %+v", len(bilder), bilder)
	}

	if bilder[0].Wahl != "common" || bilder[0].Bild != "haeufig.png" {
		t.Errorf("die erste Wahl ist %+v, erwartet common → haeufig.png", bilder[0])
	}

	// **Das Skript kommt geklammert heraus**, damit es wie jede andere
	// Bildangabe behandelt werden kann.
	if bilder[1].Wahl != "rare" || bilder[1].Bild != `{ if x then "a.png" else "b.png" }` {
		t.Errorf("die zweite Wahl ist %+v, erwartet rare → ein geklammertes Skript", bilder[1])
	}
}

// **Eine Leerzeile im Block ist kein Wahlbild.**
//
// Ein `choice images:`-Block darf Kommentare und Leerzeilen tragen — das Pack
// schreibt `# Images based on the set symbol` mitten hinein.
func TestEineLeerzeileImWahlbildblockZaehltNicht(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "wahl.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\ncard width: 10\ncard height: 10\n" +
		"card style:\n" +
		"\trarity:\n" +
		"\t\trender style: image\n" +
		"\t\tchoice images:\n" +
		"\t\t\t# Images based on the set symbol\n" +
		"\n" +
		"\t\t\tcommon: haeufig.png\n"

	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	stil, _ := paket.Stilmodell()
	if len(stil.Felder) != 1 {
		t.Fatalf("%d Feldstile", len(stil.Felder))
	}

	if bilder := stil.Felder[0].Wahlbilder; len(bilder) != 1 {
		t.Errorf("%d Wahlbilder, erwartet eines: %+v", len(bilder), bilder)
	}
}
