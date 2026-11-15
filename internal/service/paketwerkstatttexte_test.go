package service_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Texte eines Pakets sind seine bearbeitbaren Dateien samt Inhalt** —
// der Stoff, aus dem der Symbolindex entsteht (U7, #1418). Ein Bild gehört
// nicht dazu: es ist kein Text, und `Bearbeitbar` sagt das schon.
func TestDieTexteSindDieBearbeitbarenDateienMitInhalt(t *testing.T) {
	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	texte, err := werk.Texte("probe.mse-style")
	if err != nil {
		t.Fatalf("Texte: %v", err)
	}

	sort.Slice(texte, func(i, j int) bool { return texte[i].Name < texte[j].Name })

	erwartet := []service.Paketdateitext{
		{Name: "skript.txt", Inhalt: "1 + 1\n"},
		{Name: "style", Inhalt: "mse version: 2.0.0\ngame: probe\n"},
	}

	if !reflect.DeepEqual(texte, erwartet) {
		t.Errorf("gefunden:\n%+v\nerwartet:\n%+v", texte, erwartet)
	}
}

// **Ein fehlendes Paket meldet sich** — derselbe Fehler wie beim Aufzählen.
func TestDieTexteEinesFehlendenPaketsMeldenSich(t *testing.T) {
	werk := service.NeuePaketwerkstatt(t.TempDir())

	if _, err := werk.Texte("gibt-es-nicht.mse-style"); err == nil {
		t.Error("ein fehlendes Paket gab Texte")
	}

	if _, err := werk.Texte(""); err == nil {
		t.Error("ohne Paket kamen Texte")
	}
}

// **Eine Datei, die sich nicht lesen lässt, fehlt im Index — und nimmt den
// anderen nichts** (§8). Ein enges Verzeichnis dagegen scheitert schon beim
// Aufzählen, und das wird gemeldet.
func TestEineUnlesbareDateiFehltImIndexUndEinEngesPaketMeldetSich(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root liest sich jede Datei")
	}

	wurzel := werkwurzel(t)
	paket := filepath.Join(wurzel, "probe.mse-style")
	eng := filepath.Join(paket, "skript.txt")

	if err := os.Chmod(eng, 0o000); err != nil {
		t.Fatalf("engen: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(eng, 0o600) }) //nolint:errcheck,gosec // Aufräumen

	texte, err := service.NeuePaketwerkstatt(wurzel).Texte("probe.mse-style")
	if err != nil {
		t.Fatalf("Texte: %v", err)
	}

	if len(texte) != 1 || texte[0].Name != "style" {
		t.Errorf("gefunden: %+v", texte)
	}

	bilder := filepath.Join(paket, "bilder")

	if chmodfehler := os.Chmod(bilder, 0o000); chmodfehler != nil {
		t.Fatalf("engen: %v", chmodfehler)
	}

	t.Cleanup(func() { _ = os.Chmod(bilder, 0o700) }) //nolint:errcheck,gosec // Aufräumen

	if _, fehler := service.NeuePaketwerkstatt(wurzel).Texte("probe.mse-style"); fehler == nil {
		t.Error("das enge Paket gab Texte")
	}
}
