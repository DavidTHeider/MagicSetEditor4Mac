package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Bindungen des Paketverwalters (#54).
//
// **Sie reichen durch — und der Test prüft, dass sie es tun.** Ein Durchreicher,
// der einen leeren Umschlag zurückgäbe, ohne die Anwendung zu erreichen, wäre
// sonst grün; genau das hat der Nachlauf des Zweitblicks über Phase 5
// angestrichen.
func TestDiePaketbindungenReichenDurch(t *testing.T) {
	fuehrer := &paketfuehrer{pakete: []domain.Paket{
		{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
	}}

	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	bindungen := app.NeueBindungen(anwendung)

	t.Run("Pakete", func(t *testing.T) {
		antwort := bindungen.Pakete(t.Context())
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if len(antwort.Pakete) != 1 || antwort.Pakete[0].Name != "probe" {
			t.Errorf("%+v", antwort.Pakete)
		}
	})

	t.Run("PaketPlanWaehlen fragt nach einem Verzeichnis und plant", func(t *testing.T) {
		var gewaehlt app.Dateiauswahl

		anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
			gewaehlt = auswahl

			return "/quelle/neu.mse-style", nil
		})

		fuehrer.plan = service.Paketplan{Vorgang: service.PaketvorgangEinbinden, Paket: "neu.mse-style"}

		antwort := bindungen.PaketPlanWaehlen(t.Context())
		if antwort.Fehler != nil || antwort.Plan == nil || antwort.Plan.Quelle != "/quelle/neu.mse-style" {
			t.Fatalf("%+v", antwort)
		}

		// **Geplant, nicht eingebunden** — DOM-01.AK-02 (#1696).
		if fuehrer.geplant != "einbinden /quelle/neu.mse-style" || fuehrer.installiert != "" {
			t.Errorf("geplant %q, installiert %q", fuehrer.geplant, fuehrer.installiert)
		}

		// **Der Ordnerdialog, nicht der Dateidialog.** Ein MSE-Vorlagenpaket ist
		// ein Verzeichnis; der Dateidialog zeigte es als undurchdringlichen
		// Ordner, in den man nur hineingehen kann.
		if !gewaehlt.Verzeichnis {
			t.Errorf("gefragt wurde nach %+v", gewaehlt)
		}
	})

	t.Run("ein weggeklickter Dialog ist ein Abbruch", func(t *testing.T) {
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return "", nil })

		if antwort := bindungen.PaketPlanWaehlen(t.Context()); !antwort.Abgebrochen || antwort.Fehler != nil {
			t.Errorf("%+v", antwort)
		}
	})

	t.Run("ein klagender Dialog", func(t *testing.T) {
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", errors.New("der Dialog ging nicht auf")
		})

		if antwort := bindungen.PaketPlanWaehlen(t.Context()); antwort.Fehler == nil {
			t.Error("der Fehler ging verloren")
		}
	})

	t.Run("Plan, Einbinden, Ersetzen, Zurück", func(t *testing.T) {
		if antwort := bindungen.PaketPlan(t.Context(), "zurueck", "mse-game", "probe", ""); antwort.Plan == nil ||
			fuehrer.geplant != "zurueck mse-game probe" {
			t.Errorf("Plan: %+v", antwort)
		}

		if antwort := bindungen.PaketEinbindenVon(t.Context(), "/quelle/neu.mse-style"); fuehrer.installiert != "/quelle/neu.mse-style" {
			t.Errorf("Einbinden: %+v", antwort)
		}

		if antwort := bindungen.PaketErsetzen(t.Context(), "/quelle/neu.mse-style", false); antwort.Paket == nil {
			t.Errorf("Ersetzen: %+v", antwort)
		}

		if antwort := bindungen.PaketZurueck(t.Context(), "mse-game", "probe"); antwort.Paket == nil ||
			fuehrer.entfernt != "zurück mse-game probe" {
			t.Errorf("Zurück: %+v", antwort)
		}
	})

	t.Run("PaketEntfernen", func(t *testing.T) {
		if antwort := bindungen.PaketEntfernen(t.Context(), "mse-game", "probe"); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if fuehrer.entfernt != "mse-game/probe" {
			t.Errorf("entfernt: %q", fuehrer.entfernt)
		}
	})
}

// **Ein Fehler beim Auffrischen hält den Katalog nicht auf** — er wird
// gemeldet, und die Liste wird gar nicht erst gelesen.
func TestEinFehlerBeimAuffrischenBeendetDieAntwort(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{fehler: errors.New("die Platte schweigt")})

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if len(antwort.Pakete) != 0 {
		t.Errorf("trotzdem gelesen: %+v", antwort.Pakete)
	}
}

// **Ein Fehler beim Installieren kommt übersetzt an.**
func TestEinFehlerBeimInstallierenWirdUebersetzt(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{fehler: domain.ErrConflict})

	if antwort := anwendung.PaketInstallieren(t.Context(), "/quelle/x.mse-style"); antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Und ein Fehler beim Lesen der Liste ebenso.**
//
// Der zweite Weg derselben Antwort: das Auffrischen glückt, der Bestand
// schweigt erst danach.
func TestEinFehlerBeimLesenDerListeKommtAn(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{listenfehler: errors.New("der Bestand schweigt")})

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Pakete == nil {
		t.Error("nil statt leerer Liste")
	}
}

// **Die Abhängigkeiten kommen an die Oberfläche** — als Angaben, nicht als Satz.
func TestAbhaengigkeitenGebenAngaben(t *testing.T) {
	fuehrer := &paketfuehrer{abhaengigkeiten: []service.Abhaengigkeitsbefund{{
		Paket: "magic-m15.mse-style", Braucht: "magic.mse-game",
		Art: service.AbhaengigkeitZuAlt, Verlangt: "2024-01-01", Da: "2009-07-23",
	}}}

	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.Abhaengigkeiten(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Befunde) != 1 {
		t.Fatalf("Befunde: %+v", antwort.Befunde)
	}

	if antwort.Befunde[0].Art != "zualt" || antwort.Befunde[0].Da != "2009-07-23" {
		t.Errorf("Befund: %+v", antwort.Befunde[0])
	}

	// **Und über die Bindung dieselbe Auskunft** — ein Durchreicher, der ins
	// Leere ginge, wäre sonst grün.
	if gebunden := app.NeueBindungen(anwendung).Abhaengigkeiten(t.Context()); len(gebunden.Befunde) != 1 {
		t.Errorf("gebunden: %+v", gebunden)
	}
}

// **Ein Fehler wird übersetzt, und die Liste bleibt eine Liste.**
func TestEinFehlerBeiDenAbhaengigkeitenWirdUebersetzt(t *testing.T) {
	anwendung := &app.App{}
	anwendung.MitPaketen(&paketfuehrer{fehler: errors.New("der Katalog schweigt")})

	antwort := anwendung.Abhaengigkeiten(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Befunde == nil {
		t.Error("nil statt leerer Liste")
	}
}
