// **Die Fehlerwege der Freigaben** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil `freigaben_test.go` mit ihnen über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **was die Wege tun**,
// hier, **was sie melden, wenn der Bestand nicht antwortet** — und die drei
// Läufe hier haben gemeinsam, dass ihre Naht an mehr als einer Stelle klemmen
// kann.
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Übersicht, die nicht kommt, reißt jeden der drei Wege ab.**
//
// **Sie ist die Quelle für zwei Fragen zugleich** — welches Profil gemeint ist
// und wie die Nutzer heißen. Wer sie verschluckte, meldete „das Profil gibt es
// nicht", wo der Bestand nur nicht antwortet (§8).
func TestOhneNutzeruebersichtReisstJederFreigabenwegAb(t *testing.T) {
	t.Parallel()

	bauen := func() *app.App {
		anwendung := app.NeueApp("")
		anwendung.MitNutzern(&nutzerfuehrer{
			fehler: errors.New("die Platte ist voll"),
		})
		anwendung.MitProfilwechsel(&wechseldoppel{
			gezeigt: domain.Profil{ID: 10, NutzerID: 1},
		})
		anwendung.MitFreigaben(&freigabenfuehrer{rolle: domain.RolleVerwalter})

		return anwendung
	}

	faelle := map[string]func(*app.App) app.FreigabenAntwort{
		"Lesen": func(a *app.App) app.FreigabenAntwort {
			return a.Freigaben(context.Background(), "10")
		},
		"Setzen": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeSetzen(context.Background(),
				app.FreigabeangabenDTO{ProfilID: "10", NutzerID: "2"})
		},
		"Löschen": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeLoeschen(context.Background(), "10", "2")
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if antwort := ruf(bauen()); antwort.Fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Und eine Übersicht, die erst beim Abbilden abreißt, ebenso.**
//
// **Ein eigener Weg**: `profilMitKennung` glückte, `freigabenzeilen` nicht —
// und dazwischen liegt der Ruf, der die Namen holt.
func TestEineAbreissendeUebersichtBeimAbbilden(t *testing.T) {
	t.Parallel()

	nutzer := &zaehlendernutzerfuehrer{
		zeilen: []service.Nutzerzeile{{
			Nutzer:  domain.Nutzer{ID: 1, Name: "Ich"},
			Profile: []domain.Profil{{ID: 10, NutzerID: 1, Name: "Meins"}},
		}},

		// **Beim zweiten Ruf** — der erste sucht das Profil, der zweite die
		// Namen.
		abBeim: 2,
		fehler: errors.New("die Platte ist voll"),
	}

	anwendung := app.NeueApp("")
	anwendung.MitNutzern(nutzer)
	anwendung.MitProfilwechsel(&wechseldoppel{
		gezeigt: domain.Profil{ID: 10, NutzerID: 1},
	})
	anwendung.MitFreigaben(&freigabenfuehrer{rolle: domain.RolleVerwalter})

	if antwort := anwendung.Freigaben(t.Context(), "10"); antwort.Fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}

// zaehlendernutzerfuehrer klagt erst ab dem n-ten Ruf.
type zaehlendernutzerfuehrer struct {
	nutzerfuehrer

	zeilen []service.Nutzerzeile
	abBeim int
	fehler error
	rufe   int
}

func (f *zaehlendernutzerfuehrer) Uebersicht(context.Context) (
	[]service.Nutzerzeile, error,
) {
	f.rufe++

	if f.rufe >= f.abBeim {
		return nil, f.fehler
	}

	return f.zeilen, nil
}

// **Und ein Bestand, der erst hinter der Rollenfrage klagt.**
//
// **Der Fall, den ein gemeinsames Fehlerfeld verdeckt**: die Rolle glückt, das
// Lesen oder Schreiben nicht — und dahinter liegen drei Zweige, die sonst kein
// Lauf betritt (§10).
func TestEinBestandDerHinterDerRollenfrageKlagt(t *testing.T) {
	t.Parallel()

	fuehrer := func() *freigabenfuehrer {
		return &freigabenfuehrer{
			rolle:       domain.RolleVerwalter,
			datenfehler: errors.New("die Platte ist voll"),
		}
	}

	faelle := map[string]func(*app.App) app.FreigabenAntwort{
		"Lesen": func(a *app.App) app.FreigabenAntwort {
			return a.Freigaben(context.Background(), "10")
		},
		"Setzen": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeSetzen(context.Background(),
				app.FreigabeangabenDTO{
					ProfilID: "10", NutzerID: "2",
					Rolle: string(domain.RolleLeser),
				})
		},
		"Löschen": func(a *app.App) app.FreigabenAntwort {
			return a.FreigabeLoeschen(context.Background(), "10", "2")
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(mitFreigaben(fuehrer()))

			if antwort.Fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}

			if antwort.Freigaben == nil {
				t.Error("die Liste ist nil")
			}
		})
	}
}

