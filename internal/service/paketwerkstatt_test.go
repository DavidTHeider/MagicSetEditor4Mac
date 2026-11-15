// **Eine Vorlage öffnen, ändern, speichern** — Werkstück 2 von Phase 3.3
// (#1164).
//
// # Was hier entschieden wird und nicht in `msepaket`
//
// Drei Dinge, und alle drei sind fachlich:
//
//   - **Welche Dateien bearbeitbar sind.** Ein `.png` ist keine Vorlage.
//   - **Dass ein Paket nur unter der Wurzel liegen darf.** Ein Pfad von außen
//     ist keine Paketwahl.
//   - **Woran ein Konflikt erkannt wird** — der Stand, den der Aufrufer beim
//     Lesen bekommen hat.
package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vorlagenwurzel legt ein Verzeichnis mit einem Paket an.
func werkwurzel(t *testing.T) string {
	t.Helper()

	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe.mse-style")

	if err := os.MkdirAll(filepath.Join(paket, "bilder"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for name, inhalt := range map[string]string{
		"style":             "mse version: 2.0.0\ngame: probe\n",
		"skript.txt":        "1 + 1\n",
		"bilder/rahmen.png": "\x89PNG\r\n\x1a\n",
	} {
		if err := os.WriteFile(filepath.Join(paket, filepath.FromSlash(name)),
			[]byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	return wurzel
}

// **Aufgezählt wird, was sich bearbeiten lässt — und sonst nichts.**
//
// # Woher die Regel kommt
//
// Gezählt über die 526 Pakete von `full-magic-pack` am 2026-08-26: MSEs eigene
// Datendateien tragen **keine Endung** — `style`, `game`, `card_fields`,
// `include`, `locale`, 777 Dateien unter 133 verschiedenen Namen. Alles mit
// Endung ist Beiwerk: 30.245 `.png`, 3.242 `.jpg`, 55 `.ttf`, dazu 31 `.txt`.
//
// **Keine der 777 und keine der 31 trägt ein Nullbyte** — sie sind Text. Das
// ist die Reichweite dieser Messung (§10): sie gilt für dieses Pack, nicht für
// „jedes Paket".
func TestAufgezaehltWirdWasSichBearbeitenLaesst(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	dateien, fehler := werk.Dateien("probe.mse-style")
	if fehler != nil {
		t.Fatalf("Dateien: %v", fehler)
	}

	if !slices.Contains(dateien, "style") || !slices.Contains(dateien, "skript.txt") {
		t.Errorf("es fehlt etwas: %v", dateien)
	}

	if slices.Contains(dateien, "bilder/rahmen.png") {
		t.Errorf("das Bild steht in der Liste: %v", dateien)
	}
}

// **Gelesen kommt der Inhalt und ein Stand.**
func TestGelesenKommtInhaltUndStand(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	inhalt, stand, fehler := werk.Lesen("probe.mse-style", "style")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if inhalt != "mse version: 2.0.0\ngame: probe\n" {
		t.Errorf("Inhalt %q", inhalt)
	}

	if stand == "" {
		t.Error("der Stand fehlt")
	}
}

// **Geschrieben wird, was gelesen wurde — und der neue Stand kommt zurück.**
//
// **Der neue Stand ist die Zusage für den nächsten Zug**: ohne ihn müsste das
// Fenster nach jedem Speichern neu lesen, und zwischen Antwort und Lesen läge
// wieder ein Fenster (§7, Regel 2).
func TestGeschriebenWirdMitNeuemStand(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	_, stand, fehler := werk.Lesen("probe.mse-style", "style")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	neu := "mse version: 2.0.0\ngame: probe\nfull name: Probe\n"

	neuerStand, schreibfehler := werk.Schreiben(t.Context(), "probe.mse-style", "style", neu, stand)
	if schreibfehler != nil {
		t.Fatalf("Schreiben: %v", schreibfehler)
	}

	if neuerStand == stand {
		t.Error("der Stand blieb derselbe")
	}

	nachher, _, lesefehler := werk.Lesen("probe.mse-style", "style")
	if lesefehler != nil {
		t.Fatalf("Lesen: %v", lesefehler)
	}

	if nachher != neu {
		t.Errorf("gelesen %q", nachher)
	}

	// **Und der zurückgegebene Stand trägt den nächsten Zug.** Ohne diese
	// Zusicherung wäre er eine Zahl, die niemand prüft.
	if _, weiter := werk.Schreiben(t.Context(), "probe.mse-style", "style", "x\n", neuerStand); weiter != nil {
		t.Errorf("der zurückgegebene Stand trug nicht: %v", weiter)
	}
}

// **Ein veralteter Stand wird abgewiesen.**
func TestEinVeralteterStandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	_, stand, fehler := werk.Lesen("probe.mse-style", "style")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if _, err := werk.Schreiben(t.Context(), "probe.mse-style", "style", "a\n", stand); err != nil {
		t.Fatalf("erstes Schreiben: %v", err)
	}

	_, zweites := werk.Schreiben(t.Context(), "probe.mse-style", "style", "b\n", stand)
	if !errors.Is(zweites, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", zweites)
	}
}

// **Ein Paketname, der aus der Wurzel führt, wird abgewiesen — überall.**
//
// # Warum das der wichtigste Lauf dieser Datei ist
//
// Der Name kommt aus dem Fenster, von der Kommandozeile, aus einem
// MCP-Aufruf und über die Rückschleife. `../../../.ssh/id_ed25519` ist dort
// kein theoretischer Fall — §16 nennt das Härten der Pfade ausdrücklich.
//
// **Alle drei Wege**, denn ein Schutz an zweien ist keiner.
func TestEinAusbrechenderPaketnameWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	// **Ein Unterverzeichnis ist kein Ausbruch** — bezogene Sammlungen liegen
	// so (`M15-Magic-Pack/data/magic-m15.mse-style`), und wer sie abwiese,
	// erreichte kein einziges installiertes Paket. Geprüft wird, dass das
	// Ergebnis **unter der Wurzel** bleibt.
	for _, name := range []string{
		"../draussen.mse-style",
		"unter/../../draussen.mse-style",
		"..",
		"",
		"/etc/passwd",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := werk.Dateien(name); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Dateien: %v — will ErrValidation", fehler)
			}

			if _, _, fehler := werk.Lesen(name, "style"); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Lesen: %v — will ErrValidation", fehler)
			}

			if _, fehler := werk.Schreiben(t.Context(), name, "style", "x", ""); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Schreiben: %v — will ErrValidation", fehler)
			}
		})
	}
}

