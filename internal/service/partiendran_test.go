package service_test

// **Die Partienzeile sagt, ob man dran ist** — UXR-F16
// (#2031).
//
// # Wogegen hier geprüft wird
//
// **Gegen echte Stände der Regelmaschine und nicht gegen Felder, die der Lauf
// selbst setzt** (§3): der Beginn und der Zug kommen aus dem Aufbau und dem
// Sandkasten, so wie der Betrieb sie hinterlässt. Nur was sich nicht in
// wenigen Schritten erreichen lässt — eine Abwurfpflicht, eine Legendenwahl,
// eine Reihung —, wird an den Zustand geschrieben.
//
// **Dran ist, wer etwas tun muss, damit es weitergeht**: im Beginn jeder Sitz,
// der noch einen Schritt hat (CR 103.5), danach der mit der Priorität
// (CR 117.1), der aktive Spieler mit offener Abwurfpflicht (CR 514.1), wer
// eine Legende wählen soll (CR 704.5j) und wer seine Fähigkeiten ordnen soll
// (CR 603.3b) — alle in der Fassung 2026-09-25.
//
// **Und wer nicht am Tisch sitzt, ist nie dran.** Ein Sitz gehört dem Profil,
// wenn er an es gebunden ist (G8, #1465);
// wo keiner gebunden ist, sieht das Profil zu.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dasProfil ist das Profil, das in diesen Läufen fragt.
const dasProfil int64 = 7

// gebundenAn bindet die genannten Sitze an ein Profil — auf einer Kopie.
func gebundenAn(partie domain.Partie, profil int64, sitze ...string) domain.Partie {
	gebunden := slices.Clone(partie.Sitze)

	for stelle := range gebunden {
		if slices.Contains(sitze, gebunden[stelle].Kennung) {
			gebunden[stelle].ProfilID = profil
		}
	}

	partie.Sitze = gebunden

	return partie
}

// zeileFuer liest die Liste über die Leitung und gibt die einzige Zeile.
func zeileFuer(t *testing.T, partie domain.Partie, profil int64) domain.Partiezeile {
	t.Helper()

	probe := neueProbe()
	probe.partien[partie.ID] = partie

	zeilen, fehler := service.NeuePartieleitung(probe, nil).Partien(t.Context(), profil)
	if fehler != nil {
		t.Fatalf("Partien: %v", fehler)
	}

	if len(zeilen) != 1 {
		t.Fatalf("es kamen %d Zeilen", len(zeilen))
	}

	return zeilen[0]
}

// imBeginn gibt die aufgebaute Partie, in der sitz-2 so weit ist, wie die
// Schritte sagen — sitz-1 hat noch nichts getan.
func imBeginn(t *testing.T, schritte ...string) domain.Partie {
	t.Helper()

	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})
	partie := eineAufgebautePartie(t)

	for _, art := range schritte {
		_, partie = schritt(t, kasten, partie, "sitz-2", art, nil)
	}

	return partie
}

// **Im Beginn ist dran, wer noch einen Schritt hat** — CR 103.5 (Fassung
// 2026-09-25).
//
// **Vom Aufbau an**: auch wer noch keine Starthand gezogen hat, hat den
// nächsten Schritt vor sich, und die Partie wartet auf ihn. Wer behalten hat,
// wartet selbst — dann ist er nicht dran, auch wenn es sein Sitz ist.
func TestImBeginnIstDranWerNochEinenSchrittHat(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		partie domain.Partie
		dran   bool
	}{
		"die Starthand steht aus": {imBeginn(t), true},
		"die Starthand liegt da, behalten ist offen": {
			imBeginn(t, service.BefehlStarthand), true,
		},
		"nach einem Mulligan ist das Unterlegen geschuldet": {
			imBeginn(t, service.BefehlStarthand, service.BefehlMulligan), true,
		},
		"behalten, und der andere ist noch dabei": {
			imBeginn(t, service.BefehlStarthand, service.BefehlBehalten), false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zeile := zeileFuer(t, gebundenAn(fall.partie, dasProfil, "sitz-2"), dasProfil)

			if zeile.EigenerSitz != "sitz-2" || zeile.Dran != fall.dran {
				t.Errorf("eigener Sitz %q, dran=%v, erwartet dran=%v",
					zeile.EigenerSitz, zeile.Dran, fall.dran)
			}
		})
	}
}

