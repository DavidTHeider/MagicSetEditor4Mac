package app_test

// **Die Symbolbibliothek für MCP und REST** — DOM-02
// (#1697).
// Ein Fehler ist dort ein `error` und kein Feld, und eine Datei ist ein Text.

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieWerkstattReichtDieBibliothekDurch(t *testing.T) {
	db, set, leer := bibliotheksbestand(t)
	ctx := probekontext(t)
	werkstatt := app.NeueSymbolbibliothekswerkstatt(db, domain.UrheberServer)

	aufgenommen, err := werkstatt.BibliothekssymbolAufnehmen(ctx, set, "Krone")
	if err != nil {
		t.Fatalf("aufnehmen: %v", err)
	}

	krone, gestalt := aufgenommen.(app.BibliothekssymbolAntwort)
	if !gestalt {
		t.Fatalf("die Gestalt: %T", aufgenommen)
	}

	id, _ := strconv.ParseInt(krone.ID, 10, 64) //nolint:errcheck // aus der Antwort

	liste, err := werkstatt.Symbolbibliothek(ctx, "")
	symbole, _ := liste.([]app.BibliothekssymbolDTO) //nolint:errcheck // geprüft über die Länge
	if err != nil || len(symbole) != 1 || symbole[0].Vorschau != "" || len(symbole[0].Verwendet) != 1 {
		t.Fatalf("die Liste: %+v, %v", liste, err)
	}

	text, err := werkstatt.BibliothekssymbolText(ctx, id)
	gelesen, gestalt := text.(app.BibliothekssymbolTextAntwort)
	inhalt := gelesen.Inhalt

	if err != nil || !gestalt || inhalt == "" {
		t.Fatalf("als Text: %+v, %v", text, err)
	}

	for name, ruf := range map[string]func() (any, error){
		"kopieren":   func() (any, error) { return werkstatt.BibliothekssymbolKopieren(ctx, id, "Zacke") },
		"umbenennen": func() (any, error) { return werkstatt.BibliothekssymbolUmbenennen(ctx, id, "Reif") },
		"einsetzen":  func() (any, error) { return werkstatt.BibliothekssymbolEinsetzen(ctx, id, leer) },
		"aus Text":   func() (any, error) { return werkstatt.BibliothekssymbolAusText(ctx, inhalt, "Zinne") },
	} {
		if ergebnis, err := ruf(); err != nil || ergebnis == nil {
			t.Errorf("%s: %+v, %v", name, ergebnis, err)
		}
	}

	// **Verwendet ohne Bestätigung ist ein Konflikt** — als `error`.
	if _, err := werkstatt.BibliothekssymbolLoeschen(ctx, id, false); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("ohne Bestätigung: %v", err)
	}

	if folgen, err := werkstatt.BibliothekssymbolLoeschen(ctx, id, true); err != nil || folgen == nil {
		t.Errorf("mit Bestätigung: %+v, %v", folgen, err)
	}
}

// **Ohne Zuhause sagt jeder Weg es**, statt irgendwo abzulegen.
func TestDieWerkstattBrauchtEinZuhause(t *testing.T) {
	t.Setenv("HOME", "")

	werkstatt := app.NeueSymbolbibliothekswerkstatt(nil, domain.UrheberServer)

	if _, err := werkstatt.Symbolbibliothek(t.Context(), ""); err == nil {
		t.Error("ohne Zuhause ging es durch")
	}
}

// **Die Bearbeitung über die Werkstatt** — die Teile kommen als rohes JSON,
// wie beim Symboleditor über MCP.
func TestDieWerkstattBearbeitetEinSymbol(t *testing.T) {
	db, set, _ := bibliotheksbestand(t)
	ctx := probekontext(t)
	werkstatt := app.NeueSymbolbibliothekswerkstatt(db, domain.UrheberServer)

	aufgenommen, err := werkstatt.BibliothekssymbolAufnehmen(ctx, set, "Krone")
	if err != nil {
		t.Fatalf("aufnehmen: %v", err)
	}

	krone, _ := aufgenommen.(app.BibliothekssymbolAntwort) //nolint:errcheck // geprüft über die Kennung
	id, _ := strconv.ParseInt(krone.ID, 10, 64)            //nolint:errcheck // dito

	if teile, err := werkstatt.BibliothekssymbolTeile(ctx, id); err != nil || teile == nil {
		t.Fatalf("die Teile: %+v, %v", teile, err)
	}

	kreis, _ := json.Marshal(app.SymbolGrundform(app.SymbolgrundformDTO{Art: "Kreis"}).Teile) //nolint:errcheck // Teile gehen immer

	// **Der Reihe nach** und nicht über eine Map: zurücknehmen setzt das
	// Sichern voraus.
	for _, schritt := range []struct {
		name string
		ruf  func() (any, error)
	}{
		{"sichern", func() (any, error) { return werkstatt.BibliothekssymbolSichern(ctx, id, kreis) }},
		{"nachziehen", func() (any, error) { return werkstatt.BibliothekssymbolNachziehen(ctx, id) }},
		{"zurücknehmen", func() (any, error) { return werkstatt.BibliothekssymbolZuruecknehmen(ctx, id) }},
	} {
		if ergebnis, err := schritt.ruf(); err != nil || ergebnis == nil {
			t.Errorf("%s: %+v, %v", schritt.name, ergebnis, err)
		}
	}

	if _, err := werkstatt.BibliothekssymbolSichern(ctx, id, json.RawMessage(`kein json`)); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unlesbare Teile: %v", err)
	}
}
