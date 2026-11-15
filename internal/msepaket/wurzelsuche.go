package msepaket

import (
	"slices"
	"strings"
	"sync"
)

// Wurzelsuche merkt das Ergebnis von [Paketwurzeln] und prüft es bei jedem
// Ruf — PERF-02 (#2169).
//
// # Warum
//
// Die Flächen ohne Fenster fragen die Wurzeln **je Ruf**, und im vollen Pack
// heißt ein Lauf: Ordner mit rund tausend Einträgen lesen, zweimal. Ein
// warmer Ruf `lesen.skript` über `mse-mcp` brauchte dafür den größten Teil
// seiner Zeit (`docs/LEISTUNGSSCHNITTE.md`).
//
// # Wann das Gemerkte gilt
//
// **Solange jeder Ordner, den der Lauf gelesen hat, denselben Stempel
// trägt** — Änderungszeit, Statuszeit, Rechte und Knoten ([ortsstempel]) —,
// und jede Wurzel denselben oder dasselbe Fehlen. Das Ergebnis hängt allein
// an den Einträgen dieser Ordner und daran, dass sie sich lesen lassen; die
// Statuszeit wechselt bei beidem, auch wenn jemand die Änderungszeit
// zurücksetzt. Eine Frist gibt es nicht: was sich ändert, gilt beim nächsten
// Ruf.
//
// **Was nicht gemerkt wird**: ein Lauf, der einen Ordner nicht lesen konnte.
// Wird der lesbar, ändert sich seine Änderungszeit nicht, und das Gemerkte
// bliebe falsch — ein solcher Lauf wiederholt sich deshalb bei jedem Ruf, wie
// vorher. **Was sie nicht sieht**: den Inhalt eines Pakets. Der war nie ihre
// Frage; welche Datei darin steht, liest [Paketleser] je Ruf.
type Wurzelsuche struct {
	sperre  sync.Mutex
	gemerkt map[string]wurzelstand

	// laeufe zählt die Gänge über den Baum — für die Prüfung.
	laeufe int
}

// wurzelstand ist ein gemerktes Ergebnis und die Stempel, an denen es hängt.
type wurzelstand struct {
	wurzeln []string
	stempel map[string]ortsstempel
}

// grenzeWurzelsuche ist, wie viele Fragen gemerkt werden. Eine Fläche fragt
// immer dieselben drei Orte; an der Grenze wird ganz geräumt, wie beim
// Zeichenvorrat.
const grenzeWurzelsuche = 16

// NeueWurzelsuche gibt eine leere Wurzelsuche.
func NeueWurzelsuche() *Wurzelsuche {
	return &Wurzelsuche{gemerkt: map[string]wurzelstand{}}
}

// Paketwurzeln gibt dasselbe wie [Paketwurzeln] — gemerkt, solange es gilt.
//
// **Unter der Sperre**: zwei erste Rufe zugleich gingen sonst zweimal über
// den Baum.
func (s *Wurzelsuche) Paketwurzeln(wurzeln ...string) []string {
	schluessel := strings.Join(echteWurzeln(wurzeln), "\x00")

	s.sperre.Lock()
	defer s.sperre.Unlock()

	if stand, da := s.gemerkt[schluessel]; da && stand.gilt() {
		return slices.Clone(stand.wurzeln)
	}

	s.laeufe++

	lauf := &wurzellauf{stempel: map[string]ortsstempel{}, sauber: true}
	gefunden := paketwurzelnMit(lauf, wurzeln)

	if !lauf.sauber {
		delete(s.gemerkt, schluessel)

		return gefunden
	}

	if len(s.gemerkt) >= grenzeWurzelsuche {
		clear(s.gemerkt)
	}

	s.gemerkt[schluessel] = wurzelstand{wurzeln: gefunden, stempel: lauf.stempel}

	return slices.Clone(gefunden)
}

// gilt sagt, ob jeder vermerkte Ort noch denselben Stempel trägt.
func (w wurzelstand) gilt() bool {
	for pfad, stempel := range w.stempel {
		if stempelVon(pfad) != stempel {
			return false
		}
	}

	return true
}

// wurzellauf vermerkt, was ein Gang über den Baum gelesen hat.
//
// **Ein nil-Lauf vermerkt nichts** — so geht [Paketwurzeln] denselben Weg
// ohne Merker.
type wurzellauf struct {
	stempel map[string]ortsstempel
	sauber  bool
}

func (l *wurzellauf) vermerke(pfad string) {
	if l != nil {
		l.stempel[pfad] = stempelVon(pfad)
	}
}

func (l *wurzellauf) scheitert() {
	if l != nil {
		l.sauber = false
	}
}
