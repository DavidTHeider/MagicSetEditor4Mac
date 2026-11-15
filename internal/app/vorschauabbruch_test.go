package app_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Vorschau wird abgebrochen, nicht abgewartet** — die Sonderregel dieses
// Projekts (§7), und bis hierher hielt sie kein Wächter.
//
// # Was schon dastand und was nicht
//
// `vorschaulauf` bricht den vorigen Lauf ab, sobald ein neuer beginnt, und die
// Anweisungen darin sind zu 100 % durchlaufen — von jedem Test, der zweimal
// zeichnet. **Zugesichert hat die Wirkung keiner.** Eine Abdeckung sagt, dass
// eine Zeile ausgeführt wurde, nicht dass jemand hinsah
// (#590).
//
// `TestDieVorschauLaesstSichAbbrechen` prüft etwas anderes: dort ist der
// **Eltern**kontext schon abgebrochen, bevor der Lauf beginnt. Das ist die
// Zusage „ein toter Kontext wird bemerkt" — nicht „ein zweiter Lauf beendet
// den ersten".
//
// # Warum das die teurere Hälfte ist
//
// Rendern kostet. Wer während des Tippens vier Läufe anstößt und keinen davon
// beendet, rechnet vier Karten und zeigt am Ende womöglich den drittletzten
// Stand. Der Abbruch ist deshalb keine Aufräumarbeit, sondern die Regel selbst.
//
// # Und die zweite Hälfte prüft derselbe Fall
//
// Der abgebrochene Lauf ruft beim Enden sein `fertig()`. Träfe das den
// inzwischen begonnenen zweiten Lauf, brächte der zuerst Fertige den nächsten
// um — die Vorschau bliebe leer, ohne dass jemand etwas falsch gemacht hat.
// Deshalb steht am Ende die Zusicherung, dass der **zweite** Lauf durchkommt.
//
// # Und alle drei Einstiege, kreuzweise
//
// `vorschaulauf` ist **geteilt**: `Vorschau`, `Probevorschau` und
// `Stilprobevorschau` melden sich an derselben Stelle an. Das ist die Absicht —
// wer von der Kartenseite auf die Stilseite wechselt, soll den laufenden
// Kartenlauf beenden und nicht zwei Karten nebeneinander rechnen.
//
// **Ein Wächter über einen Einstieg sagte darüber nichts** (§10: welche
// anderen Beispiele nehmen denselben Weg?). Deshalb neun Fälle: jeder Einstieg
// abgelöst von jedem, die Kreuzungen eingeschlossen.
func TestEineZweiteVorschauBeendetDieErste(t *testing.T) {
	einstiege := map[string]func(*app.App) app.VorschauAntwort{
		"die Kartenvorschau": func(a *app.App) app.VorschauAntwort {
			return a.Vorschau(context.Background(), "1", 1)
		},
		"die Probevorschau": func(a *app.App) app.VorschauAntwort {
			return a.Probevorschau(context.Background(), "1", "name", "Feuerball", 1)
		},
		"die Stilprobevorschau": func(a *app.App) app.VorschauAntwort {
			return a.Stilprobevorschau(context.Background(), "1", "randfarbe", "rgb(1,2,3)", 1)
		},
	}

	for ersterName, ersterLauf := range einstiege {
		for zweiterName, zweiterLauf := range einstiege {
			t.Run(ersterName+", abgelöst von "+zweiterName, func(t *testing.T) {
				t.Parallel()

				pruefeAbloesung(t, ersterLauf, zweiterLauf)
			})
		}
	}
}

// pruefeAbloesung hält einen Lauf fest und lässt ihn von einem zweiten ablösen.
func pruefeAbloesung(t *testing.T, erst, dann func(*app.App) app.VorschauAntwort) {
	t.Helper()

	bremse := &bremsprobe{
		drin:        make(chan struct{}),
		abgebrochen: make(chan struct{}),
	}

	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{}, bremse, service.NeuerPlanbauer(),
		render.NeuerRasterer(ohneBilder{}))

	erster := make(chan app.VorschauAntwort, 1)

	go func() {
		erster <- erst(anwendung)
	}()

	// **Erst wenn der erste Lauf wirklich drinsteht.** Ein zweiter Aufruf
	// davor bräche nichts ab, und der Test wäre grün, ohne seinen Fall zu
	// enthalten (§10).
	<-bremse.drin

	zweiter := dann(anwendung)

	select {
	case <-bremse.abgebrochen:
	case <-time.After(5 * time.Second):
		t.Fatal("der erste Lauf lief weiter, obwohl ein zweiter begonnen hat")
	}

	if antwort := <-erster; antwort.Fehler == nil {
		t.Error("der abgebrochene Lauf meldete keinen Fehler")
	}

	// **Der zweite darf nicht mitsterben.**
	if zweiter.Fehler != nil {
		t.Errorf("der zweite Lauf wurde mit abgeräumt: %+v", zweiter.Fehler)
	}
}

// bremsprobe hält den **ersten** Lauf fest, bis sein Kontext endet.
//
// **Nur den ersten**: der zweite muss durchkommen, sonst prüfte der Test eine
// Verklemmung statt eines Abbruchs.
type bremsprobe struct {
	mu     sync.Mutex
	laeufe int

	// drin schließt, sobald der erste Lauf im Lesen steht.
	drin chan struct{}

	// abgebrochen schließt, wenn der erste Lauf von außen beendet wurde.
	abgebrochen chan struct{}
}

func (b *bremsprobe) KarteLesen(ctx context.Context, _ int64) (*domain.CardRow, error) {
	b.mu.Lock()
	b.laeufe++
	erster := b.laeufe == 1
	b.mu.Unlock()

	if !erster {
		return &domain.CardRow{ID: 1}, nil
	}

	close(b.drin)
	<-ctx.Done()
	close(b.abgebrochen)

	return nil, ctx.Err()
}

func (*bremsprobe) Herkuenfte(context.Context, int64) (map[string]string, error) {
	return nil, nil
}
