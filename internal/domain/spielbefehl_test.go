package domain_test

// **Befehle, Ereignisse und die Kette dazwischen** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// # Wonach hier gefragt wird
//
// Danach, dass **die Folge maßgeblich ist und der Zeitstempel ein Metadatum**,
// dass ein Abdruck sich nicht ändert, wenn man Felder anders schreibt, und
// dass die **private Ergänzung einen eigenen Nachweis** hat — sonst zwänge ein
// gemeinsamer Vollzustandshash dazu, fremde Hände offenzulegen.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func einBefehl() domain.Spielbefehl {
	return domain.Spielbefehl{
		Kennung:  "b1",
		PartieID: "p1",
		Absender: "sitz-1",
		Art:      "karte.ziehen",
		Erwartet: 3,
		Nutzlast: map[string]any{"menge": float64(1)},
		Gestellt: time.Date(2026, time.September, 15, 12, 0, 0, 0, time.UTC),
	}
}

// **Der kleinste Befehl geht durch.**
func TestDerKleinsteBefehlGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := einBefehl().Pruefe(); fehler != nil {
		t.Fatalf("der kleinste Befehl: %v", fehler)
	}
}

// **Was kein Befehl ist.**
func TestWasKeinBefehlIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Spielbefehl){
		"ohne Kennung":  func(b *domain.Spielbefehl) { b.Kennung = " " },
		"ohne Partie":   func(b *domain.Spielbefehl) { b.PartieID = "" },
		"ohne Absender": func(b *domain.Spielbefehl) { b.Absender = "" },
		"ohne Art":      func(b *domain.Spielbefehl) { b.Art = "" },

		// **Eine Fassung, die es nie gab.** Die erwartete Fassung ist der
		// Schutz gegen die doppelte Wirkung; eine negative wäre keiner.
		"mit negativer Erwartung": func(b *domain.Spielbefehl) {
			b.Erwartet = -1
		},
	}

	for name, verderben := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			befehl := einBefehl()
			verderben(&befehl)

			if fehler := befehl.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Derselbe Abdruck, gleich wie die Felder geschrieben stehen.**
func TestDerAbdruckHaengtAmInhaltUndNichtAmSchreiben(t *testing.T) {
	t.Parallel()

	erstes := domain.Ereignis{
		PartieID: "p1", Folge: 1, Art: "karte.gezogen",
		Oeffentlich: map[string]any{"sitz": "sitz-1", "menge": float64(1)},
	}

	zweites := domain.Ereignis{
		PartieID: "p1", Folge: 1, Art: "karte.gezogen",
		Oeffentlich: map[string]any{"menge": float64(1), "sitz": "sitz-1"},
	}

	if abdruck(t, erstes, "") != abdruck(t, zweites, "") {
		t.Error("die Reihenfolge der Felder ändert den Abdruck")
	}

	// **Und der Zeitstempel ändert ihn nicht** — die Folge ist maßgeblich,
	// die Uhr ist ein Metadatum (#1458). Ohne diese Zeile hinge die Kette an
	// der Wanduhr zweier Rechner.
	spaeter := erstes
	spaeter.Geschehen = time.Now()

	if abdruck(t, spaeter, "") != abdruck(t, erstes, "") {
		t.Error("der Zeitstempel geht in den Abdruck ein")
	}

	// **Der eigene Abdruck geht nicht in sich selbst ein.**
	mitAbdruck := erstes
	mitAbdruck.Kette = "irgendwas"

	if abdruck(t, mitAbdruck, "") != abdruck(t, erstes, "") {
		t.Error("die Kette rechnet sich selbst mit")
	}
}

