package service

// **Wie lange ein Ruf dauert, als Verteilung** — OPS-05.AK-01
// (#1722),
// neben dem Messwerk aus O9.
//
// # Eimer und keine Einzelwerte
//
// **Sechs Grenzen in Millisekunden und ein offener Eimer darüber**, je
// Fläche und Art. Die Zahl der Reihen ist damit fest — vier Flächen mal zwei
// Arten —, gleich wie viele Rufe es gibt; ein Speicher, der mit jedem Ruf
// wächst, wäre der unbegrenzte Datenbestand, den AK-03 ausschließt.
//
// # Was hier nicht gemessen wird
//
// **Die Bindungen des Fensters.** Sie haben keinen gemeinsamen Eingang, an
// dem eine Uhr hängen könnte; gemessen werden die zwei Server, deren Rufe
// von außen kommen — Dienst und Werkzeugserver —, und das ist die Frage, die
// das Issue stellt: „bessere Einblicke in REST/MCP".

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Laufzeiteimer sind die oberen Grenzen der Eimer in Millisekunden — der
// letzte Eimer darüber ist offen.
var Laufzeiteimer = []int64{1, 5, 25, 100, 500, 2500}

// Die Arten eines Rufs.
const (
	ArtLesend     = "lesend"
	ArtSchreibend = "schreibend"
)

// Laufzeitpunkt ist ein gemessener Ruf.
type Laufzeitpunkt struct {
	Flaeche domain.Dienstmodus

	// Weg ist das Muster des Wegs oder der Name des Werkzeugs — **nie** der
	// rohe Pfad: der trüge Kennungen und damit Nutzinhalte (AK-02).
	Weg        string
	Schreibend bool
	Dauer      time.Duration
	Ausgang    domain.Rufausgang
}

// Laufzeitreihe ist die Verteilung für eine Fläche und Art.
type Laufzeitreihe struct {
	Flaeche     domain.Dienstmodus `json:"flaeche"`
	Art         string             `json:"art"`
	Anzahl      int64              `json:"anzahl"`
	Abgewiesen  int64              `json:"abgewiesen"`
	Gescheitert int64              `json:"gescheitert"`
	SummeMs     float64            `json:"summeMs"`

	// Eimer zählt je Grenze aus [Laufzeiteimer] und einen offenen darüber.
	Eimer []int64 `json:"eimer"`
}

// Laufzeitreihengrenze ist die Zahl der Reihen, die es geben kann.
var Laufzeitreihengrenze = len(domain.Dienstmodi()) * 2

type laufzeitschluessel struct {
	flaeche domain.Dienstmodus
	art     string
}

var laufzeitwerk = struct {
	sperre sync.Mutex
	stand  map[laufzeitschluessel]*Laufzeitreihe
}{stand: map[laufzeitschluessel]*Laufzeitreihe{}}

// ErfasseLaufzeit trägt einen Ruf ein — eine unbekannte Fläche nicht.
func ErfasseLaufzeit(punkt Laufzeitpunkt) {
	if !ausDerAufzaehlung(domain.Dienstmodi(), punkt.Flaeche) {
		return
	}

	art := ArtLesend
	if punkt.Schreibend {
		art = ArtSchreibend
	}

	laufzeitwerk.sperre.Lock()
	defer laufzeitwerk.sperre.Unlock()

	schluessel := laufzeitschluessel{flaeche: punkt.Flaeche, art: art}

	reihe, da := laufzeitwerk.stand[schluessel]
	if !da {
		reihe = &Laufzeitreihe{Flaeche: punkt.Flaeche, Art: art, Eimer: make([]int64, len(Laufzeiteimer)+1)}
		laufzeitwerk.stand[schluessel] = reihe
	}

	ms := float64(punkt.Dauer) / float64(time.Millisecond)
	reihe.Anzahl++
	reihe.SummeMs += ms
	reihe.Eimer[eimerFuer(ms)]++

	switch punkt.Ausgang {
	case domain.RufAbgewiesen:
		reihe.Abgewiesen++
	case domain.RufGescheitert:
		reihe.Gescheitert++
	case domain.RufGelungen:
	}
}

// ErfasseRuf trägt einen Ruf in die Verteilung und — wenn sie an ist — in die
// Spur ein, mit der Vorgangskennung aus dem Kontext.
func ErfasseRuf(ctx context.Context, punkt Laufzeitpunkt) {
	ErfasseLaufzeit(punkt)
	spure(VorgangAus(ctx), punkt)
}

// eimerFuer gibt den Eimer einer Dauer — den ersten, dessen Grenze sie nicht
// übersteigt, sonst den offenen.
func eimerFuer(ms float64) int {
	for stelle, grenze := range Laufzeiteimer {
		if ms <= float64(grenze) {
			return stelle
		}
	}

	return len(Laufzeiteimer)
}

// Laufzeitstand gibt die Reihen, geordnet nach Fläche und Art — Kopien.
func Laufzeitstand() []Laufzeitreihe {
	laufzeitwerk.sperre.Lock()
	defer laufzeitwerk.sperre.Unlock()

	reihen := make([]Laufzeitreihe, 0, len(laufzeitwerk.stand))
	for _, reihe := range laufzeitwerk.stand {
		kopie := *reihe
		kopie.Eimer = append([]int64(nil), reihe.Eimer...)
		reihen = append(reihen, kopie)
	}

	sort.Slice(reihen, func(links, rechts int) bool {
		if reihen[links].Flaeche != reihen[rechts].Flaeche {
			return reihen[links].Flaeche < reihen[rechts].Flaeche
		}

		return reihen[links].Art < reihen[rechts].Art
	})

	return reihen
}

// LaufzeitwerkZuruecksetzen leert Verteilung und Spur und schaltet die Spur
// aus — für Läufe.
func LaufzeitwerkZuruecksetzen() {
	laufzeitwerk.sperre.Lock()
	laufzeitwerk.stand = map[laufzeitschluessel]*Laufzeitreihe{}
	laufzeitwerk.sperre.Unlock()

	spurZuruecksetzen()
}
