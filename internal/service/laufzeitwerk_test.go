package service_test

// **Wie lange ein Ruf dauert, als Verteilung** — OPS-05.AK-01
// (#1722):
// „Metriken umfassen Latenzverteilung, Fehlerrate […] mit nachvollziehbaren
// Einheiten." Eimer in Millisekunden, je Fläche und Art — eine feste Zahl
// von Reihen, gleich wie viele Rufe es gibt.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieLaufzeitenFallenInIhreEimer(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	for _, punkt := range []service.Laufzeitpunkt{
		{Flaeche: domain.ModusDienst, Dauer: 500 * time.Microsecond, Ausgang: domain.RufGelungen},
		{Flaeche: domain.ModusDienst, Dauer: 3 * time.Millisecond},
		{Flaeche: domain.ModusDienst, Dauer: 3 * time.Second, Ausgang: domain.RufGescheitert},
		{Flaeche: domain.ModusDienst, Schreibend: true, Dauer: 20 * time.Millisecond, Ausgang: domain.RufAbgewiesen},
		{Flaeche: domain.ModusMCP, Dauer: 100 * time.Millisecond},
		{Flaeche: "fremd", Dauer: time.Millisecond},
	} {
		service.ErfasseLaufzeit(punkt)
	}

	stand := service.Laufzeitstand()
	if len(stand) != 3 {
		t.Fatalf("%d Reihen: %+v", len(stand), stand)
	}

	lesend := stand[0]
	if lesend.Flaeche != domain.ModusDienst || lesend.Art != service.ArtLesend || lesend.Anzahl != 3 ||
		lesend.Gescheitert != 1 || lesend.Abgewiesen != 0 {
		t.Errorf("lesend: %+v", lesend)
	}

	// **Eimer sind obere Grenzen**: 0,5 ms in ≤ 1, 3 ms in ≤ 5, 3 s im
	// offenen letzten.
	if lesend.Eimer[0] != 1 || lesend.Eimer[1] != 1 || lesend.Eimer[len(lesend.Eimer)-1] != 1 {
		t.Errorf("Eimer: %v", lesend.Eimer)
	}

	if len(lesend.Eimer) != len(service.Laufzeiteimer)+1 || lesend.SummeMs < 3003 {
		t.Errorf("Eimer %d, Summe %v", len(lesend.Eimer), lesend.SummeMs)
	}

	if schreibend := stand[1]; schreibend.Art != service.ArtSchreibend || schreibend.Abgewiesen != 1 {
		t.Errorf("schreibend: %+v", schreibend)
	}

	if stand[2].Flaeche != domain.ModusMCP {
		t.Errorf("MCP: %+v", stand[2])
	}

	if service.Laufzeitreihengrenze != len(domain.Dienstmodi())*2 {
		t.Errorf("Grenze %d", service.Laufzeitreihengrenze)
	}
}
