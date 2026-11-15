package storage_test

// **Die Einstellungen des Menschen zum gezeigten Profil** — Werkstück 9 von
// Phase 4.12
// (#1351),
// Punkt 9.
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket: sie teilen eine
// Datenquelle im Arbeitsspeicher.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Abfrage findet den Eigner des gezeigten Profils.**
//
// **Das ist die ganze Zusage**: der Aufrufer nennt kein Profil und keinen
// Nutzer — er reicht den Kontext, und der Bestand geht den Weg vom Profil zu
// seinem Menschen.
func TestDieNutzerwerteFolgenDemGezeigtenProfil(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if err := ablage.NutzereinstellungSetzen(
		ctx, 1, domain.EinstellungErscheinungsbild, "dunkel"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	werte, fehler := ablage.NutzerwerteZumProfil(ctx)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if werte[domain.EinstellungErscheinungsbild] != "dunkel" {
		t.Errorf("gelesen wurde %v", werte)
	}
}

// **Ohne Profil im Kontext gibt es keine Antwort.**
//
// **Und das ist ein Fehler und keine leere Abbildung**: eine leere sähe aus
// wie „der Mensch hat nichts gewählt", und dann gälten die Vorgaben, obwohl
// niemand nachgesehen hat (§8).
func TestOhneProfilGibtEsKeineNutzerwerte(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.NutzerwerteZumProfil(
		t.Context()); fehler == nil { //nolint:usetesting // ohne Profil ist der Fall
		t.Error("die Abfrage kam ohne Profil durch")
	}
}
