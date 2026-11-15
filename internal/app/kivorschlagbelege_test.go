package app_test

// **Ein Vorschlag mit Belegen** — Werkstück A13.4 von Phase 5.3
// (#1452).
//
// # Der Satz, um den herum das gebaut ist
//
// **„A11 erhält dieselben Manifest-/Belegdaten wie A16; keine zweite
// Prompt-/Retrievalpipeline."**
//
// Bis A13.4 bekam ein Vorschlag **nichts** außer dem getippten Auftrag und
// dem Schema. Jetzt darf er sich auf denselben Wissensbestand stützen wie die
// Einschätzung — über **denselben** Kontextdienst, mit demselben Manifest und
// derselben Freigabe.
//
// # Und was das Werkstück trennt
//
// **Beobachtung und Vermutung.** Was das Modell vorschlägt, ist eine
// Vermutung und wird als `ki` verbucht (A11). Was es dabei zu sehen bekam,
// sind die Belege — sie stehen in der Antwort **neben** den Feldern und nicht
// als deren Beleg: kein Wort davon sagt, dass ein Wert daraus folgt.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Belegfrage bleibt alles, wie es war.**
//
// **Der Kontextdienst wird gar nicht gefragt** — ein Vorschlag ohne Belege
// ist der gewöhnliche Fall, und ein Ruf an den Index kostete Zeit für nichts.
func TestEinVorschlagOhneBelegfrageFragtDenIndexNicht(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	doppel := &recherchedoppel{paket: einProbepaket()}

	welt.anwendung.MitRecherche(doppel)

	antwort := welt.anwendung.KIVorschlagenMit(probekontext(t),
		app.KIVorschlagsbezug{
			KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
			Auftrag: "Schreibe einen Regeltext.",
		})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if doppel.gesehenWunsch.Aufgabe != "" {
		t.Errorf("der Index wurde gefragt: %+v", doppel.gesehenWunsch)
	}

	if len(antwort.Belege) != 0 {
		t.Errorf("es kamen Belege ohne Frage: %+v", antwort.Belege)
	}
}

// **Mit Belegfrage kommen die Belege — aus demselben Dienst wie A16.**
func TestEinVorschlagMitBelegfrageTraegtDieBelege(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	doppel := &recherchedoppel{paket: einProbepaket()}

	welt.anwendung.MitRecherche(doppel)

	antwort := welt.anwendung.KIVorschlagenMit(probekontext(t),
		app.KIVorschlagsbezug{
			KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
			Auftrag: "Schreibe einen Regeltext.", Belegfrage: "trampelschaden",
		})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Dieselbe Pipeline**: der Wunsch ging an den Kontextdienst, und zwar
	// mit dem Fall, der Kartentexte liest.
	if doppel.gesehenWunsch.Anfrage.Fall != domain.FallKarten {
		t.Errorf("gesucht wurde nach %q", doppel.gesehenWunsch.Anfrage.Fall)
	}

	if len(antwort.Belege) == 0 || len(antwort.Korpusstaende) == 0 {
		t.Errorf("die Belege fehlen: %+v", antwort)
	}

	// **Und der Text ging wirklich an das Modell.** Ohne diese Zeile sagte
	// der Lauf nur, dass Belege in der Antwort stehen — nicht, dass das
	// Modell sie gesehen hat.
	if len(welt.laufzeit.auftraege) != 1 {
		t.Fatalf("es gingen %d Aufträge hinaus", len(welt.laufzeit.auftraege))
	}

	// **Der Belegtext steht davor und der Auftrag dahinter** — ein Modell
	// liest von oben, und das Letzte, was es liest, soll die Aufgabe sein.
	gestellt := welt.laufzeit.auftraege[0]
	if !strings.Contains(gestellt, "[B1]") ||
		!strings.HasSuffix(gestellt, "Schreibe einen Regeltext.") {
		t.Errorf("der Auftrag war %q", gestellt)
	}
}

