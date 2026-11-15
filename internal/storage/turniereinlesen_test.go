package storage_test

// **Ein Turnier samt Verlauf einlesen** — DOM-04.AK-03
// (#1699).
//
// **In einem Zug**: Turnier, Teilnehmende samt Liste, Runden, Ergebnisse
// und die erste Zeile des Protokolls. Scheitert ein Teil, entsteht nichts.
// Die Kennungen im Aufruf sind **vorläufig** — die Paarungen verweisen über
// sie auf ihre Teilnehmenden, und der Speicher vergibt die echten.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einzulesen ist ein gespieltes Turnier mit vorläufigen Kennungen.
func einzulesen() domain.Turnier {
	return domain.Turnier{
		Name: "Freitag", Spiel: "magic", Einzelspiel: true, Notiz: "Hinterzimmer",
		Abgeschlossen: "2026-09-20T20:00:00Z",
		Teilnehmer: []domain.Teilnehmer{
			{ID: 1, Name: "Anna", Deckabzug: []domain.Abzugsplatz{
				{Bezeichnung: "Blitz", Menge: 4, Bereich: domain.BereichHauptdeck},
			}},
			{ID: 2, Name: "Bert", Ausgestiegen: true},
			{ID: 3, Name: "Cleo"},
		},
		Runden: []domain.Runde{
			{Nummer: 1, Paarungen: []domain.Paarung{
				{A: 1, B: 2, Ergebnis: &domain.Ergebnis{SpieleA: 1}},
				{A: 3},
			}},
			{Nummer: 2, Paarungen: []domain.Paarung{
				{A: 1, B: 3, Ergebnis: &domain.Ergebnis{SpieleB: 1}},
			}},
		},
	}
}

func TestEinTurnierLaesstSichSamtVerlaufEinlesen(t *testing.T) {
	ablage := turnierablage(t)

	eingelesen, err := ablage.Einlesen(probekontext(t), einzulesen())
	if err != nil {
		t.Fatalf("einlesen: %v", err)
	}

	gelesen, err := ablage.Turnier(probekontext(t), eingelesen.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Name != "Freitag" || !gelesen.Einzelspiel || gelesen.Notiz != "Hinterzimmer" ||
		gelesen.Abgeschlossen != "2026-09-20T20:00:00Z" {
		t.Errorf("der Kopf: %+v", gelesen)
	}

	if len(gelesen.Teilnehmer) != 3 || !gelesen.Teilnehmer[1].Ausgestiegen ||
		len(gelesen.Teilnehmer[0].Deckabzug) != 1 || gelesen.Teilnehmer[0].Deckabzug[0].Menge != 4 {
		t.Fatalf("die Teilnehmenden: %+v", gelesen.Teilnehmer)
	}

	namen := map[int64]string{}
	for _, teilnehmer := range gelesen.Teilnehmer {
		namen[teilnehmer.ID] = teilnehmer.Name
	}

	if len(gelesen.Runden) != 2 || len(gelesen.Runden[0].Paarungen) != 2 {
		t.Fatalf("die Runden: %+v", gelesen.Runden)
	}

	erste := gelesen.Runden[0].Paarungen[0]
	if namen[erste.A] != "Anna" || namen[erste.B] != "Bert" || erste.Ergebnis == nil || erste.Ergebnis.SpieleA != 1 {
		t.Errorf("die erste Paarung: %+v", erste)
	}

	if bye := gelesen.Runden[0].Paarungen[1]; namen[bye.A] != "Cleo" || !bye.IstBye() {
		t.Errorf("der Bye: %+v", bye)
	}

	protokoll, err := ablage.Protokoll(probekontext(t), eingelesen.ID)
	if err != nil || len(protokoll) != 1 || protokoll[0].Vorgang != domain.TurnierEingelesen ||
		protokoll[0].Angaben != "Freitag" {
		t.Errorf("das Protokoll: %+v, %v", protokoll, err)
	}
}

// **Scheitert ein Teil, entsteht nichts** — eine Paarung, deren Seite es
// nicht gibt, bricht den ganzen Zug ab.
func TestEinFehlerhaftesEinlesenLaesstNichtsZurueck(t *testing.T) {
	ablage := turnierablage(t)
	kaputt := einzulesen()
	kaputt.Runden[1].Paarungen[0].B = 99

	if _, err := ablage.Einlesen(probekontext(t), kaputt); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet ErrValidation", err)
	}

	liste, err := ablage.Turniere(probekontext(t), "")
	if err != nil || len(liste) != 0 {
		t.Errorf("nach dem Abbruch stehen %d Turniere da (%v)", len(liste), err)
	}
}

// **Eine Lücke in der Rundenfolge bricht ab** — eine eingelesene zweite Runde
// ohne erste wäre ein Verlauf, der nie stattfand.
func TestEineLueckeImVerlaufBrichtDasEinlesenAb(t *testing.T) {
	ablage := turnierablage(t)
	luecke := einzulesen()
	luecke.Runden = luecke.Runden[1:]

	if _, err := ablage.Einlesen(probekontext(t), luecke); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fehler ist %v, erwartet ErrConflict", err)
	}

	if liste, err := ablage.Turniere(probekontext(t), ""); err != nil || len(liste) != 0 {
		t.Errorf("nach dem Abbruch stehen %d Turniere da (%v)", len(liste), err)
	}
}
