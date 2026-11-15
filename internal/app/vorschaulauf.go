package app

import (
	"context"
	"sync"
)

// abloesenderLauf hält den zuletzt angeforderten Lauf und bricht den vorigen
// ab — **ein Vorgang je Bezugsobjekt** (§7, Regel 4).
//
// # Die Sonderregel aus AGENTS.md §7
//
// **Eine Kartenvorschau wird abgebrochen, nicht abgewartet.** Wer während des
// Tippens vier Änderungen macht, stößt vier Läufe an. Ohne Abbruch laufen alle
// vier zu Ende und streiten dabei um Rechenzeit, Paketzugriffe und
// Schriftdateien — der neueste, auf den jemand wartet, am längsten.
//
// # Warum hier und nicht in `render`
//
// Der teuerste Teil einer Vorschau ist nicht das Rastern, sondern das
// Auswerten der Feldskripte: über die 190 Stile von `M15-Magic-Pack` sind das
// 5.950 Feldstile. Ein Abbruch, der erst beim Rasterer ansetzt, ließe genau
// diesen Teil weiterlaufen.
//
// Das ist auch der Grund, warum `render.Vorschau` entfallen ist. Es hat diese
// Aufgabe seit Phase 4 getragen und **nie einen Produktivaufrufer gehabt** —
// ein gebauter Weg, der sich nur selbst prüft
// (#294).
//
// # Und seit dem 2026-08-29 auch für die Packarten
//
// Sie rechnen über **jede** Karte des Sets: gemessen 3,3 s im Fenster bei 107
// Karten. Wer das Set wechselt, während sie laufen, ließ den Lauf bis dahin zu
// Ende rechnen — für eine Antwort, die niemand mehr einträgt
// (#1288).
//
// **Der Typ heißt deshalb nicht mehr `vorschaulauf`.** Ein Name, der einen
// von zwei Aufrufern nennt, führt den zweiten Leser in die Irre.
//
// # Warum das keine fachliche Entscheidung ist
//
// §2 sagt, `internal/app` enthält keine fachlichen Entscheidungen. „Ein Vorgang
// je Bezugsobjekt" ist keine: es ist eine Aussage über die **Naht** zum
// Fenster, und §7 führt sie unter den Regeln für die schreibende Seite. Wo der
// Auftrag entsteht, gehört auch sein Abbruch hin.
type abloesenderLauf struct {
	mu      sync.Mutex
	nummer  uint64
	abbruch context.CancelFunc
}

// beginne meldet einen neuen Lauf an und bricht den vorigen ab.
//
// Der zurückgegebene Kontext gilt für **diesen** Lauf. Die Aufräumfunktion ist
// immer aufzurufen; sie gibt den eigenen Abbruch frei, **ohne** einen inzwischen
// begonnenen späteren Lauf zu treffen. Ohne diese Unterscheidung brächte der
// zuerst fertige Aufrufer den nächsten um, kaum dass dieser angefangen hat —
// und die Vorschau bliebe leer, obwohl niemand etwas falsch gemacht hat.
func (v *abloesenderLauf) beginne(eltern context.Context) (lauf context.Context, fertig func()) {
	ctx, abbrechen := context.WithCancel(eltern)

	v.mu.Lock()

	if v.abbruch != nil {
		v.abbruch()
	}

	v.nummer++
	meiner := v.nummer
	v.abbruch = abbrechen

	v.mu.Unlock()

	return ctx, func() {
		v.mu.Lock()

		if v.nummer == meiner {
			v.abbruch = nil
		}

		v.mu.Unlock()

		// **Außerhalb der Sperre.** `abbrechen` ist zwar harmlos, aber ein
		// Aufruf fremden Codes unter einer gehaltenen Sperre ist die Form, aus
		// der Verklemmungen entstehen — auch wenn sie es heute nicht tut.
		abbrechen()
	}
}
