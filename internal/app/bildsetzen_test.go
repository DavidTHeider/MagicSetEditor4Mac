package app_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Bildfeld lässt sich füllen** —
// #246.
//
// # Der Weg, den MSE geht
//
//	String filename = wxFileSelector(_("Open image file"), …);
//	wxImage image(filename);
//	sliceImage(image);
//	  → ImageSliceWindow s(…, image, style().getSize(), mask);
//	  → LocalFileName f = getLocalPackage().newFileName(field().name, "");
//	  → img.SaveFile(getLocalPackage().nameOut(f), wxBITMAP_TYPE_PNG);
//	  → addAction(value_action(valueP(), f));
//
// `ImageValueEditor::onLeftDClick` und `sliceImage`, `src/gui/value/image.cpp`,
// Zeilen 22–53, Fassung f9d9356 vom 2021-07-06.
//
// Also: **wählen, zuschneiden, ablegen, den Namen in das Feld schreiben** — und
// der Name ist es, der im Set steht, nicht der Pfad, aus dem das Bild kam.
func TestEinBildfeldLaesstSichFuellen(t *testing.T) {
	wurzel := t.TempDir()
	quelle := filepath.Join(t.TempDir(), "urlaub.jpg")
	schreibeJPEG(t, quelle)

	schreiber := &schreibprobe{}
	anwendung := bildanwendung(t, wurzel, schreiber, quelle)

	// **Zwei Schritte seit #576**:
	// `BildSetzen` bereitet vor, das Zuschnittfenster zeigt den Ausschnitt,
	// `BildEinsetzen` legt ab. Wer nichts ändert, reicht die Vorgabewahl
	// durch — und bekommt genau dieses Ergebnis.
	vorbereitet := anwendung.BildSetzen(t.Context(), "1", "kunst")
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	antwort := anwendung.BildEinsetzen(
		probekontext(t), "1", "1", "kunst", vorbereitet.Pfad, vorbereitet.Auswahl)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Der Name steht im Feld, nicht der Pfad.** Ein Pfad zeigte auf diesen
	// Rechner; weitergegeben öffnete das Set mit leerem Kunstfenster.
	if schreiber.feld != "kunst" || schreiber.wert != "kunst1" {
		t.Errorf("geschrieben wurde %q = %q", schreiber.feld, schreiber.wert)
	}

	// **Und die Datei liegt neben dem Set** — unter der Setkennung, nicht der
	// Kartenkennung: ein Set hat ein Verzeichnis, keine dreitausend.
	abgelegt := filepath.Join(wurzel, "7", "kunst1")
	if _, err := os.Stat(abgelegt); err != nil {
		t.Fatalf("%s fehlt: %v", abgelegt, err)
	}

	// **Zugeschnitten auf die Fläche des Feldes.** Die Quelle ist 400 × 100,
	// das Feld 100 × 80 — abgelegt gehört, was hineinpasst, nicht die Quelle.
	datei, err := os.Open(abgelegt) //nolint:gosec // eben geschriebener Testpfad
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := datei.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	bild, _, err := image.Decode(datei)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if bild.Bounds() != image.Rect(0, 0, 100, 80) {
		t.Errorf("abgelegt wurde %v statt 100 × 80", bild.Bounds())
	}
}

// **Ein abgebrochener Dialog ist kein Fehler.**
//
// Wer die Dateiwahl abbricht, hat sich entschieden — eine Meldung darüber wäre
// eine Rückfrage auf eine Antwort.
func TestEinAbgebrochenerDialogSchreibtNichts(t *testing.T) {
	schreiber := &schreibprobe{}
	anwendung := bildanwendung(t, t.TempDir(), schreiber, "")

	antwort := anwendung.BildSetzen(t.Context(), "1", "kunst")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if schreiber.feld != "" {
		t.Errorf("geschrieben wurde %q", schreiber.feld)
	}
}

