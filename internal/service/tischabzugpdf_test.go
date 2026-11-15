package service_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einBreiterAbzug gibt einen Data-URL mit einem echten PNG von 40 × 20.
func einBreiterAbzug(t *testing.T) string {
	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, 40, 20))
	bild.Set(1, 1, color.RGBA{R: 255, A: 255})

	var puffer bytes.Buffer
	if err := png.Encode(&puffer, bild); err != nil {
		t.Fatalf("das Probebild schreiben: %v", err)
	}

	return domain.Abzugsvorsatz +
		base64.StdEncoding.EncodeToString(puffer.Bytes())
}

// alsZahl liest ein Maß aus der MediaBox.
func alsZahl(t *testing.T, rohes []byte) float64 {
	t.Helper()

	wert, err := strconv.ParseFloat(string(rohes), 64)
	if err != nil {
		t.Fatalf("%q ist keine Zahl: %v", rohes, err)
	}

	return wert
}

func TestTischbildausfuhrSchreibtEinPDF(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "tisch.pdf")

	fehler := service.NeueTischbildausfuhr().
		SchreibePDF(einBreiterAbzug(t), ziel, domain.BlattA4)
	if fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", fehler)
	}

	gelesen, err := os.ReadFile(ziel) // #nosec G304 -- ein Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	// **Ein PDF fängt mit `%PDF-` an und endet mit `%%EOF`** — beides ist die
	// Formatbeschreibung und keine Wahl dieses Schreibers.
	if !bytes.HasPrefix(gelesen, []byte("%PDF-")) {
		t.Errorf("die Datei fängt mit %q an", gelesen[:min(8, len(gelesen))])
	}

	if !bytes.Contains(gelesen, []byte("%%EOF")) {
		t.Error("die Datei trägt keine Endmarke")
	}

	// **Und die Seite liegt quer** — ein breiter Abzug legt A4 quer, also
	// 297 mm breit und 210 hoch.
	//
	// **Gelesen aus der MediaBox und nicht gegen eine getippte Zahl**: wie
	// viele Nachkommastellen der Schreiber setzt, ist seine Sache, und ein
	// Lauf, der sie mitprüfte, bräche bei jeder Formatierung (§10).
	kasten := regexp.MustCompile(`/MediaBox \[0 0 ([\d.]+) ([\d.]+)\]`).
		FindSubmatch(gelesen)
	if kasten == nil {
		t.Fatal("die Seite trägt keine MediaBox")
	}

	breite, hoehe := alsZahl(t, kasten[1]), alsZahl(t, kasten[2])

	if breite <= hoehe {
		t.Errorf("die Seite ist %v × %v pt — erwartet quer", breite, hoehe)
	}

	// **297 mm sind 841,89 pt** — die Umrechnung ist die Definition des
	// Punktes und keine Wahl.
	if math.Abs(breite-841.89) > 0.1 {
		t.Errorf("die Seite ist %v pt breit, erwartet 841,89", breite)
	}
}

// **Ein unbekanntes Format ist ein Befund und kein stilles A4.**
func TestTischbildausfuhrPDFOhneFormat(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "tisch.pdf")

	fehler := service.NeueTischbildausfuhr().
		SchreibePDF(einBreiterAbzug(t), ziel, "a3")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}

	if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Error("die Datei steht da, obwohl das Format abgewiesen wurde")
	}
}

// **Ein Abzug, der kein Bild ist, kommt nicht an die Platte** — dieselbe
// Prüfung wie beim PNG, und dieser Lauf hält fest, dass der Weg sie ruft.
func TestTischbildausfuhrPDFWeistFremdesAb(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "tisch.pdf")

	fehler := service.NeueTischbildausfuhr().
		SchreibePDF("data:text/plain,hallo", ziel, domain.BlattA4)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Und ein PNG, das keines ist, ebenso** — die Signatur stimmt, der Rumpf
// nicht. Sie ist acht Bytes lang; was danach kommt, prüft erst der Dekoder
// (§8).
func TestTischbildausfuhrPDFWeistEinKaputtesBildAb(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "tisch.pdf")

	kaputt := domain.Abzugsvorsatz + base64.StdEncoding.EncodeToString(
		[]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 'x', 'y'})

	if fehler := service.NeueTischbildausfuhr().
		SchreibePDF(kaputt, ziel, domain.BlattA4); fehler == nil {
		t.Fatal("erwartet einen Fehler")
	}
}

// **Ein leerer Pfad ist keine Datei** — wie beim PNG daneben.
func TestTischbildausfuhrPDFOhnePfad(t *testing.T) {
	t.Parallel()

	fehler := service.NeueTischbildausfuhr().
		SchreibePDF(einBreiterAbzug(t), "", domain.BlattA4)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Ein Verzeichnis, das es nicht gibt, ist ein Fehler und kein halbes PDF.**
func TestTischbildausfuhrPDFOhneVerzeichnis(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "gibtsnicht", "tisch.pdf")

	if fehler := service.NeueTischbildausfuhr().
		SchreibePDF(einBreiterAbzug(t), ziel, domain.BlattA4); fehler == nil {
		t.Fatal("erwartet einen Fehler")
	}
}
