package domain

import (
	"math"
	"time"
)

// Zeitspanne rechnet eine Anzahl von Einheiten in eine Dauer um — und sagt,
// ob sie hineinpasst (J05, #2269).
//
// **Eine Dauer endet bei 2⁶³ − 1 Nanosekunden**, rund 292 Jahren, und
// `time.Duration(n) * time.Hour` läuft darüber hinaus um, ohne es zu sagen:
// 5.124.096 Stunden wurden zu 25 Minuten, 9.223.372.037 Sekunden zu einer
// negativen Frist. Eine Anzahl von außen — getippt, eingestellt, aus einer
// Kopfzeile gelesen — geht deshalb hier hindurch und nicht durch die
// Multiplikation.
//
// Eine negative Anzahl und eine Einheit von null oder weniger sind keine
// Spanne; das zweite Ergebnis ist dann `false`.
func Zeitspanne(anzahl int64, einheit time.Duration) (time.Duration, bool) {
	if anzahl < 0 || einheit <= 0 {
		return 0, false
	}

	if anzahl > math.MaxInt64/int64(einheit) {
		return 0, false
	}

	return time.Duration(anzahl) * einheit, true
}

// ZeitspanneHoechstens ist die sättigende Lesart: was nicht hineinpasst, ist
// die längste Dauer, die es gibt, und nicht eine kurze.
//
// Für Angaben, die eine **Untergrenze** setzen — „warte mindestens so lange"
// —, ist das die richtige Deutung: wer mehr verlangt, als sich ausdrücken
// lässt, bekommt das meiste und nicht ein Umlaufergebnis. Eine negative
// Anzahl ist null.
func ZeitspanneHoechstens(anzahl int64, einheit time.Duration) time.Duration {
	if anzahl <= 0 || einheit <= 0 {
		return 0
	}

	if dauer, passt := Zeitspanne(anzahl, einheit); passt {
		return dauer
	}

	return math.MaxInt64
}
