package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **`0.10.0` ist neuer als `0.9.0`** — der Fall, an dem ein Vergleich über
// Zeichenketten scheitert, und deshalb der erste, den das Werkstück nennt
// (#1100).
func TestWelcheFassungNeuerIst(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		eine, andere string
		neuer        bool
	}{
		// **Der Fall, um den es geht.** Als Text ist „0.10.0" kleiner als
		// „0.9.0", weil „1" vor „9" steht.
		{"0.10.0", "0.9.0", true},
		{"0.9.0", "0.10.0", false},

		{"0.2.0", "0.1.0", true},
		{"1.0.0", "0.99.99", true},
		{"0.2.1", "0.2.0", true},
		{"0.2.0", "0.2.0", false},
		{"0.2.0", "0.2.1", false},

		// **Ein führendes `v` gehört zum Tag, nicht zur Fassung.** GitHub
		// nennt seine Freigaben `v0.1.0`; die Anwendung nennt sich `0.1.0`.
		{"v0.2.0", "0.1.0", true},
		{"0.2.0", "v0.1.0", true},
		{"v0.2.0", "v0.2.0", false},

		// **Weniger Stellen sind Nullen.** `1.0` ist nicht neuer als `1.0.0`.
		{"1.0", "1.0.0", false},
		{"1.0.1", "1.0", true},
	} {
		t.Run(fall.eine+" gegen "+fall.andere, func(t *testing.T) {
			t.Parallel()

			neuer, err := domain.FassungIstNeuer(fall.eine, fall.andere)
			if err != nil {
				t.Fatalf("vergleichen: %v", err)
			}

			if neuer != fall.neuer {
				t.Errorf("%q neuer als %q: %v, erwartet %v",
					fall.eine, fall.andere, neuer, fall.neuer)
			}
		})
	}
}

// **Was keine Fassung ist, wird abgewiesen und nicht geraten.**
//
// Eine Freigabe, deren Marke niemand lesen kann, ist kein Anlass, den
// Nutzenden etwas zu melden — und erst recht keiner, still nichts zu tun.
func TestEineUnlesbareFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, fall := range []string{"", "v", "keine", "1.2.3.4", "1.-2.0", "1.x.0"} {
		if _, err := domain.FassungIstNeuer(fall, "0.1.0"); err == nil {
			t.Errorf("%q gilt als Fassung", fall)
		}

		if _, err := domain.FassungIstNeuer("0.1.0", fall); err == nil {
			t.Errorf("%q gilt als Fassung (als zweite)", fall)
		}
	}
}

// **Die Abweisung einer fremden Archivfassung sagt beides** — was in der
// Sicherung steht und was diese Anwendung ist
// (#1162).
//
// **Zwei Sätze und nicht einer**: eine Sicherung **ohne** Angabe ist etwas
// anderes als eine mit einer fremden, und „nennt die Fassung \"\"" wäre die
// Auskunft, die niemandem hilft.
func TestEineFremdeArchivfassungSagtBeides(t *testing.T) {
	t.Parallel()

	mitAngabe := &domain.FremdeArchivfassung{Dort: "99.0.0", Hier: "0.3.0"}

	if !strings.Contains(mitAngabe.Error(), "99.0.0") ||
		!strings.Contains(mitAngabe.Error(), "0.3.0") {
		t.Errorf("der Satz nennt nicht beide Fassungen: %q", mitAngabe.Error())
	}

	ohneAngabe := &domain.FremdeArchivfassung{Hier: "0.3.0"}

	if strings.Contains(ohneAngabe.Error(), `""`) {
		t.Errorf("der Satz nennt eine leere Fassung: %q", ohneAngabe.Error())
	}

	if !strings.Contains(ohneAngabe.Error(), "keine Fassung") {
		t.Errorf("der Satz sagt nicht, dass keine dasteht: %q", ohneAngabe.Error())
	}

	// **Und sie ist ein Sicherungsfehler.** Daran hängt, welcher Satz im
	// Fenster steht: `Uebersetze` prüft die Sicherungsausgänge **vor** den
	// allgemeinen, und ohne diese Zuordnung landete sie beim Rückfall.
	for _, fehler := range []error{mitAngabe, ohneAngabe} {
		if !errors.Is(fehler, domain.ErrSicherung) {
			t.Errorf("%v ist kein Sicherungsfehler", fehler)
		}
	}
}
