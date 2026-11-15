// **Die Symbolnaht** —
// #890,
// Werkstück 4 von Phase X.
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// symbolprobe merkt sich, was abgelegt wurde.
type symbolprobe struct {
	abgelegt msepaket.Symbolzeichnung
	gestalt  msepaket.Symbolzeichnung
	fehler   error
	setID    int64
	name     string

	// unter hält, unter welchem Namen zuletzt abgelegt wurde.
	unter string
}

func (s *symbolprobe) Ablegen(
	setID int64, zeichnung msepaket.Symbolzeichnung, unter string,
) (string, error) {
	s.setID, s.abgelegt, s.unter = setID, zeichnung, unter

	if unter != "" {
		return unter, s.fehler
	}

	return "symbol1.mse-symbol", s.fehler
}

func (s *symbolprobe) Gestalt(setID int64, name string) (msepaket.Symbolzeichnung, error) {
	s.setID, s.name = setID, name

	return s.gestalt, s.fehler
}

// **Eine gezeichnete Form geht durch die Naht und kommt zurück.**
//
// **`KurveDavor` wird abgeleitet, nicht gereicht.** Eine Strecke hat eine Art,
// und die schreibt der Punkt davor (`enforceConstraints`, `src/data/symbol.cpp`,
// Zeilen 205–212, Fassung f9d9356 vom 2021-07-06) — wer beide Seiten reichen
// ließe, könnte sie widersprüchlich setzen.
func TestEineGezeichneteFormGehtDurchDieNaht(t *testing.T) {
	t.Parallel()

	probe := &symbolprobe{}

	anwendung := &app.App{}
	anwendung.MitSymbolablage(probe)

	antwort := anwendung.SymbolSetzen(t.Context(), "7", []app.SymbolteilDTO{{
		Art: "Form", Name: "Bogen", Kombination: "merge",
		Punkte: []app.SymbolpunktDTO{
			{X: 0, Y: 0, GriffDanachX: 0.5, KurveDanach: true, Sperre: "size"},
			{X: 1, Y: 1, GriffDavorY: -0.5, Sperre: "size"},
		},
	}}, "")

	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if antwort.Name != "symbol1.mse-symbol" || probe.setID != 7 {
		t.Errorf("abgelegt wurde %q für Set %d", antwort.Name, probe.setID)
	}

	if len(probe.abgelegt.Teile) != 1 {
		t.Fatalf("die Zeichnung ist %+v", probe.abgelegt.Teile)
	}

	punkte := probe.abgelegt.Teile[0].Punkte
	if len(punkte) != 2 {
		t.Fatalf("%d Punkte", len(punkte))
	}

	// **Der zweite Punkt hat eine Kurve davor**, weil der erste eine danach
	// hat — abgeleitet, nicht gereicht.
	if !punkte[1].KurveDavor {
		t.Error("die Strecke zum zweiten Punkt ist eine Gerade")
	}

	// **Und der erste hat eine davor**, weil der Zug geschlossen ist und der
	// letzte keine danach hat.
	if punkte[0].KurveDavor {
		t.Error("die Schlussstrecke ist eine Kurve, obwohl der letzte Punkt keine setzt")
	}
}

// **Eine Symmetrie mit Kindern kommt vollständig zurück.**
func TestEineSymmetrieMitKindernKommtZurueck(t *testing.T) {
	t.Parallel()

	probe := &symbolprobe{gestalt: msepaket.Symbolzeichnung{
		Teile: []msepaket.Symbolteil{{
			Art: msepaket.TeilSymmetrie, Name: "Spiegel",
			Spiegelart: msepaket.SpiegelungAchse, Kopien: 4,
			Mitte: msepaket.Vektor{X: 0.5, Y: 0.5},
			Griff: msepaket.Vektor{X: 0, Y: -0.5},
			Teile: []msepaket.Symbolteil{{
				Art: msepaket.TeilForm, Name: "Kind",
				Punkte: []msepaket.Kontrollpunkt{{Stelle: msepaket.Vektor{X: 1, Y: 1}}},
			}},
		}},
	}}

	anwendung := &app.App{}
	anwendung.MitSymbolablage(probe)

	antwort := anwendung.Symbol(t.Context(), "7", "symbol1.mse-symbol")
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if probe.name != "symbol1.mse-symbol" {
		t.Errorf("gefragt wurde nach %q", probe.name)
	}

	if len(antwort.Teile) != 1 || antwort.Teile[0].Kopien != 4 {
		t.Fatalf("die Teile sind %+v", antwort.Teile)
	}

	if len(antwort.Teile[0].Teile) != 1 || antwort.Teile[0].Teile[0].Name != "Kind" {
		t.Errorf("die Kinder sind %+v", antwort.Teile[0].Teile)
	}
}

