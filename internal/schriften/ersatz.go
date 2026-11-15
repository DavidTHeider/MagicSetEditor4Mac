package schriften

import (
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Ersatz gibt eine Schrift, die dieses Zeichen führt.
//
// # Warum es das gibt
//
// **MSE weicht aus, weil wxWidgets es tut.** Ein Stil nennt eine Schrift;
// fehlt darin ein Zeichen, sucht die Plattform eine andere — unter macOS Core
// Text. Dieses Programm rastert selbst und hatte diesen Weg nicht: über die
// 23.168 kanonischen Karten fehlte **ein** Zeichen, `ǵ` (U+01F5), und es
// fehlte 159-mal
// (#515).
//
// # Welche Schrift genommen wird
//
// **Die erste, die es führt** — in der Reihenfolge der Orte, und die stehen
// nach Vorrang: was jemand selbst installiert hat, kommt zuerst (#140).
//
// **Das ist nicht dieselbe Wahl, die Core Text träfe.** Ein ausgewichenes
// Zeichen sieht deshalb anders aus als bei MSE auf demselben Rechner. Es steht
// aber da, und das ist der Unterschied, um den es geht (§8) — ein fehlendes
// Zeichen ist eine Lücke im Kartennamen.
//
// # Was es kostet
//
// **Jede Schriftdatei einmal öffnen, und das nur beim ersten Fehlschlag.**
// Danach steht die Antwort für die Sitzung; ein Zeichen, das keine Schrift
// führt, wird ebenfalls gemerkt, sonst liefe die Suche bei jedem Vorkommen neu.
//
// # Unter dem Schloss, wie jeder Speicher der Suche
//
// **Eine Starthand legt sieben Karten auf einmal auf den Tisch**, und ihre
// Bilder werden gleichzeitig gezeichnet. Ohne Schloss endete das am
// 2026-09-25 in „fatal error: concurrent map writes" — zwei von sechs Läufen
// der WebKit-Probe `play-beginn` verloren den Server. **Gesucht wird ohne
// Schloss**: das Durchgehen der Orte liest Dateien, und alle anderen Fragen
// an die Suche warteten sonst darauf. Zwei gleichzeitige Fehlschläge für
// dasselbe Zeichen suchen dann beide; das Ergebnis ist dasselbe.
func (s *Suche) Ersatz(zeichen rune) (render.Glyphenschrift, bool) {
	s.schloss.RLock()
	gemerkt, da := s.ersatz[zeichen]
	s.schloss.RUnlock()

	if da {
		return gemerkt, gemerkt != nil
	}

	gefunden := s.sucheZeichen(zeichen)

	s.schloss.Lock()
	defer s.schloss.Unlock()

	if s.ersatz == nil {
		s.ersatz = map[rune]render.Glyphenschrift{}
	}

	s.ersatz[zeichen] = gefunden

	return gefunden, gefunden != nil
}

// sucheZeichen geht die Orte durch und gibt die erste Schrift mit diesem
// Zeichen.
func (s *Suche) sucheZeichen(zeichen rune) render.Glyphenschrift {
	for _, ort := range s.orte {
		for _, datei := range schriftdateien(ort) {
			if schrift := gelesenMitZeichen(filepath.Join(ort, datei), zeichen); schrift != nil {
				return schrift
			}
		}
	}

	return nil
}

// gelesenMitZeichen öffnet eine Schriftdatei und gibt sie, wenn sie das
// Zeichen führt.
func gelesenMitZeichen(pfad string, zeichen rune) render.Glyphenschrift {
	// Der Pfad kommt aus der Verzeichnisliste eines Ortes dieser Suche, nicht
	// aus einer Eingabe — dieselbe Lage wie in [Suche.Suche].
	roh, err := schriftLesen(pfad)
	if err != nil {
		return nil
	}

	schrift, err := render.SchriftAusDatei(roh)
	if err != nil {
		return nil
	}

	if _, _, da := schrift.Maske(zeichen, 12); !da {
		return nil
	}

	return schrift
}
