package domain

// **Die Mischpflicht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was sie ist
//
// **Die Gegenleistung für einen Blick.** Wer eine verborgene Zone ansieht,
// zerstört ihr Geheimnis — und nur ein Mischen stellt es wieder her. Bis
// dahin steht hier, dass es aussteht.
//
// # Warum im Zustand und nicht nur im Ereignisstrom
//
// **Weil sie eine Tatsache über den Tisch ist und keine über die
// Vergangenheit.** Der Blick steht als Ereignis da; dass seitdem **nicht**
// gemischt wurde, müsste sonst jeder Empfänger aus dem Strom nachrechnen —
// und ein Zuschauer, der später dazukommt, hat den Strom gar nicht.
//
// # Und warum G1 sie nicht erzwingt
//
// **Weil der Sandkasten dem Menschen gehorcht.** Zwei Leute an einem Tisch
// mischen, weil sie es schulden, nicht weil ein Programm sie zwingt.
// **Dass** nach einer Suche gemischt werden **muss**, ist Werkstück G4
// (#1461);
// was G1 schuldet, ist der Platz, an dem G4 es ablesen kann — und die
// Sichtbarkeit, die einen Verstoß **auffällig** macht, statt ihn zu
// verhindern.

// Mischpflicht sagt, dass eine verborgene Zone angesehen und seitdem nicht
// gemischt wurde.
//
// **Öffentlich, und das ist ihr ganzer Zweck**: sie sagt dem Gegenüber, was
// jemand gesehen hat. Stünde sie nur beim Betroffenen, wäre sie eine Notiz an
// sich selbst.
type Mischpflicht struct {
	Sitz string
	Zone Zone
}

// SchuldetMischen sagt, ob dieser Sitz diese Zone noch mischen muss.
func (p Partie) SchuldetMischen(sitz string, zone Zone) bool {
	for _, pflicht := range p.Mischpflichten {
		if pflicht.Sitz == sitz && pflicht.Zone == zone {
			return true
		}
	}

	return false
}

// MerkeMischpflicht trägt eine ein — höchstens einmal je Sitz und Zone.
//
// **Höchstens einmal**, denn zweimal ansehen ist eine Pflicht und nicht zwei;
// stünde sie doppelt da, löschte ein Mischen nur die Hälfte.
func (p *Partie) MerkeMischpflicht(sitz string, zone Zone) {
	if p.SchuldetMischen(sitz, zone) {
		return
	}

	p.Mischpflichten = append(p.Mischpflichten,
		Mischpflicht{Sitz: sitz, Zone: zone})
}

// LoeseMischpflicht nimmt sie zurück.
//
// **Eine Pflicht, die es nicht gab, zu lösen ist kein Fehler**: wer mischt,
// ohne gesehen zu haben, darf das — es ist die Richtung, in der ein Irrtum
// nichts kostet.
func (p *Partie) LoeseMischpflicht(sitz string, zone Zone) {
	behalten := make([]Mischpflicht, 0, len(p.Mischpflichten))

	for _, pflicht := range p.Mischpflichten {
		if pflicht.Sitz == sitz && pflicht.Zone == zone {
			continue
		}

		behalten = append(behalten, pflicht)
	}

	p.Mischpflichten = behalten
}
