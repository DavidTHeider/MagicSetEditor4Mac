package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bogenprobe ist ein Double des Bogenlegers.
type bogenprobe struct {
	lauf   *service.Bogenlauf
	fehler error

	// probelauf ist, was der Probelauf zurückgibt (#2015); probelaeufe zählt
	// seine Aufrufe.
	probelauf    *service.Bogenprobelauf
	probelaeufe  int
	probefehler  error
	probeAuswahl domain.Kartenauswahl

	setID   int64
	pfad    string
	masse   service.Bogenmasse
	auswahl domain.Kartenauswahl
}

func (b *bogenprobe) Lege(
	_ context.Context, setID int64, pfad string, masse service.Bogenmasse,
	auswahl domain.Kartenauswahl,
) (*service.Bogenlauf, error) {
	b.setID, b.pfad, b.masse, b.auswahl = setID, pfad, masse, auswahl

	return b.lauf, b.fehler
}

func (b *bogenprobe) Probelauf(
	_ context.Context, setID int64, masse service.Bogenmasse, auswahl domain.Kartenauswahl,
) (*service.Bogenprobelauf, error) {
	b.probelaeufe++
	b.setID, b.masse, b.probeAuswahl = setID, masse, auswahl

	return b.probelauf, b.probefehler
}

func mitBogenleger(probe *bogenprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitBogenleger(probe)

	return anwendung
}

// **Was die Oberfläche einstellt, kommt beim Leger an** — und was sie nicht
// einstellt, ist die Zusage von #50: 63 × 88 mm.
func TestEinBogenGehtMitDenGewaehltenMassenHinaus(t *testing.T) {
	probe := &bogenprobe{lauf: &service.Bogenlauf{
		Pfad: "/tmp/bogen.pdf", Seiten: 2, Karten: 10,
	}}

	antwort := mitBogenleger(probe).
		BogenLegen(probekontext(t), "7", "/tmp/bogen.pdf", "a4", 3, 300, true, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.setID != 7 || probe.pfad != "/tmp/bogen.pdf" {
		t.Errorf("beim Leger kam %d, %q an", probe.setID, probe.pfad)
	}

	if probe.masse.Kartenbreite != 63 || probe.masse.Kartenhoehe != 88 {
		t.Errorf("Kartenmaß %g × %g, erwartet 63 × 88",
			probe.masse.Kartenbreite, probe.masse.Kartenhoehe)
	}

	if probe.masse.Anschnitt != 3 || probe.masse.DPI != 300 || !probe.masse.Schnittmarken {
		t.Errorf("%+v", probe.masse)
	}

	if antwort.Seiten != 2 || antwort.Karten != 10 {
		t.Errorf("%+v", antwort)
	}
}

// **Das Papierformat kommt als Name, nicht als zwei Zahlen.**
//
// „A4" ist etwas, das jemand auswählt; 210 × 297 ist etwas, das er nachschlägt.
func TestDasPapierformatBestimmtDieSeitengroesse(t *testing.T) {
	faelle := map[string][2]float64{
		"a4":     {210, 297},
		"letter": {215.9, 279.4},

		// Ein unbekannter Name fällt auf A4 zurück, statt einen Bogen ohne Maß
		// zu bauen: die Auswahl der Oberfläche kennt nur zwei, und ein dritter
		// Wert wäre ein Fehler, den niemand herbeiführen kann.
		"papyrus": {210, 297},
	}

	for papier, erwartet := range faelle {
		t.Run(papier, func(t *testing.T) {
			probe := &bogenprobe{lauf: &service.Bogenlauf{}}

			mitBogenleger(probe).BogenLegen(probekontext(t), "7", "/tmp/b.pdf", papier, 0, 300, false, nil)

			if probe.masse.Seitenbreite != erwartet[0] || probe.masse.Seitenhoehe != erwartet[1] {
				t.Errorf("%g × %g, erwartet %g × %g",
					probe.masse.Seitenbreite, probe.masse.Seitenhoehe, erwartet[0], erwartet[1])
			}
		})
	}
}

// **Was dem Bogen fehlte, steht beim Erfolg** (§8).
func TestEinBogenNenntSeineBefunde(t *testing.T) {
	probe := &bogenprobe{lauf: &service.Bogenlauf{
		Pfad:    "/tmp/bogen.pdf",
		Befunde: []domain.Befund{{Ort: "Feuerball", Meldung: "die Vorlage wird verzerrt"}},
	}}

	antwort := mitBogenleger(probe).
		BogenLegen(probekontext(t), "7", "/tmp/bogen.pdf", "a4", 0, 300, true, nil)

	if len(antwort.Befunde) != 1 || antwort.Befunde[0] == "" {
		t.Errorf("Befunde: %v", antwort.Befunde)
	}
}

// Die Fehlerwege, jeder übersetzt.
func TestBogenfehlerWerdenUebersetzt(t *testing.T) {
	faelle := map[string]func(*app.App) app.BogenAntwort{
		"eine Kennung, die keine ist": func(a *app.App) app.BogenAntwort {
			return a.BogenLegen(context.Background(), "sieben", "/tmp/b.pdf", "a4", 0, 300, true, nil)
		},
		"ein Leger, der scheitert": func(a *app.App) app.BogenAntwort {
			return a.BogenLegen(context.Background(), "7", "/tmp/b.pdf", "a4", 0, 300, true, nil)
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			antwort := ruf(mitBogenleger(&bogenprobe{fehler: errors.New("es ging nicht")}))

			if antwort.Fehler == nil {
				t.Fatal("kein Fehler")
			}

			if antwort.Befunde == nil {
				t.Error("Befunde ist null — das Frontend läuft darüber, ohne zu prüfen")
			}
		})
	}
}
