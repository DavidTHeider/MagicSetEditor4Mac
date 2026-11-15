package domain

// **Wie lange ein Nutzer nach Fehlversuchen warten muss** — SEC-10
// (#2151,
// Vermerk 0178).
//
// # Die Zahlen
//
// **Fünf Versuche frei, danach eine Sperre, die sich mit jedem weiteren
// Fehlversuch verdoppelt** — 30 Sekunden, eine Minute, zwei, … bis zu einer
// Stunde. Wer sich vertippt, merkt davon nichts; wer rät, kommt nach der
// Sättigung auf 24 Versuche am Tag. Bei sechs Ziffern und dem Fenster ±1
// sind das drei passende von einer Million je Versuch.
//
// **Eine Entscheidung dieses Projekts, und sie weicht ab**: NIST SP 800-63B
// §5.2.2 verlangt, aufeinanderfolgende Fehlversuche je Konto auf höchstens
// hundert zu begrenzen. Hier gibt es keine solche Obergrenze — nach jeder
// Sperre geht wieder ein Versuch. Der Grund steht in Vermerk 0178: eine
// endgültige Sperre bräuchte einen Weg zurück, den diese Anwendung ohne
// Verwalter nicht hat.

import "time"

// Faktorgrenze ist der Versuch, mit dem die Sperre beginnt.
const Faktorgrenze = 5

// Die Dauer der ersten Sperre und die längste.
const (
	ersteFaktorsperre    = 30 * time.Second
	laengsteFaktorsperre = time.Hour
)

// Faktorsperre sagt, wie lange nach dem `fehlversuche`-ten Versuch gesperrt
// ist — null unter der Grenze.
func Faktorsperre(fehlversuche int) time.Duration {
	if fehlversuche < Faktorgrenze {
		return 0
	}

	sperre := ersteFaktorsperre
	for range fehlversuche - Faktorgrenze {
		sperre *= 2
		if sperre >= laengsteFaktorsperre {
			return laengsteFaktorsperre
		}
	}

	return sperre
}

// FaktorGesperrt sagt, dass Codes bis `bis` nicht geprüft werden.
//
// **Am Feld des Codes**, denn dort wartet die Fläche; der Zeitpunkt reist
// mit, damit sie ihn nennen kann.
func FaktorGesperrt(bis time.Time) error {
	return NewFieldError("faktor", SchluesselFaktorGesperrt,
		"Zu viele falsche Codes — der nächste Versuch geht ab %s.",
		bis.Local().Format("15:04:05")).
		MitWerten(map[string]string{"bis": bis.UTC().Format(time.RFC3339)})
}
