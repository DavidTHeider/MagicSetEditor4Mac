package service

// **Was mit einer offenen Sitzung geschieht** — Werkstück N1 von Phase 5.4
// (#1454).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Sende reiht eine Nachricht ein.
//
// # Drei Prüfungen vor der Schlange, und jede hat ihren Grund
//
//   - **Die Epoche**: nach einem Profilwechsel gehört dieser Griff zu einem
//     Stand, den es nicht mehr gibt (§7, Regel 2).
//   - **Das Recht**: eine Freigabe kann zurückgenommen worden sein, während
//     die Sitzung offen war — das ist der Berechtigungsentzug aus dem Issue.
//   - **Die Größe**: dieselbe Grenze wie in der Hülle. Wer sie erst beim
//     Verpacken prüfte, hätte den Speicher schon gestellt.
func (g *Sitzungsgriff) Sende(
	ctx context.Context, nachricht []byte, art Zustellung,
) error {
	if fehler := g.pruefeStand(ctx); fehler != nil {
		return fehler
	}

	if len(nachricht) > peer.Hoechstlaenge {
		return fmt.Errorf(
			"%w: die Nachricht wiegt %d Byte, erlaubt sind %d",
			domain.ErrValidation, len(nachricht), peer.Hoechstlaenge)
	}

	g.schloss.Lock()
	defer g.schloss.Unlock()

	// **Die Nachricht, die gerade hinausgeht, zählt mit** — sie steht noch in
	// der Schlange, bis sie angekommen ist (siehe [Sitzungsgriff.Leere]).
	if len(g.warten) < g.eltern.tiefe {
		g.warten = append(g.warten, wartend{nachricht: nachricht, art: art})

		return nil
	}

	// **Zuverlässig heißt: der Absender erfährt es.** Ein Rückstau, der
	// stillschweigend wüchse, wäre keiner.
	if art == Zuverlaessig {
		return fmt.Errorf(
			"%w: die Warteschlange dieser Sitzung ist voll (%d)",
			domain.ErrConflict, g.eltern.tiefe)
	}

	g.stand.Verworfen++

	// **Verwerfbar heißt: der neuere Stand ersetzt den älteren — derselben
	// Art.** Hier wich bis zum Nachlauf über Phase 6.2 die älteste Nachricht
	// gleich welcher Art, auch eine zuverlässige, deren Absender „angenommen"
	// gehört hatte. **Nie weicht die, die gerade hinausgeht**, und steht
	// nichts Verwerfbares an, weicht die neue selbst.
	for stelle, eintrag := range g.warten {
		if eintrag.art != Verwerfbar || (stelle == 0 && g.unterwegs) {
			continue
		}

		g.warten = append(append(g.warten[:stelle:stelle], g.warten[stelle+1:]...),
			wartend{nachricht: nachricht, art: art})

		return nil
	}

	return nil
}

// Leere schickt hinaus, was wartet.
//
// # Warum das ein eigener Ruf ist und keine Goroutine
//
// **Weil eine Goroutine eine Uhr bräuchte, die niemand stellt.** Wer sie
// einsetzt, hat einen Lauf, der auf echte Zeit wartet — und §3 sagt, was das
// wert ist. Der Transport (N2, N3) ruft sie, wenn die Leitung bereit ist;
// hier steht, **was** dann geschieht, und das ist prüfbar.
//
// **Eine Nachricht, die nicht hinausgeht, bleibt stehen.** Sie ist
// zuverlässig zugesagt; wer sie wegwürfe, hätte die Zusage gebrochen — und
// der Absender erführe es nicht.
func (g *Sitzungsgriff) Leere(ctx context.Context) error {
	if fehler := g.pruefeStand(ctx); fehler != nil {
		return fehler
	}

	// **Einer leert, und zwar der Reihe nach.** Hier las jeder Ruf die erste
	// Nachricht, gab das Schloss frei, schickte sie und nahm sie erst danach
	// heraus. Zwei zugleich — die Netzschleife einer Partie und eine Handlung
	// am Tisch leeren denselben Griff — schickten dieselbe Nachricht zweimal,
	// und der zweite schnitt aus einer leeren Schlange: „slice bounds out of
	// range [1:0]", ein roter Racelauf am 2026-09-24. **Das Schloss der
	// Schlange bleibt beim Schicken frei**: `Sende` soll nicht auf die
	// Leitung warten.
	g.leeren.Lock()
	defer g.leeren.Unlock()

	for {
		g.schloss.Lock()

		if len(g.warten) == 0 {
			g.schloss.Unlock()

			return nil
		}

		// **Sie bleibt stehen, bis sie angekommen ist**, und sie zählt zur
		// Tiefe. Der Zweitblick nahm sie vorher heraus; ein `Sende`
		// währenddessen belegte ihren Platz, und scheiterte sie, standen mehr
		// Nachrichten an, als die Tiefe erlaubt (Nachlauf über Phase 6.2).
		// `unterwegs` sagt `Sende`, dass diese eine nicht weichen darf.
		naechste := g.warten[0].nachricht
		g.unterwegs = true

		g.schloss.Unlock()

		fehler := g.verbindung.Sende(ctx, naechste)

		g.schloss.Lock()
		g.unterwegs = false

		// **Vorn steht noch dieselbe**: nur `Leere` nimmt vorn heraus, und
		// es leert nur einer; `Sende` rührt die Nachricht unterwegs nicht an.
		if fehler == nil {
			g.warten = g.warten[1:]
		}

		g.schloss.Unlock()

		if fehler != nil {
			return fehler
		}
	}
}

