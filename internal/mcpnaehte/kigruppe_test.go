package mcpnaehte

// **Die atomare Gruppe an der MCP-Naht** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// **Die fachliche Zusage prüft `internal/flaechen`** — dort geht ein Ruf
// durch die ganze Kette. Hier steht die Frage davor: dass die Naht die Wahl
// übersetzt und den Fehler durchreicht, statt still nichts zu tun.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Einen Auftrag, den es nicht gibt, reicht sie als Fehler durch.**
func TestDieGruppennahtReichtIhrenFehlerDurch(t *testing.T) {
	naht := &preise{db: frischerBestand(t)}

	if _, fehler := naht.KIGruppe(probekontext(t), mcp.KIGruppenbezug{
		Wahl: []mcp.KIGruppenwahlsicht{
			{Kennung: "gibtesnicht", Felder: []string{"name"}},
		},
	}); fehler == nil {
		t.Fatal("der erfundene Auftrag kam durch")
	}
}

// **Und sie hält ihre Anwendung fest — über mehrere Rufe.**
//
// # Warum das geprüft wird
//
// **Weil daran die Kennungen des Batches hängen.** `service.KIAuftragswerk`
// führt seine offenen Aufträge im Arbeitsspeicher; eine Anwendung je Ruf
// hieße, dass `schreiben.ki.batch` Kennungen zurückgibt, die beim nächsten
// Ruf nicht mehr existieren — genau der Befund, den A13.1b behoben hat.
func TestDieNahtHaeltIhreKIAnwendung(t *testing.T) {
	naht := &preise{db: frischerBestand(t)}

	erste := naht.kianwendung()
	if erste == nil {
		t.Fatal("die Anwendung entstand nicht")
	}

	if zweite := naht.kianwendung(); zweite != erste {
		t.Error("der zweite Ruf baute eine neue Anwendung")
	}
}
