package app_test

// **Der Weg eines erzeugten Bildes durch die Naht** — Werkstück A12 von
// Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass ein Vorschlag ein Vorschlag bleibt.** Das Ziel des Issues steht in
// einem Satz: „Generierte oder bearbeitete Bilder bleiben Vorschläge, bis
// Nutzende sie mit vorhandener Bildverarbeitung übernehmen."
//
// **Und dass nur der Besitzer herankommt.** Die Kennung ist der ganze
// Schutz; was sie nicht schützt, schützt niemand.

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildlaufzeit ist eine Laufzeit, die **auch** Bilder erzeugt.
//
// **Als eigener Typ und nicht als Feld am vorhandenen Doppel**: ob ein
// Anbieter Bilder kann, entscheidet die Naht über eine Typzusicherung
// ([app.KIBildquelle]). Ein Feld daneben prüfte etwas anderes als der
// Betrieb.
type bildlaufzeit struct {
	*laufzeitdoppel

	bild    app.KIRohbild
	fehler  error
	gesehen []app.KIBildwunsch
}

func (b *bildlaufzeit) Bild(_ context.Context, wunsch app.KIBildwunsch) (
	app.KIRohbild, error,
) {
	b.gesehen = append(b.gesehen, wunsch)

	if b.fehler != nil {
		return app.KIRohbild{}, b.fehler
	}

	return b.bild, nil
}

// einBildkoenner baut ein Modell, das Bilder erzeugen **und** bearbeiten kann.
func einBildkoenner(alias string) domain.KIModell {
	modell := einKoenner(alias)
	for _, faehigkeit := range []domain.KIFaehigkeit{
		domain.KIBildgenerierung, domain.KIBildbearbeitung,
	} {
		modell.Geprueft = append(modell.Geprueft, domain.KIGeprueft{
			Faehigkeit: faehigkeit, Datum: "2026-09-12",
			Beleg: "gemessen an einem Auftrag mit fester Größe",
		})
	}

	return modell
}

// einProbebild malt ein PNG.
func einProbebild(t *testing.T) []byte {
	t.Helper()

	flaeche := image.NewNRGBA(image.Rect(0, 0, 8, 12))
	flaeche.Set(0, 0, color.NRGBA{R: 255, A: 255})

	var aus bytes.Buffer
	if fehler := png.Encode(&aus, flaeche); fehler != nil {
		t.Fatalf("das Prüfbild ließ sich nicht schreiben: %v", fehler)
	}

	return aus.Bytes()
}

// einProbeJPEG malt dasselbe als JPEG.
func einProbeJPEG(t *testing.T) []byte {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.jpg")
	schreibeJPEG(t, pfad)

	inhalt, fehler := os.ReadFile(pfad) //nolint:gosec // eigene Datei im Lauf
	if fehler != nil {
		t.Fatalf("das Prüf-JPEG ließ sich nicht lesen: %v", fehler)
	}

	return inhalt
}

// bildwelt ist eine kiwelt mit Bildvorrat und bildfähiger Laufzeit.
type bildwelt struct {
	kiwelt

	laufzeitbild *bildlaufzeit
	vorrat       *service.Bildvorrat
	bild         []byte
}

// einebildwelt baut sie.
func einebildwelt(t *testing.T) bildwelt {
	t.Helper()

	welt := einekiwelt(t)

	bild := einProbebild(t)

	// **Das Modell muss die Fähigkeit belegen** (A6, #1445): eine Liste ist
	// keine Fähigkeit, und ein Auftrag an ein Modell, das es nicht kann,
	// kostet Geld für einen Fehler.
	welt.laufzeit.modelle = []domain.KIModell{einBildkoenner("gemma4:8b")}

	laufzeit := &bildlaufzeit{
		laufzeitdoppel: welt.laufzeit,
		bild: app.KIRohbild{
			Inhalt: bild, Medientyp: "image/png",
			Modell: "gpt-image-2.5-flare", Umschrift: "ein Blitz, Ölgemälde",
		},
	}

	vorrat, fehler := service.NeuerBildvorrat(t.TempDir(),
		service.UeblicheBildgrenzen(), time.Now)
	if fehler != nil {
		t.Fatalf("der Vorrat ließ sich nicht bauen: %v", fehler)
	}

	welt.anwendung.MitKILaufzeit(
		func(string) app.KILaufzeitquelle { return laufzeit })
	welt.anwendung.MitBildvorrat(vorrat)

	// **Planbauer und Ablage kommen dazu**: der Zuschnitt braucht die
	// Fläche des Feldes, das Einsetzen einen Ort. `einekiwelt` hat beides
	// nicht — dort geht es um Text.
	welt.anwendung.MitEditor(welt.karten, welt.karten, &bildplanbauer{}, nil)
	welt.anwendung.MitBildablage(service.NeueBildablage(t.TempDir()))

	return bildwelt{kiwelt: welt, laufzeitbild: laufzeit, vorrat: vorrat, bild: bild}
}

// vorschlagen lässt ein Bild erzeugen.
func (b bildwelt) vorschlagen(t *testing.T) app.KIBildvorschlagAntwort {
	t.Helper()

	return b.anwendung.KIBildVorschlagen(probekontext(t), b.karte, "kunst",
		"ollama", "gemma4:8b", "ein Blitz", domain.KIBildgroesseHoch,
		domain.KIBildgueteMittel, false, "")
}

