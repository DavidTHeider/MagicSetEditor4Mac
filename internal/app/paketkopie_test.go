// **Eine eigene Kopie über die Fassade** — FMT-02.AK-04
// (#1691).
package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// paketkopiererprobe kopiert — oder scheitert.
type paketkopiererprobe struct {
	paketfuehrer

	fehler  error
	gefragt []string
}

func (p *paketkopiererprobe) Kopieren(_ context.Context, art, name, neu string) (domain.Paket, error) {
	p.gefragt = []string{art, name, neu}
	if p.fehler != nil {
		return domain.Paket{}, p.fehler
	}

	return domain.Paket{Art: art, Name: name + "-kopie", Pfad: "/pakete/" + name + "-kopie." + art}, nil
}

func TestDieFassadeGibtDieKopie(t *testing.T) {
	probe := &paketkopiererprobe{}
	anwendung := app.NeueApp("")
	anwendung.MitPaketen(probe)

	antwort := anwendung.PaketKopieren(t.Context(), "mse-style", "probe-schlicht", "")
	if antwort.Fehler != nil || antwort.Paket == nil || antwort.Paket.Name != "probe-schlicht-kopie" ||
		antwort.Paket.Pfad != "/pakete/probe-schlicht-kopie.mse-style" {
		t.Fatalf("die Antwort: %+v, %+v", antwort.Paket, antwort.Fehler)
	}

	if probe.gefragt[0] != "mse-style" || probe.gefragt[1] != "probe-schlicht" || probe.gefragt[2] != "" {
		t.Errorf("gefragt: %v", probe.gefragt)
	}
	// **Die Hauptdatei kommt mit** — die Werkstatt öffnet sie in der Kopie,
	// und welche es ist, sagt die Paketart und nicht das Fenster.
	if antwort.Paket.Hauptdatei != "style" {
		t.Errorf("die Hauptdatei: %q", antwort.Paket.Hauptdatei)
	}
}

func TestDieFassadeNenntDenFehlerDerKopieAmFeld(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&paketkopiererprobe{fehler: domain.NewFieldError("name",
		domain.SchluesselPaketnameUngueltig, "Ein Paketname …")})

	antwort := anwendung.PaketKopieren(t.Context(), "mse-style", "probe-schlicht", "mein stil")
	if antwort.Paket != nil || antwort.Fehler == nil || antwort.Fehler.Feld != "name" ||
		antwort.Fehler.Schluessel != domain.SchluesselPaketnameUngueltig {
		t.Fatalf("die Antwort: %+v, %+v", antwort.Paket, antwort.Fehler)
	}
}

// **Der Flächenweg macht aus einem übernommenen Paket ein eigenes** — für
// Kommandozeile, MCP und Dienst. Die Saat liegt neben dem Programm und damit
// im Test nirgends; übernommen ist dieselbe Lage: fremd verwaltet.
func TestDerFlaechenwegKopiertEinUebernommenesPaket(t *testing.T) {
	weg, quelle := pflegeweg(t, "unbenutzt.mse-style")

	if _, fehler := weg.Installieren(probekontext(t), quelle); fehler != nil {
		t.Fatalf("einbinden: %v", fehler)
	}

	kopie, fehler := weg.Kopieren(probekontext(t), "mse-style", "unbenutzt", "")
	if fehler != nil || kopie.Name != "unbenutzt-kopie" {
		t.Fatalf("kopieren: %+v, %v", kopie, fehler)
	}

	liste, _, lesefehler := weg.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	gefunden := false

	for _, eines := range liste {
		if eines.Paket.Name == kopie.Name {
			gefunden = eines.Herkunftsart == string(domain.PaketEigen)
		}
	}

	if !gefunden {
		t.Errorf("die Kopie steht nicht als eigene in der Liste: %+v", liste)
	}
}
