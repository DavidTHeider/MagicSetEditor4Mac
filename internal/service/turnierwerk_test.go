package service_test

// **Die Vorrichtung der Turnierleitung** — Werkstück 4 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turnierleitung_test.go` unter der
// Zeilengrenze aus §2 bleibt.
//
// # Was das Doppel wirklich tut
//
// **Es legt ab.** Ein Doppel, das jeden Ruf mit demselben Turnier
// beantwortete, wäre großzügiger als der Betrieb (§3) — es könnte nicht
// zeigen, dass eine Runde den Stand der vorigen sieht.

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// turnierdoppel legt Turniere im Arbeitsspeicher ab.
type turnierdoppel struct {
	turniere map[int64]*domain.Turnier

	naechste int64

	angelegt   int
	angemeldet int
	gesetzt    int

	fehler error

	// setzfehler klagt **nur** beim Eintragen eines Ergebnisses (G7, #1464).
	//
	// **Ein eigener Schalter, weil `fehler` schon beim Lesen greift** — und
	// wer beim Lesen scheitert, erreicht das Eintragen nie. Ein Weg, den kein
	// Lauf betritt, ist eine Zusage ohne Prüfung (§3).
	setzfehler error

	// protokoll ist, was `Protokoll` antwortet — DOM-04.AK-04 (#1699).
	protokoll []domain.Protokolleintrag
}

func neuesTurnierdoppel() *turnierdoppel {
	return &turnierdoppel{turniere: map[int64]*domain.Turnier{}, naechste: 1}
}

func (d *turnierdoppel) Anlegen(_ context.Context, turnier domain.Turnier) (
	domain.Turnier, error,
) {
	if d.fehler != nil {
		return domain.Turnier{}, d.fehler
	}

	d.angelegt++

	turnier.ID, turnier.Fassung = d.naechste, 1
	d.naechste++

	abgelegt := turnier
	d.turniere[turnier.ID] = &abgelegt

	return turnier, nil
}

// Einlesen legt ab wie Anlegen — **mit** den vorläufigen Kennungen; was der
// Speicher daraus macht, prüft `storage` (DOM-04.AK-03, #1699).
func (d *turnierdoppel) Einlesen(ctx context.Context, turnier domain.Turnier) (
	domain.Turnier, error,
) {
	return d.Anlegen(ctx, turnier)
}

func (d *turnierdoppel) Turnier(_ context.Context, kennung int64) (
	domain.Turnier, error,
) {
	if d.fehler != nil {
		return domain.Turnier{}, d.fehler
	}

	turnier, gibts := d.turniere[kennung]
	if !gibts {
		return domain.Turnier{}, domain.ErrNotFound
	}

	return *turnier, nil
}

func (d *turnierdoppel) Turniere(_ context.Context, spiel string) (
	[]domain.Turnier, error,
) {
	alle := []domain.Turnier{}

	for _, turnier := range d.turniere {
		if spiel == "" || turnier.Spiel == spiel {
			alle = append(alle, *turnier)
		}
	}

	return alle, d.fehler
}

func (d *turnierdoppel) Anmelden(
	_ context.Context, turnierID int64, teilnehmer domain.Teilnehmer,
) (domain.Teilnehmer, error) {
	if d.fehler != nil {
		return domain.Teilnehmer{}, d.fehler
	}

	turnier, gibts := d.turniere[turnierID]
	if !gibts {
		return domain.Teilnehmer{}, domain.ErrNotFound
	}

	d.angemeldet++

	teilnehmer.ID = d.naechste
	d.naechste++
	teilnehmer.TurnierID = turnierID
	teilnehmer.Stelle = len(turnier.Teilnehmer) + 1

	turnier.Teilnehmer = append(turnier.Teilnehmer, teilnehmer)

	return teilnehmer, nil
}

func (d *turnierdoppel) Aussteigen(
	_ context.Context, teilnehmerID int64, ausgestiegen bool,
) error {
	for _, turnier := range d.turniere {
		for stelle := range turnier.Teilnehmer {
			if turnier.Teilnehmer[stelle].ID == teilnehmerID {
				turnier.Teilnehmer[stelle].Ausgestiegen = ausgestiegen

				return d.fehler
			}
		}
	}

	return domain.ErrNotFound
}

func (d *turnierdoppel) RundeLegen(
	_ context.Context, turnierID int64, erwartet int, paarungen []domain.Paarung,
) (domain.Runde, error) {
	if d.fehler != nil {
		return domain.Runde{}, d.fehler
	}

	turnier, gibts := d.turniere[turnierID]
	if !gibts {
		return domain.Runde{}, domain.ErrNotFound
	}

	// **Das Doppel prüft die erwartete Nummer wie der Speicher.** Eines, das
	// sie annähme und ignorierte, wäre großzügiger als der Betrieb (§3) — und
	// die Zusage aus Befund 8 stünde ungeprüft da.
	if erwartet != len(turnier.Runden)+1 {
		return domain.Runde{}, domain.ErrConflict
	}

	runde := domain.Runde{
		ID: d.naechste, TurnierID: turnierID, Nummer: len(turnier.Runden) + 1,
	}
	d.naechste++

	for stelle, paarung := range paarungen {
		paarung.ID = d.naechste
		d.naechste++
		paarung.RundeID = runde.ID
		paarung.Stelle = stelle + 1

		runde.Paarungen = append(runde.Paarungen, paarung)
	}

	turnier.Runden = append(turnier.Runden, runde)

	return runde, nil
}

