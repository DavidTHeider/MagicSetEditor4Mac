package app

// **Der Netzweg im Fenster und an den Flächen** — Werkstück N5 von Phase 5.4
// (#1571).
//
// # Wonach hier gefragt wird
//
// Danach, dass die Einstellungen **gedeutet** werden — eine Zahl in Sekunden
// wird eine Frist, ein leerer Wert die Vorgabe —, und danach, dass ein
// Lesefehler nichts aufhält: was hier zurückkommt, sind Vorgaben, und ohne
// sie ginge eine Anfrage **direkt** hinaus statt gar nicht.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// netzeinstellungsprobe gibt einen festen Satz Einstellungen.
type netzeinstellungsprobe struct {
	werte  map[string]string
	fehler error
}

func (n netzeinstellungsprobe) Alle(context.Context) (map[string]string, error) {
	return n.werte, n.fehler
}

func (n netzeinstellungsprobe) Setzen(context.Context, string, string) error { return nil }

func (n netzeinstellungsprobe) Stand(context.Context) (map[string]service.Wahl, error) {
	return nil, nil
}

// **Die gesetzten Werte kommen an.**
func TestDieNetzvorgabenWerdenGedeutet(t *testing.T) {
	t.Parallel()

	vorgaben := NetzvorgabenAus(map[string]string{
		domain.NetzProxyHTTPS:      "http://proxy.firma:3128",
		domain.NetzProxyAus:        "gpu.intern",
		domain.NetzFristStillstand: "300",
	})

	if vorgaben.Proxy.HTTPS != "http://proxy.firma:3128" {
		t.Errorf("der Proxy ist %q", vorgaben.Proxy.HTTPS)
	}

	if vorgaben.Proxy.Aus != "gpu.intern" {
		t.Errorf("die Ausnahmeliste ist %q", vorgaben.Proxy.Aus)
	}

	if vorgaben.Fristen.Stillstand != 300*time.Second {
		t.Errorf("die Stillstandsfrist ist %v", vorgaben.Fristen.Stillstand)
	}

	// **Was nicht gesetzt ist, bleibt die übliche** — und das ist die
	// Zusage, an der eine halbe Einstellung sonst die anderen mitnähme.
	if vorgaben.Fristen.Kopf != netz.UeblicheFristen().Kopf {
		t.Errorf("die Kopffrist ist %v", vorgaben.Fristen.Kopf)
	}
}

// **Ein Lesefehler hält nichts auf.**
//
// **Weil das Ergebnis Vorgaben sind.** Ohne sie ginge eine Anfrage direkt
// hinaus statt gar nicht — und das ist das Verhalten, das diese Anwendung
// ohne jede Einstellung ohnehin hat. Eine Netzanfrage an einem unlesbaren
// Bestand scheitern zu lassen wäre eine zweite, schlechtere Meldung über
// dasselbe Problem.
func TestEinLesefehlerGibtDieUeblichenVorgaben(t *testing.T) {
	t.Parallel()

	for name, anwendung := range map[string]*App{
		"ohne Einstellungen": NeueApp(""),
		"mit Lesefehler": func() *App {
			a := NeueApp("")
			a.MitEinstellungen(netzeinstellungsprobe{
				fehler: errors.New("die Platte schweigt"),
			})

			return a
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorgaben := anwendung.netzvorgaben(t.Context())

			if vorgaben.Fristen != netz.UeblicheFristen() {
				t.Errorf("die Fristen sind %+v", vorgaben.Fristen)
			}

			if vorgaben.Proxy != (netz.Proxyvorgabe{}) {
				t.Errorf("der Proxy ist %+v", vorgaben.Proxy)
			}
		})
	}
}

// **Die Quelle wird je Anfrage gefragt und nicht einmal.**
//
// **Das ist die Zusage, für die es die Funktion gibt**: wer im Fenster einen
// Proxy einträgt, erwartet nicht, die Anwendung neu starten zu müssen.
func TestDieProxyquelleLiestJedesMalNeu(t *testing.T) {
	anwendung := NeueApp("")

	werte := map[string]string{}
	anwendung.MitEinstellungen(netzeinstellungsprobe{werte: werte})

	quelle := anwendung.Proxyquelle()

	if erst := quelle(t.Context()); erst.HTTPS != "" {
		t.Fatalf("vor der Einstellung steht %q", erst.HTTPS)
	}

	werte[domain.NetzProxyHTTPS] = "http://spaeter:3128"

	if dann := quelle(t.Context()); dann.HTTPS != "http://spaeter:3128" {
		t.Errorf("nach der Einstellung steht %q", dann.HTTPS)
	}
}

