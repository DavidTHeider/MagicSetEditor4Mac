package domain_test

// **Der Handelsvertrag** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was hier gemessen wird
//
// **Dass eine Zusage an einer Revision hängt und nicht an einem Gespräch.**
// Ein Angebot ist unveränderlich: wer etwas ändert, erzeugt die **nächste**
// Revision, und die trägt den Abdruck ihrer Eltern. Damit ist „ich habe dem
// zugestimmt" eine Aussage über einen Stand und nicht über eine Absicht.
//
// # Warum der Abdruck jedes fachliche Feld bindet
//
// **Weil eine Unterschrift sonst weniger zusagt, als sie aussieht.** Wer über
// „drei Karten" unterschreibt und nicht über ihren Zustand, hat einer Menge
// zugestimmt und keinem Handel. [TestDerAbdruckHaengtAnJedemFachlichenFeld]
// zählt jedes Feld einzeln nach — das ist die Gegenprobe zur Aufzählung in
// der Unterschriftsform (§10).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einAngebot ist das kleinste, das prüft.
func einAngebot() domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,

		Vorgang:   "v1",
		Revision:  1,
		Art:       domain.HandelTausch,
		Initiator: "anna",
		Parteien:  []string{"anna", "bert"},
		Zustand:   domain.HandelEntwurf,
		Positionen: []domain.Handelsposition{{
			Von:      "anna",
			Printing: "abc-123",
			Menge:    2,
			Guete:    "NM",
			Sprache:  "de",
			Finish:   "nonfoil",
			Abbild:   "abdruck-der-zeile",
		}},
	}
}

// **Ein Angebot, das nicht prüft, entsteht nicht.**
func TestEinAngebotWirdGeprueft(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Handelsangebot){
		"ohne Vertragsfassung": func(a *domain.Handelsangebot) {
			a.Vertragsfassung = 0
		},
		"ohne Vorgang":       func(a *domain.Handelsangebot) { a.Vorgang = "" },
		"ohne Revision":      func(a *domain.Handelsangebot) { a.Revision = 0 },
		"ohne Art":           func(a *domain.Handelsangebot) { a.Art = "" },
		"mit fremder Art":    func(a *domain.Handelsangebot) { a.Art = "schenken" },
		"ohne Initiator":     func(a *domain.Handelsangebot) { a.Initiator = "" },
		"mit einer Partei":   func(a *domain.Handelsangebot) { a.Parteien = []string{"anna"} },
		"ohne den Initiator": func(a *domain.Handelsangebot) { a.Parteien = []string{"bert", "cara"} },
		"ohne Zustand":       func(a *domain.Handelsangebot) { a.Zustand = "" },
		"mit fremdem Zustand": func(a *domain.Handelsangebot) {
			a.Zustand = "fast fertig"
		},
		"ohne Positionen": func(a *domain.Handelsangebot) {
			a.Positionen = nil
		},
		"mit einer Position ohne Menge": func(a *domain.Handelsangebot) {
			a.Positionen[0].Menge = 0
		},
		"mit einer Position ohne Druck": func(a *domain.Handelsangebot) {
			a.Positionen[0].Printing = ""
			a.Positionen[0].DruckID = 0
		},
		"mit einer Position einer fremden Partei": func(a *domain.Handelsangebot) {
			a.Positionen[0].Von = "cara"
		},
		"mit Elternabdruck in der ersten Revision": func(a *domain.Handelsangebot) {
			a.Eltern = "abdruck"
		},
		"ohne Elternabdruck in der zweiten": func(a *domain.Handelsangebot) {
			a.Revision = 2
		},
	}

	for name, verbiegen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			angebot := einAngebot()
			verbiegen(&angebot)

			if fehler := angebot.Pruefe(); fehler == nil {
				t.Errorf("%q kommt durch", name)
			}
		})
	}

	if fehler := einAngebot().Pruefe(); fehler != nil {
		t.Errorf("das gute Angebot wird abgewiesen: %v", fehler)
	}
}