// **Das Doppel hält die Fassung so streng wie der Bestand.**
//
// Eine Vorrichtung, die großzügiger ist als der Betrieb, prüft die
// Vorrichtung und nicht den Weg (§3).
func (d *turnierdoppel) ErgebnisSetzen(
	_ context.Context, fassung, paarungID int64, stand *domain.Ergebnis,
) error {
	d.gesetzt++

	if d.fehler != nil {
		return d.fehler
	}

	if d.setzfehler != nil {
		return d.setzfehler
	}

	for _, turnier := range d.turniere {
		for runde := range turnier.Runden {
			for stelle := range turnier.Runden[runde].Paarungen {
				if turnier.Runden[runde].Paarungen[stelle].ID != paarungID {
					continue
				}

				if turnier.Fassung != fassung {
					return fmt.Errorf(
						"%w: dieses Turnier steht bei Fassung %d, geschrieben "+
							"wurde gegen %d", domain.ErrConflict,
						turnier.Fassung, fassung)
				}

				turnier.Runden[runde].Paarungen[stelle].Ergebnis = stand
				turnier.Fassung++

				return nil
			}
		}
	}

	return domain.ErrNotFound
}

func (d *turnierdoppel) Loeschen(_ context.Context, kennung int64) error {
	if _, gibts := d.turniere[kennung]; !gibts {
		return domain.ErrNotFound
	}

	delete(d.turniere, kennung)

	return d.fehler
}

// Abschliessen setzt ein Datum — die Regeln dazu hält der Dienst.
func (d *turnierdoppel) Abschliessen(_ context.Context, turnierID, _ int64) error {
	turnier, gibts := d.turniere[turnierID]
	if !gibts {
		return domain.ErrNotFound
	}

	turnier.Abgeschlossen = "2026-09-27T12:00:00Z"
	turnier.Fassung++

	return nil
}

// Wiedereroeffnen nimmt das Datum weg.
func (d *turnierdoppel) Wiedereroeffnen(_ context.Context, turnierID, _ int64) error {
	turnier, gibts := d.turniere[turnierID]
	if !gibts {
		return domain.ErrNotFound
	}

	turnier.Abgeschlossen = ""
	turnier.Fassung++

	return nil
}

// Protokoll gibt, was belegt ist.
func (d *turnierdoppel) Protokoll(_ context.Context, turnierID int64) (
	[]domain.Protokolleintrag, error,
) {
	if _, gibts := d.turniere[turnierID]; !gibts {
		return nil, domain.ErrNotFound
	}

	return d.protokoll, nil
}

// deckabzugsdoppel gibt ein festes Deck.
type deckabzugsdoppel struct {
	deck   domain.Deck
	fehler error
}

func (d *deckabzugsdoppel) Deck(_ context.Context, _ int64) (domain.Deck, error) {
	return d.deck, d.fehler
}

// kartennamendoppel gibt abgelegte Kartennamen.
type kartennamendoppel struct {
	namen  map[int64]string
	fehler error

	// felder sind die Feldwerte, die `FeldZurKarte` gibt (#1376).
	felder map[string]string

	// feldfehler scheitert **nur** beim Feld, nicht beim Namen. Beides an
	// einem Feld hängen zu lassen hieße, die beiden Wege nur zusammen prüfen
	// zu können — und der zweite fällt still aus (#1376).
	feldfehler error
}

func (d *kartennamendoppel) FeldZurKarte(
	_ context.Context, _ int64, feld string,
) (string, error) {
	if d.feldfehler != nil {
		return "", d.feldfehler
	}

	return d.felder[feld], d.fehler
}

func (d *kartennamendoppel) NameZurKarte(_ context.Context, karteID int64) (
	string, error,
) {
	return d.namen[karteID], d.fehler
}

// turnierwerk hält alles zusammen, was ein Lauf braucht.
type turnierwerk struct {
	turniere *turnierdoppel
	decks    *deckabzugsdoppel
	karten   *kartennamendoppel

	leitung *service.Turnierleitung
}

func neuesTurnierwerk(t *testing.T) *turnierwerk {
	t.Helper()

	werk := &turnierwerk{
		turniere: neuesTurnierdoppel(),
		decks:    &deckabzugsdoppel{},
		karten:   &kartennamendoppel{namen: map[int64]string{}},
	}

	werk.leitung = service.NeueTurnierleitung(werk.turniere, werk.decks, werk.karten)

	return werk
}

// turnierAnlegen legt ein brauchbares Turnier an.
func (w *turnierwerk) turnierAnlegen(t *testing.T) domain.Turnier {
	t.Helper()

	turnier, err := w.leitung.Anlegen(t.Context(),
		domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	return turnier
}

// anmelden meldet Namen an und gibt ihre Kennungen.
func (w *turnierwerk) anmelden(
	t *testing.T, turnierID int64, namen ...string,
) []int64 {
	t.Helper()

	kennungen := make([]int64, 0, len(namen))

	for _, name := range namen {
		angemeldet, err := w.leitung.Anmelden(t.Context(), turnierID,
			domain.Teilnehmer{Name: name})
		if err != nil {
			t.Fatalf("%s anmelden: %v", name, err)
		}

		kennungen = append(kennungen, angemeldet.ID)
	}

	return kennungen
}

// fassung liest die aktuelle Fassung aus dem Doppel.
func (w *turnierwerk) fassung(t *testing.T, turnierID int64) int64 {
	t.Helper()

	turnier, gibts := w.turniere.turniere[turnierID]
	if !gibts {
		t.Fatalf("das Turnier %d gibt es im Doppel nicht", turnierID)
	}

	return turnier.Fassung
}
