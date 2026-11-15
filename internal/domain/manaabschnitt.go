package domain

// **Mana kommt hinzu und geht wieder** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **`manavorrat.go` sagt, was ein Vorrat ist**, `manazahlung.go`, wie man ihn
// ausgibt — hier steht, **wann er entsteht und wann er verschwindet**. Das ist
// eine Frage an den Zug und nicht an das Mana.

import "fmt"

// ManaHinzu legt Mana in den Vorrat eines Sitzes — CR 106.4.
//
// **Eine Menge und kein Stand**: „zwei rote dazu" überlebt eine Antwort, die
// sich mit einer anderen kreuzt (§7, Regel 2). Derselbe Grund wie bei
// [Partie.Leben] nebenan.
func (p Partie) ManaHinzu(wer string, art Manaart, um int) (Partie, error) {
	if _, bekannt := ManaartAus(string(art)); !bekannt {
		return p, fmt.Errorf("%w: %q ist keine Manaart (CR 106.1b)",
			ErrValidation, art)
	}

	for stelle := range p.Sitze {
		if p.Sitze[stelle].Kennung != wer {
			continue
		}

		// **Die Liste wird kopiert, bevor sie sich ändert.** Eine Partie ist
		// ein Wert; ihre Sitze liegen in einer geteilten Reihe, und wer
		// darin schriebe, änderte jede Kopie mit — dasselbe wie bei
		// [Partie.merkeLeergezogen].
		sitze := make([]Sitz, len(p.Sitze))
		copy(sitze, p.Sitze)
		sitze[stelle].Manavorrat = sitze[stelle].Manavorrat.Mehr(art, um)
		p.Sitze = sitze

		return p, nil
	}

	return p, fmt.Errorf("%w: den Sitz %q gibt es in dieser Partie nicht",
		ErrNotFound, wer)
}

// Manaleeren leert jeden Vorrat — CR 500.5.
//
// # Warum alle auf einmal und nicht der aktive Spieler
//
// **Weil CR 500.5 von „a player's mana pool" spricht und der Abschnitt für
// alle endet.** Wer nur den aktiven leerte, ließe dem anderen sein Mana über
// den Abschnitt hinweg — und der bekäme damit etwas, das die Regel niemandem
// gibt.
//
// **Sie gibt immer eine Partie zurück, auch wenn nichts zu tun war.** Ein
// zweiter Rückgabewert „hat sich etwas geändert" stünde hier ohne Verbraucher:
// das Leeren ist eine zugbasierte Handlung (CR 703.4q) und keine
// zustandsbasierte Aktion, die sich wiederholt.
func (p Partie) Manaleeren() Partie {
	leerBereits := true

	for _, sitz := range p.Sitze {
		if !sitz.Manavorrat.Leer() {
			leerBereits = false

			break
		}
	}

	if leerBereits {
		return p
	}

	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)

	for stelle := range sitze {
		sitze[stelle].Manavorrat = nil
	}

	p.Sitze = sitze

	return p
}