// **Nach dem Beginn ist dran, wer Priorität hat** — CR 117.1 (Fassung
// 2026-09-25) —, und sie wandert mit dem Passen.
func TestNachDemBeginnIstDranWerPrioritaetHat(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie(t)
	if partie.Zug.Prioritaet != "sitz-1" {
		t.Fatalf("die Priorität hat %q", partie.Zug.Prioritaet)
	}

	gepasst := partie
	_, gepasst = schritt(t, service.NeuerSandkasten(festeMischung{}), gepasst,
		"sitz-1", service.BefehlPassen, nil)

	if gepasst.Zug.Prioritaet != "sitz-2" {
		t.Fatalf("nach dem Passen hat sie %q", gepasst.Zug.Prioritaet)
	}

	for name, fall := range map[string]struct {
		partie domain.Partie
		sitz   string
		dran   bool
	}{
		"der Sitz mit der Priorität":     {partie, "sitz-1", true},
		"der andere hat sie nicht":       {partie, "sitz-2", false},
		"nach dem Passen der nächste":    {gepasst, "sitz-2", true},
		"nach dem Passen nicht mehr der": {gepasst, "sitz-1", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zeile := zeileFuer(t, gebundenAn(fall.partie, dasProfil, fall.sitz), dasProfil)

			if zeile.EigenerSitz != fall.sitz || zeile.Dran != fall.dran {
				t.Errorf("eigener Sitz %q, dran=%v, erwartet dran=%v",
					zeile.EigenerSitz, zeile.Dran, fall.dran)
			}
		})
	}
}

// **Offene Entscheidungen machen dran, auch ohne Priorität** — CR 514.1,
// CR 704.5j und CR 603.3b (Fassung 2026-09-25).
//
// **Der Abwurf trifft den aktiven Spieler und keinen sonst**; die Legendenwahl
// den Beherrscher; die Reihung den, den die Partie nennt. Jede der drei hält
// das Spiel an, bis sie beantwortet ist — und hat in dieser Zeit keine
// Priorität vergeben.
func TestOffeneEntscheidungenMachenDran(t *testing.T) {
	t.Parallel()

	ohnePrioritaet := func(aendern func(*domain.Partie)) domain.Partie {
		partie := eineZugpartie(t)
		partie.Zug.Prioritaet = ""
		aendern(&partie)

		return partie
	}

	for name, fall := range map[string]struct {
		partie domain.Partie
		dran   bool
	}{
		"eine Abwurfpflicht des aktiven Spielers": {
			ohnePrioritaet(func(p *domain.Partie) {
				p.Zug.Aktiv, p.Zug.Abwurf = "sitz-2", 2
			}), true,
		},
		"eine Abwurfpflicht des anderen": {
			ohnePrioritaet(func(p *domain.Partie) {
				p.Zug.Aktiv, p.Zug.Abwurf = "sitz-1", 2
			}), false,
		},
		"eine Legendenwahl": {
			ohnePrioritaet(func(p *domain.Partie) {
				p.Legendenwahlen = []domain.Legendenwahl{{Sitz: "sitz-2", Name: "Urza"}}
			}), true,
		},
		"eine Legendenwahl des anderen": {
			ohnePrioritaet(func(p *domain.Partie) {
				p.Legendenwahlen = []domain.Legendenwahl{{Sitz: "sitz-1", Name: "Urza"}}
			}), false,
		},
		"eine Reihung": {
			ohnePrioritaet(func(p *domain.Partie) { p.Ausloeserreihung = "sitz-2" }),
			true,
		},
		"nichts, und niemand hat Priorität": {
			ohnePrioritaet(func(*domain.Partie) {}), false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zeile := zeileFuer(t, gebundenAn(fall.partie, dasProfil, "sitz-2"), dasProfil)

			if zeile.Dran != fall.dran {
				t.Errorf("dran=%v, erwartet %v", zeile.Dran, fall.dran)
			}
		})
	}
}
