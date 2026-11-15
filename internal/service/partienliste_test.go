package service_test

// **Die Liste der eigenen Partien** —
// #1600.
//
// # Wogegen hier geprüft wird
//
//   - Die Liste gibt **Zeilen** und keine Partien — was in einer Partie
//     liegt, sagt die Projektion (G1).
//   - Ein Fehler des Speichers kommt durch und wird nicht zu einer leeren
//     Liste, die wie „du hast keine“ aussähe (§8).
//   - **Keine Partie ist eine leere Liste und kein `nil`.**

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDiePartienlisteGibtZeilen(t *testing.T) {
	t.Parallel()

	probe := neueProbe()
	probe.partien["p1"] = domain.Partie{
		ID: "p1", Fassung: 4, Spiel: "magic", Stand: domain.StandLaeuft,
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "David"},
			{Kennung: "sitz-2", Name: "Gegenüber"},
		},
		Objekte: []domain.Spielobjekt{{Kennung: "sitz-1/k1/0"}},
	}

	zeilen, fehler := service.NeuePartieleitung(probe, nil).Partien(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Partien: %v", fehler)
	}

	if len(zeilen) != 1 {
		t.Fatalf("es kamen %d Zeilen", len(zeilen))
	}

	switch {
	case zeilen[0].ID != "p1":
		t.Errorf("die Kennung ist %q", zeilen[0].ID)
	case zeilen[0].Objekte != 1:
		t.Errorf("gezählt sind %d Objekte", zeilen[0].Objekte)
	case len(zeilen[0].Sitze) != 2:
		t.Errorf("die Sitze sind %v", zeilen[0].Sitze)
	}
}

// **Keine Partie ist eine leere Liste und kein `nil`.**
func TestOhnePartienKommtEineLeereListe(t *testing.T) {
	t.Parallel()

	zeilen, fehler := service.NeuePartieleitung(neueProbe(), nil).Partien(t.Context(), 7)

	switch {
	case fehler != nil:
		t.Fatalf("Partien: %v", fehler)
	case zeilen == nil:
		t.Error("die Liste ist nil statt leer")
	case len(zeilen) != 0:
		t.Errorf("es kamen %d Zeilen", len(zeilen))
	}
}

// **Ein Fehler des Speichers kommt durch.**
//
// **Und wird nicht zur leeren Liste**: „ich konnte nicht nachsehen“ und „du
// hast keine Partie“ sind zwei verschiedene Auskünfte mit zwei verschiedenen
// nächsten Schritten (§8).
func TestEinFehlerBeimPartienlesenKommtDurch(t *testing.T) {
	t.Parallel()

	probe := neueProbe()
	probe.klagt = errors.New("der Bestand ist zu")

	if _, fehler := service.NeuePartieleitung(probe, nil).Partien(
		t.Context(), 7); fehler == nil {
		t.Error("der Fehler wurde verschluckt")
	}
}
