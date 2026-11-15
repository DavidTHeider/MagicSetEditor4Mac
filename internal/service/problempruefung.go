package service

// **Die Problemprüfung über ein ganzes Set** — PERF02
// (#2352),
// Befunde A-14, B-13 und Z-03 des Audits 0.9.0.
//
// # Warum es das gibt
//
// **Billiges Zählen und volle Prüfung sind zwei Dinge** (Z-03): die
// Übersicht zählt Karten über SQL in Millisekunden; welche davon ein Problem
// haben, sagt erst das Zeichnen jeder Karte. Diese Prüfung läuft deshalb
// **nicht bei jedem Blick auf die Übersicht**, sondern einmal im Hintergrund,
// und ihr Ergebnis bleibt mit seinem Abdruck stehen, bis jemand neu prüft.
//
// # Wie
//
// **Synchron, mit Meldung je Karte** ([Problempruefung.Pruefen]): die Fläche
// entscheidet, ob sie wartet — der Dienst über seine Aufträge (`Prefer:
// respond-async`), die Kommandozeile mit Fortschritt auf dem Fehlerkanal.
// **Im Hintergrund des Fensters** startet sie erst AP-8
// (#2365),
// das die Zahl zeigt. Karten werden von [Problempruefung.arbeiter] Arbeitern
// geprüft; eine Karte, die sich nicht prüfen lässt, **hat ein Problem** und
// verschwindet nicht aus der Zählung.
//
// # Was es nicht weiß
//
// **Wie geprüft wird** — das ist [Problempruefung.pruefen], gesetzt von der
// Wurzel: Plan und Rasterung über dieselben Teile wie die Vorschau.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Problemarbeiter ist, wie viele Karten zugleich geprüft werden — so viele,
// wie die Pforte des Rasterers Plätze hat (#2020).
const Problemarbeiter = 2

// ProblemstaendeHoechstens ist, wie viele Sets ihren Stand behalten — der
// älteste geht (Zweitblick über Phase 3, Befund 15): ein Dienst, der über
// Wochen Sets anlegt, prüft und löscht, wüchse sonst mit jedem.
const ProblemstaendeHoechstens = 64

// errAbgeloest ist der Grund, mit dem ein neuerer Lauf den älteren abbricht.
var errAbgeloest = fmt.Errorf("%w: eine neuere Prüfung desselben Sets hat diese abgelöst", domain.ErrConflict)

// Problempruefung prüft die Karten eines Sets und hält je Set den letzten
// Stand.
type Problempruefung struct {
	karten   func(ctx context.Context, setID int64) ([]*domain.CardRow, error)
	pruefen  func(ctx context.Context, karte *domain.CardRow) ([]string, error)
	abdruck  func(ctx context.Context, setID int64) (string, error)
	uhr      func() time.Time
	arbeiter int

	mu      sync.Mutex
	staende map[int64]domain.Problemstand

	// laufend ist je Set der Lauf, der gerade prüft — **einer je Set** (§7,
	// Regel 4). Ein neuerer bricht ihn ab, und nur der Lauf, der hier steht,
	// legt seinen Stand ab (Regel 2; Zweitblick über Phase 3, Befund 4).
	laufend map[int64]laufmarke
	nummer  uint64
}

// laufmarke ist ein laufender Lauf: seine Nummer und sein Abbruch.
type laufmarke struct {
	nummer    uint64
	abbrechen context.CancelCauseFunc
}

// NeueProblempruefung baut die Prüfung über ihren Nähten.
func NeueProblempruefung(
	karten func(context.Context, int64) ([]*domain.CardRow, error),
	pruefen func(context.Context, *domain.CardRow) ([]string, error),
	abdruck func(context.Context, int64) (string, error),
	uhr func() time.Time,
) *Problempruefung {
	return &Problempruefung{
		karten: karten, pruefen: pruefen, abdruck: abdruck, uhr: uhr,
		arbeiter: Problemarbeiter, staende: map[int64]domain.Problemstand{},
		laufend: map[int64]laufmarke{},
	}
}

