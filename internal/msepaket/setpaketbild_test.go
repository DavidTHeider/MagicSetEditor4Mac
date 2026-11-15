package msepaket_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Das Bild einer Karte liegt im Set-Paket, und das Set liegt irgendwo.**
//
// Ein Stilpaket wird über die Paketwurzeln gefunden; ein Set nicht — es liegt
// dort, wo jemand es abgelegt hat. Deshalb nennt dieser Verweis den **Ort**
// (#380).
func TestEinBildKommtAusDemSetpaketAnSeinemOrt(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "Mein Set.mse-set")
	legePNG(t, filepath.Join(ort, "image1"), color.NRGBA{R: 9, G: 8, B: 7, A: 255})

	verweis := msepaket.SetpaketVerweis(ort, "image1")

	bild, err := msepaket.NeueBildquelle().Bild(verweis)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen := ecke(t, bild); gelesen != (color.NRGBA{R: 9, G: 8, B: 7, A: 255}) {
		t.Errorf("Ecke %+v", gelesen)
	}
}

// **Auch ein gepacktes Set gibt sein Bild her.** Ein `.mse-set` ist im Regelfall
// ein Zip; das Verzeichnis daneben ist die entpackte Form.
func TestEinBildKommtAusDemGepacktenSetpaket(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "Gepackt.mse-set")
	legeZipMitBild(t, ort, "image1", color.NRGBA{R: 1, G: 2, B: 3, A: 255})

	bild, err := msepaket.NeueBildquelle().Bild(msepaket.SetpaketVerweis(ort, "image1"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen := ecke(t, bild); gelesen != (color.NRGBA{R: 1, G: 2, B: 3, A: 255}) {
		t.Errorf("Ecke %+v", gelesen)
	}
}

// **Zwei Sets mit gleichnamiger Datei bleiben zwei Bilder.**
//
// Der Grund, warum der Verweis den **ganzen Ort** trägt und nicht den
// Paketnamen: der Bildspeicher merkt sich nach Verweis, und zwei Sets heißen
// erfahrungsgemäß beide `image1`. Ein Speicher nach Namen zeigte auf der einen
// Karte die Illustration der anderen.
func TestZweiSetsTeilenSichKeinBild(t *testing.T) {
	erstes := filepath.Join(t.TempDir(), "Set.mse-set")
	zweites := filepath.Join(t.TempDir(), "Set.mse-set")
	legePNG(t, filepath.Join(erstes, "image1"), color.NRGBA{R: 200, A: 255})
	legePNG(t, filepath.Join(zweites, "image1"), color.NRGBA{B: 200, A: 255})

	quelle := msepaket.NeueBildquelle()

	eins, err := quelle.Bild(msepaket.SetpaketVerweis(erstes, "image1"))
	if err != nil {
		t.Fatalf("erstes lesen: %v", err)
	}

	zwei, err := quelle.Bild(msepaket.SetpaketVerweis(zweites, "image1"))
	if err != nil {
		t.Fatalf("zweites lesen: %v", err)
	}

	if ecke(t, eins) == ecke(t, zwei) {
		t.Errorf("beide Sets geben dieselbe Ecke %+v", ecke(t, eins))
	}
}

// **Eine fehlende Datei ist ein Fehler und kein Absturz** (§8).
func TestEinFehlendesSetbildMeldetSich(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "Leer.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := msepaket.NeueBildquelle().Bild(msepaket.SetpaketVerweis(ort, "image1")); err == nil {
		t.Fatal("ein fehlendes Bild wurde nicht gemeldet")
	}
}

// **Ein Verweis, der nichts benennt, wird abgewiesen.**
//
// Und zwar hier, nicht erst am Dateisystem: `..` in der Datei führte sonst aus
// dem Set heraus, und ein leerer Ort machte aus `filepath.Join` einen
// **relativen** Pfad ins Arbeitsverzeichnis des Prozesses.
func TestEinKaputterSetverweisWirdAbgewiesen(t *testing.T) {
	faelle := map[string]string{
		"ohne Ort":         msepaket.SetpaketVerweis("", "image1"),
		"ohne Datei":       msepaket.SetpaketVerweis("/tmp/Set.mse-set", ""),
		"Ausbruch über ..": msepaket.SetpaketVerweis("/tmp/Set.mse-set", "../geheim"),
		"Ausbruch über . mitten": msepaket.SetpaketVerweis("/tmp/Set.mse-set",
			"bilder/../../geheim"),
	}

	for name, verweis := range faelle {
		t.Run(name, func(t *testing.T) {
			_, err := msepaket.NeueBildquelle().Bild(verweis)
			if !errors.Is(err, msepaket.ErrPaketpfad) {
				t.Errorf("Fehler %v, erwartet ErrPaketpfad", err)
			}
		})
	}
}

