package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stilwaehler ist die Naht zu `service` — eine Attrappe.
type stilwaehler struct {
	angebote  []service.Stilangebot
	folgen    *service.Folgen
	optionen  []service.Stiloption
	felder    []domain.Field
	setfelder []domain.Field
	fehler    error

	gewechselt     string
	gesetzt        string
	zurueckgesetzt string
}

func (s *stilwaehler) Setfelder(context.Context, int64) ([]domain.Field, error) {
	return s.setfelder, s.fehler
}

func (s *stilwaehler) Stile(context.Context, int64) ([]service.Stilangebot, error) {
	return s.angebote, s.fehler
}

func (s *stilwaehler) Folgen(context.Context, int64, string) (*service.Folgen, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return s.folgen, nil
}

func (s *stilwaehler) Wechseln(_ context.Context, _ int64, paket string) error {
	s.gewechselt = paket

	return s.fehler
}

func (s *stilwaehler) Optionen(context.Context, int64) ([]service.Stiloption, error) {
	return s.optionen, s.fehler
}

func (s *stilwaehler) Kartenfelder(context.Context, int64) ([]domain.Field, error) {
	return s.felder, s.fehler
}

// **Ohne Spielskript.** Die Doppel prüfen die Naht, nicht die Feldskripte des
// Spiels; was mit Skript herauskommt, prüft `internal/service` (#1238).
func (s *stilwaehler) Spielskript(context.Context, int64) (string, error) {
	return "", nil
}

// Stilpaket kennt kein Paket — die Läufe darüber fragen nach Stilen, nicht
// nach der Ausfuhr.
func (s *stilwaehler) Stilpaket(context.Context, int64, string) (msepaket.Stil, error) {
	return msepaket.Stil{}, domain.ErrNotFound
}

func (s *stilwaehler) OptionSetzen(_ context.Context, _ int64, feld, wert string) error {
	s.gesetzt = feld + "=" + wert

	return s.fehler
}

func (s *stilwaehler) OptionZuruecksetzen(_ context.Context, _ int64, feld string) error {
	s.zurueckgesetzt = feld

	return s.fehler
}

// mitStilen baut eine Anwendung über einer Attrappe.
func mitStilen(waehler *stilwaehler) *app.App {
	anwendung := &app.App{}
	anwendung.MitStilwahl(waehler)

	return anwendung
}

// **Die Stile kommen an die Oberfläche.**
func TestStileGibtDieAngebote(t *testing.T) {
	anwendung := mitStilen(&stilwaehler{angebote: []service.Stilangebot{
		{
			Paket: "probe-schlicht", Anzeigename: "Schlicht", Spiel: "probe",
			Version: "2026-08-15", EigenesSpiel: true, Gewaehlt: true,
		},
	}})

	antwort := anwendung.Stile(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Stile) != 1 || !antwort.Stile[0].EigenesSpiel || !antwort.Stile[0].Gewaehlt {
		t.Errorf("Stile: %+v", antwort.Stile)
	}

	// **Der Anzeigename reist mit** (UX-05, S07.NAT-02, #1677).
	if antwort.Stile[0].Anzeigename != "Schlicht" {
		t.Errorf("Anzeigename: %q", antwort.Stile[0].Anzeigename)
	}
}

// **Die Folgen kommen als Angaben, nicht als Satz** (§9).
func TestStilfolgenGibtAngaben(t *testing.T) {
	anwendung := mitStilen(&stilwaehler{folgen: &service.Folgen{
		UngedeuteteFelder: []string{"flavor"},
		VerloreneOptionen: []string{"hintergrund"},
		Spielwechsel:      true,
	}})

	antwort := anwendung.Stilfolgen(t.Context(), "1", "fremd")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !antwort.Folgen.Spielwechsel || len(antwort.Folgen.UngedeuteteFelder) != 1 {
		t.Errorf("Folgen: %+v", antwort.Folgen)
	}
}

