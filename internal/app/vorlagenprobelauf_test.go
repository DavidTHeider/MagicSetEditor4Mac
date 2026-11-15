// Der Probelauf einer Exportvorlage — UXR-F7
// (#2015).
//
// **Ein Weg, zwei Ausgänge**: der Probelauf wertet die Vorlage über denselben
// Weg aus wie der echte Lauf, und was dort geschrieben würde, trägt sich hier
// in eine Ablage ein. Die Tests fahren beide mit derselben Eingabe und
// vergleichen — Fehler wie Schreibwirkung.
package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitVorlagenUndKarten ist `mitVorlagen` mit zwei Karten im ersten Set.
func mitVorlagenUndKarten(t *testing.T, probe *vorlagenprobe) *app.App {
	t.Helper()

	anwendung := mitVorlagen(t, probe)

	for range 2 {
		if angelegt := anwendung.KarteAnlegen(t.Context(), "1"); angelegt.Fehler != nil {
			t.Fatalf("Karte anlegen: %+v", angelegt.Fehler)
		}
	}

	return anwendung
}

// **Der Probelauf nennt Karten, Größe und Namen — und legt nichts an.**
func TestDerVorlagenprobelaufNenntKartenGroesseUndName(t *testing.T) {
	t.Parallel()

	anwendung := mitVorlagenUndKarten(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "fertig",
	})

	antwort := anwendung.VorlageProbelauf(t.Context(), "1", "probe.mse-export-template", "", nil, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Karten != 2 || antwort.Bytes != 6 || antwort.Dateiname != "ausfuhr.txt" ||
		antwort.Beigaben != 0 || antwort.Verzeichnis != "" || antwort.Geschaetzt {
		t.Errorf("%+v", antwort)
	}

	if antwort.Befunde == nil {
		t.Error("Befunde ist null — das Frontend läuft darüber, ohne zu prüfen")
	}
}

// **Eine Auswahl zählt: nur die gewählten Karten laufen.**
func TestDerVorlagenprobelaufZaehltNurDieGewaehlten(t *testing.T) {
	t.Parallel()

	anwendung := mitVorlagenUndKarten(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "x",
	})

	zeilen := anwendung.Kartenzeilen(t.Context(), "1")
	if zeilen.Fehler != nil || len(zeilen.Zeilen) != 2 {
		t.Fatalf("Zeilen: %+v", zeilen)
	}

	antwort := anwendung.VorlageProbelauf(t.Context(), "1", "probe.mse-export-template", "",
		nil, []string{zeilen.Zeilen[1].ID})
	if antwort.Fehler != nil || antwort.Karten != 1 {
		t.Errorf("%+v", antwort)
	}
}

// **Was die Vorlage neben ihre Ausgabe schriebe, wird gezählt und nicht
// geschrieben.**
//
// Die Vorlage legt hier einen Text, ein Set und ein Kartenbild ab, wie
// `magic-spoiler` es tut. Im Probelauf liegt danach **nichts** im Ordner, auch
// kein `-files`-Verzeichnis; die Antwort zählt drei Dateien, kennt die Größe
// des Textes und sagt, dass die Summe geschätzt ist.
func TestDerVorlagenprobelaufSchreibtNichts(t *testing.T) {
	t.Parallel()

	vorlage := probevorlage()
	vorlage.Verzeichnis = true

	probe := &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{vorlage},
		ausgabe:  "fertig",
	}

	probe.waehrend = func(mittel service.Ausfuhrmittel) {
		if _, err := mittel.Beigaben.Lege("liste.txt", []byte("abc")); err != nil {
			t.Errorf("Lege: %v", err)
		}

		if _, err := mittel.Beigaben.LegeSet(t.Context(), "set.mse-set"); err != nil {
			t.Errorf("LegeSet: %v", err)
		}

		if _, err := mittel.Bilder.Schreibe(t.Context(), 1, mittel.Bildziel,
			"karte.png", service.FormatPNG, 300); err != nil {
			t.Errorf("Schreibe: %v", err)
		}
	}

	anwendung := mitVorlagenUndKarten(t, probe)
	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "meine.txt")

	antwort := anwendung.VorlageProbelauf(t.Context(), "1", "probe.mse-export-template",
		ziel, nil, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Beigaben != 3 || antwort.BeigabenBytes != 3 || !antwort.Geschaetzt ||
		antwort.Verzeichnis != "meine-files" || antwort.Dateiname != "meine.txt" {
		t.Errorf("%+v", antwort)
	}

	if eintraege, err := os.ReadDir(ordner); err != nil || len(eintraege) != 0 {
		t.Errorf("im Ordner liegt etwas: %v (%v)", eintraege, err)
	}
}

