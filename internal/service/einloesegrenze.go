package service

// **Die Ratengrenze der Einlösung** — NET-01b.AK-02
// (#1954).
//
// # Drei Grenzen, und jede hat ihren Angreifer
//
//   - **je Absender**: wer mit einer Kennung Nonces rät;
//   - **je Einladung**: wer Kennungen würfelt, um an einer bekannten
//     Einladung zu rütteln — eine Kennung kostet einen Schlüssel, also
//     nichts;
//   - **insgesamt**: wer beides würfelt. Ohne sie wüchsen die zwei Listen
//     darüber mit jedem Versuch, und ein Fremder bestimmte, wie viel
//     Speicher dieser Posten braucht. **Sie wird zuerst geprüft**, und was
//     sie abweist, wird nicht vermerkt — so bleiben alle drei Listen unter
//     [EinloesungenInsgesamt] Einträgen.
//
// # Warum im Arbeitsspeicher
//
// **Weil sie eine Flut abhält und keine Geschichte führt.** Nach einem
// Neustart beginnt sie leer; wer das ausnutzt, muss den Einladenden zum
// Neustart bringen — und hat dann wieder dieselben Grenzen vor sich.

import (
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Grenzen je [Einloesefenster].
//
// **Fünf je Absender und Einladung**: wer sich vertippt oder dessen Leitung
// abreißt, versucht es zwei-, dreimal; wer es fünfmal in zehn Minuten
// versucht, rät. **Sechzig insgesamt**: mehr Einladungen löst niemand in
// zehn Minuten ein, der sie von Hand verschickt hat.
const (
	EinloesungenJeAbsender  = 5
	EinloesungenJeEinladung = 5
	EinloesungenInsgesamt   = 60
	Einloesefenster         = 10 * time.Minute
)

// Einloesegrenze zählt Einlöseversuche.
type Einloesegrenze struct {
	schloss sync.Mutex
	jetzt   func() time.Time

	alle        []time.Time
	absender    map[string][]time.Time
	einladungen map[string][]time.Time
}

// NeueEinloesegrenze baut sie.
func NeueEinloesegrenze(jetzt func() time.Time) *Einloesegrenze {
	return &Einloesegrenze{
		jetzt: jetzt, absender: map[string][]time.Time{},
		einladungen: map[string][]time.Time{},
	}
}

// Versuch vermerkt einen Versuch — oder weist ihn mit
// [domain.ErrEinloesenGedrosselt] ab.
func (g *Einloesegrenze) Versuch(absender, nonce string) error {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	jetzt := g.jetzt()
	grenze := jetzt.Add(-Einloesefenster)

	g.alle = juenger(g.alle, grenze)
	lichten(g.absender, grenze)
	lichten(g.einladungen, grenze)

	if len(g.alle) >= EinloesungenInsgesamt ||
		len(g.absender[absender]) >= EinloesungenJeAbsender ||
		len(g.einladungen[nonce]) >= EinloesungenJeEinladung {
		return domain.ErrEinloesenGedrosselt
	}

	g.alle = append(g.alle, jetzt)
	g.absender[absender] = append(g.absender[absender], jetzt)
	g.einladungen[nonce] = append(g.einladungen[nonce], jetzt)

	return nil
}

// juenger behält, was nach `grenze` liegt.
func juenger(zeiten []time.Time, grenze time.Time) []time.Time {
	behalten := zeiten[:0]

	for _, zeit := range zeiten {
		if zeit.After(grenze) {
			behalten = append(behalten, zeit)
		}
	}

	return behalten
}

// lichten räumt abgelaufene Versuche und leere Einträge.
func lichten(liste map[string][]time.Time, grenze time.Time) {
	for schluessel, zeiten := range liste {
		if behalten := juenger(zeiten, grenze); len(behalten) > 0 {
			liste[schluessel] = behalten
		} else {
			delete(liste, schluessel)
		}
	}
}
