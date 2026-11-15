package domain_test

// **Wie eine Runde gepaart wird** — Werkstück 4 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Was nachgelesen ist und was nicht
//
// **Nachgelesen** (MTR §10.4): ohne andere Ankündigung gilt für ein
// Turnier das Schweizer Paarungsverfahren. Und Appendix C sagt, was ein
// Bye wert ist.
//
// **Nicht nachgelesen**: *wie* Swiss paart. Die MTR beschreibt den Algorithmus
// nicht — sie nennt ihn. Was hier steht, ist deshalb die übliche Gestalt, und
// jede Einzelheit davon ist eine **eigene** Entscheidung: nach Stand sortieren,
// benachbart paaren, eine Wiederholung durch Herunterpaaren vermeiden, und der
// Bye geht an den Letzten ohne Bye.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die erste Runde paart nach der Anmeldung.**
//
// Es gibt noch keinen Stand; jede andere Regel wäre erfunden. **Und sie ist
// wiederholbar**: eine Auslosung machte jeden Lauf zu einem anderen, und ein
// Turnier, das sich nach einem Absturz anders fortsetzt, ist keines.
func TestDieErsteRundePaartNachDerAnmeldung(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(4)

	plan := domain.Paarungsplan(turnier)

	if len(plan) != 2 {
		t.Fatalf("%d Paarungen, erwartet 2", len(plan))
	}

	if plan[0].A != 1 || plan[0].B != 2 || plan[1].A != 3 || plan[1].B != 4 {
		t.Errorf("gepaart wurde %+v", plan)
	}
}

// **Bei ungerader Zahl bekommt der Letzte den Bye.**
//
// **Eigene Regel** (Vermerk 0062, Entscheidung 8): die MTR sagt, was ein Bye
// wert ist, nicht, wer ihn bekommt. Gewählt ist der mit den wenigsten
// Matchpunkten, der noch keinen hatte.
func TestBeiUngeraderZahlBekommtDerLetzteDenBye(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(3)

	plan := domain.Paarungsplan(turnier)

	if len(plan) != 2 {
		t.Fatalf("%d Paarungen, erwartet 2", len(plan))
	}

	if !plan[1].IstBye() || plan[1].A != 3 {
		t.Errorf("der Bye ging an %+v, erwartet Teilnehmer 3", plan[1])
	}
}

// **Wer schon einen Bye hatte, bekommt keinen zweiten** — solange es einen
// gibt, der noch keinen hatte.
//
// # Warum dieser Aufbau so umständlich ist
//
// Weil ein einfacher nichts beweist. Der erste Anlauf gab dem Letzten im
// Stand ohnehin keinen früheren Bye — der Plan kam mit **und ohne** die Regel
// gleich heraus, und die Gegenprobe blieb grün.
//
// **Hier steht der Bye-Empfänger hinten.** Nach zwei Runden haben 3 und 2 je
// einen Bye gehabt und stehen unten; nur 1 hat noch keinen. Ohne die Regel
// bekäme 3 seinen zweiten.
func TestEinZweiterByeGehtAnJemandAnderen(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(3)

	runde(&turnier, spiel(1, 2, 2, 0), domain.Paarung{A: 3})
	runde(&turnier, spiel(1, 3, 2, 0), domain.Paarung{A: 2})

	hatte := map[int64]bool{2: true, 3: true}

	plan := domain.Paarungsplan(turnier)

	var bye int64

	for _, paarung := range plan {
		if paarung.IstBye() {
			bye = paarung.A
		}
	}

	if bye == 0 {
		t.Fatalf("kein Bye im Plan: %+v", plan)
	}

	if hatte[bye] {
		t.Errorf("der Bye ging an %d — der hatte schon einen; 1 noch nicht "+
			"(%+v)", bye, plan)
	}
}

// **Die zweite Runde paart nach dem Stand.**
func TestDieZweiteRundePaartNachDemStand(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(4)

	// A und C gewinnen, B und D verlieren.
	runde(&turnier, spiel(1, 2, 2, 0), spiel(3, 4, 2, 0))

	plan := domain.Paarungsplan(turnier)

	// **Die Sieger gegeneinander, die Verlierer gegeneinander.**
	if plan[0].A != 1 || plan[0].B != 3 {
		t.Errorf("die erste Paarung ist %+v, erwartet 1 gegen 3", plan[0])
	}

	if plan[1].A != 2 || plan[1].B != 4 {
		t.Errorf("die zweite Paarung ist %+v, erwartet 2 gegen 4", plan[1])
	}
}

