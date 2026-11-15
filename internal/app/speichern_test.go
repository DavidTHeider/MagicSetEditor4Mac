package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// speicherprobe ist ein Double des Speicherers.
type speicherprobe struct {
	speicherung *service.Speicherung
	fehler      error
	setID       int64
	ziel        string
}

func (s *speicherprobe) Schreibe(
	_ context.Context, setID int64, ziel string,
) (*service.Speicherung, error) {
	s.setID, s.ziel = setID, ziel

	return s.speicherung, s.fehler
}

func mitSpeicherer(probe *speicherprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitSpeicherer(probe)

	return anwendung
}

// **Was nicht mitkam, kommt in der Antwort an.**
//
// Ein Speichern, das die Bilder verliert und nichts sagt, sieht aus wie eines,
// das geglückt ist — bis jemand die Datei öffnet (§8).
func TestSetSpeichernNenntDieBefunde(t *testing.T) {
	probe := &speicherprobe{speicherung: &service.Speicherung{
		Pfad:    "/tmp/wieder.mse-set",
		Befunde: []domain.Befund{{Datei: "probe.mse-set", Ort: "image1", Meldung: "nicht lesbar"}},
	}}

	antwort := mitSpeicherer(probe).SetSpeichern(probekontext(t), "7", "/tmp/wieder.mse-set")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1", len(antwort.Befunde))
	}

	if antwort.Befunde[0] == "" {
		t.Error("ein leerer Befund sagt nichts")
	}
}

// Eine Kennung, die keine Zahl ist, wird übersetzt und nicht geraten.
func TestSetSpeichernMitUnsinnigerKennung(t *testing.T) {
	antwort := mitSpeicherer(&speicherprobe{}).SetSpeichern(probekontext(t), "sieben", "")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler — welches Set wäre geschrieben worden?")
	}
}

// Ein Fehler des Dienstes wird übersetzt, nicht durchgereicht.
func TestSetSpeichernUebersetztDenFehler(t *testing.T) {
	probe := &speicherprobe{fehler: errors.New("das Ziel anlegen: kein Platz")}

	antwort := mitSpeicherer(probe).SetSpeichern(probekontext(t), "7", "/tmp/wieder.mse-set")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Befunde == nil {
		t.Error("Befunde ist null — das Frontend läuft darüber, ohne zu prüfen")
	}
}

// **Ein gescheiterter Arbeitsortwechsel nimmt den Pfad nicht mit** — Befund 1
// des Nachlaufs über Phase 4.4
// (#1291).
//
// `Setschreiber.Schreibe` gibt die `Speicherung` **und** den Fehler zurück,
// wenn nur `HerkunftSetzen` scheitert: die Datei liegt dann schon am Ziel
// (§7 Regel 3). Wer nur den Fehler zeigte, ließe den Aufrufer im Unklaren
// darüber, ob geschrieben wurde.
func TestEinGescheiterterArbeitsortwechselNenntDenPfad(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitSpeicherer(halberSpeicherer{})

	antwort := anwendung.SetSpeichern(t.Context(), "7", "/tmp/halb.mse-set")

	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if antwort.Pfad != "/tmp/halb.mse-set" {
		t.Errorf("der Pfad ging verloren: %q", antwort.Pfad)
	}
}

// halberSpeicherer schreibt die Datei und scheitert danach.
type halberSpeicherer struct{}

func (halberSpeicherer) Schreibe(_ context.Context, _ int64, ziel string) (
	*service.Speicherung, error,
) {
	return &service.Speicherung{Pfad: ziel, Befunde: nil},
		errors.New("den Arbeitsort setzen: die Datenbank ist zu")
}
