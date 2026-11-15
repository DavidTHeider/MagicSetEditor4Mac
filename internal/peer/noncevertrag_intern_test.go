package peer

// **Der Vertrag der Nonceliste** — SEC-15
// (#2144,
// Vermerk 0176).
//
// Entschieden ist **die begrenzte Zusatzwache und keine strikte Frist**: über
// der Grenze geht die älteste Zufallszahl, auch vor ihrem Ablauf. Eine
// strikte Frist wiese bei voller Liste jedes neue Angebot ab — dann sperrte
// ein Fremder mit 100.000 Angeboten in einer Stunde jeden Kontakt aus. Der
// Schutz gegen ein Wiedereinspielen ruht deshalb auf der Kanalbindung (je
// Verbindung ein anderer Kanal, `leitung/gegenstelle_test.go`); die Liste
// fängt dasselbe Angebot auf **derselben** Verbindung, solange sie es hält.

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDieNoncelisteAnIhrerGrenze(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		eingetragen int
		ersteNoch   bool
	}{
		"eins unter der Grenze": {eingetragen: 2, ersteNoch: true},
		"genau auf der Grenze":  {eingetragen: 3, ersteNoch: true},
		"eins darüber":          {eingetragen: 4, ersteNoch: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jetzt := time.Date(2026, 10, 7, 3, 0, 0, 0, time.UTC)
			liste := NeueNonceliste(func() time.Time { return jetzt }, time.Hour)
			liste.hoechstens = 3

			for nummer := range fall.eingetragen {
				if fehler := liste.Merke(fmt.Sprintf("n%d", nummer)); fehler != nil {
					t.Fatalf("n%d: %v", nummer, fehler)
				}
			}

			// **Die erste ist noch eine Wiederholung** — bis die Grenze sie
			// verdrängt.
			if abgewiesen := liste.Merke("n0") != nil; abgewiesen != fall.ersteNoch {
				t.Errorf("die erste abgewiesen: %v, erwartet %v", abgewiesen, fall.ersteNoch)
			}
		})
	}
}

// **Genau auf der Frist gilt sie noch** — erst danach geht sie.
func TestDieFristIstEinschliesslich(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 7, 3, 0, 0, 0, time.UTC)
	liste := NeueNonceliste(func() time.Time { return jetzt }, time.Hour)

	if fehler := liste.Merke("n"); fehler != nil {
		t.Fatal(fehler)
	}

	jetzt = jetzt.Add(time.Hour)
	if liste.Merke("n") == nil {
		t.Error("auf der Frist wurde dieselbe Zahl genommen")
	}

	jetzt = jetzt.Add(time.Nanosecond)
	if fehler := liste.Merke("n"); fehler != nil {
		t.Errorf("nach der Frist: %v", fehler)
	}
}

// **Zwei, die gleichzeitig dieselbe Zahl bringen: einer kommt durch.**
func TestGleichzeitigKommtEineZahlNurEinmalDurch(t *testing.T) {
	t.Parallel()

	liste := NeueNonceliste(time.Now, time.Hour)

	var (
		warten  sync.WaitGroup
		schloss sync.Mutex
		durch   int
	)

	for range 32 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			if liste.Merke("dieselbe") == nil {
				schloss.Lock()
				durch++
				schloss.Unlock()
			}
		}()
	}

	warten.Wait()

	if durch != 1 {
		t.Errorf("%d kamen durch", durch)
	}
}
