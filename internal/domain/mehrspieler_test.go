package domain_test

// **Mehr als zwei am Tisch** — Werkstück M13 der Regelmaschine
// (#1615).
//
// # Was eine Mehrspielerpartie von einer zu zweit unterscheidet
//
// **CR 800.4**: sie läuft weiter, wenn jemand geht. Bei zweien ist mit dem
// Gehen alles vorbei; bei dreien geht es reihum weiter — **ohne den, der
// gegangen ist**.
//
// **Und das ist keine Feinheit, sondern der Kern**: eine Zugreihenfolge, die
// einen Gegangenen mitzählt, bleibt bei ihm stehen. Niemand hat Priorität,
// niemand kann passen, und die Partie hängt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineDreierpartie gibt einen Tisch mit drei Sitzen, im ersten Zug.
func eineDreierpartie(t *testing.T) domain.Partie {
	t.Helper()

	partie := domain.Partie{
		Spiel:  "magic",
		Profil: domain.Grundprofil("magic"),
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten", Fassung: domain.Grundprofilfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Wirt: "sitz-1",
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "Anna", Deck: einDeckabzug()},
			{Kennung: "sitz-2", Name: "Bert", Deck: einDeckabzug()},
			{Kennung: "sitz-3", Name: "Carla", Deck: einDeckabzug()},
		},
	}.MitStartleben()

	return partie.ZugBeginnen("sitz-1")
}

// **Die Zugreihenfolge geht reihum** — CR 102.1 mit CR 800.
func TestDieZugreihenfolgeGehtReihum(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)

	if partie.Zug.Aktiv != "sitz-1" {
		t.Fatalf("aktiv ist %q", partie.Zug.Aktiv)
	}

	zweiter := partie.ZugBeenden().Weiter()
	if zweiter.Zug.Aktiv != "sitz-2" {
		t.Errorf("nach dem ersten Zug ist %q dran", zweiter.Zug.Aktiv)
	}

	dritter := zweiter.ZugBeenden().Weiter()
	if dritter.Zug.Aktiv != "sitz-3" {
		t.Errorf("nach dem zweiten Zug ist %q dran", dritter.Zug.Aktiv)
	}
}

// **Wer gegangen ist, wird übersprungen** — CR 800.4.
//
// **Die härteste Stelle der Mehrspielerregeln**: eine Reihenfolge, die einen
// Gegangenen mitzählt, gibt ihm Priorität, und niemand kann sie ihm nehmen —
// die Partie hängt.
func TestWerGegangenIstWirdUebersprungen(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t).Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
	})

	danach := partie.ZugBeenden().Weiter()

	if danach.Zug.Aktiv != "sitz-3" {
		t.Errorf("nach sitz-1 ist %q dran — sitz-2 ist gegangen",
			danach.Zug.Aktiv)
	}
}

// **Und die Priorität ebenso** — CR 117.4 mit CR 800.4.
func TestDiePrioritaetUeberspringtDenGegangenen(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t).Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
	})

	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	if danach.Zug.Prioritaet != "sitz-3" {
		t.Errorf("nach sitz-1 hat %q Priorität", danach.Zug.Prioritaet)
	}
}

// **Alle passen nacheinander — auch zu dritt** — CR 117.4.
//
// **G9 hat es für einen und zwei belegt**; hier steht es für drei, und der
// Unterschied ist sichtbar: nach zwei Passen ist der Abschnitt **nicht**
// vorbei.
func TestAlleDreiPassenNacheinander(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	vorher := partie.Zug.Abschnitt

	erst, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("erstes passen: %v", fehler)
	}

	zweit, fehler := erst.Passen("sitz-2")
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	if zweit.Zug.Abschnitt != vorher {
		t.Fatalf("nach zwei Passen ist der Abschnitt schon %+v",
			zweit.Zug.Abschnitt)
	}

	dritt, fehler := zweit.Passen("sitz-3")
	if fehler != nil {
		t.Fatalf("drittes passen: %v", fehler)
	}

	if dritt.Zug.Abschnitt == vorher {
		t.Errorf("nach drei Passen steht der Abschnitt noch bei %+v",
			dritt.Zug.Abschnitt)
	}
}

// **Und ein Gegangener zählt beim Passen nicht mit** — CR 117.4 mit
// CR 800.4.
//
// **Sonst wartete der Abschnitt auf eine Antwort, die niemand geben kann.**
func TestEinGegangenerZaehltBeimPassenNichtMit(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t).Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
	})
	vorher := partie.Zug.Abschnitt

	erst, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("erstes passen: %v", fehler)
	}

	zweit, fehler := erst.Passen("sitz-3")
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	if zweit.Zug.Abschnitt == vorher {
		t.Errorf("zwei von drei haben nicht gereicht: %+v",
			zweit.Zug.Abschnitt)
	}
}

// **Kampfschaden an einen Gegangenen wird nicht zugeteilt** — CR 800.4e.
func TestKampfschadenAnEinenGegangenenFaelltAus(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	partie.Sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "baer", Name: "Bär", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"name": "Bär", "super type": "Creature",
				"power": "2", "toughness": "2",
			}}},
		}},
	}
	partie.Objekte = []domain.Spielobjekt{{
		Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "baer", Seit: 0,
	}}

	partie = partie.Scheidenaus([]domain.Ausgeschieden{
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
	})

	partie.Zug.Abschnitt = domain.Abschnitt{
		Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker,
	}
	partie.Zug.Prioritaet = "sitz-1"
	partie.Zug.Kampf = domain.Kampfstand{
		Erklaert: true, Geblockt: true,
		Angriffe: []domain.Angriff{
			{Kreatur: "sitz-1/baer/0", Ziel: "sitz-2"},
		},
	}

	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach, fehler = danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	zweiter := danach.SitzVon("sitz-2")
	if zweiter == nil {
		t.Fatal("den zweiten Sitz gibt es nicht")
	}

	if zweiter.Leben != 20 {
		t.Errorf("der Gegangene hat %d Leben — CR 800.4e teilt ihm keinen "+
			"Schaden zu", zweiter.Leben)
	}
}

// **Sind alle draußen, bleibt der erste Sitz** — §8.
//
// **Der Fall gibt es im Spiel nicht** (CR 104.2a beendet die Partie, sobald
// einer übrig ist), und er steht trotzdem hier: ein von Hand geänderter
// Bestand kann ihn herstellen, und eine leere Kennung ließe den Aufrufer
// raten.
func TestSindAlleDraussenBleibtDerErsteSitz(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t)
	partie.Ausgang.Ausgeschieden = []domain.Ausgeschieden{
		{Sitz: "sitz-1", Grund: domain.GrundAufgabe},
		{Sitz: "sitz-2", Grund: domain.GrundAufgabe},
		{Sitz: "sitz-3", Grund: domain.GrundAufgabe},
	}

	danach := partie.ZugBeenden().Weiter()

	if danach.Zug.Aktiv != "sitz-1" {
		t.Errorf("aktiv ist %q", danach.Zug.Aktiv)
	}
}