// **Auch die Dienstwahl kennt den Netzweg — und die Flächen ohne Fenster.**
//
// **Ohne diesen Weg gälte ein Proxy nur im Fenster** (Werkstück 2 des
// Issues): `dienstklient`, `mtgjsonbulk` und der Scryfall-Klient sind
// Funktionen des Pakets und bekommen ihre [Dienstwahl] hereingereicht.
func TestDieDienstwahlKenntDenNetzweg(t *testing.T) {
	t.Parallel()

	wahl := NeueEinstellungswahl(netzeinstellungsprobe{werte: map[string]string{
		domain.NetzProxyHTTP: "http://ausderwahl:3128",
	}})

	if gab := wahl.Netzvorgaben(t.Context()); gab.Proxy.HTTP != "http://ausderwahl:3128" {
		t.Errorf("die Wahl gibt %q", gab.Proxy.HTTP)
	}

	// **Und über den Weg, den die Funktionen des Pakets nehmen.**
	quelle := proxyquelleAus(wahl)
	if quelle == nil {
		t.Fatal("die Wahl gibt keine Quelle her")
	}

	if gab := quelle(t.Context()); gab.HTTP != "http://ausderwahl:3128" {
		t.Errorf("die Quelle gibt %q", gab.HTTP)
	}

	// **Ein Lesefehler gibt die üblichen** — dieselbe Regel wie oben.
	kaputt := NeueEinstellungswahl(netzeinstellungsprobe{
		fehler: errors.New("die Platte schweigt"),
	})

	if gab := kaputt.Netzvorgaben(t.Context()); gab.Fristen != netz.UeblicheFristen() {
		t.Errorf("die Fristen sind %+v", gab.Fristen)
	}
}

// **Eine Wahl, die keine Einstellungen kennt, gibt keine Quelle.**
//
// **Und das ist richtig**: dann gilt die Umgebung, wie überall. Eine Quelle,
// die eine leere Vorgabe zurückgäbe, sähe genauso aus und verspräche mehr.
func TestEineWahlOhneEinstellungenGibtKeineQuelle(t *testing.T) {
	t.Parallel()

	if quelle := proxyquelleAus(wahlOhneEinstellungen{}); quelle != nil {
		t.Error("eine Wahl ohne Einstellungen gab eine Quelle")
	}
}

// wahlOhneEinstellungen beantwortet nur den Schalter.
type wahlOhneEinstellungen struct{}

func (wahlOhneEinstellungen) IstAn(context.Context, domain.Dienst) (bool, error) {
	return true, nil
}

// **Die Fläche ohne Fenster steht und liest dieselben Einstellungen.**
//
// **Werkstück 2 des Issues, wörtlich**: `mse`, `mse-mcp` und `mse-rest` sind
// Klienten wie das Fenster. Eine zweite Deutung dort wäre die Gestalt, in der
// ein Proxy an einer Fläche gilt und an der anderen nicht (§16).
func TestDieNetzflaecheLiestDieselbenEinstellungen(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	ctx := riegelkontext(t)

	anwendung := NetzflaecheUeber(db)

	// **Ohne Einstellung stehen die üblichen Fristen da.**
	vorher := anwendung.Netzstand(ctx)
	if vorher.Fehler != nil {
		t.Fatalf("%+v", vorher.Fehler)
	}

	if vorher.FristKopf != int(netz.UeblicheFristen().Kopf.Seconds()) {
		t.Errorf("die Kopffrist ist %d", vorher.FristKopf)
	}

	if vorher.ProxyBenutzt {
		t.Error("ohne Einstellung gilt ein Proxy als benutzt")
	}

	// **Gesetzt wird über denselben Weg wie überall** — und die Fläche liest
	// ihn.
	if antwort := anwendung.EinstellungSetzen(
		ctx, domain.NetzProxyHTTPS, "http://ausdemBestand:3128"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	nachher := anwendung.Netzstand(ctx)
	if nachher.Fehler != nil {
		t.Fatalf("%+v", nachher.Fehler)
	}

	if nachher.ProxyHTTPS.Wert != "http://ausdemBestand:3128" ||
		nachher.ProxyHTTPS.Herkunft != string(netz.AusDerEinstellung) {
		t.Errorf("gelesen wurde %+v", nachher.ProxyHTTPS)
	}

	if !nachher.ProxyBenutzt {
		t.Error("ein gesetzter Proxy gilt als nicht benutzt")
	}
}
