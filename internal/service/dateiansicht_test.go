package service_test

// **Eine Datei ansehen, ohne dass etwas geschieht** — UX-13.AK-01/03
// (#1685).

import (
	"archive/zip"
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func hinlegen(t *testing.T, pfad string, inhalt []byte) string {
	t.Helper()

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(pfad, inhalt, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	return pfad
}

func ansehen(t *testing.T, pfad string) service.Dateiansicht {
	t.Helper()

	ansicht, fehler := service.DateiAnsehen(pfad)
	if fehler != nil {
		t.Fatalf("ansehen: %v", fehler)
	}

	return ansicht
}

func TestEinTextErscheintAlsTextUndHoechstensBisZurGrenze(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	kurz := ansehen(t, hinlegen(t, filepath.Join(ort, "liste.txt"), []byte("4 Blitzschlag\n\xff kaputt")))
	if kurz.Regel.Vorschau != domain.VorschauText || kurz.Gekuerzt || !strings.HasPrefix(kurz.Text, "4 Blitzschlag") ||
		!strings.Contains(kurz.Text, "kaputt") || strings.Contains(kurz.Text, "\xff") {
		t.Errorf("kurz: %+v", kurz)
	}

	lang := ansehen(t, hinlegen(t, filepath.Join(ort, "lang.csv"), bytes.Repeat([]byte("a"), 300<<10)))
	if !lang.Gekuerzt || len(lang.Text) != 256<<10 || lang.Groesse != 300<<10 {
		t.Errorf("lang: gekürzt %v, %d Zeichen von %d", lang.Gekuerzt, len(lang.Text), lang.Groesse)
	}

	// **HTML bleibt Text** — auch mit Skript darin.
	html := ansehen(t, hinlegen(t, filepath.Join(ort, "set.html"), []byte("<script>alert(1)</script>")))
	if html.Regel.Vorschau != domain.VorschauText || html.Text != "<script>alert(1)</script>" || html.Regel.Oeffnen {
		t.Errorf("html: %+v", html)
	}
}

func TestEinBildErscheintMitSeinenMassen(t *testing.T) {
	t.Parallel()

	var roh bytes.Buffer

	bild := image.NewRGBA(image.Rect(0, 0, 3, 2))
	bild.Set(0, 0, color.White)

	if fehler := png.Encode(&roh, bild); fehler != nil {
		t.Fatal(fehler)
	}

	ansicht := ansehen(t, hinlegen(t, filepath.Join(t.TempDir(), "karte.png"), roh.Bytes()))
	if ansicht.Bildtyp != "image/png" || ansicht.Breite != 3 || ansicht.Hoehe != 2 || !bytes.Equal(ansicht.Bild, roh.Bytes()) {
		t.Errorf("%+v", ansicht)
	}

	// **Ein SVG kommt als Bild** — das Fenster zeigt es über `<img>`, und
	// dort läuft kein Skript.
	svg := ansehen(t, hinlegen(t, filepath.Join(t.TempDir(), "symbol.svg"), []byte("<svg><script>x()</script></svg>")))
	if svg.Bildtyp != "image/svg+xml" || len(svg.Bild) == 0 || svg.Text != "" {
		t.Errorf("svg: %+v", svg)
	}

	// **Kaputt ist kein Fehler** — die Maße fehlen dann, der Rest steht da.
	kaputt := ansehen(t, hinlegen(t, filepath.Join(t.TempDir(), "kaputt.jpg"), []byte("kein jpeg")))
	if kaputt.Breite != 0 || kaputt.Bildtyp != "image/jpeg" {
		t.Errorf("kaputt: %+v", kaputt)
	}
}

func TestZuGrossZeigtNurWasBekanntIst(t *testing.T) {
	t.Parallel()

	gross := ansehen(t, hinlegen(t, filepath.Join(t.TempDir(), "riesig.svg"), bytes.Repeat([]byte(" "), 1<<20+1)))
	if !gross.ZuGross || len(gross.Bild) != 0 {
		t.Errorf("%+v", gross)
	}
}

func TestEinPaketErscheintAlsListeUndNichtsWirdGefolgt(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	paket := filepath.Join(ort, "magic.mse-game")

	hinlegen(t, filepath.Join(paket, "game"), []byte("mse version: 2.0.0"))
	hinlegen(t, filepath.Join(paket, "bilder", "rahmen.png"), []byte("x"))

	if fehler := os.Symlink("/etc", filepath.Join(paket, "fremd")); fehler != nil {
		t.Fatal(fehler)
	}

	ansicht := ansehen(t, paket)
	if !ansicht.Verzeichnis || ansicht.Regel.Art != domain.DateiartPaket || ansicht.Unsicher != 1 {
		t.Fatalf("%+v", ansicht)
	}

	namen := map[string]service.Dateieintrag{}
	for _, eintrag := range ansicht.Eintraege {
		namen[eintrag.Name] = eintrag
	}

	if namen["game"].Groesse != 18 || !namen["bilder"].Verzeichnis || !namen["fremd"].Unsicher {
		t.Errorf("%+v", namen)
	}

	if _, gefolgt := namen["fremd/hosts"]; gefolgt {
		t.Error("der Verweis wurde gefolgt")
	}
}

func TestEinArchivWirdGelesenUndNichtAusgepackt(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	var roh bytes.Buffer

	schreiber := zip.NewWriter(&roh)
	for _, name := range []string{"bestand.db", "beiwerk/7/kunst.png", "../../boese", "/absolut"} {
		teil, fehler := schreiber.Create(name)
		if fehler != nil {
			t.Fatal(fehler)
		}

		if _, fehler := teil.Write([]byte("inhalt")); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	ansicht := ansehen(t, hinlegen(t, filepath.Join(ort, "stand.msesicherung"), roh.Bytes()))
	if len(ansicht.Eintraege) != 4 || ansicht.Unsicher != 2 {
		t.Errorf("%+v", ansicht)
	}

	if eintraege, _ := os.ReadDir(ort); len(eintraege) != 1 { //nolint:errcheck // der Vergleich sagt es
		t.Errorf("es wurde ausgepackt: %v", eintraege)
	}

	// **Eine verschlüsselte Sicherung sagt, dass sie es ist.**
	geheim := ansehen(t, hinlegen(t, filepath.Join(ort, "geheim.msesicherung"), []byte("age-encryption.org/v1\n->")))
	if !geheim.Verschluesselt || len(geheim.Eintraege) != 0 {
		t.Errorf("geheim: %+v", geheim)
	}

	// **Und ein kaputtes Archiv ist keines** — ohne Einträge, ohne Fehler.
	kaputt := ansehen(t, hinlegen(t, filepath.Join(ort, "kaputt.zip"), []byte("PK kaputt")))
	if len(kaputt.Eintraege) != 0 || !kaputt.Unlesbar {
		t.Errorf("kaputt: %+v", kaputt)
	}
}

func TestEinPDFNenntSeineSeiten(t *testing.T) {
	t.Parallel()

	inhalt := "%PDF-1.7\n1 0 obj << /Type /Pages /Count 2 >>\n2 0 obj << /Type /Page >>\n3 0 obj <</Type/Page>>\n"

	ansicht := ansehen(t, hinlegen(t, filepath.Join(t.TempDir(), "bogen.pdf"), []byte(inhalt)))
	if ansicht.Seiten != 2 || ansicht.Regel.Vorschau != domain.VorschauKeine || !ansicht.Regel.Oeffnen {
		t.Errorf("%+v", ansicht)
	}

	falsch := ansehen(t, hinlegen(t, filepath.Join(t.TempDir(), "falsch.pdf"), []byte("kein pdf")))
	if falsch.Seiten != 0 || !falsch.Unlesbar {
		t.Errorf("falsch: %+v", falsch)
	}
}

func TestEineUnbekannteDateiZeigtNurIhreAngaben(t *testing.T) {
	t.Parallel()

	ansicht := ansehen(t, hinlegen(t, filepath.Join(t.TempDir(), "start.command"), []byte("#!/bin/sh\nrm -rf ~")))
	if ansicht.Regel.Art != domain.DateiartUnbekannt || ansicht.Text != "" || ansicht.Groesse != 18 || ansicht.Geaendert.IsZero() {
		t.Errorf("%+v", ansicht)
	}

	// **Ein gewöhnlicher Ordner auch.**
	ordner := ansehen(t, t.TempDir())
	if !ordner.Verzeichnis || len(ordner.Eintraege) != 0 {
		t.Errorf("ordner: %+v", ordner)
	}

	if _, fehler := service.DateiAnsehen(filepath.Join(t.TempDir(), "fehlt")); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("fehlt: %v", fehler)
	}
}

// **Was sich nicht lesen lässt, meldet sich** — der Ordner davor als Fehler,
// die Datei selbst als Ansicht ohne Inhalt.
func TestEineUnlesbareDateiMeldetSich(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	datei := hinlegen(t, filepath.Join(ort, "zu", "liste.txt"), []byte("geheim"))

	if fehler := os.Chmod(datei, 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	if ansicht := ansehen(t, datei); ansicht.Text != "" || ansicht.Groesse != 6 {
		t.Errorf("unlesbar: %+v", ansicht)
	}

	if fehler := os.Chmod(filepath.Join(ort, "zu"), 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(filepath.Join(ort, "zu"), 0o700) }) //nolint:errcheck,gosec // Aufräumen

	if _, fehler := service.DateiAnsehen(datei); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("versperrt: %v", fehler)
	}
}

// **An macOS geht nur ein ganzer Pfad, und geöffnet wird nur, was die Regel
// erlaubt** — UX-13.AK-04.
func TestDieUebergabeAnMacOSFolgtDerRegel(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	bild := hinlegen(t, filepath.Join(ort, "karte.png"), []byte("x"))
	skript := hinlegen(t, filepath.Join(ort, "start.command"), []byte("x"))

	for _, fall := range []struct {
		name    string
		pfad    string
		oeffnen bool
		art     error
	}{
		{"ein Bild zum Zeigen", bild, false, nil},
		{"ein Bild zum Öffnen", bild, true, nil},
		{"ein Skript zum Zeigen", skript, false, nil},
		{"ein Skript zum Öffnen", skript, true, domain.ErrValidation},
		{"ein Ordner zum Öffnen", ort, true, domain.ErrValidation},
		{"ein Ordner zum Zeigen", ort, false, nil},
		{"ein relativer Pfad", "-a Rechner", false, domain.ErrValidation},
		{"eine fehlende Datei", filepath.Join(ort, "fehlt.png"), false, domain.ErrNotFound},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := service.DateiUebergabePruefen(fall.pfad, fall.oeffnen)
			if fall.art == nil && fehler != nil || fall.art != nil && !errors.Is(fehler, fall.art) {
				t.Errorf("%v, erwartet %v", fehler, fall.art)
			}
		})
	}
}

// **Ein Paketordner mit einem verschlossenen Unterordner** zeigt, was er
// zeigen kann — der Unterordner steht da, sein Inhalt fehlt.
func TestEinVerschlossenerUnterordnerBrichtDieListeNichtAb(t *testing.T) {
	t.Parallel()

	paket := filepath.Join(t.TempDir(), "magic.mse-style")
	hinlegen(t, filepath.Join(paket, "style"), []byte("x"))
	hinlegen(t, filepath.Join(paket, "zu", "geheim.png"), []byte("x"))

	if fehler := os.Chmod(filepath.Join(paket, "zu"), 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(filepath.Join(paket, "zu"), 0o700) }) //nolint:errcheck,gosec // Aufräumen

	ansicht := ansehen(t, paket)

	namen := make([]string, 0, len(ansicht.Eintraege))
	for _, eintrag := range ansicht.Eintraege {
		namen = append(namen, eintrag.Name)
	}

	if strings.Join(namen, ",") != "style,zu" {
		t.Errorf("%v", namen)
	}
}
