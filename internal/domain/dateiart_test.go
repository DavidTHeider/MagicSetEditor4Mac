package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestJedeDateiHatEineSichereAnsicht hält UX-13.AK-01
// (#1685): für die sieben MSE-Paketarten und die häufigen Bild-, Text- und
// PDF-Ausgaben steht fest, welche Vorschau sinnvoll **und sicher** ist — und
// eine unbekannte Datei zeigt nur, was über sie bekannt ist.
func TestJedeDateiHatEineSichereAnsicht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		art      domain.Dateiart
		vorschau domain.Vorschauart
		oeffnen  bool
	}{
		{"kraftwerk.mse-set", domain.DateiartPaket, domain.VorschauListe, false},
		{"magic.mse-game", domain.DateiartPaket, domain.VorschauListe, false},
		{"magic-m15.mse-style", domain.DateiartPaket, domain.VorschauListe, false},
		{"symbole.mse-symbol-font", domain.DateiartPaket, domain.VorschauListe, false},
		{"teile.mse-include", domain.DateiartPaket, domain.VorschauListe, false},
		{"de.mse-locale", domain.DateiartPaket, domain.VorschauListe, false},
		{"html.mse-export-template", domain.DateiartPaket, domain.VorschauListe, false},
		{"Karte.PNG", domain.DateiartBild, domain.VorschauBild, true},
		{"karte.jpeg", domain.DateiartBild, domain.VorschauBild, true},
		{"karte.webp", domain.DateiartBild, domain.VorschauBild, true},
		{"karte.tiff", domain.DateiartBild, domain.VorschauBild, true},
		{"karte.svg", domain.DateiartVektor, domain.VorschauBild, true},
		{"bogen.pdf", domain.DateiartPDF, domain.VorschauKeine, true},
		{"liste.txt", domain.DateiartText, domain.VorschauText, true},
		{"karten.json", domain.DateiartText, domain.VorschauText, true},
		{"sammlung.csv", domain.DateiartText, domain.VorschauText, true},
		// **HTML als Text und nie ausgeführt** — auch nicht über „Öffnen":
		// ein Browser führte seine Skripte aus.
		{"set.html", domain.DateiartText, domain.VorschauText, false},
		{"stand.msesicherung", domain.DateiartArchiv, domain.VorschauListe, false},
		{"tausch.msepackage", domain.DateiartArchiv, domain.VorschauListe, false},
		{"paket.zip", domain.DateiartArchiv, domain.VorschauListe, false},
		// **Was ausgeführt werden könnte, bekommt keine Vorschau und kein
		// Öffnen** — nur, was über die Datei bekannt ist.
		{"skript.command", domain.DateiartUnbekannt, domain.VorschauKeine, false},
		{"Programm.app", domain.DateiartUnbekannt, domain.VorschauKeine, false},
		{"ohne-endung", domain.DateiartUnbekannt, domain.VorschauKeine, false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			regel := domain.AnsichtsregelFuer(fall.name)
			if regel.Art != fall.art || regel.Vorschau != fall.vorschau || regel.Oeffnen != fall.oeffnen {
				t.Errorf("%+v", regel)
			}

			if regel.Vorschau != domain.VorschauKeine && regel.Hoechstens <= 0 {
				t.Errorf("eine Vorschau ohne Grenze: %+v", regel)
			}
		})
	}
}
