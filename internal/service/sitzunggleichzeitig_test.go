package service_test

// **Zwei Rufe auf dieselbe Sache wählen einmal und nicht zweimal** — Befund 3
// des Zweitblicks über Phase 5.4.
//
// # Was der Lauf daneben misst und was nicht
//
// **`TestEineZweiteSitzungAufDieselbeSacheIstEinKonflikt` prüft einen
// bereits bestehenden ersten Lauf** — und das ist richtig, aber es ist nicht
// die Frage aus §7 Regel 4. Die lautet: was geschieht, wenn beide Rufe
// **gleichzeitig** kommen? Zwischen der Prüfung und dem Eintrag lagen Wählen
// **und** Handschlag; beide gingen hinaus, und das Gegenüber erlebte zwei
// Verbindungsaufbauten, während der zweite hier mit `ErrConflict` endete.
//
// **Den Preis eines zu späten Neins zahlt der andere.**
//
// # Wie die Gegenprobe aussieht
//
// **Sie hängt, und das ist die Aussage.** Ohne die Belegung in
// `Peersitzungen.entsteht` geht der zweite Ruf bis auf die Leitung und wartet
// dort auf einen Handschlag, den niemand annimmt — auf der anderen Seite
// steht genau ein `Annehmen`. Gemessen am 2026-09-15: mit der Belegung 1,4 s,
// ohne sie die volle Frist von 45 s.
//
// **Aufgeschrieben, weil ein hängender Lauf nicht sagt, was er gefunden
// hat.** Wer die Belegung später herausnimmt, sieht einen Verklemmer und
// sucht ihn im Lauf statt im Weg; dieser Absatz ist die Abkürzung dorthin.

import (
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der zweite Ruf kommt gar nicht bis zur Leitung.**
func TestZweiGleichzeitigeRufeWaehlenNurEinmal(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	drin := make(chan struct{})
	weiter := make(chan struct{})

	// **Festgehalten wird *in* der Naht** — ein `time.Sleep` wäre ein
	// Wettlauf mit der Maschine, und ein schneller Rechner machte diesen
	// Lauf grün, ohne etwas zu messen.
	//
	// **Und nur der erste wird festgehalten.** Wer jeden hielte, bekäme als
	// Gegenprobe einen Verklemmer statt eines Befunds — und ein Lauf, der
	// hängt, sagt nicht, was er gefunden hat.
	var (
		schranke sync.Mutex
		offen    = true
	)

	paar.hiesigerWaehler.dazwischen = func() {
		schranke.Lock()
		meiner := offen
		offen = false
		schranke.Unlock()

		if meiner {
			close(drin)
			<-weiter
		}
	}

	var (
		warten     sync.WaitGroup
		dortFehler error
		hierFehler error
	)

	warten.Add(2)

	go func() {
		defer warten.Done()

		_, dortFehler = paar.dort.Annehmen(t.Context(), paar.dortigeLeitung)
	}()

	go func() {
		defer warten.Done()

		_, hierFehler = paar.hier.Oeffne(
			t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen)
	}()

	<-drin

	// **Der zweite Ruf trifft eine Sitzung, die noch entsteht.**
	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der zweite Ruf gab %v", fehler)
	}

	close(weiter)
	warten.Wait()

	if hierFehler != nil || dortFehler != nil {
		t.Fatalf("der erste Ruf: %v / %v", hierFehler, dortFehler)
	}

	// **Und gewählt wurde genau einmal.** Das ist die Zusage: nicht „der
	// zweite bekommt einen Fehler", sondern „das Gegenüber sieht ihn nie".
	if paar.hiesigerWaehler.gerufen != 1 {
		t.Errorf("gewählt wurde %dmal", paar.hiesigerWaehler.gerufen)
	}
}
