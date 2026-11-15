package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stilwaehlerAmPruefstueck baut einen Wähler über `testdaten/probe-stil`.
func stilwaehlerAmPruefstueck(t *testing.T, werte *stilablage) *service.Stilwaehler {
	t.Helper()

	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
		{
			Art: "mse-style", Name: "probe-stil", Version: "2026-08-15",
			Pfad: "/p/probe-stil.mse-style",
		},
	}}

	return service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Game: "probe", Style: "probe-stil"},
		}},
		katalog, werte, &feldliste{}, testdatenwurzel(t),
	)
}

// **Eine Mehrfachwahl bietet ihre Wahlen an**
// (#637).
//
// **In der Reihenfolge der Erklärung** — an ihr hängt die Normalform, und die
// Kästchen sollen in derselben Folge stehen wie der gespeicherte Wert.
func TestEineMehrfachwahlBietetIhreWahlenAn(t *testing.T) {
	optionen, err := stilwaehlerAmPruefstueck(t, &stilablage{}).Optionen(t.Context(), 7)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	for _, option := range optionen {
		if option.Name != "rahmen" {
			continue
		}

		vergleicheWahlen(t, option.Moeglichkeiten, []domain.Wahlmoeglichkeit{
			{Wert: "nyx", Beschriftung: "nyx"},
			{Wert: "schnee", Beschriftung: "schnee"},
			{Wert: "fahrzeug", Beschriftung: "fahrzeug"},
		})

		// **Und das `initial:` kommt in Normalform an**, nicht in der
		// Reihenfolge der Datei: dort steht `fahrzeug, nyx`.
		if option.Wert != "nyx, fahrzeug" {
			t.Errorf("Wert %q, erwartet `nyx, fahrzeug`", option.Wert)
		}

		return
	}

	t.Fatal("die Stiloption `rahmen` fehlt")
}

// **Gespeichert wird die Normalform**, nicht die Reihenfolge des Anklickens.
//
// **ported.** `MultipleChoiceValue::update` ruft `normalForm` nach jedem
// Durchgang (`src/data/field/multiple_choice.cpp`, Zeilen 54–60, Fassung
// f9d9356 vom 2021-07-06); die Reihenfolge ist die der Erklärung
// (ebenda, Zeilen 82–113).
func TestEineMehrfachwahlWirdInNormalformGespeichert(t *testing.T) {
	ablageWerte := &stilablage{}

	if err := stilwaehlerAmPruefstueck(t, ablageWerte).
		OptionSetzen(t.Context(), 7, "rahmen", "fahrzeug, nyx"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if gespeichert := wertVon(ablageWerte, "rahmen"); gespeichert != "nyx, fahrzeug" {
		t.Errorf("gespeichert %q, erwartet `nyx, fahrzeug`", gespeichert)
	}
}

// **Ist nichts gewählt, gilt `empty choice:`** — und ohne Angabe die leere
// Zeichenkette (`multiple_choice.cpp`, Zeilen 110–112, dieselbe Fassung).
func TestEineLeereMehrfachwahlSchreibtIhrenLeernamen(t *testing.T) {
	for feld, erwartet := range map[string]string{
		"rahmen":  "",
		"leerbar": "nichts",
	} {
		t.Run(feld, func(t *testing.T) {
			ablageWerte := &stilablage{}

			if err := stilwaehlerAmPruefstueck(t, ablageWerte).
				OptionSetzen(t.Context(), 7, feld, ""); err != nil {
				t.Fatalf("setzen: %v", err)
			}

			if gespeichert := wertVon(ablageWerte, feld); gespeichert != erwartet {
				t.Errorf("gespeichert %q, erwartet %q", gespeichert, erwartet)
			}
		})
	}
}

// **Eine Wahl, die es nicht gibt, fällt weg** — sie kommt aus einer Datei, die
// eine andere Fassung der Vorlage kannte. Sie stehen zu lassen hieße, einen
// Wert zu führen, den kein Skript liest.
func TestEineUnbekannteWahlFaelltWeg(t *testing.T) {
	ablageWerte := &stilablage{}

	if err := stilwaehlerAmPruefstueck(t, ablageWerte).
		OptionSetzen(t.Context(), 7, "rahmen", "nyx, erfunden"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if gespeichert := wertVon(ablageWerte, "rahmen"); gespeichert != "nyx" {
		t.Errorf("gespeichert %q, erwartet `nyx`", gespeichert)
	}
}

// **Eine andere Art wird nicht angefasst** — ein Text mit Komma bleibt, was er
// ist.
func TestEineTextoptionWirdNichtNormalisiert(t *testing.T) {
	ablageWerte := &stilablage{}

	if err := stilwaehlerAmPruefstueck(t, ablageWerte).
		OptionSetzen(t.Context(), 7, "aufschrift", "b, a"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if gespeichert := wertVon(ablageWerte, "aufschrift"); gespeichert != "b, a" {
		t.Errorf("gespeichert %q, erwartet `b, a`", gespeichert)
	}
}

// wertVon holt einen gespeicherten Stilwert.
func wertVon(ablageWerte *stilablage, feld string) string {
	for _, wert := range ablageWerte.werte {
		if wert.Field == feld {
			return wert.Text
		}
	}

	return "<nicht gesetzt>"
}
