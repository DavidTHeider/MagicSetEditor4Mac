package service_test

// **Wo die Zeile nicht „dran" sagt** — UXR-F16
// (#2031).
//
// **Eine eigene Datei**, weil `partiendran_test.go` den Fall prüft, in dem
// jemand dran ist, und hier die Ränder stehen: der freie Modus, die anderen
// Stände, wer nicht am Tisch sitzt, wer ging, und was bei mehreren Sitzen
// eines Profils gilt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Im freien Modus ist nie jemand dran.**
//
// **Weil die Regelmaschine dort keine Priorität kennt**: die Partien aus G1
// bis G8 sind Sandkästen, in denen jeder jederzeit alles darf (§15, Punkt 4),
// und wer sich ihre Priorität auswählt, sieht einen Namen, den niemand
// vergeben hat. Zwei Wege führen dorthin — kein laufender Zug, und ein
// Profil, das die Priorität nicht durchsetzt. Beide tragen hier eine
// Priorität, die stehengeblieben sein könnte: sie darf nichts bedeuten.
func TestImFreienModusIstNieJemandDran(t *testing.T) {
	t.Parallel()

	ohneZug := eineZugpartie(t)
	ohneZug.Zug = domain.Zugstand{Prioritaet: "sitz-2"}

	abgewaehlt := eineZugpartie(t)
	abgewaehlt.Zug.Prioritaet = "sitz-2"
	abgewaehlt.Profil.Matrix = domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichPrioritaet},
	}.Wirksam(abgewaehlt.Profil.Matrix)

	for name, partie := range map[string]domain.Partie{
		"ohne laufenden Zug": ohneZug, "mit abgewählter Priorität": abgewaehlt,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zeile := zeileFuer(t, gebundenAn(partie, dasProfil, "sitz-2"), dasProfil)

			if zeile.Dran {
				t.Error("im freien Modus war jemand dran")
			}

			if zeile.EigenerSitz != "sitz-2" {
				t.Errorf("der eigene Sitz ist %q", zeile.EigenerSitz)
			}
		})
	}
}

// **Eine angehaltene oder beendete Partie wartet auf niemanden.**
func TestPausierteUndBeendetePartienWartenAufNiemanden(t *testing.T) {
	t.Parallel()

	for name, stand := range map[string]domain.Partiestand{
		"pausiert": domain.StandPausiert, "beendet": domain.StandBeendet,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := eineZugpartie(t)
			partie.Zug.Prioritaet = "sitz-2"
			partie.Stand = stand

			zeile := zeileFuer(t, gebundenAn(partie, dasProfil, "sitz-2"), dasProfil)

			if zeile.Dran || zeile.EigenerSitz != "sitz-2" {
				t.Errorf("%s: eigener Sitz %q, dran=%v", name, zeile.EigenerSitz, zeile.Dran)
			}
		})
	}
}

// **Wer nicht am Tisch sitzt, ist nie dran** — und hat keinen eigenen Sitz.
//
// **Drei Wege dahin**, und jeder trägt eine Partie, in der gerade Priorität
// vergeben ist: ein Sitz, der an ein anderes Profil gebunden ist; keiner
// gebunden (zwei Menschen am Küchentisch tippen zwei Namen); und die Null,
// die an jedem ungebundenen Sitz steht und nie einen findet.
func TestWerNichtAmTischSitztIstNieDran(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie(t)
	partie.Zug.Prioritaet = "sitz-2"

	for name, fall := range map[string]struct {
		partie domain.Partie
		profil int64
	}{
		"ein anderes Profil sitzt dort": {gebundenAn(partie, 9, "sitz-2"), dasProfil},
		"niemand ist gebunden":          {partie, dasProfil},
		"die Null findet keinen Sitz":   {partie, 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zeile := zeileFuer(t, fall.partie, fall.profil)

			if zeile.EigenerSitz != "" || zeile.Dran {
				t.Errorf("eigener Sitz %q, dran=%v", zeile.EigenerSitz, zeile.Dran)
			}
		})
	}
}

