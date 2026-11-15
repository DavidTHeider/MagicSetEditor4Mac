package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Einen Befehl ausführen — mit Vorschau, Bestätigung und Abbruch** —
// Werkstück U12 von Phase 5.1 (#1423).

func TestEinLesenderBefehlLaeuftUndBringtSeinenText(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("").MitBefehlen(probebefehle())

	// **Die Argumente bleiben Daten** — Anführungszeichen, Semikolon, Shellzeichen.
	ergebnis := anwendung.BefehlAusfuehren(t.Context(), "lesen.sets", `{"suche":"a\"; rm -rf / # $(x)"}`, false)
	if ergebnis.Fehler != nil || !strings.Contains(ergebnis.Text, `rm -rf / # $(x)`) || ergebnis.Name != "lesen.sets" {
		t.Fatalf("%+v", ergebnis)
	}

	// Ohne Argumente heißt: ein leeres Objekt.
	if ergebnis := anwendung.BefehlAusfuehren(t.Context(), "lesen.sets", "", false); ergebnis.Fehler != nil ||
		!strings.Contains(ergebnis.Text, `"argumente":{}`) {
		t.Errorf("ohne Argumente: %+v", ergebnis)
	}

	faelle := map[string]struct {
		name, argumente string
		bestaetigt      bool
		code            string
	}{
		"unbekannt":       {"lesen.nichts", "{}", false, "nichtGefunden"},
		"kein JSON":       {"lesen.sets", "{", false, "validierung"},
		"kein Objekt":     {"lesen.sets", "[1]", false, "validierung"},
		"gefährlich ohne": {"schreiben.karte.loeschen", "{}", false, "validierung"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ergebnis := anwendung.BefehlAusfuehren(t.Context(), fall.name, fall.argumente, fall.bestaetigt)
			if ergebnis.Fehler == nil || string(ergebnis.Fehler.Code) != fall.code {
				t.Errorf("%+v", ergebnis)
			}
		})
	}

	// Bestätigt läuft auch das Gefährliche — das Fenster darf schreiben.
	if ergebnis := anwendung.BefehlAusfuehren(t.Context(), "schreiben.karte.loeschen", "{}", true); ergebnis.Fehler != nil || ergebnis.Text != "gelöscht" {
		t.Errorf("bestätigt: %+v", ergebnis)
	}

	// Der Fehler eines Werkzeugs kommt übersetzt an — mit der Dauer bis dahin.
	if ergebnis := app.NeueBindungen(anwendung).BefehlAusfuehren(t.Context(), "lesen.kaputt", "{}", false); ergebnis.Fehler == nil ||
		ergebnis.Name != "lesen.kaputt" {
		t.Errorf("Fensterbindung: %+v", ergebnis)
	}
}

func TestEinLaufendenBefehlKannManAbbrechenUndKeinenZweitenStarten(t *testing.T) {
	t.Parallel()

	begonnen := make(chan struct{})
	anwendung := app.NeueApp("").MitBefehlen([]app.Befehl{{
		Name: "lesen.langsam",
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			close(begonnen)
			<-ctx.Done()

			return "", ctx.Err()
		},
	}})

	fertig := make(chan app.BefehlsergebnisAntwort, 1)

	go func() { fertig <- anwendung.BefehlAusfuehren(t.Context(), "lesen.langsam", "{}", false) }()

	<-begonnen

	// **Ein Vorgang zur Zeit** (§7, Regel 4).
	if zweiter := anwendung.BefehlAusfuehren(t.Context(), "lesen.langsam", "{}", false); zweiter.Fehler == nil || zweiter.Fehler.Code != "konflikt" {
		t.Errorf("zweiter Lauf: %+v", zweiter)
	}

	if antwort := app.NeueBindungen(anwendung).BefehlAbbrechen(t.Context()); antwort.Fehler != nil {
		t.Errorf("abbrechen: %+v", antwort)
	}

	select {
	case ergebnis := <-fertig:
		if ergebnis.Fehler == nil || ergebnis.Fehler.Code != "abgebrochen" {
			t.Errorf("abgebrochen: %+v", ergebnis)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("der Abbruch kam nicht an")
	}

	// Ohne laufenden Befehl ist Abbrechen nichts — und kein Fehler.
	if antwort := anwendung.BefehlAbbrechen(t.Context()); antwort.Fehler != nil {
		t.Errorf("Abbruch ohne Lauf: %+v", antwort)
	}
}

// **Wer nicht schreiben darf, ruft keinen schreibenden Befehl** — die
// Autorisierung steht vor der Bestätigung und vor dem Ruf.
func TestOhneSchreibrechtBleibtEinSchreibenderBefehlAus(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)
	anwendung := app.NeueApp("").MitBefehlen(probebefehle())
	anwendung.MitAutorisierung(service.NeueAutorisierung(
		stolperzeiger{fehler: errors.New("der Zeiger klemmt")},
		service.NeueFreigabenverwaltung(storage.NeueFreigabenablage(db), storage.NeueNutzerablage(db))))

	if ergebnis := anwendung.BefehlAusfuehren(probekontext(t), "schreiben.feld", "{}", false); ergebnis.Fehler == nil {
		t.Errorf("ohne Schreibrecht: %+v", ergebnis)
	}

	// Lesen geht weiter — die Autorisierung fragt nur, wer schreibt.
	if ergebnis := anwendung.BefehlAusfuehren(probekontext(t), "lesen.sets", "{}", false); ergebnis.Fehler != nil {
		t.Errorf("lesend: %+v", ergebnis)
	}
}
