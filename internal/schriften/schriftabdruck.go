package schriften

// **Der Abdruck einer Schrift** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wofür es ihn gibt
//
// **Damit ein gerastertes Kartenbild merkt, wenn sich eine Schrift ändert.**
// Ein Zeichenplan trägt **Namen** — „MPlantin" —, keine Datei; wer nur ihn
// abdruckt, hat einen Schlüssel, der eine neu installierte Schriftfassung
// nicht bemerkt. Der Text sähe danach anders aus, und der Spieltisch zeigte
// das Bild von vorhin.
//
// # Warum der Abdruck über die **Datei** geht und nicht über den Namen
//
// **Weil der Name die Frage ist und die Datei die Antwort.** Welche Datei das
// Betriebssystem für „MPlantin" findet, hängt davon ab, was installiert ist
// und welche Pakete daliegen — und genau diese Antwort gehört in den
// Schlüssel.
//
// # Und warum ein Fehlschlag ein Wert ist
//
// **Weil eine fehlende Schrift ein Befund ist und kein Abbruch.** Der Rasterer
// zeichnet dann mit dem Ersatz und sagt es; dieses Bild soll wiedererkannt
// werden. Taucht die Schrift auf, ändert sich der Abdruck — und die Karte wird
// neu gezeichnet.

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

// Abdruck gibt den Abdruck der Schriftdatei zu einem Namen.
//
// **Aufrecht und fett getrennt**, denn beide werden gesucht, beide können
// verschiedene Dateien sein, und beide zeichnen.
func (s *Suche) Abdruck(name string, fett bool) string {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	schluessel := name
	if fett {
		schluessel += fettzusatz
	}

	if abdruck, da := s.abdruecke()[schluessel]; da {
		return abdruck
	}

	abdruck := s.lies(name, fett)
	s.abdruecke()[schluessel] = abdruck

	return abdruck
}

// lies sucht die Datei und druckt ihre Bytes ab — **unter dem Schloss**.
func (s *Suche) lies(name string, fett bool) string {
	ort, datei, err := s.schnitt(name, fett)
	if err != nil {
		return schriftFehlt(err.Error())
	}

	roh, err := schriftLesen(filepath.Join(ort, datei))
	if err != nil {
		return schriftFehlt(err.Error())
	}

	summe := sha256.Sum256(roh)

	return hex.EncodeToString(summe[:])
}

// abdruecke gibt den Speicher — **unter dem Schloss**, und legt ihn an.
func (s *Suche) abdruecke() map[string]string {
	if s.gemerkteAbdruecke == nil {
		s.gemerkteAbdruecke = map[string]string{}
	}

	return s.gemerkteAbdruecke
}

// schriftFehlt gibt den Wert für eine Schrift, die es nicht gibt.
//
// **Mit dem Grund darin**, damit „nicht gefunden" und „nicht lesbar" zwei
// verschiedene Schlüssel geben.
func schriftFehlt(grund string) string {
	summe := sha256.Sum256([]byte(grund))

	return "fehlt:" + hex.EncodeToString(summe[:])
}
