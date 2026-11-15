//go:build darwin

// **Der Dialog um das Diagnosepaket** — Werkstück O10 von Phase 5.2
// (#1434).
package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Endung wird angehängt, nicht verlangt** — wer im Dialog einen Namen
// tippt, tippt selten die Endung mit.
func TestDerDialogHaengtDieEndungAn(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	ordner := t.TempDir()

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return filepath.Join(ordner, "bericht"), nil
	})

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	if antwort := anwendung.DiagnosepaketSichern(t.Context(), nil); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ordner, "bericht"+domain.Diagnosepaketendung)); fehler != nil {
		t.Errorf("das Paket liegt nicht unter der Endung: %v", fehler)
	}
}

// **Abgebrochen ist kein Fehler** — und es entsteht nichts.
func TestEinAbgebrochenerPaketdialogSchreibtNichts(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	ordner := t.TempDir()

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", nil
	})

	if vorschau := anwendung.Diagnosepaket(t.Context()); vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	if antwort := anwendung.DiagnosepaketSichern(t.Context(), nil); antwort.Fehler != nil {
		t.Fatalf("der Abbruch wurde als Fehler gemeldet: %+v", antwort.Fehler)
	}

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(eintraege) != 0 {
		t.Errorf("%d Einträge — der Abbruch hat geschrieben", len(eintraege))
	}
}

// **Und ein Fehler des Dialogs kommt an** (§8).
func TestEinFehlerDesDialogsKommtAn(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("der Dialog klemmt")
	})

	if antwort := anwendung.DiagnosepaketSichern(t.Context(), nil); antwort.Fehler == nil {
		t.Error("der Fehler des Dialogs ging verloren")
	}
}

// **Ein zu großes Paket entsteht gar nicht erst** — die Grenze greift vor dem
// Schreiben, nicht danach.
func TestEinZuGrossesPaketWirdAbgewiesen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)

	// **Ein Protokoll über der Grenze.** Zweihundert Zeilen kommen ins Paket
	// (`app.Protokollzeilen`); bei 64 KB je Zeile sind das 12,8 MB — mehr als
	// die acht, die `domain.Diagnosepaketgrenze` zulässt.
	pfad, fehler := app.Protokollpfad()
	if fehler != nil {
		t.Fatalf("Protokollpfad: %v", fehler)
	}

	zeile := strings.Repeat("x", 64<<10) + "\n"
	if schreibfehler := os.WriteFile(pfad,
		[]byte(strings.Repeat(zeile, 210)), 0o600); schreibfehler != nil {
		t.Fatalf("schreiben: %v", schreibfehler)
	}

	vorschau := anwendung.Diagnosepaket(t.Context())

	switch {
	case vorschau.Fehler == nil:
		t.Fatalf("das zu große Paket kam durch: %d Byte", vorschau.Bytes)
	case vorschau.Bytes <= domain.Diagnosepaketgrenze:
		t.Errorf("die Vorschau meldet %d Byte — das ist unter der Grenze", vorschau.Bytes)
	}

	// **Und geschrieben wird trotzdem nichts** — es ist nichts gemerkt.
	ziel := filepath.Join(t.TempDir(), "probe"+domain.Diagnosepaketendung)
	if antwort := anwendung.DiagnosepaketNach(t.Context(), ziel); antwort.Fehler == nil {
		t.Error("nach dem abgewiesenen Zusammenstellen entstand ein Paket")
	}
}
