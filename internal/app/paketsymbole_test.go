package app_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func werkstattMitTexten() *werkstattprobe {
	return &werkstattprobe{texte: []service.Paketdateitext{
		{Name: "game", Inhalt: "card field:\n\tname: title\nscript:\n\tx := card.title\n"},
	}}
}

// **Die Symbole eines Pakets kommen ans Fenster — mit Art, Datei und
// Stelle** (U7, #1418).
func TestDiePaketsymboleKommenAnsFenster(t *testing.T) {
	probe := werkstattMitTexten()
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(probe)

	antwort := anwendung.Paketsymbole(t.Context(), "/pakete/probe.mse-game")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	erwartet := []app.PaketsymbolDTO{
		{Name: "title", Art: "card field", Datei: "game", Zeile: 2, Spalte: 2},
		{Name: "x", Art: "script", Datei: "game", Zeile: 4, Spalte: 2},
	}

	if !reflect.DeepEqual(antwort.Symbole, erwartet) {
		t.Errorf("gefunden %+v, erwartet %+v", antwort.Symbole, erwartet)
	}

	if probe.paket != "/pakete/probe.mse-game" {
		t.Errorf("gefragt wurde nach %q", probe.paket)
	}
}

// **Die Verwendungen eines Namens kommen mit Datei, Zeile und Zeilentext.**
func TestDieVerwendungenKommenAnsFenster(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(werkstattMitTexten())

	antwort := anwendung.Symbolverwendungen(t.Context(), "/pakete/probe.mse-game", "title")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	erwartet := []app.SymbolverwendungDTO{{Datei: "game", Zeile: 4, Text: "x := card.title"}}

	if !reflect.DeepEqual(antwort.Verwendungen, erwartet) {
		t.Errorf("gefunden %+v, erwartet %+v", antwort.Verwendungen, erwartet)
	}
}

// **Leer heißt leer, nicht `null`** — über die Wails-Bindung stünde dort
// sonst `null`, und die Seite griffe ins Leere.
func TestOhneSymboleKommenLeereListen(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(&werkstattprobe{})

	if antwort := anwendung.Paketsymbole(t.Context(), "/p"); antwort.Symbole == nil || antwort.Fehler != nil {
		t.Errorf("Symbole: %+v", antwort)
	}

	if antwort := anwendung.Symbolverwendungen(t.Context(), "/p", "x"); antwort.Verwendungen == nil ||
		antwort.Fehler != nil {
		t.Errorf("Verwendungen: %+v", antwort)
	}
}

// **Ohne Werkstatt und bei einem Fehler sagen es beide.**
func TestDieSymbolwegeMeldenIhreFehler(t *testing.T) {
	ohne := app.NeueApp("")

	if antwort := ohne.Paketsymbole(t.Context(), "/p"); antwort.Fehler == nil {
		t.Error("ohne Werkstatt kamen Symbole")
	}

	if antwort := ohne.Symbolverwendungen(t.Context(), "/p", "x"); antwort.Fehler == nil {
		t.Error("ohne Werkstatt kamen Verwendungen")
	}

	kaputt := app.NeueApp("")
	kaputt.MitPaketdateien(&werkstattprobe{fehler: errors.New("die Platte ist weg")})

	if antwort := kaputt.Paketsymbole(t.Context(), "/p"); antwort.Fehler == nil {
		t.Error("der Fehler der Werkstatt kam nicht an")
	}

	if antwort := kaputt.Symbolverwendungen(t.Context(), "/p", "x"); antwort.Fehler == nil {
		t.Error("der Fehler der Werkstatt kam nicht an")
	}
}

// **Die Bindungen gehen denselben Weg.**
func TestDieSymbolbindungenGehenDenselbenWeg(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitPaketdateien(werkstattMitTexten())
	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.Paketsymbole(t.Context(), "/p"); len(antwort.Symbole) != 2 {
		t.Errorf("Symbole: %+v", antwort)
	}

	if antwort := bindungen.Symbolverwendungen(t.Context(), "/p", "title"); len(antwort.Verwendungen) != 1 {
		t.Errorf("Verwendungen: %+v", antwort)
	}
}
