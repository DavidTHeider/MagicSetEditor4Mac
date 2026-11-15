package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestSpielmodellLiestFelderInIhrerReihenfolge(t *testing.T) {
	spiel, _ := oeffneTestpaket(t, "probe.mse-game").Spielmodell()

	if spiel.Name != "Probe" {
		t.Errorf("Spielname %q", spiel.Name)
	}

	erwartet := []struct {
		name string
		art  domain.FieldKind
	}{
		{"name", domain.KindText},
		{"farbe", domain.KindChoice},
		{"bild", domain.KindImage},
		{"legendaer", domain.KindBoolean},
		{"seltsam", domain.KindUnknown},
		{"ohne art", domain.KindUnknown},
	}

	if len(spiel.Fields) != len(erwartet) {
		t.Fatalf("%d Felder, erwartet waren %d", len(spiel.Fields), len(erwartet))
	}

	for i, fall := range erwartet {
		if spiel.Fields[i].Name != fall.name || spiel.Fields[i].Kind != fall.art {
			t.Errorf("Feld %d ist %+v, erwartet war %s/%s",
				i+1, spiel.Fields[i], fall.name, fall.art)
		}
	}
}

// Eine unbekannte Feldart wird zu KindUnknown, nicht zu einem Fehler.
//
// Ein Spiel abzuweisen, weil es ein Feld mehr hat als wir kennen, hieße: jedes
// selbst erweiterte Spiel bleibt zu. Als Text behandelt bleibt der Wert
// erhalten und der Rundlauf heil.
// Ein Kartenfeld ganz ohne `type:` ist dasselbe wie eines mit unbekannter Art.
//
// Beides landet bei KindUnknown und wird als Text behandelt. Sie zu
// unterscheiden brächte nichts: in beiden Fällen wissen wir nur, dass da ein
// Wert steht.
func TestEinFeldOhneArtWirdEbenfallsText(t *testing.T) {
	spiel, _ := oeffneTestpaket(t, "probe.mse-game").Spielmodell()

	ohneArt, bekannt := spiel.Field("ohne art")
	if !bekannt {
		t.Fatal("das Feld ohne Art fehlt")
	}
	if ohneArt.Kind != domain.KindUnknown {
		t.Errorf("Art %q, erwartet war KindUnknown", ohneArt.Kind)
	}
}

func TestEineUnbekannteFeldartWirdText(t *testing.T) {
	spiel, befunde := oeffneTestpaket(t, "probe.mse-game").Spielmodell()

	seltsam, bekannt := spiel.Field("seltsam")
	if !bekannt {
		t.Fatal("das Feld mit unbekannter Art fehlt")
	}
	if seltsam.Kind != domain.KindUnknown {
		t.Errorf("Art %q, erwartet war KindUnknown", seltsam.Kind)
	}

	for _, befund := range befunde {
		if befund.Meldung == "" {
			t.Error("ein Befund ohne Meldung")
		}
	}
}

func TestEinKartenfeldOhneNamenIstEinBefund(t *testing.T) {
	_, befunde := oeffneTestpaket(t, "probe.mse-game").Spielmodell()

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer", len(befunde))
	}
	if befunde[0].Stelle.Zeile == 0 {
		t.Error("der Befund nennt keine Zeile")
	}
}

func TestAuswahlmoeglichkeitenUndStandardwertWerdenGelesen(t *testing.T) {
	spiel, _ := oeffneTestpaket(t, "probe.mse-game").Spielmodell()

	farbe, bekannt := spiel.Field("farbe")
	if !bekannt {
		t.Fatal("das Feld farbe fehlt")
	}
	if len(farbe.Choices) != 2 || farbe.Choices[0] != "rot" || farbe.Choices[1] != "blau" {
		t.Errorf("Möglichkeiten %v", farbe.Choices)
	}
	if farbe.Default != "rot" {
		t.Errorf("Standardwert %q", farbe.Default)
	}
}

// Das gelesene Spiel besteht seine eigene Prüfung.
//
// Ein Modell, das die Domäne selbst abweisen würde, wäre kein Modell — es wäre
// eine Sammlung von Werten, die zufällig dieselben Feldnamen trägt.
func TestDasGeleseneSpielBestehtSeineEigenePruefung(t *testing.T) {
	spiel, _ := oeffneTestpaket(t, "probe.mse-game").Spielmodell()

	if err := spiel.Validate(); err != nil {
		t.Errorf("das gelesene Spiel besteht seine eigene Prüfung nicht: %v", err)
	}
}
