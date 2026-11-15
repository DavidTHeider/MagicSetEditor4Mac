package service_test

import (
	"math/rand/v2"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Zwei Portierungen von `smart_compare`, eine Ordnung** —
// #2006.
//
// `domain.Vergleiche` ordnet Kartenlisten und Statistiken,
// `msescript.KlugVergleiche` die Schlüssel von `sort_list`. Beide portieren
// dieselbe Funktion (`src/util/string.cpp`, Zeilen 208–318, Fassung f9d9356
// vom 2021-07-06), und beide dürfen einander nicht kennen (§2). **Dieser Lauf
// ist der Preis dafür**: er steht hier, weil `service` das einzige Paket ist,
// das beide sieht.
//
// # Was er hält
//
//   - Jedes Paar aus Zeichenketten bis zur Länge zwei über ein Alphabet, das
//     jeden Zweig berührt — Ziffern, Groß und klein, beide Tabellen, die
//     zerfallenden Buchstaben, die zwei bestimmten Punkte, Steuerzeichen.
//   - Zwanzigtausend zufällige Paare bis zur Länge sechs, mit festem Samen.
//
// # Was er nicht hält
//
// Ob **eine** der beiden MSE richtig nachbildet — das halten ihre eigenen
// Tabellen (`internal/domain/ordnung_test.go`,
// `internal/msescript/klugvergleich_test.go`).
func TestBeidePortierungenOrdnenGleich(t *testing.T) {
	t.Parallel()

	alphabet := []rune{
		'a', 'A', 'e', 'b', 'z', '0', '1', '2', '9', '/', ' ', 0x01,
		'Æ', 'Ä', 'Œ', 'ĳ', 'ǅ', 'Ł', 'ó', '×', 'ſ', 'ƀ', 'Ḁ', 'ẹ',
	}

	ketten := make([]string, 0, 1+(1+len(alphabet))*len(alphabet))
	ketten = append(ketten, "")
	for _, erstes := range alphabet {
		ketten = append(ketten, string(erstes))
		for _, zweites := range alphabet {
			ketten = append(ketten, string([]rune{erstes, zweites}))
		}
	}

	for _, links := range ketten {
		for _, rechts := range ketten {
			vergleicheBeide(t, links, rechts)
		}
	}

	zufall := rand.New(rand.NewPCG(2006, 1)) //nolint:gosec // ein fester Samen, damit der Lauf wiederholbar ist
	kette := func() string {
		zeichen := make([]rune, zufall.IntN(7))
		for stelle := range zeichen {
			zeichen[stelle] = alphabet[zufall.IntN(len(alphabet))]
		}

		return string(zeichen)
	}

	for range 20000 {
		vergleicheBeide(t, kette(), kette())
	}
}

// vergleicheBeide meldet ein Paar, das die beiden Portierungen verschieden ordnen.
func vergleicheBeide(t *testing.T, links, rechts string) {
	t.Helper()

	if dort, hier := domain.Vergleiche(links, rechts), msescript.KlugVergleiche(links, rechts); dort != hier {
		t.Errorf("%q gegen %q: domain.Vergleiche %d, msescript.KlugVergleiche %d",
			links, rechts, dort, hier)
	}
}