// **Ohne Naht, mit unbrauchbarer Kennung und mit einem Fehler dahinter.**
func TestDieSymbolnahtMeldetIhreFehler(t *testing.T) {
	t.Parallel()

	t.Run("ohne Naht", func(t *testing.T) {
		t.Parallel()

		anwendung := &app.App{}

		if antwort := anwendung.SymbolSetzen(t.Context(), "7", nil, ""); antwort.Fehler == nil {
			t.Error("das Setzen ohne Naht ging")
		}

		if antwort := anwendung.Symbol(t.Context(), "7", "a"); antwort.Fehler == nil {
			t.Error("das Lesen ohne Naht ging")
		}
	})

	t.Run("unbrauchbare Kennung", func(t *testing.T) {
		t.Parallel()

		anwendung := &app.App{}
		anwendung.MitSymbolablage(&symbolprobe{})

		if antwort := anwendung.SymbolSetzen(t.Context(), "keine", nil, ""); antwort.Fehler == nil {
			t.Error("das Setzen mit unbrauchbarer Kennung ging")
		}

		if antwort := anwendung.Symbol(t.Context(), "keine", "a"); antwort.Fehler == nil {
			t.Error("das Lesen mit unbrauchbarer Kennung ging")
		}
	})

	t.Run("ein Fehler der Ablage kommt durch", func(t *testing.T) {
		t.Parallel()

		anwendung := &app.App{}
		anwendung.MitSymbolablage(&symbolprobe{
			fehler: errors.New("die Platte ist voll: " + domain.ErrValidation.Error()),
		})

		if antwort := anwendung.SymbolSetzen(t.Context(), "7", nil, ""); antwort.Fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}

		if antwort := anwendung.Symbol(t.Context(), "7", "a"); antwort.Fehler == nil {
			t.Error("der Fehler kam nicht durch")
		}
	})
}

// **Die vier Grundformen kommen durch die Naht.**
//
// **Gerechnet wird in `service`** — hier steht nur, dass die Naht sie
// weiterreicht und ihre Punkte nicht unterwegs verliert.
func TestDieGrundformenKommenDurchDieNaht(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}

	for name, fall := range map[string]struct {
		form   app.SymbolgrundformDTO
		punkte int
		heisst string
	}{
		"Kreis":    {app.SymbolgrundformDTO{Art: "Kreis", Eng: true}, 4, "Kreis"},
		"Ellipse":  {app.SymbolgrundformDTO{Art: "Kreis"}, 4, "Ellipse"},
		"Rechteck": {app.SymbolgrundformDTO{Art: "Rechteck"}, 4, "Rechteck"},
		"Dreieck":  {app.SymbolgrundformDTO{Art: "Vieleck", Ecken: 3}, 3, "Dreieck"},
		"Stern":    {app.SymbolgrundformDTO{Art: "Stern", Ecken: 5}, 10, "Stern"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := anwendung.Grundform(fall.form)
			if antwort.Fehler != nil {
				t.Fatalf("Fehler: %+v", antwort.Fehler)
			}

			if len(antwort.Teile) != 1 {
				t.Fatalf("%d Teile", len(antwort.Teile))
			}

			if antwort.Teile[0].Name != fall.heisst {
				t.Errorf("die Form heißt %q", antwort.Teile[0].Name)
			}

			if hat := len(antwort.Teile[0].Punkte); hat != fall.punkte {
				t.Errorf("%d Punkte, will %d", hat, fall.punkte)
			}
		})
	}
}

// **Eine Form, die sich nicht zeichnen lässt, ist ein Fehler und kein leerer
// Teil.**
//
// In `service` gibt sie eine leere Form — dort ist es Geometrie, und der
// Aufrufer entscheidet. **An der Naht ist sie eine Fehleingabe**: die
// Oberfläche hat vier Knöpfe, und was hier ankommt, kam nicht von einem davon.
func TestEineUnzeichenbareGrundformIstEinFehler(t *testing.T) {
	t.Parallel()

	anwendung := &app.App{}

	for name, form := range map[string]app.SymbolgrundformDTO{
		"eine fremde Art":      {Art: "Kegelschnitt", Ecken: 5},
		"zu wenige Ecken":      {Art: "Vieleck", Ecken: 2},
		"ein Stern ohne Ecken": {Art: "Stern"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if antwort := anwendung.Grundform(form); antwort.Fehler == nil {
				t.Errorf("%+v kam durch", form)
			}
		})
	}
}
