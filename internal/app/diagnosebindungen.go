package app

import "context"

// **Die Bindungen der Diagnose** — Werkstück 4 von Phase 3.5
// (#1166).
//
// **Eine eigene Datei**, weil `bindungen.go` sonst über die Zeilengrenze ginge
// (§2).

// Diagnose gibt den Text, den jemand einem Fehlerbericht beilegen kann.
//
// **Er wird gezeigt, bevor er irgendwohin geht**: dieser Weg gibt eine
// Zeichenkette zurück und schickt nichts hinaus.
func (b *Bindungen) Diagnose(ctx context.Context) DiagnoseAntwort {
	return b.app(ctx).Diagnose(b.ruf(ctx))
}

// DiagnoseSichern schreibt sie in eine Datei, die jemand wählt.
func (b *Bindungen) DiagnoseSichern(ctx context.Context) SchreibAntwort {
	return b.app(ctx).DiagnoseSichern(b.ruf(ctx))
}
