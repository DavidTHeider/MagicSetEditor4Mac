// **Der Fehlerweg beim Füllen der Nachbardatei** —
// #853.
//
// **Ein interner Test**, und der Grund gehört dazu: ein `Write` auf eine
// frisch angelegte Datei scheitert nur an einer vollen Platte oder einem
// entzogenen Medium — beides in einem Test nicht herstellbar. Über eine
// Attrappe ist der Weg erreichbar, und ein Fehlerweg, den kein Test betritt,
// ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
package service

import (
	"errors"
	"strings"
	"testing"
)

// stolpernderSchreiber scheitert beim Schreiben, beim Schließen oder gar nicht.
type stolpernderSchreiber struct {
	beimSchreiben, beimSchliessen error
}

func (s stolpernderSchreiber) Write(p []byte) (int, error) {
	if s.beimSchreiben != nil {
		return 0, s.beimSchreiben
	}

	return len(p), nil
}

func (s stolpernderSchreiber) Close() error { return s.beimSchliessen }

func TestDieNachbardateiMeldetJedenIhrerFehler(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		datei    stolpernderSchreiber
		erwartet string
	}{
		{"gelingt", stolpernderSchreiber{}, ""},

		{
			"scheitert beim Schreiben",
			stolpernderSchreiber{beimSchreiben: errors.New("kein Platz")},
			"füllen",
		},

		// **Ein gepufferter Schreibvorgang scheitert erst beim Schließen.**
		// Wer den Fehler verwürfe, hielte eine halb geschriebene Datei für
		// fertig und benennte sie um.
		{
			"scheitert beim Schließen",
			stolpernderSchreiber{beimSchliessen: errors.New("kein Platz")},
			"schließen",
		},

		// **Und beides zugleich**: der Schließfehler verdeckt den ersten
		// nicht.
		{
			"scheitert zweimal",
			stolpernderSchreiber{
				beimSchreiben:  errors.New("kein Platz"),
				beimSchliessen: errors.New("und auch das nicht"),
			},
			"füllen",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := fuelleUndSchliesse(fall.datei, []byte("inhalt"))

			if fall.erwartet == "" {
				if fehler != nil {
					t.Errorf("Fehler %v, erwartet war keiner", fehler)
				}

				return
			}

			if fehler == nil || !strings.Contains(fehler.Error(), fall.erwartet) {
				t.Errorf("Fehler %v, erwartet war einer über %q", fehler, fall.erwartet)
			}
		})
	}
}
