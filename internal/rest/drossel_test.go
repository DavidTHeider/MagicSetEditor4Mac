package rest_test

import (
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Die Drossel lässt eine benannte Zahl durch und dann nicht mehr**
// (#997).
//
// # Wovor sie schützt
//
// Nicht vor Überlastung — ein Dienst mit einem Aufrufer hat keine. Sondern vor
// dem **Raten der Marke**: ohne Grenze könnte ein lokaler Prozess sie in
// beliebiger Geschwindigkeit durchprobieren, und die Länge der Marke wäre die
// einzige Hürde.
//
// # Warum die Uhr eingesetzt wird
//
// Sonst bräuchte jeder Test eine echte Wartezeit, und ein Lauf, der auf Uhren
// wartet, ist einer, der unter Last flackert (§3, §16).
func TestDieDrosselLaesstIhreZahlDurch(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	drossel := rest.NeueDrossel(3, time.Minute, func() time.Time { return jetzt })

	for versuch := 1; versuch <= 3; versuch++ {
		if !drossel.Durchlassen() {
			t.Fatalf("Versuch %d wurde abgewiesen, obwohl drei erlaubt sind", versuch)
		}
	}

	if drossel.Durchlassen() {
		t.Error("der vierte Versuch kam durch")
	}
}

// **Und nach der Frist wieder.**
//
// Ohne diesen Fall wäre eine Drossel grün, die nach der ersten Grenze für
// immer zumacht — und das wäre keine Drossel, sondern ein Riegel.
func TestNachDerFristGehtEsWeiter(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	drossel := rest.NeueDrossel(1, time.Minute, func() time.Time { return jetzt })

	if !drossel.Durchlassen() {
		t.Fatal("der erste Versuch wurde abgewiesen")
	}

	if drossel.Durchlassen() {
		t.Fatal("der zweite kam durch, obwohl einer erlaubt ist")
	}

	// **Eine Sekunde vor Ablauf noch nicht.** Ohne diesen Zwischenschritt wäre
	// auch eine Drossel grün, die die Frist gar nicht liest.
	jetzt = jetzt.Add(59 * time.Second)

	if drossel.Durchlassen() {
		t.Error("vor Ablauf der Frist kam einer durch")
	}

	jetzt = jetzt.Add(2 * time.Second)

	if !drossel.Durchlassen() {
		t.Error("nach der Frist blieb es zu")
	}
}

// **Sie hält auch, wenn viele gleichzeitig fragen.**
//
// # Warum das hierher gehört
//
// `net/http` bedient jede Anfrage in einer eigenen Goroutine. Eine Drossel
// ohne Sperre ließe unter Last mehr durch als ihre Zahl — und genau dann, wenn
// jemand rät, kämen viele gleichzeitig.
//
// **Gezählt wird das Ergebnis, nicht die Zeit.** Von 100 gleichzeitigen
// Versuchen bei einer Grenze von 10 dürfen genau 10 durchkommen.
func TestDieDrosselHaeltUnterLast(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	drossel := rest.NeueDrossel(10, time.Minute, func() time.Time { return jetzt })

	var (
		warten      sync.WaitGroup
		schranke    sync.Mutex
		durchkommen int
	)

	for range 100 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			if drossel.Durchlassen() {
				schranke.Lock()
				durchkommen++
				schranke.Unlock()
			}
		}()
	}

	warten.Wait()

	if durchkommen != 10 {
		t.Errorf("%d von 100 kamen durch, erlaubt waren 10", durchkommen)
	}
}
