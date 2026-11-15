package app_test

// **Begrenzte Metriken** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne `t.Parallel()`**: das Messwerk gehört dem Prozess, und zwei Tests,
// die es nebeneinander leeren, prüften die Zahlen des jeweils anderen.
func TestDieBindungGibtDenZaehlstandUndSeineGrenze(t *testing.T) {
	service.MesswerkZuruecksetzen()

	service.Zaehle(service.Messpunkt{
		Flaeche: domain.ModusFenster, Wirkung: domain.WirkungSchreiben,
		Ausgang: domain.AusgangGewaehrt,
	})

	antwort := app.NeueBindungen(app.NeueApp("")).Metriken(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("die Bindung gab einen Fehler: %+v", antwort.Fehler)
	case antwort.Grenze != service.Messreihengrenze:
		t.Errorf("Grenze = %d, erwartet %d", antwort.Grenze, service.Messreihengrenze)
	case len(antwort.Reihen) != 1:
		t.Fatalf("Reihen = %+v", antwort.Reihen)
	}

	reihe := antwort.Reihen[0]
	if reihe.Flaeche != "fenster" || reihe.Wirkung != "lokalSchreiben" ||
		reihe.Ausgang != "gewaehrt" || reihe.Anzahl != 1 {
		t.Errorf("Reihe = %+v", reihe)
	}
}

// **Kein Zähler heißt keine Reihen — und keine `null`.** Eine leere Liste, die
// als `null` ankäme, ließe die Seite über `undefined` stolpern (§3).
func TestOhneWirkungGibtEsEineLeereListe(t *testing.T) {
	service.MesswerkZuruecksetzen()

	antwort := app.NeueApp("").Metriken(t.Context())
	if antwort.Reihen == nil || len(antwort.Reihen) != 0 {
		t.Fatalf("Reihen = %#v", antwort.Reihen)
	}
}

// **Ein abgebrochener Ruf bekommt keine Zahlen** (§7, Regel 2).
func TestEinAbgebrochenerRufBekommtKeineZahlen(t *testing.T) {
	service.MesswerkZuruecksetzen()

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	antwort := app.NeueApp("").Metriken(ctx)

	switch {
	case antwort.Fehler == nil:
		t.Fatal("der Abbruch wurde nicht gemeldet")
	case antwort.Fehler.Code != app.CodeAbgebrochen:
		t.Errorf("Code = %q", antwort.Fehler.Code)
	case antwort.Reihen == nil:
		t.Error("auch im Abbruch gehört eine leere Liste in die Antwort")
	}
}