// **Ein anderer Inhalt, ein anderer Abdruck.**
//
// **Die Gegenprobe zum Lauf darüber**: ohne sie stünde nur, dass der Abdruck
// stabil ist — und das wäre er auch, wenn er immer dasselbe zurückgäbe.
func TestEinAndererInhaltGibtEinenAnderenAbdruck(t *testing.T) {
	t.Parallel()

	grund := domain.Ereignis{
		PartieID: "p1", Folge: 1, Art: "karte.gezogen",
		Oeffentlich: map[string]any{"sitz": "sitz-1"},
	}

	faelle := map[string]func(*domain.Ereignis){
		"andere Folge":  func(e *domain.Ereignis) { e.Folge = 2 },
		"andere Art":    func(e *domain.Ereignis) { e.Art = "karte.gelegt" },
		"andere Partie": func(e *domain.Ereignis) { e.PartieID = "p2" },
		"anderer Inhalt": func(e *domain.Ereignis) {
			e.Oeffentlich = map[string]any{"sitz": "sitz-2"}
		},
	}

	for name, aendern := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anderes := grund
			aendern(&anderes)

			if abdruck(t, anderes, "") == abdruck(t, grund, "") {
				t.Errorf("%s ändert den Abdruck nicht", name)
			}
		})
	}

	// **Und der Vorgänger geht ein** — sonst wäre es keine Kette, sondern
	// eine Liste von Abdrücken.
	if abdruck(t, grund, "aa") == abdruck(t, grund, "bb") {
		t.Error("der Vorgänger geht nicht in den Abdruck ein")
	}
}

// **Die private Ergänzung hat ihren eigenen Nachweis.**
//
// **Das ist die Zusage, die verhindert, dass Integrität zum Offenlegen
// zwingt** (#1458): wer die Kette prüfen will, braucht den öffentlichen Strom
// — und **nicht** die Hand des Gegenübers.
func TestDiePrivateErgaenzungAendertDenOeffentlichenAbdruckNicht(t *testing.T) {
	t.Parallel()

	ohne := domain.Ereignis{
		PartieID: "p1", Folge: 1, Art: "karte.gezogen",
		Oeffentlich: map[string]any{"sitz": "sitz-1"},
	}

	mit := ohne
	mit.Privat = []domain.Ergaenzung{{
		Empfaenger: "sitz-1",
		Inhalt:     map[string]any{"karte": "wald"},
	}}

	if abdruck(t, mit, "") != abdruck(t, ohne, "") {
		t.Error("die private Ergänzung geht in den öffentlichen Abdruck ein")
	}

	// **Sie hat trotzdem einen** — sonst ließe sie sich unbemerkt
	// austauschen.
	erster := ergaenzungsabdruck(t, mit.Privat[0])

	geaendert := mit.Privat[0]
	geaendert.Inhalt = map[string]any{"karte": "gebirge"}

	if ergaenzungsabdruck(t, geaendert) == erster {
		t.Error("die Ergänzung hat keinen wirksamen eigenen Nachweis")
	}

	// **Und der Empfänger geht ein** — dieselbe Karte an einen anderen ist
	// eine andere Ergänzung.
	fremd := mit.Privat[0]
	fremd.Empfaenger = "sitz-2"

	if ergaenzungsabdruck(t, fremd) == erster {
		t.Error("der Empfänger geht nicht in den Nachweis ein")
	}
}

// abdruck rechnet ihn und lässt den Lauf scheitern, wenn er es nicht kann.
func abdruck(t *testing.T, ereignis domain.Ereignis, vorher string) string {
	t.Helper()

	gerechnet, fehler := ereignis.Abdruck(vorher)
	if fehler != nil {
		t.Fatalf("Abdruck: %v", fehler)
	}

	return gerechnet
}

// ergaenzungsabdruck tut dasselbe für eine Ergänzung.
func ergaenzungsabdruck(t *testing.T, ergaenzung domain.Ergaenzung) string {
	t.Helper()

	gerechnet, fehler := ergaenzung.Abdruck()
	if fehler != nil {
		t.Fatalf("Abdruck: %v", fehler)
	}

	return gerechnet
}

// **Was sich nicht kodieren lässt, bekommt keinen Abdruck.**
//
// **Ein echter Fall und kein erfundener**: `Oeffentlich` ist ein
// `map[string]any`, und ein Aufrufer kann daraus eine fremde Eingabe bauen.
func TestWasSichNichtKodierenLaesstBekommtKeinenAbdruck(t *testing.T) {
	t.Parallel()

	ereignis := domain.Ereignis{
		PartieID: "p1", Folge: 1, Art: "karte.gezogen",
		Oeffentlich: map[string]any{"kanal": make(chan int)},
	}

	if _, fehler := ereignis.Abdruck(""); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}

	ergaenzung := domain.Ergaenzung{
		Empfaenger: "sitz-1",
		Inhalt:     map[string]any{"kanal": make(chan int)},
	}

	if _, fehler := ergaenzung.Abdruck(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}
