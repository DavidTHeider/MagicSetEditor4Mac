// Wann eine Fassung neuer ist als eine andere.
//
// # Warum das hier steht und nicht bei der Aktualisierung
//
// Zwei Verbraucher brauchen es: die Aktualisierung und der Rückspielweg
// (#1162) —
// eine **Sicherung** trägt die Fassung, die sie geschrieben hat, und der
// Rückspielweg entscheidet daran.
//
// Ihn in `internal/aktualisierung` lesen zu lassen ginge nicht: das ist der
// Netzverbraucher für GitHub, und der Wächter über die Composition Roots lässt
// ihn nur in drei benannten Dateien zu (§2). Er hat recht — was der
// Rückspielweg braucht, ist keine Aktualisierung, sondern ein **Vergleich
// zweier Werte**. Und der gehört zu den Wertetypen.

package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Fassungsstellen ist die Zahl der Stellen einer Fassung — Haupt, Neben, Korrektur.
const Fassungsstellen = 3

// FassungIstNeuer sagt, ob die erste Fassung neuer ist als die zweite.
//
// **Stellenweise als Zahlen und nicht als Text.** `0.10.0` ist neuer als
// `0.9.0`; als Zeichenkette verglichen wäre es kleiner, weil „1" vor „9"
// steht. Genau dieser Fall steht als erster im Werkstück
// (#1100).
//
// **Ein führendes `v` gehört zum Tag, nicht zur Fassung.** GitHub nennt seine
// Freigaben `v0.1.0`, die Anwendung nennt sich `0.1.0`.
//
// **Weniger Stellen sind Nullen**: `1.0` ist dasselbe wie `1.0.0`, nicht
// weniger.
func FassungIstNeuer(eine, andere string) (bool, error) {
	dort, err := stellen(eine)
	if err != nil {
		return false, err
	}

	hier, err := stellen(andere)
	if err != nil {
		return false, err
	}

	for stelle := range Fassungsstellen {
		switch {
		case dort[stelle] > hier[stelle]:
			return true, nil
		case dort[stelle] < hier[stelle]:
			return false, nil
		}
	}

	return false, nil
}

// stellen zerlegt eine Fassung in ihre Zahlen.
//
// **Ein Fehler und keine Null.** Eine Marke, die niemand lesen kann, ist kein
// Anlass, den Nutzenden etwas zu melden — und erst recht keiner, still nichts
// zu tun (§8).
func stellen(fassung string) ([Fassungsstellen]int, error) {
	var zahlen [Fassungsstellen]int

	roh := strings.TrimPrefix(strings.TrimSpace(fassung), "v")
	if roh == "" {
		return zahlen, fmt.Errorf("%w: die Fassung ist leer", ErrValidation)
	}

	teile := strings.Split(roh, ".")
	if len(teile) > Fassungsstellen {
		return zahlen, fmt.Errorf("%w: %q hat mehr als %d Stellen",
			ErrValidation, fassung, Fassungsstellen)
	}

	for stelle, teil := range teile {
		zahl, err := strconv.Atoi(teil)
		if err != nil || zahl < 0 {
			return zahlen, fmt.Errorf("%w: %q ist keine Fassung — %q ist keine Zahl",
				ErrValidation, fassung, teil)
		}

		zahlen[stelle] = zahl
	}

	return zahlen, nil
}
