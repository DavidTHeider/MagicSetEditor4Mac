package service_test

// **Der Plan vor einer Paketänderung** — DOM-01.AK-02
// (#1696):
// „Installation, Aktualisierung, Deinstallation und Rückkehr zur vorigen
// kompatiblen Version besitzen eine Vorschau betroffener Sets/Stile und
// erkennen Abhängigkeitskonflikte."

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// planplatz ist ein Katalog mit Spiel und Stil, ein Set auf beiden und ein
// verwaltetes Verzeichnis.
func planplatz(t *testing.T) *service.Paketverwalter {
	t.Helper()

	verwaltet := t.TempDir()
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "probe", Version: "2026-08-15", Pfad: filepath.Join(verwaltet, "probe.mse-game")},
		{
			Art: "mse-style", Name: "probe-schlicht", Version: "2026-08-15",
			Pfad:    filepath.Join(verwaltet, "probe-schlicht.mse-style"),
			Braucht: []domain.Bedarf{{Paket: "probe.mse-game", Version: "2026-08-15"}},
		},
	}}

	verwalter := service.NeuerPaketverwalter(katalog, verwaltet, verwaltet).
		MitSetkoepfen(setkoepfe{koepfe: []domain.SetHead{
			{Name: "Erstes", Game: "probe", Style: "schlicht"},
			{Name: "Zweites", Game: "probe"},
		}})

	return verwalter
}

func TestDerPlanEinesEntfernensNenntSetsUndWerEsBraucht(t *testing.T) {
	verwalter := planplatz(t)

	plan, err := verwalter.PlanEntfernen(t.Context(), "mse-game", "probe")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if plan.Vorgang != service.PaketvorgangEntfernen || plan.Paket != "probe.mse-game" ||
		plan.Von != "2026-08-15" || plan.Zu != "" || !plan.Moeglich {
		t.Errorf("der Plan: %+v", plan)
	}

	if !slices.Equal(plan.Sets, []string{"Erstes", "Zweites"}) {
		t.Errorf("die Sets: %v", plan.Sets)
	}

	// **Der Stil verlöre sein Spiel** — ein Konflikt, und genannt, nicht
	// mitgenommen.
	if len(plan.Konflikte) != 1 || plan.Konflikte[0].Paket != "probe-schlicht.mse-style" ||
		plan.Konflikte[0].Art != service.AbhaengigkeitFehlt {
		t.Errorf("die Konflikte: %+v", plan.Konflikte)
	}

	if _, err := verwalter.PlanEntfernen(t.Context(), "mse-game", "gibtsnicht"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Paket: %v", err)
	}
}

// **Ein Paket außerhalb des verwalteten Verzeichnisses entfernt der Verwalter
// nicht** — der Plan sagt es vorher.
func TestDerPlanSagtWasNichtGeht(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{{Art: "mse-game", Name: "saat", Pfad: "/buendel/saat/saat.mse-game"}}}

	plan, err := service.NeuerPaketverwalter(katalog, t.TempDir()).PlanEntfernen(t.Context(), "mse-game", "saat")
	if err != nil || plan.Moeglich {
		t.Errorf("das mitgelieferte Paket: %+v, %v", plan, err)
	}
}

func TestDerPlanEinesEinbindensUndEinesErsetzens(t *testing.T) {
	verwalter := planplatz(t)
	quelle := t.TempDir()

	// **Neu, aber mit einem Bedarf, den es hier nicht gibt.**
	paket(t, quelle, "probe-kunst.mse-style",
		"mse version: 2.0.0\ngame: probe\nversion: 2026-09-01\ndepends on:\n\tpackage: fehlt.mse-symbol-font\n")

	plan, err := verwalter.PlanEinbinden(t.Context(), filepath.Join(quelle, "probe-kunst.mse-style"))
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if plan.Vorgang != service.PaketvorgangEinbinden || plan.Von != "" || plan.Zu != "2026-09-01" || !plan.Moeglich ||
		len(plan.Konflikte) != 1 || plan.Konflikte[0].Braucht != "fehlt.mse-symbol-font" {
		t.Errorf("das neue Paket: %+v", plan)
	}

	// **Ein Spiel, das es schon gibt, wird ersetzt** — und eine ältere
	// Fassung ließe den Stil zurück, der die neuere verlangt.
	paket(t, quelle, "probe.mse-game", "mse version: 2.0.0\nversion: 2026-01-01\n")

	ersatz, err := verwalter.PlanEinbinden(t.Context(), filepath.Join(quelle, "probe.mse-game"))
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if ersatz.Vorgang != service.PaketvorgangErsetzen || ersatz.Von != "2026-08-15" || ersatz.Zu != "2026-01-01" ||
		!slices.Equal(ersatz.Sets, []string{"Erstes", "Zweites"}) ||
		len(ersatz.Konflikte) != 1 || ersatz.Konflikte[0].Art != service.AbhaengigkeitZuAlt {
		t.Errorf("der Ersatz: %+v", ersatz)
	}

	if _, err := verwalter.PlanEinbinden(t.Context(), filepath.Join(quelle, "gibtsnicht.mse-game")); err == nil {
		t.Error("ein Plan aus dem Nichts")
	}
}

// **Ohne vorige Fassung gibt es kein Zurück** — der Plan sagt es.
func TestDerPlanEinerRueckkehrOhneVorigeFassung(t *testing.T) {
	verwalter := planplatz(t)

	plan, err := verwalter.PlanZurueck(t.Context(), "mse-game", "probe")
	if err != nil || plan.Moeglich || plan.Vorgang != service.PaketvorgangZurueck {
		t.Errorf("ohne vorige Fassung: %+v, %v", plan, err)
	}
}

// **Ein Fehler beim Lesen kommt durch.**
func TestDerPlanReichtFehlerDurch(t *testing.T) {
	klemmt := errors.New("klemmt")

	for name, ruf := range map[string]func(*service.Paketverwalter) error{
		"entfernen": func(v *service.Paketverwalter) error {
			_, err := v.PlanEntfernen(t.Context(), "mse-game", "probe")

			return err
		},
		"einbinden": func(v *service.Paketverwalter) error {
			_, err := v.PlanEinbinden(t.Context(), t.TempDir())

			return err
		},
		"zurück": func(v *service.Paketverwalter) error {
			_, err := v.PlanZurueck(t.Context(), "mse-game", "probe")

			return err
		},
	} {
		if err := ruf(service.NeuerPaketverwalter(&ablage{lesefehler: klemmt}, t.TempDir())); err == nil {
			t.Errorf("%s: der Fehler ging verloren", name)
		}
	}

	// **Ein lesbares Paket, aber ein klemmender Katalog.**
	quelle := t.TempDir()
	paket(t, quelle, "probe-kunst.mse-style", "mse version: 2.0.0\ngame: probe\n")

	if _, err := service.NeuerPaketverwalter(&ablage{lesefehler: klemmt}, t.TempDir()).
		PlanEinbinden(t.Context(), filepath.Join(quelle, "probe-kunst.mse-style")); !errors.Is(err, klemmt) {
		t.Errorf("der Katalog klemmt beim Einbinden: %v", err)
	}

	verwalter := planplatz(t)
	verwalter.MitSetkoepfen(setkoepfe{fehler: klemmt})

	if _, err := verwalter.PlanEntfernen(t.Context(), "mse-game", "probe"); !errors.Is(err, klemmt) {
		t.Errorf("die Setköpfe klemmen: %v", err)
	}
}
