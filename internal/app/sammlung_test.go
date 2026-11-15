// **Die Sammlung am Fenster** — Werkstück 2 von Phase 3.4
// (#1165).
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sammlungsnahtdoppel beantwortet die fünf Wege.
type sammlungsnahtdoppel struct {
	exemplare []*domain.Exemplar
	fehler    error

	// gesehen hält fest, was zuletzt hereinkam.
	gesehen domain.Exemplar
}

func (s *sammlungsnahtdoppel) Liste(context.Context) ([]*domain.Exemplar, error) {
	return s.exemplare, s.fehler
}

func (s *sammlungsnahtdoppel) Fuer(context.Context, string) ([]*domain.Exemplar, error) {
	return s.exemplare, s.fehler
}

func (s *sammlungsnahtdoppel) Anlegen(_ context.Context, exemplar domain.Exemplar) (int64, error) {
	s.gesehen = exemplar

	return 7, s.fehler
}

func (s *sammlungsnahtdoppel) Aendern(_ context.Context, exemplar domain.Exemplar) error {
	s.gesehen = exemplar

	return s.fehler
}

func (s *sammlungsnahtdoppel) Loeschen(context.Context, int64, int64) error { return s.fehler }

// **Ein Exemplar kommt als DTO an — samt seinem Preis als Text.**
//
// # Warum der Preis als Zeichenkette geht
//
// Im Bestand steht er in Hundertsteln; JSON kennt nur `number`, und das ist in
// JavaScript ein `double`. Was hier geht, ist die Form, in der jemand ihn
// tippt.
func TestEinExemplarKommtAlsDTOAn(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsnahtdoppel{exemplare: []*domain.Exemplar{{
		ID: 1, Printing: "abc", Menge: 3,
		Zustand: domain.ZustandNearMint, Sprache: "en",
		Erwerbspreis: 417, Erwerbswaehrung: "EUR",
	}}}

	anwendung := app.NeueApp("")
	anwendung.MitSammlung(doppel)

	antwort := anwendung.Sammlung(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Sammlung: %+v", antwort.Fehler)
	}

	if len(antwort.Exemplare) != 1 {
		t.Fatalf("es sind %d Exemplare", len(antwort.Exemplare))
	}

	if antwort.Exemplare[0].Erwerbspreis != "4.17" {
		t.Errorf("der Preis ist %q", antwort.Exemplare[0].Erwerbspreis)
	}
}

// **Ohne Preis steht dort nichts — und nicht `0.00`.**
//
// „Kein Preis erfasst" und „hat nichts gekostet" sind zwei Aussagen; ein
// Formularfeld, das von allein `0.00` zeigt, lädt zum Falschlesen ein.
func TestOhneErwerbspreisStehtNichts(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsnahtdoppel{exemplare: []*domain.Exemplar{{
		ID: 1, Printing: "abc", Menge: 1,
	}}}

	anwendung := app.NeueApp("")
	anwendung.MitSammlung(doppel)

	if preis := anwendung.Sammlung(t.Context()).Exemplare[0].Erwerbspreis; preis != "" {
		t.Errorf("der Preis ist %q", preis)
	}
}

// **Der getippte Preis wird umgerechnet, bevor er hineingeht.**
func TestDerGetipptePreisWirdUmgerechnet(t *testing.T) {
	t.Parallel()

	doppel := &sammlungsnahtdoppel{}

	anwendung := app.NeueApp("")
	anwendung.MitSammlung(doppel)

	antwort := anwendung.ExemplarAnlegen(t.Context(), app.ExemplarDTO{
		Printing: "abc", Menge: 1, Erwerbspreis: "4.17", Erwerbswaehrung: "EUR",
	})
	if antwort.Fehler != nil {
		t.Fatalf("ExemplarAnlegen: %+v", antwort.Fehler)
	}

	if doppel.gesehen.Erwerbspreis != 417 {
		t.Errorf("hineingegangen ist %d", doppel.gesehen.Erwerbspreis)
	}

	if antwort.ID != 7 {
		t.Errorf("die Kennung ist %d", antwort.ID)
	}
}

