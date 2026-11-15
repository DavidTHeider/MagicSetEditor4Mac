// **Eine Vorgabe sieht die gespeicherten Nachbarwerte** —
// #860.
//
// # Der Unterschied zu #789
//
// `stiloptionumgebung_test.go` prüft, dass `styling` überhaupt in der Umgebung
// liegt — mit einer **ungesetzten** Nachbaroption auf ihrem Anfangswert. Der
// Nachlauf zum Zweitblick über Phase P hat angestrichen, dass damit der Fall
// geprüft war und nicht die Klasse: `Stilwaehler.Optionen` baute die Umgebung
// **nur** aus den Anfangswerten und legte die gespeicherten erst danach auf
// das Ausgabe-DTO.
//
// MSE bindet dagegen unmittelbar `set.stylingDataFor(*stylesheet)` — die
// tatsächlichen Stilwerte (`src/script/script_manager.cpp`, Zeilen 30–61, und
// `src/data/set.cpp`, Zeilen 99–105, Fassung f9d9356 vom 2021-07-06).
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nachbarwaehler gibt einen Stilwähler über `nachbarvorgabe.mse-style`.
func nachbarwaehler(t *testing.T, werte []domain.Value) *service.Stilwaehler {
	t.Helper()

	return service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Game: "probe", Style: "nachbarvorgabe"},
		}},
		&ablage{pakete: []domain.Paket{
			{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
			{
				Art: "mse-style", Name: "nachbarvorgabe", Version: "2026-08-22",
				Pfad: "/p/nachbarvorgabe.mse-style",
			},
		}},
		&stilablage{werte: werte}, &feldliste{}, testdatenwurzel(t),
	)
}

// vorgabeVon sucht die Vorgabe einer Option.
func vorgabeVon(t *testing.T, optionen []service.Stiloption, name string) string {
	t.Helper()

	for _, option := range optionen {
		if option.Name == name {
			return option.Vorgabe
		}
	}

	t.Fatalf("die Option %q steht nicht in %+v", name, optionen)

	return ""
}

func TestEineVorgabeSiehtDenGespeichertenNachbarwert(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		werte    []domain.Value
		erwartet string
	}{
		// **Ungesetzt**: die Nachbaroption steht auf ihrer ersten Wahl.
		{"ungesetzt", nil, "schwarz"},

		// **Gesetzt**: und genau das sah die Umgebung bis zum 2026-08-22
		// nicht.
		{
			"gespeichert",
			[]domain.Value{{Field: "hintergrund", Text: "dunkel"}},
			"weiß",
		},

		// **Ein Wert, den der heutige Stil nicht mehr führt.** Ein Set kann
		// ihn tragen — jemand hat den Stil gewechselt —, und MSEs
		// `styling_data` hält ihn ebenso. Er darf die anderen nicht
		// verdrängen.
		{
			"ein Wert ohne Feld im Stil",
			[]domain.Value{
				{Field: "gibtsnichtmehr", Text: "irgendwas"},
				{Field: "hintergrund", Text: "dunkel"},
			},
			"weiß",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			optionen, err := nachbarwaehler(t, fall.werte).Optionen(t.Context(), 7)
			if err != nil {
				t.Fatalf("Optionen: %v", err)
			}

			if gegeben := vorgabeVon(t, optionen, "schriftfarbe"); gegeben != fall.erwartet {
				t.Errorf("Vorgabe %q, erwartet war %q", gegeben, fall.erwartet)
			}
		})
	}
}