// **Wechseln und Setzen reichen durch.**
func TestWechselnUndSetzenReichenDurch(t *testing.T) {
	waehler := &stilwaehler{}
	anwendung := mitStilen(waehler)

	if antwort := anwendung.StilWechseln(t.Context(), "1", "probe-bild"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if waehler.gewechselt != "probe-bild" {
		t.Errorf("gewechselt: %q", waehler.gewechselt)
	}

	if antwort := anwendung.StiloptionSetzen(t.Context(), "1", "hintergrund", "dunkel"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if waehler.gesetzt != "hintergrund=dunkel" {
		t.Errorf("gesetzt: %q", waehler.gesetzt)
	}
}

// **Die Optionen kommen mit Art, Beschreibung und Wahlmöglichkeiten.**
//
// Und eine Option ohne Auswahl gibt eine **leere Liste**, nicht `null` — die
// Fehlerklasse zwischen den Prüfebenen (§3).
func TestStiloptionenGebenArtUndWahl(t *testing.T) {
	anwendung := mitStilen(&stilwaehler{optionen: []service.Stiloption{
		{
			StyleField: domain.StyleField{
				Name: "hintergrund", Kind: domain.KindChoice,
				Description: "Der Grund.", Choices: []string{"hell", "dunkel"},
			},
			Wert: "dunkel",
		},
		{
			StyleField: domain.StyleField{Name: "rahmen", Kind: domain.KindBoolean},
			Wert:       "no",
		},
	}})

	antwort := anwendung.Stiloptionen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Optionen) != 2 {
		t.Fatalf("Optionen: %+v", antwort.Optionen)
	}

	if antwort.Optionen[0].Art != "choice" || antwort.Optionen[0].Beschreibung != "Der Grund." {
		t.Errorf("erste Option: %+v", antwort.Optionen[0])
	}

	if antwort.Optionen[1].Moeglichkeiten == nil {
		t.Error("nil statt leerer Wahlliste")
	}
}

// **Eine krumme Kennung wird abgewiesen — auf jedem der fünf Wege.**
func TestEineKrummeKennungWirdAbgewiesen(t *testing.T) {
	anwendung := mitStilen(&stilwaehler{})

	if antwort := anwendung.Stile(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("Stile")
	}

	if antwort := anwendung.Stilfolgen(t.Context(), "keine Zahl", "x"); antwort.Fehler == nil {
		t.Error("Stilfolgen")
	}

	if antwort := anwendung.StilWechseln(t.Context(), "keine Zahl", "x"); antwort.Fehler == nil {
		t.Error("StilWechseln")
	}

	if antwort := anwendung.Stiloptionen(t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("Stiloptionen")
	}

	if antwort := anwendung.StiloptionSetzen(t.Context(), "keine Zahl", "a", "b"); antwort.Fehler == nil {
		t.Error("StiloptionSetzen")
	}
}

// **Ein Fehler aus dem Dienst wird übersetzt — und Listen bleiben Listen.**
func TestEinFehlerAusDerStilwahlWirdUebersetzt(t *testing.T) {
	anwendung := mitStilen(&stilwaehler{fehler: errors.New("der Bestand schweigt")})

	stile := anwendung.Stile(t.Context(), "1")
	if stile.Fehler == nil || stile.Stile == nil {
		t.Errorf("Stile: %+v", stile)
	}

	if antwort := anwendung.Stilfolgen(t.Context(), "1", "x"); antwort.Fehler == nil {
		t.Error("Stilfolgen")
	}

	optionen := anwendung.Stiloptionen(t.Context(), "1")
	if optionen.Fehler == nil || optionen.Optionen == nil {
		t.Errorf("Stiloptionen: %+v", optionen)
	}

	if antwort := anwendung.StilWechseln(t.Context(), "1", "x"); antwort.Fehler == nil {
		t.Error("StilWechseln")
	}

	if antwort := anwendung.StiloptionSetzen(t.Context(), "1", "a", "b"); antwort.Fehler == nil {
		t.Error("StiloptionSetzen")
	}
}

// **Die Bindungen reichen durch** — und zwar zur Anwendung, nicht ins Leere.
func TestDieStilbindungenReichenDurch(t *testing.T) {
	waehler := &stilwaehler{
		angebote: []service.Stilangebot{{Paket: "probe-schlicht"}},
		folgen:   &service.Folgen{},
		optionen: []service.Stiloption{{StyleField: domain.StyleField{Name: "hintergrund"}}},
	}
	bindungen := app.NeueBindungen(mitStilen(waehler))

	if antwort := bindungen.Stile(t.Context(), "1"); len(antwort.Stile) != 1 {
		t.Errorf("Stile: %+v", antwort)
	}

	if antwort := bindungen.Stilfolgen(t.Context(), "1", "x"); antwort.Folgen == nil {
		t.Errorf("Stilfolgen: %+v", antwort)
	}

	if antwort := bindungen.StilWechseln(t.Context(), "1", "probe-bild"); antwort.Fehler != nil {
		t.Errorf("StilWechseln: %+v", antwort)
	}

	if antwort := bindungen.Stiloptionen(t.Context(), "1"); len(antwort.Optionen) != 1 {
		t.Errorf("Stiloptionen: %+v", antwort)
	}

	if antwort := bindungen.StiloptionSetzen(t.Context(), "1", "hintergrund", "hell"); antwort.Fehler != nil {
		t.Errorf("StiloptionSetzen: %+v", antwort)
	}

	if antwort := bindungen.StiloptionZuruecksetzen(t.Context(), "1", "randfarbe"); antwort.Fehler != nil {
		t.Errorf("StiloptionZuruecksetzen: %+v", antwort)
	}

	if waehler.zurueckgesetzt != "randfarbe" {
		t.Errorf("zurückgesetzt wurde %q", waehler.zurueckgesetzt)
	}

	if waehler.gewechselt != "probe-bild" || waehler.gesetzt != "hintergrund=hell" {
		t.Errorf("die Bindungen kamen nicht an: %q, %q", waehler.gewechselt, waehler.gesetzt)
	}
}