// **Zwei, die schon gegeneinander gespielt haben, werden nicht erneut
// gepaart.**
//
// # Wie das geht
//
// **Durch Herunterpaaren.** Wer im Stand als Nächster käme und schon gespielt
// hat, wird übersprungen; genommen wird der Nächstbeste. Das ist die übliche
// Gestalt, und sie ist **eine eigene Entscheidung** — die MTR beschreibt sie
// nicht.
func TestEineWiederholungWirdVermieden(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(4)

	// Runde 1: 1 schlägt 2, 3 schlägt 4.
	runde(&turnier, spiel(1, 2, 2, 0), spiel(3, 4, 2, 0))
	// Runde 2: 1 schlägt 3, 2 schlägt 4. Stand: 1 (6), 2 (3), 3 (3), 4 (0).
	runde(&turnier, spiel(1, 3, 2, 0), spiel(2, 4, 2, 0))

	plan := domain.Paarungsplan(turnier)

	for _, paarung := range plan {
		schon := map[[2]int64]bool{
			{1, 2}: true, {2, 1}: true,
			{3, 4}: true, {4, 3}: true,
			{1, 3}: true, {3, 1}: true,
			{2, 4}: true, {4, 2}: true,
		}

		if schon[[2]int64{paarung.A, paarung.B}] {
			t.Errorf("%d gegen %d haben schon gespielt: %+v",
				paarung.A, paarung.B, plan)
		}
	}
}

// **Geht es nicht ohne Wiederholung, wird wiederholt.**
//
// Bei drei Teilnehmern und drei Runden hat jeder gegen jeden gespielt. Eine
// vierte Runde muss dann wiederholen — **eine Runde ohne Paarungen wäre die
// schlechtere Antwort**, denn sie sagt dem Turnierleiter nichts.
func TestOhneAusweichWirdWiederholt(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(2)

	runde(&turnier, spiel(1, 2, 2, 0))

	plan := domain.Paarungsplan(turnier)

	if len(plan) != 1 || plan[0].IstBye() {
		t.Fatalf("gepaart wurde %+v — erwartet eine Wiederholung", plan)
	}
}

// **Wer ausgestiegen ist, wird nicht mehr gepaart.**
func TestEinAusgestiegenerWirdNichtGepaart(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(3)
	turnier.Teilnehmer[2].Ausgestiegen = true

	plan := domain.Paarungsplan(turnier)

	for _, paarung := range plan {
		if paarung.A == 3 || paarung.B == 3 {
			t.Errorf("der Ausgestiegene wurde gepaart: %+v", plan)
		}
	}

	if len(plan) != 1 {
		t.Errorf("%d Paarungen, erwartet eine", len(plan))
	}
}

// **Bleibt einer übrig, weil die anderen ausgestiegen sind, bekommt er den
// Bye.**
func TestEinUebrigerBekommtDenBye(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(2)
	turnier.Teilnehmer[1].Ausgestiegen = true

	plan := domain.Paarungsplan(turnier)

	if len(plan) != 1 || !plan[0].IstBye() || plan[0].A != 1 {
		t.Errorf("gepaart wurde %+v — erwartet ein Bye für 1", plan)
	}
}

// **Niemand mehr da, keine Paarung.**
func TestOhneTeilnehmerGibtEsKeinePaarung(t *testing.T) {
	t.Parallel()

	if plan := domain.Paarungsplan(domain.Turnier{}); len(plan) != 0 {
		t.Errorf("gepaart wurde %+v — erwartet nichts", plan)
	}
}

// **Hatten alle schon einen Bye, bekommt ihn der Letzte im Stand.**
//
// # Warum es diesen Zweig gibt
//
// Weil eine Runde, die wegen eines Byes ausfiele, die schlechtere Antwort
// wäre: sie sagt dem Turnierleiter nichts. Bei drei Teilnehmern und drei
// Runden hat jeder einmal ausgesetzt — die vierte Runde braucht trotzdem
// einen Plan.
func TestHattenAlleEinenByeBekommtIhnDerLetzte(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(3)

	// Drei Runden, in denen reihum jeder einmal aussetzt.
	runde(&turnier, spiel(1, 2, 2, 0), domain.Paarung{A: 3})
	runde(&turnier, spiel(1, 3, 2, 0), domain.Paarung{A: 2})
	runde(&turnier, spiel(2, 3, 2, 0), domain.Paarung{A: 1})

	plan := domain.Paarungsplan(turnier)

	var byes int

	for _, paarung := range plan {
		if paarung.IstBye() {
			byes++
		}
	}

	if byes != 1 {
		t.Errorf("%d Byes im Plan, erwartet einen: %+v", byes, plan)
	}

	// **Und es ist der Letzte im Stand.** 1 hat sechs Punkte, 2 hat drei,
	// 3 hat drei — 3 steht hinten, weil 2 ihn geschlagen hat.
	if letzte := plan[len(plan)-1]; !letzte.IstBye() || letzte.A != 3 {
		t.Errorf("der Bye ging an %+v, erwartet Teilnehmer 3", letzte)
	}
}

