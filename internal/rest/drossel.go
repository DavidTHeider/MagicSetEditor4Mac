package rest

import (
	"sync"
	"time"
)

// Drossel begrenzt, wie viele Anfragen in einem Zeitfenster durchkommen.
//
// # Warum ein festes Fenster und kein Eimer
//
// Ein Zeitfenster zählt und setzt zurück; ein Eimer läuft ständig nach. Der
// Unterschied zählt bei einer gleichmäßigen Last — die es hier nicht gibt.
// Wogegen die Drossel steht, ist das **Raten** der Marke, und dagegen ist das
// feste Fenster genauso wirksam und in einem Satz zu erklären.
//
// **Der Preis steht dazu**: an der Fensterkante kommen im ungünstigsten Fall
// zweimal `grenze` Anfragen kurz hintereinander durch. Für das Raten einer
// 64-Zeichen-Marke ändert das nichts.
type Drossel struct {
	grenze  int
	fenster time.Duration
	jetzt   func() time.Time

	schranke sync.Mutex
	seit     time.Time
	gezaehlt int
}

// NeueDrossel baut eine Drossel.
//
// **Die Uhr wird eingesetzt** (§16): sonst braucht jeder Test eine echte
// Wartezeit, und ein Lauf, der auf Uhren wartet, flackert unter Last.
func NeueDrossel(grenze int, fenster time.Duration, jetzt func() time.Time) *Drossel {
	return &Drossel{grenze: grenze, fenster: fenster, jetzt: jetzt, seit: jetzt()}
}

// Durchlassen sagt, ob diese Anfrage noch in das Fenster passt — und zählt sie.
//
// **Fragen und Zählen in einem Schritt.** Getrennt hätten zwei gleichzeitige
// Aufrufer beide ein Ja bekommen, bevor einer gezählt hat; `net/http` bedient
// jede Anfrage in einer eigenen Goroutine, das ist also der Normalfall.
func (d *Drossel) Durchlassen() bool {
	d.schranke.Lock()
	defer d.schranke.Unlock()

	if jetzt := d.jetzt(); jetzt.Sub(d.seit) >= d.fenster {
		d.seit = jetzt
		d.gezaehlt = 0
	}

	if d.gezaehlt >= d.grenze {
		return false
	}

	d.gezaehlt++

	return true
}
