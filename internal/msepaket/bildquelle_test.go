package msepaket_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// gemaltesBild gibt ein winziges Bild mit einer erkennbaren Ecke.
//
// **Erkennbar und nicht einfarbig**: ein einfarbiges Bild sagt nicht, ob die
// Quelle die richtige Datei gelesen hat oder irgendeine.
func gemaltesBild(ecke color.NRGBA) image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	bild.SetNRGBA(0, 0, ecke)

	return bild
}

// legePNG schreibt ein Bild als PNG an seinen Ort, samt Verzeichnissen.
func legePNG(t *testing.T, pfad string, ecke color.NRGBA) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	var puffer bytes.Buffer
	if err := png.Encode(&puffer, gemaltesBild(ecke)); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	if err := os.WriteFile(pfad, puffer.Bytes(), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// ecke liest den Bildpunkt oben links.
func ecke(t *testing.T, bild image.Image) color.NRGBA {
	t.Helper()

	rot, gruen, blau, alpha := bild.At(bild.Bounds().Min.X, bild.Bounds().Min.Y).RGBA()

	return color.NRGBA{
		R: uint8(rot >> 8), G: uint8(gruen >> 8), B: uint8(blau >> 8), A: uint8(alpha >> 8),
	}
}

// **Ein Bild kommt aus dem Paket, das sein Pfad nennt.**
//
// Der Plan trägt Paketpfade und keine Dateisystempfade
// ([render.Ebene]): `/probe-bild.mse-style/rahmen.png` sagt **Paket** und
// **Datei**, und wo das Paket liegt, weiß erst diese Quelle.
func TestEinBildKommtAusSeinemPaket(t *testing.T) {
	wurzel := t.TempDir()
	legePNG(t, filepath.Join(wurzel, "probe-bild.mse-style", "rahmen.png"),
		color.NRGBA{R: 12, G: 34, B: 56, A: 255})

	bild, err := msepaket.NeueBildquelle(wurzel).Bild("/probe-bild.mse-style/rahmen.png")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen := ecke(t, bild); gelesen != (color.NRGBA{R: 12, G: 34, B: 56, A: 255}) {
		t.Errorf("Ecke %+v", gelesen)
	}
}

// **Mehrere Wurzeln, und die erste gewinnt.**
//
// Die Saat liegt im Programm, bezogene Pakete liegen bei den Nutzenden. Beide
// Orte sind gültig, und die Reihenfolge ist die Aussage: wer ein Paket bezogen
// hat, meint das bezogene.
func TestDieWurzelnWerdenDerReiheNachDurchsucht(t *testing.T) {
	erste, zweite := t.TempDir(), t.TempDir()

	legePNG(t, filepath.Join(erste, "beide.mse-style", "x.png"),
		color.NRGBA{R: 1, A: 255})
	legePNG(t, filepath.Join(zweite, "beide.mse-style", "x.png"),
		color.NRGBA{R: 2, A: 255})
	legePNG(t, filepath.Join(zweite, "nur-hinten.mse-style", "x.png"),
		color.NRGBA{R: 3, A: 255})

	quelle := msepaket.NeueBildquelle(erste, zweite)

	vorn, err := quelle.Bild("/beide.mse-style/x.png")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if ecke(t, vorn).R != 1 {
		t.Errorf("die zweite Wurzel hat gewonnen: %+v", ecke(t, vorn))
	}

	hinten, err := quelle.Bild("/nur-hinten.mse-style/x.png")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if ecke(t, hinten).R != 3 {
		t.Errorf("Ecke %+v", ecke(t, hinten))
	}
}

// **Ein Paket darf ein Archiv sein.**
//
// `msepaket.Oeffne` kann beides seit Phase 2, und ein bezogenes Paket liegt
// mal so und mal so vor. Eine Bildquelle, die nur Verzeichnisse läse, fände
// die Hälfte nicht.
func TestEinBildKommtAuchAusEinemArchiv(t *testing.T) {
	wurzel := t.TempDir()
	pfad := filepath.Join(wurzel, "gepackt.mse-style")

	datei, err := os.Create(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	archiv := zip.NewWriter(datei)

	eintrag, err := archiv.Create("rahmen.png")
	if err != nil {
		t.Fatalf("Eintrag: %v", err)
	}

	if fehler := png.Encode(eintrag, gemaltesBild(color.NRGBA{G: 200, A: 255})); fehler != nil {
		t.Fatalf("kodieren: %v", fehler)
	}

	if fehler := errors.Join(archiv.Close(), datei.Close()); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	bild, err := msepaket.NeueBildquelle(wurzel).Bild("/gepackt.mse-style/rahmen.png")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if ecke(t, bild).G != 200 {
		t.Errorf("Ecke %+v", ecke(t, bild))
	}
}

// **JPEG gehört dazu.** `magic-m15.mse-style` setzt seinen Kartengrund über
// `template_suffix := [card: "card.jpg"]` — wer nur PNG entschlüsselte,
// zeichnete bei jeder M15-Karte den Hintergrund nicht.
func TestEinBildDarfAuchJPEGSein(t *testing.T) {
	wurzel := t.TempDir()
	ordner := filepath.Join(wurzel, "probe-bild.mse-style")

	if err := os.MkdirAll(ordner, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	var puffer bytes.Buffer
	if err := jpeg.Encode(&puffer, gemaltesBild(color.NRGBA{B: 255, A: 255}), nil); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ordner, "grund.jpg"), puffer.Bytes(), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := msepaket.NeueBildquelle(wurzel).Bild("/probe-bild.mse-style/grund.jpg"); err != nil {
		t.Fatalf("lesen: %v", err)
	}
}

// **Ein Bild, das es nicht gibt, wird benannt** — mit dem Verweis und mit dem,
// was auf jedem Weg schiefging.
//
// Ein `errKeinBild` ohne die einzelnen Fehlschläge sagte „nicht gefunden" und
// verschwiege, dass eine der Wurzeln unlesbar ist. Das sind zwei verschiedene
// Lagen mit zwei verschiedenen Abhilfen.
func TestEinFehlendesBildWirdBenannt(t *testing.T) {
	// **Zwei Lagen, und beide gehören geprüft**: das Paket fehlt ganz, oder es
	// liegt da und die Datei darin fehlt. Der erste Anlauf prüfte nur die
	// erste — und der Leseweg, der die zweite meldet, lief in keinem Test.
	ohnePaket := t.TempDir()

	mitPaket := t.TempDir()
	legePNG(t, filepath.Join(mitPaket, "probe-bild.mse-style", "andere.png"),
		color.NRGBA{A: 255})

	for name, wurzel := range map[string]string{
		"das Paket fehlt": ohnePaket, "die Datei im Paket fehlt": mitPaket,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := msepaket.NeueBildquelle(wurzel).Bild("/probe-bild.mse-style/fehlt.png")

			if !errors.Is(err, msepaket.ErrKeinBild) {
				t.Fatalf("Fehler %v, erwartet ErrKeinBild", err)
			}

			if !bytes.Contains([]byte(err.Error()), []byte("fehlt.png")) {
				t.Errorf("die Meldung nennt die Datei nicht: %v", err)
			}
		})
	}
}

