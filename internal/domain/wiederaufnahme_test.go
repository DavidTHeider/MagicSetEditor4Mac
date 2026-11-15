package domain_test

// **Wie ein Gast zurückkommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// Das Issue nennt drei Wege und zwei Sperren: „kleine Lücke über Events,
// große über autorisierten Snapshot plus Folgestrom. Doppelte/geänderte IDs,
// Lücken, falsche Hashes und alte Epochs blockieren statt weiterzuspielen."
//
// **Die Entscheidung ist eine Rechnung** und keine Sache der Leitung: sie
// hängt an vier Zahlen und einem Hash, und sie muss auf beiden Seiten
// dieselbe sein.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const (
	ketteA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	ketteB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func einWunsch(teile ...func(*domain.Wiederaufnahmewunsch)) domain.Wiederaufnahmewunsch {
	wunsch := domain.Wiederaufnahmewunsch{
		Partie: "p1",
		Epoche: 2,
		Folge:  10,
		Kette:  ketteA,
	}

	for _, teil := range teile {
		teil(&wunsch)
	}

	return wunsch
}

func einStand(teile ...func(*domain.Wirtsstand)) domain.Wirtsstand {
	stand := domain.Wirtsstand{
		Epoche:       2,
		Folge:        10,
		Kette:        ketteA,
		AelteteFolge: 1,
	}

	for _, teil := range teile {
		teil(&stand)
	}

	return stand
}

// **Wer nichts verpasst hat, spielt weiter.**
func TestWerNichtsVerpasstHatSpieltWeiter(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(einWunsch(), einStand())

	if weg.Art != domain.WegWeiter {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegWeiter)
	}
}

// **Eine kleine Lücke wird über die Ereignisse geschlossen** — sie liegen
// noch vor, und der Gast bekommt sie nachgereicht.
func TestEineKleineLueckeGehtUeberDieEreignisse(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) { w.Folge = 7 }),
		einStand(),
	)

	switch {
	case weg.Art != domain.WegEreignisse:
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegEreignisse)
	case weg.AbFolge != 8:
		t.Errorf("ab Folge %d, erwartet 8", weg.AbFolge)
	}
}

// **Eine große geht über einen Abzug** — was der Gast verpasst hat, liegt
// nicht mehr vollständig vor.
//
// **Die Grenze ist nicht eine Zahl, sondern was der Wirt noch hat.** Wer sie
// an eine feste Zahl bände, bekäme bei einem kurzen Protokoll eine Lücke, die
// er für klein hält, und der Gast liefe in ein Loch.
func TestEineGrosseLueckeGehtUeberEinenAbzug(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) { w.Folge = 2 }),
		einStand(func(s *domain.Wirtsstand) { s.AelteteFolge = 6 }),
	)

	if weg.Art != domain.WegAbzug {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegAbzug)
	}
}

// **Eine alte Epoche kommt nicht zurück.**
func TestEineAlteEpocheBricht(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) { w.Epoche = 1 }),
		einStand(),
	)

	if weg.Art != domain.WegAbbruch {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegAbbruch)
	}
}

// **Eine neue Epoche beim Gast ist ein Widerspruch** — nur der Wirt zählt
// sie hoch, und ein Gast, der weiter ist als der Wirt, hat mit jemand anderem
// gespielt.
func TestEineNeuereEpocheAlsDerWirtBricht(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) { w.Epoche = 3 }),
		einStand(),
	)

	if weg.Art != domain.WegAbbruch {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegAbbruch)
	}
}

// **Ein falscher Hash bricht ab** — und zwar auch dann, wenn die Zahlen
// stimmen.
//
// Das ist der Fall, für den es den Hash überhaupt gibt: zwei Stände, die
// dieselbe Folge tragen und verschiedene Geschichten haben.
func TestEineFalscheKetteBricht(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) { w.Kette = ketteB }),
		einStand(),
	)

	if weg.Art != domain.WegAbbruch {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegAbbruch)
	}
}

