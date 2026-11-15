package service_test

// **Die Stufen und die Schlüsselklassen** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Aus `spieltextur_test.go` herausgelöst**, als es über die Zeilengrenze aus
// §2 ging. Der Schnitt folgt der Frage: dort, **was** der Dienst tut, hier,
// **woran** er ein Bild wiedererkennt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Fünf benannte Stufen, und jede hat eine Breite.**
func TestFuenfBenannteTexturgueten(t *testing.T) {
	t.Parallel()

	gueten := service.Texturgueten()

	if len(gueten) != 5 {
		t.Fatalf("%d Stufen, erwartet fünf: %v", len(gueten), gueten)
	}

	// **Aufsteigend** — sonst hieße „niedrig" nichts.
	vorher := 0

	for _, guete := range gueten {
		breite := guete.Breite()
		if breite <= vorher {
			t.Errorf("%s ist %d breit, die Stufe davor %d", guete, breite, vorher)
		}

		vorher = breite
	}

	if !service.IstTexturguete(service.GueteTisch) ||
		service.IstTexturguete("riesig") {
		t.Error("IstTexturguete trennt nicht")
	}
}

// **Jede Schlüsselklasse verwirft das Bild.**
func TestJedeSchluesselklasseVerwirftDasBild(t *testing.T) {
	t.Parallel()

	for name, anders := range map[string]func(
		*planprobe, *ressourcenprobe, *service.Texturwunsch,
	){
		"die Kartenseite": func(_ *planprobe, _ *ressourcenprobe, w *service.Texturwunsch) {
			w.Seite = string(service.SeiteZweite)
		},
		"die Zahl der Gesichter": func(_ *planprobe, _ *ressourcenprobe, w *service.Texturwunsch) {
			w.Gesichter = 2
		},
		"die Güte": func(_ *planprobe, _ *ressourcenprobe, w *service.Texturwunsch) {
			w.Guete = service.GueteBetrachten
		},
		"der Stil": func(p *planprobe, _ *ressourcenprobe, _ *service.Texturwunsch) {
			p.plan = texturplan("andererrahmen.png", "MPlantin")
		},
		"das Bild": func(_ *planprobe, r *ressourcenprobe, _ *service.Texturwunsch) {
			r.bilder["rahmen.png"] = "bild-2"
		},
		"die Schrift": func(_ *planprobe, r *ressourcenprobe, _ *service.Texturwunsch) {
			r.schriften["MPlantin"] = "schrift-2"
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plaene := &planprobe{plan: texturplan("rahmen.png", "MPlantin")}
			quelle := ressourcenprobe{
				bilder:    map[string]string{"rahmen.png": "bild-1"},
				schriften: map[string]string{"MPlantin": "schrift-1"},
			}
			wunsch := service.Texturwunsch{
				Seite: string(service.SeiteVorne),
				Guete: service.GueteTisch, Gesichter: 1,
			}

			vorher := service.Texturschluessel(plaene.plan, quelle, wunsch)

			anders(plaene, &quelle, &wunsch)

			nachher := service.Texturschluessel(plaene.plan, quelle, wunsch)

			if vorher == nachher {
				t.Errorf("%s ändert den Schlüssel nicht", name)
			}
		})
	}
}

// **Eine fehlende Datei gibt einen anderen Schlüssel.**
//
// # Warum das die richtige Antwort ist und kein Fehler
//
// **Weil die Karte trotzdem gezeichnet wird** — mit einem Befund daneben. Ein
// Fehler hier machte aus einem erklärten Mangel einen Abbruch, und die Karte
// verschwände ganz. Was zählt, ist, dass der Schlüssel ein **anderer** ist:
// taucht die Datei auf, wird neu gezeichnet.
func TestEineFehlendeDateiGibtEinenAnderenSchluessel(t *testing.T) {
	t.Parallel()

	plan := texturplan("rahmen.png", "MPlantin")
	wunsch := service.Texturwunsch{
		Seite: string(service.SeiteVorne),
		Guete: service.GueteTisch, Gesichter: 1,
	}

	vollstaendig := service.Texturschluessel(plan, ressourcenprobe{
		bilder:    map[string]string{"rahmen.png": "bild-1"},
		schriften: map[string]string{"MPlantin": "schrift-1"},
	}, wunsch)

	ohneBild := service.Texturschluessel(plan, ressourcenprobe{
		schriften: map[string]string{"MPlantin": "schrift-1"},
	}, wunsch)

	ohneSchrift := service.Texturschluessel(plan, ressourcenprobe{
		bilder: map[string]string{"rahmen.png": "bild-1"},
	}, wunsch)

	for name, schluessel := range map[string]string{
		"das fehlende Bild":    ohneBild,
		"die fehlende Schrift": ohneSchrift,
	} {
		if schluessel == vollstaendig {
			t.Errorf("%s ändert den Schlüssel nicht", name)
		}
	}

	if ohneBild == ohneSchrift {
		t.Error("ein fehlendes Bild und eine fehlende Schrift geben denselben Schlüssel")
	}
}

// **Eine unbekannte Güte wird abgewiesen, statt still die kleinste zu nehmen.**
func TestEineUnbekannteGueteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	_, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		service.SeiteVorne, "riesig")

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet eine Abweisung", fehler)
	}
}

// **Und eine unbekannte Seite ebenso.**
func TestEineUnbekannteSeiteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk, _, _, _ := eineWerkstatt(t)

	_, fehler := werk.Textur(t.Context(), eineTexturkarte(1),
		"schraeg", service.GueteTisch)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet eine Abweisung", fehler)
	}
}