// **Ein Paket in einem Unterverzeichnis wird gefunden.**
//
// # Warum das ein eigener Lauf ist
//
// Weil der erste Anlauf einen **reinen Namen** verlangte und damit kein
// einziges installiertes Paket erreichte: bezogen wird eine Sammlung, und
// darin liegt `M15-Magic-Pack/data/magic-m15.mse-style`. Gefunden bei der
// Sichtprüfung im laufenden Fenster — 327 Pakete in der Auswahl, und bei jedem
// stand „hat keine Datei zum Bearbeiten".
func TestEinPaketImUnterverzeichnisWirdGefunden(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	tief := filepath.Join("sammlung", "data", "probe.mse-style")

	if err := os.MkdirAll(filepath.Join(wurzel, tief), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(wurzel, tief, "style"),
		[]byte("mse version: 2.0.0\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	werk := service.NeuePaketwerkstatt(wurzel)

	// **Beide Formen**: das Fenster reicht den vollen Pfad aus dem
	// Paketkatalog weiter, die Kommandozeile tippt den relativen.
	for name, angabe := range map[string]string{
		"relativ": filepath.ToSlash(tief),
		"absolut": filepath.Join(wurzel, tief),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dateien, fehler := werk.Dateien(angabe)
			if fehler != nil {
				t.Fatalf("Dateien: %v", fehler)
			}

			if !slices.Contains(dateien, "style") {
				t.Errorf("style fehlt in %v", dateien)
			}
		})
	}
}

// **Und eine Datei, die nicht bearbeitbar ist, wird nicht geschrieben.**
//
// Die Gegenprobe zur Aufzählung: ohne sie wäre „bearbeitbar" eine Angabe für
// die Anzeige, und wer den Namen kennt, schriebe trotzdem ins Bild.
func TestEinBildWirdNichtBeschrieben(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	if _, _, fehler := werk.Lesen("probe.mse-style", "bilder/rahmen.png"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Lesen: %v — will ErrValidation", fehler)
	}

	if _, fehler := werk.Schreiben(t.Context(), "probe.mse-style", "bilder/rahmen.png", "x", ""); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Schreiben: %v — will ErrValidation", fehler)
	}
}

// **Ein Paket, das es nicht gibt, meldet sich auf jedem Weg.**
func TestEinFehlendesPaketMeldetSich(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(t.TempDir())

	if _, fehler := werk.Dateien("nichts.mse-style"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Dateien: %v — will ErrNotFound", fehler)
	}

	if _, _, fehler := werk.Lesen("nichts.mse-style", "style"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Lesen: %v — will ErrNotFound", fehler)
	}
}

// **Und eine Datei, die es im Paket nicht gibt, ebenso.**
func TestEineFehlendeVorlageMeldetSich(t *testing.T) {
	t.Parallel()

	werk := service.NeuePaketwerkstatt(werkwurzel(t))

	if _, _, fehler := werk.Lesen("probe.mse-style", "gibtesnicht"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Lesen: %v — will ErrNotFound", fehler)
	}
}

// **Ein Paket, dessen Dateien sich nicht aufzählen lassen, meldet es.**
//
// Ohne Leserecht am Verzeichnis geht das Öffnen noch, das Gehen nicht. Der
// Fall ist der eines Pakets, das jemand eng abgelegt hat.
func TestEinEngesPaketMeldetSichBeimAufzaehlen(t *testing.T) {
	t.Parallel()

	wurzel := werkwurzel(t)
	eng := filepath.Join(wurzel, "probe.mse-style", "bilder")

	if err := os.Chmod(eng, 0o000); err != nil {
		t.Fatalf("engen: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(eng, 0o700) }) //nolint:errcheck,gosec // Aufräumen

	if _, fehler := service.NeuePaketwerkstatt(wurzel).Dateien("probe.mse-style"); fehler == nil {
		t.Error("das enge Paket zählte sich auf")
	}
}
