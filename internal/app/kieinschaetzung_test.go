package app_test

// **Der belegte Regelhinweis** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Was hier geprüft wird
//
// **Der Weg und nicht die Prüfung.** Ob eine erfundene Regelnummer ein Befund
// ist, steht in `internal/domain`; hier steht, dass sie einen **bekommt** —
// dass also der Weg vom Kontextpaket über das Modell bis zur Prüfung wirklich
// zusammenhängt. Bis A13.2a hing er nicht: `domain.PruefeBelege` hatte keinen
// Produktivaufrufer.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// hinweiswelt ist eine KI-Welt samt Kontextdienst.
type hinweiswelt struct {
	kiwelt

	recherche *recherchedoppel
}

// einehinweiswelt baut sie — mit zwei Belegen im Paket.
func einehinweiswelt(t *testing.T) hinweiswelt {
	t.Helper()

	welt := einekiwelt(t)

	doppel := &recherchedoppel{paket: service.Kontextpaket{
		Manifest: domain.Kontextmanifest{
			PaketID: "p1", Zaehlweise: "probe",
			Belege: []domain.Beleg{{
				Kennung: "B1", Quelle: domain.QuelleRegelwerk,
				Titel:       "509.1 Blocks erklären",
				Ausschnitt:  "509.1. Zuerst prüft das Spiel, ob die Blocks legal sind.",
				Regelnummer: "509.1",
			}},
			Korpusstaende: []domain.Korpusstand{{
				Quelle: domain.QuelleRegelwerk, Revision: "2026-06-07",
			}},
		},
		Kontext: "[B1] 509.1. Zuerst prüft das Spiel …",
	}}

	welt.anwendung.MitRecherche(doppel)

	return hinweiswelt{kiwelt: welt, recherche: doppel}
}

