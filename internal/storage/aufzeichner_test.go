package storage

// **Ein Treiber, der mitschreibt** — Hinweis 12 des Nachlaufs zum Zweitblick
// über Phase 5.2 (#1435).
//
// # Wofür er da ist
//
// Manche Zusagen sind Aussagen über die **Reihenfolge**, nicht über ein
// Ergebnis: „der Vergleich liegt im Schreibgeschäft" ist eine. Ein Test gegen
// SQLite sieht sie nicht — er sieht das Ergebnis, und das ist dasselbe, ob der
// Vergleich davor oder darin lag. Erst wenn ein zweiter Schreiber dazwischen
// kommt, gehen die beiden auseinander, und den kann kein Aufrufer von außen
// hineinlegen.
//
// **Also wird die Reihenfolge selbst geprüft.** Dieser Treiber antwortet wie
// der Stolpertreiber ohne Stolperstelle und schreibt nebenher auf, was er
// gefragt wird — mit `BEGIN`, `COMMIT` und `ROLLBACK` dazwischen.
//
// **Was er nicht prüft**, gehört danebengeschrieben (§10): er sagt nichts
// darüber, ob SQLite diese Reihenfolge auch einhält — nur, dass der Code sie
// verlangt. Dass ein Geschäft hält, was es verspricht, ist eine Zusage von
// SQLite und wird hier nicht nachgemessen.

import (
	"database/sql"
	"sync"
)

func init() {
	sql.Register("aufzeichner", stolpertreiber{stelle: beimAufzeichnen})
}

// aufgezeichnet sammelt, was der Treiber gefragt wurde.
//
// **Als Paketvariable, weil ein Treiber keinen Zustand mitbekommt** — dieselbe
// Überlegung wie bei `zweiterSchreibzaehler`. Die Tests, die ihn benutzen,
// setzen ihn zurück und laufen **nicht** nebenläufig.
var (
	aufgezeichnet []string
	aufzeichnung  sync.Mutex
)

// aufzeichnen hält eine Anweisung fest — nur für den Aufzeichner.
func aufzeichnen(stelle stolperstelle, abfrage string) {
	if stelle != beimAufzeichnen {
		return
	}

	aufzeichnung.Lock()
	defer aufzeichnung.Unlock()

	aufgezeichnet = append(aufgezeichnet, abfrage)
}

// mitschrift gibt das Aufgezeichnete zurück und beginnt von vorn.
func mitschrift() []string {
	aufzeichnung.Lock()
	defer aufzeichnung.Unlock()

	geschrieben := aufgezeichnet
	aufgezeichnet = nil

	return geschrieben
}
