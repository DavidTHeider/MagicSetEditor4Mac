// **Preise holen, ablegen und zusammenzählen** — Werkstück 3 von Phase 3.4
// (#1165).
package service_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// preisdoppel ist eine Ablage im Arbeitsspeicher.
type preisdoppel struct {
	staende []domain.Preisstand
	fehler  error

	// lesezaehler zählt die **lesenden** Rufe.
	//
	// **Er steht hier und nicht in einem zweiten Doppel** (#1292): die Zusage
	// „das Portfolio fragt einmal" ist eine über diese Ablage, und ein Doppel
	// daneben müsste jeden Weg noch einmal nachbauen.
	//
	// **Atomar, weil die Läufe parallel sind.** Ein `int` daneben war ein
	// Wettlauf — vom Race-Detektor am 2026-08-31 gefunden, und zwar in einem
	// Lauf, der mit dem Zähler nichts zu tun hat: `Portfolio` ruft `Neueste`
	// aus mehreren Tests zugleich.
	lesezaehler atomic.Int64
}

func (p *preisdoppel) Eintragen(
	_ context.Context, staende ...domain.Preisstand,
) error {
	if p.fehler != nil {
		return p.fehler
	}

	// **Alle oder keiner — auch hier.** Ein Doppel, das den halben Zug
	// hinterließe, wo der Bestand ihn zurücknimmt, wäre großzügiger als der
	// Betrieb (§3).
	p.staende = append(p.staende, staende...)

	return nil
}

func (p *preisdoppel) Neueste(_ context.Context, printing string) (
	[]domain.Preisstand, error,
) {
	p.lesezaehler.Add(1)

	if p.fehler != nil {
		return nil, p.fehler
	}

	gefunden := []domain.Preisstand{}

	for _, stand := range p.staende {
		if stand.Printing == printing {
			gefunden = append(gefunden, stand)
		}
	}

	return gefunden, nil
}

// NeuesteAlle gibt alles, was abgelegt ist.
//
// **Ohne die Auswahl „je Reihe der jüngste"**, die der echte Speicher trifft.
// Das ist der Unterschied, den eine Vorrichtung machen darf: hier legt jeder
// Test ab, was er meint, und ein zweiter Stand derselben Reihe kommt in keinem
// vor. Wo es darauf ankommt, prüft `internal/storage` gegen SQLite (§3).
func (p *preisdoppel) NeuesteAlle(_ context.Context) (
	[]domain.Preisstand, error,
) {
	p.lesezaehler.Add(1)

	if p.fehler != nil {
		return nil, p.fehler
	}

	return p.staende, nil
}

func (p *preisdoppel) Verlauf(
	_ context.Context, printing string, _ domain.Preisstand,
) ([]domain.Preisstand, error) {
	if p.fehler != nil {
		return nil, p.fehler
	}

	return p.Neueste(context.Background(), printing) //nolint:contextcheck // das Doppel filtert nur
}

// quellendoppel gibt feste Preise.
type quellendoppel struct {
	// preise sind die Preise, wie Scryfall sie führt.
	//
	// **Sie werden über [scryfall.Preis.AlsStand] neutral gemacht** (A3,
	// #1442) und nicht von Hand zusammengesetzt: so geht jeder Lauf durch
	// dieselbe Zuordnung wie der Betrieb (§3).
	preise  []scryfall.Preis
	staende []domain.Preisstand
	quelle  domain.Preisquelle
	alter   time.Duration
	befunde []domain.Befund
	fehler  error
}

func (q quellendoppel) PreiseFuerDruck(_ context.Context, kennung string) (
	[]domain.Preisstand, time.Duration, []domain.Befund, error,
) {
	if q.staende != nil {
		return q.staende, q.alter, q.befunde, q.fehler
	}

	staende := make([]domain.Preisstand, 0, len(q.preise))
	for _, preis := range q.preise {
		staende = append(staende, preis.AlsStand(kennung))
	}

	return staende, q.alter, q.befunde, q.fehler
}

// Name gibt die Quelle, in deren Namen dieses Doppel antwortet.
func (q quellendoppel) Name() domain.Preisquelle {
	if q.quelle != "" {
		return q.quelle
	}

	return domain.PreisquelleScryfall
}

// probeuhr ist eine feste Uhr.
func probeuhr() func() time.Time {
	fest := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)

	return func() time.Time { return fest }
}

// **Geholte Preise werden abgelegt und kommen zurück.**
func TestGeholtePreiseWerdenAbgelegt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{}
	pflege := service.NeuePreispflege(ablage, &sammlungsdoppel{}, quellendoppel{
		preise: []scryfall.Preis{
			{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 399},
			{Waehrung: "USD", Finish: domain.FinishFoil, Betrag: 1200},
		},
	}, probeuhr())

	staende, _, _, err := pflege.Holen(t.Context(), "abc")
	if err != nil {
		t.Fatalf("Holen: %v", err)
	}

	if len(staende) != 2 || len(ablage.staende) != 2 {
		t.Fatalf("%d zurück, %d abgelegt", len(staende), len(ablage.staende))
	}

	if ablage.staende[0].Printing != "abc" ||
		ablage.staende[0].Quelle != domain.PreisquelleScryfall {
		t.Errorf("abgelegt wurde %+v", ablage.staende[0])
	}
}