// Empfange liest die nächste Nachricht dieser Sitzung.
//
// # Warum der Stand **nach** dem Lesen geprüft wird und nicht davor
//
// **Weil die Antwort zu dem Stand gehört, aus dem sie stammt** (§7, Regel 2).
// Eine Prüfung vor dem Lesen beantwortete die Frage „galt die Sitzung, als wir
// uns hingesetzt haben?" — und zwischen dem Hinsetzen und der Nachricht liegt
// beliebig viel Zeit: ein Profilwechsel, ein zurückgenommenes Recht. Gefragt
// ist, ob sie gilt, **wenn die Nachricht da ist**.
//
// **Und es ist genau eine Prüfung je Nachricht**, nicht zwei. Eine zweite
// davor kostete bei jeder einen weiteren Gang in den Bestand und beantwortete
// eine Frage, die niemand gestellt hat.
//
// **Was das offenlässt, steht hier**: eine Sitzung, deren Recht entzogen wird,
// während niemand schickt, bleibt am Lauschen, bis der Kontext endet. Es geht
// dabei nichts hinaus und nichts herein — aber es ist ein Wartender, den
// niemand mehr braucht, und beenden kann ihn nur sein Aufrufer.
func (g *Sitzungsgriff) Empfange(ctx context.Context) ([]byte, error) {
	nachricht, fehler := g.verbindung.Empfange(ctx)
	if fehler != nil {
		return nil, fehler
	}

	if fehler := g.pruefeStand(ctx); fehler != nil {
		return nil, fehler
	}

	return nachricht, nil
}

// Empfangen sagt, ob eine eingetroffene Nachricht noch gilt.
//
// **Sie gibt einen Wahrheitswert und keinen Fehler.** „Zu spät" ist kein
// Fehlschlag, sondern die richtige Antwort auf eine Nachricht, die zu einem
// Profil gehört, das nicht mehr offen ist — und ein Fehler zwänge jeden
// Adapter, ihn wegzufangen.
func (g *Sitzungsgriff) Empfangen([]byte) bool {
	g.schloss.Lock()
	zu := g.zu
	epoche := g.epoche
	g.schloss.Unlock()

	if zu {
		return false
	}

	g.eltern.schloss.Lock()
	defer g.eltern.schloss.Unlock()

	return g.eltern.epoche == epoche
}

// Stand gibt, was über diese Sitzung feststeht.
func (g *Sitzungsgriff) Stand() Sitzungsstand { return g.jetzigerStand() }

// Schliesse beendet die Sitzung.
func (g *Sitzungsgriff) Schliesse() error {
	g.eltern.vergiss(g.stand.Kennung, g)

	return g.schliesseStill()
}

// schliesseStill legt die Leitung nieder, ohne die Übersicht anzufassen.
//
// **Getrennt, weil [Peersitzungen.Pausiere] die Übersicht in einem Zug
// räumt** — und ein zweiter Griff hinein wäre ein zweiter Stand.
func (g *Sitzungsgriff) schliesseStill() error {
	g.schloss.Lock()

	if g.zu {
		g.schloss.Unlock()

		return nil
	}

	g.zu = true

	g.schloss.Unlock()

	return g.verbindung.Schliesse()
}

// pruefeStand hält den Griff gegen den heutigen Stand.
func (g *Sitzungsgriff) pruefeStand(ctx context.Context) error {
	g.schloss.Lock()
	zu := g.zu
	epoche := g.epoche
	g.schloss.Unlock()

	g.eltern.schloss.Lock()
	jetzige := g.eltern.epoche
	g.eltern.schloss.Unlock()

	if zu || jetzige != epoche {
		return fmt.Errorf(
			"%w: diese Sitzung gehört nicht mehr zum offenen Profil",
			domain.ErrConflict)
	}

	darf, fehler := g.eltern.rechte.Darf(
		ctx, g.stand.Gegenueber, g.stand.Zweck)
	if fehler != nil {
		return fehler
	}

	if !darf {
		return fmt.Errorf(
			"%w: %s darf %q nicht mehr", domain.ErrValidation,
			g.stand.Fingerabdruck, g.stand.Zweck)
	}

	return nil
}

// jetzigerStand gibt den Stand samt der Zahl der wartenden Nachrichten.
func (g *Sitzungsgriff) jetzigerStand() Sitzungsstand {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	stand := g.stand
	stand.Wartend = len(g.warten)

	return stand
}
