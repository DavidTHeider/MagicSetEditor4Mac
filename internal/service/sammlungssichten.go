package service

// **Die Fragen an eine geordnete Sammlung** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
//
// # Warum das hier steht und nicht in SQL
//
// Weil die Sammlung ohnehin ganz geladen wird — sie hängt an keinem Set und
// wächst mit dem, was jemand besitzt, nicht mit dem Pack. Dieselbe Begründung
// wie bei der Sicht im Fenster, und dieselbe Grenze: bei einer Sammlung, die
// nicht in den Arbeitsspeicher passt, gehört beides ins SQL.
//
// **Gemessen ist es**: die ganze Liste zu laden kostet bei 5.000 Exemplaren
// 4,5 ms (`docs/SAMMLUNG.md`).

import (
	"context"
	"fmt"
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Besitzstand sagt, was von einem Druck da ist — und was gewünscht.
type Besitzstand struct {
	// Stueck ist die Zahl der **besessenen** Stücke.
	Stueck int

	// Gewuenscht ist die Zahl der gewünschten Stücke.
	//
	// **Getrennt gezählt** (0057): ein Wunsch ist kein Besitz, und beides in
	// einer Zahl wäre die Antwort auf keine Frage.
	Gewuenscht int

	// Zeilen sind die Exemplare dieses Drucks — Besitz und Wunsch.
	Zeilen []*domain.Exemplar
}

// Besitzt sagt, ob mindestens ein Stück da ist.
func (b Besitzstand) Besitzt() bool { return b.Stueck > 0 }

// Lagerortzeile ist ein Lagerort samt dem, was dort liegt.
type Lagerortzeile struct {
	// Lagerort ist der Freitext — leer heißt „ohne Angabe".
	Lagerort string

	// Zeilen ist die Zahl der Exemplarzeilen dort.
	Zeilen int

	// Stueck ist die Zahl der Stücke dort.
	Stueck int
}

// Dublettenzeile ist ein Exemplar, von dem mehr da ist als gewollt.
type Dublettenzeile struct {
	Exemplar *domain.Exemplar

	// Ueberzaehlig ist, wie viele über der behaltenen Zahl liegen.
	Ueberzaehlig int
}

// Sammlungssichten beantwortet Fragen an die Sammlung.
type Sammlungssichten struct {
	ablage Sammlungsablage

	// reserviert ist die Naht zum Handeln (T2, #1467) — **wahlfrei**, und
	// ohne sie ist festgehalten null. Dieselbe Bauform wie bei
	// [Deckzuordnungspflege.MitReservierungen].
	reserviert Reservierungen
}

// NeueSammlungssichten baut sie.
func NeueSammlungssichten(ablage Sammlungsablage) *Sammlungssichten {
	return &Sammlungssichten{ablage: ablage}
}

// MitReservierungen setzt die Naht zum Handeln (T2, #1467, 0120).
//
// **Damit ein zugesagtes Stück nicht als überzählig dasteht.** „Überzählig"
// heißt „kann weg"; was einem Handel zugesagt ist, kann nicht weg — und wer es
// weiter anböte, böte es ein zweites Mal an.
func (s *Sammlungssichten) MitReservierungen(reserviert Reservierungen) *Sammlungssichten {
	s.reserviert = reserviert

	return s
}

// Besitz sagt, was von einem Druck da ist.
//
// **Die Naht zum Deckbau** (#1292): „besitze ich diese Karte?" ist die Frage,
// die Phase 4.6 stellt, wenn ein Deck eine Karte nennt.
func (s *Sammlungssichten) Besitz(ctx context.Context, printing string) (
	Besitzstand, error,
) {
	zeilen, err := s.ablage.Fuer(ctx, printing)
	if err != nil {
		return Besitzstand{}, err
	}

	stand := Besitzstand{Zeilen: zeilen}

	for _, zeile := range zeilen {
		if zeile.Besitzt() {
			stand.Stueck += zeile.Menge

			continue
		}

		stand.Gewuenscht += zeile.Menge
	}

	return stand, nil
}

// Lagerorte gibt die Lagerorte samt dem, was dort liegt — nach Namen geordnet.
//
// **Wünsche zählen nicht mit**: sie liegen nirgends (0057).
func (s *Sammlungssichten) Lagerorte(ctx context.Context) ([]Lagerortzeile, error) {
	exemplare, err := s.ablage.Liste(ctx)
	if err != nil {
		return nil, err
	}

	nachOrt := map[string]*Lagerortzeile{}

	for _, exemplar := range exemplare {
		if !exemplar.Besitzt() {
			continue
		}

		zeile, da := nachOrt[exemplar.Lagerort]
		if !da {
			zeile = &Lagerortzeile{Lagerort: exemplar.Lagerort}
			nachOrt[exemplar.Lagerort] = zeile
		}

		zeile.Zeilen++
		zeile.Stueck += exemplar.Menge
	}

	orte := make([]Lagerortzeile, 0, len(nachOrt))
	for _, zeile := range nachOrt {
		orte = append(orte, *zeile)
	}

	// **Nach Namen und nicht nach Menge.** Wer einen Lagerort sucht, sucht
	// seinen Namen; eine Reihenfolge nach Größe ließe ihn jedes Mal woanders
	// stehen.
	sort.Slice(orte, func(links, rechts int) bool {
		return orte[links].Lagerort < orte[rechts].Lagerort
	})

	return orte, nil
}

// Dubletten gibt die Exemplare, von denen mehr da ist als `behalten`.
//
// # Warum die Zahl hereinkommt
//
// **Weil „Dublette" kein Zustand ist, sondern ein Verhältnis** (0057): *mehr
// als ich behalten will*. Meist 1, bei Ländern 4, beim Deckbau anders. Eine
// Spalte im Bestand müsste bei jeder Mengenänderung nachgeführt werden und
// wäre falsch, sobald jemand die Zahl anders meint.
//
// **Wünsche sind keine Dubletten**: man hat sie nicht.
func (s *Sammlungssichten) Dubletten(ctx context.Context, behalten int) (
	[]Dublettenzeile, error,
) {
	if behalten < 0 {
		return nil, fmt.Errorf(
			"%w: %d ist keine Stückzahl, die man behält", domain.ErrValidation, behalten)
	}

	exemplare, err := s.ablage.Liste(ctx)
	if err != nil {
		return nil, err
	}

	// **Was zugesagt ist, ist nicht überzählig** (T2, #1467). Die Definition
	// bleibt dieselbe — „mehr als ich behalten will", je Zeile und nicht über
	// eine Drucksumme; die Zusage **zieht ab**, sie stellt nicht um.
	festgehalten, zusagenfehler := s.festgehalten(ctx)
	if zusagenfehler != nil {
		return nil, zusagenfehler
	}

	dubletten := []Dublettenzeile{}

	for _, exemplar := range exemplare {
		uebrig := exemplar.Menge - behalten - festgehalten[exemplar.ID]
		if !exemplar.Besitzt() || uebrig <= 0 {
			continue
		}

		dubletten = append(dubletten, Dublettenzeile{
			Exemplar: exemplar, Ueberzaehlig: uebrig,
		})
	}

	return dubletten, nil
}

// festgehalten sagt, was je Exemplar einem Handel zugesagt ist.
//
// **Ohne Naht ist es nichts**, und das ist keine Annahme, sondern der Stand:
// wer das Handeln nicht verdrahtet hat, hat keine Zusagen.
func (s *Sammlungssichten) festgehalten(ctx context.Context) (map[int64]int, error) {
	if s.reserviert == nil {
		return map[int64]int{}, nil
	}

	return s.reserviert.ReserviertJeExemplar(ctx)
}

// Wuensche gibt die Zeilen, die kein Besitz sind.
func (s *Sammlungssichten) Wuensche(ctx context.Context) ([]*domain.Exemplar, error) {
	exemplare, err := s.ablage.Liste(ctx)
	if err != nil {
		return nil, err
	}

	wuensche := []*domain.Exemplar{}

	for _, exemplar := range exemplare {
		if !exemplar.Besitzt() {
			wuensche = append(wuensche, exemplar)
		}
	}

	return wuensche, nil
}
