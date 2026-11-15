package app

import (
	"context"
	"fmt"
)

// panikFehler macht aus einer gefangenen Panik den Fehler, den das Fenster
// zeigt — F02 (#2341).
//
// **Die Klasse ist »unbekannt«, mit der Kennung dieses Vorgangs**: eine Panik
// ist ein Fehler dieses Programms und keiner der Eingabe. Der Satz ist der
// Rückfallsatz; der Wert der Panik steht nur im Protokoll ([Uebersetze]
// schreibt ihn dorthin) und nie im Fenster — ein Paketpfad oder ein
// englischer Laufzeitsatz hilft dort niemandem.
func panikFehler(ctx context.Context, wert any) *Fehler {
	return Uebersetze(ctx, fmt.Errorf("eine Panik: %v", wert))
}
