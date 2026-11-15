package app_test

// **Der Abzug eines Tisches an der Naht** — PLAY-05.AK-03 und AK-04
// (#1713).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abzugsprobe merkt sich, was ankam.
type abzugsprobe struct {
	daten  string
	pfad   string
	blatt  string
	fehler error
}

func (a *abzugsprobe) Schreibe(datenURL, pfad string) error {
	a.daten, a.pfad = datenURL, pfad

	return a.fehler
}

func (a *abzugsprobe) SchreibePDF(datenURL, pfad, blatt string) error {
	a.daten, a.pfad, a.blatt = datenURL, pfad, blatt

	return a.fehler
}

// mitAbzug baut eine Anwendung mit Dateidialog und Schreiber.
func mitAbzug(gewaehlt string, wahlfehler error, probe *abzugsprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return gewaehlt, wahlfehler
	})
	anwendung.MitTischabzugschreiber(probe)

	return anwendung
}

// **Der gewählte Ort geht an den Schreiber, und der Abzug mit ihm.**
func TestTischabzugSpeichernSchreibtAnDenGewaehltenOrt(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{}
	antwort := app.NeueBindungen(mitAbzug("/tmp/tisch.png", nil, probe)).
		TischabzugSpeichern(t.Context(), "data:image/png;base64,AAA")

	if antwort.Fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", antwort.Fehler)
	}

	if antwort.Pfad != "/tmp/tisch.png" || probe.pfad != "/tmp/tisch.png" {
		t.Errorf("Pfad %q, beim Schreiber %q", antwort.Pfad, probe.pfad)
	}

	if probe.daten != "data:image/png;base64,AAA" {
		t.Errorf("beim Schreiber kam %q an", probe.daten)
	}
}

// **Ein weggeklickter Dialog ist ein Abbruch und kein Fehler** — dieselbe
// Unterscheidung wie bei `SetSpeichernUnter`.
func TestTischabzugSpeichernAbbruch(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{}
	antwort := app.NeueBindungen(mitAbzug("", nil, probe)).
		TischabzugSpeichern(t.Context(), "data:image/png;base64,AAA")

	if !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Fatalf("erwartet Abbruch ohne Fehler, bekam %+v", antwort)
	}

	if probe.pfad != "" {
		t.Errorf("geschrieben wurde trotzdem, nach %q", probe.pfad)
	}
}

// **Ein Dialog, der scheitert, meldet sich** — und schreibt nichts.
func TestTischabzugSpeichernDialogfehler(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{}
	antwort := app.NeueBindungen(
		mitAbzug("", errors.New("kein Fenster"), probe),
	).TischabzugSpeichern(t.Context(), "data:image/png;base64,AAA")

	if antwort.Fehler == nil {
		t.Fatal("erwartet einen Fehler")
	}

	if probe.pfad != "" {
		t.Errorf("geschrieben wurde trotzdem, nach %q", probe.pfad)
	}
}

// **Und ein Abzug, den der Dienst abweist, kommt als Meldung zurück.**
func TestTischabzugSchreibenMeldetDenDienstfehler(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{fehler: domain.ErrValidation}
	antwort := mitAbzug("/tmp/tisch.png", nil, probe).
		TischabzugSchreiben(context.Background(), "unsinn", "/tmp/tisch.png")

	if antwort.Fehler == nil {
		t.Fatal("erwartet einen Fehler")
	}

	if antwort.Pfad != "" {
		t.Errorf("Pfad %q, erwartet keinen", antwort.Pfad)
	}
}

// **Der Weg als PDF trägt das Blatt mit** — PLAY-05.AK-03
// (#1713).
func TestTischabzugPDFSpeichernReichtDasBlattWeiter(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{}
	antwort := app.NeueBindungen(mitAbzug("/tmp/tisch.pdf", nil, probe)).
		TischabzugPDFSpeichern(t.Context(), "data:image/png;base64,AAA", "letter")

	if antwort.Fehler != nil {
		t.Fatalf("unerwarteter Fehler: %v", antwort.Fehler)
	}

	if probe.blatt != "letter" || probe.pfad != "/tmp/tisch.pdf" {
		t.Errorf("beim Schreiber kam %q auf %q an", probe.blatt, probe.pfad)
	}
}

// **Und ein weggeklickter Dialog schreibt nichts** — wie beim PNG daneben.
func TestTischabzugPDFSpeichernAbbruch(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{}
	antwort := app.NeueBindungen(mitAbzug("", nil, probe)).
		TischabzugPDFSpeichern(t.Context(), "data:image/png;base64,AAA", "a4")

	if !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Fatalf("erwartet Abbruch ohne Fehler, bekam %+v", antwort)
	}

	if probe.pfad != "" {
		t.Errorf("geschrieben wurde trotzdem, nach %q", probe.pfad)
	}
}

// **Ein Dialog, der scheitert, meldet sich** — und schreibt nichts.
func TestTischabzugPDFSpeichernDialogfehler(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{}
	antwort := app.NeueBindungen(
		mitAbzug("", errors.New("kein Fenster"), probe),
	).TischabzugPDFSpeichern(t.Context(), "data:image/png;base64,AAA", "a4")

	if antwort.Fehler == nil {
		t.Fatal("erwartet einen Fehler")
	}

	if probe.pfad != "" {
		t.Errorf("geschrieben wurde trotzdem, nach %q", probe.pfad)
	}
}

// **Und ein Abzug, den der Dienst abweist, kommt als Meldung zurück.**
func TestTischabzugPDFSchreibenMeldetDenDienstfehler(t *testing.T) {
	t.Parallel()

	probe := &abzugsprobe{fehler: domain.ErrValidation}
	antwort := mitAbzug("/tmp/tisch.pdf", nil, probe).TischabzugPDFSchreiben(
		context.Background(), "unsinn", "/tmp/tisch.pdf", "a4")

	if antwort.Fehler == nil {
		t.Fatal("erwartet einen Fehler")
	}

	if antwort.Pfad != "" {
		t.Errorf("Pfad %q, erwartet keinen", antwort.Pfad)
	}
}
