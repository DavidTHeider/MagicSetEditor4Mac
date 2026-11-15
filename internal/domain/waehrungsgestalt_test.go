package domain_test

// **Ein Währungscode hat an jeder Stelle dieselbe Gestalt** — UX-06.AK-02 und
// AK-04 (#1678).
//
// # Warum es diese Läufe gibt
//
// **Drei Stellen prüften ihn, und jede anders.** Die Einstellung zählte
// Runen, das Kostenjournal Bytes — und ließ damit `€` durch, das in UTF-8
// drei davon hat —, und das Exemplar prüfte die Gestalt gar nicht: `EURO`
// stand als Währung eines Erwerbspreises im Bestand. Dieselbe Angabe hieß
// damit an drei Flächen dreierlei.
//
// **Plausibel heißt nicht bekannt** (AK-04): drei lateinische Buchstaben sind
// die Gestalt eines Codes nach ISO 4217, und mehr wird hier nicht behauptet.
// Ob es die Währung gibt, weiß keine dieser Stellen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// waehrungsfaelle sind die Eingaben, die jede Stelle gleich beurteilen muss.
var waehrungsfaelle = []struct {
	name    string
	eingabe string
	gestalt bool
}{
	{"drei Buchstaben", "EUR", true},
	{"vier Buchstaben", "EURO", false},
	{"zwei Buchstaben", "EU", false},
	{"das Eurozeichen hat drei Bytes und ist keine", "€", false},
	{"Ziffern sind keine", "978", false},
	{"ein Umlaut ist kein lateinischer Buchstabe", "ÄUR", false},
	{"ein Leerzeichen in der Mitte", "E R", false},
}

func TestEinErwerbspreisTraegtEinenWaehrungscode(t *testing.T) {
	for _, fall := range waehrungsfaelle {
		t.Run(fall.name, func(t *testing.T) {
			exemplar := domain.Exemplar{
				Printing: "abc", Menge: 1, Sprache: "en",
				Erwerbspreis: 150, Erwerbswaehrung: fall.eingabe,
			}
			exemplar.Normalisieren()

			befund := befundAm(exemplar.Validate(), "erwerbswaehrung")
			if fall.gestalt && befund != nil {
				t.Errorf("%q abgewiesen: %v", fall.eingabe, befund)
			}

			if !fall.gestalt && befund == nil {
				t.Errorf("%q als Währung angenommen", fall.eingabe)
			}
		})
	}
}

// **Klein geschrieben bleibt beim Exemplar erlaubt** — `Normalisieren` hebt
// es an (0057); abgewiesen wird erst, was auch groß keine Gestalt hat.
func TestEinKleinGeschriebenerCodeAmExemplarWirdGross(t *testing.T) {
	exemplar := domain.Exemplar{
		Printing: "abc", Menge: 1, Sprache: "en",
		Erwerbspreis: 150, Erwerbswaehrung: " eur ",
	}
	exemplar.Normalisieren()

	if befunde := exemplar.Validate(); len(befunde) != 0 {
		t.Fatalf("abgewiesen: %v", befunde)
	}

	if exemplar.Erwerbswaehrung != "EUR" {
		t.Errorf("Währung ist %q, erwartet EUR", exemplar.Erwerbswaehrung)
	}
}

func TestEinBetragTraegtEinenWaehrungscode(t *testing.T) {
	for _, fall := range waehrungsfaelle {
		t.Run(fall.name, func(t *testing.T) {
			err := domain.Betrag{Wert: 150, Waehrung: fall.eingabe}.Pruefe()

			if fall.gestalt && err != nil {
				t.Errorf("%q abgewiesen: %v", fall.eingabe, err)
			}

			if !fall.gestalt && !errors.Is(err, domain.ErrValidation) {
				t.Errorf("%q als Währung angenommen: %v", fall.eingabe, err)
			}
		})
	}
}

// **Und die Einstellung urteilt genauso** — sie war die Stelle, die es schon
// richtig machte, und bleibt der Maßstab.
func TestDieEinstellungUrteiltWieDieAnderen(t *testing.T) {
	for _, fall := range waehrungsfaelle {
		t.Run(fall.name, func(t *testing.T) {
			err := domain.PruefeWaehrungscode(fall.eingabe)

			if fall.gestalt != (err == nil) {
				t.Errorf("%q: Gestalt %v, Urteil %v", fall.eingabe, fall.gestalt, err)
			}
		})
	}
}

// befundAm sucht den Befund an einem Feld.
func befundAm(befunde []*domain.FieldError, feld string) *domain.FieldError {
	for _, befund := range befunde {
		if befund.Field == feld {
			return befund
		}
	}

	return nil
}
