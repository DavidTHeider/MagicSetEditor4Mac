// Der Probelauf des Bogenlegers — UXR-F7
// (#2015).
//
// **Ein Weg, zwei Ausgänge**: der Probelauf und der echte Lauf gehen durch
// denselben Plan, bis vor das Zeichnen. Ein Probelauf, der andere Fehler gäbe
// als der Lauf, den er ankündigt, wäre eine Zusage, die der Knopf danach bricht.
package service_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Probelauf rechnet, was „Wie es wird" zeigt.**
//
// Zehn Karten auf A4 mit 10 mm Rand sind drei Spalten und drei Zeilen — neun
// je Seite, die zehnte beginnt die zweite. Die Breite in Bildpunkten ist die,
// mit der der echte Lauf rastert: 63 mm bei 300 dpi.
func TestDerProbelaufRechnetKartenJeBogenUndBogenzahl(t *testing.T) {
	t.Parallel()

	probe, err := bogenleger(t, 10).Probelauf(
		t.Context(), 7, service.StandardBogen(), domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("Probelauf: %v", err)
	}

	erwartet := map[string][2]int{
		"Karten":          {probe.Karten, 10},
		"Seiten":          {probe.Seiten, 2},
		"Spalten":         {probe.Spalten, 3},
		"Zeilen":          {probe.Zeilen, 3},
		"je Seite":        {probe.JeSeite, 9},
		"auf der letzten": {probe.AufLetzterSeite, 1},
		"Bildbreite":      {probe.Bildbreite, 744},
	}

	for name, paar := range erwartet {
		if paar[0] != paar[1] {
			t.Errorf("%s: %d, erwartet %d", name, paar[0], paar[1])
		}
	}

	// **Die Maße reisen mit**, damit die Oberfläche „A4 · 300 dpi" aus
	// derselben Antwort sagt und nicht aus ihrem eigenen Zustand.
	if probe.Masse != service.StandardBogen() {
		t.Errorf("Maße: %+v", probe.Masse)
	}
}

// **Eine volle letzte Seite ist keine halbe.**
func TestEinVollerBogenHatNeunAufDerLetztenSeite(t *testing.T) {
	t.Parallel()

	probe, err := bogenleger(t, 18).Probelauf(
		t.Context(), 7, service.StandardBogen(), domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("Probelauf: %v", err)
	}

	if probe.Seiten != 2 || probe.AufLetzterSeite != 9 {
		t.Errorf("%d Seiten, %d auf der letzten; erwartet 2 und 9",
			probe.Seiten, probe.AufLetzterSeite)
	}
}

// **Die Auswahl gilt im Probelauf wie im Lauf.**
func TestDerProbelaufZaehltNurDieGewaehlten(t *testing.T) {
	t.Parallel()

	probe, err := bogenleger(t, 12).Probelauf(t.Context(), 7, service.StandardBogen(),
		domain.Kartenauswahl{Karten: []int64{2, 5}})
	if err != nil {
		t.Fatalf("Probelauf: %v", err)
	}

	if probe.Karten != 2 || probe.Seiten != 1 || probe.AufLetzterSeite != 2 {
		t.Errorf("%+v, erwartet 2 Karten auf einer Seite", probe)
	}
}

// **Anschnitt und Papier ändern die Rechnung.**
//
// Mit 3 mm Anschnitt ist eine Zelle 69 × 94 mm; auf Letter (215,9 × 279,4 mm)
// bleiben bei 10 mm Rand zwei Spalten und zwei Zeilen — vier je Seite.
func TestDerProbelaufFolgtPapierUndAnschnitt(t *testing.T) {
	t.Parallel()

	masse := service.StandardBogen()
	masse.Seitenbreite, masse.Seitenhoehe = 215.9, 279.4
	masse.Anschnitt = 3

	probe, err := bogenleger(t, 5).Probelauf(t.Context(), 7, masse, domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("Probelauf: %v", err)
	}

	if probe.JeSeite != 4 || probe.Seiten != 2 {
		t.Errorf("%d je Seite, %d Seiten; erwartet 4 und 2", probe.JeSeite, probe.Seiten)
	}

	// (63 + 2·3) mm bei 300 dpi
	if probe.Bildbreite != 815 {
		t.Errorf("Bildbreite %d, erwartet 815", probe.Bildbreite)
	}
}

