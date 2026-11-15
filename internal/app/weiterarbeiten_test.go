package app_test

// **Weiterarbeiten am Fenster** — Werkstück U11 von Phase 5.1 (#1422).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type weiterarbeitsprobe struct {
	uebersicht domain.Weiterarbeiten
	fehler     error
	vermerkt   []domain.Arbeitsschritt
	angeheftet []domain.Arbeitsobjekt
	geloest    []string
}

func (w *weiterarbeitsprobe) Vermerken(_ context.Context, objekt domain.Arbeitsobjekt, vorgang domain.Arbeitsvorgang) error {
	w.vermerkt = append(w.vermerkt, domain.Arbeitsschritt{Objekt: objekt, Vorgang: vorgang})

	return w.fehler
}

func (w *weiterarbeitsprobe) Anheften(_ context.Context, objekt domain.Arbeitsobjekt) error {
	w.angeheftet = append(w.angeheftet, objekt)

	return w.fehler
}

func (w *weiterarbeitsprobe) Loesen(_ context.Context, art domain.Arbeitsart, kennung string) (bool, error) {
	w.geloest = append(w.geloest, string(art)+":"+kennung)

	return true, w.fehler
}

func (w *weiterarbeitsprobe) Uebersicht(context.Context) (domain.Weiterarbeiten, error) {
	return w.uebersicht, w.fehler
}

// koepfeprobe ist ein Bestand mit zwei Sets.
type koepfeprobe struct{ fehler error }

func (k koepfeprobe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 7, Name: "Probeset"}, {ID: 8, Name: ""}}, k.fehler
}

func (koepfeprobe) Karten(context.Context, int64) ([]*domain.CardRow, error) { return nil, nil }

func (koepfeprobe) Rohbloecke(context.Context, int64) ([]domain.RawBlock, error) { return nil, nil }

type benutzungsspion struct{ benutzt []int64 }

func (b *benutzungsspion) SetBenutzt(_ context.Context, setID int64) error {
	b.benutzt = append(b.benutzt, setID)

	return nil
}

