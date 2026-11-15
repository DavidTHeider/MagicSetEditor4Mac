package rest

// **Wem eine Ausgabe gehört** — SEC-P6
// (#2002),
// Befund P6 in `docs/PENTEST.md`.
//
// # Der Befund
//
// **Der Name war das einzige Recht.** `GET` und `DELETE /ausgang/{name}`
// fragten nur, ob es die Datei gibt; jeder gültige Aufrufer — auch ein
// Lesetoken eines anderen Profils — holte ab, wer den Namen kannte. Der Name
// ist gewürfelt (80 Bit), aber er steht im `Location` einer Antwort und im
// Ergebnis eines Auftrags, und beides kann weiterwandern.
//
// # Wie bei den Aufträgen
//
// **Die Ausgabe gehört dem Principal, der sie anstieß** — `auftraege.go`,
// `eigener`. Ein anderer bekommt dasselbe `404` wie für einen Namen, den es
// nicht gibt: ein `403` verriete, dass es ihn gibt.
//
// # Warum hier und nicht in der Ablage
//
// **Der Principal ist eine Angabe dieses Dienstes** — die Ablage in
// `internal/app` kennt ihn nicht, und sie soll ihn nicht kennen müssen. Der
// Vermerk lebt im Arbeitsspeicher wie das Auftragsbuch: ein neu gestarteter
// Dienst kennt keinen Eigner mehr, und die Ausgaben davor holt **niemand**
// mehr ab — die Ablage räumt sie nach einer Stunde.

import (
	"sync"
	"time"
)

// vermerkfrist ist, wie lange ein Vermerk bleibt — **länger als die Ablage
// eine Ausgabe hält** (eine Stunde, `internal/app/dienstablage.go`). Ein
// Vermerk, der vor seiner Datei ginge, sperrte den Eigner aus; einer, der
// nie ginge, wüchse mit jeder Ausgabe eines Dienstes, der Wochen läuft.
const vermerkfrist = 24 * time.Hour

// ausgangsbuch merkt sich, wem welche Ausgabe gehört.
type ausgangsbuch struct {
	sperre sync.Mutex
	jetzt  func() time.Time
	eigner map[string]ausgangsvermerk
}

// ausgangsvermerk ist ein Eigner und seit wann.
type ausgangsvermerk struct {
	principal string
	seit      time.Time
}

// neuesAusgangsbuch baut eines über einer Uhr.
func neuesAusgangsbuch(jetzt func() time.Time) *ausgangsbuch {
	return &ausgangsbuch{jetzt: jetzt, eigner: map[string]ausgangsvermerk{}}
}

// vermerke trägt den Eigner einer neuen Ausgabe ein — und räumt ab, was
// älter als die Frist ist.
func (b *ausgangsbuch) vermerke(name, principal string) {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	jetzt := b.jetzt()

	for alt, vermerk := range b.eigner {
		if jetzt.Sub(vermerk.seit) > vermerkfrist {
			delete(b.eigner, alt)
		}
	}

	b.eigner[name] = ausgangsvermerk{principal: principal, seit: jetzt}
}

// gehoert sagt, ob die Ausgabe diesem Principal gehört.
//
// **Ohne Vermerk gehört sie niemandem** — auch eine Datei, die ein früherer
// Lauf dieses Dienstes hinterließ.
func (b *ausgangsbuch) gehoert(name, principal string) bool {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	vermerk, da := b.eigner[name]

	return da && vermerk.principal == principal
}

// vergiss nimmt den Vermerk einer vergessenen Ausgabe weg.
func (b *ausgangsbuch) vergiss(name string) {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	delete(b.eigner, name)
}
