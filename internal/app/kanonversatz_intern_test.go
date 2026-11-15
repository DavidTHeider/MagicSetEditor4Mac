//go:build fremdpakete

// Die Versatzprobe und der Abstand je Vorlage
// (#522).
//
// Zwei Fragen, die derselbe Mittelwert verdeckt:
//
//  1. **Malfehler oder Platzierungsfehler?** Ein Abstand von 27 kann heißen,
//     dass überall die falsche Farbe steht — oder dass alles um einen
//     Bildpunkt daneben liegt.
//  2. **Ist eine Vorlagenklasse als Klasse weiter weg?** Das ist die zweite
//     Hälfte von #498:
//     `Jaya Ballard` ist ein Planeswalker und steht fast doppelt so weit weg
//     wie die Mitte.
package app

import (
	"image"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Versatzreichweite ist der Suchbereich der Versatzprobe.
//
// **Zwei und nicht eins.** Bei eins läge jeder gefundene Versatz am Rand des
// Bereichs, und die Probe könnte nie sagen, ob sie fertig ist — sie meldete
// überall „mindestens 1". Mit zwei gibt es eine Stufe dazwischen, an der sich
// das ablesen lässt ([render.Versatz.AmRand]).
const Versatzreichweite = 2

// versatzstand zählt Versatzproben über viele Karten zusammen.
type versatzstand struct {
	karten int

	// mitVersatz ist die Zahl der Karten, bei denen ein Versatz den Abstand
	// überhaupt verkleinert hat.
	mitVersatz int

	// amRand zählt die Karten, deren bester Versatz auf dem Rand des
	// Suchbereichs lag — bei ihnen ist die Zahl eine Untergrenze (§10).
	amRand int

	// gewinne sind die Beträge, um die der Abstand fiel.
	gewinne []float64

	// stellen zählt, welcher Versatz wie oft gewonnen hat.
	stellen map[image.Point]int

	// beispiele nennt je Versatz ein paar Karten **beim Namen**
	// (#1029).
	//
	// **Ohne sie ist die Tabelle eine Zählung ohne Anfassbares.** „48 Karten
	// bei (0,−1)" sagt, dass es sie gibt; welche es sind, sagt erst der Name —
	// und erst am Namen lässt sich ansehen, was sie gemeinsam haben.
	beispiele map[image.Point][]string
}

// BeispieleJeVersatz ist, wie viele Namen je Versatz gemerkt werden.
//
// **Fünf**: genug, um eine Gemeinsamkeit zu sehen, wenig genug, dass die
// Ausgabe lesbar bleibt.
const BeispieleJeVersatz = 5

func neuerVersatzstand() *versatzstand {
	return &versatzstand{
		stellen:   map[image.Point]int{},
		beispiele: map[image.Point][]string{},
	}
}

func (s *versatzstand) nimm(versatz render.Versatz, name string) {
	stelle := image.Pt(versatz.X, versatz.Y)

	s.karten++
	s.stellen[stelle]++

	// **Nur die versetzten bekommen Namen.** (0,0) ist der Normalfall; fünf
	// Namen dafür wären fünf beliebige Karten.
	if stelle != (image.Point{}) && len(s.beispiele[stelle]) < BeispieleJeVersatz {
		s.beispiele[stelle] = append(s.beispiele[stelle], name)
	}

	if versatz.AmRand() {
		s.amRand++
	}

	if versatz.Gewinn() > 0 {
		s.mitVersatz++
		s.gewinne = append(s.gewinne, versatz.Gewinn())
	}
}

func (s *versatzstand) uebernimm(anderer *versatzstand) {
	s.karten += anderer.karten
	s.mitVersatz += anderer.mitVersatz
	s.amRand += anderer.amRand
	s.gewinne = append(s.gewinne, anderer.gewinne...)

	for stelle, anzahl := range anderer.stellen {
		s.stellen[stelle] += anzahl
	}

	for stelle, namen := range anderer.beispiele {
		fehlend := BeispieleJeVersatz - len(s.beispiele[stelle])
		if fehlend > 0 {
			s.beispiele[stelle] = append(s.beispiele[stelle],
				namen[:min(fehlend, len(namen))]...)
		}
	}
}

// melde schreibt die Versatztabelle in den Testlauf.
func (s *versatzstand) melde(t *testing.T, was string) {
	t.Helper()

	if s.karten == 0 {
		return
	}

	var summe float64
	for _, gewinn := range s.gewinne {
		summe += gewinn
	}

	mittlererGewinn := 0.0
	if len(s.gewinne) > 0 {
		mittlererGewinn = summe / float64(len(s.gewinne))
	}

	t.Logf("  %s — Versatzprobe über %d Karten, Reichweite ±%d:",
		was, s.karten, Versatzreichweite)
	t.Logf("      %d Karten gewinnen durch einen Versatz, im Mittel %.3f Abstand",
		s.mitVersatz, mittlererGewinn)

	if s.amRand > 0 {
		t.Logf("      %d davon am Rand des Suchbereichs — dort ist die Zahl "+
			"eine Untergrenze", s.amRand)
	}

	stellen := make([]image.Point, 0, len(s.stellen))
	for stelle := range s.stellen {
		stellen = append(stellen, stelle)
	}

	sort.Slice(stellen, func(i, j int) bool {
		return s.stellen[stellen[i]] > s.stellen[stellen[j]]
	})

	for _, stelle := range stellen {
		if namen := s.beispiele[stelle]; len(namen) > 0 {
			t.Logf("      (%+d,%+d) bei %d Karten — z. B. %s",
				stelle.X, stelle.Y, s.stellen[stelle], strings.Join(namen, ", "))

			continue
		}

		t.Logf("      (%+d,%+d) bei %d Karten", stelle.X, stelle.Y, s.stellen[stelle])
	}
}

// vorlagenstand sammelt Abstände je Vorlage.
type vorlagenstand map[string][]float64

func (v vorlagenstand) nimm(vorlage string, abstand float64) {
	v[vorlage] = append(v[vorlage], abstand)
}

// melde schreibt die Tabelle je Vorlage — **nach Mittelwert geordnet**, denn
// die Frage lautet, welche Klasse am weitesten weg ist.
//
// **Vorlagen unter [MindestensJeVorlage] Karten stehen zusammengefasst.** Ein
// Mittelwert über zwei Karten sagt nichts über eine Klasse, und in einer
// sortierten Liste stünde er ganz oben (§10).
func (v vorlagenstand) melde(t *testing.T) {
	t.Helper()

	if len(v) == 0 {
		return
	}

	namen := make([]string, 0, len(v))

	seltene, seltenSumme := 0, 0.0

	for name, abstaende := range v {
		if len(abstaende) < MindestensJeVorlage {
			seltene += len(abstaende)

			for _, a := range abstaende {
				seltenSumme += a
			}

			continue
		}

		namen = append(namen, name)
	}

	sort.Slice(namen, func(i, j int) bool {
		return mittel(v[namen[i]]) > mittel(v[namen[j]])
	})

	t.Logf("  Abstand je Vorlage (ab %d Karten):", MindestensJeVorlage)

	for _, name := range namen {
		abstaende := append([]float64(nil), v[name]...)
		sort.Float64s(abstaende)

		t.Logf("      %-34s %3d Karten  Mittel %7.3f  schlechteste %7.3f",
			name, len(abstaende), mittel(abstaende), abstaende[len(abstaende)-1])
	}

	if seltene > 0 {
		t.Logf("      %-34s %3d Karten  Mittel %7.3f", "(seltenere Vorlagen zusammen)",
			seltene, seltenSumme/float64(seltene))
	}
}

// MindestensJeVorlage ist die Zahl an Karten, ab der ein Mittelwert je Vorlage
// ausgewiesen wird.
const MindestensJeVorlage = 5

func mittel(werte []float64) float64 {
	if len(werte) == 0 {
		return 0
	}

	var summe float64
	for _, wert := range werte {
		summe += wert
	}

	return summe / float64(len(werte))
}

// vorlageDerKarte gibt den Namen des Stils, mit dem diese Karte gezeichnet wird.
//
// **Der eigene Stil der Karte, sonst der des Sets** — dieselbe Regel wie
// [service.Stilquelle.StilFuerKarte], nachgelesen an `src/data/set.cpp`,
// Zeilen 90–93, Fassung f9d9356 vom 2021-07-06.
//
// **Und derselbe Feldname**: `service.Kartenstilfeld`. Ihn hier als
// Zeichenkette zu wiederholen hieße, zwei Stellen zu haben, die dasselbe
// wissen sollen.
func vorlageDerKarte(karte KarteDTO, desSets string) string {
	for _, wert := range karte.Werte {
		if wert.Feld == service.Kartenstilfeld {
			if eigener := strings.TrimSpace(wert.Wert); eigener != "" {
				return eigener
			}
		}
	}

	return desSets
}

// gewaehlterStil gibt das Stilpaket, das dieses Set gewählt hat.
//
// **Als Rückfall und nicht als Regel**: über die kanonischen Sets trägt jede
// Karte ihren eigenen `stylesheet:` — in `Red` sind es 3.619-mal
// `m15-altered`, 50-mal `m15-mainframe-dfc`, 33-mal
// `m15-mainframe-planeswalker`. Der Rückfall greift dort nie; er steht da,
// weil ein Set ohne Angabe an der Karte möglich ist.
func gewaehlterStil(t *testing.T, anwendung *App, setID string) string {
	t.Helper()

	antwort := anwendung.Stile(t.Context(), setID)
	if antwort.Fehler != nil {
		return "(Stil des Sets unbekannt)"
	}

	for _, stil := range antwort.Stile {
		if stil.Gewaehlt {
			return stil.Paket
		}
	}

	return "(Set ohne gewählten Stil)"
}
