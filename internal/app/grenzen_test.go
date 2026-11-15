package app_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Grenzen dieses Prozesses** — Werkstück O5 von Phase 5.2 (#1429).

func TestDieGrenzenKommenAusDemMaschinenhaushalt(t *testing.T) {
	t.Parallel()

	vorgabe := domain.Vorgabegrenzen()
	grenzen := app.GrenzenauskunftVon(vorgabe)

	if grenzen.Parallel["rendern"] != vorgabe.Parallel[domain.RessourceRendern] ||
		grenzen.Parallel["einfuhr"] != vorgabe.Parallel[domain.RessourceEinfuhr] ||
		grenzen.AntwortBytes != vorgabe.Antwort ||
		grenzen.Principal != (app.KontingentDTO{Burst: vorgabe.Principal.Burst, Nachfuellen: vorgabe.Principal.Nachfuellen, FensterSekunden: 60}) ||
		grenzen.Global.Burst != vorgabe.Global.Burst || grenzen.Reserve != vorgabe.Reserve ||
		grenzen.Kosten["lesen"] != 1 || grenzen.Kosten["masse"] != domain.KostenMasse.Kosten() ||
		grenzen.Kontingente != domain.KontingenteAn {
		t.Errorf("Grenzen: %+v", grenzen)
	}

	// Jede Kostenklasse steht drin — und keine, die es nicht gibt.
	if len(grenzen.Kosten) != len(domain.Kostenklassen()) {
		t.Errorf("Kosten: %v", grenzen.Kosten)
	}

	anwendung := app.NeueApp("")
	if antwort := anwendung.Grenzen(t.Context()); antwort.Fehler != nil || antwort.Grenzen.Reserve != vorgabe.Reserve ||
		antwort.Grenzen.Kontingente != domain.KontingenteAn {
		t.Errorf("App: %+v", antwort)
	}

	if antwort := app.NeueBindungen(anwendung).Grenzen(t.Context()); antwort.Fehler != nil || antwort.Grenzen.Principal.Burst != vorgabe.Principal.Burst {
		t.Errorf("Fensterbindung: %+v", antwort)
	}
}

// **Auftrags- und Verbindungsgrenze und die Kostentafel stehen in der
// Auskunft** — #2003. Bis zum 2026-10-02 nannte keine Fläche die vier
// Aufträge und die sechzehn Verbindungen, und was ein Ruf kostet, stand nur
// je Klasse da.
func TestDieGrenzenNennenAuftraegeVerbindungenUndKostentafel(t *testing.T) {
	t.Parallel()

	grenzen := app.GrenzenauskunftVon(domain.Vorgabegrenzen())

	if grenzen.Verbindungen != 16 ||
		grenzen.Auftraege != (app.AuftragsgrenzenDTO{Laufend: 4, Behalten: 16, AufbewahrungSekunden: 3600}) {
		t.Errorf("Verbindungen %d, Aufträge %+v", grenzen.Verbindungen, grenzen.Auftraege)
	}

	if len(grenzen.Kostentafel) != len(domain.Kostentafel()) ||
		grenzen.Kostentafel[0].Faehigkeit != "Kartenbilder schreiben" || grenzen.Kostentafel[0].Kosten != 5 {
		t.Errorf("Kostentafel: %+v", grenzen.Kostentafel)
	}

	roh, fehler := json.Marshal(grenzen)
	if fehler != nil || !strings.Contains(string(roh), `"verbindungen":16`) ||
		!strings.Contains(string(roh), `"auftraege":{"laufend":4,"behalten":16,"aufbewahrungSekunden":3600}`) ||
		strings.Contains(string(roh), `null`) {
		t.Errorf("JSON: %s, %v", roh, fehler)
	}
}
