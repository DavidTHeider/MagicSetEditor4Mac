// **Profilvorgaben über die Fassade** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vorgabenfuehrer ist eine Verwaltung im Arbeitsspeicher.
type vorgabenfuehrer struct {
	staende []service.Vorgabenstand

	// gesehen ist, was das letzte Setzen zu sehen bekam.
	gesehen domain.Profilvorgabe

	// profil ist die Kennung, mit der zuletzt gerufen wurde.
	profil int64

	fehler error
}

func (f *vorgabenfuehrer) Stand(_ context.Context, profilID int64) (
	[]service.Vorgabenstand, error,
) {
	f.profil = profilID

	return f.staende, f.fehler
}

func (f *vorgabenfuehrer) Setzen(
	_ context.Context, profilID int64, vorgabe domain.Profilvorgabe,
) error {
	f.profil = profilID
	f.gesehen = vorgabe

	return f.fehler
}

// mitVorgaben baut eine Anwendung über einer Verwaltung.
func mitVorgaben(fuehrer *vorgabenfuehrer) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitVorgaben(fuehrer)

	return anwendung
}

// **Der Stand kommt an, samt Angebot.**
func TestDerVorgabenstandKommtAn(t *testing.T) {
	t.Parallel()

	fuehrer := &vorgabenfuehrer{staende: []service.Vorgabenstand{{
		Art: "deck", Mehrwertig: true, Werte: []string{"7"},
		Moeglich: []service.Vorgabewahl{{Wert: "7", Beschriftung: "Meins"}},
		Verwaist: []string{},
	}}}

	antwort := mitVorgaben(fuehrer).Profilvorgaben(t.Context(), "3")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Vorgaben) != 1 || antwort.Vorgaben[0].Art != "deck" {
		t.Fatalf("%+v", antwort.Vorgaben)
	}

	// **Die Kennung reist als Text und kommt als Zahl unten an.**
	if fuehrer.profil != 3 {
		t.Errorf("gerufen wurde mit %d", fuehrer.profil)
	}
}

// **Setzen reicht die Art durch und gibt den ganzen Stand zurück.**
func TestVorgabeSetzenGibtDenGanzenStand(t *testing.T) {
	t.Parallel()

	fuehrer := &vorgabenfuehrer{staende: []service.Vorgabenstand{
		{Art: "format"}, {Art: "deck"},
	}}

	antwort := mitVorgaben(fuehrer).ProfilvorgabeSetzen(t.Context(),
		app.VorgabeangabenDTO{ProfilID: "3", Art: "deck", Werte: []string{"7"}})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if fuehrer.gesehen.Art != domain.VorgabeDeck ||
		len(fuehrer.gesehen.Werte) != 1 {
		t.Errorf("gesehen = %+v", fuehrer.gesehen)
	}

	// **Zwei Stände zurück und nicht einer** (§7, Regel 2): die Seite zeigt
	// fünf nebeneinander, und wer nur die geänderte bekäme, müsste danach noch
	// einmal lesen.
	if len(antwort.Vorgaben) != 2 {
		t.Errorf("zurück kamen %d Stände", len(antwort.Vorgaben))
	}
}

// **Eine fehlende Liste ist eine leere und kein `nil`.**
//
// **Über die JSON-Naht würde daraus `null`**, und die Seite griffe in etwas,
// das es nicht gibt (§3, „zwischen den Prüfebenen").
func TestEineLeereVorgabenlisteIstEineListe(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*app.App) app.VorgabenAntwort{
		"krumme Kennung beim Lesen": func(a *app.App) app.VorgabenAntwort {
			return a.Profilvorgaben(context.Background(), "keine-zahl")
		},
		"krumme Kennung beim Setzen": func(a *app.App) app.VorgabenAntwort {
			return a.ProfilvorgabeSetzen(context.Background(),
				app.VorgabeangabenDTO{ProfilID: "keine-zahl", Art: "deck"})
		},
		"stolpernder Dienst beim Lesen": func(a *app.App) app.VorgabenAntwort {
			return a.Profilvorgaben(context.Background(), "1")
		},
		"stolpernder Dienst beim Setzen": func(a *app.App) app.VorgabenAntwort {
			return a.ProfilvorgabeSetzen(context.Background(),
				app.VorgabeangabenDTO{ProfilID: "1", Art: "deck"})
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(mitVorgaben(&vorgabenfuehrer{
				fehler: errors.New("die Platte ist voll"),
			}))

			if antwort.Fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}

			if antwort.Vorgaben == nil {
				t.Error("die Liste ist nil")
			}
		})
	}
}

// **Eine fehlende Werteliste räumt ab und reist nicht als `nil` weiter.**
func TestEineFehlendeWertelisteRaeumtAb(t *testing.T) {
	t.Parallel()

	fuehrer := &vorgabenfuehrer{}

	mitVorgaben(fuehrer).ProfilvorgabeSetzen(t.Context(),
		app.VorgabeangabenDTO{ProfilID: "1", Art: "deck"})

	if fuehrer.gesehen.Werte == nil {
		t.Error("die Werte kamen als nil unten an")
	}

	if len(fuehrer.gesehen.Werte) != 0 {
		t.Errorf("gesehen = %+v", fuehrer.gesehen.Werte)
	}
}

// **Und die Bindungen des Fensters reichen durch.**
func TestDieVorgabenbindungenReichenDurch(t *testing.T) {
	t.Parallel()

	fuehrer := &vorgabenfuehrer{staende: []service.Vorgabenstand{{Art: "deck"}}}
	bindungen := app.NeueBindungen(mitVorgaben(fuehrer))

	if antwort := bindungen.Profilvorgaben(t.Context(), "1"); len(antwort.Vorgaben) != 1 {
		t.Errorf("Profilvorgaben: %+v", antwort)
	}

	if antwort := bindungen.ProfilvorgabeSetzen(t.Context(), app.VorgabeangabenDTO{
		ProfilID: "1", Art: "deck", Werte: []string{"7"},
	}); antwort.Fehler != nil {
		t.Errorf("ProfilvorgabeSetzen: %+v", antwort.Fehler)
	}

	if fuehrer.gesehen.Art != domain.VorgabeDeck {
		t.Errorf("gesehen = %+v", fuehrer.gesehen)
	}
}

// **Und der Flächenweg reicht beide durch** — über einen echten Bestand.
func TestDerVorgabenwegUeberDieFlaeche(t *testing.T) {
	t.Parallel()

	flaeche := app.NutzerflaecheUeber(frischerBestand(t))

	antwort := flaeche.Profilvorgaben(probekontext(t), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Vorgaben) != 5 {
		t.Fatalf("es sind %d Arten", len(antwort.Vorgaben))
	}

	// **Ein Wert, den es nicht gibt, wird abgewiesen** — und zwar gegen das
	// Angebot dieses Bestands, das leer ist.
	gesetzt := flaeche.ProfilvorgabeSetzen(probekontext(t),
		app.VorgabeangabenDTO{ProfilID: "1", Art: "deck", Werte: []string{"7"}})
	if gesetzt.Fehler == nil {
		t.Error("ein Deck, das es nicht gibt, kam durch")
	}

	// **Und eine leere Liste geht durch** — sie räumt ab, und dafür braucht es
	// kein Angebot.
	geraeumt := flaeche.ProfilvorgabeSetzen(probekontext(t),
		app.VorgabeangabenDTO{ProfilID: "1", Art: "deck"})
	if geraeumt.Fehler != nil {
		t.Errorf("das Abräumen scheiterte: %+v", geraeumt.Fehler)
	}
}
