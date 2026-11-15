// **Der Weg vom Naht-Fehler zurück zum Go-Fehler** — Werkstück
// #1391.
//
// # Warum das geprüft gehört
//
// An dieser Umkehrung hängen die Rückgabewerte aller drei Flächen ohne
// Fenster: die Kommandozeile setzt daraus ihren Code, MCP sein `isError`, der
// Dienst seinen Status. Alle drei fragen mit `errors.Is` — und wenn die
// Klasse dabei verlorengeht, antwortet der Dienst 500 auf eine Validierung.
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAlsFehlerGibtKlasseUndSatzZurueck(t *testing.T) {
	for name, fall := range map[string]struct {
		code   app.Fehlercode
		klasse error
	}{
		"nicht gefunden": {app.CodeNichtGefunden, domain.ErrNotFound},
		"Konflikt":       {app.CodeKonflikt, domain.ErrConflict},
		"Validierung":    {app.CodeValidierung, domain.ErrValidation},
	} {
		t.Run(name, func(t *testing.T) {
			fehler := (&app.Fehler{Code: fall.code, Meldung: "der Satz"}).AlsFehler()

			if !errors.Is(fehler, fall.klasse) {
				t.Errorf("Klasse fehlt: %v", fehler)
			}

			// **Der Satz steht einmal da.** Ein `fmt.Errorf("%w: %s", …)`
			// setzte das Präfix der Klasse davor — und `Meldung` trägt es
			// schon, weil `mitSatz` `fehler.Error()` hineinlegt. Genau das
			// war der erste Anlauf, und die Ausgabe eines Wächters hat es
			// gezeigt: „Validierung fehlgeschlagen: Validierung
			// fehlgeschlagen: …".
			if fehler.Error() != "der Satz" {
				t.Errorf("Satz %q, erwartet %q", fehler.Error(), "der Satz")
			}
		})
	}
}

// **Ein unbekannter Code bekommt keine Klasse.**
//
// [app.CodeUnbekannt] heißt genau das: wir wissen nicht, was es war. Ihn auf
// `ErrValidation` abzubilden hieße zu behaupten, der Aufrufer sei schuld.
func TestAlsFehlerGibtDemUnbekanntenKeineKlasse(t *testing.T) {
	fehler := (&app.Fehler{Code: app.CodeUnbekannt, Meldung: "irgendetwas"}).AlsFehler()

	for _, klasse := range []error{
		domain.ErrValidation, domain.ErrNotFound, domain.ErrConflict,
	} {
		if errors.Is(fehler, klasse) {
			t.Errorf("%v trägt eine Klasse, die es nicht kennt: %v", fehler, klasse)
		}
	}

	if fehler.Error() != "irgendetwas" {
		t.Errorf("Satz %q", fehler.Error())
	}
}

// **Und kein Fehler bleibt keiner.**
func TestAlsFehlerVonNilIstNil(t *testing.T) {
	var keiner *app.Fehler

	if fehler := keiner.AlsFehler(); fehler != nil {
		t.Errorf("aus nichts wurde %v", fehler)
	}
}
