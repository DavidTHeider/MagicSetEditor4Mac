package msepaket_test

// **Der Abdruck eines Paketbildes** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass ein ausgetauschter Rahmen auffällt.** Ein Zeichenplan trägt Pfade;
// ohne diesen Abdruck hätte ein Schlüssel über ihm keine Ahnung, dass die
// Datei dahinter eine andere ist — und der Spieltisch zeigte das Bild von
// vorhin.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// einPaketMitBild legt ein Paket mit einer Datei an und gibt die Wurzel.
func einPaketMitBild(t *testing.T, inhalt string) string {
	t.Helper()

	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe.mse-style")

	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatalf("das Paket anlegen: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(paket, "rahmen.png"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("das Bild schreiben: %v", err)
	}

	return wurzel
}

// **Dieselbe Datei gibt denselben Abdruck.**
func TestDieselbeDateiGibtDenselbenAbdruck(t *testing.T) {
	t.Parallel()

	abdruecke := msepaket.NeueBildabdruecke(einPaketMitBild(t, "die Bytes"))

	erster := abdruecke.Abdruck("/probe.mse-style/rahmen.png")
	zweiter := abdruecke.Abdruck("/probe.mse-style/rahmen.png")

	if erster != zweiter {
		t.Errorf("zwei Abdrücke: %q und %q", erster, zweiter)
	}

	if len(erster) != 64 || strings.HasPrefix(erster, "fehlt:") {
		t.Errorf("der Abdruck %q sieht nicht aus wie einer", erster)
	}
}

// **Eine andere Datei gibt einen anderen Abdruck.**
//
// **Die Aussage dieses Werkstücks**: ein ausgetauschter Rahmen verwirft das
// Bild. Ohne diesen Lauf wäre der Abdruck nur eine Behauptung.
func TestEineAndereDateiGibtEinenAnderenAbdruck(t *testing.T) {
	t.Parallel()

	erster := msepaket.NeueBildabdruecke(einPaketMitBild(t, "die Bytes")).
		Abdruck("/probe.mse-style/rahmen.png")

	zweiter := msepaket.NeueBildabdruecke(einPaketMitBild(t, "andere Bytes")).
		Abdruck("/probe.mse-style/rahmen.png")

	if erster == zweiter {
		t.Error("zwei Dateien, ein Abdruck")
	}
}

// **Eine fehlende Datei ist ein Wert und kein Fehler.**
//
// **Sonst verschwände die Karte ganz.** Wer eine Karte mit fehlendem Rahmen
// zeichnet, bekommt ein Bild mit einem Befund daneben — und dieses Bild soll
// wiedererkannt werden.
func TestEineFehlendeDateiIstEinWert(t *testing.T) {
	t.Parallel()

	abdruecke := msepaket.NeueBildabdruecke(t.TempDir())

	abdruck := abdruecke.Abdruck("/gibtesnicht.mse-style/rahmen.png")

	if !strings.HasPrefix(abdruck, "fehlt:") {
		t.Errorf("der Abdruck %q sagt nicht, dass die Datei fehlt", abdruck)
	}

	// **Und er ist ein anderer als der einer vorhandenen Datei** — taucht sie
	// auf, ändert sich der Schlüssel.
	vorhanden := msepaket.NeueBildabdruecke(einPaketMitBild(t, "da")).
		Abdruck("/probe.mse-style/rahmen.png")

	if abdruck == vorhanden {
		t.Error("fehlend und vorhanden geben denselben Abdruck")
	}
}

// **Ein unbrauchbarer Verweis ebenso.**
func TestEinUnbrauchbarerVerweisIstEinWert(t *testing.T) {
	t.Parallel()

	abdruck := msepaket.NeueBildabdruecke(t.TempDir()).Abdruck("../hinaus.png")

	if !strings.HasPrefix(abdruck, "fehlt:") {
		t.Errorf("der Abdruck %q sagt nicht, dass etwas fehlt", abdruck)
	}
}