// **Ein gewöhnlicher Paketpfad bleibt einer.** Die neue Form steht daneben und
// nicht an ihrer Stelle.
func TestZerlegeSetpaketverweisLaesstPaketpfadeInRuhe(t *testing.T) {
	if _, _, ist := msepaket.ZerlegeSetpaketverweis("/probe.mse-style/rahmen.png"); ist {
		t.Error("ein Paketpfad wurde als Setverweis gelesen")
	}

	ort, datei, ist := msepaket.ZerlegeSetpaketverweis(
		msepaket.SetpaketVerweis("/a/Mein Set.mse-set", "image1"))
	if !ist || ort != "/a/Mein Set.mse-set" || datei != "image1" {
		t.Errorf("Ort %q, Datei %q, ist %v", ort, datei, ist)
	}
}

// legeZipMitBild schreibt ein gepacktes Paket mit einem Bild darin.
func legeZipMitBild(t *testing.T, pfad, name string, ecke color.NRGBA) {
	t.Helper()

	var bildpuffer bytes.Buffer
	if err := png.Encode(&bildpuffer, gemaltesBild(ecke)); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	var puffer bytes.Buffer

	schreiber := zip.NewWriter(&puffer)

	eintrag, err := schreiber.Create(name)
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}

	if _, err := eintrag.Write(bildpuffer.Bytes()); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := schreiber.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if err := os.WriteFile(pfad, puffer.Bytes(), 0o600); err != nil {
		t.Fatalf("ablegen: %v", err)
	}
}

// **Ein Verweis mit Schema, aber ohne Trenner, ist keiner.**
//
// `setpaket:/Orte/S.mse-set` nennt kein Ziel. Er wird hier abgewiesen und nicht
// als Paketpfad weitergereicht — sonst suchte die Quelle ein Paket namens
// `setpaket:`.
func TestEinSetverweisOhneTrennerWirdAbgewiesen(t *testing.T) {
	ort, datei, ist := msepaket.ZerlegeSetpaketverweis("setpaket:/Orte/S.mse-set")
	if !ist || ort != "" || datei != "" {
		t.Errorf("Ort %q, Datei %q, ist %v", ort, datei, ist)
	}

	if _, err := msepaket.NeueBildquelle().Bild("setpaket:/Orte/S.mse-set"); !errors.Is(
		err, msepaket.ErrPaketpfad) {
		t.Errorf("Fehler %v, erwartet ErrPaketpfad", err)
	}
}

// **Das zweite Lesen kommt aus dem Speicher.**
//
// Die Gegenprobe steckt im Aufbau: nach dem ersten Lesen verschwindet die
// Datei. Wer trotzdem ein Bild bekommt, hat es gemerkt — und genau das trägt
// die Vorschau, die bei jedem Tastendruck neu zeichnet (§7).
func TestEinSetbildWirdGemerkt(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "Gemerkt.mse-set")
	pfad := filepath.Join(ort, "image1")
	legePNG(t, pfad, color.NRGBA{G: 128, A: 255})

	quelle := msepaket.NeueBildquelle()
	verweis := msepaket.SetpaketVerweis(ort, "image1")

	if _, err := quelle.Bild(verweis); err != nil {
		t.Fatalf("erstes Lesen: %v", err)
	}

	if err := os.Remove(pfad); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	bild, err := quelle.Bild(verweis)
	if err != nil {
		t.Fatalf("zweites Lesen: %v", err)
	}

	if gelesen := ecke(t, bild); gelesen != (color.NRGBA{G: 128, A: 255}) {
		t.Errorf("Ecke %+v", gelesen)
	}
}

// **Eine Datei, die kein Bild ist, wird benannt** — und nicht als leeres Bild
// gezeichnet.
func TestEinUnleserlichesSetbildMeldetSich(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "Kaputt.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "image1"), []byte("kein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	_, err := msepaket.NeueBildquelle().Bild(msepaket.SetpaketVerweis(ort, "image1"))
	if err == nil || !strings.Contains(err.Error(), "entschlüsseln") {
		t.Errorf("Fehler %v, erwartet einen über das Entschlüsseln", err)
	}
}

// **Leerraum am Ende eines Bildwerts gehört zum Dateinamen** — gefunden von
// `FuzzPaketpfad` (QA-06b,
// #1967).
//
// Der Zerteiler schneidet einen Wert nur links, wie MSE2 (`trim_left`,
// `src/util/io/reader.cpp` Zeile 224, Fassung f9d9356 vom 2021-07-06), und
// `image: image1 ` nennt dort die Datei `image1 `. Der Verweis schnitt beim
// Zerlegen auch rechts und öffnete eine andere Datei, als die Karte nennt.
func TestEinSetbildMitLeerraumAmEndeIstSeineDatei(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "Mein Set.mse-set")
	legePNG(t, filepath.Join(ort, "image1 "), color.NRGBA{R: 200, A: 255})
	legePNG(t, filepath.Join(ort, "image1"), color.NRGBA{B: 200, A: 255})

	bild, err := msepaket.NeueBildquelle().Bild(msepaket.SetpaketVerweis(ort, "image1 "))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen := ecke(t, bild); gelesen != (color.NRGBA{R: 200, A: 255}) {
		t.Errorf("gelesen wurde das Bild mit der Ecke %+v — das von „image1“", gelesen)
	}
}