// **Eine Datei, die kein Bild ist, wird gemeldet** (§8) — und nichts
// geschrieben.
func TestEineUnlesbareDateiWirdGemeldet(t *testing.T) {
	quelle := filepath.Join(t.TempDir(), "kein-bild.jpg")
	if err := os.WriteFile(quelle, []byte("guten Tag"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	schreiber := &schreibprobe{}
	anwendung := bildanwendung(t, t.TempDir(), schreiber, quelle)

	if antwort := anwendung.BildSetzen(t.Context(), "1", "kunst"); antwort.Fehler == nil {
		t.Error("die unlesbare Datei kam durch")
	}

	if schreiber.feld != "" {
		t.Errorf("geschrieben wurde %q", schreiber.feld)
	}
}

// **Ein Feld, das nicht auf der Karte liegt, hat keine Größe** — und damit
// nichts, worauf zugeschnitten würde.
func TestEinFeldOhneFlaecheWirdAbgewiesen(t *testing.T) {
	quelle := filepath.Join(t.TempDir(), "urlaub.jpg")
	schreibeJPEG(t, quelle)

	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{}, quelle)

	if antwort := anwendung.BildSetzen(t.Context(), "1", "gibtsnicht"); antwort.Fehler == nil {
		t.Error("ein Feld ohne Fläche kam durch")
	}
}

// **Und eine unlesbare Kennung** — dieselbe Wache wie überall.
func TestBildSetzenWeistEineUnlesbareKennungAb(t *testing.T) {
	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{}, "")

	if antwort := anwendung.BildSetzen(t.Context(), "keine zahl", "kunst"); antwort.Fehler == nil {
		t.Error("eine unlesbare Kennung kam durch")
	}
}

// bildanwendung baut eine Anwendung, deren Dateidialog `gewaehlt` antwortet.
func bildanwendung(
	t *testing.T, wurzel string, schreiber *schreibprobe, gewaehlt string,
) *app.App {
	t.Helper()

	anwendung := app.NeueApp("")
	anwendung.MitEditor(schreiber,
		karteprobe{karte: &domain.CardRow{ID: 1, SetID: 7}},
		&bildplanbauer{}, render.NeuerRasterer(ohneBilder{}))
	anwendung.MitBildablage(service.NeueBildablage(wurzel))
	// **Ein leerer Pfad heißt „abgebrochen"** — dieselbe Übereinkunft wie bei
	// `SetOeffnen` (`oeffnen.go`, Zeilen 72–77): „nichts gewählt" ist eine
	// Aussage und kein Fehler.
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return gewaehlt, nil
	})

	return anwendung
}

// bildplanbauer legt ein Feld `kunst` auf 100 × 80.
type bildplanbauer struct {
	ohneWortlisten
	ohneWahlbilder
}

func (bildplanbauer) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{
		Breite: 100, Hoehe: 140,
		Feldflaechen: []render.Feldflaeche{{
			Feld:   "kunst",
			Rahmen: render.Rechteck{Breite: 100, Hoehe: 80},
		}, {
			// **Ein zweites Bildfeld** (A12, #1451): eine Karte hat
			// Illustration **und** Wasserzeichen, und ohne das zweite
			// ließe sich nicht prüfen, dass ein Vorschlag nur in das Feld
			// geht, für das er erzeugt wurde.
			Feld:   "wasserzeichen",
			Rahmen: render.Rechteck{Breite: 20, Hoehe: 20},
		}},
	}, nil, nil
}

func (bildplanbauer) Kartenpalette(context.Context, *domain.CardRow, string) (
	[]service.Paletteneintrag, bool, []domain.Befund, error,
) {
	return nil, false, nil, nil
}

// schreibeJPEG legt ein Probebild an — **400 × 100**, also viermal so breit
// wie hoch. Das Feld im [bildplanbauer] ist 100 × 80; damit greift der
// Zuschnitt sichtbar, statt die Quelle nur umzurechnen.
func schreibeJPEG(t *testing.T, pfad string) {
	const breite, hoehe = 400, 100

	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, breite, hoehe))
	for x := range breite {
		for y := range hoehe {
			bild.Set(x, y, color.RGBA{R: uint8(x % 256), A: 255}) //nolint:gosec // Testbild
		}
	}

	var gepackt bytes.Buffer
	if err := jpeg.Encode(&gepackt, bild, nil); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	if err := os.WriteFile(pfad, gepackt.Bytes(), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// **Ein leerer Feldname wird abgewiesen** — dieselbe Wache wie beim Schreiben:
// er ergäbe eine Zeile, die niemand wiederfindet.
func TestBildSetzenWeistEinenLeerenFeldnamenAb(t *testing.T) {
	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{}, "")

	if antwort := anwendung.BildSetzen(t.Context(), "1", "   "); antwort.Fehler == nil {
		t.Error("ein leerer Feldname kam durch")
	}
}

// **Eine Karte, die es nicht gibt, ist ein Befund.**
func TestBildSetzenMeldetEineFehlendeKarte(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{}, karteprobe{fehler: domain.ErrNotFound},
		&bildplanbauer{}, render.NeuerRasterer(ohneBilder{}))
	anwendung.MitBildablage(service.NeueBildablage(t.TempDir()))

	if antwort := anwendung.BildSetzen(t.Context(), "1", "kunst"); antwort.Fehler == nil {
		t.Error("die fehlende Karte kam durch")
	}
}