// **Ein erzeugtes Bild ist ein Vorschlag und keine Änderung.**
func TestEinErzeugtesBildAendertDieKarteNicht(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	vorher := welt.kartenwerte(t)

	antwort := welt.vorschlagen(t)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Kennung == "" || antwort.Vorschau == "" {
		t.Fatalf("die Antwort ist %+v", antwort.KIBildvorschlagDTO)
	}

	if antwort.Breite != 8 || antwort.Hoehe != 12 {
		t.Errorf("die Maße sind %d×%d", antwort.Breite, antwort.Hoehe)
	}

	// **Die Umschrift ist Herkunft**: gezeichnet wurde, was hier steht.
	if antwort.Umschrift != "ein Blitz, Ölgemälde" {
		t.Errorf("die Umschrift ist %q", antwort.Umschrift)
	}

	// **Und die Angaben nennen nur, was feststeht.**
	if len(antwort.Angaben) != 4 {
		t.Errorf("die Angaben sind %+v", antwort.Angaben)
	}

	if nachher := welt.kartenwerte(t); nachher != vorher {
		t.Errorf("die Karte wurde geändert: %q statt %q", nachher, vorher)
	}
}

// **Was der Dienst nicht sagt, steht nicht in den Angaben.**
//
// **Die Ergänzung des Issues verlangt es**: „Fehlende Angaben als unbekannt
// führen." Eine Angabe `guete: ""` behauptete, es gäbe eine und sie sei leer.
func TestFehlendeAngabenStehenNichtDa(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.laufzeitbild.bild.Umschrift = ""

	antwort := welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "ollama", "gemma4:8b", "ein Blitz", "", "", false, "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, angabe := range antwort.Angaben {
		if angabe.Name != "medientyp" {
			t.Errorf("die Angabe %q steht da, obwohl niemand sie genannt hat",
				angabe.Name)
		}
	}
}

// **Ein unvollständiger Auftrag geht gar nicht erst hinaus.**
//
// **Vor dem Ruf und nicht danach**: jeder dieser Fälle kostete sonst Geld für
// einen Fehler, den diese Seite selbst kennt.
func TestEinUnvollstaendigerBildauftragGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		feld, aufgabe, groesse, guete string
	}{
		"ohne Feld":       {"", "ein Blitz", domain.KIBildgroesseHoch, ""},
		"ohne Aufgabe":    {"kunst", "", domain.KIBildgroesseHoch, ""},
		"erfundene Größe": {"kunst", "ein Blitz", "4000x4000", ""},
		"erfundene Güte":  {"kunst", "ein Blitz", "", "beste"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			welt := einebildwelt(t)

			antwort := welt.anwendung.KIBildVorschlagen(probekontext(t),
				welt.karte, fall.feld, "ollama", "gemma4:8b", fall.aufgabe,
				fall.groesse, fall.guete, false, "")
			if antwort.Fehler == nil {
				t.Fatalf("%q kam durch", name)
			}

			if len(welt.laufzeitbild.gesehen) != 0 {
				t.Errorf("es ging trotzdem ein Auftrag hinaus: %+v",
					welt.laufzeitbild.gesehen)
			}
		})
	}
}

// **Ein Anbieter, der keine Bilder erzeugt, sagt es.**
//
// **Und die Meldung nennt ihn.** Drei von vier Anbietern können es nicht
// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md));
// „geht nicht" ohne Namen hilft dabei niemandem.
func TestEinAnbieterOhneBilderSagtEs(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorrat, fehler := service.NeuerBildvorrat(t.TempDir(),
		service.UeblicheBildgrenzen(), time.Now)
	if fehler != nil {
		t.Fatalf("der Vorrat ließ sich nicht bauen: %v", fehler)
	}

	welt.anwendung.MitBildvorrat(vorrat)

	antwort := welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "ollama", "gemma4:8b", "ein Blitz", "", "", false, "")
	if antwort.Fehler == nil {
		t.Fatal("die Laufzeit ohne Bildweg kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "ollama") {
		t.Errorf("die Meldung nennt den Anbieter nicht: %s",
			antwort.Fehler.Meldung)
	}
}

// **Ohne Vorrat gibt es keine Bildvorschläge — und das steht da.**
func TestOhneVorratGibtEsKeineBildvorschlaege(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	if antwort := welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "ollama", "gemma4:8b", "ein Blitz", "", "",
		false, ""); antwort.Fehler == nil {
		t.Error("der Vorschlag gelang ohne Vorrat")
	}

	if antwort := welt.anwendung.KIBildVerwerfen(
		probekontext(t), "egal"); antwort.Fehler == nil {
		t.Error("das Verwerfen gelang ohne Vorrat")
	}

	if antwort := welt.anwendung.KIBildZuschnitt(probekontext(t), "4", "kunst",
		"egal"); antwort.Fehler == nil {
		t.Error("der Zuschnitt gelang ohne Vorrat")
	}
}

// **Scheitert der Anbieter, entsteht kein Vorschlag.**
func TestEinGescheiterterAnbieterHinterlaesstNichts(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.laufzeitbild.fehler = domain.MitGrund(domain.ErrConflict, nil,
		"der Dienst hat abgebrochen")

	if antwort := welt.vorschlagen(t); antwort.Fehler == nil {
		t.Fatal("der gescheiterte Auftrag kam als Vorschlag zurück")
	}

	if weg := welt.vorrat.Aufraeumen(); weg != 0 {
		t.Errorf("es liegen %d Einträge im Vorrat", weg)
	}
}

// **Ein Bild, das der Vorrat abweist, wird gemeldet.**
//
// **Der Fall ist die Bildbombe**, eine Ebene höher: was die Schranken nicht
// passiert, darf nicht als Vorschlag erscheinen.
func TestEinAbgewiesenesBildWirdGemeldet(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.laufzeitbild.bild.Inhalt = []byte("das ist kein Bild")

	if antwort := welt.vorschlagen(t); antwort.Fehler == nil {
		t.Fatal("das Nichtbild kam als Vorschlag zurück")
	}
}
