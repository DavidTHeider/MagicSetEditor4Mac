package service_test

// **Weiterarbeiten** — Werkstück U11 von Phase 5.1 (#1422): die Übersicht
// führt nur, was es noch gibt, und nichts aus einem fremden Profil — das
// schneidet die Ablage; hier wird geprüft, dass Verschwundenes fällt und
// vergessen wird, und dass ein Vermerk seine Uhr trägt.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type verlaufsablage struct {
	schritte   []domain.Arbeitsschritt
	angeheftet []domain.Arbeitsobjekt
	vergessen  []string
	fehler     error
}

func (v *verlaufsablage) Vermerken(_ context.Context, schritt domain.Arbeitsschritt) error {
	if v.fehler != nil {
		return v.fehler
	}

	v.schritte = append(v.schritte, schritt)

	return nil
}

func (v *verlaufsablage) Schritte(context.Context) ([]domain.Arbeitsschritt, error) {
	return v.schritte, v.fehler
}

func (v *verlaufsablage) Anheften(_ context.Context, objekt domain.Arbeitsobjekt) error {
	if v.fehler != nil {
		return v.fehler
	}

	v.angeheftet = append(v.angeheftet, objekt)

	return nil
}

func (v *verlaufsablage) Loesen(_ context.Context, art domain.Arbeitsart, kennung string) (bool, error) {
	if v.fehler != nil {
		return false, v.fehler
	}

	for i, eines := range v.angeheftet {
		if eines.Art == art && eines.Kennung == kennung {
			v.angeheftet = append(v.angeheftet[:i], v.angeheftet[i+1:]...)

			return true, nil
		}
	}

	return false, nil
}

func (v *verlaufsablage) Angeheftete(context.Context) ([]domain.Arbeitsobjekt, error) {
	return v.angeheftet, v.fehler
}

func (v *verlaufsablage) Vergessen(_ context.Context, art domain.Arbeitsart, kennung string) error {
	v.vergessen = append(v.vergessen, string(art)+":"+kennung)

	return v.fehler
}

// bestandsprobe sagt, welche Objekte es gibt.
type bestandsprobe struct {
	gibt   map[string]bool
	fehler error
}

func (b bestandsprobe) Gibt(_ context.Context, objekt domain.Arbeitsobjekt) (bool, error) {
	return b.gibt[string(objekt.Art)+":"+objekt.Kennung], b.fehler
}

func festeUhr() time.Time { return time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC) }

func TestDieUebersichtFuehrtNurWasEsNochGibt(t *testing.T) {
	t.Parallel()

	ablage := &verlaufsablage{
		schritte: []domain.Arbeitsschritt{
			{Objekt: domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "1", Name: "Bleibt"}, Vorgang: domain.VorgangGeoeffnet, Zeitpunkt: festeUhr()},
			{Objekt: domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "2", Name: "Gelöscht"}, Vorgang: domain.VorgangGeoeffnet, Zeitpunkt: festeUhr().Add(time.Hour)},
			{Objekt: domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "3", Name: "Deck"}, Vorgang: domain.VorgangAusgegeben, Zeitpunkt: festeUhr().Add(-time.Hour)},
		},
		angeheftet: []domain.Arbeitsobjekt{
			{Art: domain.ArbeitsartSet, Kennung: "2", Name: "Gelöscht"},
			{Art: domain.ArbeitsartDeck, Kennung: "3", Name: "Deck"},
		},
	}
	bestand := bestandsprobe{gibt: map[string]bool{"set:1": true, "deck:3": true}}

	dienst := service.NeuesWeiterarbeiten(ablage, bestand, festeUhr)

	uebersicht, err := dienst.Uebersicht(context.Background())
	if err != nil {
		t.Fatalf("Übersicht: %v", err)
	}

	// **Das Gelöschte war das jüngste** — und ist trotzdem nicht »Zuletzt«.
	if uebersicht.Zuletzt == nil || uebersicht.Zuletzt.Objekt.Kennung != "1" {
		t.Errorf("Zuletzt ist %+v — erwartet Set 1", uebersicht.Zuletzt)
	}

	if len(uebersicht.Geoeffnet) != 1 || len(uebersicht.Ausgegeben) != 1 || len(uebersicht.Angeheftet) != 1 {
		t.Errorf("Übersicht: %+v", uebersicht)
	}

	// **Und vergessen wurde es** — einmal, nicht je Zeile.
	if len(ablage.vergessen) != 1 || ablage.vergessen[0] != "set:2" {
		t.Errorf("vergessen wurde %v", ablage.vergessen)
	}
}

