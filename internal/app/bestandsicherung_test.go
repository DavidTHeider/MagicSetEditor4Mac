// **Der Bestand lässt sich aus dem Fenster sichern** (#799).
package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// sicherungsprobe merkt sich, wohin gesichert wurde — und legt eine Datei an.
//
// **Sie schreibt wirklich**, seit die Sicherung ein Archiv ist
// (#1162):
// der Bestand kommt in eine Nachbardatei und von dort ins Archiv. Ein Double,
// das nichts anlegt, prüfte einen Weg, den es so nicht gibt (§3).
type sicherungsprobe struct {
	ziele  []string
	fehler error
}

func (s *sicherungsprobe) Sichern(_ context.Context, ziel string) error {
	s.ziele = append(s.ziele, ziel)

	if s.fehler != nil {
		return s.fehler
	}

	return os.WriteFile(ziel, []byte("SQLite format 3\x00 tut so"), 0o600)
}

// mitSicherung baut eine Anwendung mit Dateidialog und Sicherer.
func mitSicherung(
	t *testing.T, gewaehlt string, wahlfehler error, sicherer *sicherungsprobe,
) *app.App {
	t.Helper()

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return gewaehlt, wahlfehler
	})
	anwendung.MitBestandssicherer(sicherer)

	// **Ein eigenes Bilderverzeichnis**, damit der Lauf nicht das echte
	// dieses Rechners einliest (§3).
	bilder := t.TempDir()
	anwendung.MitBilderort(func() (string, error) { return bilder, nil })

	return anwendung
}

// **Am gewählten Ort steht danach eine Sicherung.**
//
// **Nicht mehr „der Pfad geht an den Sicherer"**: seit #1162 bekommt der
// Sicherer eine **Nachbardatei**, und was am gewählten Ort entsteht, ist das
// Archiv. Der Test prüft deshalb das Ergebnis und nicht den Zwischenschritt.
func TestBestandSichernSchreibtAnDenGewaehltenOrt(t *testing.T) {
	sicherer := &sicherungsprobe{}
	ziel := filepath.Join(t.TempDir(), "mse-sicherung"+sicherung.Endung)

	antwort := mitSicherung(t, ziel, nil, sicherer).BestandSichern(probekontext(t), "", false)
	if antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	if antwort.Pfad != ziel {
		t.Errorf("die Antwort nennt %q", antwort.Pfad)
	}

	if !sicherung.IstArchiv(ziel) {
		t.Fatal("am gewählten Ort steht kein Archiv")
	}

	beschreibung, err := sicherung.Entpacken(ziel, t.TempDir())
	if err != nil {
		t.Fatalf("entpacken: %v", err)
	}

	if len(beschreibung.Teile) != 1 || beschreibung.Teile[0].Name != app.Bestandsname {
		t.Errorf("die Sicherung trägt %+v", beschreibung.Teile)
	}

	// **Und sie sagt, welche Fassung sie geschrieben hat.**
	if beschreibung.Fassung != app.Fassung {
		t.Errorf("die Sicherung nennt die Fassung %q", beschreibung.Fassung)
	}
}

