package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// **Eine Datei ansehen und an macOS geben** — UX-13.AK-01/03/04
// (#1685).

func TestDasFensterSiehtEinBildAlsDatenadresse(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "symbol.svg")
	if fehler := os.WriteFile(pfad, []byte("<svg/>"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	ansicht := NeueApp("").DateiAnsehen(t.Context(), pfad)
	if ansicht.Fehler != nil || ansicht.Bild != "data:image/svg+xml;base64,PHN2Zy8+" || ansicht.Art != "vektor" ||
		!ansicht.Oeffnen || ansicht.Eintraege == nil {
		t.Errorf("%+v", ansicht)
	}

	// **Die Flächen ohne Fenster bekommen keine Bilddaten** — sie haben die
	// Datei selbst; eine Antwort von 20 MB wäre eine Kopie und keine Ansicht.
	if ohne := DateiansichtAus(pfad); ohne.Bild != "" || ohne.Name != "symbol.svg" {
		t.Errorf("ohne Bild: %+v", ohne)
	}

	if fehlt := NeueApp("").DateiAnsehen(t.Context(), filepath.Join(t.TempDir(), "fehlt")); fehlt.Fehler == nil {
		t.Error("eine fehlende Datei ohne Fehler")
	}

	if fehlt := DateiansichtAus(filepath.Join(t.TempDir(), "fehlt")); fehlt.Fehler == nil || fehlt.Eintraege == nil {
		t.Errorf("ohne Fenster: %+v", fehlt)
	}

	// **Ein Paket kommt mit seinen Einträgen.**
	paket := filepath.Join(t.TempDir(), "magic.mse-game")
	if fehler := os.MkdirAll(paket, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(filepath.Join(paket, "game"), []byte("x"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if liste := DateiansichtAus(paket); len(liste.Eintraege) != 1 || liste.Eintraege[0].Name != "game" || liste.Oeffnen {
		t.Errorf("paket: %+v", liste)
	}
}

// aufmacher merkt sich, womit macOS gerufen würde.
type aufmacher struct {
	gerufen [][]string
	fehler  error
}

func (a *aufmacher) aufmachen(_ context.Context, argumente ...string) error {
	a.gerufen = append(a.gerufen, argumente)

	return a.fehler
}

func TestImFinderZeigenUndOeffnenFolgenDerRegel(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	bild := filepath.Join(ort, "karte.png")
	skript := filepath.Join(ort, "start.command")

	for _, pfad := range []string{bild, skript} {
		if fehler := os.WriteFile(pfad, []byte("x"), 0o600); fehler != nil {
			t.Fatal(fehler)
		}
	}

	anwendung := NeueApp("")
	gerufen := &aufmacher{}

	// **Im Finder zeigen geht für jede Datei**, denn es startet nichts.
	for _, pfad := range []string{bild, skript} {
		if antwort := anwendung.dateiZeigen(t.Context(), pfad, gerufen.aufmachen); antwort.Fehler != nil {
			t.Errorf("zeigen %s: %+v", pfad, antwort.Fehler)
		}
	}

	if antwort := anwendung.dateiOeffnen(t.Context(), bild, gerufen.aufmachen); antwort.Fehler != nil {
		t.Errorf("öffnen: %+v", antwort.Fehler)
	}

	// **Öffnen nur, was die Regel erlaubt** — ein `.command` startete ein
	// Programm, ein relativer Pfad könnte eine Option sein, ein Ordner ein
	// Programmbündel.
	for _, pfad := range []string{skript, "-a Rechner", ort} {
		if antwort := anwendung.dateiOeffnen(t.Context(), pfad, gerufen.aufmachen); antwort.Fehler == nil {
			t.Errorf("geöffnet: %s", pfad)
		}
	}

	if antwort := anwendung.dateiZeigen(t.Context(), "relativ.png", gerufen.aufmachen); antwort.Fehler == nil {
		t.Error("ein relativer Pfad wurde gezeigt")
	}

	erwartet := [][]string{{"-R", bild}, {"-R", skript}, {bild}}
	if !slices.EqualFunc(gerufen.gerufen, erwartet, slices.Equal) {
		t.Errorf("gerufen: %q", gerufen.gerufen)
	}

	// **Und ein Fehler von macOS kommt an.**
	gerufen.fehler = errors.New("kein Finder")
	if antwort := anwendung.dateiZeigen(t.Context(), bild, gerufen.aufmachen); antwort.Fehler == nil ||
		!strings.Contains(antwort.Fehler.Meldung, "Finder") {
		t.Errorf("Fehler: %+v", antwort.Fehler)
	}
}

// **Eine Tilde ist am Fenster das Zuhause** — die Orte der Hilfe stehen so da
// (`schwaerze`), und wer „Ansehen" daneben drückt, meint die Datei.
func TestEineTildeIstAmFensterDasZuhause(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	pfad := filepath.Join(zuhause, "liste.txt")
	if fehler := os.WriteFile(pfad, []byte("x"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if ansicht := NeueApp("").DateiAnsehen(t.Context(), "~/liste.txt"); ansicht.Fehler != nil || ansicht.Pfad != pfad {
		t.Errorf("%+v", ansicht)
	}

	gerufen := &aufmacher{}
	if antwort := NeueApp("").dateiZeigen(t.Context(), "~/liste.txt", gerufen.aufmachen); antwort.Fehler != nil ||
		!slices.EqualFunc(gerufen.gerufen, [][]string{{"-R", pfad}}, slices.Equal) {
		t.Errorf("zeigen: %+v %q", antwort.Fehler, gerufen.gerufen)
	}

	// **Nur am Anfang**, und **nur am Fenster**: der Dienst prüft den Pfad
	// vor dem Ansehen, und eine Tilde danach aufzulösen hieße, die Prüfung
	// an einem anderen Pfad vorzunehmen als dem gelesenen.
	if ansicht := DateiansichtAus("~/liste.txt"); ansicht.Fehler == nil {
		t.Error("ohne Fenster wurde die Tilde aufgelöst")
	}

	if antwort := NeueApp("").dateiZeigen(t.Context(), "a/~/liste.txt", gerufen.aufmachen); antwort.Fehler == nil {
		t.Error("eine Tilde in der Mitte wurde aufgelöst")
	}

	// **Ohne Zuhause bleibt die Tilde stehen** — und der Pfad ist keiner.
	t.Setenv("HOME", "")

	if ansicht := NeueApp("").DateiAnsehen(t.Context(), "~/liste.txt"); ansicht.Fehler == nil {
		t.Error("ohne Zuhause wurde die Tilde aufgelöst")
	}
}
