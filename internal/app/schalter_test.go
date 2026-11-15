package app_test

// **Bereiche und Funktionen abschalten, an der Anwendung** — OPS-10
// (#1908).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// anwendungMit baut eine Anwendung, deren Einstellungen diese Werte tragen.
func anwendungMit(werte map[string]string) *app.Bindungen {
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{werte: werte})

	return app.NeueBindungen(anwendung)
}

// abweisungAus ruft eine Bindung und gibt den Fehler, den ihr Tor geworfen
// hat — oder nil, wenn sie durchkam.
func abweisungAus(t *testing.T, ruf func()) (fehler *app.Fehler) {
	t.Helper()

	defer func() {
		grund := recover()
		if grund == nil {
			return
		}

		panik, istFehler := grund.(error)
		if !istFehler {
			t.Fatalf("das Tor warf %T: %v", grund, grund)
		}

		roh, markiert := strings.CutPrefix(panik.Error(), app.Fehlermarke)
		if !markiert {
			t.Fatalf("die Panik trägt keine Marke: %q", panik.Error())
		}

		fehler = &app.Fehler{}
		if entpackt := json.Unmarshal([]byte(roh), fehler); entpackt != nil {
			t.Fatalf("die Panik trägt keinen Fehler: %v", entpackt)
		}
	}()

	ruf()

	return nil
}

// **Ein abgeschalteter Dienst hält seine Bindung an** — mit dem Code, dem
// Schlüssel und dem Namen des Schalters, in einer Gestalt, die das Fenster
// über die Panik hinweg wieder auspacken kann (OPS-10.AK-02).
func TestEinAbgeschalteterWegHaeltSeineBindungAn(t *testing.T) {
	t.Parallel()

	bindungen := anwendungMit(map[string]string{
		domain.SchalterEinstellung("dienst"): domain.SchalterAus,
	})

	fehler := abweisungAus(t, func() { bindungen.RESTStand(t.Context()) })
	if fehler == nil {
		t.Fatal("RESTStand kam bei abgeschaltetem Dienst durch")
	}

	if fehler.Code != app.CodeAbgeschaltet || fehler.Schluessel != domain.SchluesselBereichAbgeschaltet ||
		fehler.Werte["schalter"] != "dienst" {
		t.Errorf("%+v", fehler)
	}
}

// **Und ohne Wahl, mit „an“ oder bei einem Lesefehler kommt sie durch** —
// ein Schalter sichert nichts, und ein Bestand, der nicht antwortet, soll
// das Fenster nicht leer machen (Vermerk 0158).
func TestEinOffenerWegKommtDurch(t *testing.T) {
	t.Parallel()

	for name, probe := range map[string]*einstellungsprobe{
		"ohne Wahl":      {werte: map[string]string{}},
		"mit an":         {werte: map[string]string{domain.SchalterEinstellung("dienst"): domain.SchalterAn}},
		"anderer aus":    {werte: map[string]string{domain.SchalterEinstellung("spielen"): domain.SchalterAus}},
		"mit Lesefehler": {fehler: errors.New("kein Bestand")},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := app.NeueApp("")
			anwendung.MitEinstellungen(probe)

			if fehler := abweisungAus(t, func() { app.NeueBindungen(anwendung).RESTStand(t.Context()) }); fehler != nil {
				t.Errorf("abgewiesen: %+v", fehler)
			}
		})
	}
}

