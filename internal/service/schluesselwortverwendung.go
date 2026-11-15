// **Welche Schlüsselwörter eine Karte benutzt** —
// #1354.
//
// # Was portiert ist
//
// `Card::keyword_usage` (`src/data/card.hpp` Zeile 59, Fassung `f9d9356` vom
// 2021-07-06) ist ein `vector<pair<const Value*, const Keyword*>>` — je
// Verwendung ein Paar aus **Feld** und **Schlüsselwort**, in der Reihenfolge
// des Eintragens.
//
// Drei Stellen fassen ihn an, und alle drei stehen hier:
//
//   - `expand_keyword` trägt ein, **nachdem** der Erinnerungstext
//     zusammengesetzt ist (`src/data/keyword.cpp` Zeilen 749–751);
//   - `remove_from_stats` räumt vor jedem Feld dessen alte Einträge weg
//     (ebd. Zeilen 766–773);
//   - `SCRIPT_FUNCTION(keyword_usage)` fügt die Namen mit `", "` zusammen
//     (`src/script/functions/basic.cpp` Zeilen 695–717).
//
// # Warum das Feld mitgeführt wird
//
// Weil `remove_from_stats` es braucht: eine Karte hat mehrere Textfelder, und
// jedes wird für sich neu gerechnet. Ohne den Feldnamen räumte ein Neuzeichnen
// von `rule text` auch weg, was `rule text 2` beigetragen hat.

package service

import (
	"context"
	"strings"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Schluesselwortverwendung hält fest, welche Schlüsselwörter eine Karte
// benutzt.
//
// **Der Nullwert ist brauchbar** — eine Karte ohne Verwendung.
//
// # Warum eine Sperre
//
// **Eine Kartenvorschau wird abgebrochen, nicht abgewartet** (§7): wer tippt,
// stößt mehrere Läufe an, und der vorige rechnet noch, während der nächste
// beginnt. Beide setzen Schlüsselwörter ein, und beide vermerken.
//
// Ohne sie meldete `go test -race ./internal/...` am 2026-09-03 zwei bis drei
// Rennen je Lauf, alle in `Vermerke` — ein `append` auf einen Slice, den zwei
// Goroutinen zugleich anfassen. Der Fehler war **nicht neu**: derselbe Lauf
// über den Stand davor meldete ihn ebenso, sobald genug parallel lief. Ein
// grüner Lauf davor war Glück, kein Beleg
// (#1325).
//
// **Sie deckt auch `laufend` ab.** MSEs `value_being_updated()` ist ein
// dynamisches Argument je Auswertung; hier steht es an der geteilten Gestalt,
// und zwei Läufe an verschiedenen Feldern überschrieben es gegenseitig.
type Schluesselwortverwendung struct {
	sperre sync.Mutex

	eintraege []schluesselwortvermerk

	// laufend ist das Feld, dessen Skript gerade rechnet — MSEs
	// `value_being_updated()`, ein dynamisches Argument
	// (`src/data/keyword.cpp` Zeile 278, Fassung `f9d9356` vom 2021-07-06).
	laufend string
}

// schluesselwortvermerk ist MSEs `pair<const Value*, const Keyword*>`.
type schluesselwortvermerk struct {
	feld string
	wort string
}

// Vermerke trägt eine Verwendung ein.
func (v *Schluesselwortverwendung) Vermerke(feld, wort string) {
	v.sperre.Lock()
	defer v.sperre.Unlock()

	v.eintraege = append(v.eintraege, schluesselwortvermerk{feld: feld, wort: wort})
}

// Raeume nimmt heraus, was dieses Feld beigetragen hat.
//
// **Vor dem Einsetzen, nicht danach** — MSEs `KeywordDatabase::expand` beginnt
// damit (Zeile 779). Ohne das wüchse die Liste bei jedem Neuzeichnen.
func (v *Schluesselwortverwendung) Raeume(feld string) {
	v.sperre.Lock()
	defer v.sperre.Unlock()

	behalten := v.eintraege[:0]

	for _, eintrag := range v.eintraege {
		if eintrag.feld != feld {
			behalten = append(behalten, eintrag)
		}
	}

	v.eintraege = behalten
}

// Liste fügt die Namen zusammen — `keyword_usage`.
//
// **Mit `einmalig` fällt weg, was schon dastand.** MSE vergleicht dabei den
// **Zeiger** auf das Schlüsselwort und nicht seinen Namen; hier gibt es keine
// Zeiger, und zwei Schlüsselwörter mit demselben Namen wären ohnehin dasselbe
// für jeden, der die Liste liest.
func (v *Schluesselwortverwendung) Liste(einmalig bool) string {
	if v == nil {
		return ""
	}

	v.sperre.Lock()
	defer v.sperre.Unlock()

	namen := make([]string, 0, len(v.eintraege))

	for _, eintrag := range v.eintraege {
		if einmalig && enthaelt(namen, eintrag.wort) {
			continue
		}

		namen = append(namen, eintrag.wort)
	}

	return strings.Join(namen, ", ")
}

// BeiFeld sagt an, dass jetzt das Skript dieses Feldes rechnet.
//
// **Und räumt weg, was das Feld zuletzt beigetragen hat** — MSEs
// `remove_from_stats` beim Eintritt in `KeywordDatabase::expand` (Zeile 779).
// Ohne das wüchse die Liste bei jedem Neuzeichnen.
func (v *Schluesselwortverwendung) BeiFeld(feld string) {
	v.sperre.Lock()
	v.laufend = feld
	v.sperre.Unlock()

	v.Raeume(feld)
}

// vermerkeAmLaufendenFeld trägt eine Verwendung für das laufende Feld ein.
func (v *Schluesselwortverwendung) vermerkeAmLaufendenFeld(wort string) {
	v.sperre.Lock()
	feld := v.laufend
	v.sperre.Unlock()

	v.Vermerke(feld, wort)
}

// **`enthaelt` steht in `bildpassung.go`** — eine zweite Fassung daneben
// liefe mit jener auseinander (§2).

// schluesselwortverwendungsfrage baut die Funktion `keyword_usage`.
//
// Portiert aus `SCRIPT_FUNCTION(keyword_usage)`
// (`src/script/functions/basic.cpp` Zeilen 695–717, Fassung `f9d9356` vom
// 2021-07-06).
//
// **`card:` wird nicht gelesen.** MSE nimmt es als Pflichtargument
// (`SCRIPT_PARAM_C(CardP, card)`) und liest daran die Karte, deren Vermerk
// gefragt ist. Hier hängt der Vermerk an der **Umgebung** der Karte, die
// gerade rechnet — ein `card:`-Argument könnte auf keine andere zeigen. Die
// Dimension `keywords` ruft ohnehin `keyword_usage(unique: true)` ohne Karte.
func schluesselwortverwendungsfrage(vermerk *Schluesselwortverwendung) msescript.Wert {
	return msescript.NeueEingebaute(func(_ context.Context, ruf msescript.Ruf) (
		msescript.Wert, error,
	) {
		einmalig := false

		if wert, da := ruf.Benannter("unique"); da {
			gesetzt, brauchbar := wert.AlsWahrheit()
			einmalig = brauchbar && gesetzt
		}

		return msescript.Text(vermerk.Liste(einmalig)), nil
	})
}
