package leitung_test

// **Was eine gesicherte Leitung über ihr Gegenüber weiß** — SEC-16 und SEC-15
// (#2145,
// #2144).
//
//   - **Die Kennung aus dem Zeugnis**, auf beiden Seiten — sie darf dem
//     Handschlag nicht widersprechen (`service.eineIdentitaet`).
//   - **Je Verbindung ein anderer Kanal**, auch zwischen denselben beiden:
//     darauf ruht der Schutz gegen ein Wiedereinspielen, wenn die
//     Nonceliste eine Zufallszahl schon vergessen hat (Vermerk 0176).

import (
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

func TestEineGesicherteLeitungKenntIhrGegenueberUndIhrenKanal(t *testing.T) {
	t.Parallel()

	hiesige, dortige := zweiIdentitaeten(t)
	horcher := einHorcher(t, dortige)

	verbinde := func() (*leitung.Leitung, *leitung.Leitung) {
		var (
			warten     sync.WaitGroup
			genommen   *leitung.Leitung
			nimmfehler error
		)

		warten.Add(1)

		go func() {
			defer warten.Done()

			genommen, nimmfehler = horcher.Nimm(t.Context())
		}()

		gewaehlt, fehler := leitung.Waehle(t.Context(), horcher.Adresse(), hiesige, dortige.Kennung())

		warten.Wait()

		if fehler != nil || nimmfehler != nil {
			t.Fatalf("Waehle: %v, Nimm: %v", fehler, nimmfehler)
		}

		t.Cleanup(func() {
			_ = gewaehlt.Schliesse() //nolint:errcheck // nur ein Lauf
			_ = genommen.Schliesse() //nolint:errcheck // nur ein Lauf
		})

		return gewaehlt, genommen
	}

	erste, ersteDort := verbinde()

	if erste.Gegenstelle() != dortige.Kennung() || ersteDort.Gegenstelle() != hiesige.Kennung() {
		t.Errorf("gesehen: wählend %s, horchend %s", erste.Gegenstelle(), ersteDort.Gegenstelle())
	}

	zweite, _ := verbinde()

	if erste.Kanal() == zweite.Kanal() {
		t.Errorf("zwei Verbindungen, ein Kanal: %s", erste.Kanal())
	}

	// **Eine Röhre im selben Prozess hat kein Zeugnis.**
	roehre, _ := leitung.Schleife("kanal-roehre")
	if roehre.Gegenstelle() != "" {
		t.Errorf("die Röhre nennt %s", roehre.Gegenstelle())
	}
}
