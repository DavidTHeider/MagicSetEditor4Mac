package app_test

// **Ein Bild ohne Dialog** — für das Ziehen und Ablegen
// (#246).
//
// Eine eigene Datei, weil `bildsetzen_test.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht der Weg über den Dateidialog,
// hier der über einen Pfad, der schon feststeht.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Bild lässt sich auch ohne Dialog einsetzen.**
//
// # Wofür
//
// Wer eine Datei in das Fenster **zieht**, hat sie schon gewählt. Wails gibt
// dabei die **absoluten Pfade** (`runtime.OnFileDrop`), und ein Dialog danach
// wäre eine Rückfrage auf eine Antwort.
//
// **Derselbe Weg dahinter**: lesen, zuschneiden, ablegen, den Namen schreiben.
// Zwei Wege wären zwei Stellen, an denen der Zuschnitt auseinanderläuft.
//
// **Seit #576
// sind es zwei Schritte**: `BildAusDatei` **bereitet** vor, das Fenster zeigt
// den Ausschnitt, `BildEinsetzen` legt ab. Vorher schrieb dieser Aufruf selbst
// — eine benannte Änderung (§15, Satz 4), ohne die niemand den Ausschnitt zu
// sehen bekäme.
func TestEinBildLaesstSichOhneDialogEinsetzen(t *testing.T) {
	quelle := filepath.Join(t.TempDir(), "urlaub.jpg")
	schreibeJPEG(t, quelle)

	wurzel := t.TempDir()
	schreiber := &schreibprobe{}

	// **Ohne Dateiwahl.** Wird sie doch gerufen, ist das der Befund.
	anwendung := app.NeueApp("")
	anwendung.MitEditor(schreiber,
		karteprobe{karte: &domain.CardRow{ID: 1, SetID: 7}},
		&bildplanbauer{}, render.NeuerRasterer(ohneBilder{}))
	anwendung.MitBildablage(service.NeueBildablage(wurzel))
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		t.Error("der Dateidialog ging auf, obwohl ein Pfad dastand")

		return "", nil
	})

	vorbereitet := anwendung.BildAusDatei(t.Context(), "1", "kunst", quelle)
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	antwort := anwendung.BildEinsetzen(
		probekontext(t), "1", "1", "kunst", vorbereitet.Pfad, vorbereitet.Auswahl)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if schreiber.feld != "kunst" || schreiber.wert != "kunst1" {
		t.Errorf("geschrieben wurde %q = %q", schreiber.feld, schreiber.wert)
	}

	if _, err := os.Stat(filepath.Join(wurzel, "7", "kunst1")); err != nil {
		t.Errorf("das Bild fehlt: %v", err)
	}
}

// **Eine unlesbare Kennung wird auch hier abgewiesen** — dieselbe Wache für
// beide Wege, an einer Stelle ([App.bildziel]).
func TestBildAusDateiWeistEineUnlesbareKennungAb(t *testing.T) {
	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{}, "")

	if antwort := anwendung.BildAusDatei(
		probekontext(t), "keine zahl", "kunst", "/tmp/egal.png"); antwort.Fehler == nil {
		t.Error("eine unlesbare Kennung kam durch")
	}
}

// **Ein leerer Pfad ist ein Befund, kein Abbruch.**
//
// Beim Dialog heißt ein leerer Pfad „niemand hat gewählt"; hier gibt es
// niemanden, der wählt — die Pfade kommen aus dem Fenster. Ein leerer wäre ein
// Aufruf, den so niemand gemeint hat.
func TestEinLeererPfadWirdAbgewiesen(t *testing.T) {
	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{}, "")

	if antwort := anwendung.BildAusDatei(t.Context(), "1", "kunst", ""); antwort.Fehler == nil {
		t.Error("der leere Pfad kam durch")
	}
}