// **Ohne Freigabe bleibt der eigene Bestand außen.**
//
// **Auch hier** (A13.4): der Kartenfall liest die eigenen Karten, und ein
// Vorschlag, der sie ungefragt mitschickte, wäre genau der Weg, den das
// Issue ausschließt.
func TestEinVorschlagOhneFreigabeLaesstDenBestandAussen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	doppel := &recherchedoppel{paket: einProbepaket()}

	welt.anwendung.MitRecherche(doppel)

	welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
		Auftrag: "Schreibe.", Belegfrage: "trampelschaden",
	})

	if len(doppel.gesehenWunsch.Anfrage.Freigegeben) != 0 {
		t.Errorf("freigegeben war %v",
			doppel.gesehenWunsch.Anfrage.Freigegeben)
	}

	// **Und mit Kreuz kommt er dazu** — die Gegenprobe.
	welt.anwendung.KIVorschlagenMit(probekontext(t), app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
		Auftrag: "Schreibe.", Belegfrage: "trampelschaden",
		Freigegeben: []string{string(domain.QuelleEigeneKarte)},
	})

	if len(doppel.gesehenWunsch.Anfrage.Freigegeben) != 1 {
		t.Errorf("freigegeben war %v",
			doppel.gesehenWunsch.Anfrage.Freigegeben)
	}
}

// **Jeder Fehlerweg des Belegvorschlags reicht durch.**
//
// # Warum das drei Fälle sind und nicht einer
//
// **Weil sie an drei verschiedenen Stellen entstehen** — vor dem Paket, beim
// Paket und danach. Ein Lauf, der nur einen prüfte, sagte nichts über die
// anderen; und wer einen Fehler verschluckte, bekäme einen Vorschlag ohne
// die Belege, die er verlangt hat — ohne zu erfahren, dass sie fehlen.
func TestJederFehlerwegDesBelegvorschlagsReichtDurch(t *testing.T) {
	t.Parallel()

	t.Run("ohne Recherche", func(t *testing.T) {
		t.Parallel()

		welt := einekiwelt(t)

		// **Die Recherche ausdrücklich abhängen** — `KIAnwendungUeber`
		// verdrahtet sie, und ohne diese Zeile prüfte der Lauf den
		// gewöhnlichen Weg.
		welt.anwendung.MitRecherche(nil)

		if antwort := welt.anwendung.KIVorschlagenMit(probekontext(t),
			app.KIVorschlagsbezug{
				KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
				Auftrag: "x", Belegfrage: "trampelschaden",
			}); antwort.Fehler == nil {
			t.Fatal("ohne Recherche kam ein Vorschlag mit Belegen")
		}
	})

	t.Run("die Recherche klemmt", func(t *testing.T) {
		t.Parallel()

		welt := einekiwelt(t)
		welt.anwendung.MitRecherche(&recherchedoppel{
			paketfehler: domain.MitGrund(domain.ErrValidation, nil,
				"der Index ist nicht gebaut"),
		})

		if antwort := welt.anwendung.KIVorschlagenMit(probekontext(t),
			app.KIVorschlagsbezug{
				KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
				Auftrag: "x", Belegfrage: "trampelschaden",
			}); antwort.Fehler == nil {
			t.Fatal("der Fehler des Index kam nicht durch")
		}
	})

	t.Run("ohne Profil", func(t *testing.T) {
		t.Parallel()

		welt := einekiwelt(t)

		if antwort := welt.anwendung.KIVorschlagenMit(t.Context(),
			app.KIVorschlagsbezug{
				KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
				Auftrag: "x",
			}); antwort.Fehler == nil {
			t.Fatal("der Ruf ohne Profil kam durch")
		}
	})

	t.Run("ein Anbieter, den es nicht gibt", func(t *testing.T) {
		t.Parallel()

		welt := einekiwelt(t)

		if antwort := welt.anwendung.KIVorschlagenMit(probekontext(t),
			app.KIVorschlagsbezug{
				KarteID: welt.karte, Anbieter: "gibtesnicht",
				Alias: "gemma4:8b", Auftrag: "x",
			}); antwort.Fehler == nil {
			t.Fatal("der unbekannte Anbieter kam durch")
		}
	})

	t.Run("das Modell klemmt", func(t *testing.T) {
		t.Parallel()

		welt := einekiwelt(t)
		welt.laufzeit.redefehler = domain.MitGrund(domain.ErrValidation, nil,
			"keine Antwort")

		if antwort := welt.anwendung.KIVorschlagenMit(probekontext(t),
			app.KIVorschlagsbezug{
				KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
				Auftrag: "x",
			}); antwort.Fehler == nil {
			t.Fatal("der Fehler des Modells kam nicht durch")
		}
	})
}
