package domain

// **Wofür eine Peerverbindung geöffnet wird** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum der Zweck hier steht und nicht im Protokoll
//
// **Weil er eine Erlaubnis ist und keine Leitung.** `teilen` und `handel`
// laufen über dasselbe Protokoll und sind zwei sehr verschiedene Dinge: das
// eine überträgt ein Objekt, das andere ändert zwei Bestände. Was ein Kontakt
// darf, ist eine fachliche Aussage — und die gehört dorthin, wo die
// Fachregeln stehen.
//
// **Und weil es sonst zwei Listen gäbe.** Der Rechtsatz eines Kontakts und die
// Zuordnung auf ein Protokoll müssten beide alle Zwecke kennen; zwei
// Aufzählungen laufen auseinander, und die erste, die es merkt, ist keine.

import "fmt"

// Zweck ist, wozu eine Peersitzung berechtigt.
type Zweck string

// Die vier Zwecke, die dieses Programm kennt.
const (
	// ZweckPaarung ist das Kennenlernen — und sonst nichts.
	//
	// **Er gestattet weder Einfuhr noch Handel.** Das Issue sagt es
	// wörtlich: „Pairing allein gestattet keinen Import oder Handel".
	ZweckPaarung Zweck = "paarung"

	// ZweckTeilen überträgt Objekte in eine Richtung.
	ZweckTeilen Zweck = "teilen"

	// ZweckHandel ändert zwei Bestände.
	ZweckHandel Zweck = "handel"

	// ZweckPartie spielt eine Partie.
	ZweckPartie Zweck = "partie"
)

// DieZwecke ist die Aufzählung in fester Reihenfolge.
//
// **Fest, weil sie an Flächen erscheint**: eine Liste, deren Reihenfolge bei
// jedem Lauf eine andere ist, lässt sich weder anzeigen noch vergleichen.
var DieZwecke = []Zweck{
	ZweckPaarung, ZweckTeilen, ZweckHandel, ZweckPartie,
}

// PruefeZweck hält einen fremden Wert gegen die Aufzählung.
func PruefeZweck(zweck Zweck) error {
	for _, gilt := range DieZwecke {
		if zweck == gilt {
			return nil
		}
	}

	return fmt.Errorf(
		"%w: %q ist kein Zweck dieses Programms", ErrValidation, zweck)
}