// **Und ein Verweis auf ein Setpaket wird dort gesucht.**
//
// **Nicht in den Wurzeln**: das Bild einer Karte liegt im Set und nirgends
// sonst — dieselbe Regel wie bei [msepaket.Bildquelle.Bild].
func TestEinSetpaketverweisWirdImSetGesucht(t *testing.T) {
	t.Parallel()

	wurzel := einPaketMitBild(t, "im Set")
	ort := filepath.Join(wurzel, "probe.mse-style")

	// **Ein Setpaketverweis trägt sein Schema und seinen Trenner** — er nennt
	// den Ort des Sets und die Datei darin, und gesucht wird nur dort.
	abdruck := msepaket.NeueBildabdruecke().
		Abdruck("setpaket:" + ort + "!rahmen.png")

	if strings.HasPrefix(abdruck, "fehlt:") {
		t.Errorf("das Bild im Setpaket wurde nicht gefunden: %q", abdruck)
	}
}

// **Eine unlesbare Wurzel ist ein anderer Wert als eine fehlende Datei.**
//
// # Warum das zählt
//
// **Weil es zwei Lagen mit zwei Abhilfen sind** — derselbe Satz steht an
// [msepaket.Bildquelle.Bild]. Und für den Schlüssel heißt es: wird die Wurzel
// wieder lesbar, ändert sich der Abdruck, und die Karte wird neu gezeichnet.
func TestEineUnlesbareWurzelGibtEinenAnderenWert(t *testing.T) {
	t.Parallel()

	// **Zwei Wurzeln, und die erste ist eine Datei statt eines
	// Verzeichnisses** — sie klagt, und die zweite hat das Bild.
	unlesbar := filepath.Join(t.TempDir(), "keinverzeichnis")
	if err := os.WriteFile(unlesbar, []byte("x"), 0o600); err != nil {
		t.Fatalf("die Sperre anlegen: %v", err)
	}

	gefunden := msepaket.NeueBildabdruecke(unlesbar, einPaketMitBild(t, "da")).
		Abdruck("/probe.mse-style/rahmen.png")

	if strings.HasPrefix(gefunden, "fehlt:") {
		t.Errorf("die zweite Wurzel wurde nicht mehr gefragt: %q", gefunden)
	}

	// **Und ohne die zweite Wurzel steht der Grund im Wert.**
	geklagt := msepaket.NeueBildabdruecke(unlesbar).
		Abdruck("/probe.mse-style/rahmen.png")

	if !strings.HasPrefix(geklagt, "fehlt:") {
		t.Errorf("die unlesbare Wurzel gab %q", geklagt)
	}

	if geklagt == msepaket.NeueBildabdruecke(t.TempDir()).
		Abdruck("/probe.mse-style/rahmen.png") {
		t.Error("unlesbar und nicht vorhanden geben denselben Abdruck")
	}
}

// **Ein Setpaketverweis auf nichts ist ebenso ein Wert.**
func TestEinLeererSetpaketverweisIstEinWert(t *testing.T) {
	t.Parallel()

	abdruck := msepaket.NeueBildabdruecke().Abdruck(
		"setpaket:" + filepath.Join(t.TempDir(), "gibtesnicht.mse-set") +
			"!bild.png")

	if !strings.HasPrefix(abdruck, "fehlt:") {
		t.Errorf("der Abdruck %q sagt nicht, dass etwas fehlt", abdruck)
	}
}

// **`Hat` sagt, ob es die Datei gibt — ohne dass ein Aufrufer die Marke
// kennen muss.**
//
// # Warum es das Verfahren gibt
//
// **Weil `fehlt:` eine Sache dieses Pakets ist.** Ein Aufrufer, der den
// Vorsatz selbst prüfte, hinge an einer Zeichenkette, die hier steht — und am
// Tag, an dem sie sich ändert, fände er jede Datei. Gefragt wird deshalb
// hier, wo die Marke gesetzt wird (G2,
// #1459).
func TestHatSagtObEsDieDateiGibt(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "spiel.mse-game")

	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(paket, "da.png"), []byte("bytes"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	abdruecke := msepaket.NeueBildabdruecke(wurzel)

	if !abdruecke.Hat("/spiel.mse-game/da.png") {
		t.Error("eine vorhandene Datei gilt als fehlend")
	}

	if abdruecke.Hat("/spiel.mse-game/weg.png") {
		t.Error("eine fehlende Datei gilt als vorhanden")
	}

	// **Und ein Verweis, der gar keiner ist, fehlt ebenfalls** — statt einen
	// Fehler zu werfen, den niemand fängt.
	if abdruecke.Hat("kein Paketpfad") {
		t.Error("ein unsinniger Verweis gilt als vorhanden")
	}
}