// **Dieselben Namen werden abgewiesen wie im echten Lauf.**
//
// Ein Name mit Pfad im `copy_file`, ein Bildziel ohne Format: beide Wege
// bekommen von denselben Schreibern denselben Satz.
func TestDerVorlagenprobelaufWeistDieNamenDesLaufsAb(t *testing.T) {
	t.Parallel()

	aufrufe := map[string]func(mittel service.Ausfuhrmittel) error{
		"Beigabe mit Pfad": func(mittel service.Ausfuhrmittel) error {
			_, err := mittel.Beigaben.Lege("../ausbruch.txt", nil)

			return err
		},
		"Set mit Pfad": func(mittel service.Ausfuhrmittel) error {
			_, err := mittel.Beigaben.LegeSet(context.Background(), "a/b.mse-set")

			return err
		},
		"Bild ohne Format": func(mittel service.Ausfuhrmittel) error {
			_, err := mittel.Bilder.Schreibe(context.Background(), 1, mittel.Bildziel,
				"karte.gif", "gif", 300)

			return err
		},
		"Bild mit Pfad": func(mittel service.Ausfuhrmittel) error {
			_, err := mittel.Bilder.Schreibe(context.Background(), 1, mittel.Bildziel,
				"../karte.png", service.FormatPNG, 300)

			return err
		},
		"Bild über der Breitengrenze": func(mittel service.Ausfuhrmittel) error {
			_, err := mittel.Bilder.Schreibe(context.Background(), 1, mittel.Bildziel,
				"karte.png", service.FormatPNG, service.GrenzeZielbreite+1)

			return err
		},
	}

	for name, aufruf := range aufrufe {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var lauf, probelauf error

			probe := &vorlagenprobe{vorlagen: []msepaket.Exportvorlage{probevorlage()}}
			probe.waehrend = func(mittel service.Ausfuhrmittel) {
				if lauf == nil && probelauf == nil {
					lauf = aufruf(mittel)

					return
				}

				probelauf = aufruf(mittel)
			}

			anwendung := mitVorlagenUndKarten(t, probe)

			// Bildausführer: der echte Lauf braucht einen, der Probelauf nicht.
			anwendung.MitBildausfuhr(&bildausfuehrerDouble{})

			echt := anwendung.VorlageAusfuehren(t.Context(), "1", "probe.mse-export-template",
				filepath.Join(t.TempDir(), "a.txt"), nil, nil)
			if echt.Fehler != nil {
				t.Fatalf("Lauf: %+v", echt.Fehler)
			}

			if proben := anwendung.VorlageProbelauf(t.Context(), "1", "probe.mse-export-template",
				filepath.Join(t.TempDir(), "b.txt"), nil, nil); proben.Fehler != nil {
				t.Fatalf("Probelauf: %+v", proben.Fehler)
			}

			if lauf == nil || probelauf == nil {
				t.Fatalf("Lauf %v, Probelauf %v", lauf, probelauf)
			}

			if lauf.Error() != probelauf.Error() {
				t.Errorf("Lauf %q, Probelauf %q", lauf, probelauf)
			}
		})
	}
}

// bildausfuehrerDouble lässt `Schreibe` an der Prüfung des Ziels scheitern,
// wie der Dienst es tut, und schreibt sonst nichts.
type bildausfuehrerDouble struct{}

func (*bildausfuehrerDouble) Namen(
	context.Context, int64, string, service.Bildformat, string, domain.Kartenauswahl,
) ([]service.Bildname, error) {
	return nil, nil
}

func (*bildausfuehrerDouble) Schreibe(
	_ context.Context, _ int64, _, datei string, format service.Bildformat, zielbreite int,
	_ string, _ render.Guete,
) (*service.Bildschreibung, error) {
	if err := service.PruefeBildziel(datei, format, zielbreite); err != nil {
		return nil, err
	}

	return &service.Bildschreibung{}, nil
}

