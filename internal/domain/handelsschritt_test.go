package domain_test

// **Die Schritte eines Handels** — Werkstück `T1` von Phase 5.6
// (#1466),
// fortgeschrieben in `T6`
// (#1471).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein zweiter Beschluss sammelt eigene Stimmen.**
//
// # Der Fund, für den dieser Lauf steht
//
// **Der Weg endete vor dem letzten Schritt.** Die Matrix kennt **zwei**
// Schritte, die alle Parteien brauchen — `zustimmen` und `bestaetigen` —, und
// die Stimmen standen in **einem** Feld. Wer zugestimmt hatte, stand darin;
// beim Bestätigen fand `mitStimme` ihn wieder, hielt es für eine Wiederholung
// und gab denselben Stand zurück. `beidseitigVerbucht` war damit **nicht
// erreichbar** — angestrichen von `T6`, dem ersten Werkstück, das den Weg zu
// Ende geht (§10: unbetretene Zweige sind Zusagen, die niemand hält).
func TestEinZweiterBeschlussSammeltEigeneStimmen(t *testing.T) {
	t.Parallel()

	angebot := domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-1", Revision: 1,
		Art: domain.HandelTausch, Initiator: "aich",
		Parteien: []string{"aich", "bdu"}, Zustand: domain.HandelAngeboten,
	}

	// **Beide stimmen zu** — danach steht der Vorgang auf `angenommen`.
	for _, wer := range angebot.Parteien {
		geschritten, fehler := angebot.Schreite(domain.SchrittZustimmen, wer)
		if fehler != nil {
			t.Fatalf("zustimmen durch %s: %v", wer, fehler)
		}

		angebot = geschritten
	}

	if angebot.Zustand != domain.HandelAngenommen {
		t.Fatalf("der Vorgang steht auf %q", angebot.Zustand)
	}

	// **Und der Beschluss ist danach verbraucht.** Eine Stimme, die stehen
	// bliebe, zählte beim nächsten Beschluss mit — und der wäre entschieden,
	// bevor jemand etwas gesagt hat.
	if len(angebot.Zustimmungen) != 0 {
		t.Errorf("nach dem Wechsel stehen noch %d Stimmen da: %v",
			len(angebot.Zustimmungen), angebot.Zustimmungen)
	}

	// **Bis zum Abgleich, wo der zweite Beschluss ansteht.**
	//
	// **Die Übergabe ist selbst ein Beschluss** — beide Parteien bestätigen
	// sie, und deshalb steht sie hier mit zwei Stimmen.
	for _, schritt := range []domain.Handelsschritt{
		domain.SchrittVorbereiten, domain.SchrittUebergeben,
		domain.SchrittUebergeben, domain.SchrittVerbuchen,
		domain.SchrittAbgleichen,
	} {
		wer := "aich"
		if angebot.Zustand == domain.HandelVorbereitet &&
			len(angebot.Zustimmungen) == 1 {
			wer = "bdu"
		}

		geschritten, fehler := angebot.Schreite(schritt, wer)
		if fehler != nil {
			t.Fatalf("%s durch %s: %v", schritt, wer, fehler)
		}

		angebot = geschritten
	}

	if angebot.Zustand != domain.HandelAbgleichOffen {
		t.Fatalf("der Vorgang steht auf %q", angebot.Zustand)
	}

	// **Die erste Bestätigung wechselt noch nicht.**
	erste, fehler := angebot.Schreite(domain.SchrittBestaetigen, "aich")
	if fehler != nil {
		t.Fatalf("erste Bestätigung: %v", fehler)
	}

	if erste.Zustand != domain.HandelAbgleichOffen {
		t.Errorf("eine Stimme genügte: %q", erste.Zustand)
	}

	// **Die zweite schon** — und das ist die Stelle, die vorher nie erreicht
	// wurde.
	zweite, fehler := erste.Schreite(domain.SchrittBestaetigen, "bdu")
	if fehler != nil {
		t.Fatalf("zweite Bestätigung: %v", fehler)
	}

	if zweite.Zustand != domain.HandelBeidseitigVerbucht {
		t.Errorf("der Vorgang steht auf %q", zweite.Zustand)
	}
}

// **Eine Übergabe bestätigen beide.**
//
// # Warum das ein Beschluss ist und kein Schritt
//
// **Weil eine Übergabe zwei Seiten hat.** Das Issue zu `T6`
// (#1471)
// verlangt vor dem Abschluss „lokale **und empfangene**
// Übergabebestätigung"; solange eine fehlt, hat einer Karten weggegeben und
// der andere weiß nichts davon. Ein Zustand, den eine Seite allein herstellt,
// hieße „übergeben" und meinte „ich sage, dass ich übergeben habe".
//
// **Bis zum 2026-09-18 genügte eine Stimme** — der Übergang stand als
// gewöhnlicher Schritt in der Matrix. Angestrichen beim Bau von `T6`, dem
// ersten Werkstück, das aus diesem Zustand heraus etwas tut.
func TestEineUebergabeBestaetigenBeide(t *testing.T) {
	t.Parallel()

	angebot := domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-1", Revision: 1,
		Art: domain.HandelTausch, Initiator: "aich",
		Parteien: []string{"aich", "bdu"}, Zustand: domain.HandelVorbereitet,
	}

	// **Eine Stimme wechselt nicht.**
	erste, fehler := angebot.Schreite(domain.SchrittUebergeben, "aich")
	if fehler != nil {
		t.Fatalf("erste Übergabe: %v", fehler)
	}

	if erste.Zustand != domain.HandelVorbereitet {
		t.Errorf("eine Stimme genügte: %q", erste.Zustand)
	}

	// **Und dieselbe noch einmal auch nicht** (§7, Regel 4).
	nochmal, fehler := erste.Schreite(domain.SchrittUebergeben, "aich")
	if fehler != nil {
		t.Fatalf("Wiederholung: %v", fehler)
	}

	if nochmal.Zustand != domain.HandelVorbereitet {
		t.Errorf("die Wiederholung wechselte: %q", nochmal.Zustand)
	}

	// **Erst die zweite.**
	zweite, fehler := erste.Schreite(domain.SchrittUebergeben, "bdu")
	if fehler != nil {
		t.Fatalf("zweite Übergabe: %v", fehler)
	}

	if zweite.Zustand != domain.HandelUebergabeOffen {
		t.Errorf("der Vorgang steht auf %q", zweite.Zustand)
	}
}
