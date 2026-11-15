package mcp

import (
	"sync"
	"time"
)

// Die Grenze für falsche Marken — **dieselbe wie beim Dienst** (`rest.grenze`,
// `rest.fenster`): sechzig Fehlversuche in der Minute. SEC-12
// (#2141).
const (
	fehlversuchsgrenze  = 60
	fehlversuchsfenster = time.Minute
)

// fehlversuche begrenzt das Raten der Marke.
//
// **Nur Fehlversuche zählen.** Wer die Marke hat, arbeitet im Budget seiner
// Werkzeuge (`MitKontingent`) weiter, auch wenn ein anderer gerade rät — ein
// gemeinsames Budget ließe einen lokalen Prozess, der nie eine Marke hatte,
// den rechtmäßigen Aufrufer aussperren. Dieselbe Trennung wie beim Dienst
// (Zweitblick über Phase U).
//
// **Ein festes Fenster**, wie `rest.Drossel`: gegen das Raten einer Marke von
// 64 Zeichen ist es so wirksam wie ein Eimer und in einem Satz zu erklären.
// Die Regel steht hier ein zweites Mal, weil dieses Paket ein Blatt ist (§2).
type fehlversuche struct {
	jetzt func() time.Time

	schloss  sync.Mutex
	seit     time.Time
	gezaehlt int
}

// zaehle trägt einen Fehlversuch ein und sagt, ob er noch ins Fenster passt.
func (f *fehlversuche) zaehle() bool {
	f.schloss.Lock()
	defer f.schloss.Unlock()

	if jetzt := f.jetzt(); jetzt.Sub(f.seit) >= fehlversuchsfenster {
		f.seit = jetzt
		f.gezaehlt = 0
	}

	if f.gezaehlt >= fehlversuchsgrenze {
		return false
	}

	f.gezaehlt++

	return true
}

// MitUhr setzt die Uhr, nach der falsche Marken gezählt werden.
//
// **Eingesetzt und nicht genommen** (§16): ein Lauf über ein Zeitfenster, der
// auf die echte Uhr wartet, flackert.
func (s *Server) MitUhr(jetzt func() time.Time) *Server {
	s.uhr = jetzt

	return s
}
