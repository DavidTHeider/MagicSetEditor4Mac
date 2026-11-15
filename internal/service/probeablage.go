package service

import "sync"

// Probeablage hält fest, was eine Ausfuhr **neben ihre Hauptdatei** geschrieben
// hätte — und schreibt es nicht
// (#2015,
// UXR-F7).
//
// # Wozu es sie gibt
//
// Eine Exportvorlage schreibt **während** der Auswertung: `write_image_file`,
// `copy_file`, `write_text_file`, `write_set_file`, die Symbolbilder. Ein
// Probelauf, der die Vorlage auswertet, ohne dass sie schreibt, setzt an diese
// Stellen eine Ablage — dieselben Funktionen, dieselben Namensprüfungen, und
// am Ende der Schritt, der auf die Platte ginge, ein Eintrag hier.
//
// **Der Name ist der Schlüssel.** Eine Vorlage, die denselben Namen zweimal
// nennt, meint dieselbe Datei — der echte Lauf überschriebe sie, und gezählt
// wird, was auf der Platte läge.
type Probeablage struct {
	mu sync.Mutex

	// groessen sind die Bytes je Name; negativ heißt: unbekannt.
	groessen map[string]int
}

// NeueProbeablage baut eine leere Ablage.
func NeueProbeablage() *Probeablage {
	return &Probeablage{groessen: map[string]int{}}
}

// Lege hält eine Datei mit bekannter Größe fest.
func (p *Probeablage) Lege(name string, bytes int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.groessen[name] = bytes
}

// LegeOhneGroesse hält eine Datei fest, deren Größe der Probelauf nicht
// kennt — ein Kartenbild entsteht erst beim Rastern, ein Set erst beim Packen.
func (p *Probeablage) LegeOhneGroesse(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.groessen[name] = -1
}

// Probestand ist, was die Ablage über die Dateien neben der Hauptdatei weiß.
type Probestand struct {
	// Dateien sind die Namen, die geschrieben würden.
	Dateien int

	// Bytes ist die Summe der **bekannten** Größen.
	Bytes int

	// Geschaetzt sagt, dass mindestens eine Datei dabei ist, deren Größe der
	// Probelauf nicht kennt — die Summe ist dann eine Untergrenze.
	Geschaetzt bool
}

// Stand zählt zusammen, was festgehalten wurde.
func (p *Probeablage) Stand() Probestand {
	p.mu.Lock()
	defer p.mu.Unlock()

	stand := Probestand{Dateien: len(p.groessen)}

	for _, bytes := range p.groessen {
		if bytes < 0 {
			stand.Geschaetzt = true

			continue
		}

		stand.Bytes += bytes
	}

	return stand
}