func TestDieUebersichtKommtAlsUmschlag(t *testing.T) {
	t.Parallel()

	wann := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.FixedZone("Berlin", 2*3600))
	schritt := domain.Arbeitsschritt{
		Objekt:  domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "7", Name: "Probeset"},
		Vorgang: domain.VorgangGeoeffnet, Zeitpunkt: wann,
	}
	probe := &weiterarbeitsprobe{uebersicht: domain.Weiterarbeiten{
		Zuletzt: &schritt, Geoeffnet: []domain.Arbeitsschritt{schritt},
		Eingefuehrt: []domain.Arbeitsschritt{}, Ausgegeben: []domain.Arbeitsschritt{},
		Angeheftet: []domain.Arbeitsobjekt{{Art: domain.ArbeitsartSet, Kennung: "7", Name: "Probeset"}},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitWeiterarbeiten(probe)

	antwort := anwendung.Weiterarbeiten(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Zuletzt == nil || antwort.Zuletzt.Objekt.Name != "Probeset" || !antwort.Zuletzt.Objekt.Angeheftet {
		t.Errorf("Zuletzt ist %+v", antwort.Zuletzt)
	}

	// **In UTC und RFC 3339**, wie die Preisstände.
	if antwort.Zuletzt.Zeitpunkt != "2026-09-09T10:00:00Z" {
		t.Errorf("der Zeitpunkt ist %q", antwort.Zuletzt.Zeitpunkt)
	}

	if len(antwort.Geoeffnet) != 1 || len(antwort.Eingefuehrt) != 0 || antwort.Eingefuehrt == nil ||
		len(antwort.Angeheftet) != 1 || !antwort.Angeheftet[0].Angeheftet {
		t.Errorf("Übersicht: %+v", antwort)
	}

	probe.fehler = errors.New("die Ablage antwortet nicht")

	if fehlend := anwendung.Weiterarbeiten(t.Context()); fehlend.Fehler == nil || fehlend.Geoeffnet == nil {
		t.Errorf("bei kaputter Ablage: %+v", fehlend)
	}
}

func TestOhneNahtMeldenDieBindungenEinenFehler(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	if antwort := anwendung.Weiterarbeiten(t.Context()); antwort.Fehler == nil || antwort.Angeheftet == nil {
		t.Errorf("Weiterarbeiten ohne Naht: %+v", antwort)
	}

	if antwort := anwendung.Anheften(t.Context(), "set", "7"); antwort.Fehler == nil {
		t.Error("Anheften ohne Naht ging durch")
	}

	if antwort := anwendung.Loesen(t.Context(), "set", "7"); antwort.Fehler == nil {
		t.Error("Lösen ohne Naht ging durch")
	}

	// **DeckBenutzt ist Bestenfalls-Arbeit** — ohne Naht kein Fehler.
	if antwort := anwendung.DeckBenutzt(t.Context(), "3"); antwort.Fehler != nil {
		t.Errorf("DeckBenutzt ohne Naht: %+v", antwort.Fehler)
	}

	if antwort := anwendung.DeckBenutzt(t.Context(), "drei"); antwort.Fehler == nil {
		t.Error("eine unlesbare Deckkennung ging durch")
	}
}

func TestAnheftenPrueftUndHoltDenNamen(t *testing.T) {
	t.Parallel()

	probe := &weiterarbeitsprobe{}
	anwendung := app.NeueApp("")
	anwendung.MitWeiterarbeiten(probe)
	anwendung.MitBestand(koepfeprobe{})

	if antwort := anwendung.Anheften(t.Context(), "set", "7"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(probe.angeheftet) != 1 || probe.angeheftet[0].Name != "Probeset" {
		t.Errorf("angeheftet wurde %+v — erwartet mit Namen", probe.angeheftet)
	}

	// **Ohne Namen bleibt die Kennung** — Set 8 hat keinen, Set 9 gibt es nicht.
	for _, kennung := range []string{"8", "9"} {
		if antwort := anwendung.Anheften(t.Context(), "set", kennung); antwort.Fehler != nil {
			t.Fatalf("%s: %+v", kennung, antwort.Fehler)
		}
	}

	if probe.angeheftet[1].Name != "8" || probe.angeheftet[2].Name != "9" {
		t.Errorf("die Namen sind %q und %q", probe.angeheftet[1].Name, probe.angeheftet[2].Name)
	}

	if antwort := anwendung.Anheften(t.Context(), "karte", "1"); antwort.Fehler == nil ||
		antwort.Fehler.Code != "validierung" {
		t.Errorf("eine erfundene Art gab %+v", antwort.Fehler)
	}

	probe.fehler = errors.New("kaputt")

	if antwort := anwendung.Anheften(t.Context(), "set", "7"); antwort.Fehler == nil {
		t.Error("ein Fehler der Naht kam nicht an")
	}

	if antwort := anwendung.Loesen(t.Context(), "set", "7"); antwort.Fehler == nil {
		t.Error("ein Fehler beim Lösen kam nicht an")
	}

	probe.fehler = nil

	if antwort := anwendung.Loesen(t.Context(), "set", "7"); antwort.Fehler != nil || probe.geloest[len(probe.geloest)-1] != "set:7" {
		t.Errorf("Lösen: %+v, %v", antwort.Fehler, probe.geloest)
	}
}

func TestSetUndDeckBenutztVermerken(t *testing.T) {
	t.Parallel()

	probe := &weiterarbeitsprobe{}
	benutzung := &benutzungsspion{}
	anwendung := app.NeueApp("")
	anwendung.MitWeiterarbeiten(probe)
	anwendung.MitBestand(koepfeprobe{})
	anwendung.MitSetbenutzung(benutzung)
	anwendung.MitDecknaht(&deckprobe{deck: domain.Deck{Name: "Mono Rot"}})

	if antwort := anwendung.SetBenutzt(t.Context(), "7"); antwort.Fehler != nil {
		t.Fatalf("SetBenutzt: %+v", antwort.Fehler)
	}

	if antwort := anwendung.DeckBenutzt(t.Context(), "3"); antwort.Fehler != nil {
		t.Fatalf("DeckBenutzt: %+v", antwort.Fehler)
	}

	if len(probe.vermerkt) != 2 {
		t.Fatalf("vermerkt wurde %+v", probe.vermerkt)
	}

	if probe.vermerkt[0].Objekt.Name != "Probeset" || probe.vermerkt[0].Vorgang != domain.VorgangGeoeffnet ||
		probe.vermerkt[1].Objekt.Art != domain.ArbeitsartDeck || probe.vermerkt[1].Objekt.Name != "Mono Rot" {
		t.Errorf("vermerkt wurde %+v", probe.vermerkt)
	}

	// **Ein Fehler beim Vermerken hält den Weg nicht an.**
	probe.fehler = errors.New("die Ablage antwortet nicht")

	if antwort := anwendung.SetBenutzt(t.Context(), "7"); antwort.Fehler != nil {
		t.Errorf("SetBenutzt scheiterte am Vermerk: %+v", antwort.Fehler)
	}

	if len(benutzung.benutzt) != 2 {
		t.Errorf("benutzt wurde %v", benutzung.benutzt)
	}
}

// **Die Fensterbindungen reichen mit dem Kontext der Anwendung durch.**
func TestDieFensterbindungenDerUebersichtReichenDurch(t *testing.T) {
	t.Parallel()

	probe := &weiterarbeitsprobe{}
	anwendung := app.NeueApp("")
	anwendung.MitWeiterarbeiten(probe)
	anwendung.MitBestand(koepfeprobe{})
	anwendung.MitDecknaht(&deckprobe{deck: domain.Deck{Name: "Mono Rot"}})
	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.Weiterarbeiten(t.Context()); antwort.Fehler != nil {
		t.Errorf("Weiterarbeiten: %+v", antwort.Fehler)
	}

	if antwort := bindungen.Anheften(t.Context(), "set", "7"); antwort.Fehler != nil {
		t.Errorf("Anheften: %+v", antwort.Fehler)
	}

	if antwort := bindungen.Loesen(t.Context(), "set", "7"); antwort.Fehler != nil {
		t.Errorf("Loesen: %+v", antwort.Fehler)
	}

	if antwort := bindungen.DeckBenutzt(t.Context(), "3"); antwort.Fehler != nil {
		t.Errorf("DeckBenutzt: %+v", antwort.Fehler)
	}

	if len(probe.angeheftet) != 1 || len(probe.geloest) != 1 || len(probe.vermerkt) != 1 {
		t.Errorf("die Naht sah %d/%d/%d Rufe", len(probe.angeheftet), len(probe.geloest), len(probe.vermerkt))
	}
}
