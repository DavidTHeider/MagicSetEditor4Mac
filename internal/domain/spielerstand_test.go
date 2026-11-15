package domain_test

// **Was einen Sitz zum Spieler macht** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Regeln, nicht gegen eine Erwartung.** Jede Zahl hier steht in
// der CR, und die Nummer steht daneben — eine erfundene sähe aus wie ein
// Beleg (§11).
//
//   - **Leben** beginnt bei dem, was das Regelprofil sagt (CR 119.1 und
//     CR 103.4, Fassung 2026-08-19: gewöhnlich 20).
//   - **Vergiftungsmarken** beginnen bei null (CR 122.1f).
//   - **Markierter Schaden** liegt am Objekt und nicht am Spieler
//     (CR 120.3).
//   - Die **Handgrenze** ist eine Zahl dieses Regelwerks und kein Feld am
//     Sitz (CR 402.2) — warum, steht bei [domain.Handgrenze].

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Sitz fängt mit dem Leben seines Regelprofils an.**
func TestEinSitzFaengtMitDemLebenDesProfilsAn(t *testing.T) {
	t.Parallel()

	partie := einePartie()

	// **Aus dem eingefrorenen Profil und nicht aus einer Zahl hier**: das
	// Profil reist mit der Partie, damit ein Neustart dieselbe Antwort gibt
	// (G4, #1461).
	if gesehen := partie.Profil.Startleben; gesehen != 20 {
		t.Fatalf("das Grundprofil nennt %d Startleben", gesehen)
	}

	gesetzt := partie.MitStartleben()

	for _, sitz := range gesetzt.Sitze {
		if sitz.Leben != 20 {
			t.Errorf("der Sitz %q hat %d Leben", sitz.Kennung, sitz.Leben)
		}

		if sitz.Vergiftung != 0 {
			t.Errorf("der Sitz %q ist vergiftet: %d", sitz.Kennung, sitz.Vergiftung)
		}
	}
}

// **Ein Regelwerk ohne Leben gibt keines** — CR 119.1 gilt für Magic, nicht
// für jedes Spiel, das dieses Programm trägt.
func TestEinRegelwerkOhneLebenGibtKeines(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Profil.Startleben = 0

	for _, sitz := range partie.MitStartleben().Sitze {
		if sitz.Leben != 0 {
			t.Errorf("der Sitz %q hat %d Leben", sitz.Kennung, sitz.Leben)
		}
	}
}

// **Die Handgrenze ist sieben** — CR 402.2 (Fassung 2026-08-19).
func TestDieHandgrenzeIstSieben(t *testing.T) {
	t.Parallel()

	if domain.Handgrenze != 7 {
		t.Errorf("die Handgrenze ist %d", domain.Handgrenze)
	}

	// **Und die Starthand ist dieselbe Zahl aus einem anderen Grund**
	// (CR 103.5): sie steht eigens da, weil ein Effekt die eine ändern kann,
	// ohne die andere anzufassen.
	if domain.Starthand != 7 {
		t.Errorf("die Starthand ist %d", domain.Starthand)
	}
}

// **Negatives Leben ist erlaubt, negative Marken sind es nicht.**
//
// **Weil das zwei verschiedene Dinge sind**: ein Spieler bei −3 hat verloren
// (CR 704.5a) — das ist eine zustandsbasierte Aktion und kommt mit Phase 5.8;
// bis dahin steht die Zahl da, wie sie ist. **Marken unter null gibt es
// nicht**: CR 122.1 kennt keine negative Anzahl.
func TestNegativeMarkenWerdenAbgewiesen(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Sitze[0].Leben = -3

	if fehler := partie.Pruefe(); fehler != nil {
		t.Errorf("ein Spieler bei -3 Leben wurde abgewiesen: %v", fehler)
	}

	partie.Sitze[0].Vergiftung = -1

	if fehler := partie.Pruefe(); fehler == nil {
		t.Error("eine negative Vergiftungsmarke kam durch")
	}
}

// **Markierter Schaden liegt am Objekt** — CR 120.3 (Fassung 2026-08-19).
//
// **Und geprüft wird er dort**, nicht an der Partie: [Partie.Pruefe] sieht
// die Objekte gar nicht an — sie werden am Befehlsrand geprüft, weil dort
// jede Änderung entsteht. Eine zweite Prüfung wäre dieselbe Aussage an zwei
// Orten (§2).
//
// **Und nicht unter null**: „drei Schaden entfernen" von einem Objekt ohne
// Schaden ist kein negativer Stand.
func TestSchadenAmObjektWirdGeprueft(t *testing.T) {
	t.Parallel()

	objekt := domain.Spielobjekt{
		Kennung: "sitz-1/k1/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "wald", Schaden: 2,
	}

	if fehler := objekt.Pruefe(); fehler != nil {
		t.Errorf("zwei markierter Schaden wurden abgewiesen: %v", fehler)
	}

	objekt.Schaden = -1

	if fehler := objekt.Pruefe(); fehler == nil {
		t.Error("negativer Schaden kam durch")
	}
}