func TestEinVermerkTraegtDieUhr(t *testing.T) {
	t.Parallel()

	ablage := &verlaufsablage{}
	dienst := service.NeuesWeiterarbeiten(ablage, bestandsprobe{}, festeUhr)

	objekt := domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "1", Name: "Probe"}
	if err := dienst.Vermerken(context.Background(), objekt, domain.VorgangEingefuehrt); err != nil {
		t.Fatalf("vermerken: %v", err)
	}

	if len(ablage.schritte) != 1 || !ablage.schritte[0].Zeitpunkt.Equal(festeUhr()) ||
		ablage.schritte[0].Vorgang != domain.VorgangEingefuehrt {
		t.Errorf("vermerkt wurde %+v", ablage.schritte)
	}

	// **Ein ungültiges Objekt kommt nicht an die Ablage.**
	if err := dienst.Vermerken(context.Background(), domain.Arbeitsobjekt{}, domain.VorgangGeoeffnet); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ein leeres Objekt gab %v", err)
	}

	if err := dienst.Anheften(context.Background(), domain.Arbeitsobjekt{Art: "x", Kennung: "1"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("eine erfundene Art gab %v", err)
	}
}

func TestAnheftenBrauchtDasObjektUndLoesenSagtObEsDaWar(t *testing.T) {
	t.Parallel()

	ablage := &verlaufsablage{}
	bestand := bestandsprobe{gibt: map[string]bool{"deck:3": true}}
	dienst := service.NeuesWeiterarbeiten(ablage, bestand, festeUhr)

	// **Was es nicht gibt, lässt sich nicht anheften.**
	if err := dienst.Anheften(context.Background(), domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "9"}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Objekt gab %v", err)
	}

	if err := dienst.Anheften(context.Background(), domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "3", Name: "Deck"}); err != nil {
		t.Fatalf("anheften: %v", err)
	}

	war, err := dienst.Loesen(context.Background(), domain.ArbeitsartDeck, "3")
	if err != nil || !war {
		t.Errorf("lösen: %v, %v", war, err)
	}

	if war, err := dienst.Loesen(context.Background(), domain.ArbeitsartDeck, "3"); err != nil || war {
		t.Errorf("zweites Lösen: %v, %v", war, err)
	}
}

func TestDieUebersichtMeldetIhreFehler(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Ablage antwortet nicht")

	if _, err := service.NeuesWeiterarbeiten(&verlaufsablage{fehler: kaputt}, bestandsprobe{}, festeUhr).
		Uebersicht(context.Background()); !errors.Is(err, kaputt) {
		t.Errorf("eine kaputte Ablage gab %v", err)
	}

	// **Angeheftetes scheitert, Schritte nicht.**
	halb := &halbeAblage{verlaufsablage: &verlaufsablage{}, fehler: kaputt}
	if _, err := service.NeuesWeiterarbeiten(halb, bestandsprobe{}, festeUhr).Uebersicht(context.Background()); !errors.Is(err, kaputt) {
		t.Errorf("ein kaputtes Angeheftetes gab %v", err)
	}

	// **Der Bestand antwortet nicht** — dann steht die Übersicht, statt zu raten.
	ablage := &verlaufsablage{schritte: []domain.Arbeitsschritt{
		{Objekt: domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "1"}, Vorgang: domain.VorgangGeoeffnet, Zeitpunkt: festeUhr()},
	}}
	if _, err := service.NeuesWeiterarbeiten(ablage, bestandsprobe{fehler: kaputt}, festeUhr).Uebersicht(context.Background()); !errors.Is(err, kaputt) {
		t.Errorf("ein kaputter Bestand gab %v", err)
	}

	if err := service.NeuesWeiterarbeiten(ablage, bestandsprobe{fehler: kaputt}, festeUhr).
		Anheften(context.Background(), domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "1"}); !errors.Is(err, kaputt) {
		t.Errorf("Anheften bei kaputtem Bestand gab %v", err)
	}

	// **Auch beim Angehefteten**, wenn die Schritte schon durch sind.
	nurAngeheftet := &verlaufsablage{angeheftet: []domain.Arbeitsobjekt{{Art: domain.ArbeitsartDeck, Kennung: "3"}}}
	if _, err := service.NeuesWeiterarbeiten(nurAngeheftet, bestandsprobe{fehler: kaputt}, festeUhr).Uebersicht(context.Background()); !errors.Is(err, kaputt) {
		t.Errorf("ein kaputter Bestand beim Angehefteten gab %v", err)
	}

	// **Vergessen ist Bestenfalls-Arbeit** (§7, Regel 3): scheitert es, steht
	// die Übersicht trotzdem — ohne das Verschwundene.
	stur := &sturVergessend{verlaufsablage: ablage, fehler: kaputt}
	uebersicht, err := service.NeuesWeiterarbeiten(stur, bestandsprobe{}, festeUhr).Uebersicht(context.Background())
	if err != nil || len(uebersicht.Geoeffnet) != 0 {
		t.Errorf("bei sturem Vergessen: %+v, %v", uebersicht, err)
	}
}

// halbeAblage liest Schritte, scheitert aber am Angehefteten.
type halbeAblage struct {
	*verlaufsablage
	fehler error
}

func (h *halbeAblage) Angeheftete(context.Context) ([]domain.Arbeitsobjekt, error) {
	return nil, h.fehler
}

// sturVergessend vergisst nicht.
type sturVergessend struct {
	*verlaufsablage
	fehler error
}

func (s *sturVergessend) Vergessen(context.Context, domain.Arbeitsart, string) error { return s.fehler }
