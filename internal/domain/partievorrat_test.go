package domain_test

// **Der eigene Vorrat** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Eine eigene Datei neben `partiesicht_test.go`**, weil jene über die
// Zeilengrenze aus §2 ging — und weil es eine eigene Frage ist: was die
// Projektion **hergibt**, damit das Fenster vorladen kann, ist etwas anderes
// als das, was sie **zurückhält**.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der eigene Vorrat steht in der eigenen Sicht — und nur dort.**
//
// # Wofür es ihn gibt
//
// **Damit die Bilder da sind, bevor die Karte kommt** (#1459: „Eigene Decks,
// Rückseite … vorladen"). Die Objekte in der Bibliothek tragen keine
// Kennung, und das bleibt so — ohne diese Liste gäbe es nichts zum Vorladen.
//
// # Und warum das die Bibliothek nicht verrät
//
// **Weil zwei verschiedene Dinge geheim sind.** Die **Reihenfolge** ist das
// Geheimnis; **welche Karten im eigenen Deck liegen**, hat der Eigner selbst
// aufgeschrieben. Die Liste ist deshalb sortiert und ohne Wiederholung: aus
// ihr folgt nichts über die Lage einer Karte.
func TestDerVorratStehtNurInDerEigenenSicht(t *testing.T) {
	t.Parallel()

	partie := einePartieMitVorrat()

	eigene := partie.Sicht("sitz1")
	if len(eigene.Vorrat) == 0 {
		t.Fatal("der eigene Sitz bekommt keinen Vorrat")
	}

	// **Sortiert und ohne Wiederholung** — vier Wälder sind ein Bild.
	if !slices.IsSorted(eigene.Vorrat) {
		t.Errorf("Vorrat %v ist nicht sortiert", eigene.Vorrat)
	}

	if len(slices.Compact(slices.Clone(eigene.Vorrat))) != len(eigene.Vorrat) {
		t.Errorf("Vorrat %v hat Wiederholungen", eigene.Vorrat)
	}

	// **Die Kennungen sind die des eigenen Decks** — 3 und 9, und die 7 des
	// Gegenübers steht nicht darin.
	if !slices.Equal(eigene.Vorrat, []int64{3, 9}) {
		t.Errorf("Vorrat %v, erwartet [3 9]", eigene.Vorrat)
	}

	// **Jeder Sitz bekommt seinen eigenen und keinen fremden.**
	if fremde := partie.Sicht("sitz2").Vorrat; !slices.Equal(fremde, []int64{7}) {
		t.Errorf("sitz2 bekommt %v, erwartet [7]", fremde)
	}

	// **Und wer nicht am Tisch sitzt, bekommt keinen.**
	for _, draussen := range []string{"zuschauer", ""} {
		if vorrat := partie.Sicht(draussen).Vorrat; len(vorrat) != 0 {
			t.Errorf("%q bekommt den Vorrat %v", draussen, vorrat)
		}
	}
}

// **Und er ist eine leere Liste und kein `nil`.**
//
// **Im Fenster stünde sonst `null`**, und jede Schleife darüber fiele um —
// derselbe Satz wie bei den Mischpflichten daneben.
func TestDerVorratIstNieNil(t *testing.T) {
	t.Parallel()

	if vorrat := einePartieMitVorrat().Sicht("wer auch immer").Vorrat; vorrat == nil {
		t.Error("nil statt einer leeren Liste")
	}
}

// einePartieMitVorrat gibt eine Partie, deren erster Sitz Karten mit Kennung
// führt — und deren zweiter keine.
func einePartieMitVorrat() domain.Partie {
	mitKennung := func(name string, id int64, menge int) domain.Abzugskarte {
		return domain.Abzugskarte{
			Kennung: name, Name: name, Menge: menge,
			Seiten: []domain.Abzugsseite{{Name: name, KarteID: id}},
		}
	}

	return domain.Partie{
		ID: "p1", Spiel: "magic", Stand: domain.StandLaeuft, Wirt: "sitz1",
		Sitze: []domain.Sitz{
			{
				Kennung: "sitz1", Name: "A",
				Deck: domain.Partiedeck{
					Abdruck: "a",
					Karten: []domain.Abzugskarte{
						mitKennung("Wald", 9, 4),
						mitKennung("Berg", 3, 2),
						// **Ein echter Druck trägt keine Kennung.**
						{
							Kennung: "d", Name: "Druck", Menge: 1,
							Seiten: []domain.Abzugsseite{{Name: "Druck"}},
						},
					},
				},
			},
			{
				Kennung: "sitz2", Name: "B",
				Deck: domain.Partiedeck{
					Abdruck: "b",
					Karten:  []domain.Abzugskarte{mitKennung("Insel", 7, 1)},
				},
			},
		},
		Objekte: []domain.Spielobjekt{},
	}
}
