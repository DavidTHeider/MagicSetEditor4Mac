package service_test

import (
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne `t.Parallel()`, und das ist der Punkt**: das Messwerk ist ein Wert
// des Prozesses (siehe dort). Zwei Tests, die es nebeneinander leeren, prüften
// die Zahlen des jeweils anderen.
func TestDasMesswerkZaehltNachFlaecheWirkungUndAusgang(t *testing.T) {
	service.MesswerkZuruecksetzen()

	service.Zaehle(service.Messpunkt{
		Flaeche: domain.ModusDienst, Wirkung: domain.WirkungSchreiben,
		Ausgang: domain.AusgangGewaehrt,
	})
	service.Zaehle(service.Messpunkt{
		Flaeche: domain.ModusDienst, Wirkung: domain.WirkungSchreiben,
		Ausgang: domain.AusgangGewaehrt,
	})
	service.Zaehle(service.Messpunkt{
		Flaeche: domain.ModusFenster, Wirkung: domain.WirkungLesen,
		Ausgang: domain.AusgangAbgewiesen,
	})

	stand := service.Messstand()
	if len(stand) != 2 {
		t.Fatalf("Reihen = %d, erwartet 2: %+v", len(stand), stand)
	}

	// **Feste Reihenfolge**: „dienst" steht vor „fenster".
	if stand[0].Flaeche != domain.ModusDienst || stand[0].Anzahl != 2 {
		t.Errorf("erste Reihe = %+v", stand[0])
	}

	if stand[1].Flaeche != domain.ModusFenster || stand[1].Anzahl != 1 {
		t.Errorf("zweite Reihe = %+v", stand[1])
	}
}

// **Die Zusage des Werkstücks**: tausend verschiedene Objekte erzeugen keine
// tausend Reihen. Gezählt wird mit einer Kartenkennung in jeder der drei
// Dimensionen — genau das, was das Issue verbietet.
func TestVieleObjekteErzeugenKeineNeuenReihen(t *testing.T) {
	service.MesswerkZuruecksetzen()

	for nummer := range 1000 {
		kennung := fmt.Sprintf("karte-%d", nummer)

		service.Zaehle(service.Messpunkt{
			Flaeche: domain.Dienstmodus(kennung), Wirkung: domain.WirkungLesen,
			Ausgang: domain.AusgangGewaehrt,
		})
		service.Zaehle(service.Messpunkt{
			Flaeche: domain.ModusMCP, Wirkung: domain.Wirkung(kennung),
			Ausgang: domain.AusgangGewaehrt,
		})
		service.Zaehle(service.Messpunkt{
			Flaeche: domain.ModusMCP, Wirkung: domain.WirkungLesen,
			Ausgang: domain.Ausgang(kennung),
		})
	}

	if stand := service.Messstand(); len(stand) != 0 {
		t.Fatalf("Reihen = %d, erwartet 0: %+v", len(stand), stand)
	}
}

// **Und die Obergrenze ist wirklich vierzig** — alle vierzig Punkte einmal,
// und danach ist keiner mehr übrig.
func TestAlleFlaechenWirkungenUndAusgaengeErgebenJedeReihe(t *testing.T) {
	service.MesswerkZuruecksetzen()

	for _, flaeche := range domain.Dienstmodi() {
		for _, wirkung := range domain.Wirkungen() {
			for _, ausgang := range domain.Ausgaenge() {
				service.Zaehle(service.Messpunkt{
					Flaeche: flaeche, Wirkung: wirkung, Ausgang: ausgang,
				})
			}
		}
	}

	if stand := service.Messstand(); len(stand) != service.Messreihengrenze {
		t.Fatalf("Reihen = %d, erwartet %d", len(stand), service.Messreihengrenze)
	}
}
