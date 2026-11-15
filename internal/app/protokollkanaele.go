package app

// **Jeder Kanal schreibt für sich** — Befunde 3 und 10 des Zweitblicks über
// Phase 6.4 (#1665).
//
// # Warum kein `io.MultiWriter`
//
// **Er hört beim ersten Fehler auf** („If a listed writer returns an error,
// that overall write operation stops", `io.MultiWriter`, nachgelesen in Go
// 1.27.1). Stand die Datei vorn und war ihr Laufwerk fort, bekamen Unified
// Log und Standardfehlerausgabe nichts mehr — genau das, was OPS-04.AK-02
// ausschließt.
//
// # Und der Stand weiß es
//
// **Ein Kanal, der beim Schreiben scheitert, heißt im Stand nicht mehr
// „an"**, und sein Grund steht dabei — wie bei einer Datei, die sich nicht
// öffnen ließ.

import (
	"sync"
)

// kanalschreiber ist ein Kanal, der sich seinen ersten Fehler merkt.
type kanalschreiber struct {
	name   string
	ziel   interface{ Write([]byte) (int, error) }
	sperre sync.Mutex
	grund  string
}

// schreiben gibt eine Zeile an das Ziel und merkt sich, wie es ausging.
//
// **Der jüngste Ausgang zählt** — Nachlauf über Phase 6.4, N13: hier blieb
// der erste Fehler kleben, und ein einmaliger Zeitablauf von `logger` nannte
// das Unified Log bis zum nächsten Einrichten „fehlt", obwohl es längst
// wieder schrieb.
func (k *kanalschreiber) schreiben(zeile []byte) {
	_, fehler := k.ziel.Write(zeile)

	k.sperre.Lock()
	defer k.sperre.Unlock()

	k.grund = ""
	if fehler != nil {
		k.grund = schwaerze(fehler.Error())
	}
}

// Write schreibt in diesen Kanal — **ein Fehler erreicht den Schreibenden
// nicht**: ein Protokoll, das beim Protokollieren scheitert, darf den Aufrufer
// nicht scheitern lassen, und die anderen Kanäle schreiben weiter.
func (k *kanalschreiber) Write(zeile []byte) (int, error) {
	k.schreiben(zeile)

	return len(zeile), nil
}

func (k *kanalschreiber) gescheitert() string {
	k.sperre.Lock()
	defer k.sperre.Unlock()

	return k.grund
}

// verteiler sind die Kanäle einer Regel — jeder mit seinem eigenen Schreiber
// (Nachlauf über Phase 6.4, N5), und jeder merkt sich seinen Ausgang.
type verteiler []*kanalschreiber

// mitSchreibfehlern trägt die Fehler der Kanäle in einen Stand ein.
func mitSchreibfehlern(stand Protokollstand, kanaele verteiler) Protokollstand {
	stand.Kanaele = append([]Protokollkanal(nil), stand.Kanaele...)

	for _, kanal := range kanaele {
		grund := kanal.gescheitert()
		if grund == "" {
			continue
		}

		for stelle := range stand.Kanaele {
			if stand.Kanaele[stelle].Name == kanal.name {
				stand.Kanaele[stelle] = Protokollkanal{Name: kanal.name, Grund: grund}
			}
		}
	}

	return stand
}
