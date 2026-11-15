package app

// **Die Abdrücke der Ressourcen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass zwei verschiedene Dateien zwei verschiedene Abdrücke geben** — sonst
// zeigte ein ausgetauschter Rahmen das Bild von vorhin, und zwar still.
//
// **Dass beide Schriftschnitte mitgehen.** Ein Stil nennt `italic name:` für
// den kursiven Schnitt, und MSE setzt fett über ein **Gewicht**; ein
// Schlüssel, der nur die aufrechte Schrift kennte, ließe einen
// ausgetauschten fetten Schnitt durch.
//
// **Und dass der Rücken am Dateinamen hängt** — nicht am `icon:` des Pakets,
// denn fünf von sieben Spielpaketen nennen dort eine Beispiel-**Vorderseite**
// (gemessen am 2026-09-15 über `Full-Magic-Pack`).
//
// **Paketintern**, weil `spielressourcen` es ist: von außen ließe sich nur
// beobachten, dass ein Schlüssel sich ändert — und **warum**, stünde nirgends.

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// einPaketMit legt ein Paketverzeichnis mit einer Datei an und gibt die Wurzel.
//
// **Ein echtes Verzeichnis und keine Zeichenkette** (§3): `msepaket` liest
// Pakete, und eine Vorrichtung, die den Weg dorthin abkürzte, prüfte den Weg
// nicht.
func einPaketMit(t *testing.T, paket, datei string, inhalt []byte) string {
	t.Helper()

	wurzel := t.TempDir()
	ort := filepath.Join(wurzel, paket)

	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("%s anlegen: %v", ort, err)
	}

	if err := os.WriteFile(filepath.Join(ort, datei), inhalt, 0o600); err != nil {
		t.Fatalf("%s schreiben: %v", datei, err)
	}

	return wurzel
}

// **Zwei verschiedene Bilder geben zwei verschiedene Abdrücke.**
func TestZweiBilderGebenZweiAbdruecke(t *testing.T) {
	t.Parallel()

	wurzel := einPaketMit(t, "spiel.mse-game", "rahmen.png", []byte("eins"))
	quelle := spielressourcen{
		bilder:    msepaket.NeueBildabdruecke(wurzel),
		schriften: schriften.Neue(nil),
	}

	erster := quelle.Bildabdruck("/spiel.mse-game/rahmen.png")

	zweiter := spielressourcen{
		bilder: msepaket.NeueBildabdruecke(
			einPaketMit(t, "spiel.mse-game", "rahmen.png", []byte("zwei"))),
		schriften: schriften.Neue(nil),
	}.Bildabdruck("spiel.mse-game/rahmen.png")

	if erster == "" {
		t.Error("kein Abdruck")
	}

	if erster == zweiter {
		t.Errorf("beide geben %q", erster)
	}

	// **Und eine Datei, die es nicht gibt, gibt etwas Drittes.**
	if fehlend := quelle.Bildabdruck("/spiel.mse-game/gibtesnicht.png"); fehlend == erster {
		t.Error("ein fehlendes Bild gibt denselben Abdruck wie ein vorhandenes")
	}
}

// **Der Schriftabdruck trägt beide Schnitte.**
//
// **Gemessen am Trennzeichen**: aufrecht und fett stehen darin, und wer nur
// einen nähme, hätte die Hälfte der Frage beantwortet.
func TestDerSchriftabdruckTraegtBeideSchnitte(t *testing.T) {
	t.Parallel()

	suche := schriften.Neue(nil)
	quelle := spielressourcen{
		bilder: msepaket.NeueBildabdruecke(t.TempDir()), schriften: suche,
	}

	abdruck := quelle.Schriftabdruck("MPlantin")

	teile := strings.Split(abdruck, "\x1f")
	if len(teile) != 2 {
		t.Fatalf("%q zerfällt in %d Teile, erwartet zwei", abdruck, len(teile))
	}

	if teile[0] != suche.Abdruck("MPlantin", false) ||
		teile[1] != suche.Abdruck("MPlantin", true) {
		t.Errorf("die Teile von %q sind nicht die beiden Schnitte", abdruck)
	}
}

// **Der Rücken hängt am Dateinamen, nicht am Symbol des Pakets.**
func TestDerRueckenHaengtAmDateinamen(t *testing.T) {
	t.Parallel()

	wurzel := einPaketMit(t, "magic.mse-game", Rueckenbilddatei, einPNG())
	quelle := spielressourcen{
		bilder:    msepaket.NeueBildabdruecke(wurzel),
		schriften: schriften.Neue(nil),
	}

	if gesehen := quelle.Rueckenbild("magic"); gesehen != "/magic.mse-game/"+Rueckenbilddatei {
		t.Errorf("Rueckenbild %q", gesehen)
	}

	// **Ein Spiel ohne Rücken gibt nichts** — dann zeichnet `service` einen.
	if gesehen := quelle.Rueckenbild("saat"); gesehen != "" {
		t.Errorf("ein Spiel ohne Rücken gibt %q", gesehen)
	}

	// **Und ein leerer Spielname fragt gar nicht erst.**
	if gesehen := quelle.Rueckenbild(""); gesehen != "" {
		t.Errorf("der leere Spielname gibt %q", gesehen)
	}
}

// einPNG gibt ein paar Bytes, die als Datei taugen.
func einPNG() []byte { return []byte("\x89PNG\r\n\x1a\nirgendwas") }

// **Und er findet ihn auch in einem gepackten Paket.**
//
// **Beide Gestalten gibt es**: ein bezogenes Paket liegt als Verzeichnis da,
// ein `.mse-game` aus MSE2 als ZIP. Wer nur die erste prüfte, hätte für die
// zweite eine Behauptung.
func TestDerRueckenWirdAuchImGepacktenPaketGefunden(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	pfad := filepath.Join(wurzel, "magic.mse-game")

	datei, err := os.Create(pfad) //nolint:gosec // ein Pfad aus t.TempDir
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	packer := zip.NewWriter(datei)

	eintrag, err := packer.Create(Rueckenbilddatei)
	if err != nil {
		t.Fatalf("eintragen: %v", err)
	}

	if _, err := eintrag.Write(einPNG()); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	quelle := spielressourcen{
		bilder:    msepaket.NeueBildabdruecke(wurzel),
		schriften: schriften.Neue(nil),
	}

	if gesehen := quelle.Rueckenbild("magic"); gesehen == "" {
		t.Error("im gepackten Paket wurde keiner gefunden")
	}
}
