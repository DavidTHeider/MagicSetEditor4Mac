package service_test

// **Die Autoersetzungen eines Spiels sehen und pflegen** — Werkstück E von
// Phase 4.11
// (#1341).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `GameSettings::initDefaults` (`src/data/settings.cpp`, Zeilen 88–104,
// Fassung f9d9356 vom 2021-07-06). Verglichen wird über das **Muster**; eine
// Regel des Pakets ist **nicht** `custom`; was jemand angelegt hat, bleibt es.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ersetzablage ist ein Doppel für `service.Autoersetzungsablage`.
type ersetzablage struct {
	regeln []domain.Autoersetzungsregel
	an     bool
	fehler error

	// regelfehler scheitert **nur** am Lesen der Regeln — der Schalter steht
	// dann noch.
	regelfehler error

	gesetzt     domain.Autoersetzungsregel
	geloescht   string
	zurueck     bool
	geschaltet  *bool
	gesetztFuer string
}

func (e *ersetzablage) Regeln(_ context.Context, _ string) (
	[]domain.Autoersetzungsregel, error,
) {
	if e.regelfehler != nil {
		return nil, e.regelfehler
	}

	return e.regeln, e.fehler
}

func (e *ersetzablage) Setzen(
	_ context.Context, spiel string, regel domain.Autoersetzungsregel,
) error {
	e.gesetzt, e.gesetztFuer = regel, spiel

	return e.fehler
}

func (e *ersetzablage) Loeschen(_ context.Context, _, muster string) error {
	e.geloescht = muster

	return e.fehler
}

func (e *ersetzablage) Zuruecksetzen(_ context.Context, _ string) error {
	e.zurueck = true

	return e.fehler
}

func (e *ersetzablage) Schalter(_ context.Context, _ string) (bool, error) {
	return e.an, e.fehler
}

func (e *ersetzablage) SchalterSetzen(_ context.Context, _ string, an bool) error {
	e.geschaltet = &an

	return e.fehler
}

// spielprobeErsetzung gibt Spielnamen und Paketregeln.
type spielprobeErsetzung struct {
	spiel   string
	regeln  []msepaket.AutoErsetzung
	fehler  error
	paketno error
}

func (s *spielprobeErsetzung) SpielVon(_ context.Context, _ int64) (string, error) {
	return s.spiel, s.fehler
}

// Autoersetzungen erfüllt zusätzlich `service.Paketregelquelle`.
func (s *spielprobeErsetzung) Autoersetzungen(ctx context.Context, setID int64) (
	[]msepaket.AutoErsetzung, error,
) {
	return s.Paketregeln(ctx, setID)
}

func (s *spielprobeErsetzung) Paketregeln(_ context.Context, _ int64) (
	[]msepaket.AutoErsetzung, error,
) {
	return s.regeln, s.paketno
}

func probeersetzung() (*ersetzablage, *spielprobeErsetzung) {
	return &ersetzablage{an: true}, &spielprobeErsetzung{
		spiel: "magic",
		regeln: []msepaket.AutoErsetzung{
			{Muster: "(W)", Ersatz: "<sym>W</sym>", GanzesWort: true},
			{Muster: "(U)", Ersatz: "<sym>U</sym>", GanzesWort: true},
		},
	}
}

// **Ohne eigene Zeile stehen die Regeln des Pakets da — und keine ist eigen.**
func TestOhneEigeneZeileStehenDieRegelnDesPakets(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()

	liste, err := service.NeueAutoersetzungspflege(ablage, spiele).
		Liste(t.Context(), 1)
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if len(liste.Regeln) != 2 || !liste.An {
		t.Fatalf("die Liste lautet %+v", liste)
	}

	for _, regel := range liste.Regeln {
		if regel.Eigen {
			t.Errorf("die Regel %q gilt als eigen", regel.Muster)
		}

		// **Und sie ist an** — das Paket schaltet nur ab, wenn es `enabled:
		// false` schreibt, und `msepaket` nimmt solche gar nicht erst auf.
		if !regel.Aktiv {
			t.Errorf("die Regel %q gilt als abgeschaltet", regel.Muster)
		}
	}
}

