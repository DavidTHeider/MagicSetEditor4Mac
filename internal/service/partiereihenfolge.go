package service

// **Der Wirt legt die Reihenfolge fest** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wofür es diese Stelle gibt
//
// Das Issue verlangt es als Erstes: „G1-Host verarbeitet gleichzeitig
// eintreffende Commands in **einer** Reihenfolge."
//
// **Zwei Spieler, die im selben Augenblick ziehen, ziehen nacheinander** — und
// welcher zuerst, entscheidet der Wirt. Ohne diese Stelle hätten zwei
// Teilnehmer zwei verschiedene Geschichten derselben Partie, und die Hashkette
// aus G1 fiele bei der ersten Gleichzeitigkeit auseinander.
//
// # Warum sie so klein ist
//
// **Weil sie nur eine Frage beantwortet**: welche Nummer bekommt das nächste
// Ereignis? Was daraus folgt — senden, ketten, aufbewahren — steht anderswo.
// Eine Stelle, die zugleich zählte und schickte, wäre unter dem
// Race-Detektor nicht mehr zu lesen.

import "sync"

// Reihenfolge vergibt Folgenummern innerhalb einer Epoche.
//
// **Sicher für mehrere Goroutinen** — das ist ihr ganzer Zweck.
type Reihenfolge struct {
	schloss sync.Mutex
	epoche  uint64
	folge   uint64
}

// NeueReihenfolge fängt in einer Epoche an, ohne eine Folge vergeben zu haben.
//
// **Die Epoche kommt von außen**, denn sie überlebt einen Neustart: nach ihm
// ist es die gespeicherte, um eins erhöht. Hier gezählt würde sie bei jedem
// Start wieder eins, und zwei Läufe wären nicht mehr zu unterscheiden.
func NeueReihenfolge(epoche uint64) *Reihenfolge {
	return &Reihenfolge{epoche: epoche}
}

// Naechste gibt die nächste Nummer — genau einmal.
func (r *Reihenfolge) Naechste() (epoche, folge uint64) {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	r.folge++

	return r.epoche, r.folge
}

// Stand liest ab, ohne weiterzuzählen.
//
// **Ein Gast, der nach dem Stand fragt, soll ihn nicht verändern** — sonst
// entstünde eine Lücke allein durchs Hinsehen.
func (r *Reihenfolge) Stand() (epoche, folge uint64) {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	return r.epoche, r.folge
}

// NeueEpoche fängt von vorn an.
//
// **Die Folge beginnt wieder bei eins**, und die Epoche trennt die beiden
// Läufe — ein Gast aus der alten kommt damit nicht zurück, und das ist die
// Zusage aus `domain.Wiederaufnahme`.
func (r *Reihenfolge) NeueEpoche() uint64 {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	r.epoche++
	r.folge = 0

	return r.epoche
}
