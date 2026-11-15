package app

// **Der Abbruch erreicht das Fenster und das Protokoll, und der saubere
// Schluss räumt** — F04 (#2343).

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEinAbbruchStehtAmFensterUndImProtokoll(t *testing.T) {
	var gesammelt bytes.Buffer

	vorher := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&gesammelt, nil)))
	t.Cleanup(func() { slog.SetDefault(vorher) })

	ordner := t.TempDir()
	legeMarke(t, ordner, 1<<30, "")

	// **Und einer mit Spur** — beide Arten stehen im Protokoll.
	if err := os.WriteFile(filepath.Join(ordner, "absturz-1-2.txt"), []byte("panic: Probe\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	legeMarke(t, ordner, 1<<30+1, "absturz-1-2.txt")

	anwendung := gestartet(t)
	if err := anwendung.StartmarkeSetzen(ordner); err != nil {
		t.Fatal(err)
	}

	antwort := anwendung.Wiederanlauf(t.Context())
	if len(antwort.Abbrueche) != 2 || !antwort.Abbrueche[0].Gestartet.Equal(jetzt.Add(-time.Hour)) {
		t.Fatalf("Abbrüche %+v", antwort.Abbrueche)
	}

	if zeilen := gesammelt.String(); !strings.Contains(zeilen, `art="mit Absturzspur"`) || !strings.Contains(zeilen, `art="ohne Absturzspur"`) {
		t.Errorf("der Abbruch steht nicht im Protokoll:\n%s", gesammelt.String())
	}

	frist, aufhoeren := context.WithTimeout(context.Background(), 2*time.Second)
	defer aufhoeren()

	_ = anwendung.Schliessen(frist, nil) //nolint:errcheck // geprüft wird die Marke

	if marken, _ := filepath.Glob(filepath.Join(ordner, "fenster-*.json")); len(marken) != 0 { //nolint:errcheck // ein festes Muster
		t.Errorf("nach dem Schluss liegen noch Marken: %v", marken)
	}
}

func TestOhneSitzungIstDerWiederanlaufLeer(t *testing.T) {
	anwendung := gestartet(t)

	antwort := anwendung.Wiederanlauf(t.Context())
	if antwort.Abbrueche == nil || len(antwort.Abbrueche) != 0 {
		t.Errorf("Abbrüche %#v, erwartet eine leere Liste", antwort.Abbrueche)
	}
}

func TestEinBeginnOhneOrdnerScheitertLaut(t *testing.T) {
	datei := filepath.Join(t.TempDir(), "datei")
	if err := os.WriteFile(datei, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := gestartet(t).StartmarkeSetzen(datei); err == nil {
		t.Error("ein Beginn auf einer Datei als Ordner gelang")
	}
}

func TestEinAbgebrochenerRufBekommtSeinenAbbruch(t *testing.T) {
	abgebrochen, aufhoeren := context.WithCancel(t.Context())
	aufhoeren()

	if antwort := gestartet(t).Wiederanlauf(abgebrochen); antwort.Fehler == nil || antwort.Fehler.Code != CodeAbgebrochen {
		t.Errorf("Antwort %+v, erwartet der Abbruch", antwort)
	}
}
