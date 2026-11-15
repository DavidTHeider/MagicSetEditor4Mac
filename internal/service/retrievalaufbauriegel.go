package service

// **Der Riegel über dem Indexaufbau** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Zwei Riegel und ein Stand
//
// **Einer im Prozess und einer darüber.** Der erste hält das Fenster davon
// ab, zweimal zugleich zu bauen; der zweite hält einen **anderen Prozess**
// davon ab — `mse indexaufbau` weiß nichts vom Fenster, und beide schreiben
// in denselben Bestand.
//
// **Der Stand überlebt den Lauf**, weil ein Fehlerzustand, der mit ihm
// verschwände, für niemanden lesbar wäre.
//
// # Warum das neben dem Aufbau steht
//
// **Weil es zwei Fragen sind**: dort „wie entsteht eine Generation?", hier
// „wer darf gerade?". Zusammen liefen beide über die Zeilengrenze aus §2 —
// die Grenze ist der Anlass des Schnitts und nicht sein Grund.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Abbrechen hält den laufenden Aufbau an.
//
// **Ohne zu warten.** Wer abbricht, will nicht wissen, wann es vorbei ist; er
// will, dass es aufhört. Der Lauf verwirft seine Generation selbst
// ([Indexaufbau.lauf]), und [Indexaufbau.Stand] sagt danach, was war.
func (a *Indexaufbau) Abbrechen() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.abbrechen == nil {
		return false
	}

	a.abbrechen()

	return true
}

// Stand gibt, was gerade gilt.
func (a *Indexaufbau) Stand() Aufbaustand {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.stand
}

// prozesssperre nimmt die Absprache über die Prozesse — oder weist ab.
//
// **Ohne Sperre gelingt sie.** Ein Prozess ohne Absprache ist der Stand vor
// dieser Zeile, und das steht hier statt in einer stillen Rückgabe.
func (a *Indexaufbau) prozesssperre() (func(), error) {
	if a.sperre == nil {
		return func() {}, nil
	}

	freigeben, genommen, fehler := a.sperre.Nehmen()
	if fehler != nil {
		return nil, fehler
	}

	if !genommen {
		return nil, domain.MitGrund(domain.ErrConflict, nil,
			"ein anderer Prozess baut gerade an diesem Index — zwei zugleich "+
				"schrieben in denselben Bestand")
	}

	return freigeben, nil
}

// beginnen setzt den Riegel — oder weist ab.
func (a *Indexaufbau) beginnen(abbrechen context.CancelFunc) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.stand.Laeuft {
		return domain.MitGrund(domain.ErrConflict, nil,
			"es läuft bereits ein Indexaufbau (Generation %d, %d Abschnitte "+
				"geschrieben) — zwei zugleich bauen zweimal denselben Stand",
			a.stand.Generation, a.stand.Geschrieben)
	}

	a.stand = Aufbaustand{Laeuft: true, Begonnen: time.Now()}
	a.abbrechen = abbrechen

	return nil
}

// beenden hebt den Riegel und schreibt den Ausgang fest.
func (a *Indexaufbau) beenden(generation int64, geschrieben int, fehler error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stand.Laeuft = false
	a.abbrechen = nil
	a.stand.Generation = generation
	a.stand.Geschrieben = geschrieben
	a.stand.Beendet = time.Now()

	if fehler != nil {
		// **Der Satz und nicht die Kette** (§8, [domain.SatzVon]): der Stand
		// wird über jede Fläche gezeigt, und ein `%w`-gewickelter Pfad hat
		// dort nichts zu suchen.
		a.stand.Fehler = domain.SatzVon(fehler)
		if a.stand.Fehler == "" {
			a.stand.Fehler = "Der Aufbau ist gescheitert."
		}
	}
}

// fortschritt meldet, wie weit der laufende Aufbau ist.
func (a *Indexaufbau) fortschritt(generation int64, geschrieben int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stand.Generation = generation
	a.stand.Geschrieben = geschrieben
}
