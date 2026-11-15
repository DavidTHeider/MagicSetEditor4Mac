package msescript

import (
	"regexp"
	"sync"
	"time"
)

// **Übersetzte Muster werden gemerkt** —
// #2006.
//
// # Warum
//
// Gemessen am 2026-10-03 an der Ausfuhr von `Black.mse-set`: 45 von 56
// Sekunden der Kartenskripte im Hauptfaden gingen in `regexp.Compile`
// (CPU-Profil, unter Last). `replace@(match: …)` und seine Geschwister
// bekommen ihr Muster als Zeichenkette, und die Übersetzung lief bei jedem
// Aufruf von vorn. Der Test mit der Begründung und der Fundstelle in MSE steht
// in `mustermerken_intern_test.go`.
//
// # Was gemerkt wird, und wie lange
//
// Die Quelle ist der Schlüssel, die Übersetzung der Wert. **Eine Übersetzung
// ist unveränderlich** — [Muster] hat keine Methode, die es schreibt, und
// Gos `regexp` wie `regexp2` sind für den gleichzeitigen Gebrauch gemacht.
//
// **Die Frist gehört zur Rückfalllinie** ([musterFrist]): wer sie ändert,
// bekommt kein Muster mit der alten. Ein RE2-Muster hängt an ihr nicht.
//
// **Der Speicher ist begrenzt.** Ein Skript kann Muster aus Kartennamen bauen
// (`regex_escape(card.name)`), und ein Dienst läuft lange. Ist er voll, wird er
// ganz geleert — einfacher als eine Verdrängungsordnung, und er füllt sich in
// einer Ausfuhr mit den wenigen Dutzend Mustern des Spiels sofort wieder.
// **Ein Fehler wird nicht gemerkt**: ein unlesbares Muster kommt jedes Mal als
// Fehler.

// musterspeichergrenze ist die Zahl der Übersetzungen, ab der der Speicher
// geleert wird — je Art.
const musterspeichergrenze = 4096

// musterablage hält die Übersetzungen zweier Arten: die [Muster] der
// MSE-Dialekte und die festen RE2-Ausdrücke der Auszeichnungsfunktionen.
type musterablage struct {
	sperre sync.Mutex

	muster     map[string]*Muster
	ausdruecke map[string]*regexp.Regexp
}

// musterspeicher ist der eine Speicher des Pakets.
var musterspeicher musterablage

// hole gibt die Übersetzung zu dieser Quelle — **wenn sie zu dieser Frist
// passt**.
func (a *musterablage) hole(quelle string, frist time.Duration) (*Muster, bool) {
	a.sperre.Lock()
	defer a.sperre.Unlock()

	gemerkt, da := a.muster[quelle]
	if !da || (gemerkt.breit != nil && gemerkt.frist != frist) {
		return nil, false
	}

	return gemerkt, true
}

// merke legt eine Übersetzung ab; ein voller Speicher wird vorher geleert.
func (a *musterablage) merke(quelle string, muster *Muster) {
	a.sperre.Lock()
	defer a.sperre.Unlock()

	if a.muster == nil || len(a.muster) >= musterspeichergrenze {
		a.muster = make(map[string]*Muster)
	}

	a.muster[quelle] = muster
}

// leeren vergisst alles — für die Tests.
func (a *musterablage) leeren() {
	a.sperre.Lock()
	defer a.sperre.Unlock()

	a.muster, a.ausdruecke = nil, nil
}

// groesse zählt die Übersetzungen beider Arten.
func (a *musterablage) groesse() int {
	a.sperre.Lock()
	defer a.sperre.Unlock()

	return len(a.muster) + len(a.ausdruecke)
}

// gemerkterAusdruck übersetzt einen RE2-Ausdruck einmal.
//
// **`MustCompile` und kein Fehlerzweig**, wie an den Aufrufern: der Name geht
// dort durch `QuoteMeta`, und was herauskommt, ist immer ein gültiges Muster.
func gemerkterAusdruck(ausdruck string) *regexp.Regexp {
	musterspeicher.sperre.Lock()
	defer musterspeicher.sperre.Unlock()

	if gemerkt, da := musterspeicher.ausdruecke[ausdruck]; da {
		return gemerkt
	}

	if musterspeicher.ausdruecke == nil || len(musterspeicher.ausdruecke) >= musterspeichergrenze {
		musterspeicher.ausdruecke = make(map[string]*regexp.Regexp)
	}

	uebersetzt := regexp.MustCompile(ausdruck)
	musterspeicher.ausdruecke[ausdruck] = uebersetzt

	return uebersetzt
}
