package service_test

// **Was der Wirt in einem Zug hinausgibt, passt nicht in eine Schlange** —
// gefunden am Messlauf „Zwei Macs auf einem" zu Phase 6.7
// (#1668).
//
// Der Griff reiht ein und lässt höchstens `Sitzungstiefe` Nachrichten warten
// — ein Rückstau, der den Absender erreicht. Der Wirt reichte einem Gast, der
// sich an einen laufenden Tisch setzte, Antwort, **jedes** fehlende Ereignis
// und das Bild in einem Zug nach und leerte erst danach: ab gut sechs
// Ereignissen war die Schlange voll, und die Sitzung endete.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schlangendoppel reiht ein wie der Griff — mit Tiefe — und gibt beim Leeren
// hinaus.
type schlangendoppel struct {
	schloss sync.Mutex
	tiefe   int
	warten  [][]byte
	hinaus  [][]byte
	klage   error
}

func (s *schlangendoppel) Sende(_ context.Context, nachricht []byte, _ service.Zustellung) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	if len(s.warten) >= s.tiefe {
		return fmt.Errorf("%w: die Warteschlange ist voll (%d)", domain.ErrConflict, s.tiefe)
	}

	s.warten = append(s.warten, nachricht)

	return nil
}

func (s *schlangendoppel) Leere(context.Context) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	if s.klage != nil {
		return s.klage
	}

	s.hinaus = append(s.hinaus, s.warten...)
	s.warten = nil

	return nil
}

func eineLuecke(laenge int) *nehmerdoppel {
	nehmer := &nehmerdoppel{stand: domain.Wirtsstand{
		Epoche: domain.Partieepoche, Folge: uint64(laenge), Kette: "zuletzt", AelteteFolge: 1,
	}}

	for folge := 1; folge <= laenge; folge++ {
		nehmer.nachholt = append(nehmer.nachholt, domain.Ereignis{Folge: int64(folge), Art: "objekt.tappen"})
	}

	return nehmer
}

func einWirtMitSchlange(nehmer service.Befehlsnehmer, tiefe int) (*service.Partiewirt, *schlangendoppel) {
	schlange := &schlangendoppel{tiefe: tiefe}
	sitzung := service.NeuePartiesitzung("p1", "wirt", schlange, service.NeueReihenfolge(1))

	return service.NeuerPartiewirt("p1", "gast", nehmer, sitzung), schlange
}

// **Eine Lücke, die länger ist als die Schlange, geht ganz hinaus.**
func TestEineLangeLueckePasstDurchEineKurzeSchlange(t *testing.T) {
	t.Parallel()

	const tiefe = 8

	wirt, schlange := einWirtMitSchlange(eineLuecke(3*tiefe), tiefe)

	if fehler := wirt.Angekommen(t.Context(), einRueckkehrwunsch(t, 0, "")); fehler != nil {
		t.Fatalf("angekommen: %v", fehler)
	}

	// **Was am Ende noch wartet, leert die Schleife** — wie im Betrieb.
	if fehler := schlange.Leere(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	ereignisse := 0

	for _, roh := range schlange.hinaus {
		if nachrichtenartVon(t, roh) == domain.ArtEreignis {
			ereignisse++
		}
	}

	if ereignisse != 3*tiefe {
		t.Errorf("%d Ereignisse gingen hinaus, erwartet %d", ereignisse, 3*tiefe)
	}
}

// **Klagt das Hinausgeben, kommt die Klage zurück** — und nicht die volle
// Schlange danach.
func TestWasDasHinausgebenBeimNachreichenKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	wirt, schlange := einWirtMitSchlange(eineLuecke(2), 8)
	klage := errors.New("die Leitung ist weg")
	schlange.klage = klage

	if fehler := wirt.Angekommen(t.Context(), einRueckkehrwunsch(t, 0, "")); !errors.Is(fehler, klage) {
		t.Errorf("angekommen: %v, erwartet die Klage der Leitung", fehler)
	}
}

// **Eine Naht, die nicht hinausgeben kann, hat nichts zu leeren** — die
// Doppel der übrigen Läufe senden sofort.
func TestEineNahtOhneLeerenHatNichtsZuLeeren(t *testing.T) {
	t.Parallel()

	sitzung := service.NeuePartiesitzung("p1", "wirt", &nahtdoppel{}, service.NeueReihenfolge(1))

	if fehler := sitzung.Leere(t.Context()); fehler != nil {
		t.Errorf("Leere: %v", fehler)
	}
}

func nachrichtenartVon(t *testing.T, roh []byte) domain.Nachrichtenart {
	t.Helper()

	var nachricht domain.Partienachricht
	if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
		t.Fatalf("keine Nachricht: %v", fehler)
	}

	return nachricht.Art
}
