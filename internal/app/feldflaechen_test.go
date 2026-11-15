package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau sagt, welches Feld wo liegt**
// (#246).
//
// # Wofür
//
// AGENTS.md §6: **die Karte ist das Hauptobjekt.** Eine Vorschau, in der sich
// nichts anklicken lässt, ist eine Abbildung des Hauptobjekts, nicht das
// Hauptobjekt. Der Plan kennt die Rechtecke — bis
// #246 wusste
// er nicht, wem sie gehören.
//
// # In Bildpunkten und nicht in Kartenpunkten
//
// Das Frontend legt die Flächen über das PNG. Käme die Maßzahl aus dem Plan,
// müsste es die Vergrößerung selbst nachrechnen — und läge um genau den Faktor
// daneben, den es vergessen hat.
func TestDieVorschauNenntDieFeldflaechen(t *testing.T) {
	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	antwort := editor(&schreibprobe{}, karte).Vorschau(probekontext(t), "1", 1)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	flaechen := antwort.Vorschau.Flaechen
	if len(flaechen) == 0 {
		t.Fatal("keine Flächen")
	}

	fuerName := 0

	for _, flaeche := range flaechen {
		if flaeche.Feld == "" {
			t.Errorf("eine Fläche ohne Feld: %+v", flaeche)
		}

		if flaeche.Feld == "name" {
			fuerName++

			if flaeche.Breite <= 0 || flaeche.Hoehe <= 0 {
				t.Errorf("%+v", flaeche)
			}
		}
	}

	if fuerName == 0 {
		t.Error("keine Fläche für das Feld name")
	}
}

// **Die Vergrößerung geht in die Zahlen ein.**
//
// Wer sie im Frontend nachrechnete, läge um genau den Faktor daneben, den er
// vergessen hat — und die anklickbare Fläche säße neben dem, was man sieht.
func TestDieFeldflaechenFolgenDerVergroesserung(t *testing.T) {
	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	anwendung := editor(&schreibprobe{}, karte)

	einfach := anwendung.Vorschau(t.Context(), "1", 1)
	doppelt := anwendung.Vorschau(t.Context(), "1", 2)

	if einfach.Fehler != nil || doppelt.Fehler != nil {
		t.Fatalf("%+v %+v", einfach.Fehler, doppelt.Fehler)
	}

	if len(einfach.Vorschau.Flaechen) != len(doppelt.Vorschau.Flaechen) {
		t.Fatalf("%d gegen %d Flächen",
			len(einfach.Vorschau.Flaechen), len(doppelt.Vorschau.Flaechen))
	}

	erste, zweite := einfach.Vorschau.Flaechen[0], doppelt.Vorschau.Flaechen[0]
	if zweite.Breite != erste.Breite*2 || zweite.Hoehe != erste.Hoehe*2 {
		t.Errorf("%+v gegen %+v", erste, zweite)
	}
}
