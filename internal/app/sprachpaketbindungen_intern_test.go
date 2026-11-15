package app

// **Eine eigene Übersetzung am Fenster, mit Dialog** — LOC-02
// (#1688).

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func mitSprachpaketen(t *testing.T, gewaehlt *string, wahlfehler error) *Bindungen {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.MitSprachpaketen(NeueSprachpaketeUeber(probedatenbank(t)))
	anwendung.fenster = probekontext(t)
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return *gewaehlt, wahlfehler
	})

	return NeueBindungen(anwendung)
}

func TestEineUebersetzungGehtUeberDenDialogHinausUndZurueck(t *testing.T) {
	var gewaehlt string

	bindungen := mitSprachpaketen(t, &gewaehlt, nil)

	if angelegt := bindungen.SprachpaketAnlegen(t.Context(), "da", "Dansk", "en"); angelegt.Fehler != nil {
		t.Fatalf("angelegt: %+v", angelegt.Fehler)
	}

	lizenz := "CC0-1.0"
	if geaendert := bindungen.SprachpaketAendern(t.Context(), "da", SprachpaketangabenDTO{Lizenz: &lizenz},
		map[string]string{"allgemein.arbeitsplatz.keinSet": "Intet sæt åbent"}); geaendert.Fehler != nil {
		t.Fatalf("geändert: %+v", geaendert.Fehler)
	}

	if liste := bindungen.Sprachpakete(t.Context()); len(liste.Pakete) != 1 {
		t.Errorf("die Liste: %+v", liste)
	}

	// **Weggeklickt ist kein Fehler.**
	if abgebrochen := bindungen.SprachpaketAusgabeSpeichern(t.Context(), "da"); !abgebrochen.Abgebrochen ||
		abgebrochen.Fehler != nil {
		t.Errorf("weggeklickt: %+v", abgebrochen)
	}

	gewaehlt = filepath.Join(t.TempDir(), "da.json")

	if gespeichert := bindungen.SprachpaketAusgabeSpeichern(t.Context(), "da"); gespeichert.Fehler != nil ||
		gespeichert.Pfad != gewaehlt {
		t.Fatalf("gespeichert: %+v", gespeichert)
	}

	if geloescht := bindungen.SprachpaketLoeschen(t.Context(), "da"); geloescht.Fehler != nil {
		t.Fatalf("gelöscht: %+v", geloescht.Fehler)
	}

	if weg := bindungen.Sprachpaket(t.Context(), "da"); weg.Fehler == nil {
		t.Error("das gelöschte Paket ist noch da")
	}

	eingelesen := bindungen.SprachpaketEinlesenOeffnen(t.Context(), false)
	if eingelesen.Fehler != nil || eingelesen.Paket.Uebersetzt != 1 {
		t.Fatalf("eingelesen: %+v", eingelesen)
	}

	if ersetzt := bindungen.SprachpaketEinlesenOeffnen(t.Context(), true); ersetzt.Fehler != nil {
		t.Errorf("ersetzt: %+v", ersetzt.Fehler)
	}
}

// **Ein Dialog, der scheitert, ist ein Befund** — mit leeren Listen.
func TestEinGescheiterterUebersetzungsdialogIstEinBefund(t *testing.T) {
	var gewaehlt string

	bindungen := mitSprachpaketen(t, &gewaehlt, errors.New("kein Fenster"))

	for name, antwort := range map[string]SprachpaketAntwort{
		"speichern": bindungen.SprachpaketAusgabeSpeichern(t.Context(), "da"),
		"öffnen":    bindungen.SprachpaketEinlesenOeffnen(t.Context(), false),
	} {
		if antwort.Fehler == nil || antwort.Texte == nil || antwort.Abgebrochen {
			t.Errorf("%s: %+v", name, antwort)
		}
	}
}
