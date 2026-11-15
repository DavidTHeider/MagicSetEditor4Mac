//go:build darwin

package app_test

// **Ein Leistungsprofil, gezielt und begrenzt** — OPS-05.AK-03
// (#1722):
// „Profiling/Tracing ist gezielt aktivierbar und verursacht keinen
// unbegrenzten lokalen Datenbestand." Ein CPU-Profil über eine gewählte Zahl
// von Sekunden, neben dem Bestand, und nur die letzten liegen bleiben.

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinLeistungsprofilLiegtDanachNebenDemBestand(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if antwort := anwendung.EinstellungSetzen(t.Context(), domain.MessungProfile, "2"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Zwei ältere liegen schon da** — nach dem neuen bleiben zwei, und das
	// älteste geht.
	ordner := filepath.Join(filepath.Dir(bestand), app.Profilordner)
	if fehler := os.MkdirAll(ordner, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	for stelle, name := range []string{"mse-20260101-000000.pprof", "mse-20260102-000000.pprof"} {
		pfad := filepath.Join(ordner, name)
		if fehler := os.WriteFile(pfad, []byte("alt"), 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		zeit := time.Date(2026, 1, 1+stelle, 0, 0, 0, 0, time.UTC)
		if fehler := os.Chtimes(pfad, zeit, zeit); fehler != nil {
			t.Fatal(fehler)
		}
	}

	antwort := anwendung.Leistungsprofil(t.Context(), 1)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	roh, fehler := os.ReadFile(antwort.Datei)
	if fehler != nil || filepath.Dir(antwort.Datei) != ordner || antwort.Bytes != int64(len(roh)) ||
		antwort.Sekunden != 1 {
		t.Fatalf("%+v, %v", antwort, fehler)
	}

	// **Ein pprof-Profil ist gzip** — die ersten zwei Bytes sagen es.
	if !bytes.HasPrefix(roh, []byte{0x1f, 0x8b}) {
		t.Errorf("kein Profil: %x", roh[:min(len(roh), 4)])
	}

	eintraege, lesefehler := os.ReadDir(ordner)
	if lesefehler != nil || len(eintraege) != 2 || eintraege[0].Name() != "mse-20260102-000000.pprof" {
		t.Errorf("im Ordner: %v", eintraege)
	}
}

func TestEinLeistungsprofilHatGrenzen(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	for _, sekunden := range []int{0, 61} {
		if antwort := anwendung.Leistungsprofil(t.Context(), sekunden); antwort.Fehler == nil ||
			antwort.Fehler.Code != app.CodeValidierung {
			t.Errorf("%d s: %+v", sekunden, antwort.Fehler)
		}
	}

	// **Ein abgebrochener Lauf hinterlässt nichts** — die Datei wäre ein
	// halbes Profil, und wer sie läse, hielte sie für ein ganzes.
	ctx, abbrechen := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer abbrechen()

	antwort := anwendung.Leistungsprofil(ctx, 30)
	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeAbgebrochen || antwort.Datei != "" {
		t.Errorf("abgebrochen: %+v", antwort)
	}
}

// **Eines zur Zeit** — ein CPU-Profil gehört dem ganzen Prozess.
func TestZweiLeistungsprofileZugleichGehenNicht(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	erstes := make(chan app.LeistungsprofilAntwort)

	go func() { erstes <- anwendung.Leistungsprofil(t.Context(), 1) }()

	// **Gewartet wird auf die Wirkung**: das erste läuft, sobald es als
	// laufend gilt.
	for !anwendung.LeistungsprofilLaeuft() {
		time.Sleep(time.Millisecond)
	}

	zweites := anwendung.Leistungsprofil(t.Context(), 1)
	if zweites.Fehler == nil || zweites.Fehler.Code != app.CodeKonflikt {
		t.Errorf("zweites: %+v", zweites.Fehler)
	}

	if antwort := <-erstes; antwort.Fehler != nil {
		t.Errorf("erstes: %+v", antwort.Fehler)
	}
}

// **Ohne Bestand kein Ort** — und kein Profil ins Arbeitsverzeichnis.
func TestOhneBestandGibtEsKeinLeistungsprofil(t *testing.T) {
	if antwort := app.NeueApp("").Leistungsprofil(t.Context(), 1); antwort.Fehler == nil ||
		antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("%+v", antwort.Fehler)
	}
}

// **Ein fremdes Profil sperrt auch** — Go lässt je Prozess nur eines zu.
func TestEinFremdesProfilSperrt(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	if fehler := pprof.StartCPUProfile(io.Discard); fehler != nil {
		t.Skipf("dieser Lauf profiliert schon selbst: %v", fehler)
	}

	antwort := anwendung.Leistungsprofil(t.Context(), 1)
	pprof.StopCPUProfile()

	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("%+v", antwort.Fehler)
	}
}

// **Ein Ort, an dem nichts entstehen kann, ist ein Fehler** — und kein
// Profil, das es nicht gibt.
func TestEinVersperrterOrtIstEinFehler(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	sperre := filepath.Join(filepath.Dir(bestand), app.Profilordner)
	if fehler := os.WriteFile(sperre, []byte("keine Mappe"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if antwort := anwendung.Leistungsprofil(t.Context(), 1); antwort.Fehler == nil || antwort.Datei != "" {
		t.Errorf("%+v", antwort)
	}
}

// **Ein Ordner, der sich nicht lesen lässt, verhindert das Aufräumen und
// nicht das Profil** — geschrieben ist geschrieben (§7, Regel 3).
func TestEinUnlesbarerOrdnerLaesstDasProfilStehen(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	ordner := filepath.Join(filepath.Dir(bestand), app.Profilordner)
	if fehler := os.Mkdir(ordner, 0o300); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if fehler := os.Chmod(ordner, 0o700); fehler != nil { //nolint:gosec // ein Ordner braucht das Ausführrecht, sonst räumt t.TempDir ihn nicht weg
			t.Error(fehler)
		}
	})

	if antwort := anwendung.Leistungsprofil(t.Context(), 1); antwort.Fehler != nil || antwort.Bytes == 0 {
		t.Errorf("%+v", antwort)
	}
}
