package app_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Testkarten der Werkstatt kommen mit Kennung, Name und Nummer** —
// Werkstück U7 von Phase 5.1 (#1418). Der Name ist der Wert des Feldes
// `name`; wer keinen hat, kommt mit leerem Namen und seiner Nummer.
func TestDieProbekartenKommenAnsFenster(t *testing.T) {
	probe := &zeichnerprobe{karten: []*domain.CardRow{
		{ID: 7, SetID: 1, Number: 1, Values: []domain.Value{{Field: "name", Text: "Langtext"}}},
		{ID: 9, SetID: 1, Number: 2},
	}}
	anwendung := mitZeichner(probe)

	antwort := anwendung.Probekarten(t.Context(), "/pakete/probe.mse-style")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	erwartet := []app.ProbekarteDTO{
		{ID: "7", Name: "Langtext", Nummer: 1},
		{ID: "9", Name: "", Nummer: 2},
	}

	if !reflect.DeepEqual(antwort.Karten, erwartet) {
		t.Errorf("gefunden %+v, erwartet %+v", antwort.Karten, erwartet)
	}
}

// **Leer heißt leer, nicht `null`; ohne Zeichner und bei einem Fehler sagt
// es die Antwort.**
func TestDieProbekartenMeldenIhreFehler(t *testing.T) {
	if antwort := mitZeichner(&zeichnerprobe{}).Probekarten(t.Context(), "/p"); antwort.Karten == nil ||
		antwort.Fehler != nil {
		t.Errorf("ohne Karten: %+v", antwort)
	}

	if antwort := app.NeueApp("").Probekarten(t.Context(), "/p"); antwort.Fehler == nil {
		t.Error("ohne Zeichner kamen Karten")
	}

	kaputt := mitZeichner(&zeichnerprobe{fehler: errors.New("der Bestand ist zu")})

	if antwort := kaputt.Probekarten(t.Context(), "/p"); antwort.Fehler == nil {
		t.Error("der Fehler des Zeichners kam nicht an")
	}
}

// **Die Bindung geht denselben Weg.**
func TestDieProbekartenbindungGehtDenselbenWeg(t *testing.T) {
	anwendung := mitZeichner(&zeichnerprobe{karten: []*domain.CardRow{{ID: 7, SetID: 1}}})

	if antwort := app.NeueBindungen(anwendung).Probekarten(t.Context(), "/p"); len(antwort.Karten) != 1 {
		t.Errorf("%+v", antwort)
	}
}