// MitArbeitern setzt, wie viele Karten zugleich geprüft werden — **die Wurzel
// weiß es**: der Plan einer Karte steht nicht hinter der Pforte des Rasterers
// und kostet so viel wie ihre Rasterung (PERF02, gemessen an `Black`).
// Weniger als einer ist einer.
func (p *Problempruefung) MitArbeitern(n int) *Problempruefung {
	p.arbeiter = max(n, 1)

	return p
}

// Stand gibt den letzten Stand eines Sets und den heutigen Abdruck — ein
// Set ohne Prüfung gibt den Nullwert, dessen Urteil »ungeprüft« ist.
//
// **Ein Set, das es nicht mehr gibt, nimmt seinen Stand mit** (Befund 15) —
// beim ersten Fragen danach, nicht beim Löschen: bis dahin hält ihn die
// Grenze von [ProblemstaendeHoechstens] (Nachlauf über Phase 3).
func (p *Problempruefung) Stand(ctx context.Context, setID int64) (domain.Problemstand, string, error) {
	heute, err := p.abdruck(ctx, setID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			p.mu.Lock()
			delete(p.staende, setID)
			p.mu.Unlock()
		}

		return domain.Problemstand{}, "", fmt.Errorf("den Abdruck des Sets bilden: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	return p.staende[setID], heute, nil
}

// Pruefen prüft ein Set und wartet darauf — für die Flächen ohne Fenster.
// `melde` hört nach jeder Karte, wie viele geprüft sind.
//
// **Ein Lauf je Set**: ein zweiter löst den ersten ab, und der erste endet
// mit [domain.ErrConflict], ohne etwas abzulegen. Zurück kommt **der eigene
// Stand** — nicht, was gerade abgelegt ist.
func (p *Problempruefung) Pruefen(
	ctx context.Context, setID int64, melde func(geprueft, gesamt int),
) (domain.Problemstand, error) {
	ctx, nummer, ende := p.anmelden(ctx, setID)
	defer ende()

	karten, abdruck, err := p.eingaben(ctx, setID)
	if err != nil {
		return domain.Problemstand{}, err
	}

	return p.lauf(ctx, setID, nummer, karten, abdruck, melde)
}

// anmelden trägt einen Lauf ein und bricht den bisherigen ab.
func (p *Problempruefung) anmelden(
	ctx context.Context, setID int64,
) (laufctx context.Context, nummer uint64, abmelden func()) {
	laufctx, abbrechen := context.WithCancelCause(ctx)

	p.mu.Lock()
	defer p.mu.Unlock()

	if bisher, da := p.laufend[setID]; da {
		bisher.abbrechen(errAbgeloest)
	}

	p.nummer++
	nummer = p.nummer
	p.laufend[setID] = laufmarke{nummer: nummer, abbrechen: abbrechen}

	return laufctx, nummer, func() {
		p.mu.Lock()
		if p.laufend[setID].nummer == nummer {
			delete(p.laufend, setID)
		}
		p.mu.Unlock()

		abbrechen(nil)
	}
}

// eingaben liest den Abdruck und die Karten, mit denen eine Prüfung beginnt —
// **der Abdruck zuerst** (Befund 3): ändert sich das Set danach, trägt der
// Stand den alten und heißt veraltet. Andersherum trüge er den neuen über
// alte Karten und hieße aktuell.
func (p *Problempruefung) eingaben(ctx context.Context, setID int64) ([]*domain.CardRow, string, error) {
	abdruck, err := p.abdruck(ctx, setID)
	if err != nil {
		return nil, "", fmt.Errorf("den Abdruck des Sets bilden: %w", err)
	}

	karten, err := p.karten(ctx, setID)
	if err != nil {
		return nil, "", fmt.Errorf("die Karten des Sets lesen: %w", err)
	}

	return karten, abdruck, nil
}

// lauf prüft die Karten mit den Arbeitern und legt den Stand ab — auch den
// eines Abbruchs, der dann unvollständig heißt.
func (p *Problempruefung) lauf(
	ctx context.Context, setID int64, nummer uint64, karten []*domain.CardRow, abdruck string, melde func(int, int),
) (domain.Problemstand, error) {
	stand := domain.Problemstand{SetID: setID, Abdruck: abdruck, Begonnen: p.uhr(), Gesamt: len(karten)}

	var (
		mu     sync.Mutex
		gruppe sync.WaitGroup
	)

	arbeit := make(chan *domain.CardRow)

	for range p.arbeiter {
		gruppe.Go(func() {
			for karte := range arbeit {
				befunde, err := p.pruefen(ctx, karte)
				if err != nil && ctx.Err() != nil {
					continue
				}

				if err != nil {
					befunde = append(befunde, err.Error())
				}

				// **Gemeldet unter dem Schloss**: aus zwei Arbeitern zugleich
				// kam die Zählung ungeordnet an, und eine Meldung ging verloren.
				mu.Lock()
				stand.Geprueft++

				if len(befunde) > 0 {
					stand.Karten = append(stand.Karten, domain.Kartenproblem{KarteID: karte.ID, Befunde: befunde})
				}

				if melde != nil {
					melde(stand.Geprueft, len(karten))
				}

				// **Und über den Kontext** (Zweitblick über Phase 3, Befund
				// 11): ein Auftrag des Dienstes und ein Werkzeug über MCP
				// hören hier zu.
				Melde(ctx, domain.Fortschritt{Erledigt: stand.Geprueft, Gesamt: len(karten), Einheit: domain.EinheitKarte})
				mu.Unlock()
			}
		})
	}

	// **Die Arbeiter leeren den Kanal bis zum Ende**, auch nach einem Abbruch
	// — dann kehrt ihr Prüfer sofort zurück. Ein Senden kann deshalb nicht
	// hängen bleiben, und ein `select` darauf würfelte nur, welcher Zweig
	// betreten wird.
	for _, karte := range karten {
		if ctx.Err() != nil {
			break
		}

		arbeit <- karte
	}

	close(arbeit)
	gruppe.Wait()

	slices.SortFunc(stand.Karten, func(a, b domain.Kartenproblem) int { return int(a.KarteID - b.KarteID) })
	stand.Beendet = p.uhr()
	stand.Vollstaendig = ctx.Err() == nil && stand.Geprueft == len(karten)

	p.ablegen(setID, nummer, stand)

	if errors.Is(context.Cause(ctx), errAbgeloest) {
		return stand, fmt.Errorf("die Prüfung von Set %d: %w", setID, errAbgeloest)
	}

	if err := ctx.Err(); err != nil {
		return stand, fmt.Errorf("die Prüfung von Set %d: %w", setID, err)
	}

	return stand, nil
}

// ablegen legt den Stand eines Laufs ab — **nur den des Laufs, der noch
// eingetragen ist**, und einen unvollständigen nicht über einen vollständigen
// desselben Abdrucks: der gilt weiter.
func (p *Problempruefung) ablegen(setID int64, nummer uint64, stand domain.Problemstand) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.laufend[setID].nummer != nummer {
		return
	}

	if bisher, da := p.staende[setID]; da && !stand.Vollstaendig && bisher.Vollstaendig && bisher.Abdruck == stand.Abdruck {
		return
	}

	p.staende[setID] = stand

	// **Über der Grenze geht der älteste** — der, der am längsten fertig ist;
	// nie der eben abgelegte, auch wenn die Uhr zweimal dasselbe sagt.
	if len(p.staende) > ProblemstaendeHoechstens {
		var aeltester *domain.Problemstand

		for kennung, gemerkt := range p.staende {
			if kennung != setID && (aeltester == nil || gemerkt.Beendet.Before(aeltester.Beendet)) {
				aeltester = &gemerkt
			}
		}

		delete(p.staende, aeltester.SetID)
	}
}
