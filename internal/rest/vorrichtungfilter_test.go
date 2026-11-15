// **Der Schnellfilter am willigen Doppel** — Werkstück U5 von Phase 5.1
// (#1416),
// in eigener Datei, weil `vorrichtungbestand_test.go` an der Zeilengrenze
// steht (§2). Das Doppel nennt genau eine Karte: die Zwei, die auch die
// Kartenliste führt.
package rest_test

import "context"

func (williger) KartenFiltern(context.Context, int64, string) ([]int64, error) {
	return []int64{2}, nil
}
