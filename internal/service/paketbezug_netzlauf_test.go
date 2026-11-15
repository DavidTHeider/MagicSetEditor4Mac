//go:build netzlauf

// **Der Paketbezug am Original** — #2049, Entscheidung 2 in
// `docs/INTEGRATIONEN.md`: bis hierher stand der Bezug in keiner Prüfliste,
// und ob er heute so holt wie gebaut, sagte kein Lauf.
//
// # Was er prüft
//
// **Den Weg der Paketseite**, über [service.Paketbezieher]: der kleinste der
// vier bekannten Packs (`Basic-M15-Magic-Pack`, rund 316 MB) am **gepinnten**
// Stand aus `bezug.Quellen()`, von `codeload.github.com` — Archiv, Prüfsumme
// beim Lesen, sicheres Auspacken, Einsetzen. **Und dass ein Zweig auch am
// Original kein Stand ist**: `main` schlägt GitHub als Marke nach und findet
// keine.
//
// # Warum er nicht in der Pipeline läuft
//
// **Er lädt über 300 MB aus dem Netz** — freigegeben am 2026-10-04 für
// #2049, in ein Wegwerfverzeichnis. In `make ci` geht kein Aufruf hinaus
// (§16).
//
//	go test -count=1 -tags netzlauf -run Original -v ./internal/service/
package service_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kleinsterPack ist die Quelle, gegen die der Lauf fährt.
func kleinsterPack(t *testing.T) bezug.Quelle {
	t.Helper()

	for _, quelle := range bezug.Quellen() {
		if quelle.Name == "Basic-M15-Magic-Pack" {
			return quelle
		}
	}

	t.Fatal("Basic-M15-Magic-Pack steht nicht mehr in bezug.Quellen()")

	return bezug.Quelle{}
}

func TestDerBezugAmOriginal(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ctx, abbrechen := context.WithTimeout(t.Context(), 20*time.Minute)
	defer abbrechen()

	quelle := kleinsterPack(t)
	wurzel := t.TempDir()

	anfang := time.Now()

	ergebnis, fehler := service.NeuerPaketbezieher(wurzel, "MagicSetEditor4Mac-Netzlauf").
		Beziehen(ctx, quelle.Adresse, quelle.Stand, quelle.Unterverzeichnis, quelle.Name)
	if fehler != nil {
		t.Fatalf("beziehen: %v", fehler)
	}

	dauer := time.Since(anfang)

	if ergebnis.Stand != quelle.Stand || !strings.HasPrefix(ergebnis.Pruefsumme, "sha256:") {
		t.Fatalf("Stand %q, Prüfsumme %q", ergebnis.Stand, ergebnis.Pruefsumme)
	}

	var dateien, stile int

	var bytes int64

	if fehler := filepath.WalkDir(ergebnis.Ort, func(_ string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil {
			return fehler
		}

		if eintrag.IsDir() && strings.HasSuffix(eintrag.Name(), ".mse-style") {
			stile++
		}

		if eintrag.Type().IsRegular() {
			info, infofehler := eintrag.Info()
			if infofehler != nil {
				return infofehler
			}

			dateien++
			bytes += info.Size()
		}

		return nil
	}); fehler != nil {
		t.Fatalf("den Baum lesen: %v", fehler)
	}

	// **Die Schriften kommen mit** (#263): sie liegen neben `data`, und ein
	// Bezug nur von `data` gäbe Vorlagen ohne ihre Schriften.
	if _, fehler := os.Stat(filepath.Join(ergebnis.Ort, "Magic - Fonts")); fehler != nil {
		t.Errorf("ohne »Magic - Fonts«: %v", fehler)
	}

	if stile == 0 {
		t.Error("kein einziger Stil im Pack")
	}

	t.Logf("Stand %s, Prüfsumme %s, %d Dateien, %d Stile, %.1f MB ausgepackt, %d übergangen, %.1f s",
		ergebnis.Stand, ergebnis.Pruefsumme, dateien, stile, float64(bytes)/1e6,
		len(ergebnis.Uebergangen), dauer.Seconds())

	for _, uebergangen := range ergebnis.Uebergangen {
		t.Logf("übergangen: %s", uebergangen)
	}
}

func TestEinZweigIstAmOriginalKeinStand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := kleinsterPack(t)

	_, fehler := service.NeuerPaketbezieher(t.TempDir(), "MagicSetEditor4Mac-Netzlauf").
		Beziehen(t.Context(), quelle.Adresse, "main", quelle.Unterverzeichnis, quelle.Name)

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "stand" {
		t.Fatalf("erwartet ein Befund am Stand, bekam %v", fehler)
	}

	t.Logf("GitHub kennt »main« nicht als Freigabe: %v", fehler)
}
