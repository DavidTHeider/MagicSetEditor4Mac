package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein eingesetztes Bild kommt beim Speichern mit** —
// #246,
// entschieden in
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md).
//
// # Wofür
//
// Ohne diese Naht wäre das eingesetzte Bild nur auf **diesem** Rechner zu
// sehen: das gespeicherte Set nennt `image1`, und die Datei bliebe im
// Beiwerkverzeichnis zurück. Weitergegeben öffnete es mit leerem Kunstfenster
// — genau der Fall, den `beiwerk` für das Herkunftspaket schon abwendet.
//
// # Und die Ablage schlägt das Herkunftspaket
//
// Wer ein Bild einsetzt, ersetzt damit das, was vorher dastand. Käme das Paket
// zuletzt, überschriebe es die neue Datei mit der alten — und das Speichern
// nähme die Änderung zurück, die es sichern soll.
func TestEinEingesetztesBildKommtMit(t *testing.T) {
	t.Parallel()

	// **Ein Paketverzeichnis heißt wie ein Paket.** `msepaket.Oeffne` erkennt
	// ein Verzeichnis an seiner Endung; ein nackter Ordner ist keines.
	herkunft := filepath.Join(t.TempDir(), "Mein Set.mse-set")
	ablage := t.TempDir()
	ziel := t.TempDir()

	if err := os.MkdirAll(herkunft, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	schreibe(t, filepath.Join(herkunft, "set"), "mse version: 2.0.0")
	schreibe(t, filepath.Join(herkunft, "image1"), "das alte Bild")
	schreibe(t, filepath.Join(herkunft, "image2"), "ein Bild, das bleibt")
	schreibe(t, filepath.Join(ablage, "image1"), "das neue Bild")
	schreibe(t, filepath.Join(ablage, "image3"), "ein neues Bild")

	if befunde := service.Beiwerk(herkunft, ablage, "set", ziel); len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	faelle := []struct {
		datei, inhalt, warum string
	}{
		{"image1", "das neue Bild", "die Ablage schlägt das Herkunftspaket"},
		{"image2", "ein Bild, das bleibt", "was nur im Paket liegt, kommt von dort"},
		{"image3", "ein neues Bild", "und was nur in der Ablage liegt, von dort"},
	}

	for _, fall := range faelle {
		t.Run(fall.datei, func(t *testing.T) {
			inhalt, err := os.ReadFile(filepath.Join(ziel, fall.datei)) //nolint:gosec // Testpfad
			if err != nil {
				t.Fatalf("%s fehlt: %v", fall.datei, err)
			}

			if string(inhalt) != fall.inhalt {
				t.Errorf("%q statt %q — %s", inhalt, fall.inhalt, fall.warum)
			}
		})
	}

	// **Die Hauptdatei kommt nicht mit.** Sie wird gleich danach geschrieben,
	// und zwar aus dem Bestand — die alte hierher zu kopieren hieße, sie
	// zweimal zu schreiben und die falsche zuletzt.
	if _, err := os.Stat(filepath.Join(ziel, "set")); err == nil {
		t.Error("die Hauptdatei ist mitgekommen")
	}
}

// **Ohne Ablage bleibt es beim Herkunftspaket** — der Weg, den es bisher gab.
func TestOhneAblageKommtNurDasPaketMit(t *testing.T) {
	t.Parallel()

	herkunft := filepath.Join(t.TempDir(), "Mein Set.mse-set")
	ziel := t.TempDir()

	if err := os.MkdirAll(herkunft, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	schreibe(t, filepath.Join(herkunft, "set"), "mse version: 2.0.0")
	schreibe(t, filepath.Join(herkunft, "image1"), "aus dem Paket")

	if befunde := service.Beiwerk(herkunft, "", "set", ziel); len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	inhalt, err := os.ReadFile(filepath.Join(ziel, "image1")) //nolint:gosec // Testpfad
	if err != nil {
		t.Fatalf("image1 fehlt: %v", err)
	}

	if string(inhalt) != "aus dem Paket" {
		t.Errorf("%q", inhalt)
	}
}

// **Und ohne Herkunftspaket kommt die Ablage allein.** Ein Set, das in dieser
// Anwendung entstanden ist, hat kein Paket — seine Bilder gibt es trotzdem.
func TestOhnePaketKommtDieAblageAllein(t *testing.T) {
	t.Parallel()

	ablage := t.TempDir()
	ziel := t.TempDir()

	schreibe(t, filepath.Join(ablage, "image1"), "eingesetzt")

	if befunde := service.Beiwerk("", ablage, "set", ziel); len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if _, err := os.Stat(filepath.Join(ziel, "image1")); err != nil {
		t.Errorf("image1 fehlt: %v", err)
	}
}

func schreibe(t *testing.T, pfad, inhalt string) {
	t.Helper()

	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// **Ein Verzeichnis, das sich nicht lesen lässt, wird gemeldet** — und die
// Hauptdatei kommt trotzdem heraus (§8).
//
// **Aber nicht, dass es fehlt.** In den allermeisten Sets hat niemand ein Bild
// eingesetzt; ein Befund je Speichern wäre eine Meldung über den Normalfall.
func TestEineUnlesbareAblageWirdGemeldetUndEineFehlendeNicht(t *testing.T) {
	t.Parallel()

	ziel := t.TempDir()

	t.Run("fehlend ist kein Befund", func(t *testing.T) {
		fehlt := filepath.Join(t.TempDir(), "gibtsnicht")

		if befunde := service.Beiwerk("", fehlt, "set", ziel); len(befunde) > 0 {
			t.Errorf("Befunde: %v", befunde)
		}
	})

	t.Run("unlesbar ist einer", func(t *testing.T) {
		// **Eine Datei statt eines Verzeichnisses.** `ReadDir` scheitert daran
		// auf jeder Plattform, anders als an einem Rechtebit.
		keinVerzeichnis := filepath.Join(t.TempDir(), "ablage")
		schreibe(t, keinVerzeichnis, "keine Ablage")

		befunde := service.Beiwerk("", keinVerzeichnis, "set", ziel)
		if len(befunde) != 1 {
			t.Fatalf("Befunde: %v", befunde)
		}

		if !strings.Contains(befunde[0].Meldung, "Bildablage") {
			t.Errorf("%q", befunde[0].Meldung)
		}
	})
}

// **Eine Datei, die sich nicht kopieren lässt, ist ein Befund je Datei** — die
// anderen kommen mit. Eine Teilliste ist besser als keine (§8).
func TestEineUnkopierbareDateiIstEinBefundJeDatei(t *testing.T) {
	t.Parallel()

	ablage := t.TempDir()
	ziel := t.TempDir()

	schreibe(t, filepath.Join(ablage, "image1"), "geht")

	// **Ein Unterverzeichnis wird übergangen.** [service.Bildablage] legt flach
	// ab; eines darin käme von außen, und wir kopierten dann etwas, das
	// niemand hier hingelegt hat.
	if err := os.MkdirAll(filepath.Join(ablage, "unterordner"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Und ein Ziel, das schon ein Verzeichnis ist**, lässt sich nicht
	// überschreiben — derselbe Weg wie eine volle Platte, nur reproduzierbar.
	schreibe(t, filepath.Join(ablage, "image2"), "geht nicht")

	if err := os.MkdirAll(filepath.Join(ziel, "image2"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	befunde := service.Beiwerk("", ablage, "set", ziel)
	if len(befunde) != 1 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if befunde[0].Ort != "image2" {
		t.Errorf("der Befund nennt %q", befunde[0].Ort)
	}

	if _, err := os.Stat(filepath.Join(ziel, "image1")); err != nil {
		t.Errorf("die brauchbare Datei ist mit ausgefallen: %v", err)
	}

	if _, err := os.Stat(filepath.Join(ziel, "unterordner")); err == nil {
		t.Error("ein Unterverzeichnis ist mitgekommen")
	}
}

// **Eine Datei, die es nicht mehr gibt, wird gemeldet.**
//
// Zwischen `ReadDir` und `ReadFile` liegt ein Augenblick; wer in dieser Zeit
// die Ablage aufräumt, bekommt einen Befund und kein stilles Fehlen.
func TestEineVerschwundeneDateiWirdGemeldet(t *testing.T) {
	t.Parallel()

	if fehler := service.KopiereDatei(
		filepath.Join(t.TempDir(), "gibtsnicht"),
		filepath.Join(t.TempDir(), "egal"),
	); fehler == nil {
		t.Error("die fehlende Datei kam durch")
	}
}

// **Der Setschreiber gibt weiter, was in der Ablage liegt.**
//
// Ohne diese Naht bliebe ein eingesetztes Bild auf **diesem** Rechner: das
// gespeicherte Set nennt `image1`, und die Datei stünde im
// Beiwerkverzeichnis zurück.
func TestDerSetschreiberGibtDieAblageWeiter(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	// Das Verzeichnis der Ablage anlegen und ein Bild hineinlegen.
	if _, err := ablage.Ablegen(1, "kunst", probebild()); err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	ziel := filepath.Join(t.TempDir(), "woanders.mse-set")
	schreiber := service.NeuerSetschreiber(
		&archivleser{herkunft: gepacktesSet(t)}, &ortsschreiber{}).MitBildablage(ablage)

	if _, err := schreiber.Schreibe(t.Context(), 1, ziel); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := os.Stat(filepath.Join(ziel, "kunst1")); err != nil {
		t.Errorf("das eingesetzte Bild ist nicht mitgekommen: %v", err)
	}
}

// **Und ohne Ablage bleibt alles, wie es war** — der Weg, den es bisher gab.
func TestOhneAblageSchreibtDerSetschreiberWieBisher(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "woanders.mse-set")
	schreiber := service.NeuerSetschreiber(
		&archivleser{herkunft: gepacktesSet(t)}, &ortsschreiber{})

	if _, err := schreiber.Schreibe(t.Context(), 1, ziel); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := os.Stat(filepath.Join(ziel, "set")); err != nil {
		t.Errorf("die Hauptdatei fehlt: %v", err)
	}
}
