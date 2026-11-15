// Der Probelauf des Druckbogens am Fenster — UXR-F7
// (#2015).
package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einProbelauf ist ein Ergebnis wie das des Dienstes für zehn Karten.
func einProbelauf() *service.Bogenprobelauf {
	masse := service.StandardBogen()

	return &service.Bogenprobelauf{
		Karten: 10, Seiten: 2, Spalten: 3, Zeilen: 3, JeSeite: 9,
		AufLetzterSeite: 1, Bildbreite: 744, Masse: masse,
	}
}

// **Der Probelauf rechnet mit denselben Maßen wie der Lauf.**
//
// Dieselben Wahlen gehen einmal durch `BogenLegen` und einmal durch
// `BogenProbelauf`; beim Leger kommt beide Male dasselbe an. Wer die Maße an
// zwei Stellen baute, hätte zwei Bögen: den, den „Wie es wird" zeigt, und den,
// den der Knopf legt.
func TestDerProbelaufRechnetMitDenMassenDesLaufs(t *testing.T) {
	t.Parallel()

	for _, papier := range []string{"a4", "letter", "papyrus"} {
		t.Run(papier, func(t *testing.T) {
			t.Parallel()

			lauf := &bogenprobe{lauf: &service.Bogenlauf{}, probelauf: einProbelauf()}
			anwendung := mitBogenleger(lauf)

			anwendung.BogenLegen(probekontext(t), "7", "/tmp/b.pdf", papier, 3, 150, false, []string{"2"})
			gelegt, legeauswahl := lauf.masse, lauf.auswahl

			antwort := anwendung.BogenProbelauf(probekontext(t), "7", papier, 3, 150, false, []string{"2"})
			if antwort.Fehler != nil {
				t.Fatalf("%+v", antwort.Fehler)
			}

			if lauf.masse != gelegt {
				t.Errorf("Probelauf rechnete mit %+v, der Lauf mit %+v", lauf.masse, gelegt)
			}

			if len(lauf.probeAuswahl.Karten) != 1 || lauf.probeAuswahl.Karten[0] != 2 ||
				len(legeauswahl.Karten) != 1 {
				t.Errorf("Auswahl Probelauf %v, Lauf %v", lauf.probeAuswahl, legeauswahl)
			}
		})
	}
}

// **Die Antwort trägt, was „Wie es wird" zeigt.**
func TestDieAntwortDesBogenprobelaufsTraegtDieZahlen(t *testing.T) {
	t.Parallel()

	antwort := mitBogenleger(&bogenprobe{probelauf: einProbelauf()}).
		BogenProbelauf(probekontext(t), "7", "a4", 0, 300, true, nil)

	erwartet := app.BogenprobelaufAntwort{
		Karten: 10, Seiten: 2, Spalten: 3, Zeilen: 3, JeSeite: 9, AufLetzterSeite: 1,
		Bildbreite: 744, Seitenbreite: 210, Seitenhoehe: 297,
		Kartenbreite: 63, Kartenhoehe: 88, Rand: 10, DPI: 300, Schnittmarken: true,
	}

	if antwort != erwartet {
		t.Errorf("%+v\nerwartet %+v", antwort, erwartet)
	}
}

// **Dieselben Fehler, übersetzt wie die des Laufs.**
//
// Jede Eingabe geht durch beide Wege; verglichen wird der übersetzte Fehler.
func TestDerBogenprobelaufGibtDieFehlerDesLaufs(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		setID   string
		auswahl []string
		leger   *bogenprobe
	}{
		"eine Kennung, die keine ist": {setID: "sieben", leger: &bogenprobe{}},
		"eine Auswahl, die keine ist": {setID: "7", auswahl: []string{"x"}, leger: &bogenprobe{}},
		"ein Leger, der scheitert": {
			setID: "7",
			leger: &bogenprobe{
				fehler:      errors.New("es ging nicht"),
				probefehler: errors.New("es ging nicht"),
			},
		},
		"ein Set, das es nicht gibt": {
			setID: "7",
			leger: &bogenprobe{
				fehler:      domain.ErrNotFound,
				probefehler: domain.ErrNotFound,
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := mitBogenleger(fall.leger)

			lauf := anwendung.BogenLegen(
				probekontext(t), fall.setID, "/tmp/b.pdf", "a4", 0, 300, true, fall.auswahl)
			probe := anwendung.BogenProbelauf(
				probekontext(t), fall.setID, "a4", 0, 300, true, fall.auswahl)

			if lauf.Fehler == nil || probe.Fehler == nil {
				t.Fatalf("Lauf %+v, Probelauf %+v", lauf.Fehler, probe.Fehler)
			}

			if !reflect.DeepEqual(ohneVorgang(lauf.Fehler), ohneVorgang(probe.Fehler)) {
				t.Errorf("Lauf %+v, Probelauf %+v", *lauf.Fehler, *probe.Fehler)
			}
		})
	}
}

// **Die Bindung fragt nach keinem Ort.** Ein Probelauf schreibt nichts, also
// gibt es keinen Dialog — und der Leger bekommt keinen Pfad.
func TestDieBindungZumBogenprobelaufFragtNachKeinemOrt(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	lauf := &bogenprobe{probelauf: einProbelauf()}
	anwendung.MitBogenleger(lauf)

	gefragt := false

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		gefragt = true

		return "", nil
	})

	antwort := app.NeueBindungen(anwendung).BogenProbelauf(t.Context(), "7", "a4", 0, 300, true, nil)
	if antwort.Fehler != nil || antwort.Seiten != 2 {
		t.Fatalf("%+v", antwort)
	}

	if gefragt || lauf.pfad != "" {
		t.Error("der Probelauf hat nach einem Ort gefragt oder geschrieben")
	}
}

// ohneVorgang nimmt einem Fehler die Kennung, die jeder Aufruf neu würfelt: sie
// steht im Protokoll, damit man den Aufruf wiederfindet, und ist nie gleich.
func ohneVorgang(fehler *app.Fehler) app.Fehler {
	gleich := *fehler
	gleich.Vorgang = ""
	gleich.Werte = map[string]string{}

	for schluessel, wert := range fehler.Werte {
		if schluessel != "vorgang" {
			gleich.Werte[schluessel] = wert
		}
	}

	return gleich
}
