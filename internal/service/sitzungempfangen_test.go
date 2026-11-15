package service_test

// **Was an einer offenen Sitzung hereinkommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum es diesen Weg erst jetzt gibt
//
// **Weil bis dahin niemand gelauscht hat.** N1 und N2 haben das Senden gebaut
// und `Empfangen` daneben gestellt — die Frage „gilt das noch?" ohne jemanden,
// der sie stellt. Wer die Partie über eine Leitung führen will, braucht die
// andere Richtung, und sie gehört an dieselbe Stelle: der Griff ist der
// einzige Ort, an dem Epoche und Recht zusammen bekannt sind.
//
// # Die Zusage, für die es diese Läufe gibt
//
// **Der Stand wird nach dem Lesen geprüft und nicht davor** (§7, Regel 2).
// Zwischen dem Hinsetzen und der Nachricht liegt beliebig viel Zeit; gefragt
// ist, ob die Sitzung gilt, **wenn die Nachricht da ist**. Der zweite Lauf
// hier ist genau diese Gegenprobe — er nimmt das Recht weg, **nachdem** die
// Nachricht unterwegs ist.

import (
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was gesendet wurde, kommt am anderen Ende an.**
func TestWasGesendetWurdeKommtAmAnderenEndeAn(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	hier, dort := paar.eineSitzung(t, domain.ZweckPartie)

	fertig := unterwegs(t, hier, []byte(`{"art":"ereignis"}`))
	defer fertig()

	angekommen, fehler := dort.Empfange(t.Context())
	if fehler != nil {
		t.Fatalf("Empfange: %v", fehler)
	}

	if string(angekommen) != `{"art":"ereignis"}` {
		t.Errorf("es kam %q an", angekommen)
	}
}

// **Ein entzogenes Recht hält die Nachricht zurück — nachdem sie da ist.**
//
// **Das ist der Punkt und nicht bloß eine zweite Zusicherung**: die Nachricht
// liegt schon auf der Leitung, als das Recht fällt. Eine Prüfung **vor** dem
// Lesen ließe sie durch, weil sie zu einem Zeitpunkt fragte, an dem noch alles
// stimmte — und genau davor warnt §7, Regel 2.
func TestEinEntzogenesRechtHaeltDieNachrichtZurueck(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	hier, dort := paar.eineSitzung(t, domain.ZweckPartie)

	fertig := unterwegs(t, hier, []byte(`{"art":"ereignis"}`))
	defer fertig()

	// **Erst jetzt fällt das Recht.** Die Nachricht ist längst unterwegs.
	paar.dortigeRechte.darf = false

	angekommen, fehler := dort.Empfange(t.Context())

	switch {
	case !errors.Is(fehler, domain.ErrValidation):
		t.Fatalf("es kam %v", fehler)
	case angekommen != nil:
		t.Errorf("die Nachricht kam trotzdem durch: %q", angekommen)
	}
}

// **Und was die Leitung klagt, klagt der Griff weiter.**
//
// **Ohne Deutung**: eine geschlossene Sitzung ist kein Rechtsproblem, und wer
// beides gleich meldete, schickte den Leser auf die falsche Fährte (§8).
func TestEineGeschlosseneSitzungEmpfaengtNichtsMehr(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	_, dort := paar.eineSitzung(t, domain.ZweckPartie)

	if fehler := dort.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if _, fehler := dort.Empfange(t.Context()); fehler == nil {
		t.Error("eine geschlossene Sitzung hat etwas empfangen")
	}
}

// unterwegs schickt eine Nachricht nebenher hinaus.
//
// **Zwei Rufe und nicht einer**: `Sende` reiht ein, `Leere` schickt. Der Grund
// steht bei `service.Sitzungsgriff.Leere` — eine Goroutine bräuchte eine Uhr,
// die niemand stellt, und deshalb startet sie hier der Lauf.
//
// **Nebenher, weil die Leitung ohne Puffer ist**: `Leere` kommt erst zurück,
// wenn die Gegenseite gelesen hat. Wer sie in derselben Goroutine riefe wie
// `Empfange`, wartete auf sich selbst.
func unterwegs(
	t *testing.T, griff *service.Sitzungsgriff, nachricht []byte,
) func() {
	t.Helper()

	var (
		warten sync.WaitGroup
		fehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		if fehler = griff.Sende(
			t.Context(), nachricht, service.Zuverlaessig,
		); fehler != nil {
			return
		}

		fehler = griff.Leere(t.Context())
	}()

	return func() {
		warten.Wait()

		// **Der Fehler wird gelesen und nicht bloß gesammelt.** Ein Senden,
		// das still scheitert, machte aus jedem Lauf hier eine Aussage über
		// eine Nachricht, die es nie gab (§10).
		if fehler != nil {
			t.Errorf("das Senden nebenher: %v", fehler)
		}
	}
}
