package service_test

// **Die Einwilligung, dass später jemand mitliest** — Werkstück G5 von Phase
// 5.5 (#1462).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einTischMitGeheimnis gibt eine Leitung über eine Partie, in der zwei Sitze
// je eine private Ergänzung haben.
func einTischMitGeheimnis(t *testing.T) (*service.Partieleitung, *partieprobe) {
	t.Helper()

	speicher := neueProbe()
	speicher.partien["p1"] = domain.Partie{
		ID: "p1", Stand: domain.StandLaeuft,
		Sitze: []domain.Sitz{{Kennung: "anna"}, {Kennung: "ben"}},
	}
	speicher.ereignisse["p1"] = []domain.Ereignis{{
		PartieID: "p1", Folge: 1, Art: "ziehen",
		Oeffentlich: map[string]any{"sitz": "anna"},
		Privat: []domain.Ergaenzung{
			{Empfaenger: "anna", Inhalt: map[string]any{"karte": "Wald"}},
			{Empfaenger: "ben", Inhalt: map[string]any{"karte": "Insel"}},
		},
		Kette: "abc",
	}}

	return service.NeuePartieleitung(
		speicher, service.NeuerSandkasten(service.NeuerZufall())), speicher
}

func TestOhneEinwilligungBleibtDasGeheimnisEigen(t *testing.T) {
	t.Parallel()

	leitung, _ := einTischMitGeheimnis(t)

	strom, fehler := leitung.Verlauf(t.Context(), "p1", "ben")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(strom[0].Privat) != 1 {
		t.Fatalf("erwartet eine Ergänzung, bekam %v", strom[0].Privat)
	}
}

func TestEineEinwilligungOeffnetDenVerlauf(t *testing.T) {
	t.Parallel()

	leitung, _ := einTischMitGeheimnis(t)

	fehler := leitung.Offenlegen(t.Context(), "p1",
		domain.Offenlegung{Von: "anna", An: []string{"ben"}})
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	strom, fehler := leitung.Verlauf(t.Context(), "p1", "ben")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(strom[0].Privat) != 2 {
		t.Fatalf("erwartet beide Ergänzungen, bekam %v", strom[0].Privat)
	}
}

func TestEineEinwilligungWirktNichtBeimNachreichen(t *testing.T) {
	t.Parallel()

	leitung, _ := einTischMitGeheimnis(t)

	if fehler := leitung.Offenlegen(t.Context(), "p1",
		domain.Offenlegung{Von: "anna", An: []string{"ben"}}); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	strom, fehler := leitung.Nachreichen(t.Context(), "p1", "ben", 0)
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(strom[0].Privat) != 1 {
		t.Fatalf("der laufende Tisch hat die Nachlese-Einwilligung "+
			"mitgerechnet: %v", strom[0].Privat)
	}
}

func TestEineZuruecknahmeSchliesstDieTuer(t *testing.T) {
	t.Parallel()

	leitung, _ := einTischMitGeheimnis(t)

	if fehler := leitung.Offenlegen(t.Context(), "p1",
		domain.Offenlegung{Von: "anna", An: []string{"ben"}}); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if fehler := leitung.Zuruecknehmen(
		t.Context(), "p1", "anna"); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	strom, fehler := leitung.Verlauf(t.Context(), "p1", "ben")
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(strom[0].Privat) != 1 {
		t.Fatalf("erwartet eine Ergänzung, bekam %v", strom[0].Privat)
	}
}

func TestEineZweiteEinwilligungErsetztDieErste(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischMitGeheimnis(t)

	for _, an := range [][]string{{"ben"}, {"cem"}} {
		if fehler := leitung.Offenlegen(t.Context(), "p1",
			domain.Offenlegung{Von: "anna", An: an}); fehler != nil {
			t.Fatalf("Fehler: %v", fehler)
		}
	}

	gaben := speicher.freigaben["p1"]
	if len(gaben) != 1 || len(gaben[0].An) != 1 || gaben[0].An[0] != "cem" {
		t.Fatalf("erwartet nur die zweite Liste, bekam %v", gaben)
	}
}

func TestNurEinSitzLegtOffen(t *testing.T) {
	t.Parallel()

	leitung, _ := einTischMitGeheimnis(t)

	fehler := leitung.Offenlegen(t.Context(), "p1",
		domain.Offenlegung{Von: "zuschauer-1", An: []string{"ben"}})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

func TestEineLaufendePartieWirdNichtVergessen(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischMitGeheimnis(t)

	fehler := leitung.Vergessen(t.Context(), "p1")

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
	}

	if _, steht := speicher.partien["p1"]; !steht {
		t.Fatal("die Partie wurde trotzdem gelöscht")
	}
}

func TestEineBeendetePartieWirdVergessen(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischMitGeheimnis(t)

	partie := speicher.partien["p1"]
	partie.Stand = domain.StandBeendet
	speicher.partien["p1"] = partie

	if fehler := leitung.Vergessen(t.Context(), "p1"); fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if _, steht := speicher.partien["p1"]; steht {
		t.Fatal("die Partie steht noch da")
	}

	if _, steht := speicher.ereignisse["p1"]; steht {
		t.Fatal("der Strom steht noch da")
	}
}

// **Jeder Fehlerweg der Offenlegung wird betreten** (§3).
func TestJederFehlerwegDerOffenlegungWirdBetreten(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischMitGeheimnis(t)

	// **Eine Einwilligung, die keine ist.**
	if fehler := leitung.Offenlegen(t.Context(), "p1",
		domain.Offenlegung{Von: "anna"}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die leere Empfängerliste kam durch: %v", fehler)
	}

	// **Eine Partie, die es nicht gibt** — auf allen vier Wegen.
	gabe := domain.Offenlegung{Von: "anna", An: []string{"ben"}}

	wege := map[string]func(string) error{
		"offenlegen": func(kennung string) error {
			return leitung.Offenlegen(t.Context(), kennung, gabe)
		},
		"zurücknehmen": func(kennung string) error {
			return leitung.Zuruecknehmen(t.Context(), kennung, "anna")
		},
		"lesen": func(kennung string) error {
			_, fehler := leitung.Freigaben(t.Context(), kennung)

			return fehler
		},
		"vergessen": func(kennung string) error {
			return leitung.Vergessen(t.Context(), kennung)
		},
	}

	for name, weg := range wege {
		if fehler := weg("gibtesnicht"); !errors.Is(
			fehler, domain.ErrNotFound) {
			t.Errorf("%s: erwartet ErrNotFound, bekam %v", name, fehler)
		}
	}

	// **Und eine klagende Ablage auf dem Weg zu den Freigaben.**
	speicher.stolpert = map[string]error{"freigaben": domain.ErrConflict}

	if _, fehler := leitung.Freigaben(t.Context(), "p1"); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("die Klage kam nicht an: %v", fehler)
	}

	if _, fehler := leitung.Verlauf(t.Context(), "p1", "anna"); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("die Klage im Verlauf kam nicht an: %v", fehler)
	}
}

// **Und der Erfolgsweg gibt, was eingetragen wurde.**
func TestDieFreigabenKommenZurueck(t *testing.T) {
	t.Parallel()

	leitung, _ := einTischMitGeheimnis(t)

	if fehler := leitung.Offenlegen(t.Context(), "p1",
		domain.Offenlegung{Von: "anna", An: []string{"ben"}}); fehler != nil {
		t.Fatalf("offenlegen: %v", fehler)
	}

	gaben, fehler := leitung.Freigaben(t.Context(), "p1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gaben) != 1 || gaben[0].Von != "anna" {
		t.Fatalf("die Freigaben sind %v", gaben)
	}
}
