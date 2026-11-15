package app

// **Ein Paket, das nach dem Start entsteht, ist für jede Naht da** — FMT-02,
// S10.NAT-03 (#1691).
//
// Gemessen am nativen Fenster (Hüllenprobe `fmt02f`): eine eigene Kopie, im
// frischen Zuhause angelegt, zeichnete die Karte im Set — und die Vorschau der
// Werkstatt meldete „das Stilpaket liegt in keinem Paketverzeichnis". Ihre
// Wurzeln standen seit dem Start fest, und ein leeres Paketverzeichnis trägt
// kein Paket; `Paketwurzeln` nennt es deshalb nicht. Dieselbe Klasse war im
// Composition Root schon behoben (Zweitblick über Phase 7) — in sechs Nähten
// daneben nicht.

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinSpaeterAngelegterStilZeichnetInDerWerkstatt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad, err := Datenbankpfad()
	if err != nil {
		t.Fatal(err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatal(err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Spät", Game: "probe", Style: "probe-schlicht-kopie"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	// **Erst starten, dann die Pakete** — das Paketverzeichnis ist beim Start
	// leer, wie in einem frischen Zuhause.
	anwendung := NeueApp(pfad)
	if fehler := anwendung.Starten(probekontext(t)); fehler != nil {
		t.Fatal(fehler)
	}

	wurzel, err := Paketverzeichnis()
	if err != nil {
		t.Fatal(err)
	}

	saat, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatal(err)
	}

	for quelle, ziel := range map[string]string{
		"probe.mse-game": "probe.mse-game", "probe-schlicht.mse-style": "probe-schlicht-kopie.mse-style",
	} {
		if fehler := os.CopyFS(filepath.Join(wurzel, ziel), os.DirFS(filepath.Join(saat, quelle))); fehler != nil {
			t.Fatal(fehler)
		}
	}

	paket := filepath.Join(wurzel, "probe-schlicht-kopie.mse-style")

	roh, err := os.ReadFile(filepath.Join(paket, "style")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	antwort := anwendung.Vorlagenvorschau(probekontext(t), paket, "style", string(roh), 1, "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, befund := range antwort.Vorschau.Befunde {
		if strings.Contains(befund, "keinem Paketverzeichnis") {
			t.Fatalf("die Werkstatt sieht das Paket nicht: %s", befund)
		}
	}

	marke := strings.TrimSuffix(antwort.Vorschau.Bild[strings.LastIndex(antwort.Vorschau.Bild, "/")+1:], ".png")

	gepackt, da := anwendung.bildfach.Hole(marke)
	if !da {
		t.Fatalf("kein Bild unter %s", antwort.Vorschau.Bild)
	}

	bild, err := png.Decode(bytes.NewReader(gepackt))
	if err != nil {
		t.Fatal(err)
	}

	// **Der weiße Grund des Stils** — ohne Stil stünde dort die Farbe des
	// Platzhalters.
	if r, g, b, _ := bild.At(6, 6).RGBA(); r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("Bildpunkt (6,6) = %d,%d,%d — erwartet der weiße Grund", r>>8, g>>8, b>>8)
	}
}
