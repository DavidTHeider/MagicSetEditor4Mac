package app_test

// **Die Parameterhilfen an der Bindungsfassade** — Werkstück D von Phase 4.11
// (#1342).

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probehilfen sind zwei Arten und ein Verweis.
func probehilfen() service.Parameterhilfen {
	return service.Parameterhilfen{
		Arten: []service.Parametereinfuegung{{
			Name: "zahl", Beschreibung: "eine Zahl",
			Einfuegen: "<atom-param>zahl</atom-param>",
		}},
		Verweise: []service.Parameterverweis{{
			Nummer: 1, Art: "zahl", Einfuegen: "{param1}",
		}},
		Befunde: []domain.Befund{{Ort: "mana", Meldung: "das Skript lief nicht"}},
	}
}

// **Beide Menüs kommen durch — samt Befunden.**
func TestDieParameterhilfenKommenAnDieOberflaeche(t *testing.T) {
	t.Parallel()

	probe := &wortpflegeprobe{hilfen: probehilfen()}

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselwoertern(probe)

	antwort := anwendung.Parameterhilfen(t.Context(), "7", "3")
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if probe.setID != 7 || probe.wortID != 3 {
		t.Errorf("gefragt wurde nach %d/%d", probe.setID, probe.wortID)
	}

	if len(antwort.Arten) != 1 ||
		antwort.Arten[0].Einfuegen != "<atom-param>zahl</atom-param>" {
		t.Errorf("die Arten lauten %+v", antwort.Arten)
	}

	if len(antwort.Verweise) != 1 || antwort.Verweise[0].Nummer != 1 {
		t.Errorf("die Verweise lauten %+v", antwort.Verweise)
	}

	if len(antwort.Befunde) != 1 {
		t.Errorf("die Befunde lauten %+v", antwort.Befunde)
	}
}

// **Ohne Naht sagt die Bindung es** — und gibt drei leere Listen.
func TestOhneNahtSagtDieParameterbindungEs(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Parameterhilfen(probekontext(t), "7", "3")

	if antwort.Fehler == nil {
		t.Error("ohne Naht kamen Hilfen zurück")
	}

	if antwort.Arten == nil || antwort.Verweise == nil || antwort.Befunde == nil {
		t.Errorf("eine der Listen ist nil: %+v", antwort)
	}
}

// **Beide Kennungen werden geprüft.**
func TestDieParameterbindungPrueftBeideKennungen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselwoertern(&wortpflegeprobe{})

	if antwort := anwendung.Parameterhilfen(
		probekontext(t), "keine Zahl", "3"); antwort.Fehler == nil {
		t.Error("die Setkennung ging durch")
	}

	if antwort := anwendung.Parameterhilfen(
		probekontext(t), "7", "keine Zahl"); antwort.Fehler == nil {
		t.Error("die Wortkennung ging durch")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerParameterhilfenKommtDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselwoertern(&wortpflegeprobe{
		fehler: fmt.Errorf("%w: kein solches Wort", domain.ErrNotFound),
	})

	antwort := anwendung.Parameterhilfen(t.Context(), "7", "3")
	if antwort.Fehler == nil || !strings.Contains(antwort.Fehler.Meldung, "Wort") {
		t.Errorf("der Fehler lautet %+v", antwort.Fehler)
	}
}
