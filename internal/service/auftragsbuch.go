package service

// **Das Auftragsbuch** — Werkstück U11 von Phase 5.1
// (#1422):
// die gemeinsame Naht für alles, was länger läuft. Wer einen Auftrag
// beginnt, bekommt eine Kennung, meldet Fortschritt und beendet ihn —
// und die Übersicht liest, was läuft und was zuletzt fertig wurde.
//
// **Im Arbeitsspeicher, je Prozess**: ein Auftrag lebt so lange wie das
// Fenster, das ihn angestoßen hat. Nach einem Neustart läuft nichts mehr,
// und das ist die Wahrheit, nicht ein Verlust.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Behalten ist, wie viele beendete Aufträge das Buch nennt.
const Behalten = 5

// Auftragsbuch führt die laufenden und die zuletzt beendeten Aufträge.
type Auftragsbuch struct {
	mu       sync.Mutex
	uhr      func() time.Time
	naechste int64
	laufend  []domain.Auftrag
	beendet  []domain.Auftrag

	// abbrecher hält den Abbruch je laufendem Auftrag, der einen hat.
	abbrecher map[int64]func()
}

// NeuesAuftragsbuch baut das Buch mit einer Uhr.
func NeuesAuftragsbuch(uhr func() time.Time) *Auftragsbuch {
	return &Auftragsbuch{uhr: uhr}
}

// Beginnen trägt einen Auftrag ein und gibt seine Kennung.
func (b *Auftragsbuch) Beginnen(art, was string) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.naechste++
	b.laufend = append(b.laufend, domain.Auftrag{
		ID: b.naechste, Art: art, Was: was, Stand: domain.AuftragLaeuft, Begonnen: b.uhr(),
	})

	return b.naechste
}

// Fortschreiben setzt den Stand eines laufenden Auftrags.
func (b *Auftragsbuch) Fortschreiben(id int64, fortschritt string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.laufend {
		if b.laufend[i].ID == id {
			b.laufend[i].Fortschritt = fortschritt
		}
	}
}

// Abbrechbar gibt einem laufenden Auftrag seinen Abbruch — FMT-03.AK-02.
func (b *Auftragsbuch) Abbrechbar(id int64, abbrechen func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.laufend {
		if b.laufend[i].ID == id {
			if b.abbrecher == nil {
				b.abbrecher = map[int64]func(){}
			}

			b.abbrecher[id] = abbrechen
			b.laufend[i].Abbrechbar = true
		}
	}
}

// Zaehlen setzt den gezählten Stand eines laufenden Auftrags.
func (b *Auftragsbuch) Zaehlen(id int64, stand domain.Fortschritt) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.laufend {
		if b.laufend[i].ID == id {
			b.laufend[i].Gezaehlt = stand
		}
	}
}

// Abbrechen hält einen laufenden Auftrag an. **Beendet ist er erst, wenn sein
// Lauf es sagt** — zwischen zwei Stücken, damit nur Ganzes stehen bleibt.
func (b *Auftragsbuch) Abbrechen(id int64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, auftrag := range b.laufend {
		if auftrag.ID != id {
			continue
		}

		abbrechen, da := b.abbrecher[id]
		if !da {
			return fmt.Errorf("%w: dieser Auftrag läuft zu Ende und lässt sich nicht anhalten", domain.ErrConflict)
		}

		abbrechen()

		return nil
	}

	return fmt.Errorf("%w: der Auftrag läuft nicht mehr", domain.ErrNotFound)
}

// Beenden schließt einen Auftrag — fertig, gescheitert mit Grund oder
// abgebrochen.
func (b *Auftragsbuch) Beenden(id int64, fehler error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, auftrag := range b.laufend {
		if auftrag.ID != id {
			continue
		}

		auftrag.Beendet = b.uhr()
		auftrag.Stand = domain.AuftragFertig
		auftrag.Abbrechbar = false
		delete(b.abbrecher, id)

		switch {
		case errors.Is(fehler, context.Canceled):
			auftrag.Stand = domain.AuftragAbgebrochen
		case fehler != nil:
			auftrag.Stand = domain.AuftragGescheitert
			auftrag.Grund = fehler.Error()
		}

		b.laufend = append(b.laufend[:i], b.laufend[i+1:]...)
		b.beendet = append([]domain.Auftrag{auftrag}, b.beendet...)

		if len(b.beendet) > Behalten {
			b.beendet = b.beendet[:Behalten]
		}

		return
	}
}

// Auftraege gibt die laufenden zuerst, dann die zuletzt beendeten — Kopien.
func (b *Auftragsbuch) Auftraege() []domain.Auftrag {
	b.mu.Lock()
	defer b.mu.Unlock()

	alle := make([]domain.Auftrag, 0, len(b.laufend)+len(b.beendet))
	alle = append(alle, b.laufend...)
	alle = append(alle, b.beendet...)

	return alle
}