// **Ein Feld ohne Ausdehnung ist keine Fläche.**
//
// Ein Feldstil darf Breite oder Höhe null tragen — `magic-m15-saga` schreibt
// `width: {if is_saga() then 0 else 153}`, und dann ist das Feld unsichtbar
// (`ValueViewer::isVisible`, `src/render/value/viewer.cpp`, Zeilen 36–42,
// Fassung f9d9356 vom 2021-07-06). Ein Zuschnitt darauf ergäbe ein leeres
// Bild, das im Set aussähe wie eine Illustration.
func TestEinFeldOhneAusdehnungWirdAbgewiesen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{},
		karteprobe{karte: &domain.CardRow{ID: 1, SetID: 7}},
		&flachesfeld{}, render.NeuerRasterer(ohneBilder{}))
	anwendung.MitBildablage(service.NeueBildablage(t.TempDir()))

	if antwort := anwendung.BildSetzen(t.Context(), "1", "kunst"); antwort.Fehler == nil {
		t.Error("ein Feld ohne Ausdehnung kam durch")
	}
}

// **Ein Plan, der nicht entsteht, wird gemeldet** — und nicht auf eine
// erfundene Größe zugeschnitten.
func TestEinGescheiterterPlanWirdGemeldet(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{},
		karteprobe{karte: &domain.CardRow{ID: 1, SetID: 7}},
		&kaputterplanbauer{}, render.NeuerRasterer(ohneBilder{}))
	anwendung.MitBildablage(service.NeueBildablage(t.TempDir()))

	if antwort := anwendung.BildSetzen(t.Context(), "1", "kunst"); antwort.Fehler == nil {
		t.Error("der gescheiterte Plan kam durch")
	}
}

// **Eine Datei, die es nicht gibt, wird gemeldet.**
//
// Der Dialog gibt einen Pfad; zwischen der Wahl und dem Öffnen kann jemand die
// Datei verschieben.
func TestEineFehlendeBilddateiWirdGemeldet(t *testing.T) {
	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{},
		filepath.Join(t.TempDir(), "gibtsnicht.png"))

	if antwort := anwendung.BildSetzen(t.Context(), "1", "kunst"); antwort.Fehler == nil {
		t.Error("die fehlende Datei kam durch")
	}
}

// **Und eine Ablage, die nicht schreiben kann** (§8).
func TestEineGescheiterteAblageWirdGemeldet(t *testing.T) {
	quelle := filepath.Join(t.TempDir(), "urlaub.jpg")
	schreibeJPEG(t, quelle)

	// Eine Wurzel, unter der sich kein Verzeichnis anlegen lässt.
	wurzel := filepath.Join(t.TempDir(), "keine ablage")
	if err := os.WriteFile(wurzel, []byte("Datei statt Verzeichnis"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	anwendung := bildanwendung(t, wurzel, &schreibprobe{}, quelle)

	vorbereitet := anwendung.BildSetzen(t.Context(), "1", "kunst")
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	if antwort := anwendung.BildEinsetzen(
		probekontext(t), "1", "1", "kunst", vorbereitet.Pfad, vorbereitet.Auswahl,
	); antwort.Fehler == nil {
		t.Error("die gescheiterte Ablage kam durch")
	}
}

// flachesfeld legt `kunst` auf eine Fläche ohne Ausdehnung.
type flachesfeld struct{ bildplanbauer }

func (flachesfeld) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{
		Breite: 100, Hoehe: 140,
		Feldflaechen: []render.Feldflaeche{{Feld: "kunst"}},
	}, nil, nil
}

// kaputterplanbauer scheitert am Plan.
type kaputterplanbauer struct{ bildplanbauer }

func (kaputterplanbauer) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return nil, nil, errors.New("planmäßig gescheitert")
}

// **Ein Dateidialog, der scheitert, wird gemeldet** — und ist etwas anderes
// als ein abgebrochener.
//
// Abgebrochen heißt „nichts gewählt" und kommt als leerer Pfad; ein Fehler
// heißt, dass der Dialog gar nicht erst aufging. Wer beides gleich behandelte,
// verschwiege den zweiten Fall.
func TestEinGescheiterterDateidialogWirdGemeldet(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{},
		karteprobe{karte: &domain.CardRow{ID: 1, SetID: 7}},
		&bildplanbauer{}, render.NeuerRasterer(ohneBilder{}))
	anwendung.MitBildablage(service.NeueBildablage(t.TempDir()))
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("kein Fenster")
	})

	if antwort := anwendung.BildSetzen(t.Context(), "1", "kunst"); antwort.Fehler == nil {
		t.Error("der gescheiterte Dialog kam durch")
	}
}
