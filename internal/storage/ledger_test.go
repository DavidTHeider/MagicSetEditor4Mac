package storage_test

// **Das Buch der verbuchten Wanderungen** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **Hier und nicht nur über `internal/app`.** [storage.Ledger] wird von der
// Verträglichkeitsnaht gerufen, und dort ist sie eine Zeile; was sie
// **zurückgibt**, entscheidet dieses Paket — und ein Lauf, der sie nur
// mitnimmt, weil ein anderer Weg über sie führt, prüft ihre Aussage nicht
// (§10).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Name und Prüfsumme, und die Summe ist der Punkt.**
//
// Eine nachträglich geänderte Migration trägt dieselbe Nummer und einen
// anderen Inhalt; ohne die Summe wäre der Ledger eine Liste von Namen und
// beantwortete die Frage nicht, für die es ihn gibt.
func TestDerLedgerNenntJedeWanderungMitIhrerSumme(t *testing.T) {
	db := frischeDatenbank(t)

	buch, fehler := storage.Ledger(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("der Ledger ist nicht lesbar: %v", fehler)
	}

	if len(buch) == 0 {
		t.Fatal("nach der Migration steht nichts im Buch")
	}

	summe, verbucht := buch["0001_grundlage.sql"]
	if !verbucht {
		t.Fatalf("die erste Wanderung fehlt im Buch: %v", buch)
	}

	if len(summe) != 64 {
		t.Errorf("die Summe von 0001_grundlage.sql ist %q — erwartet war ein SHA-256 aus 64 Zeichen", summe)
	}

	// **Und die Zahl stimmt mit der anderen Auskunft überein.** Zwei
	// Listen über dieselbe Sache, die auseinanderlaufen, wären schlimmer
	// als eine.
	namen, namensfehler := storage.AngewendeteMigrationen(probekontext(t), db)
	if namensfehler != nil {
		t.Fatalf("die Buchführung ist nicht lesbar: %v", namensfehler)
	}

	if len(buch) != len(namen) {
		t.Errorf("das Buch führt %d Wanderungen, die Buchführung %d", len(buch), len(namen))
	}
}

// **Der Ledger migriert nicht** — wer fragt, ob eine Fassung diesen Bestand
// öffnen kann, soll ihn dabei nicht verändern.
func TestDerLedgerLegtNichtsAnUndAendertNichts(t *testing.T) {
	db := frischeDatenbank(t)

	vorher := angelegtAm(t, db)

	if _, fehler := storage.Ledger(probekontext(t), db); fehler != nil {
		t.Fatalf("der Ledger ist nicht lesbar: %v", fehler)
	}

	if nachher := angelegtAm(t, db); nachher != vorher {
		t.Errorf("das Lesen hat den Bestand geändert: %q → %q", vorher, nachher)
	}
}

// **Und ohne Buchführung sagt er das, statt eine leere Liste zu geben.** Eine
// leere Antwort hieße „keine Wanderung verbucht" — und darauf gäbe
// `domain.Vertraeglichkeit` ein `vertraeglich`, das nichts belegt.
func TestOhneBuchfuehrungMeldetDerLedgerEinenFehler(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t), `DROP TABLE migrationen`); fehler != nil {
		t.Fatalf("die Buchführung wegräumen: %v", fehler)
	}

	if _, fehler := storage.Ledger(probekontext(t), db); fehler == nil {
		t.Error("ohne Buchführung kam eine Antwort zurück")
	}
}
