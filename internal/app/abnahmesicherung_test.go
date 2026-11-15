// **Der Abnahmelauf für die vollständige Sicherung** (#1162).
//
// # Warum er neben den anderen steht
//
// Jeder Lauf in `vollrueckspielen_test.go` setzt das Bilderverzeichnis über
// [app.App.MitBilderort] — und prüft damit den Weg, nicht den **Ort**. Wäre
// `beiwerk/` in der Anwendung woanders, als die Sicherung sucht, blieben alle
// grün, und die Sicherung käme auf einem zweiten Rechner ohne Bilder an.
//
// Genau das ist der Fall, den #1162
// beschreibt: **auf demselben Rechner fällt es nie auf**, weil das alte
// Verzeichnis noch daliegt.
//
// Dieser Lauf nimmt deshalb die **echten** Orte — [app.Datenbankpfad] und
// [app.Bilderverzeichnis] — und stellt den Rechnerwechsel her, indem er das
// ganze Anwendungsverzeichnis löscht.
//
// **Was er nicht zeigt**: das Fenster. Der Dateidialog ist eine Naht, und wer
// im Menü „Bestand sichern" wählt, geht durch dieselbe Bindung — dass sie im
// Menü hängt, prüft `vollstaendigkeit_intern_test.go`.
package app_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein gelöschtes Anwendungsverzeichnis lässt sich aus der Sicherung allein
// wiederherstellen — mit Bild.**
func TestDerAbnahmelaufStelltAllesAusDerSicherungHer(t *testing.T) {
	// **Ein eigenes Zuhause.** Ohne das liefe dieser Lauf gegen den Bestand
	// dieses Rechners — und löschte ihn (§3).
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	bestand, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Datenbankpfad: %v", err)
	}

	if !strings.HasPrefix(bestand, zuhause) {
		t.Fatalf("der Bestand liegt außerhalb des Testzuhauses: %q", bestand)
	}

	anwendungsordner := filepath.Dir(bestand)
	if fehler := os.MkdirAll(anwendungsordner, 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	anwendung := app.NeueApp(bestand)
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	// **Ein Bild wird eingesetzt, wie die Anwendung es tut** — über
	// [service.Bildablage], nicht mit `os.WriteFile` an einen geratenen Ort.
	// Ein Lauf, der das Bild selbst hinlegt, prüfte seine eigene Annahme.
	bilder, err := app.Bilderverzeichnis()
	if err != nil {
		t.Fatalf("Bilderverzeichnis: %v", err)
	}

	ablage := service.NeueBildablage(bilder)

	name, err := ablage.Ablegen(42, "kunst", einfarbig())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	// **Den Ort nennt die Ablage**, nicht dieser Lauf: `Ablegen` gibt den
	// Namen, unter dem das Bild im Set steht, und `Verzeichnis` sagt, wo er
	// gilt. Wer den Pfad hier zusammensetzte, prüfte seine eigene Annahme.
	abgelegt := filepath.Join(ablage.Verzeichnis(42), name)

	vorher, err := os.ReadFile(abgelegt) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("das eingesetzte Bild lesen: %v", err)
	}

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	// **Und ein Exemplar der Sammlung** (#1165).
	//
	// Das Issue verlangt ausdrücklich, dass die Sicherung sie mitnimmt — und
	// das ist der Grund, warum sie **keine eigene Datei** bekommen hat
	// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)):
	// sie liegt im Bestand, und der wird gesichert. Ohne diese Zeilen wäre das
	// eine Behauptung über eine Datei, die niemand aufgemacht hat.
	if antwort := anwendung.ExemplarAnlegen(t.Context(), app.ExemplarDTO{
		Printing: "sicherung-druck-0001", Menge: 3,
	}); antwort.Fehler != nil {
		t.Fatalf("ExemplarAnlegen: %+v", antwort.Fehler)
	}

	// **Die Sicherung liegt außerhalb** — sie ist das Einzige, was den
	// Rechnerwechsel übersteht.
	sicherungsdatei := filepath.Join(t.TempDir(), "abnahme.msesicherung")

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return sicherungsdatei, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	// **Der Rechnerwechsel**: das ganze Anwendungsverzeichnis ist fort.
	if fehler := os.RemoveAll(anwendungsordner); fehler != nil {
		t.Fatalf("den Anwendungsordner löschen: %v", fehler)
	}

	if fehler := os.MkdirAll(anwendungsordner, 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Eine frische Anwendung**, wie sie auf dem zweiten Rechner startet.
	frisch := app.NeueApp(bestand)
	if fehler := frisch.Starten(probekontext(t)); fehler != nil {
		t.Fatalf("frisch starten: %v", fehler)
	}

	t.Cleanup(func() { _ = frisch.Beenden() }) //nolint:errcheck // Aufräumen

	antwort := frisch.SicherungZurueckspielen(probekontext(t), sicherungsdatei, "")
	if antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	// **Die Antwort nennt die neue Epoche** (J05, #2269): das Fenster
	// verwirft danach eine Kennung, die aus dem alten Bestand gemeldet wurde.
	if antwort.Bestandsepoche == 0 {
		t.Error("die Antwort nennt keine Epoche des Bestands")
	}

	// **Das Bild steht wieder an seinem Ort** — an dem, den die Ablage nennt,
	// und nicht an einem, den dieser Lauf sich ausgedacht hat.
	nachher, err := os.ReadFile(abgelegt) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("das eingesetzte Bild fehlt nach dem Zurückspielen: %v", err)
	}

	if !bytes.Equal(nachher, vorher) {
		t.Error("das Bild kam anders zurück")
	}

	// **Und der Bestand dazu.**
	if got := merkmal(t, frisch); got != "de" {
		t.Errorf("die Einstellung kam als %q zurück", got)
	}

	// **Samt der Sammlung** (#1165).
	sammlung := frisch.Sammlung(probekontext(t))
	if sammlung.Fehler != nil {
		t.Fatalf("Sammlung: %+v", sammlung.Fehler)
	}

	if len(sammlung.Exemplare) != 1 ||
		sammlung.Exemplare[0].Printing != "sicherung-druck-0001" {
		t.Errorf("die Sammlung kam als %+v zurück", sammlung.Exemplare)
	}
}

// einfarbig gibt ein kleines Bild — der Inhalt ist gleichgültig, die Ablage
// nicht.
func einfarbig() image.Image {
	flaeche := image.NewRGBA(image.Rect(0, 0, 4, 4))
	draw.Draw(flaeche, flaeche.Bounds(), &image.Uniform{C: color.RGBA{R: 200, A: 255}},
		image.Point{}, draw.Src)

	return flaeche
}
