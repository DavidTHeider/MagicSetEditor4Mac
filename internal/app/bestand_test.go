package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// leseprobe ist ein Bestandsleser aus dem Arbeitsspeicher.
type leseprobe struct {
	koepfe  []domain.SetHead
	karten  []*domain.CardRow
	bloecke []domain.RawBlock
	fehler  error
}

func (l leseprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return l.koepfe, l.fehler
}

func (l leseprobe) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return l.karten, l.fehler
}

func (l leseprobe) Rohbloecke(context.Context, int64) ([]domain.RawBlock, error) {
	return l.bloecke, l.fehler
}

func mitLeser(probe leseprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitBestand(probe)

	return anwendung
}

func TestSetsKommenAlsDTO(t *testing.T) {
	anwendung := mitLeser(leseprobe{koepfe: []domain.SetHead{
		{ID: 42, Name: "Mein Set", Game: "magic", Style: "m15", Origin: "/a.mse-set", Cards: 7},
	}})

	antwort := anwendung.Sets(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	sets := antwort.Sets

	if len(sets) != 1 {
		t.Fatalf("%d Sets", len(sets))
	}

	// **Die Kennung ist eine Zeichenkette.** JavaScript kennt ganze Zahlen genau
	// nur bis 2^53; eine Kennung, die dort ankommt und nicht mehr dieselbe ist,
	// fällt erst auf, wenn jemand ein Set nicht mehr öffnen kann.
	if sets[0].ID != "42" || sets[0].Karten != 7 {
		t.Errorf("%+v", sets[0])
	}
}

// **Eine leere Antwort ist `[]`, nicht `null`.**
//
// Das ist die Lücke zwischen den Prüfebenen, vor der AGENTS.md §3 warnt: die
// Go-Tests sehen die Struct-Marken nicht, die Frontend-Tests gehen an der
// Serialisierung vorbei — beide sind grün, und die Seite bleibt leer. Dieser
// Test geht deshalb **durch die Serialisierung**.
func TestLeereAntwortenSindListenUndKeinNull(t *testing.T) {
	anwendung := mitLeser(leseprobe{})

	faelle := map[string]any{
		"Sets":       anwendung.Sets(t.Context()).Sets,
		"Karten":     anwendung.Karten(t.Context(), "1").Karten,
		"Vorbehalte": anwendung.Vorbehalte(t.Context(), "1").Vorbehalte,
	}

	for name, wert := range faelle {
		t.Run(name, func(t *testing.T) {
			gebaut, err := json.Marshal(wert)
			if err != nil {
				t.Fatalf("serialisieren: %v", err)
			}

			if string(gebaut) != "[]" {
				t.Errorf("%s — erwartet war []", gebaut)
			}
		})
	}
}

// Und eine Karte ohne Feldwerte trägt eine leere Liste, kein `null`.
func TestEineKarteOhneWerteTraegtEineLeereListe(t *testing.T) {
	anwendung := mitLeser(leseprobe{karten: []*domain.CardRow{{ID: 1, Number: 0}}})

	antwort := anwendung.Karten(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	gebaut, err := json.Marshal(antwort.Karten)
	if err != nil {
		t.Fatalf("serialisieren: %v", err)
	}

	if !strings.Contains(string(gebaut), `"werte":[]`) {
		t.Errorf("%s — erwartet war eine leere Liste", gebaut)
	}
}

func TestKartenUndVorbehalteKommenDurch(t *testing.T) {
	anwendung := mitLeser(leseprobe{
		karten: []*domain.CardRow{
			{ID: 5, Number: 0, Values: []domain.Value{{Field: "name", Text: "Feuerball"}}},
		},
		bloecke: []domain.RawBlock{{File: "set", Line: 12, Text: "unverstanden: wert"}},
	})

	karten := anwendung.Karten(t.Context(), "1")
	if karten.Fehler != nil || len(karten.Karten) != 1 || karten.Karten[0].ID != "5" {
		t.Fatalf("%+v", karten)
	}

	if len(karten.Karten[0].Werte) != 1 || karten.Karten[0].Werte[0].Feld != "name" {
		t.Errorf("%+v", karten.Karten[0].Werte)
	}

	vorbehalte := anwendung.Vorbehalte(t.Context(), "1")
	if vorbehalte.Fehler != nil || len(vorbehalte.Vorbehalte) != 1 ||
		vorbehalte.Vorbehalte[0].Zeile != 12 {
		t.Errorf("%+v", vorbehalte)
	}
}

// Eine unlesbare Kennung ist ein Validierungsfehler, kein „nicht gefunden".
//
// Wer `abc` schickt, hat nicht ein Set verfehlt, sondern gar keines genannt.
func TestEineUnlesbareKennungIstEinValidierungsfehler(t *testing.T) {
	anwendung := mitLeser(leseprobe{})

	for _, kennung := range []string{"abc", "", "1.5", "9999999999999999999999"} {
		t.Run(kennung, func(t *testing.T) {
			befund := anwendung.Karten(t.Context(), kennung).Fehler
			if befund == nil {
				t.Fatal("die Kennung kam durch")
			}

			if befund.Code != app.CodeValidierung || befund.Feld != "id" {
				t.Errorf("%+v", befund)
			}

			if anwendung.Vorbehalte(t.Context(), kennung).Fehler == nil {
				t.Error("Vorbehalte nahm die Kennung an")
			}
		})
	}
}

// Ein Fehler aus dem Speicher wird übersetzt, nicht durchgereicht.
//
// **Rohe Meldungen erreichen die Oberfläche nie** (§8).
func TestEinSpeicherfehlerWirdUebersetzt(t *testing.T) {
	roh := errors.New("SQL logic error: no such table: cards")
	anwendung := mitLeser(leseprobe{fehler: roh})

	for name, ruf := range map[string]func() *app.Fehler{
		"Sets":       func() *app.Fehler { return anwendung.Sets(t.Context()).Fehler },
		"Karten":     func() *app.Fehler { return anwendung.Karten(t.Context(), "1").Fehler },
		"Vorbehalte": func() *app.Fehler { return anwendung.Vorbehalte(t.Context(), "1").Fehler },
	} {
		t.Run(name, func(t *testing.T) {
			befund := ruf()
			if befund == nil {
				t.Fatal("der Fehler ging verloren")
			}

			if befund.Code != app.CodeUnbekannt {
				t.Errorf("Code %q", befund.Code)
			}

			if strings.Contains(befund.Meldung, "no such table") {
				t.Errorf("die rohe Meldung kam durch: %q", befund.Meldung)
			}
		})
	}
}