// **Der Abdruck hängt an jedem fachlichen Feld.**
//
// **Die Aufzählung ist die Zusage**, und dieser Lauf ist ihre Gegenprobe: ein
// Feld, das die Unterschriftsform vergisst, kommt hier als grüner Fall
// zurück — und ein grüner Fall ist hier ein **Fehler**.
func TestDerAbdruckHaengtAnJedemFachlichenFeld(t *testing.T) {
	t.Parallel()

	aendern := map[string]func(*domain.Handelsangebot){
		"der Vorgang":    func(a *domain.Handelsangebot) { a.Vorgang = "v2" },
		"die Revision":   func(a *domain.Handelsangebot) { a.Revision = 2; a.Eltern = "x" },
		"die Art":        func(a *domain.Handelsangebot) { a.Art = domain.HandelTeilen },
		"der Initiator":  func(a *domain.Handelsangebot) { a.Initiator = "bert" },
		"die Parteien":   func(a *domain.Handelsangebot) { a.Parteien = []string{"anna", "cara"} },
		"die Seite":      func(a *domain.Handelsangebot) { a.Positionen[0].Von = "bert" },
		"der Druck":      func(a *domain.Handelsangebot) { a.Positionen[0].Printing = "xyz-9" },
		"die Druckzeile": func(a *domain.Handelsangebot) { a.Positionen[0].DruckID = 7 },
		"die Menge":      func(a *domain.Handelsangebot) { a.Positionen[0].Menge = 3 },
		"die Güte":       func(a *domain.Handelsangebot) { a.Positionen[0].Guete = "LP" },
		"die Sprache":    func(a *domain.Handelsangebot) { a.Positionen[0].Sprache = "en" },
		"das Finish":     func(a *domain.Handelsangebot) { a.Positionen[0].Finish = "foil" },
		"das Abbild":     func(a *domain.Handelsangebot) { a.Positionen[0].Abbild = "anders" },
		"eine zweite Position": func(a *domain.Handelsangebot) {
			a.Positionen = append(a.Positionen, a.Positionen[0])
		},
	}

	vorher := einAngebot().Abdruck()

	for name, verbiegen := range aendern {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			angebot := einAngebot()
			verbiegen(&angebot)

			if angebot.Abdruck() == vorher {
				t.Errorf("%s ändert den Abdruck nicht", name)
			}
		})
	}
}

// **Der Zustand und die Zustimmungen gehen nicht ein.**
//
// # Warum nicht
//
// **Weil sie sich am selben Angebot ändern.** Eine Zustimmung ist eine
// Aussage **über** eine Revision; ginge sie in deren Abdruck ein, änderte das
// Zustimmen den Stand, dem zugestimmt wurde. Derselbe Satz wie bei
// [Deckleihe.Unterschriftsform] über die Unterschrift selbst.
func TestZustandUndZustimmungGehenNichtInDenAbdruck(t *testing.T) {
	t.Parallel()

	angebot := einAngebot()
	vorher := angebot.Abdruck()

	angebot.Zustand = domain.HandelAngeboten
	angebot.Zustimmungen = []string{"anna"}

	if angebot.Abdruck() != vorher {
		t.Error("der Zustand oder die Zustimmung ändert den Abdruck")
	}
}

// **Die Unterschriftsform nennt ihre Domäne und trennt mit dem Nullbyte.**
//
// **Ohne das Domänenwort** ließe sich dieselbe Unterschrift für ein anderes
// Dokument mit denselben Feldern wiederverwenden; ohne einen Trenner, der in
// keinem Feld vorkommt, hätten zwei verschiedene Angebote dieselbe Form
// (dieselbe Überlegung wie bei [Deckleihe.Unterschriftsform]).
func TestDieUnterschriftsformNenntIhreDomaene(t *testing.T) {
	t.Parallel()

	form := string(einAngebot().Unterschriftsform())

	if !strings.HasPrefix(form, "handelsangebot\x00") {
		t.Errorf("die Form fängt an mit %q", form[:min(30, len(form))])
	}

	if !strings.Contains(form, "\x00") {
		t.Error("die Form trennt nicht mit dem Nullbyte")
	}
}
