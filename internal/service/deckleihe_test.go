package service_test

// **Die digitale Deckleihe** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Wogegen hier geprüft wird
//
//   - Der **Abzug wird eingefroren**: wer sein Deck danach ändert oder
//     löscht, ändert die Leihe nicht (zweite Abnahme des Issues).
//   - Die **Unterschrift** hängt am Verleiher; ein geändertes Dokument wird
//     abgewiesen.
//   - **Vier Abweisungen** vor dem Start: anderer Peer, falsche Partie,
//     falscher Abdruck, abgelaufen.
//   - Ein **Widerruf** verhindert neue Starts und lässt eine laufende Nutzung
//     nach dem vereinbarten Modus.

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

var leihzeit = time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)

// leihprobe hält Leihen im Arbeitsspeicher.
type leihprobe struct {
	leihen map[string]domain.Leihstand

	klagt error

	// schreibklage klagt **nur** beim Schreiben.
	//
	// **Ein eigener Schalter, weil `klagt` schon beim Lesen greift** — und
	// wer beim Lesen scheitert, erreicht das Schreiben nie. Ein Weg, den
	// kein Lauf betritt, ist eine Zusage ohne Prüfung (§3).
	schreibklage error
}

func neueLeihprobe() *leihprobe {
	return &leihprobe{leihen: map[string]domain.Leihstand{}}
}

func (p *leihprobe) LeiheSchreiben(
	_ context.Context, stand domain.Leihstand,
) error {
	if p.klagt != nil {
		return p.klagt
	}

	if p.schreibklage != nil {
		return p.schreibklage
	}

	p.leihen[stand.Leihe.Kennung] = stand

	return nil
}

func (p *leihprobe) Leihe(
	_ context.Context, kennung string,
) (domain.Leihstand, error) {
	if p.klagt != nil {
		return domain.Leihstand{}, p.klagt
	}

	stand, da := p.leihen[kennung]
	if !da {
		return domain.Leihstand{}, fmt.Errorf(
			"%w: die Leihe %q gibt es nicht", domain.ErrNotFound, kennung)
	}

	return stand, nil
}

func (p *leihprobe) Leihen(
	_ context.Context,
) ([]domain.Leihstand, error) {
	if p.klagt != nil {
		return nil, p.klagt
	}

	gesammelt := make([]domain.Leihstand, 0, len(p.leihen))
	for _, stand := range p.leihen {
		gesammelt = append(gesammelt, stand)
	}

	sort.Slice(gesammelt, func(eins, zwei int) bool {
		return gesammelt[eins].Leihe.Kennung < gesammelt[zwei].Leihe.Kennung
	})

	return gesammelt, nil
}

// zweiPeers gibt zwei Identitäten — Verleiher und Entleiher.
func zweiPeers(t *testing.T) (verleiher, entleiher peer.Identitaet) {
	t.Helper()

	erste, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("die erste Identität: %v", fehler)
	}

	zweite, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("die zweite Identität: %v", fehler)
	}

	return erste, zweite
}

// eineLeihstelle baut sie samt Deck.
func eineLeihstelle(
	t *testing.T, ich peer.Identitaet,
) (*service.Leihstelle, *leihprobe, *deckprobe) {
	t.Helper()

	ablage, decks := neueLeihprobe(), &deckprobe{
		decks: map[int64]domain.Deck{7: {
			ID: 7, Name: "Wald", Spiel: "magic",
			Plaetze: []domain.Deckplatz{{
				KarteID: 1, Bezeichnung: "Wald", Menge: 40,
			}},
		}},
	}

	stelle := service.NeueLeihstelle(ablage, decks, ausweisprobe{ich: ich},
		func() time.Time { return leihzeit })

	return stelle, ablage, decks
}

// **Eine angebotene Leihe trägt den Abzug und ist unterschrieben.**
func TestEineLeiheWirdAusgestelltUndUnterschrieben(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	stelle, _, _ := eineLeihstelle(t, anna)

	leihe, abzug, fehler := stelle.Anbieten(t.Context(), service.Leihwunsch{
		Entleiher: string(bert.Kennung()), DeckID: 7,
		Bindung: domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
		Rechte:  []domain.Leihrecht{domain.RechtSpielen},
		Dauer:   24 * time.Hour, Modus: domain.LaufZuEnde,
	})
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	if leihe.Verleiher != string(anna.Kennung()) {
		t.Errorf("der Verleiher ist %q", leihe.Verleiher)
	}

	if leihe.Abdruck == "" || leihe.Abdruck != abzug.Abdruck {
		t.Fatalf("der Abdruck ist %q, der des Abzugs %q",
			leihe.Abdruck, abzug.Abdruck)
	}

	if fehler := service.PruefeLeihe(leihe); fehler != nil {
		t.Fatalf("die eigene Unterschrift hält nicht: %v", fehler)
	}
}

// **Ein geändertes Dokument wird abgewiesen.**
func TestEineGeaenderteLeiheWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	stelle, _, _ := eineLeihstelle(t, anna)

	leihe, _, fehler := stelle.Anbieten(t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	leihe.Bis = leihe.Bis.Add(10000 * time.Hour)

	if fehler := service.PruefeLeihe(leihe); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Der Abzug wird eingefroren** — die zweite Abnahme des Issues.
func TestEinGeloeschtesDeckAendertDieLeiheNicht(t *testing.T) {
	t.Parallel()

	anna, bert := zweiPeers(t)
	stelle, ablage, decks := eineLeihstelle(t, anna)

	leihe, abzug, fehler := stelle.Anbieten(t.Context(), einLeihwunsch(bert))
	if fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	// **Das Ursprungsdeck verschwindet.**
	delete(decks.decks, 7)

	stand, fehler := ablage.Leihe(t.Context(), leihe.Kennung)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if stand.Abzug.Abdruck != abzug.Abdruck ||
		len(stand.Abzug.Karten) != len(abzug.Karten) {
		t.Fatalf("der eingefrorene Abzug hat sich geändert: %+v", stand.Abzug)
	}
}

// einWunsch gibt den gewöhnlichen Wunsch an diesen Entleiher.
func einLeihwunsch(entleiher peer.Identitaet) service.Leihwunsch {
	return service.Leihwunsch{
		Entleiher: string(entleiher.Kennung()), DeckID: 7,
		Bindung: domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
		Rechte:  []domain.Leihrecht{domain.RechtSpielen},
		Dauer:   24 * time.Hour, Modus: domain.LaufZuEnde,
	}
}
