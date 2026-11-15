// **Ein Doppelklick im Finder öffnet das Set** —
// #891,
// Werkstück 1 von Phase Y.
//
// # Warum es diese Naht gibt
//
// macOS übergibt eine doppelgeklickte Datei **nicht** in `os.Args`, sondern
// als Apple-Event `odoc` an die laufende Anwendung. Wails reicht es über
// `mac.Options.OnFileOpen` weiter
// (`internal/frontend/desktop/darwin/AppDelegate.m`, Fassung v2.13.0). Wer
// nur `os.Args` läse, bekäme vom Finder nie etwas.
//
// **Und `os.Args` gilt trotzdem**: aus einer Shell gestartet steht der Pfad
// dort, und das ist der Weg, den ein Skript nimmt.
package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine übergebene Datei wird geöffnet und gemeldet.**
func TestEineUebergebeneDateiWirdGeoeffnet(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	var gemeldet []any

	anwendung.MitMeldung(func(_ context.Context, ereignis string, daten ...any) {
		if ereignis == app.EreignisDateiGeoeffnet {
			gemeldet = daten
		}
	})

	anwendung.Bereit(t.Context())
	anwendung.DateiOeffnen(probeset)

	// **Kennung und Epoche des Bestands** (J05, #2269): an der Epoche erkennt
	// das Fenster eine Meldung aus einem Bestand, der inzwischen getauscht ist.
	if len(gemeldet) != 2 {
		t.Fatalf("gemeldet wurde %v — erwartet waren Setkennung und Epoche", gemeldet)
	}

	if kennung, istText := gemeldet[0].(string); !istText || kennung == "" {
		t.Errorf("gemeldet wurde %#v — erwartet war eine Setkennung", gemeldet[0])
	}

	if _, istZahl := gemeldet[1].(uint64); !istZahl {
		t.Errorf("gemeldet wurde %#v — erwartet war die Epoche des Bestands", gemeldet[1])
	}
}

// **Eine Datei, die vor dem Fenster kommt, wird nachgeholt.**
//
// # Warum das der Normalfall ist und nicht die Ausnahme
//
// Der Finder schickt das `odoc`-Ereignis **beim Start** — es kann eintreffen,
// bevor Wails `OnStartup` gerufen hat und die Anwendung einen Kontext hat.
// Wer es dann meldet, meldet ins Leere: es hört niemand zu, und die Datei
// bleibt ungeöffnet, ohne dass irgendetwas fehlschlüge.
func TestEineDateiVorDemFensterWirdNachgeholt(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	var gemeldet []any

	anwendung.MitMeldung(func(_ context.Context, ereignis string, daten ...any) {
		if ereignis == app.EreignisDateiGeoeffnet {
			gemeldet = daten
		}
	})

	// **Vor `Bereit`** — so kommt es vom Finder.
	anwendung.DateiOeffnen(probeset)

	if len(gemeldet) != 0 {
		t.Fatalf("ohne Fenster wurde schon gemeldet: %v", gemeldet)
	}

	anwendung.Bereit(t.Context())

	if len(gemeldet) != 2 {
		t.Errorf("nach dem Fenster wurde %v gemeldet", gemeldet)
	}
}

// **Zwei Dateien vor dem Fenster werden beide geöffnet.**
//
// Wails ruft `OnFileOpen` je Datei; wer im Finder zwei Sets markiert und
// öffnet, löst zwei Aufrufe aus — beide vor `Bereit`. Bis zum 2026-08-25
// überschrieb der zweite den ersten, und **eine der beiden Dateien wurde
// wortlos nicht geöffnet** (Zweitblick über Phase Y, Befund 3).
//
// **Und die letzte gewinnt die Anzeige** — dieselbe Ordnung wie bei Aufrufen
// nach dem Fenster.
func TestZweiDateienVorDemFensterWerdenBeideGeoeffnet(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	var gemeldet []string

	anwendung.MitMeldung(func(_ context.Context, ereignis string, daten ...any) {
		if ereignis == app.EreignisDateiGeoeffnet {
			if kennung, istText := daten[0].(string); istText {
				gemeldet = append(gemeldet, kennung)
			}
		}
	})

	// **Zweimal dasselbe Set geht nicht** — der zweite Aufruf fände es im
	// Bestand und gäbe dieselbe Kennung. Zwei verschiedene Dateien sagen, ob
	// beide angekommen sind.
	// **Kopiert und nicht verknüpft**: `probe.mse-set` ist ein
	// **Verzeichnis** — MSE-Pakete sind beides, gepackt und ausgepackt.
	zweites := filepath.Join(t.TempDir(), "zweites.mse-set")
	if err := os.CopyFS(zweites, os.DirFS(probeset)); err != nil {
		t.Fatalf("zweite Probe anlegen: %v", err)
	}

	anwendung.DateiOeffnen(probeset)
	anwendung.DateiOeffnen(zweites)

	if len(gemeldet) != 0 {
		t.Fatalf("ohne Fenster wurde schon gemeldet: %v", gemeldet)
	}

	anwendung.Bereit(t.Context())

	if len(gemeldet) != 2 {
		t.Fatalf("gemeldet wurden %d Sets, erwartet 2: %v", len(gemeldet), gemeldet)
	}

	if gemeldet[0] == gemeldet[1] {
		t.Errorf("beide Meldungen nennen dieselbe Kennung %q — "+
			"die zweite Datei wurde nicht eingefahren", gemeldet[0])
	}
}

// **Eine Datei, die sich nicht lesen lässt, meldet ihren Grund.**
//
// Nicht stumm bleiben: wer eine Datei doppelklickt und nichts geschieht, sucht
// den Fehler bei sich (§8).
func TestEineUnlesbareDateiMeldetIhrenGrund(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	var gemeldet []any

	anwendung.MitMeldung(func(_ context.Context, ereignis string, daten ...any) {
		if ereignis == app.EreignisDateiFehler {
			gemeldet = daten
		}
	})

	fehlt := filepath.Join(t.TempDir(), "gibt-es-nicht.mse-set")

	anwendung.Bereit(t.Context())
	anwendung.DateiOeffnen(fehlt)

	if len(gemeldet) != 2 {
		t.Fatalf("gemeldet wurde %v — erwartet waren Pfad und Fehler", gemeldet)
	}

	// **Der Pfad, damit das Fenster sagen kann, welche Datei gemeint ist.**
	if pfad, istText := gemeldet[0].(string); !istText || pfad != fehlt {
		t.Errorf("gemeldet wurde der Pfad %#v", gemeldet[0])
	}

	// **Und der Fehler mit seinem Schlüssel** — den Satz baut das Fenster, denn
	// Go kennt die Sprache der Sitzung nicht (§9).
	befund, istFehler := gemeldet[1].(*app.Fehler)
	if !istFehler {
		t.Fatalf("gemeldet wurde %#v — erwartet war ein Fehler", gemeldet[1])
	}

	if befund.Schluessel != "nichtGefunden" {
		t.Errorf("der Schlüssel ist %q", befund.Schluessel)
	}
}

// **Ein leerer Pfad tut nichts.**
//
// Er entsteht, wenn niemand eine Datei übergeben hat — der gewöhnliche Start.
func TestEinLeererPfadTutNichts(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	var gemeldet int

	anwendung.MitMeldung(func(context.Context, string, ...any) { gemeldet++ })

	anwendung.Bereit(t.Context())
	anwendung.DateiOeffnen("")

	if gemeldet != 0 {
		t.Errorf("ein leerer Pfad meldete %d Ereignisse", gemeldet)
	}
}