// **Eine Zeile des Bestands überlagert die gleichnamige des Pakets — und
// bleibt nicht eigen.**
//
// Das ist MSEs `ar2->custom = false`: das Muster gehört dem Paket, auch wenn
// der Ersatz aus dem Bestand kommt.
func TestEineZeileUeberlagertDieGleichnamigeDesPakets(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()
	ablage.regeln = []domain.Autoersetzungsregel{
		{Muster: "(W)", Ersatz: "WEISS", GanzesWort: false, Aktiv: false, Eigen: true},
	}

	liste, err := service.NeueAutoersetzungspflege(ablage, spiele).
		Liste(t.Context(), 1)
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if len(liste.Regeln) != 2 {
		t.Fatalf("%d Regeln: %+v", len(liste.Regeln), liste.Regeln)
	}

	// **Geordnet nach dem Muster**: `(U)` vor `(W)`.
	weiss := liste.Regeln[1]

	if weiss.Muster != "(W)" || weiss.Ersatz != "WEISS" ||
		weiss.Aktiv || weiss.GanzesWort {
		t.Errorf("die überlagerte Regel lautet %+v", weiss)
	}

	if weiss.Eigen {
		t.Error("die überlagerte Regel gilt als eigen — ihr Muster gehört dem Paket")
	}
}

// **Eine Zeile ohne Gegenstück im Paket ist eigen.**
func TestEineZeileOhneGegenstueckIstEigen(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()
	ablage.regeln = []domain.Autoersetzungsregel{
		{Muster: "!!", Ersatz: "…", Aktiv: true},
	}

	liste, _ := service.NeueAutoersetzungspflege(ablage, spiele). //nolint:errcheck // unten geprüft
									Liste(t.Context(), 1)

	if len(liste.Regeln) != 3 {
		t.Fatalf("%d Regeln: %+v", len(liste.Regeln), liste.Regeln)
	}

	// **`!!` steht vorn**: geordnet wird nach dem Muster, und `!` kommt vor `(`.
	if liste.Regeln[0].Muster != "!!" || !liste.Regeln[0].Eigen {
		t.Errorf("die erste Regel lautet %+v", liste.Regeln[0])
	}
}

// **Ohne Spielpaket bleiben die eigenen Regeln stehen.**
//
// §8: ein Set ohne Spielpaket ist auf einem frischen Rechner der Normalfall,
// und wer dabei seine eigenen Regeln verlöre, verlöre sie ohne Grund.
func TestOhneSpielpaketBleibenDieEigenenRegeln(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()
	spiele.regeln, spiele.paketno = nil, errors.New("kein Spielpaket")
	ablage.regeln = []domain.Autoersetzungsregel{{Muster: "!!", Ersatz: "…", Aktiv: true}}

	liste, err := service.NeueAutoersetzungspflege(ablage, spiele).
		Liste(t.Context(), 1)
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if len(liste.Regeln) != 1 || !liste.Regeln[0].Eigen {
		t.Errorf("die Regeln lauten %+v", liste.Regeln)
	}
}

// **Der Dienst entscheidet, ob eine Regel eigen ist — nicht der Aufrufer.**
func TestDerDienstEntscheidetUeberEigen(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()

	pflege := service.NeueAutoersetzungspflege(ablage, spiele)

	// **Ein Muster des Pakets ist nicht eigen — auch wenn es so behauptet
	// wird.**
	if err := pflege.Setzen(t.Context(), 1, domain.Autoersetzungsregel{
		Muster: "(W)", Ersatz: "x", Eigen: true,
	}); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	if ablage.gesetzt.Eigen {
		t.Errorf("die gesetzte Regel lautet %+v", ablage.gesetzt)
	}

	if ablage.gesetztFuer != "magic" {
		t.Errorf("gesetzt wurde für %q", ablage.gesetztFuer)
	}

	// **Und ein fremdes Muster ist eigen — auch wenn es bestritten wird.**
	if err := pflege.Setzen(t.Context(), 1, domain.Autoersetzungsregel{
		Muster: "!!", Ersatz: "x", Eigen: false,
	}); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	if !ablage.gesetzt.Eigen {
		t.Errorf("die gesetzte Regel lautet %+v", ablage.gesetzt)
	}
}

// **Eine Regel des Pakets lässt sich nicht löschen.**
//
// MSEs `remove->Enable(ar && ar->custom)`. Sie gehört dem Paket; was hier
// ginge, wäre sie abzuschalten.
func TestEineRegelDesPaketsLaesstSichNichtLoeschen(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()

	pflege := service.NeueAutoersetzungspflege(ablage, spiele)

	if err := pflege.Loeschen(t.Context(), 1, "(W)"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("der Fehler lautet %v", err)
	}

	if ablage.geloescht != "" {
		t.Errorf("gelöscht wurde %q", ablage.geloescht)
	}

	// **Eine eigene schon.**
	if err := pflege.Loeschen(t.Context(), 1, "!!"); err != nil {
		t.Fatalf("Loeschen: %v", err)
	}

	if ablage.geloescht != "!!" {
		t.Errorf("gelöscht wurde %q", ablage.geloescht)
	}
}