// **Ein Stand aus dem Zwischenspeicher gibt sich nicht frischer, als er ist.**
//
// # Warum das die entscheidende Zeile ist
//
// „Ohne Netz arbeitet die Anwendung weiter, und der letzte Stand sagt, wie alt
// er ist" (§16, Satz 5). Wer den Zeitpunkt des **Eintrags** nimmt statt den des
// **Abrufs**, macht aus einem drei Tage alten Preis einen von jetzt — und
// genau darauf verlässt sich die Übersicht.
func TestEinAlterStandBleibtAlt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{}
	uhr := probeuhr()

	pflege := service.NeuePreispflege(ablage, &sammlungsdoppel{}, quellendoppel{
		preise: []scryfall.Preis{{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 399}},
		alter:  50 * time.Hour,
	}, uhr)

	staende, alter, _, err := pflege.Holen(t.Context(), "abc")
	if err != nil {
		t.Fatalf("Holen: %v", err)
	}

	if alter != 50*time.Hour {
		t.Errorf("das Alter der Antwort ist %v", alter)
	}

	if gemessen := staende[0].Alter(uhr()); gemessen != 50*time.Hour {
		t.Errorf("der Stand gibt sich %v alt", gemessen)
	}
}

// **Befunde der Quelle reisen mit — auch wenn Preise ankommen.**
func TestBefundeDerQuelleReisenMit(t *testing.T) {
	t.Parallel()

	pflege := service.NeuePreispflege(&preisdoppel{}, &sammlungsdoppel{}, quellendoppel{
		preise:  []scryfall.Preis{{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 1}},
		befunde: []domain.Befund{{Meldung: "einer ist nicht zu deuten"}},
	}, probeuhr())

	_, _, befunde, err := pflege.Holen(t.Context(), "abc")
	if err != nil {
		t.Fatalf("Holen: %v", err)
	}

	if len(befunde) != 1 {
		t.Errorf("es kamen %d Befunde", len(befunde))
	}
}

// **Ein Fehler der Quelle bringt die Befunde trotzdem mit.**
//
// Sonst verschwände die Auskunft „drei Preise waren nicht zu deuten" hinter
// dem einen Satz „es ging nicht".
func TestEinFehlerDerQuelleBringtBefundeMit(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Gegenstelle schweigt")

	pflege := service.NeuePreispflege(&preisdoppel{}, &sammlungsdoppel{}, quellendoppel{
		befunde: []domain.Befund{{Meldung: "ein Befund"}},
		fehler:  gestolpert,
	}, probeuhr())

	_, _, befunde, err := pflege.Holen(t.Context(), "abc")

	if !errors.Is(err, gestolpert) {
		t.Errorf("der Fehler ist %v", err)
	}

	if len(befunde) != 1 {
		t.Errorf("es kamen %d Befunde", len(befunde))
	}
}

// **Und ein Fehler der Ablage bricht ab.**
func TestEinFehlerBeimAblegenBrichtAb(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Ablage stolpert")

	pflege := service.NeuePreispflege(&preisdoppel{fehler: gestolpert}, &sammlungsdoppel{}, quellendoppel{
		preise: []scryfall.Preis{{Waehrung: "EUR", Finish: domain.FinishNonfoil, Betrag: 1}},
	}, probeuhr())

	if _, _, _, err := pflege.Holen(t.Context(), "abc"); !errors.Is(err, gestolpert) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Neueste und Verlauf reichen durch — samt Fehler.**
func TestNeuesteUndVerlaufReichenDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Ablage stolpert")
	pflege := service.NeuePreispflege(&preisdoppel{fehler: gestolpert},
		&sammlungsdoppel{}, quellendoppel{}, probeuhr())

	if _, err := pflege.Neueste(t.Context(), "abc"); !errors.Is(err, gestolpert) {
		t.Errorf("Neueste: %v", err)
	}

	if _, err := pflege.Verlauf(t.Context(), "abc",
		reihenangabe("EUR", domain.FinishNonfoil)); !errors.Is(err, gestolpert) {
		t.Errorf("Verlauf: %v", err)
	}
}

// reihenangabe benennt eine Reihe von Scryfall (A3, #1442).
//
// **Als Helfer, weil sie sieben Teile hat**: sieben Zeilen je Aufruf sagten
// nichts über den Lauf, in dem sie stehen.
func reihenangabe(waehrung string, finish domain.Finish) domain.Preisstand {
	return domain.Preisstand{
		Quelle:   domain.PreisquelleScryfall,
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Waehrung: waehrung,
		Finish:   finish,
	}
}