// **Ein Verweis, der kein Paketpfad ist, wird abgewiesen.**
//
// `rahmen.png` ohne Paket ist keine Kleinigkeit, die man auffangen könnte:
// welches Paket gemeint ist, entscheidet der Stil, und diese Quelle darf es
// nicht raten. `..` ebenso — das ist der Ausbruch, den
// [msepaket.ZerlegePaketpfad] seit Phase 3 abweist.
func TestEinVerweisOhnePaketWirdAbgewiesen(t *testing.T) {
	quelle := msepaket.NeueBildquelle(t.TempDir())

	for _, verweis := range []string{"rahmen.png", "/../heimlich.png", "/probe.mse-style/../../x"} {
		if _, err := quelle.Bild(verweis); !errors.Is(err, msepaket.ErrPaketpfad) {
			t.Errorf("%q: Fehler %v, erwartet ErrPaketpfad", verweis, err)
		}
	}
}

// **Eine Datei, die kein Bild ist, wird als solche gemeldet** — und nicht als
// fehlend. Fremde Daten sind der Normalfall (§8), und ein halb geschriebenes
// PNG ist ein anderer Befund als ein fehlendes.
func TestEineUnlesbareBilddateiWirdGemeldet(t *testing.T) {
	wurzel := t.TempDir()
	ordner := filepath.Join(wurzel, "probe-bild.mse-style")

	if err := os.MkdirAll(ordner, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ordner, "halb.png"), []byte("\x89PNG kaputt"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	_, err := msepaket.NeueBildquelle(wurzel).Bild("/probe-bild.mse-style/halb.png")

	if err == nil || errors.Is(err, msepaket.ErrKeinBild) {
		t.Errorf("Fehler %v, erwartet einen über die Bilddaten", err)
	}
}

// **Eine leere Wurzel wird verworfen, nicht durchsucht.**
//
// `app.Paketverzeichnis()` gibt einen leeren Pfad, solange niemand ein Paket
// bezogen hat. `filepath.Join("", "x.mse-style")` ergäbe daraus einen
// **relativen** Pfad — gesucht würde dann im Arbeitsverzeichnis des Prozesses,
// und was dort gefunden wird, hat mit Paketen nichts zu tun. Genau diese Falle
// hat der Zweitblick über Phase 9 schon einmal beim Paketbezieher gefunden.
func TestEineLeereWurzelWirdVerworfen(t *testing.T) {
	_, err := msepaket.NeueBildquelle("", t.TempDir()).Bild("/probe.mse-style/x.png")

	if !errors.Is(err, msepaket.ErrKeinBild) {
		t.Fatalf("Fehler %v, erwartet ErrKeinBild", err)
	}

	// Die leere Wurzel taucht in keinem Versuch auf — sonst stünde ein
	// relativer Pfad in der Meldung.
	if bytes.Contains([]byte(err.Error()), []byte(`"probe.mse-style"`)) {
		t.Errorf("die leere Wurzel wurde durchsucht: %v", err)
	}
}

// **`Gibtes` fragt, ob eine Datei in einem Paket liegt — ohne sie zu lesen.**
//
// Die Vorlagen fragen danach:
//
//	if exists_in_package(path) then path else "/magic-modules.mse-include/loyalty/" + name
//
// `magic.mse-game/script`, Zeile 3614. **126 Befunde** über die 190 Stile von
// `M15-Magic-Pack` lauteten auf diesen Namen
// (#270).
//
// **Ohne zu entschlüsseln**: gefragt wird nach jeder Datei, nicht nur nach
// Bildern — die Vorlagen prüfen auch Einschübe. Ein Leser, der das Bild
// dekodierte, sagte bei einer Textdatei „nein", obwohl sie dasteht.
func TestGibtesFragtObEineDateiDaIst(t *testing.T) {
	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe.mse-include")

	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(paket, "text"), []byte("kein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle := msepaket.NeueBildquelle(wurzel)

	faelle := map[string]bool{
		"/probe.mse-include/text":       true,
		"/probe.mse-include/gibtsnicht": false,
		"/gibtsnicht.mse-include/text":  false,

		// **Ein unbrauchbarer Verweis ist ein Nein**, kein Fehler: die Vorlage
		// fragt, um sich zu entscheiden, und eine Ausnahme brächte sie um ihre
		// Entscheidung.
		"":            false,
		"../draussen": false,
	}

	for verweis, erwartet := range faelle {
		if haben := quelle.Gibtes(verweis); haben != erwartet {
			t.Errorf("Gibtes(%q) ist %v, erwartet %v", verweis, haben, erwartet)
		}
	}
}

// **Ein schon gelesenes Bild ist da, ohne noch einmal nachzusehen.**
//
// [Bildquelle.Bild] merkt sich, was es gelesen hat; [Bildquelle.Gibtes] fragt
// zuerst dort. Ohne diesen Schritt liefe jede Existenzfrage über die Platte —
// und `magic.mse-game` stellt sie in **jedem** Bildskript.
func TestGibtesFragtZuerstDenSpeicher(t *testing.T) {
	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe.mse-include")

	legePNG(t, filepath.Join(paket, "bild.png"), color.NRGBA{R: 200, A: 255})

	quelle := msepaket.NeueBildquelle(wurzel)

	if _, err := quelle.Bild("/probe.mse-include/bild.png"); err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if err := os.RemoveAll(paket); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if !quelle.Gibtes("/probe.mse-include/bild.png") {
		t.Error("das gemerkte Bild gilt nicht als vorhanden")
	}
}