// **Die Schalter selbst sind lesbar, auch wenn einer aus ist** — sie sind
// Kern, und ohne sie fände niemand den Weg zurück.
func TestDieSchalterSindLesbar(t *testing.T) {
	t.Parallel()

	bindungen := anwendungMit(map[string]string{
		domain.SchalterEinstellung("spielen"): domain.SchalterAus,
	})

	var antwort app.SchalterAntwort

	if fehler := abweisungAus(t, func() { antwort = bindungen.Schalter(t.Context()) }); fehler != nil {
		t.Fatalf("abgewiesen: %+v", fehler)
	}

	if len(antwort.Schalter) != len(domain.Schalterverzeichnis()) || antwort.Fehler != nil {
		t.Fatalf("%+v", antwort)
	}

	for _, schalter := range antwort.Schalter {
		if schalter.An == (schalter.Kennung == "spielen") || schalter.Reife != string(domain.ReifeStabil) ||
			schalter.Vorgabe != domain.SchalterAn {
			t.Errorf("%+v", schalter)
		}
	}
}

// **Ein Lesefehler kommt als Fehler an** und nicht als „alles an“: die
// Seite, die die Schalter zeigt, soll ihn sagen.
func TestDieSchalterSagenIhrenLesefehler(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{fehler: errors.New("kein Bestand")})

	if antwort := anwendung.Schalter(context.Background()); antwort.Fehler == nil || antwort.Schalter == nil {
		t.Errorf("%+v", antwort)
	}
}

// **Die Abweisung wird überall gleich übersetzt**, und ihr Code führt
// zurück auf die Art (`nahtfehler.go`).
func TestDieAbschaltungWirdUebersetzt(t *testing.T) {
	t.Parallel()

	fehler := app.Uebersetze(context.Background(), &domain.Abschaltung{Schalter: "markt"})

	if fehler.Code != app.CodeAbgeschaltet || fehler.Schluessel != domain.SchluesselBereichAbgeschaltet ||
		fehler.Werte["schalter"] != "markt" || !strings.Contains(fehler.Meldung, "abgeschaltet") {
		t.Errorf("%+v", fehler)
	}

	if !errors.Is(fehler.AlsFehler(), domain.ErrAbgeschaltet) {
		t.Errorf("der Code findet nicht zurück: %v", fehler.AlsFehler())
	}
}

// **Ein Befehl trifft seinen Weg** — der längste, der passt, und ein
// Platzhalter wie `<id>` steht für jedes Wort.
func TestEinBefehlTrifftSeinenWeg(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		befehl string
		rest   []string
		weg    string
	}{
		{"deck", []string{"neu", "Mein Deck"}, "mse deck neu"},
		{"deck", []string{"42"}, "mse deck <id>"},
		{"turniere", nil, "mse turniere"},
		{"sets", nil, ""},
		{"gibtsnicht", []string{"x"}, ""},
	} {
		weg, gefuehrt := app.Befehlsweg(fall.befehl, fall.rest)
		if weg != fall.weg || gefuehrt != (fall.weg != "") {
			t.Errorf("%s %v: %q %v, erwartet %q", fall.befehl, fall.rest, weg, gefuehrt, fall.weg)
		}
	}
}

// **Über einem echten Bestand** — so bauen die Wurzeln ohne Fenster die
// Wache: ohne Wahl alles an, mit „aus“ ein abgewiesener Weg.
func TestDieWacheLiestDenBestand(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)

	ctx, fehler := storage.MitGezeigtemProfil(probekontext(t), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	wache := app.SchalterwacheUeber(db)
	if stand := wache.Stand(ctx); stand.Fehler != nil || len(stand.Schalter) != len(domain.Schalterverzeichnis()) {
		t.Fatalf("%+v", stand)
	}

	if fehler := wache.Pruefe(ctx, domain.FlaecheFenster, "PartieBefehl"); fehler != nil {
		t.Fatalf("ohne Wahl: %v", fehler)
	}

	if fehler := service.NeuerEinstellungsverwalter(storage.NeueEinstellungen(db)).
		Setzen(ctx, domain.SchalterEinstellung("spielen"), domain.SchalterAus); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := wache.Pruefe(ctx, domain.FlaecheFenster, "PartieBefehl"); !errors.Is(fehler, domain.ErrAbgeschaltet) {
		t.Errorf("mit spielen aus: %v", fehler)
	}
}
