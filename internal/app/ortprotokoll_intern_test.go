// **Ein Ort, der sich nicht bestimmen lässt, hält den Start nicht auf — und
// verschwindet auch nicht** —
// #857.
//
// Bis zum 2026-08-22 stand für das Bilderverzeichnis ein `_`: wer später ein
// Bild einsetzte und eine unverständliche Meldung bekam, fand nirgends den
// Grund. Der Fehler des Paketverzeichnisses wurde aufgehoben, aber erst beim
// ersten Bezug gemeldet.
//
// **Ein interner Test**, weil beide Orte Nähte am Verbund sind und keine
// ausgestellte Naht: `os.UserHomeDir` lässt sich in einem Test nicht
// wegnehmen, und eine exportierte Setzmethode dafür wäre eine Schnittstelle,
// die nur der Test benutzt.
package app

import (
	"bytes"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
)

func TestEinUnbestimmbarerOrtStehtImProtokoll(t *testing.T) {
	var gesammelt bytes.Buffer

	vorher := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&gesammelt, nil)))

	t.Cleanup(func() { slog.SetDefault(vorher) })

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.bilderort = func() (string, error) {
		return "", errors.New("kein Bilderverzeichnis")
	}
	anwendung.paketort = func() (string, error) {
		return "", errors.New("kein Paketverzeichnis")
	}

	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	for _, teil := range []string{"kein Bilderverzeichnis", "kein Paketverzeichnis"} {
		if !strings.Contains(gesammelt.String(), teil) {
			t.Errorf("%q fehlt im Protokoll:\n%s", teil, gesammelt.String())
		}
	}
}