// **Und die eingesetzten Bilder liegen darin.**
//
// # Der Befund, für den es diesen Test gibt
//
// Bis 0.2.0 schrieb „Bestand sichern" genau eine `.db`. Die eingesetzten
// Kartenbilder liegen unter `beiwerk/`, und
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md) sagt
// selbst, ein Set sei ohne dieses Verzeichnis unvollständig.
//
// **Auf demselben Rechner fällt das nie auf** — das alte `beiwerk/` liegt ja
// noch da. Angestrichen vom Audit für 0.3.0 (Z6-01).
func TestDieSicherungTraegtDieEingesetztenBilder(t *testing.T) {
	sicherer := &sicherungsprobe{}
	ziel := filepath.Join(t.TempDir(), "mit-bildern"+sicherung.Endung)

	anwendung := mitSicherung(t, ziel, nil, sicherer)

	bilder := bilderortVon(t, anwendung)

	if err := os.MkdirAll(filepath.Join(bilder, "4"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "4", "kunst.png"),
		[]byte("so tut ein PNG"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	nach := t.TempDir()
	if _, err := sicherung.Entpacken(ziel, nach); err != nil {
		t.Fatalf("entpacken: %v", err)
	}

	roh, err := os.ReadFile(filepath.Join(nach, app.Beiwerkordner, "4", "kunst.png")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("das Bild fehlt in der Sicherung: %v", err)
	}

	if string(roh) != "so tut ein PNG" {
		t.Errorf("das Bild kam anders heraus: %q", roh)
	}
}

// bilderortVon gibt das Bilderverzeichnis einer Probeanwendung.
func bilderortVon(t *testing.T, anwendung *app.App) string {
	t.Helper()

	ort, err := anwendung.Bilderort()
	if err != nil {
		t.Fatalf("Bilderort: %v", err)
	}

	return ort
}

func TestEinAbbruchImDialogIstKeinFehler(t *testing.T) {
	sicherer := &sicherungsprobe{}

	antwort := mitSicherung(t, "", nil, sicherer).BestandSichern(probekontext(t), "", false)
	if antwort.Fehler != nil {
		t.Fatalf("Abbruch gab einen Fehler: %+v", antwort.Fehler)
	}

	if antwort.Pfad != "" {
		t.Errorf("Pfad %q, erwartet war keiner", antwort.Pfad)
	}

	if len(sicherer.ziele) != 0 {
		t.Errorf("trotz Abbruch gesichert: %v", sicherer.ziele)
	}
}

// **Ein Fehler des Dialogs wird gemeldet.**
func TestEinFehlerDesDialogsWirdGemeldet(t *testing.T) {
	sicherer := &sicherungsprobe{}

	antwort := mitSicherung(t, "", errors.New("kein Fenster"), sicherer).
		BestandSichern(probekontext(t), "", false)
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if len(sicherer.ziele) != 0 {
		t.Errorf("trotz Dialogfehler gesichert: %v", sicherer.ziele)
	}
}

// **Und einer des Sicherers ebenso — übersetzt, nicht roh** (§8).
func TestEinFehlerDesSicherersKommtUebersetztAn(t *testing.T) {
	sicherer := &sicherungsprobe{fehler: domain.ErrValidation}

	antwort := mitSicherung(t, "/tmp/x.db", nil, sicherer).BestandSichern(probekontext(t), "", false)
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Fehler.Code != "validierung" {
		t.Errorf("Code %q, erwartet war validierung", antwort.Fehler.Code)
	}

	if strings.Contains(antwort.Fehler.Meldung, "SQL") {
		t.Errorf("die rohe Meldung kam durch: %s", antwort.Fehler.Meldung)
	}
}

// **Die Bindung reicht durch** — mit dem Kontext des Fensters.
//
// Ohne diesen Test bliebe die Bindung ungemessen, und ein `b.app().SetSpeichern`
// statt `b.app().BestandSichern` fiele niemandem auf: beide geben eine Antwort
// mit einem Pfad.
func TestDieBindungSichertDenBestand(t *testing.T) {
	sicherer := &sicherungsprobe{}
	ziel := filepath.Join(t.TempDir(), "aus-der-bindung"+sicherung.Endung)

	anwendung := mitSicherung(t, ziel, nil, sicherer)

	if antwort := app.NeueBindungen(anwendung).BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	// **Am Ziel steht das Archiv**, und der Sicherer hat **eine**
	// Nachbardatei geschrieben — nicht das Ziel selbst.
	if !sicherung.IstArchiv(ziel) {
		t.Error("die Bindung legte kein Archiv ab")
	}

	if len(sicherer.ziele) != 1 || filepath.Base(sicherer.ziele[0]) != app.Bestandsname {
		t.Errorf("gesichert nach %v", sicherer.ziele)
	}
}
