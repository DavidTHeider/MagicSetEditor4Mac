package app_test

// **Ein zweiter Packartenlauf beendet den ersten** — §7, Regel 4
// (#1288).
//
// # Warum das hier steht und nicht bei der Vorschau
//
// `Packarten` rechnet über **jede** Karte des Sets: bei `Vanguard.mse-set` mit
// 107 Karten sind das 3,3 s im Fenster, gemessen am 2026-08-29 an der
// laufenden Brücke. Wer währenddessen das Set wechselt, ließ diesen Lauf bis
// dahin zu Ende rechnen — für eine Antwort, die `useLaden` gar nicht mehr
// einträgt (§7, Regel 2).
//
// **Der Abbruch war schon halb da**: `Packbereich.gefilterteKarten` fragt je
// Karte `ctx.Err()`. Was fehlte, war ein Kontext, den je zweiter Aufruf
// beendet — die Bindung reichte den Fensterkontext durch, und der endet erst
// mit dem Fenster.
//
// # Was dieser Lauf **nicht** hält
//
// **Wie lange `Packarten` braucht.** Eine Schranke aus der Geschwindigkeit
// dieses Rechners wäre eine Zusage über den Rechner. Was das Fenster tut,
// während gerechnet wird, zeigt `useLaden` und nicht dieser Test.

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// packbremse hält den **ersten** Lauf fest, bis sein Kontext endet.
//
// **Nur den ersten**: der zweite muss durchkommen, sonst prüfte der Test eine
// Verklemmung statt eines Abbruchs. Dieselbe Gestalt wie `bremsprobe` bei der
// Vorschau, und aus demselben Grund.
type packbremse struct {
	mu     sync.Mutex
	laeufe int

	drin        chan struct{}
	abgebrochen chan struct{}
}

func (p *packbremse) Angebot(ctx context.Context, _ int64) (
	[]service.Packangebot, []domain.Befund, error,
) {
	p.mu.Lock()
	p.laeufe++
	erster := p.laeufe == 1
	p.mu.Unlock()

	if !erster {
		return []service.Packangebot{{Name: "durch"}}, nil, nil
	}

	close(p.drin)
	<-ctx.Done()
	close(p.abgebrochen)

	return nil, nil, ctx.Err()
}

func (p *packbremse) Ziehen(context.Context, int64, string, uint64) (
	[]*domain.CardRow, []domain.Befund, error,
) {
	return nil, nil, nil
}

func TestEinZweiterPackartenlaufBeendetDenErsten(t *testing.T) {
	bremse := &packbremse{
		drin:        make(chan struct{}),
		abgebrochen: make(chan struct{}),
	}

	anwendung := app.NeueApp("")
	anwendung.MitPack(bremse)

	bindungen := app.NeueBindungen(anwendung)

	erster := make(chan app.PackAntwort, 1)

	go func() {
		erster <- bindungen.Packarten(t.Context(), "1")
	}()

	// **Erst wenn der erste Lauf wirklich drinsteht.** Ein zweiter Aufruf
	// davor bräche nichts ab, und der Test wäre grün, ohne seinen Fall zu
	// enthalten (§10).
	<-bremse.drin

	zweiter := bindungen.Packarten(t.Context(), "1")

	select {
	case <-bremse.abgebrochen:
	case <-time.After(5 * time.Second):
		t.Fatal("der erste Lauf lief weiter, obwohl ein zweiter begonnen hat")
	}

	antwort := <-erster
	if antwort.Fehler == nil {
		t.Fatal("der abgebrochene Lauf meldete keinen Fehler")
	}

	// **Und er meldet `abgebrochen`, nicht `unbekannt`.** Das ist der
	// Unterschied, den das Fenster liest: `useLaden` lässt bei `abgebrochen`
	// den vorigen Stand stehen und holt einmal nach; bei einem unbekannten
	// Fehler stünde „Es ist ein unerwarteter Fehler aufgetreten" in der
	// Lasche (#288).
	if antwort.Fehler.Code != app.CodeAbgebrochen {
		t.Errorf("der abgebrochene Lauf meldete %q statt %q",
			antwort.Fehler.Code, app.CodeAbgebrochen)
	}

	// **Der zweite darf nicht mitsterben.**
	if zweiter.Fehler != nil {
		t.Errorf("der zweite Lauf wurde mit abgeräumt: %+v", zweiter.Fehler)
	}

	if len(zweiter.Arten) != 1 {
		t.Errorf("der zweite Lauf gab %d Arten statt einer", len(zweiter.Arten))
	}
}
