package netz

import "time"

// GesperrtBis sagt, bis wann diese Drossel gesperrt ist — die Nullzeit, wenn
// sie es nicht ist.
//
// # Wofür
//
// **Für einen Aufrufer, der das Ende nennen muss** (Nachlauf über Phase 0,
// Befund 2, #2333). Ein 429 sperrt schon in [Hole], ein Dienst mit eigener
// Deutung sperrt danach noch einmal, und [Drossel.Sperre] behält die
// längere. Wer seine eigene Zahl meldet statt dieser, meldet ein Ende, zu
// dem noch nichts hinausgeht.
//
// **Was eine andere Anwendung gesperrt hat, steht hier nicht** — das liest
// [Drossel.Warte] aus der Absprache, bevor es sendet.
func (d *Drossel) GesperrtBis() time.Time {
	d.schranke.Lock()
	defer d.schranke.Unlock()

	if !d.gesperrtBis.After(d.jetzt()) {
		return time.Time{}
	}

	return d.gesperrtBis
}

// SperreGegenstelle hält alles bis zu einem Zeitpunkt an — wie
// [Drossel.SperreFuer] für das ganze Werk, aber mit einem Ende statt einer
// Dauer.
//
// **Für ein Ende, das die Gegenstelle selbst nennt** (Nachlauf über Phase 0,
// Befund 2, #2333): GitHub schickt `x-ratelimit-reset` als Zeitpunkt. Ihn in
// eine Dauer und zurück zu rechnen, verschob ihn um die Zeit dazwischen.
//
// **Eine Drossel ohne Werk sperrt nur sich selbst** — dieselbe Regel wie
// dort.
func (d *Drossel) SperreGegenstelle(bis time.Time) {
	if d.werk == nil {
		d.Sperre(bis)

		return
	}

	d.werk.SperreAlle(bis)
}