// **Der Hash wird nur geprüft, wo er etwas belegen kann.**
//
// Bei einer Lücke kennt der Wirt das Glied des Gastes nicht mehr — er hat ein
// späteres. Geprüft wird dort die **Folge**, und die Kette trägt beim
// Nachreichen.
func TestBeiEinerLueckeWirdDieKetteNichtVerglichen(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) {
			w.Folge = 7
			w.Kette = ketteB
		}),
		einStand(),
	)

	if weg.Art != domain.WegEreignisse {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegEreignisse)
	}
}

// **Ein Gast, der weiter ist als der Wirt, bricht ab.**
func TestWerWeiterIstAlsDerWirtBricht(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) { w.Folge = 12 }),
		einStand(),
	)

	if weg.Art != domain.WegAbbruch {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegAbbruch)
	}
}

// **Ein Neuling bekommt einen Abzug** — er war noch nie dabei.
func TestEinNeulingBekommtEinenAbzug(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) {
			w.Folge = 0
			w.Kette = ""
		}),
		einStand(func(s *domain.Wirtsstand) { s.AelteteFolge = 5 }),
	)

	if weg.Art != domain.WegAbzug {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegAbzug)
	}
}

// **Und ein Neuling bei einer Partie, die noch nichts gesehen hat, spielt
// einfach mit** — es gibt nichts nachzureichen.
func TestEinNeulingAnEinerFrischenPartieSpieltMit(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) {
			w.Folge = 0
			w.Kette = ""
		}),
		einStand(func(s *domain.Wirtsstand) {
			s.Folge = 0
			s.Kette = ""
			s.AelteteFolge = 1
		}),
	)

	if weg.Art != domain.WegWeiter {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegWeiter)
	}
}

// **Jeder Weg sagt, warum** — eine Zahl allein hilft niemandem beim Lesen
// eines Protokolls.
func TestJederWegNenntSeinenGrund(t *testing.T) {
	t.Parallel()

	abbruch := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) { w.Kette = ketteB }),
		einStand(),
	)

	if strings.TrimSpace(abbruch.Grund) == "" {
		t.Error("ein Abbruch ohne Grund")
	}
}

// **Jede Art hat einen Namen** — für Meldungen und Protokolle.
func TestJedeWiederaufnahmeartHatEinenNamen(t *testing.T) {
	t.Parallel()

	gesehen := map[string]bool{}

	for _, art := range []domain.Wiederaufnahmeart{
		domain.WegWeiter, domain.WegEreignisse,
		domain.WegAbzug, domain.WegAbbruch,
	} {
		name := art.String()
		switch {
		case name == "" || name == "unbekannt":
			t.Errorf("%d hat keinen Namen", art)
		case gesehen[name]:
			t.Errorf("%q kommt zweimal vor", name)
		}

		gesehen[name] = true
	}

	// **Und was keine Art ist, sagt das auch** — statt eine leere Zeile ins
	// Protokoll zu schreiben.
	if domain.Wiederaufnahmeart(99).String() != "unbekannt" {
		t.Error("eine fremde Zahl gibt keinen Namen her")
	}
}

// **Ein Neuling bei einem Wirt in späterer Epoche bekommt trotzdem einen
// Abzug** — er hat keine Epoche zu vergleichen, nur einen leeren Stand.
func TestEinNeulingBeiSpaeterEpoche(t *testing.T) {
	t.Parallel()

	weg := domain.Wiederaufnahme(
		einWunsch(func(w *domain.Wiederaufnahmewunsch) {
			w.Epoche = 0
			w.Folge = 0
			w.Kette = ""
		}),
		einStand(func(s *domain.Wirtsstand) {
			s.Epoche = 9
			s.AelteteFolge = 4
		}),
	)

	if weg.Art != domain.WegAbzug {
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegAbzug)
	}
}
