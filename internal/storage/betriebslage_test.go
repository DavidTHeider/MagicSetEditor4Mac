package storage_test

// **Die Lage des Bestands** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein offener Bestand antwortet auf die Probe.**
func TestEinOffenerBestandAntwortetAufDieProbe(t *testing.T) {
	lage := storage.NeueBetriebslage(frischeDatenbank(t))

	if err := lage.Probe(probekontext(t)); err != nil {
		t.Errorf("Probe: %v", err)
	}

	if lage.WanderungLaeuft() {
		t.Error("die Wanderung galt als laufend")
	}
}

// **Ein geschlossener antwortet nicht** — und sagt es.
func TestEinGeschlossenerBestandAntwortetNicht(t *testing.T) {
	db := frischeDatenbank(t)

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if err := storage.NeueBetriebslage(db).Probe(probekontext(t)); err == nil {
		t.Error("die Probe schwieg")
	}
}

// **Während der Wanderung sagt sie es** — und danach nicht mehr.
//
// **Die Marke gehört dem Bestand und nicht dem Aufrufer**: wer sie von außen
// setzte, könnte sie stehen lassen, und dann bliebe der Prozess für immer
// „nicht bereit".
func TestWaehrendDerWanderungSagtSieEs(t *testing.T) {
	lage := storage.NeueBetriebslage(frischeDatenbank(t))

	fertig := lage.Wandert()

	if !lage.WanderungLaeuft() {
		t.Error("die Wanderung galt nicht als laufend")
	}

	fertig()

	if lage.WanderungLaeuft() {
		t.Error("die Wanderung galt danach noch als laufend")
	}
}