// **Dieselben Fehler wie der echte Lauf — über denselben Weg.**
func TestDerVorlagenprobelaufGibtDieFehlerDesLaufs(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		setID, paket string
		auswahl      []string
		probe        *vorlagenprobe
	}{
		"eine Kennung, die keine ist": {
			setID: "keine Zahl", paket: "probe.mse-export-template",
			probe: &vorlagenprobe{},
		},
		"eine Vorlage, die es nicht gibt": {
			setID: "1", paket: "gibtesnicht",
			probe: &vorlagenprobe{vorlagen: []msepaket.Exportvorlage{probevorlage()}},
		},
		"ein Lesefehler beim Suchen": {
			setID: "1", paket: "p", probe: &vorlagenprobe{fehler: errors.New("kaputt")},
		},
		"eine Auswahl, die keine ist": {
			setID: "1", paket: "probe.mse-export-template", auswahl: []string{"x"},
			probe: &vorlagenprobe{vorlagen: []msepaket.Exportvorlage{probevorlage()}},
		},
		"eine Auswahl, die keine Karte trifft": {
			setID: "1", paket: "probe.mse-export-template", auswahl: []string{"99999"},
			probe: &vorlagenprobe{vorlagen: []msepaket.Exportvorlage{probevorlage()}},
		},
		"ein Set, das es nicht gibt": {
			setID: "99", paket: "probe.mse-export-template",
			probe: &vorlagenprobe{vorlagen: []msepaket.Exportvorlage{probevorlage()}},
		},
		"eine Vorlage, die scheitert": {
			setID: "1", paket: "probe.mse-export-template",
			probe: &vorlagenprobe{
				vorlagen:   []msepaket.Exportvorlage{probevorlage()},
				lauffehler: errors.New(`der Name "write_image_file" ist hier nicht bekannt`),
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := mitVorlagenUndKarten(t, fall.probe)

			echt := anwendung.VorlageAusfuehren(t.Context(), fall.setID, fall.paket,
				filepath.Join(t.TempDir(), "a.txt"), nil, fall.auswahl)
			probe := anwendung.VorlageProbelauf(t.Context(), fall.setID, fall.paket, "",
				nil, fall.auswahl)

			if echt.Fehler == nil || probe.Fehler == nil {
				t.Fatalf("Lauf %+v, Probelauf %+v", echt.Fehler, probe.Fehler)
			}

			if !reflect.DeepEqual(ohneVorgang(echt.Fehler), ohneVorgang(probe.Fehler)) {
				t.Errorf("Lauf %+v, Probelauf %+v", echt.Fehler, probe.Fehler)
			}

			if !reflect.DeepEqual(echt.Befunde, probe.Befunde) {
				t.Errorf("Befunde: Lauf %v, Probelauf %v", echt.Befunde, probe.Befunde)
			}
		})
	}
}

// **Ein Lesefehler der Karten ist derselbe Satz in beiden Wegen.**
func TestDerVorlagenprobelaufMeldetDieMeldungenDerAuswertung(t *testing.T) {
	t.Parallel()

	anwendung := mitVorlagenUndKarten(t, &vorlagenprobe{
		vorlagen:    []msepaket.Exportvorlage{probevorlage()},
		ausgabe:     "fertig",
		laufbefunde: []domain.Befund{{Ort: "mit namen", Meldung: "die Vorlage führt keine Wahlmöglichkeit dieses Namens"}},
	})

	antwort := anwendung.VorlageProbelauf(t.Context(), "1", "probe.mse-export-template", "", nil, nil)
	if antwort.Fehler != nil || len(antwort.Befunde) != 1 {
		t.Errorf("%+v", antwort)
	}
}

// **Eine leere Ausgabe wird gesagt, wie im echten Lauf.**
func TestEineLeereAusgabeWirdImProbelaufGesagt(t *testing.T) {
	t.Parallel()

	anwendung := mitVorlagenUndKarten(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})

	echt := anwendung.VorlageAusfuehren(t.Context(), "1", "probe.mse-export-template",
		filepath.Join(t.TempDir(), "a.txt"), nil, nil)
	probe := anwendung.VorlageProbelauf(t.Context(), "1", "probe.mse-export-template", "", nil, nil)

	if probe.Bytes != 0 || !reflect.DeepEqual(echt.Befunde, probe.Befunde) || len(probe.Befunde) != 1 {
		t.Errorf("Lauf %v, Probelauf %+v", echt.Befunde, probe)
	}
}

// **Die Bindung fragt nach keinem Ort und nennt den Vorschlag des Dialogs.**
func TestDieBindungZumVorlagenprobelaufFragtNachKeinemOrt(t *testing.T) {
	t.Parallel()

	anwendung := mitVorlagenUndKarten(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "fertig",
	})

	gefragt := false

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		gefragt = true

		return "", nil
	})

	antwort := app.NeueBindungen(anwendung).VorlageProbelauf(t.Context(),
		"1", "probe.mse-export-template", map[string]string{"mit namen": "no"}, nil)
	if antwort.Fehler != nil || antwort.Dateiname != "ausfuhr.txt" || antwort.Karten != 2 {
		t.Fatalf("%+v", antwort)
	}

	if gefragt {
		t.Error("der Probelauf hat nach einem Ort gefragt")
	}
}