// **Eine Wiederholung, die sich vermeiden ließe, wird vermieden — auch wenn
// sie erst weiter hinten aufträte.**
//
// # Der Befund
//
// Der erste Algorithmus nahm den Ersten und für ihn den nächstbesten Gegner,
// gegen den er noch nicht gespielt hatte. Er sah nicht, was diese Wahl für
// die **Übrigen** bedeutet.
//
// **Bei der Rangfolge 1, 2, 3, 4, wenn nur 3–4 schon gespielt hat**, wählt
// die kurzsichtige Suche 1–2 — und muss danach 3–4 wiederholen. 1–3 und 2–4
// wären ganz ohne Wiederholung gegangen.
//
// Angestrichen von Befund 5 des Zweitblicks über Phase 4.7. §10.4 der MTR
// nennt Swiss und beschreibt es nicht; dass keine vermeidbare Wiederholung
// entsteht, ist deshalb eine **eigene** Zusage — und eine, die eingelöst
// werden muss, wenn sie dasteht.
func TestEineVermeidbareWiederholungWeiterHintenWirdVermieden(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(4)

	// **Der Aufbau, der den Fall wirklich herstellt.**
	//
	// Der erste Anlauf gab einfach 3 gegen 4 ein Spiel — und der Test blieb
	// grün: der Sieger stieg im Stand nach oben, und die Rangfolge war nicht
	// mehr 1, 2, 3, 4. Ein Fall, den die Rangliste umsortiert, ist keiner.
	//
	// **Hier bleiben 3 und 4 unten und beieinander**: ihr Spiel endet
	// unentschieden, und 1 und 2 stehen über ihnen, weil ein Bye drei Punkte
	// gibt (Appendix C). Damit steht die Rangfolge 1, 2, 3, 4 fest, und
	// gespielt hat nur 3 gegen 4.
	runde(&turnier, domain.Paarung{A: 1})
	runde(&turnier, domain.Paarung{A: 2})
	runde(&turnier, domain.Paarung{
		A: 3, B: 4,
		Ergebnis: &domain.Ergebnis{Unentschieden: 1},
	})

	// **Erst die Bedingung, dann die Zusage.** Ohne sie prüfte der Lauf eine
	// Rangfolge, die es nicht gibt.
	rangliste := domain.Rangliste(turnier)
	stand := make([]int64, 0, len(rangliste))

	for _, rang := range rangliste {
		stand = append(stand, rang.TeilnehmerID)
	}

	if len(stand) != 4 || stand[0] != 1 || stand[1] != 2 ||
		stand[2] != 3 || stand[3] != 4 {
		t.Fatalf("die Rangfolge ist %v — der Fall trägt nicht", stand)
	}

	plan := domain.Paarungsplan(turnier)

	for _, paarung := range plan {
		if (paarung.A == 3 && paarung.B == 4) || (paarung.A == 4 && paarung.B == 3) {
			t.Errorf("3 gegen 4 wird wiederholt, obwohl 1–3 und 2–4 gingen: %+v",
				plan)
		}
	}
}

// **Und wenn es wirklich nicht anders geht, wird wiederholt.**
//
// Die Gegenprobe zum Test darüber: eine Suche, die auf eine vollständige
// Lösung besteht, gäbe bei drei Teilnehmern nach drei Runden **keine**
// Paarung — und eine Runde ohne Tische sagt dem Turnierleiter nichts.
func TestOhneVollstaendigeLoesungWirdTrotzdemGepaart(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(4)

	// Jeder gegen jeden: danach ist jede Paarung eine Wiederholung.
	runde(&turnier, spiel(1, 2, 2, 0), spiel(3, 4, 2, 0))
	runde(&turnier, spiel(1, 3, 2, 0), spiel(2, 4, 2, 0))
	runde(&turnier, spiel(1, 4, 2, 0), spiel(2, 3, 2, 0))

	plan := domain.Paarungsplan(turnier)

	if len(plan) != 2 {
		t.Errorf("%d Paarungen, erwartet 2 — auch eine Wiederholung ist eine "+
			"Runde", len(plan))
	}
}
