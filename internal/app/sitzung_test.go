package app_test

// **Die lokale Sitzung am Fenster** — ID-02
// (#1701):
// gesperrt erreicht keine Abfrage den Bestand, und ⌘Q hebt die Sperre nicht
// auf.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

const probekennwort = "ein-langes-kennwort"

func TestGesperrtErreichtNichtsDenBestand(t *testing.T) {
	anwendung := faktoranwendung(t)
	ctx := t.Context()

	stand := anwendung.Sitzungsstand(ctx)
	if stand.Zustand != "angemeldet" || stand.Kennwort || stand.Sensor || stand.Fehler != nil {
		t.Fatalf("Beginn: %+v", stand)
	}

	// **Ohne Kennwort und ohne Sensor sperrt nichts.**
	if antwort := anwendung.Sperren(ctx); antwort.Fehler == nil {
		t.Error("ohne Kennwort gesperrt")
	}

	if antwort := anwendung.KennwortWechseln(ctx, "", probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if antwort := anwendung.Sperren(ctx); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if stand := anwendung.Sitzungsstand(ctx); stand.Zustand != "gesperrt" || !stand.Kennwort {
		t.Errorf("gesperrt: %+v", stand)
	}

	if sets := anwendung.Sets(ctx); sets.Fehler == nil {
		t.Error("gesperrt las eine Bindung den Bestand")
	}

	if antwort := anwendung.Entsperren(ctx, "falsch"); antwort.Fehler == nil || antwort.Fehler.Feld != "kennwort" {
		t.Errorf("falsches Kennwort: %+v", antwort.Fehler)
	}

	if antwort := anwendung.EntsperrenMitSensor(ctx); antwort.Fehler == nil {
		t.Error("ohne eingerichteten Sensor entsperrt")
	}

	if antwort := anwendung.Entsperren(ctx, probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if sets := anwendung.Sets(ctx); sets.Fehler != nil {
		t.Errorf("entsperrt: %v", sets.Fehler)
	}
}

func TestEineSperreUeberlebtDenNeustart(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	ctx := t.Context()

	erste := app.NeueApp(ort)
	erste.MitSchluesselbund(neuerBundspeicher())

	if fehler := erste.Starten(ctx); fehler != nil {
		t.Fatal(fehler)
	}

	for _, antwort := range []app.SchreibAntwort{
		erste.KennwortWechseln(ctx, "", probekennwort), erste.Sperren(ctx),
	} {
		if antwort.Fehler != nil {
			t.Fatal(antwort.Fehler)
		}
	}

	if fehler := erste.Beenden(); fehler != nil {
		t.Fatal(fehler)
	}

	zweite := app.NeueApp(ort)
	zweite.MitSchluesselbund(neuerBundspeicher())

	if fehler := zweite.Starten(ctx); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if fehler := zweite.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	if stand := zweite.Sitzungsstand(ctx); stand.Zustand != "gesperrt" {
		t.Errorf("nach dem Neustart: %+v", stand)
	}
}

func TestAbmeldenUndLokalAnmelden(t *testing.T) {
	anwendung := faktoranwendung(t)
	ctx := t.Context()

	nutzerID := anwendung.Sitzungsstand(ctx).NutzerID
	if antwort := anwendung.KennwortWechseln(ctx, "", probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if antwort := anwendung.Abmelden(ctx); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if stand := anwendung.Sitzungsstand(ctx); stand.Zustand != "abgemeldet" || stand.NutzerID != "0" {
		t.Errorf("abgemeldet: %+v", stand)
	}

	if antwort := anwendung.Abmelden(ctx); antwort.Fehler == nil {
		t.Error("zweimal abgemeldet")
	}

	for name, antwort := range map[string]app.SchreibAntwort{
		"keine Kennung":    anwendung.LokalAnmelden(ctx, "x", probekennwort, ""),
		"falsches Wort":    anwendung.LokalAnmelden(ctx, nutzerID, "falsch", ""),
		"Zettel ohne Zahl": anwendung.KennwortWiederherstellen(ctx, "x", "code", "neu-und-lang-genug"),
		"falscher Zettel":  anwendung.KennwortWiederherstellen(ctx, nutzerID, "code", "neu-und-lang-genug"),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s: angenommen", name)
		}
	}

	if antwort := anwendung.LokalAnmelden(ctx, nutzerID, probekennwort, ""); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if stand := anwendung.Sitzungsstand(ctx); stand.Zustand != "angemeldet" || stand.NutzerID != nutzerID {
		t.Errorf("angemeldet: %+v", stand)
	}
}

// **Ohne Fenster keine Sitzung** — die Flächen ohne Fenster bringen ihre
// eigene Anmeldung mit.
func TestOhneFensterGibtEsKeineSitzung(t *testing.T) {
	anwendung := app.NeueApp("")

	if stand := anwendung.Sitzungsstand(t.Context()); stand.Zustand != "angemeldet" {
		t.Errorf("%+v", stand)
	}

	if antwort := anwendung.Sperren(t.Context()); antwort.Fehler == nil || antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("%+v", antwort.Fehler)
	}
}
