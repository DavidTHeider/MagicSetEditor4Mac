package app_test

// **Der Plan, Ersetzen und Zurück an der Naht** — DOM-01.AK-02/AK-03
// (#1696).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestJederPlanKommtAlsDTO(t *testing.T) {
	fuehrer := &paketfuehrer{plan: service.Paketplan{
		Vorgang: service.PaketvorgangErsetzen, Paket: "probe.mse-game", Von: "1", Zu: "2",
		Geaendert: true, Moeglich: true, Sets: []string{"Erstes"},
		Konflikte: []service.Abhaengigkeitsbefund{{
			Paket: "probe-schlicht.mse-style", Braucht: "probe.mse-game", Art: service.AbhaengigkeitZuAlt,
			Verlangt: "3", Da: "2",
		}},
	}}

	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.PaketPlan(t.Context(), "einbinden", "", "", "/quelle/probe.mse-game")
	if antwort.Fehler != nil || antwort.Plan == nil || fuehrer.geplant != "einbinden /quelle/probe.mse-game" {
		t.Fatalf("der Plan: %+v, gefragt %q", antwort, fuehrer.geplant)
	}

	plan := antwort.Plan
	if plan.Vorgang != "ersetzen" || plan.Quelle != "/quelle/probe.mse-game" || plan.Von != "1" || plan.Zu != "2" ||
		!plan.Geaendert || !plan.Moeglich || len(plan.Sets) != 1 || len(plan.Konflikte) != 1 ||
		plan.Konflikte[0].Art != "zualt" || plan.Konflikte[0].Verlangt != "3" {
		t.Errorf("das DTO: %+v", plan)
	}

	// **Beim Entfernen und Zurück gibt es keine Quelle** — auch wenn eine
	// Fläche eine mitschickt.
	fuehrer.plan.Vorgang = service.PaketvorgangEntfernen
	if entfernen := anwendung.PaketPlan(t.Context(), "entfernen", "mse-game", "probe", "/irgendwo"); entfernen.Plan == nil ||
		entfernen.Plan.Quelle != "" || fuehrer.geplant != "entfernen mse-game probe" {
		t.Errorf("entfernen: %+v", entfernen)
	}

	if zurueck := anwendung.PaketPlan(t.Context(), "zurueck", "mse-game", "probe", ""); zurueck.Plan == nil ||
		fuehrer.geplant != "zurueck mse-game probe" {
		t.Errorf("zurück: %+v", zurueck)
	}

	if krumm := anwendung.PaketPlan(t.Context(), "aufraeumen", "", "", ""); krumm.Fehler == nil || krumm.Fehler.Feld != "vorgang" {
		t.Errorf("ein Vorgang, den es nicht gibt: %+v", krumm)
	}

	fuehrer.fehler = domain.ErrNotFound
	if fehlt := anwendung.PaketPlan(t.Context(), "entfernen", "mse-game", "weg", ""); fehlt.Fehler == nil ||
		fehlt.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("ein fehlendes Paket: %+v", fehlt)
	}
}

func TestErsetzenUndZurueckReichenDurch(t *testing.T) {
	fuehrer := &paketfuehrer{}
	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	if ersetzt := anwendung.PaketErsetzen(t.Context(), "/quelle", true); ersetzt.Fehler != nil ||
		ersetzt.Paket == nil || fuehrer.installiert != "/quelle (bestätigt)" {
		t.Errorf("ersetzen: %+v, %q", ersetzt, fuehrer.installiert)
	}

	if zurueck := anwendung.PaketZurueck(t.Context(), "mse-style", "fremd"); zurueck.Fehler != nil ||
		zurueck.Paket == nil || fuehrer.entfernt != "zurück mse-style fremd" {
		t.Errorf("zurück: %+v", zurueck)
	}

	fuehrer.fehler = errors.Join(domain.ErrConflict)
	if klemmt := anwendung.PaketErsetzen(t.Context(), "/quelle", false); klemmt.Fehler == nil {
		t.Errorf("der Konflikt ging verloren: %+v", klemmt)
	}
}

// **Der Paketweg der Flächen ohne Fenster geht denselben Weg** — Plan,
// Ersetzen und Zurück reichen durch, und die Auswahlliste der Vorgaben
// bekommt die Pakete ohne Übersicht.
func TestDerPaketwegPlantErsetztUndSetztZurueck(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, nahtfehler := app.PaketnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	fuehrer := &paketfuehrer{
		plan:   service.Paketplan{Vorgang: service.PaketvorgangZurueck, Paket: "probe.mse-game"},
		pakete: []domain.Paket{{Art: "mse-game", Name: "probe"}},
	}
	weg.MitVerwalter(fuehrer)

	if plan, err := weg.Plan(t.Context(), "zurueck", "mse-game", "probe", ""); err != nil ||
		plan.Vorgang != "zurueck" || fuehrer.geplant != "zurueck mse-game probe" {
		t.Errorf("der Plan: %+v, %v", plan, err)
	}

	if paket, err := weg.Ersetzen(t.Context(), "/q", true); err != nil ||
		paket.Name != "ersetzt" || fuehrer.installiert != "/q (bestätigt)" {
		t.Errorf("ersetzen: %+v, %v", paket, err)
	}

	if paket, err := weg.Zurueck(t.Context(), "mse-game", "probe"); err != nil ||
		paket.Name != "probe" || fuehrer.entfernt != "zurück mse-game probe" {
		t.Errorf("zurück: %+v, %v", paket, err)
	}

	if pakete, err := weg.Pakete(t.Context()); err != nil || len(pakete) != 1 {
		t.Errorf("die Pakete: %+v, %v", pakete, err)
	}

	weg.MitVerwalter(&paketfuehrer{auffrischfehler: errors.New("klemmt")})

	if _, err := weg.Pakete(t.Context()); err == nil {
		t.Error("ein Fehler beim Auffrischen ging verloren")
	}
}