// modellSagt stellt die Antwort des Modells ein.
func (w hinweiswelt) modellSagt(satz string) {
	w.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"einschaetzung":` + mussTextJSON(satz) +
			`,"einschraenkungen":["Nur die mitgegebenen Stellen."]}`,
		Modell: "gemma4:8b",
	}
}

// mussTextJSON macht aus einem Satz ein JSON-Literal.
func mussTextJSON(satz string) string {
	return `"` + strings.ReplaceAll(satz, `"`, `\"`) + `"`
}

// hinweis fragt nach einem Regelhinweis.
func (w hinweiswelt) hinweis(t *testing.T) app.KIEinschaetzungAntwort {
	t.Helper()

	return w.anwendung.KIEinschaetzen(probekontext(t),
		app.KIEinschaetzungsbezug{
			Art: string(domain.EinschaetzungRegel), KarteID: w.karte,
			Anbieter: "ollama", Alias: "gemma4:8b",
			Frage: "Wie wirkt die Fähigkeit beim Blocken?",
		})
}

// **Ein Hinweis, der nur vergebene Belege nennt, gilt als belegt — und
// bringt sie mit.**
func TestEinRegelhinweisNenntSeineBelege(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Der Block ist legal [B1].")

	antwort := welt.hinweis(t)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !antwort.Belegt {
		t.Errorf("nicht belegt: %+v", antwort.Befunde)
	}

	if len(antwort.Belege) != 1 || antwort.Belege[0].Kennung != "B1" {
		t.Errorf("die Belege sind %+v", antwort.Belege)
	}

	if len(antwort.Korpusstaende) != 1 {
		t.Errorf("die Stände sind %+v", antwort.Korpusstaende)
	}

	if !strings.Contains(antwort.Vorbehalt, "keine amtliche") {
		t.Errorf("der Vorbehalt fehlt: %q", antwort.Vorbehalt)
	}

	if antwort.Einschraenkungen == nil {
		t.Error("die Einschränkungen fehlen")
	}

	// **Und das Modell bekam die Anweisung dieser Aufgabe.**
	//
	// **Nicht die des Kartenvorschlags**: „Du füllst Felder einer
	// Sammelkarte" wäre die erste Zeile, die das Modell liest — vor einer
	// Frage, die von etwas anderem handelt.
	if len(welt.laufzeit.anweisungen) != 1 ||
		welt.laufzeit.anweisungen[0] == app.KIKartenanweisung {
		t.Errorf("die Anweisungen sind %+v", welt.laufzeit.anweisungen)
	}
}

// **Und die Recherche fragt den Fall, der alle drei Korpora liest.**
func TestEinRegelhinweisFragtDenRegelhinweisfall(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Der Block ist legal [B1].")

	if antwort := welt.hinweis(t); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if fall := welt.recherche.gesehenWunsch.Anfrage.Fall; fall != domain.FallRegelhinweis {
		t.Errorf("gefragt wurde %q", fall)
	}
}

// **Eine erfundene Regelnummer steht als Befund da — und der Satz bleibt.**
func TestEineErfundeneRegelnummerStehtAlsBefund(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Das steht in 702.19b und in [B1].")

	antwort := welt.hinweis(t)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Belegt {
		t.Fatal("die erfundene Nummer kam als belegt durch")
	}

	if !strings.Contains(antwort.Satz, "702.19b") {
		t.Errorf("der Satz wurde verstümmelt: %q", antwort.Satz)
	}

	if len(antwort.Befunde) != 1 ||
		antwort.Befunde[0].Art != domain.BefundErfundeneRegelnummer {
		t.Errorf("die Befunde sind %+v", antwort.Befunde)
	}
}

// **Und ein Hinweis ändert die Karte nicht.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil das Issue es als Abnahme führt**: „Keine Regel-/Sammlungs-/
// Spielzustandsänderung beim Analysieren." Eine Zusage, die nur daher gilt,
// dass niemand einen Schreibweg gebaut hat, ist keine — sie hält, bis jemand
// einen baut.
func TestEineEinschaetzungSchreibtNichtsAnDieKarte(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Der Block ist legal [B1].")

	vorher := welt.kartenwerte(t)

	if antwort := welt.hinweis(t); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if nachher := welt.kartenwerte(t); nachher != vorher {
		t.Errorf("die Karte trägt %q statt %q", nachher, vorher)
	}
}

// **Eine Art, die es nicht gibt, wird abgewiesen — vor jedem Ruf.**
func TestEineUnbekannteArtWirdAbgewiesen(t *testing.T) {
	welt := einehinweiswelt(t)

	antwort := welt.anwendung.KIEinschaetzen(probekontext(t),
		app.KIEinschaetzungsbezug{
			Art: "regelchen", KarteID: welt.karte, Anbieter: "ollama",
			Alias: "gemma4:8b", Frage: "Was gilt?",
		})
	if antwort.Fehler == nil {
		t.Fatal("die erfundene Art kam durch")
	}

	if gefragt := len(welt.laufzeit.adressen); gefragt != 0 {
		t.Errorf("das Modell wurde %d-mal gefragt", gefragt)
	}
}

// **Eine Balancefrage geht über die Ähnlichkeitsrecherche.**
func TestEineBalancefrageGehtUeberDieAehnlichkeit(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Sie kostet eins mehr [B1].")

	antwort := welt.anwendung.KIEinschaetzen(probekontext(t),
		app.KIEinschaetzungsbezug{
			Art: string(domain.EinschaetzungBalance), KarteID: welt.karte,
			Anbieter: "ollama", Alias: "gemma4:8b",
			Frage: "Wie steht sie da?",
		})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if fall := welt.recherche.gesehenWunsch.Anfrage.Fall; fall != domain.FallAehnlich {
		t.Errorf("gefragt wurde %q", fall)
	}

	// **Und die Ausgangskarte reist mit** — eine Ähnlichkeitsfrage ohne sie
	// wäre gar keine.
	if welt.recherche.gesehenWunsch.Anfrage.AusgangskarteID != welt.karte {
		t.Errorf("die Ausgangskarte ist %d",
			welt.recherche.gesehenWunsch.Anfrage.AusgangskarteID)
	}
}

// **Jeder Fehlerweg reicht durch.**
func TestJederFehlerwegDerEinschaetzungReichtDurch(t *testing.T) {
	for name, fall := range map[string]struct {
		richten func(*testing.T, hinweiswelt)
		kern    string
	}{
		"ohne Kontextdienst": {
			func(_ *testing.T, welt hinweiswelt) {
				welt.anwendung.MitRecherche(nil)
			}, "Bestand",
		},
		"die Recherche klemmt": {
			func(_ *testing.T, welt hinweiswelt) {
				welt.recherche.paketfehler = fmt.Errorf(
					"%w: kein Index", domain.ErrNotFound)
			}, "kein Index",
		},
		"das Modell klemmt": {
			func(_ *testing.T, welt hinweiswelt) {
				welt.laufzeit.redefehler = fmt.Errorf(
					"%w: keine Antwort", domain.ErrValidation)
			}, "keine Antwort",
		},
		"das Modell gibt kein JSON": {
			func(_ *testing.T, welt hinweiswelt) {
				welt.laufzeit.antwort = app.KIRohantwort{Inhalt: "kein JSON"}
			}, "kein JSON",
		},
		"die Antwort ist abgeschnitten": {
			func(_ *testing.T, welt hinweiswelt) {
				welt.laufzeit.antwort = app.KIRohantwort{
					Inhalt: `{"einschaetzung":"halb`, Abgeschnitten: true,
				}
			}, "abgeschnitten",
		},
		"das Modell sagt nichts": {
			func(_ *testing.T, welt hinweiswelt) {
				welt.laufzeit.antwort = app.KIRohantwort{
					Inhalt: `{"einschaetzung":"  "}`,
				}
			}, "nichts gesagt",
		},
		"die Laufzeit ist aus": {
			func(t *testing.T, welt hinweiswelt) {
				t.Helper()

				if fehler := storage.NeueEinstellungen(welt.db).Setzen(
					probekontext(t), domain.KILaufzeit, "aus"); fehler != nil {
					t.Fatalf("abschalten: %v", fehler)
				}
			}, "eingeschaltet",
		},
		"den Alias gibt es nicht": {
			func(_ *testing.T, welt hinweiswelt) {
				welt.laufzeit.modelle = nil
			}, "den Alias",
		},
	} {
		t.Run(name, func(t *testing.T) {
			welt := einehinweiswelt(t)
			welt.modellSagt("Der Block ist legal [B1].")
			fall.richten(t, welt)

			antwort := welt.hinweis(t)
			if antwort.Fehler == nil {
				t.Fatalf("der Fehler kam nicht durch: %+v", antwort)
			}

			if !strings.Contains(antwort.Fehler.Meldung, fall.kern) {
				t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
			}
		})
	}
}

