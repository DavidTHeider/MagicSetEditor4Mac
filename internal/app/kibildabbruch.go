package app

// **Ein laufender Bildauftrag lässt sich abbrechen** — Werkstück A17 von
// Phase 5.3
// (#1532).
//
// # Der Befund
//
// **Es gab keinen Weg dazu.** `KIBild.tsx` kennt einen Zustand `laeuft`, der
// Knöpfe abschaltet — und nichts, was den Lauf beendet. Wer einen Auftrag
// anstößt, den er nicht wollte, wartet ihn ab und bezahlt ihn; und ein
// Bildauftrag dauert Sekunden bis Minuten.
//
// **A15 hat den Weg längst** ([App.IndexAbbrechen]); der Bildweg hatte ihn
// nicht.
//
// # Warum je Karte und nicht einer für alles
//
// **Weil das Fenster je Karte anstößt.** `useVorgang` führt den Schlüssel
// `kibild/<karteID>`, und ein Abbruch, der jeden Auftrag beendete, nähme
// jemandem den weg, den er an einer anderen Karte angestoßen hat.
//
// # Und warum der Abbruch keinen Fehler gibt
//
// **„Es lief nichts" ist eine Auskunft** — dieselbe Gestalt wie bei
// [App.IndexAbbrechen]. Ein Fehlschlag hieße, den Weg gäbe es nicht.

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bildlauf ist **ein** Auftrag und sein Abbruchweg.
//
// **Ein Zeiger, damit er eine Identität hat.** Zwei Läufe derselben Karte
// unterscheiden sich sonst durch nichts: eine `context.CancelFunc` lässt sich
// nicht vergleichen, und `func` ist in Go nicht einmal gleichheitsfähig.
type bildlauf struct{ abbrechen context.CancelFunc }

// bildlaeufe hält je Karte den Abbruch ihres laufenden Auftrags.
//
// **Ein Auftrag je Karte, und der zweite wird abgewiesen** (§7, Regel 4).
// `useVorgang` hält ihn am Fenster auf; die Go-Seite verlässt sich nicht
// darauf, denn die Bindung ist auch direkt zu rufen.
//
// # Der Weg dahin, in zwei Schritten
//
// **Erst löschte das Ende eines Laufs jeden Eintrag der Karte** — auch den
// eines zweiten. Darüber stand als Kommentar „nur den eigenen", und der Rumpf
// hielt es nicht: der zweite Auftrag verlor seinen Abbruchweg, sobald der
// erste endete (Befund 3 des Zweitblicks über Phase 5.3).
//
// **Dann gehörte der Eintrag dem jüngeren Lauf** — und der ältere war
// weiterhin nicht anzuhalten: der Fall war behoben, die Klasse nicht (Befund
// 3 des Nachlaufs). Jetzt kommt der zweite gar nicht erst hinein, und zwar
// **bevor** irgendetwas Geld kostet.
type bildlaeufe struct {
	sperre sync.Mutex
	offen  map[string]*bildlauf
}

// anmelden hängt einen Auftrag ein und gibt seinen abbrechbaren Kontext.
//
// **Läuft schon einer an dieser Karte, gibt es einen Konflikt** — und keinen
// Rückgeber: wer nichts angemeldet hat, hat nichts abzumelden.
func (b *bildlaeufe) anmelden(ctx context.Context, karteID int64) (
	abbrechbar context.Context, fertig func(), fehler error,
) {
	schluessel := bildschluessel(karteID)

	b.sperre.Lock()

	if b.offen == nil {
		b.offen = map[string]*bildlauf{}
	}

	if _, laeuft := b.offen[schluessel]; laeuft {
		b.sperre.Unlock()

		return nil, nil, fmt.Errorf(
			"%w: an dieser Karte entsteht schon ein Bild — warte es ab oder "+
				"brich es ab", domain.ErrConflict)
	}

	abbrechbar, abbrechen := context.WithCancel(ctx)
	meiner := &bildlauf{abbrechen: abbrechen}

	b.offen[schluessel] = meiner

	b.sperre.Unlock()

	// **Der Rückgeber räumt auf und bricht ab.** Beides, und in dieser
	// Reihenfolge: ein `cancel`, das nie liefe, wäre ein Kontext, der bis zum
	// Programmende hängt — `go vet` streicht genau das an.
	return abbrechbar, func() {
		b.sperre.Lock()

		// **Nur den eigenen Eintrag entfernen** — und „eigen" heißt: es ist
		// noch derselbe Lauf. Seit ein zweiter abgewiesen wird, kann hier
		// kein fremder mehr stehen; die Prüfung bleibt, weil sie die Zusage
		// **hält** und nicht bloß behauptet.
		if b.offen[schluessel] == meiner {
			delete(b.offen, schluessel)
		}

		b.sperre.Unlock()

		abbrechen()
	}, nil
}

// abbrechen beendet den Auftrag einer Karte — falls einer läuft.
func (b *bildlaeufe) abbrechen(karteID int64) bool {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	lauf, laeuft := b.offen[bildschluessel(karteID)]
	if !laeuft {
		return false
	}

	lauf.abbrechen()

	return true
}

// alleAbbrechen beendet jeden laufenden Auftrag — `ki` aus (OPS-10b,
// #1978).
func (b *bildlaeufe) alleAbbrechen() {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	for _, lauf := range b.offen {
		lauf.abbrechen()
	}
}

// bildschluessel ist die Kennung eines Bildlaufs.
//
// **Dieselbe Gestalt wie im Fenster** (`kibild/<karteID>`), damit beide
// Seiten über dasselbe reden.
func bildschluessel(karteID int64) string {
	return "kibild/" + strconv.FormatInt(karteID, 10)
}

// KIBildAbbrechen hält den laufenden Bildauftrag einer Karte an.
//
// **Der Rückgabewert sagt, ob einer lief** — und ist kein Fehler, wenn
// keiner lief.
func (a *App) KIBildAbbrechen(karteID int64) bool {
	return a.bildlaeufe.abbrechen(karteID)
}
