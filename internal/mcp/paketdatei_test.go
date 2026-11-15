// **Die drei Werkzeuge der Paketwerkstatt** —
// #1164.
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// werkstattprobe antwortet fest — oder scheitert.
type werkstattprobe struct {
	dateien []string
	beiwerk []string
	inhalt  string
	stand   string
	fehler  error

	// gesehen hält fest, was ankam.
	paket, datei, wert, erwartet string

	// neu ist der neue Name beim Umbenennen (FMT-02).
	neu string

	// befund ist, was die Prüfung melden soll — Zeile 0 heißt: nichts.
	befundzeile, befundspalte int
	befundmeldung             string
}

func (w *werkstattprobe) Pruefen(inhalt string) (zeile, spalte int, meldung string) {
	w.wert = inhalt

	return w.befundzeile, w.befundspalte, w.befundmeldung
}

func (w *werkstattprobe) DateienUndBeiwerk(_ context.Context, paket string) (dateien, beiwerk []string, fehler error) {
	w.paket = paket

	return w.dateien, w.beiwerk, w.fehler
}

func (w *werkstattprobe) Anlegen(_ context.Context, paket, datei, inhalt string) (string, error) {
	w.paket, w.datei, w.wert = paket, datei, inhalt

	return w.stand, w.fehler
}

func (w *werkstattprobe) Umbenennen(_ context.Context, paket, alt, neu, erwartet string) error {
	w.paket, w.datei, w.neu, w.erwartet = paket, alt, neu, erwartet

	return w.fehler
}

func (w *werkstattprobe) Entfernen(_ context.Context, paket, datei, erwartet string) error {
	w.paket, w.datei, w.erwartet = paket, datei, erwartet

	return w.fehler
}

func (w *werkstattprobe) Lesen(_ context.Context, paket, datei string) (
	inhalt, stand string, fehler error,
) {
	w.paket, w.datei = paket, datei

	return w.inhalt, w.stand, w.fehler
}

func (w *werkstattprobe) Schreiben(
	_ context.Context, paket, datei, inhalt, erwartet string,
) (string, error) {
	w.paket, w.datei, w.wert, w.erwartet = paket, datei, inhalt, erwartet

	return w.stand, w.fehler
}

// rufeWerkstatt sucht ein Werkzeug der Werkstatt und ruft es.
func rufeWerkstatt(
	t *testing.T, werkstatt mcp.Paketwerkstatt, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     &konsolenprobe{},
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
		Werkstatt:   werkstatt,
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("es gibt kein Werkzeug %s", name)

	return "", nil
}

// **Die drei Werkzeuge reichen durch, was sie bekommen.**
func TestDieWerkstattwerkzeugeReichenDurch(t *testing.T) {
	t.Parallel()

	t.Run("aufzählen", func(t *testing.T) {
		t.Parallel()

		probe := &werkstattprobe{dateien: []string{"style", "card_fields"}}

		antwort, fehler := rufeWerkstatt(t, probe, "lesen.paketdateien",
			`{"paket": "probe.mse-style"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if probe.paket != "probe.mse-style" {
			t.Errorf("weitergereicht: %q", probe.paket)
		}

		if !strings.Contains(antwort, "card_fields") {
			t.Errorf("Antwort %s", antwort)
		}
	})

	t.Run("lesen", func(t *testing.T) {
		t.Parallel()

		probe := &werkstattprobe{inhalt: "mse version: 2.0.0\n", stand: "abc"}

		antwort, fehler := rufeWerkstatt(t, probe, "lesen.paketdatei",
			`{"paket": "probe.mse-style", "datei": "style"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if probe.datei != "style" {
			t.Errorf("weitergereicht: %q", probe.datei)
		}

		// **Der Stand steht in der Antwort.** Ohne ihn hätte ein Modell nichts,
		// womit es schreiben könnte.
		if !strings.Contains(antwort, "abc") {
			t.Errorf("Antwort %s", antwort)
		}
	})

	// **Die Prüfung, mit Befund und ohne** — Werkstück B von Phase 4.9
	// (#1324).
	//
	// Beide Zweige, weil `Gefunden: false` die andere Aussage ist: ein
	// Modell, das nur den Fundfall bekäme, wüsste nach einer heilen Datei
	// nicht, ob geprüft wurde.
	t.Run("prüfen mit Befund", func(t *testing.T) {
		t.Parallel()

		probe := &werkstattprobe{befundzeile: 3, befundspalte: 7, befundmeldung: "hier fehlt ein Wert"}

		antwort, fehler := rufeWerkstatt(t, probe, "lesen.skriptbefund",
			`{"inhalt": "script: f(1, 2"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if probe.wert != "script: f(1, 2" {
			t.Errorf("weitergereicht: %q", probe.wert)
		}

		for _, teil := range []string{"true", "3", "7", "hier fehlt ein Wert"} {
			if !strings.Contains(antwort, teil) {
				t.Errorf("Antwort %s nennt %q nicht", antwort, teil)
			}
		}
	})

	t.Run("prüfen ohne Befund", func(t *testing.T) {
		t.Parallel()

		antwort, fehler := rufeWerkstatt(t, &werkstattprobe{}, "lesen.skriptbefund",
			`{"inhalt": "script: 1 + 2"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if !strings.Contains(antwort, "false") {
			t.Errorf("Antwort %s", antwort)
		}
	})

	t.Run("schreiben", func(t *testing.T) {
		t.Parallel()

		probe := &werkstattprobe{stand: "neu"}

		antwort, fehler := rufeWerkstatt(t, probe, "schreiben.paketdatei",
			`{"paket": "probe.mse-style", "datei": "style", "inhalt": "x\n", "stand": "alt"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if probe.wert != "x\n" || probe.erwartet != "alt" {
			t.Errorf("weitergereicht: %+v", probe)
		}

		if !strings.Contains(antwort, "neu") {
			t.Errorf("Antwort %s", antwort)
		}
	})
}

// **Eine leere Liste und kein `null`** — dieselbe Zusage wie am Fenster.
func TestEinPaketOhneVorlagenGibtEineLeereListeUeberMCP(t *testing.T) {
	t.Parallel()

	antwort, fehler := rufeWerkstatt(t, &werkstattprobe{dateien: nil},
		"lesen.paketdateien", `{"paket": "probe.mse-style"}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("Antwort %s", antwort)
	}
}

// **Die Fehlerwege aller drei.**
//
// **Jeder Fall bekommt seine eigene Probe.** Eine geteilte schriebe aus
// mehreren parallelen Läufen in dieselben Felder — der Rennmelder hat genau
// das gefunden, bevor dieser Kommentar hier stand.
func TestDieFehlerwegeDerWerkstattwerkzeuge(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		werkzeug  string
		scheitert bool
		argumente string
	}{
		"aufzählen: kein JSON": {"lesen.paketdateien", false, `{`},
		"aufzählen: gibt auf":  {"lesen.paketdateien", true, `{"paket": "p"}`},
		"lesen: kein JSON":     {"lesen.paketdatei", false, `{`},
		"prüfen: kein JSON":    {"lesen.skriptbefund", false, `{`},
		"lesen: gibt auf":      {"lesen.paketdatei", true, `{"paket": "p", "datei": "d"}`},
		"schreiben: kein JSON": {"schreiben.paketdatei", false, `{`},
		"schreiben: gibt auf":  {"schreiben.paketdatei", true, `{"paket": "p", "datei": "d"}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &werkstattprobe{}
			if fall.scheitert {
				probe.fehler = errors.New("keine Platte")
			}

			if _, fehler := rufeWerkstatt(t, probe,
				fall.werkzeug, fall.argumente); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