// **Ohne Profil im Ruf geht nichts.**
func TestEineEinschaetzungOhneProfilGehtNicht(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Der Block ist legal [B1].")

	antwort := welt.anwendung.KIEinschaetzen(t.Context(),
		app.KIEinschaetzungsbezug{
			Art: string(domain.EinschaetzungRegel), KarteID: welt.karte,
			Anbieter: "ollama", Alias: "gemma4:8b", Frage: "Was gilt?",
		})
	if antwort.Fehler == nil {
		t.Fatal("der Ruf ohne Profil kam durch")
	}
}

// **Und die Fensterbindung reicht durch.**
func TestDieEinschaetzungsbindungReichtDurch(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Der Block ist legal [B1].")

	// **Das Fenster kennt sein Profil aus dem Wechsel** — ohne ihn trüge
	// [app.App.kontext] keines, und die Bindung scheiterte daran statt an
	// dem, was sie prüfen soll.
	if gewechselt := welt.anwendung.ProfilWechseln(probekontext(t), "1", ""); gewechselt.Fehler != nil {
		t.Fatalf("wechseln: %+v", gewechselt.Fehler)
	}

	if antwort := app.NeueBindungen(welt.anwendung).KIEinschaetzen(t.Context(),
		app.KIEinschaetzungsbezug{
			Art: string(domain.EinschaetzungRegel), KarteID: welt.karte,
			Anbieter: "ollama", Alias: "gemma4:8b", Frage: "Was gilt?",
		}); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}
}
