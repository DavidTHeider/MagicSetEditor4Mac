package app_test

// **Die Naht zwischen Dienst und Fläche** — Werkstück B von Phase 4.11
// (#1337).
//
// Geprüft wird die Übersetzung, die zwischen beiden steht: aus Angaben werden
// DTOs, aus `int64`-Kennungen werden Zeichenketten, und aus einem `nil`-Slice
// Befunde wird eine leere Liste — das JSON dahinter hat für „nichts" nur eine
// brauchbare Gestalt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nahtspiel gibt ein Spiel mit einem Kartenskript.
type nahtspiel struct {
	skripte []service.Kartenskriptangabe
	fehler  error
}

func (n *nahtspiel) Kartenskripte(_ context.Context, _ int64) (
	[]service.Kartenskriptangabe, []domain.Field, string, error,
) {
	return n.skripte, []domain.Field{{Name: "name", Kind: domain.KindText}}, "", n.fehler
}

// nahtanleger legt an und gibt aufsteigende Kennungen.
type nahtanleger struct{ fehler error }

func (n *nahtanleger) KartenAnlegen(
	_ context.Context, _ int64, karten [][]domain.Feldstand,
) ([]int64, error) {
	if n.fehler != nil {
		return nil, n.fehler
	}

	kennungen := make([]int64, 0, len(karten))
	for i := range karten {
		kennungen = append(kennungen, int64(i+1))
	}

	return kennungen, nil
}

func nahtueber(spiel *nahtspiel, anleger *nahtanleger) *app.Kartenskriptnaht {
	return app.NeueKartenskriptnaht(service.NeuerKartenskriptdienst(anleger, spiel))
}

// **Aus Angaben werden DTOs — mit Kürzel und Beschreibung.**
func TestDieKartenskriptnahtGibtDasAngebotAlsDTO(t *testing.T) {
	t.Parallel()

	naht := nahtueber(&nahtspiel{skripte: []service.Kartenskriptangabe{
		{Name: "Zwei Farben", Kuerzel: "Z", Beschreibung: "zwei Karten", Skript: `[]`},
	}}, &nahtanleger{})

	skripte, fehler := naht.Angebot(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("Angebot: %v", fehler)
	}

	if len(skripte) != 1 {
		t.Fatalf("%d Skripte", len(skripte))
	}

	if skripte[0] != (app.KartenskriptDTO{
		Name: "Zwei Farben", Kuerzel: "Z", Beschreibung: "zwei Karten",
	}) {
		t.Errorf("das DTO lautet %+v", skripte[0])
	}
}

// **Ein leeres Angebot ist eine leere Liste, kein `nil`.**
func TestDieKartenskriptnahtGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	skripte, fehler := nahtueber(&nahtspiel{}, &nahtanleger{}).Angebot(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("Angebot: %v", fehler)
	}

	if skripte == nil {
		t.Error("die Liste ist nil")
	}
}

// **Ein Fehler des Dienstes kommt durch — und keine halbe Liste.**
func TestDieKartenskriptnahtReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Spielpaket")

	naht := nahtueber(&nahtspiel{fehler: eigener}, &nahtanleger{})

	skripte, fehler := naht.Angebot(probekontext(t), 1)
	if !errors.Is(fehler, eigener) || skripte != nil {
		t.Errorf("Angebot gab %+v / %v", skripte, fehler)
	}

	lauf, laufFehler := naht.Ausfuehren(probekontext(t), 1, "egal")
	if !errors.Is(laufFehler, eigener) || lauf != nil {
		t.Errorf("Ausfuehren gab %+v / %v", lauf, laufFehler)
	}
}

// **Ein Lauf gibt Kennungen als Text und Befunde als Liste.**
//
// Die leere Liste ist der Punkt: `nil` würde im JSON zu `null`, und die Seite
// müsste beides können.
func TestDieKartenskriptnahtGibtDenLaufAlsDTO(t *testing.T) {
	t.Parallel()

	naht := nahtueber(&nahtspiel{skripte: []service.Kartenskriptangabe{{
		Name:   "Zwei",
		Skript: `[ new_card([name: "Eins"]), new_card([name: "Zwei"]) ]`,
	}}}, &nahtanleger{})

	lauf, fehler := naht.Ausfuehren(probekontext(t), 1, "Zwei")
	if fehler != nil {
		t.Fatalf("Ausfuehren: %v", fehler)
	}

	if len(lauf.Karten) != 2 || lauf.Karten[0] != "1" || lauf.Karten[1] != "2" {
		t.Errorf("die Kennungen lauten %+v", lauf.Karten)
	}

	if lauf.Befunde == nil {
		t.Error("die Befunde sind nil")
	}
}
