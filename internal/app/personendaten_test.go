package app_test

// **Der Export personenbezogener Daten am Fenster** — Werkstück 8 von
// Phase 4.12
// (#1351),
// Punkt 10.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Auskunft nennt den Menschen, seine Profile und ihren Stichtag.**
func TestDiePersonenauskunftNenntWasDasteht(t *testing.T) {
	antwort := angelaufeneAnwendung(t).Personendaten(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Daten.Nutzer.Name == "" {
		t.Error("der Name fehlt")
	}

	if len(antwort.Daten.Profile) == 0 {
		t.Error("die Profile fehlen")
	}

	// **Eine Auskunft ohne Datum ist keine** — sie sagt, was am Stichtag galt.
	if antwort.Daten.Erhoben.IsZero() {
		t.Error("der Stichtag fehlt")
	}
}

// **Sie geht als Datei hinaus — und trägt kein Geheimnis.**
//
// # Warum das hier steht und nicht nur im Dienst
//
// **Weil die Datei der Weg ist, auf dem sie das Programm verlässt.** Der
// Dienst hält die Zusage über einen Verbund; ob sie **geschrieben** auch gilt,
// sagt erst ein Blick in die Datei — das ist die Naht, und zwei grüne Seiten
// sind kein Beleg über sie.
func TestDiePersonenauskunftGehtOhneGeheimnisHinaus(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ziel := filepath.Join(t.TempDir(), "meine-daten.json")

	// **Ein echtes Kennwort geht hinein** — und muss nicht wieder heraus.
	//
	// **Ohne diesen Schritt prüfte der Lauf nichts**: der erste Nutzer eines
	// frischen Bestands hat gar keines, und eine Datei, in der nichts steht,
	// nennt es auch nicht. Das ist §10: eine Gegenprobe, die nichts ändert,
	// ist keine.
	const geheim = "ein sehr geheimes Wort"

	angelegt := anwendung.KontoEintragen(t.Context(), "wer@example.org", geheim, "Wer")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if antwort := anwendung.PersonendatenSchreiben(
		t.Context(), angelegt.Nutzer.ID, ziel); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	roh, fehler := os.ReadFile(ziel) // #nosec G304 -- ein Pfad aus t.TempDir()
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Weder das Wort noch seine gehütete Form.** Das zweite ist der Fall,
	// den man vergisst: ein Argon2id-Hash ist kein Kennwort und trotzdem
	// nichts, was in eine Auskunft gehört.
	for _, verboten := range []string{geheim, "argon2", "otpauth", "$"} {
		if strings.Contains(string(roh), verboten) {
			t.Errorf("die Datei nennt %q", verboten)
		}
	}

	// **Dass es eines gibt, steht drin** — die Auskunft verschweigt es nicht.
	var gepruefte struct {
		Nutzer struct {
			HatKennwort bool `json:"hatKennwort"`
		} `json:"nutzer"`
	}

	if err := json.Unmarshal(roh, &gepruefte); err != nil {
		t.Fatalf("die Datei lesen: %v", err)
	}

	if !gepruefte.Nutzer.HatKennwort {
		t.Error("die Auskunft verschweigt, dass ein Kennwort gesetzt ist")
	}

	// **Und sie ist lesbares JSON** — eine Auskunft, die niemand aufmachen
	// kann, ist keine.
	var gelesen map[string]any
	if err := json.Unmarshal(roh, &gelesen); err != nil {
		t.Errorf("die Datei ist kein JSON: %v", err)
	}

	if gelesen["erhoben"] == nil {
		t.Error("der Stichtag fehlt in der Datei")
	}
}

// **Eine kaputte Kennung, ein fehlender Nutzer und ein unmögliches Ziel
// werden gemeldet.**
func TestDiePersonenauskunftWeistUnsinnAb(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	if antwort := anwendung.Personendaten(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("die Kennung kam durch")
	}

	if antwort := anwendung.Personendaten(t.Context(), "4711"); antwort.Fehler == nil {
		t.Error("der fehlende Nutzer kam durch")
	}

	if antwort := anwendung.PersonendatenSchreiben(
		t.Context(), "1", ""); antwort.Fehler == nil {
		t.Error("das leere Ziel kam durch")
	}

	// **Ein Verzeichnis, das es nicht gibt** — die Auskunft legt keines an.
	if antwort := anwendung.PersonendatenSchreiben(t.Context(), "1",
		filepath.Join(t.TempDir(), "gibt", "es", "nicht.json")); antwort.Fehler == nil {
		t.Error("der unmögliche Ort kam durch")
	}

	if antwort := anwendung.PersonendatenSchreiben(
		t.Context(), "keine Zahl", filepath.Join(t.TempDir(), "x.json")); antwort.Fehler == nil {
		t.Error("die Kennung kam beim Schreiben durch")
	}
}

// **Der Dialog gibt das Ziel — und ein Abbruch ist kein Fehler.**
func TestDerAuskunftsdialogSichertOderBrichtAb(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ziel := filepath.Join(t.TempDir(), "meine-daten.json")

	anwendung.MitDateiwahl(func(_ context.Context, _ app.Dateiauswahl) (string, error) {
		return ziel, nil
	})

	if antwort := anwendung.PersonendatenSichern(t.Context(), "1"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}

	anwendung.MitDateiwahl(func(_ context.Context, _ app.Dateiauswahl) (string, error) {
		return "", nil
	})

	if antwort := anwendung.PersonendatenSichern(t.Context(), "1"); antwort.Fehler != nil {
		t.Errorf("der Abbruch wurde gemeldet: %+v", antwort.Fehler)
	}

	anwendung.MitDateiwahl(func(_ context.Context, _ app.Dateiauswahl) (string, error) {
		return "", errors.New("kein Fenster")
	})

	if antwort := anwendung.PersonendatenSichern(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}