// **Und der Flächenweg reicht alle drei durch** — über einen echten Bestand.
//
// **Über einen echten und nicht über ein Doppel**: genau das ist die Zusage
// dieser Naht — die drei Flächen ohne Fenster bekommen dieselbe Anwendung
// (§16). Ein Doppel dahinter prüfte die Verdrahtung nicht, die hier gemeint
// ist.
func TestDerFreigabenwegUeberDieFlaeche(t *testing.T) {
	t.Parallel()

	flaeche := app.NutzerflaecheUeber(frischerBestand(t))

	// **Ein zweiter Nutzer, damit es überhaupt jemanden freizugeben gibt.**
	angelegt := flaeche.NutzerAnlegen(probekontext(t), "Zweiter", "")
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	zweiter := angelegt.Nutzer.ID

	antwort := flaeche.Freigaben(probekontext(t), "1")
	if antwort.Fehler != nil {
		t.Fatalf("Freigaben: %+v", antwort.Fehler)
	}

	// **Der arbeitende Mensch ist der Eigner von Profil 1** — er verwaltet es
	// ohne Zeile im Bestand.
	if antwort.EigeneRolle != string(domain.RolleVerwalter) {
		t.Errorf("eigene Rolle = %q", antwort.EigeneRolle)
	}

	gesetzt := flaeche.FreigabeSetzen(probekontext(t), app.FreigabeangabenDTO{
		ProfilID: "1", NutzerID: zweiter, Rolle: string(domain.RolleMitarbeiter),
	})
	if gesetzt.Fehler != nil {
		t.Fatalf("FreigabeSetzen: %+v", gesetzt.Fehler)
	}

	if len(gesetzt.Freigaben) != 1 {
		t.Fatalf("es sind %d Freigaben", len(gesetzt.Freigaben))
	}

	geloescht := flaeche.FreigabeLoeschen(probekontext(t), "1", zweiter)
	if geloescht.Fehler != nil {
		t.Fatalf("FreigabeLoeschen: %+v", geloescht.Fehler)
	}

	if len(geloescht.Freigaben) != 0 {
		t.Errorf("nach dem Zurückziehen sind es %d", len(geloescht.Freigaben))
	}
}

// **Und die Deckübernahme meldet jeden ihrer Fehlerwege.**
//
// **Vier Ausgänge, und jeder einzeln** — Werkstück 12 (#1351): ohne Naht,
// krumme Profilkennung, krumme Deckkennung, und das Nein der Naht selbst.
// Ein gemeinsamer Lauf zeigte nur, dass **einer** von ihnen durchreicht (§10).
func TestJederFehlerwegDerDeckuebernahmeMeldetSich(t *testing.T) {
	t.Parallel()

	// **Ohne Naht** — eine Fläche, die sie nicht gesetzt hat, sagt das, statt
	// an einem Nullzeiger abzustürzen.
	ohne := app.NeueApp("")
	if ohne.DeckUebernehmen(t.Context(), "2", "7").Fehler == nil {
		t.Error("ohne Naht kam die Übernahme durch")
	}

	mit := app.NeueApp("")
	mit.MitDeckuebernahme(&uebernahmedoppel{})

	faelle := map[string][2]string{
		"krumme Profilkennung": {"keine Zahl", "7"},
		"krumme Deckkennung":   {"2", "keine Zahl"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := mit.DeckUebernehmen(t.Context(), fall[0], fall[1])
			if antwort.Fehler == nil {
				t.Fatal("die krumme Kennung kam durch")
			}

			// **Und die Befunde sind eine leere Liste und kein `null`** — die
			// Zusage aller Antworten dieses Pakets (§3).
			if antwort.Befunde == nil {
				t.Error("die Befunde kamen als null")
			}
		})
	}
}

