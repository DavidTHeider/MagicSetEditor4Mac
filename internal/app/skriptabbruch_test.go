package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Skriptkonsole mit Abbruch, Dauer und einem Lauf je Karte** —
// Werkstück U13 von Phase 5.1
// (#1424).

// haengendeKonsole ist ein Double, das so lange rechnet, bis jemand abbricht
// — so wie ein Ausdruck, der nicht zu Ende kommt.
type haengendeKonsole struct {
	laeuft chan struct{}
}

func (k *haengendeKonsole) Namen(context.Context, int64) ([]string, error) { return nil, nil }

func (k *haengendeKonsole) Auswerten(ctx context.Context, _ int64, ausdruck string) (app.KonsolenDTO, error) {
	if ausdruck == "sofort" {
		return app.KonsolenDTO{Quelltext: "1", Text: "1", Art: "Zahl"}, nil
	}

	close(k.laeuft)
	<-ctx.Done()

	// So wickelt `service.Skriptkonsole` den Fehler des Auswerters ein:
	// `ErrValidation` außen, der Grund innen.
	return app.KonsolenDTO{}, fmt.Errorf("%w: %w", domain.ErrValidation, ctx.Err())
}

func TestEinLaufendesSkriptLaesstSichAbbrechen(t *testing.T) {
	t.Parallel()

	konsole := &haengendeKonsole{laeuft: make(chan struct{})}
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(konsole)

	antworten := make(chan app.KonsolenAntwort, 1)
	go func() { antworten <- anwendung.Skript(t.Context(), "7", "endlos") }()

	<-konsole.laeuft

	if abbruch := anwendung.SkriptAbbrechen(t.Context(), "7"); abbruch.Fehler != nil {
		t.Fatalf("Abbrechen: %+v", abbruch)
	}

	select {
	case antwort := <-antworten:
		// **Abgebrochen ist kein Fehler**: es ist, was jemand wollte.
		if !antwort.Abgebrochen || antwort.Fehler != nil {
			t.Errorf("nach dem Abbruch: %+v", antwort)
		}

		if antwort.Befunde == nil {
			t.Error("die Befunde sind nil und kämen als null an")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("der Abbruch kam beim Lauf nicht an")
	}

	// Danach läuft an derselben Karte wieder etwas.
	if antwort := anwendung.Skript(t.Context(), "7", "sofort"); antwort.Fehler != nil || antwort.Quelltext != "1" {
		t.Errorf("nach dem Abbruch: %+v", antwort)
	}
}

// **Ein Lauf je Karte** (§7, Regel 4) — und eine andere Karte ist ein anderer
// Datensatz.
func TestEinZweiterLaufAnDerselbenKarteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	konsole := &haengendeKonsole{laeuft: make(chan struct{})}
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(konsole)

	antworten := make(chan app.KonsolenAntwort, 1)
	go func() { antworten <- anwendung.Skript(t.Context(), "7", "endlos") }()

	<-konsole.laeuft

	zweiter := anwendung.Skript(t.Context(), "7", "sofort")
	if zweiter.Fehler == nil || zweiter.Fehler.Code != app.CodeKonflikt {
		t.Errorf("der zweite Lauf an Karte 7: %+v", zweiter)
	}

	if andere := anwendung.Skript(t.Context(), "8", "sofort"); andere.Fehler != nil {
		t.Errorf("Karte 8 wartet auf Karte 7: %+v", andere)
	}

	anwendung.SkriptAbbrechen(t.Context(), "7")
	<-antworten
}

func TestAbbrechenOhneLaufIstNichtsUndEineFalscheKennungEinBefund(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	if antwort := anwendung.SkriptAbbrechen(t.Context(), "7"); antwort.Fehler != nil {
		t.Errorf("ohne Lauf: %+v", antwort)
	}

	if antwort := anwendung.SkriptAbbrechen(t.Context(), "sieben"); antwort.Fehler == nil {
		t.Error("eine Kennung, die keine ist, ging durch")
	}
}

// **Die Antwort trägt ihre Dauer** — und zwar auch serialisiert, denn das
// Fenster zeigt sie.
func TestDieKonsolenantwortTraegtIhreDauer(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(&haengendeKonsole{laeuft: make(chan struct{})})

	roh, err := json.Marshal(anwendung.Skript(t.Context(), "7", "sofort"))
	if err != nil {
		t.Fatal(err)
	}

	for _, feld := range []string{`"dauerMs":`, `"abgebrochen":false`} {
		if !strings.Contains(string(roh), feld) {
			t.Errorf("%s fehlt in %s", feld, roh)
		}
	}
}

// **Ein Abbruch von außen** — der Kontext des Aufrufers endet — sieht für die
// Antwort genauso aus wie der Knopf.
func TestEinEndenderKontextIstEinAbbruch(t *testing.T) {
	t.Parallel()

	konsole := &haengendeKonsole{laeuft: make(chan struct{})}
	anwendung := app.NeueApp("")
	anwendung.MitSkriptkonsole(konsole)

	ctx, abbrechen := context.WithCancel(t.Context())
	antworten := make(chan app.KonsolenAntwort, 1)
	go func() { antworten <- anwendung.Skript(ctx, "7", "endlos") }()

	<-konsole.laeuft
	abbrechen()

	antwort := <-antworten
	if !antwort.Abgebrochen || antwort.Fehler != nil {
		t.Errorf("%+v", antwort)
	}
}

// Die Bindung reicht den Ruf durch — mit dem Kontext des Fensters.
func TestDieBindungReichtDenSkriptabbruchDurch(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueBindungen(app.NeueApp("")).SkriptAbbrechen(t.Context(), "7"); antwort.Fehler != nil {
		t.Errorf("%+v", antwort)
	}
}