// **Wer gegangen ist, ist nicht mehr dran** — CR 800.4a (Fassung 2026-09-25).
//
// **Zu dritt**, weil zu zweit die Partie mit dem Gehen endet; und der Sitz
// bleibt der eigene — er hat dort gesessen, und die Zeile soll nicht tun, als
// sei er nie dagewesen.
func TestWerGegangenIstWirdNichtMehrGefragt(t *testing.T) {
	t.Parallel()

	partie := zugpartieIn(t, 3)
	partie.Zug.Prioritaet = "sitz-2"
	partie = partie.Scheidetaus("sitz-2", domain.GrundAufgabe)

	partie.Zug.Abwurf, partie.Zug.Aktiv = 1, "sitz-2"
	partie.Ausloeserreihung = "sitz-2"
	partie.Legendenwahlen = []domain.Legendenwahl{{Sitz: "sitz-2", Name: "Urza"}}

	zeile := zeileFuer(t, gebundenAn(partie, dasProfil, "sitz-2"), dasProfil)

	if zeile.Dran || zeile.EigenerSitz != "sitz-2" {
		t.Errorf("eigener Sitz %q, dran=%v", zeile.EigenerSitz, zeile.Dran)
	}
}

// **Mehrere Sitze eines Profils: der eigene ist der, der dran ist.**
//
// **Wer gegen sich selbst spielt**, bindet beide Sitze an sein Profil (G8). Die
// Zeile nennt dann den, an dem etwas zu tun ist — und, wo keiner dran ist,
// den ersten am Tisch.
func TestBeiMehrerenSitzenDesProfilsNenntDieZeileDenDerDranIst(t *testing.T) {
	t.Parallel()

	partie := gebundenAn(eineZugpartie(t), dasProfil, "sitz-1", "sitz-2")

	zweiter := partie
	zweiter.Zug.Prioritaet = "sitz-2"

	niemand := partie
	niemand.Zug.Prioritaet = ""

	for name, fall := range map[string]struct {
		partie domain.Partie
		sitz   string
		dran   bool
	}{
		"der erste hat die Priorität": {partie, "sitz-1", true},
		"der zweite hat sie":          {zweiter, "sitz-2", true},
		"keiner hat sie":              {niemand, "sitz-1", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zeile := zeileFuer(t, fall.partie, dasProfil)

			if zeile.EigenerSitz != fall.sitz || zeile.Dran != fall.dran {
				t.Errorf("eigener Sitz %q, dran=%v, erwartet %q und %v",
					zeile.EigenerSitz, zeile.Dran, fall.sitz, fall.dran)
			}
		})
	}
}

// **Jede Partie wird für sich gerechnet.**
//
// **Zwei Partien, zwei Antworten**: eine, in der man dran ist, und eine, in der
// man es nicht ist — in der Reihenfolge, in der der Speicher sie gibt. Ein Wert,
// der über die Schleife hinaus stehenbliebe, ließe die zweite wie die erste
// aussehen.
func TestJedePartieWirdFuerSichGerechnet(t *testing.T) {
	t.Parallel()

	dran := gebundenAn(eineZugpartie(t), dasProfil, "sitz-1")
	dran.ID = "p1"

	nichtDran := gebundenAn(eineZugpartie(t), dasProfil, "sitz-2")
	nichtDran.ID = "p2"

	probe := neueProbe()
	probe.partien["p1"], probe.partien["p2"] = dran, nichtDran

	zeilen, fehler := service.NeuePartieleitung(probe, nil).Partien(t.Context(), dasProfil)
	if fehler != nil {
		t.Fatalf("Partien: %v", fehler)
	}

	if len(zeilen) != 2 || !zeilen[0].Dran || zeilen[1].Dran ||
		zeilen[0].EigenerSitz != "sitz-1" || zeilen[1].EigenerSitz != "sitz-2" {
		t.Errorf("die Zeilen sind %+v", zeilen)
	}
}