// **Dieselben Fehler wie der echte Lauf — über denselben Weg.**
//
// Jeder Fall geht mit **derselben** fehlerhaften Eingabe durch beide Wege, und
// verglichen werden die Art des Fehlers und sein Wortlaut. Das ist der
// Beleg, dass der Probelauf keinen eigenen Vorprüfer hat, der neben dem
// echten Weg läuft und mit ihm auseinanderliefe.
func TestProbelaufUndLaufGebenDieselbenFehler(t *testing.T) {
	t.Parallel()

	zuKlein := service.StandardBogen()
	zuKlein.Kartenbreite = 300

	faelle := map[string]struct {
		leger   func(t *testing.T) *service.Bogenleger
		setID   int64
		masse   service.Bogenmasse
		auswahl domain.Kartenauswahl
		art     error
	}{
		"eine Karte, die nicht aufs Papier passt": {
			leger: func(t *testing.T) *service.Bogenleger { return bogenleger(t, 3) },
			setID: 7, masse: zuKlein, art: domain.ErrValidation,
		},
		"eine Auflösung über der Grenze": {
			leger: func(t *testing.T) *service.Bogenleger { return bogenleger(t, 3) },
			setID: 7, art: domain.ErrValidation,
			masse: func() service.Bogenmasse {
				m := service.StandardBogen()
				m.DPI = service.GrenzeDPI + 1

				return m
			}(),
		},
		"ein Set, das es nicht gibt": {
			leger: func(t *testing.T) *service.Bogenleger { return bogenleger(t, 3) },
			setID: 999, masse: service.StandardBogen(), art: domain.ErrNotFound,
		},
		"ein Set ohne Karte": {
			leger: func(t *testing.T) *service.Bogenleger { return bogenleger(t, 0) },
			setID: 7, masse: service.StandardBogen(), art: domain.ErrValidation,
		},
		"eine Auswahl, die keine Karte trifft": {
			leger: func(t *testing.T) *service.Bogenleger { return bogenleger(t, 3) },
			setID: 7, masse: service.StandardBogen(),
			auswahl: domain.Kartenauswahl{Karten: []int64{99}}, art: domain.ErrNotFound,
		},
		"ein Bestand, der die Karten nicht gibt": {
			leger: func(t *testing.T) *service.Bogenleger {
				t.Helper()

				bestand := karten("Eine")
				bestand.kartenfehler = errors.New("die Datenbank ist zu")

				return service.NeuerBogenleger(kartenDouble{leserDouble: bestand},
					planDouble{}, render.NeuerRasterer(nil))
			},
			setID: 7, masse: service.StandardBogen(),
		},
		"eine Setliste, die nicht antwortet": {
			leger: func(t *testing.T) *service.Bogenleger {
				t.Helper()

				bestand := karten("Eine")
				bestand.fehler = errors.New("die Datenbank ist zu")

				return service.NeuerBogenleger(kartenDouble{leserDouble: bestand},
					planDouble{}, render.NeuerRasterer(nil))
			},
			setID: 7, masse: service.StandardBogen(),
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, lauffehler := fall.leger(t).Lege(t.Context(), fall.setID,
				filepath.Join(t.TempDir(), "bogen.pdf"), fall.masse, fall.auswahl)
			if lauffehler == nil {
				t.Fatal("der echte Lauf hat die Eingabe hingenommen")
			}

			_, probefehler := fall.leger(t).Probelauf(t.Context(), fall.setID, fall.masse, fall.auswahl)
			if probefehler == nil {
				t.Fatal("der Probelauf hat die Eingabe hingenommen")
			}

			if fall.art != nil && !errors.Is(probefehler, fall.art) {
				t.Errorf("Probelauf: %v, erwartet %v", probefehler, fall.art)
			}

			if probefehler.Error() != lauffehler.Error() {
				t.Errorf("Probelauf %q, Lauf %q", probefehler, lauffehler)
			}
		})
	}
}

// **Der Probelauf zeichnet nichts.**
//
// Ein Plan, der sich nicht bauen lässt, hält den echten Lauf auf und den
// Probelauf nicht: er kommt nie dorthin. Das ist seine Reichweite — bis vor
// das Zeichnen —, und sie steht hier, damit sie niemand für mehr hält.
func TestDerProbelaufZeichnetNichts(t *testing.T) {
	t.Parallel()

	leger := service.NeuerBogenleger(
		kartenDouble{leserDouble: karten("Eine", "Zweite")},
		planDouble{fehler: errors.New("kein Stil")}, render.NeuerRasterer(nil))

	if _, err := leger.Lege(t.Context(), 7, filepath.Join(t.TempDir(), "b.pdf"),
		service.StandardBogen(), domain.Kartenauswahl{}); err == nil {
		t.Fatal("der echte Lauf hätte am Plan scheitern müssen")
	}

	probe, err := leger.Probelauf(t.Context(), 7, service.StandardBogen(), domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("Probelauf: %v", err)
	}

	if probe.Karten != 2 {
		t.Errorf("%d Karten, erwartet 2", probe.Karten)
	}
}

// **Der Probelauf liest nur — also fragt er den Riegel nicht.**
//
// Ein Prüfer, der jedes Schreiben verweigert, hält den echten Lauf auf. Der
// Probelauf schreibt nichts, und was ein Leser über die Fläche erfährt, ist
// dieselbe Rechnung, die er an den Karten ohnehin sieht. Der Riegel fragt erst
// dort, wo etwas auf die Platte käme.
func TestDerProbelaufFragtDenSchreibriegelNicht(t *testing.T) {
	t.Parallel()

	nein := &neinsager{}
	leger := bogenleger(t, 3).MitPruefer(nein)

	if _, err := leger.Lege(t.Context(), 7, filepath.Join(t.TempDir(), "b.pdf"),
		service.StandardBogen(), domain.Kartenauswahl{}); !errors.Is(err, domain.ErrNurLesend) {
		t.Fatalf("der echte Lauf: %v, erwartet ErrNurLesend", err)
	}

	if _, err := leger.Probelauf(t.Context(), 7, service.StandardBogen(),
		domain.Kartenauswahl{}); err != nil {
		t.Fatalf("Probelauf: %v", err)
	}

	if len(nein.gefragt) != 1 {
		t.Errorf("der Riegel wurde %d-mal gefragt, erwartet 1 (nur vom echten Lauf)",
			len(nein.gefragt))
	}
}
