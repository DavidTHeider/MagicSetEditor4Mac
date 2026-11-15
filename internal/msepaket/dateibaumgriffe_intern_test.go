// **Die Fehlerwege hinter den Griffen** — Nachbardatei, Schreiben und Tausch
// scheitern auf einer gesunden Platte nicht; hier werden sie ersetzt, mit
// denselben Abbruchstellen wie beim Schreibweg (`ablegeabbruch_intern_test.go`).
package msepaket

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func dateibaumprobe(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatal(err)
	}

	for name, inhalt := range map[string]string{"style": "mse version: 2.0.0\n", "skript": "1\n"} {
		if err := os.WriteFile(filepath.Join(ort, name), []byte(inhalt), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return ort
}

func TestDieDateioperationenScheiternAnDenGriffen(t *testing.T) {
	t.Parallel()

	ohneTausch := echteHilfen()
	ohneTausch.umbenennen = func(string, string) error { return errAbbruch }

	for name, fall := range map[string]func(ort string) error{
		"Anlegen ohne Nachbardatei": func(ort string) error {
			return dateiAnlegen(context.Background(), ort, "neu", nil, hilfenMitAbbruch(beimAnlegen))
		},
		"Anlegen, Schreiben scheitert": func(ort string) error {
			return dateiAnlegen(context.Background(), ort, "neu", []byte("x"), hilfenMitAbbruch(beimEinschreiben))
		},
		"Anlegen, Tausch scheitert": func(ort string) error {
			return dateiAnlegen(context.Background(), ort, "neu", nil, ohneTausch)
		},
		"Umbenennen, Tausch scheitert": func(ort string) error {
			return dateiUmbenennen(context.Background(), ort, "skript", "neu", "", ohneTausch)
		},
		"Entfernen, Tausch scheitert": func(ort string) error {
			return dateiEntfernen(context.Background(), ort, "skript", "", ohneTausch)
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := dateibaumprobe(t)

			if err := fall(ort); !errors.Is(err, errAbbruch) {
				t.Fatalf("Fehler %v, erwartet den Abbruch", err)
			}

			// **Nichts Halbes bleibt**: keine neue Datei, keine Nachbardatei,
			// und `skript` steht, wo es stand.
			eintraege, lesefehler := os.ReadDir(ort)
			if lesefehler != nil || len(eintraege) != 2 {
				t.Errorf("%d Einträge, erwartet style und skript", len(eintraege))
			}
		})
	}
}
