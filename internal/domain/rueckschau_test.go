package domain_test

// **Was ein Teilnehmer später sehen darf** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Die Zusage, für die es diese Rechnung gibt
//
// Das Issue sagt sie zweimal: „kein Replay exportiert unbemerkt fremde private
// Informationen" und „Replay/Chat/Logs eines Teilnehmers enthalten nur seine
// freigegebenen Informationen."
//
// **Die Gefahr ist nicht der laufende Tisch, sondern das Protokoll danach.**
// Am Tisch rechnet `Partie.Sicht` je Sitz; ein Ereignisstrom liegt vollständig
// vor, und wer ihn ungeprüft ausgibt, verrät im Nachhinein, was während des
// Spiels niemand sehen durfte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func einStrom() []domain.Ereignis {
	return []domain.Ereignis{
		{
			PartieID:    "p1",
			Folge:       1,
			Art:         "objekt.ziehen",
			Oeffentlich: map[string]any{"sitz": "a", "zone": "hand"},
			Privat: []domain.Ergaenzung{
				{Empfaenger: "a", Inhalt: map[string]any{"karte": "Wald"}},
			},
			Kette: "glied-1",
		},
		{
			PartieID:    "p1",
			Folge:       2,
			Art:         "objekt.ziehen",
			Oeffentlich: map[string]any{"sitz": "b", "zone": "hand"},
			Privat: []domain.Ergaenzung{
				{Empfaenger: "b", Inhalt: map[string]any{"karte": "Berg"}},
			},
			Kette: "glied-2",
		},
	}
}

// **Jeder sieht jedes Ereignis** — die Reihenfolge ist öffentlich.
//
// Dass jemand gezogen hat, ist kein Geheimnis; **was** er gezogen hat, schon.
func TestJederSiehtJedesEreignis(t *testing.T) {
	t.Parallel()

	fuerA := domain.RueckschauFuer(einStrom(), "a", nil)

	if len(fuerA) != 2 {
		t.Fatalf("%d Ereignisse, erwartet 2", len(fuerA))
	}
}

// **Und nur seine eigenen Ergänzungen.**
func TestNurDieEigenenErgaenzungen(t *testing.T) {
	t.Parallel()

	fuerA := domain.RueckschauFuer(einStrom(), "a", nil)

	switch {
	case len(fuerA[0].Privat) != 1:
		t.Errorf("%d Ergänzungen am eigenen Zug, erwartet 1", len(fuerA[0].Privat))
	case len(fuerA[1].Privat) != 0:
		t.Errorf("%d Ergänzungen am fremden Zug, erwartet 0", len(fuerA[1].Privat))
	}
}

// **Die Gegenprobe: der andere sieht das Gegenteil.**
//
// Ohne sie sagte der Lauf darüber nur, dass gefiltert wird — nicht, dass
// **richtig** gefiltert wird (§10).
func TestDerAndereSiehtDasGegenteil(t *testing.T) {
	t.Parallel()

	fuerB := domain.RueckschauFuer(einStrom(), "b", nil)

	switch {
	case len(fuerB[0].Privat) != 0:
		t.Errorf("%d Ergänzungen am fremden Zug, erwartet 0", len(fuerB[0].Privat))
	case len(fuerB[1].Privat) != 1:
		t.Errorf("%d Ergänzungen am eigenen Zug, erwartet 1", len(fuerB[1].Privat))
	}
}

// **Ein Zuschauer sieht nichts Privates** — von niemandem.
func TestEinZuschauerSiehtNichtsPrivates(t *testing.T) {
	t.Parallel()

	for _, ereignis := range domain.RueckschauFuer(einStrom(), "zuschauer-1", nil) {
		if len(ereignis.Privat) != 0 {
			t.Errorf("Folge %d trägt %d Ergänzungen",
				ereignis.Folge, len(ereignis.Privat))
		}
	}
}

// **Und ein leerer Empfänger auch nicht.**
//
// Das ist der Fall, der beim Ausführen eines Protokolls entsteht, wenn
// niemand angegeben wurde — er darf nicht versehentlich alles zeigen.
func TestEinLeererEmpfaengerSiehtNichtsPrivates(t *testing.T) {
	t.Parallel()

	for _, ereignis := range domain.RueckschauFuer(einStrom(), "", nil) {
		if len(ereignis.Privat) != 0 {
			t.Errorf("Folge %d trägt %d Ergänzungen",
				ereignis.Folge, len(ereignis.Privat))
		}
	}
}

// **Die Kette bleibt** — sie ist öffentlich und der Beleg, dass nichts fehlt.
//
// Wer sie aus einer Rückschau striche, nähme dem Leser die Möglichkeit
// nachzurechnen, dass er einen vollständigen Verlauf vor sich hat.
func TestDieKetteBleibtInDerRueckschau(t *testing.T) {
	t.Parallel()

	fuerA := domain.RueckschauFuer(einStrom(), "a", nil)

	if fuerA[0].Kette != "glied-1" || fuerA[1].Kette != "glied-2" {
		t.Errorf("die Kette fehlt: %q, %q", fuerA[0].Kette, fuerA[1].Kette)
	}
}

// **Der Strom selbst bleibt unberührt.**
//
// Wer ihn beim Filtern änderte, hätte nach der ersten Rückschau nur noch
// diese eine — und die nächste Sicht sähe ein Protokoll, aus dem schon etwas
// entfernt wurde.
func TestDerStromBleibtUnberuehrt(t *testing.T) {
	t.Parallel()

	strom := einStrom()
	domain.RueckschauFuer(strom, "a", nil)

	if len(strom[1].Privat) != 1 {
		t.Errorf("der Strom hat %d Ergänzungen, erwartet 1", len(strom[1].Privat))
	}
}

// **Ein leerer Strom gibt einen leeren zurück** — und nicht `nil`, damit eine
// Fläche ihn ohne Fallunterscheidung ausgeben kann.
func TestEinLeererStromGibtEinenLeeren(t *testing.T) {
	t.Parallel()

	zurueck := domain.RueckschauFuer(nil, "a", nil)

	if zurueck == nil || len(zurueck) != 0 {
		t.Errorf("%v, erwartet eine leere Liste", zurueck)
	}
}
