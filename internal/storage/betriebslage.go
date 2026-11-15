package storage

// **Die Lage des Bestands: antwortet er, wandert er** — Werkstück O9 von
// Phase 5.2 (#1433).
//
// # Warum ein eigener Typ und nicht zwei Funktionen
//
// **Weil die Wanderungsmarke einen Ort braucht.** Ob gerade migriert wird,
// steht in keiner Tabelle — es ist ein Zustand **dieses Prozesses**, und er
// gehört zu der Verbindung, um die es geht. Ein Paketzustand daneben wäre
// derselbe für zwei geöffnete Bestände.
//
// # Was die Probe ist
//
// **`SELECT 1`, und sonst nichts.** Sie beantwortet „bist du da und
// antwortest du" — nicht „ist alles in Ordnung". Eine Prüfung, die Zeilen
// läse, wäre bei jedem Aufruf teurer als die Arbeit, die sie schützen soll.

import (
	"context"
	"database/sql"
	"sync/atomic"
)

// Betriebslage beantwortet, was `service.Betriebsschau` vom Bestand braucht.
type Betriebslage struct {
	db *sql.DB

	// wandert zählt die laufenden Wanderungen.
	//
	// **Ein Zähler und kein Schalter**: zwei Wanderungen zugleich gibt es
	// nicht, aber ein Schalter, den die zweite Rückgabe zurücksetzt, wäre
	// eine Zusage, die von der Reihenfolge abhinge.
	wandert atomic.Int64
}

// NeueBetriebslage baut sie über einer Verbindung.
func NeueBetriebslage(db *sql.DB) *Betriebslage { return &Betriebslage{db: db} }

// Probe fragt den Bestand etwas Triviales.
func (b *Betriebslage) Probe(ctx context.Context) error {
	var eins int

	if err := b.db.QueryRowContext(ctx, `SELECT 1`).Scan(&eins); err != nil {
		return uebersetze(err, "den Bestand ansprechen")
	}

	return nil
}

// WanderungLaeuft sagt, ob gerade migriert wird.
func (b *Betriebslage) WanderungLaeuft() bool { return b.wandert.Load() > 0 }

// Wandert markiert den Beginn einer Wanderung und gibt zurück, wie sie endet.
//
// **Die Rückgabe ist der Weg zurück** — dieselbe Form wie `defer`: wer sie
// vergisst, lässt den Prozess für immer „nicht bereit" melden, und das fällt
// sofort auf. Ein Schalter, den der Aufrufer selbst zurücksetzen müsste, fiele
// erst auf, wenn jemand die Bereitschaft abfragt.
func (b *Betriebslage) Wandert() func() {
	b.wandert.Add(1)

	return func() { b.wandert.Add(-1) }
}