// **Zurücksetzen und der Schalter kommen durch.**
func TestZuruecksetzenUndSchalterKommenDurch(t *testing.T) {
	t.Parallel()

	ablage, spiele := probeersetzung()

	pflege := service.NeueAutoersetzungspflege(ablage, spiele)

	if err := pflege.Zuruecksetzen(t.Context(), 1); err != nil || !ablage.zurueck {
		t.Errorf("Zuruecksetzen: %v / %v", err, ablage.zurueck)
	}

	if err := pflege.SchalterSetzen(t.Context(), 1, false); err != nil {
		t.Fatalf("SchalterSetzen: %v", err)
	}

	if ablage.geschaltet == nil || *ablage.geschaltet {
		t.Errorf("geschaltet wurde %v", ablage.geschaltet)
	}
}

// **Ein Set, dessen Spiel niemand kennt, hält jeden Weg an.**
func TestOhneSpielHaeltJederWegAn(t *testing.T) {
	t.Parallel()

	eigener := errors.New("das Set gibt es nicht")

	ablage, spiele := probeersetzung()
	spiele.fehler = eigener

	pflege := service.NeueAutoersetzungspflege(ablage, spiele)

	if _, err := pflege.Liste(t.Context(), 1); !errors.Is(err, eigener) {
		t.Errorf("Liste gab %v", err)
	}

	if err := pflege.Setzen(t.Context(), 1, domain.Autoersetzungsregel{
		Muster: "x",
	}); !errors.Is(err, eigener) {
		t.Errorf("Setzen gab %v", err)
	}

	if err := pflege.Loeschen(t.Context(), 1, "x"); !errors.Is(err, eigener) {
		t.Errorf("Loeschen gab %v", err)
	}

	if err := pflege.Zuruecksetzen(t.Context(), 1); !errors.Is(err, eigener) {
		t.Errorf("Zuruecksetzen gab %v", err)
	}

	if err := pflege.SchalterSetzen(t.Context(), 1, true); !errors.Is(err, eigener) {
		t.Errorf("SchalterSetzen gab %v", err)
	}
}

// **Und ein Lesefehler des Bestands ebenso.**
func TestEinLesefehlerHaeltDieAutoersetzungslisteAn(t *testing.T) {
	t.Parallel()

	eigener := errors.New("der Bestand antwortet nicht")

	ablage, spiele := probeersetzung()
	ablage.fehler = eigener

	if _, err := service.NeueAutoersetzungspflege(ablage, spiele).
		Liste(t.Context(), 1); !errors.Is(err, eigener) {
		t.Errorf("Liste gab %v", err)
	}
}

// **Die Spielregeln finden das Spiel eines Sets — und ein fehlendes ist ein
// Befund.**
func TestDieSpielregelnFindenDasSpielEinesSets(t *testing.T) {
	t.Parallel()

	regeln := service.NeueSpielregeln(&koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Game: "magic"}, {ID: 8, Game: "probe"},
	}}, &spielprobeErsetzung{regeln: []msepaket.AutoErsetzung{{Muster: "(W)"}}})

	spiel, err := regeln.SpielVon(t.Context(), 8)
	if err != nil || spiel != "probe" {
		t.Errorf("gefunden wurde %q (%v)", spiel, err)
	}

	if _, err := regeln.SpielVon(t.Context(), 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("der Fehler lautet %v", err)
	}

	// **Und die Paketregeln kommen durch.**
	ausDemPaket, paketfehler := regeln.Paketregeln(t.Context(), 7)
	if paketfehler != nil || len(ausDemPaket) != 1 {
		t.Errorf("die Paketregeln lauten %+v (%v)", ausDemPaket, paketfehler)
	}
}

// **Ein Lesefehler an den Setköpfen kommt durch.**
func TestEinLesefehlerAnDenSetkoepfenKommtDurchDieSpielregeln(t *testing.T) {
	t.Parallel()

	eigener := errors.New("der Bestand antwortet nicht")

	regeln := service.NeueSpielregeln(
		&koepfequelle{fehler: eigener}, &spielprobeErsetzung{})

	if _, err := regeln.SpielVon(t.Context(), 7); !errors.Is(err, eigener) {
		t.Errorf("der Fehler lautet %v", err)
	}
}
