package service

// **Der Abdruck sieht, was das Zeichnen liest — und aus demselben Gang** —
// #2423, Nachlauf über Phase 3 von 0.9.0, Befunde 1 und 2.
//
// Bis dahin faltete [Stilspeicher.Ressourcenstand] die Pakete der Setvorlage
// und die Bildablage. **Es fehlten** die Pakete der Symbolschriften — sie lädt
// ein eigener Leser — und die eigenen Stile der Karten. **Und ein frischer
// Abdruck konnte eine alte Vorlage beglaubigen**: innerhalb der
// [stempelpause] nahm das Zeichnen die gemerkte, der Abdruck ging frisch.
// Jede der Lagen hieß „aktuell".

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kartenstilprobe nennt die eigenen Stile der Karten eines Sets.
type kartenstilprobe struct{ stile []string }

func (k kartenstilprobe) Kartenstile(context.Context, int64) ([]string, error) { return k.stile, nil }

// wandernderKopf führt Set 7 und lässt seine Herkunft ändern.
type wandernderKopf struct {
	sperre   sync.Mutex
	herkunft string
}

func (w *wandernderKopf) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	w.sperre.Lock()
	defer w.sperre.Unlock()

	return []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe", Origin: w.herkunft}}, nil
}

func (*wandernderKopf) Setwerte(context.Context, int64) ([]domain.Value, error) { return nil, nil }

// ganzerStempelplatz legt neben Spiel und Stil eine Symbolschrift und einen
// zweiten Stil als Kartenstil — und lässt die Setvorlage die Schrift nennen.
func ganzerStempelplatz(t *testing.T) (wurzel string, speicher *Stilspeicher) {
	t.Helper()

	wurzel, speicher = stempelplatz(t)

	for quelle, ziel := range map[string]string{
		filepath.Join("..", "..", "testdaten", "probe-symbole.mse-symbol-font"): "probe-symbole.mse-symbol-font",
		filepath.Join("..", "..", "saat", "probe-schlicht.mse-style"):           "probe-zweit.mse-style",
	} {
		if err := os.CopyFS(filepath.Join(wurzel, ziel), os.DirFS(quelle)); err != nil {
			t.Fatal(err)
		}
	}

	zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"),
		"\t\t\tcolor: black\n", "\t\t\tcolor: black\n\t\tsymbol font:\n\t\t\tname: probe-symbole\n\t\t\tsize: 12\n")

	return wurzel, speicher
}

func TestDerRessourcenstandSiehtDieSymbolschrift(t *testing.T) {
	wurzel, speicher := ganzerStempelplatz(t)

	vorher := speicher.Ressourcenstand(t.Context(), 7)

	zeileErsetzen(t, filepath.Join(wurzel, "probe-symbole.mse-symbol-font", "symbol-font"),
		"image font size: 12", "image font size: 13")

	if speicher.Ressourcenstand(t.Context(), 7) == vorher {
		t.Error("eine geänderte Datei in der Symbolschrift blieb ungesehen")
	}
}

func TestDerRessourcenstandSiehtDenKartenstil(t *testing.T) {
	wurzel, speicher := ganzerStempelplatz(t)
	speicher.MitKartenstilen(kartenstilprobe{stile: []string{"zweit"}})

	vorher := speicher.Ressourcenstand(t.Context(), 7)

	zeileErsetzen(t, filepath.Join(wurzel, "probe-zweit.mse-style", "style"), "card width: 375", "card width: 380")

	if speicher.Ressourcenstand(t.Context(), 7) == vorher {
		t.Error("eine geänderte Datei im Kartenstil blieb ungesehen")
	}
}

