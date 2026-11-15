package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine kaputte Datei ist nie ein leeres, gelungenes Set** — FMT-06.AK-04
// (#1695).
//
// # Der Bestand dafür
//
// `testdaten/fehlerdateien/` — **selbst geschriebene** Minimaldateien, je eine
// Fehlerklasse (`testdaten/LIESMICH.md`). Fremde Pakete dürfen nicht ins
// Repositorium (§11); was dort gefunden wird, wird hier nachgebaut.
//
// # Was jede ergeben muss
//
// Einen **benannten** Fehler — mit der Stelle, wo es eine gibt — oder ein Set
// samt dem, was nicht gelesen wurde. **Nie** „kein Fehler, keine Karte, kein
// Befund". Und nie „ein unerwarteter Fehler": bis FMT-06 kamen ungültiges
// UTF-8, UTF-16 und ein abgeschnittenes Archiv genau so an, und die Zeile,
// die `msepaket` nennt, erreichte niemanden.
func TestEineKaputteDateiIstNieEinLeeresSet(t *testing.T) {
	faelle := map[string]struct {
		code  string
		steht string // ein Wort, das im Satz stehen muss
	}{
		"kein-utf8.mse-set":     {"validierung", "Zeile 5"},
		"utf16.mse-set":         {"validierung", "UTF-16"},
		"abgeschnitten.mse-set": {"validierung", "Archiv"},
		"leer.mse-set":          {"validierung", "Spiel"},
		"nur-bom.mse-set":       {"validierung", "Spiel"},
		"kaputte-karte.mse-set": {"", ""},
	}

	ordner := filepath.Join("..", "..", "testdaten", "fehlerdateien")

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Jede Datei im Ordner steht in der Tabelle** — eine, die hinzukommt,
	// ohne eine Erwartung zu bekommen, prüfte nichts.
	if len(eintraege) != len(faelle) {
		t.Errorf("%d Dateien im Ordner, %d in der Tabelle", len(eintraege), len(faelle))
	}

	for _, eintrag := range eintraege {
		t.Run(eintrag.Name(), func(t *testing.T) {
			erwartet, bekannt := faelle[eintrag.Name()]
			if !bekannt {
				t.Fatalf("%s hat keine Erwartung", eintrag.Name())
			}

			pfad, err := filepath.Abs(filepath.Join(ordner, eintrag.Name()))
			if err != nil {
				t.Fatalf("Pfad: %v", err)
			}

			anwendung := gestartet(t)
			antwort := anwendung.SetOeffnen(t.Context(), pfad)

			if antwort.Fehler == nil && antwort.Karten == 0 && len(antwort.Befunde) == 0 {
				t.Fatal("kein Fehler, keine Karte, kein Befund — ein leeres Set aus einer kaputten Datei")
			}

			if erwartet.code == "" {
				pruefeMitgefuehrt(t, anwendung, antwort)

				return
			}

			if antwort.Fehler == nil {
				t.Fatalf("erwartet war ein Fehler %q, gekommen ist ein Set mit %d Karten",
					erwartet.code, antwort.Karten)
			}

			if string(antwort.Fehler.Code) != erwartet.code {
				t.Errorf("Code %q, erwartet %q (%s)", antwort.Fehler.Code, erwartet.code, antwort.Fehler.Meldung)
			}

			if !strings.Contains(antwort.Fehler.Meldung, erwartet.steht) {
				t.Errorf("der Satz nennt %q nicht: %s", erwartet.steht, antwort.Fehler.Meldung)
			}

			// **Das Fenster setzt den Grund in „Das ging nicht: …" ein**, und
			// dort soll er mit der Datei beginnen — nicht mit dem Vorgang
			// „Set öffnen:", den der Satz schon sagt.
			if grund, gesetzt := antwort.Fehler.Werte["grund"]; gesetzt &&
				!strings.HasPrefix(grund, eintrag.Name()) {
				t.Errorf("der Grund beginnt nicht mit der Datei: %s", grund)
			}
		})
	}
}

// pruefeMitgefuehrt sagt zu, dass die Zeile ohne Doppelpunkt nicht still
// verschwindet: sie reist als Vorbehalt mit (Vermerk 0014 — der Zerteiler
// kann sie nicht abweisen, der Import führt sie mit).
func pruefeMitgefuehrt(t *testing.T, anwendung *app.App, antwort app.OeffnenAntwort) {
	t.Helper()

	if antwort.Fehler != nil {
		t.Fatalf("öffnen: %+v", antwort.Fehler)
	}

	vorbehalte := anwendung.Vorbehalte(t.Context(), antwort.SetID)

	for _, vorbehalt := range vorbehalte.Vorbehalte {
		if strings.Contains(vorbehalt.Inhalt, "name Ohne Doppelpunkt") {
			return
		}
	}

	t.Errorf("die Zeile ohne Doppelpunkt steht in keinem Vorbehalt: %+v", vorbehalte)
}