// **Ein Nein der Naht kommt als Fehler und nicht als Befund.**
//
// **Der Unterschied ist der ganze Punkt** (§8): ein Befund sagt „das konnte
// nicht mit", ein Fehler „es ist nichts entstanden".
func TestEinNeinDerUebernahmeIstKeinBefund(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitDeckuebernahme(&uebernahmedoppel{
		fehler: errors.New("dafür fehlt dir die Rolle"),
	})

	antwort := anwendung.DeckUebernehmen(t.Context(), "2", "7")
	if antwort.Fehler == nil {
		t.Fatal("das Nein kam nicht durch")
	}

	if antwort.Deck != nil || len(antwort.Befunde) != 0 {
		t.Errorf("die Antwort trägt %+v", antwort)
	}
}

// **Und der Flächenweg reicht sie durch.**
func TestDerUebernahmewegUeberDieFlaeche(t *testing.T) {
	t.Parallel()

	// **Ohne Freigabe im frischen Bestand** — es gibt nur einen Nutzer, und
	// aus dem eigenen Profil übernimmt niemand.
	antwort := app.NutzerflaecheUeber(frischerBestand(t)).
		DeckUebernehmen(probekontext(t), "1", "1")

	if antwort.Fehler == nil {
		t.Error("die Übernahme aus dem eigenen Profil kam durch")
	}
}

// uebernahmedoppel ist eine Übernahme im Arbeitsspeicher.
type uebernahmedoppel struct {
	fehler  error
	befunde []string
}

func (d *uebernahmedoppel) DeckUebernehmen(
	_ context.Context, _, _ int64,
) (domain.Deck, []string, error) {
	if d.fehler != nil {
		return domain.Deck{}, nil, d.fehler
	}

	return domain.Deck{ID: 99, Name: "Nachtlauf", Spiel: "probe"},
		d.befunde, nil
}

// DeckeZurUebernahme nennt ein Deck — oder das Nein der Naht.
func (d *uebernahmedoppel) DeckeZurUebernahme(
	_ context.Context, _ int64,
) ([]domain.Deck, error) {
	if d.fehler != nil {
		return nil, d.fehler
	}

	return []domain.Deck{{ID: 7, Name: "Nachtlauf", Spiel: "probe"}}, nil
}

// **Und der Erfolgsweg gibt das Deck samt seiner Befunde.**
//
// # Warum er eigens dasteht
//
// **Weil die Läufe darüber alle scheitern** — und ein Weg, der jeden Fehler
// meldet und im Erfolg nichts zurückgibt, wäre dort grün (§10). Das ist
// dieselbe Lücke, die bei den schreibenden MCP-Werkzeugen schon einmal
// gefunden wurde (Commit `27f79352`).
func TestDerErfolgswegDerUebernahmeGibtDasDeck(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitDeckuebernahme(&uebernahmedoppel{
		befunde: []string{"Das Format bleibt im anderen Profil."},
	})

	antwort := anwendung.DeckUebernehmen(t.Context(), "2", "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Deck == nil || antwort.Deck.Name != "Nachtlauf" {
		t.Fatalf("angekommen ist %+v", antwort.Deck)
	}

	if len(antwort.Befunde) != 1 {
		t.Errorf("es kamen %d Befunde", len(antwort.Befunde))
	}
}

// **Und ohne Befunde eine leere Liste und kein `null`.**
//
// **Die Zusage aller Antworten dieses Pakets** (§3): über die Bindungen käme
// sonst ein `null` heraus, wo die Oberfläche eine Liste erwartet.
func TestEineUebernahmeOhneBefundeGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitDeckuebernahme(&uebernahmedoppel{})

	antwort := anwendung.DeckUebernehmen(t.Context(), "2", "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Befunde == nil {
		t.Error("die Befunde kamen als null")
	}
}