// **Ein unlesbarer Preis ist ein Befund am Feld** (§6).
//
// Wer sich vertippt hat, soll es am Feld sehen — und nicht später an einer
// Summe, die um vier Euro danebenliegt.
func TestEinUnlesbarerPreisHaengtAmFeld(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSammlung(&sammlungsnahtdoppel{})

	for name, ruf := range map[string]func() *app.Fehler{
		"Anlegen": func() *app.Fehler {
			return anwendung.ExemplarAnlegen(t.Context(), app.ExemplarDTO{
				Printing: "abc", Menge: 1, Erwerbspreis: "4,17",
			}).Fehler
		},
		"Aendern": func() *app.Fehler {
			return anwendung.ExemplarAendern(t.Context(), app.ExemplarDTO{
				ID: 1, Printing: "abc", Menge: 1, Erwerbspreis: "4,17",
			})
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := ruf()
			if fehler == nil {
				t.Fatal("der Preis kam durch")
			}

			if fehler.Feld != "erwerbspreis" {
				t.Errorf("der Befund hängt an %q", fehler.Feld)
			}
		})
	}
}

// **Ohne Naht sagen alle fünf Wege es — und geben leere Listen** (§3).
func TestOhneSammlungsnahtSagenAlleWegeEs(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	liste := anwendung.Sammlung(t.Context())
	if liste.Fehler == nil || liste.Exemplare == nil {
		t.Errorf("Sammlung: %+v", liste)
	}

	fuer := anwendung.SammlungFuer(t.Context(), "abc")
	if fuer.Fehler == nil || fuer.Exemplare == nil {
		t.Errorf("SammlungFuer: %+v", fuer)
	}

	if anwendung.ExemplarAnlegen(t.Context(), app.ExemplarDTO{}).Fehler == nil {
		t.Error("ExemplarAnlegen kam durch")
	}

	if anwendung.ExemplarAendern(t.Context(), app.ExemplarDTO{}) == nil {
		t.Error("ExemplarAendern kam durch")
	}

	if anwendung.ExemplarLoeschen(t.Context(), 1, 1) == nil {
		t.Error("ExemplarLoeschen kam durch")
	}
}

// **Ein Fehler der Naht kommt übersetzt an.**
func TestEinFehlerDerSammlungsnahtKommtAn(t *testing.T) {
	t.Parallel()

	// **Je Fall eine eigene Anwendung und ein eigenes Doppel.**
	//
	// Sie liefen nebenläufig über **ein** Doppel, und zwei davon schreiben
	// `gesehen` — `-race` hat es angestrichen. Ein geteiltes Doppel unter
	// `t.Parallel()` ist kein Fehler dieses Laufs, sondern eine Falle für
	// jeden nächsten: wer dem Doppel etwas Merkendes hinzufügt, macht damit
	// stillschweigend jeden Fall unsicher.
	stolpert := func() *app.App {
		anwendung := app.NeueApp("")
		anwendung.MitSammlung(&sammlungsnahtdoppel{
			fehler: errors.New("die Naht stolpert"),
		})

		return anwendung
	}

	for name, ruf := range map[string]func(anwendung *app.App) *app.Fehler{
		"Sammlung": func(anwendung *app.App) *app.Fehler {
			return anwendung.Sammlung(t.Context()).Fehler
		},
		"SammlungFuer": func(anwendung *app.App) *app.Fehler {
			return anwendung.SammlungFuer(t.Context(), "abc").Fehler
		},
		"ExemplarAnlegen": func(anwendung *app.App) *app.Fehler {
			return anwendung.ExemplarAnlegen(t.Context(),
				app.ExemplarDTO{Printing: "abc", Menge: 1}).Fehler
		},
		"ExemplarAendern": func(anwendung *app.App) *app.Fehler {
			return anwendung.ExemplarAendern(t.Context(),
				app.ExemplarDTO{ID: 1, Printing: "abc", Menge: 1})
		},
		"ExemplarLoeschen": func(anwendung *app.App) *app.Fehler {
			return anwendung.ExemplarLoeschen(t.Context(), 1, 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if ruf(stolpert()) == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Eine leere Sammlung gibt eine leere Liste und kein `nil`** (§3).
func TestEineLeereSammlungGibtEineLeereListeAmFenster(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSammlung(&sammlungsnahtdoppel{})

	if antwort := anwendung.Sammlung(t.Context()); antwort.Exemplare == nil {
		t.Error("die Liste ist nil")
	}
}
