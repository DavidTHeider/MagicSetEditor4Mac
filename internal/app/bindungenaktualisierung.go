package app

import "context"

// **Die Aktualisierung am Fenster** —
// #1100.
//
// # Warum das hier nachkommt
//
// Stufe 1 stand als Verfahren auf `*App` da und war **von nirgends zu
// erreichen**: gebunden wird `*Bindungen` (`main.go`), und dort fehlte sie.
// Dieselbe Gestalt wie #852
// und #1013
// — ein Paket mit Tests, voller Abdeckung und ohne Aufrufer ist keine
// Anwendungsfunktion (§10, erste Dimension).
//
// Aufgefallen beim Bauen von Stufe 2: `bindungsaufrufer_test.go` hält jede
// **Bindung** gegen einen Aufrufer im Fenster — ein Verfahren auf `*App`
// sieht er nicht, und genau davor stand Stufe 1.

// NachNeuerFassungSehen fragt nach einer neueren Fassung.
func (b *Bindungen) NachNeuerFassungSehen(ctx context.Context) FassungsAntwort {
	return b.app(ctx).NachNeuerFassungSehen(b.ruf(ctx))
}

// NeueFassungLaden holt die neueste Freigabe ins Downloadverzeichnis.
func (b *Bindungen) NeueFassungLaden(ctx context.Context) LadeAntwort {
	return b.app(ctx).NeueFassungLaden(b.ruf(ctx))
}
