package service_test

// **Massenvorgänge über eine Auswahl** — Werkstück 6 von Phase 4.5
// (#1292).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dreiZeilen legt drei Exemplare an und gibt Ablage und Zeilen.
func dreiZeilen(t *testing.T) (*sammlungsdoppel, []service.Massenzeile) {
	t.Helper()

	ablage := &sammlungsdoppel{}
	zeilen := make([]service.Massenzeile, 0, 3)

	for _, zustand := range []domain.Zustand{
		domain.ZustandNearMint, domain.ZustandPlayed, domain.ZustandGood,
	} {
		id, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
			Printing: probeprinting, Menge: 1, Zustand: zustand,
			Sprache: domain.Vorgabesprache, Lagerort: "Kiste 4",
		})
		if fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		zeilen = append(zeilen, service.Massenzeile{ID: id, Fassung: 1})
	}

	return ablage, zeilen
}

// **Ein Lagerort für alle drei auf einmal.**
func TestEinLagerortFuerDreiZeilen(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)

	ordner := "Ordner Rot"

	bericht, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{Zeilen: zeilen, Lagerort: &ordner})
	if fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	if bericht.Geaendert != 3 {
		t.Fatalf("%d geändert, erwartet 3: %v", bericht.Geaendert, bericht.Befunde)
	}

	gelesen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	for _, exemplar := range gelesen {
		if exemplar.Lagerort != "Ordner Rot" {
			t.Errorf("Exemplar %d liegt in %q", exemplar.ID, exemplar.Lagerort)
		}
	}
}

// **Ein leerer Lagerort ist eine Angabe** — „nimm die Ortsangabe weg".
//
// Das ist der Fall, für den die Änderungen Zeiger sind: ohne sie wäre er von
// „lass den Lagerort in Ruhe" nicht zu unterscheiden.
func TestEinLeererLagerortIstEineAngabe(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)

	leer := ""

	if _, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{Zeilen: zeilen[:1], Lagerort: &leer}); fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	gelesen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	for _, exemplar := range gelesen {
		if exemplar.ID == zeilen[0].ID && exemplar.Lagerort != "" {
			t.Errorf("der Lagerort ist %q geblieben", exemplar.Lagerort)
		}
	}
}

// **Was nicht genannt ist, bleibt.**
func TestWasNichtGenanntIstBleibt(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)

	menge := 7

	if _, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{Zeilen: zeilen[:1], Menge: &menge}); fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	gelesen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	for _, exemplar := range gelesen {
		if exemplar.ID != zeilen[0].ID {
			continue
		}

		if exemplar.Menge != 7 {
			t.Errorf("Menge %d, erwartet 7", exemplar.Menge)
		}

		// **Zustand und Lagerort stehen noch** — sie waren nicht genannt.
		if exemplar.Zustand != domain.ZustandNearMint || exemplar.Lagerort != "Kiste 4" {
			t.Errorf("was nicht genannt war, hat sich geändert: %+v", exemplar)
		}
	}
}

// **Eine veraltete Fassung hält die anderen nicht auf** (§8).
//
// Und sie kommt als Befund mit ihrer Kennung heraus — sonst weiß niemand,
// welche Zeile übergangen wurde.
func TestEineVeralteteFassungWirdZumBefund(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)

	// **Die mittlere Zeile bekommt eine Fassung, die es nicht gibt.**
	zeilen[1].Fassung = 99

	menge := 5

	bericht, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{Zeilen: zeilen, Menge: &menge})
	if fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	if bericht.Geaendert != 2 {
		t.Errorf("%d geändert, erwartet 2", bericht.Geaendert)
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(bericht.Befunde), bericht.Befunde)
	}

	if !strings.Contains(bericht.Befunde[0].Ort, "Exemplar") {
		t.Errorf("der Befund zeigt auf %q", bericht.Befunde[0].Ort)
	}
}

// **Eine Kennung, die es nicht gibt, ebenso.**
func TestEineUnbekannteKennungWirdZumBefund(t *testing.T) {
	t.Parallel()

	ablage, _ := dreiZeilen(t)

	menge := 5

	bericht, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{
			Zeilen: []service.Massenzeile{{ID: 999, Fassung: 1}}, Menge: &menge,
		})
	if fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	if bericht.Geaendert != 0 || len(bericht.Befunde) != 1 {
		t.Errorf("%d geändert, %d Befunde", bericht.Geaendert, len(bericht.Befunde))
	}
}

// **Ein unbrauchbarer Wert wird zum Befund und schreibt nicht.**
func TestEinUnbrauchbarerWertInDerMasse(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)

	erfunden := domain.Zustand("wunderbar")

	bericht, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{Zeilen: zeilen, Zustand: &erfunden})
	if fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	if bericht.Geaendert != 0 {
		t.Errorf("%d geändert, erwartet 0", bericht.Geaendert)
	}

	if len(bericht.Befunde) != 3 {
		t.Errorf("%d Befunde, erwartet 3", len(bericht.Befunde))
	}
}

// **Ohne Zeilen und ohne Änderung ist es ein fachliches Nein.**
func TestMassenaenderungOhneZeilenOderAenderung(t *testing.T) {
	t.Parallel()

	pflege := service.NeueMassenpflege(&sammlungsdoppel{})

	if _, fehler := pflege.Aendern(t.Context(),
		service.Massenaenderung{}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Zeilen: %v", fehler)
	}

	if _, fehler := pflege.Aendern(t.Context(), service.Massenaenderung{
		Zeilen: []service.Massenzeile{{ID: 1, Fassung: 1}},
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Änderung: %v", fehler)
	}

	if _, fehler := pflege.Loeschen(t.Context(),
		nil); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Löschen ohne Zeilen: %v", fehler)
	}
}

// **Löschen in Masse — gegen die gelesene Fassung.**
func TestLoeschenInMasse(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)

	zeilen[2].Fassung = 99

	bericht, fehler := service.NeueMassenpflege(ablage).Loeschen(t.Context(), zeilen)
	if fehler != nil {
		t.Fatalf("Loeschen: %v", fehler)
	}

	if bericht.Geaendert != 2 || len(bericht.Befunde) != 1 {
		t.Errorf("%d gelöscht, %d Befunde", bericht.Geaendert, len(bericht.Befunde))
	}

	gelesen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	if len(gelesen) != 1 {
		t.Errorf("%d Zeilen übrig, erwartet 1", len(gelesen))
	}
}

// **Ein Lesefehler kommt durch und schreibt nichts.**
func TestEinLesefehlerInDerMasse(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")
	menge := 3

	_, fehler := service.NeueMassenpflege(&sammlungsdoppel{fehler: gestolpert}).
		Aendern(t.Context(), service.Massenaenderung{
			Zeilen: []service.Massenzeile{{ID: 1, Fassung: 1}}, Menge: &menge,
		})

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}
}

// **Der Wunsch lässt sich in Masse setzen** — eine Kiste wird zur Wunschliste.
func TestDerWunschLaesstSichInMasseSetzen(t *testing.T) {
	t.Parallel()

	ablage, zeilen := dreiZeilen(t)

	wunsch := true

	if _, fehler := service.NeueMassenpflege(ablage).Aendern(t.Context(),
		service.Massenaenderung{Zeilen: zeilen, Wunsch: &wunsch}); fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	gelesen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	for _, exemplar := range gelesen {
		if exemplar.Besitzt() {
			t.Errorf("Exemplar %d gilt noch als Besitz", exemplar.ID)
		}
	}
}