// **Der Abdruck geht nicht an der Pause vorbei**: was er beglaubigt, zeichnet
// der nächste Ruf auch — sonst stünde ein neuer Abdruck über Befunden aus der
// alten Vorlage.
func TestDerAbdruckUndDasZeichnenSehenDieselbeVorlage(t *testing.T) {
	wurzel, speicher := ganzerStempelplatz(t)
	stempelpause = time.Hour

	if vorlage, _, _ := speicher.StilFuerSet(t.Context(), 7); vorlage.Stil.Breite != 375 {
		t.Fatalf("die Breite ist %d", vorlage.Stil.Breite)
	}

	vorher := speicher.Ressourcenstand(t.Context(), 7)

	zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"), "card width: 375", "card width: 376")

	if speicher.Ressourcenstand(t.Context(), 7) == vorher {
		t.Fatal("der Abdruck blieb, obwohl sich der Stil änderte")
	}

	if vorlage, _, _ := speicher.StilFuerSet(t.Context(), 7); vorlage.Stil.Breite != 376 {
		t.Errorf("der Abdruck ist neu, gezeichnet würde mit der Breite %d", vorlage.Stil.Breite)
	}
}

// **Eine andere Herkunft macht die gemerkte Vorlage ungültig** — auch in der
// Pause: an ihr hängt, wo die Bilder des Sets liegen.
func TestEineAndereHerkunftGiltNichtAlsGemerkt(t *testing.T) {
	_, speicher := stempelplatz(t)
	stempelpause = time.Hour

	kopf := &wandernderKopf{herkunft: "/alt/probe.mse-set"}
	speicher.koepfe = kopf

	if vorlage, _, _ := speicher.StilFuerSet(t.Context(), 7); vorlage.Setpaket != "/alt/probe.mse-set" {
		t.Fatalf("das Setpaket ist %q", vorlage.Setpaket)
	}

	kopf.sperre.Lock()
	kopf.herkunft = "/neu/probe.mse-set"
	kopf.sperre.Unlock()

	if vorlage, _, _ := speicher.StilFuerSet(t.Context(), 7); vorlage.Setpaket != "/neu/probe.mse-set" {
		t.Errorf("gezeichnet würde aus %q", vorlage.Setpaket)
	}
}

// kaputteKartenstile kann die Kartenstile nicht lesen.
type kaputteKartenstile struct{}

func (kaputteKartenstile) Kartenstile(context.Context, int64) ([]string, error) {
	return nil, os.ErrPermission
}

// **Ein Lesefehler steht im Stand** — und der nächste Abdruck, der lesen
// kann, ist ein anderer; ein still verkürzter Stand bliebe gleich.
func TestKartenstileDieNichtZuLesenSindStehenImStand(t *testing.T) {
	_, speicher := ganzerStempelplatz(t)
	speicher.MitKartenstilen(kaputteKartenstile{})

	gescheitert := speicher.Ressourcenstand(t.Context(), 7)

	speicher.MitKartenstilen(kartenstilprobe{stile: []string{" ", ""}})

	if gelesen := speicher.Ressourcenstand(t.Context(), 7); gelesen == gescheitert {
		t.Errorf("ein gescheitertes Lesen gab denselben Stand: %q", gelesen)
	}
}

// **Eine Symbolschrift, die es nicht gibt, steht als fehlend im Stand** —
// entsteht sie, ist es ein anderer.
func TestEineFehlendeSymbolschriftZaehltAlsFehlend(t *testing.T) {
	wurzel, speicher := ganzerStempelplatz(t)

	if err := os.RemoveAll(filepath.Join(wurzel, "probe-symbole.mse-symbol-font")); err != nil {
		t.Fatal(err)
	}

	ohne := speicher.Ressourcenstand(t.Context(), 7)

	if err := os.CopyFS(filepath.Join(wurzel, "probe-symbole.mse-symbol-font"),
		os.DirFS(filepath.Join("..", "..", "testdaten", "probe-symbole.mse-symbol-font"))); err != nil {
		t.Fatal(err)
	}

	if speicher.Ressourcenstand(t.Context(), 7) == ohne {
		t.Error("eine entstandene Symbolschrift blieb ungesehen")
	}
}
