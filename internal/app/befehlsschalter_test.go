package app_test

// **Die Kommandozentrale fragt die Schalter des Werkzeugs** — Befund 1 des
// Gesamtreviews über 0.6.0 (#1660).
//
// Die Befehle rufen das Werkzeug direkt; die Schalterprüfung für Werkzeuge
// sitzt im MCP-**Server**, und die Bindungswache sah nur den Kernweg
// `BefehlAusfuehren`. War `spielen` aus, legte die Palette trotzdem ein Deck
// an — die Kommandozentrale aus 5.1 und die Schalter aus 6.12 waren je für
// sich geprüft.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// deckbefehle sind ein Werkzeug aus `spielen` und eines aus dem Kern.
func deckbefehle(gerufen *bool) []app.Befehl {
	return []app.Befehl{
		{
			Name: "schreiben.deck.anlegen", Schreibend: true, Idempotent: true,
			Ruf: func(context.Context, json.RawMessage) (string, error) {
				*gerufen = true

				return "angelegt", nil
			},
		},
		{
			Name: "lesen.sets",
			Ruf:  func(context.Context, json.RawMessage) (string, error) { return "[]", nil },
		},
	}
}

func TestEinAbgeschalteterBefehlLaeuftNicht(t *testing.T) {
	t.Parallel()

	gerufen := false
	anwendung := app.NeueApp("").MitBefehlen(deckbefehle(&gerufen))
	anwendung.MitEinstellungen(&einstellungsprobe{werte: map[string]string{
		domain.SchalterEinstellung("spielen"): domain.SchalterAus,
	}})

	ergebnis := anwendung.BefehlAusfuehren(t.Context(), "schreiben.deck.anlegen", "{}", true)

	switch {
	case gerufen:
		t.Fatal("das Werkzeug lief, obwohl `spielen` aus ist")
	case ergebnis.Fehler == nil || ergebnis.Fehler.Code != app.CodeAbgeschaltet:
		t.Errorf("es kam %+v", ergebnis.Fehler)
	}

	// **Dass jemand zuhört** (§3): der Kern läuft.
	if kern := anwendung.BefehlAusfuehren(t.Context(), "lesen.sets", "{}", false); kern.Fehler != nil {
		t.Errorf("der Kernbefehl wurde abgewiesen: %+v", kern.Fehler)
	}
}

// **Und die Palette bietet ihn gar nicht erst an.**
func TestDerKatalogNenntKeinenAbgeschaltetenBefehl(t *testing.T) {
	t.Parallel()

	gerufen := false
	anwendung := app.NeueApp("").MitBefehlen(deckbefehle(&gerufen))
	anwendung.MitEinstellungen(&einstellungsprobe{werte: map[string]string{
		domain.SchalterEinstellung("spielen"): domain.SchalterAus,
	}})

	namen := map[string]bool{}
	for _, befehl := range anwendung.Befehlskatalog(t.Context()).Befehle {
		namen[befehl.Name] = true
	}

	if namen["schreiben.deck.anlegen"] || !namen["lesen.sets"] {
		t.Errorf("der Katalog nennt %v", namen)
	}
}

// **Ein Bestand, der nicht antwortet, leert die Palette nicht** — ein
// Schalter sichert nichts (Vermerk 0158), und vor dem Ruf fragt die Wache
// noch einmal.
func TestBeiEinemLesefehlerNenntDerKatalogAlles(t *testing.T) {
	t.Parallel()

	gerufen := false
	anwendung := app.NeueApp("").MitBefehlen(deckbefehle(&gerufen))
	anwendung.MitEinstellungen(&einstellungsprobe{fehler: errors.New("kein Bestand")})

	if befehle := anwendung.Befehlskatalog(t.Context()).Befehle; len(befehle) != 2 {
		t.Errorf("der Katalog nennt %d Befehle", len(befehle))
	}
}
