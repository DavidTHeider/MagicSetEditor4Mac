package service_test

// **Aufgeben** — Werkstück M17 der Regelmaschine
// (#1618).
//
// **Eine eigene Datei**, weil `zugbefehle_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge — und weil der Schnitt der Frage folgt: dort geht
// es um den Zug, hier um sein **Ende**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Aufgeben geht jederzeit und beendet die Partie** — CR 104.3a, CR 104.2a
// (M17, #1618).
func TestAufgebenBeendetDiePartie(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	// **Der Sitz ohne Priorität gibt auf** — „jederzeit" heißt: ohne zu
	// warten. Der Lauf prüft damit beides auf einmal.
	ereignisse, danach, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
			Art: service.BefehlAufgeben,
		})
	if fehler != nil {
		t.Fatalf("aufgeben: %v", fehler)
	}

	switch {
	case danach.Ausgang.Sieger != "sitz-1":
		t.Errorf("Sieger ist %q", danach.Ausgang.Sieger)
	case danach.Stand != domain.StandBeendet:
		t.Errorf("der Stand ist %q", danach.Stand)
	case ereignisse[0].Oeffentlich["sieger"] != "sitz-1":
		t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
	}
}

// **Wer nicht am Tisch sitzt, gibt nicht auf.**
func TestEinZuschauerGibtNichtAuf(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "z1", Kennung: "b1",
		Art: service.BefehlAufgeben,
	})

	if fehler == nil {
		t.Error("ein Zuschauer hat aufgegeben")
	}
}

// **Wer auf null Leben fällt, verliert** — CR 704.5a mit CR 104.3b, geprüft
// von [domain.Partie.Zustandspruefung] (M17, #1618).
//
// **Über den Befehl und nicht über die Domäne unmittelbar**: die Abnahme des
// Issues lautet „ein Spieler bei 0 Leben verliert, der andere gewinnt, und
// der Grund steht am Tisch" — das ist eine Aussage über den **Weg**.
func TestWerAufNullLebenFaelltVerliert(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	ereignisse, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlLeben,
		Nutzlast: map[string]any{"um": float64(-partie.Sitze[1].Leben)},
	})
	if fehler != nil {
		t.Fatalf("leben: %v", fehler)
	}

	switch {
	case danach.Stand != domain.StandBeendet:
		t.Errorf("der Stand ist %q", danach.Stand)
	case danach.Ausgang.Sieger != "sitz-1":
		t.Errorf("Sieger ist %q", danach.Ausgang.Sieger)
	case len(danach.Ausgang.Ausgeschieden) != 1 ||
		danach.Ausgang.Ausgeschieden[0].Grund != domain.GrundLeben:
		t.Errorf("ausgeschieden: %+v", danach.Ausgang.Ausgeschieden)
	case len(ereignisse) != 1:
		t.Errorf("%d Ereignisse statt einem", len(ereignisse))
	}
}

// **Und ein Verlust, der keiner ist, beendet nichts.**
//
// Die Gegenprobe zum Lauf darüber: derselbe Weg, ein Leben weniger als nötig.
func TestEinLebenMehrBeendetNichts(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlLeben,
		Nutzlast: map[string]any{"um": float64(1 - partie.Sitze[1].Leben)},
	})
	if fehler != nil {
		t.Fatalf("leben: %v", fehler)
	}

	if danach.Stand == domain.StandBeendet {
		t.Errorf("bei einem Leben ist die Partie beendet")
	}
}

// **Zehn Vergiftungsmarken sind zehn** — CR 704.5c mit CR 104.3d.
func TestZehnVergiftungsmarkenBeendenDiePartie(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, danach, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlVergiftung,
		Nutzlast: map[string]any{"um": float64(domain.Vergiftungsgrenze)},
	})
	if fehler != nil {
		t.Fatalf("vergiftung: %v", fehler)
	}

	if len(danach.Ausgang.Ausgeschieden) != 1 ||
		danach.Ausgang.Ausgeschieden[0].Grund != domain.GrundGift {
		t.Fatalf("ausgeschieden: %+v", danach.Ausgang.Ausgeschieden)
	}
}
