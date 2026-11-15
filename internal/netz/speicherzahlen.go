package netz

// **Wie oft der Speicher antwortet** — OPS-05.AK-01
// (#1722).
//
// **Über den ganzen Prozess und nicht je Speicher**, aus demselben Grund wie
// das Messwerk in `service`: gefragt wird, was **dieser Prozess** aus dem
// Speicher beantwortet hat, und ein Zähler je Speicher müsste von jeder
// Wurzel durchgereicht werden, die einen baut. Vier Zahlen, keine Schlüssel —
// ein Zähler je Eintrag wüchse mit den Karten.

import "sync/atomic"

// Speicherstand sind die Zahlen der Speicher dieses Prozesses.
type Speicherstand struct {
	// Treffer: ein frischer Stand lag da, und niemand fragte das Netz.
	Treffer int64 `json:"treffer"`

	// Geholt: es wurde gefragt, und die Antwort liegt jetzt da.
	Geholt int64 `json:"geholt"`

	// Veraltet: das Netz scheiterte, und der letzte Stand sprang ein (§16,
	// Satz 5).
	Veraltet int64 `json:"veraltet"`

	// Fehlgeschlagen: das Netz scheiterte, und es gab keinen Stand.
	Fehlgeschlagen int64 `json:"fehlgeschlagen"`
}

var speicherzaehler struct {
	treffer, geholt, veraltet, fehlgeschlagen atomic.Int64
}

// Speicherzahlen gibt den Stand seit dem Start des Prozesses.
func Speicherzahlen() Speicherstand {
	return Speicherstand{
		Treffer:        speicherzaehler.treffer.Load(),
		Geholt:         speicherzaehler.geholt.Load(),
		Veraltet:       speicherzaehler.veraltet.Load(),
		Fehlgeschlagen: speicherzaehler.fehlgeschlagen.Load(),
	}
}
