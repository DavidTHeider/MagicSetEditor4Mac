package domain_test

// **Was nach einem Abriss fehlt** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass zwei Seiten sich über den Abdruck einigen und nicht über die Zahl.**
// „Wir sind beide bei Revision 3" sagt nichts, solange nicht feststeht, dass
// es **dieselbe** Revision 3 ist — zwei Verläufe können auseinandergelaufen
// sein, und genau dann darf nichts nachgeliefert werden.
//
// **Dass ein widersprüchlicher Verlauf offen bleibt.** Ihn stillschweigend zu
// überschreiben hieße, einen Stand zu verlieren, von dem niemand mehr weiß.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einFortschritt baut, was eine Seite als bestätigt kennt.
func einFortschritt(anders func(*domain.Handelsfortschritt)) domain.Handelsfortschritt {
	fortschritt := domain.Handelsfortschritt{
		Vorgang: "v-1", Revision: 3, Abdruck: "abdruck-3", Folge: 7,
	}

	if anders != nil {
		anders(&fortschritt)
	}

	return fortschritt
}

func TestWasNachEinemAbrissFehlt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		meiner, seiner domain.Handelsfortschritt
		abFolge        int64
		klaerung       bool
	}{
		// **Gleichauf** — nichts fehlt.
		"gleichauf": {einFortschritt(nil), einFortschritt(nil), 8, false},

		// **Er hinkt hinterher** — ab seiner nächsten Folge wird geliefert.
		"er ist zurück": {
			einFortschritt(nil),
			einFortschritt(func(f *domain.Handelsfortschritt) {
				f.Revision, f.Abdruck, f.Folge = 2, "abdruck-2", 4
			}),
			5, false,
		},

		// **Er ist weiter** — dann fehlt **mir** etwas, und nachzuliefern
		// habe ich nichts. Die Folge ist trotzdem seine plus eins: was ich
		// danach noch habe, geht hinaus.
		"er ist weiter": {
			einFortschritt(func(f *domain.Handelsfortschritt) {
				f.Revision, f.Abdruck, f.Folge = 2, "abdruck-2", 4
			}),
			einFortschritt(nil),
			8, false,
		},

		// **Dieselbe Revision, ein anderer Abdruck** — die Verläufe sind
		// auseinandergelaufen. Hier wird nichts nachgeliefert.
		"derselbe Stand, ein anderer Abdruck": {
			einFortschritt(nil),
			einFortschritt(func(f *domain.Handelsfortschritt) {
				f.Abdruck = "ein anderer"
			}),
			0, true,
		},

		// **Und ein fremder Vorgang ist keine Wiederaufnahme.**
		"ein fremder Vorgang": {
			einFortschritt(nil),
			einFortschritt(func(f *domain.Handelsfortschritt) { f.Vorgang = "v-9" }),
			0, true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			abFolge, klaerung := domain.Nachzuliefern(fall.meiner, fall.seiner)

			if klaerung != fall.klaerung {
				t.Fatalf("Klärung ist %v", klaerung)
			}

			if abFolge != fall.abFolge {
				t.Errorf("geliefert wird ab %d, erwartet war %d", abFolge, fall.abFolge)
			}
		})
	}
}

// **Ein Abdruck, den eine Seite gar nicht kennt, ist keine Klärung.**
//
// **Sondern ein Anfang**: wer bei Revision null steht, hat noch nichts
// bestätigt, und dann gibt es auch nichts, was widersprechen könnte.
func TestEinAnfangIstKeineKlaerung(t *testing.T) {
	t.Parallel()

	anfang := domain.Handelsfortschritt{Vorgang: "v-1"}

	abFolge, klaerung := domain.Nachzuliefern(einFortschritt(nil), anfang)

	if klaerung {
		t.Fatal("ein Anfang gilt als Klärung")
	}

	if abFolge != 1 {
		t.Errorf("geliefert wird ab %d", abFolge)
	}
}
