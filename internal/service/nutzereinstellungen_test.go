package service_test

// **Die Einstellungen, die dem Menschen gehören** — Werkstück 8 von
// Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Antwort ist vollständig, auch wenn nichts gesetzt ist.**
//
// **Sonst zwänge sie jede Fläche, die Vorgabe noch einmal zu kennen** — und
// die zweite Liste liefe zurück (§10, gefunden am `bildkonflikt`).
func TestDieAntwortIstVollstaendig(t *testing.T) {
	t.Parallel()

	alle, fehler := service.NeueNutzereinstellungen(&einstellungsdoppel{}).
		Alle(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	for _, schluessel := range domain.Nutzereinstellungsreihe() {
		if alle[schluessel] != domain.NutzerVorgabe(schluessel) {
			t.Errorf("%s ist %q, erwartet war die Vorgabe %q",
				schluessel, alle[schluessel], domain.NutzerVorgabe(schluessel))
		}
	}
}

// **Ein gesetzter Wert sticht die Vorgabe.**
func TestEinGesetzterWertStichtDieVorgabe(t *testing.T) {
	t.Parallel()

	dienst := service.NeueNutzereinstellungen(&einstellungsdoppel{})

	if fehler := dienst.Setzen(t.Context(), 1,
		domain.NutzerKontrast, domain.KontrastHoch); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	wert, fehler := dienst.Wert(t.Context(), 1, domain.NutzerKontrast)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if wert != domain.KontrastHoch {
		t.Errorf("gelesen wurde %q", wert)
	}
}

// **Ein unbekanntes Paar kommt gar nicht erst in den Bestand.**
func TestEinUnbekanntesPaarKommtNichtInDenBestand(t *testing.T) {
	t.Parallel()

	ablage := &einstellungsdoppel{}
	dienst := service.NeueNutzereinstellungen(ablage)

	if fehler := dienst.Setzen(t.Context(), 1, "gibtsnicht", "egal"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}

	if len(ablage.werte) != 0 {
		t.Errorf("es steht etwas da: %v", ablage.werte)
	}
}

// **Ein kaputter Bestand reißt an jedem Weg ab.**
func TestEinKaputterBestandReisstAnJederEinstellungAb(t *testing.T) {
	t.Parallel()

	wege := map[string]func(*service.Nutzereinstellungen) error{
		"beim Lesen": func(d *service.Nutzereinstellungen) error {
			_, fehler := d.Alle(t.Context(), 1)

			return fehler
		},
		"beim einzelnen Lesen": func(d *service.Nutzereinstellungen) error {
			_, fehler := d.Wert(t.Context(), 1, domain.NutzerKontrast)

			return fehler
		},
		"beim Setzen": func(d *service.Nutzereinstellungen) error {
			return d.Setzen(t.Context(), 1, domain.NutzerKontrast, domain.KontrastHoch)
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dienst := service.NeueNutzereinstellungen(
				&einstellungsdoppel{fehler: errStummerBestand})

			if fehler := weg(dienst); !errors.Is(fehler, errStummerBestand) {
				t.Errorf("durchgelassen wurde: %v", fehler)
			}
		})
	}
}
