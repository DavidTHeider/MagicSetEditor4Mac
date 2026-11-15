package app

// **Der Tausch des Bestands und sein Tor** — A02 und A05
// (#2216,
// #2219).
//
// Zurückspielen, Verschlüsseln und Neuanfang tauschen den Bestand unter der
// laufenden Anwendung aus: schließen, allein sein, tauschen, öffnen — und
// wenn das Öffnen scheitert, zurückdrehen und noch einmal öffnen. **Das Tor
// steht die ganze Zeit auf [torTausch]** und geht erst wieder auf, wenn der
// ganze Vorgang durch ist ([App.imTausch]). Bis zum Zweitblick über Phase A
// ging es mit dem ersten Öffnen auf, und eine Rücknahme danach lief bei
// offenem Tor.
//
// Aus `lebenszyklus.go` herausgelöst, weil der Schnitt einer Frage folgt (§2):
// dort, wie die Anwendung endet, hier, wie sie ihren Bestand wechselt.

import (
	"context"
	"errors"
)

// imTausch führt einen Tausch des Bestands aus — **exklusiv**: steht das Tor
// nicht offen, weil schon getauscht wird oder das Programm endet, wird
// abgewiesen, bevor etwas angefasst ist (Befund 2 des zweiten Durchgangs des
// Zweitblicks über Phase A: zwei Aufrufe passierten das offene Tor, und der
// zweite räumte mit ab).
func (a *App) imTausch(tausch func() error) error {
	if !a.lauf.tor.CompareAndSwap(torOffen, torTausch) {
		if a.lauf.tor.Load() == torEndet {
			return errAnwendungEndet
		}

		return errBestandWirdGetauscht
	}

	// **Der ganze Rahmen hält `wechsel`** — Restbefund des Nachlaufs über
	// Phase A: das Schließen zum Tausch nahm die Sperre nicht, und ein Ende
	// in dieser Zeit räumte gleichzeitig ab. So wartet ein Ende den ganzen
	// Tausch ab, in seiner Frist ([App.Schliessen]).
	a.lauf.wechsel.Lock()
	defer a.lauf.wechsel.Unlock()

	defer a.torWiederOeffnen()

	// **Kam das Ende, während der Tausch auf die Sperre wartete, fasst er
	// nichts mehr an** — sonst tauschte er Dateien unter einem Programm, das
	// gerade endet.
	if a.lauf.tor.Load() == torEndet {
		return errAnwendungEndet
	}

	// **Auch ein gescheiterter Tausch zählt**: er hat den Bestand geschlossen
	// und wieder geöffnet, und was vorher galt, ist nicht mehr sicher dasselbe.
	defer a.lauf.epoche.Add(1)

	return tausch()
}

// zumTauschSchliessen schließt den Bestand, damit er getauscht werden kann —
// **mit denselben Schritten wie der Schluss**, das Tor aber auf [torTausch].
//
// **Scheitert das Schließen, wird gleich wieder geöffnet** — Befund des
// Zweitblicks über Phase A: [App.abraeumen] schließt den Bestand auch dann,
// wenn die eigenen Läufe ihre Frist überschreiten, und der Bestandsweg bricht
// bei einem Fehler ab, ohne zu öffnen.
func (a *App) zumTauschSchliessen() error {
	if fehler := a.schliessenFuerTausch(); fehler != nil {
		return errors.Join(fehler, a.wiederOeffnen(context.Background()))
	}

	return nil
}

// schliessenFuerTausch stellt das Tor auf [torTausch], wo es offen stand,
// und räumt ab — in der [Schlussfrist].
func (a *App) schliessenFuerTausch() error {
	a.lauf.tor.CompareAndSwap(torOffen, torTausch)

	ctx, aufhoeren := context.WithTimeout(context.Background(), Schlussfrist)
	defer aufhoeren()

	return a.abraeumen(ctx, nil)
}

// wiederOeffnen öffnet den Bestand während eines Tauschs — **das Tor bleibt,
// wie es steht**; es öffnet [App.imTausch], wenn alles durch ist. **Die
// Sperre `wechsel` hält der Rahmen**, nicht dieser Schritt.
//
// **Nach dem Beginn des Schlusses wird nicht mehr geöffnet**, und ein
// Schluss, der während des Öffnens kam und nicht warten konnte, findet den
// Bestand danach geschlossen: das Wiederöffnen schließt dann selbst.
func (a *App) wiederOeffnen(ctx context.Context) error {
	if a.lauf.tor.Load() == torEndet {
		return errAnwendungEndet
	}

	fehler := a.Starten(ctx)

	// **Und dann ist auch die Sitzung zu Ende** — der Schluss, der nicht
	// warten konnte, hat ihre Marke stehen lassen.
	if a.lauf.tor.Load() == torEndet {
		return errors.Join(fehler, errAnwendungEndet, a.bestandSchliessen(), a.sitzungBeenden())
	}

	return fehler
}

// torWiederOeffnen öffnet das Tor nach einem Tausch — **nur dann**.
func (a *App) torWiederOeffnen() { a.lauf.tor.CompareAndSwap(torTausch, torOffen) }

// abbrechen räumt einen halben Start ab — **und lässt das Tor, wie es vorher
// stand**: nach einem gescheiterten Start bleibt die Anwendung für den
// Rückweg da (Startbefund, Sicherung zurückspielen, #962), mitten in einem
// Tausch bleibt es beim Tausch, und ein Schluss bleibt ein Schluss.
//
// **Ohne neuen Start**, auch wenn das Schließen scheitert: `abbrechen` läuft
// mitten in einem Start, und ein zweiter dort riefe sich selbst.
func (a *App) abbrechen(grund error) error {
	vorher := a.lauf.tor.Load()
	fehler := a.schliessenFuerTausch()

	if vorher == torOffen {
		a.torWiederOeffnen()
	}

	return errors.Join(grund, fehler)
}
