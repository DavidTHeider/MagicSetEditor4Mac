package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieFassungswacheEndetVorEinerNeuerenFassung(t *testing.T) {
	t.Parallel()

	db, err := storage.Oeffnen(t.Context(), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := storage.Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	var beendet atomic.Int32

	wache := Fassungswache(db, func() { beendet.Add(1) })

	frei, fehler := wache(t.Context())
	if fehler != nil || beendet.Load() != 0 {
		t.Fatalf("der eigene Stand: %v, beendet %d", fehler, beendet.Load())
	}

	frei()

	verbucht, _ := storage.AngewendeteMigrationen(t.Context(), db) //nolint:errcheck // gezählt
	name := fmt.Sprintf("%04d_kuenftig.sql", len(verbucht)+1)

	if fehler := storage.MigriereMit(t.Context(), db,
		fstest.MapFS{"migrationen/" + name: {Data: []byte("CREATE TABLE kuenftig (wert INTEGER) STRICT;")}}, nil); fehler != nil {
		t.Fatal(fehler)
	}

	_, erster := wache(t.Context())
	_, zweiter := wache(t.Context())

	var veraltet *domain.ProzessVeraltet
	if !errors.As(erster, &veraltet) || !errors.As(zweiter, &veraltet) {
		t.Fatalf("nach der Wanderung einer neueren Fassung: %v, %v", erster, zweiter)
	}

	if beendet.Load() != 1 {
		t.Errorf("beendet %d-mal, erwartet einmal", beendet.Load())
	}

	uebersetzt := Uebersetze(t.Context(), erster)
	if uebersetzt.Code != CodeBelegt || uebersetzt.Schluessel != domain.SchluesselProzessVeraltet || uebersetzt.Werte["unbekannt"] != name {
		t.Errorf("übersetzt: %+v", uebersetzt)
	}
}

// **Ein Lesefehler ist kein Ja** — Zweitblick über Phase 2, Befund 3: hier
// stand „das Schreiben sagt es selbst deutlicher", und ein Prozess, der nicht
// wusste, ob er den Bestand kennt, schrieb hinein. **Er endet dafür nicht**:
// beendet wird nur, wer nachweislich veraltet ist.
func TestDieFassungswacheVerweigertBeiEinemLesefehler(t *testing.T) {
	t.Parallel()

	db, err := storage.Oeffnen(t.Context(), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	_ = db.Close() //nolint:errcheck // geschlossen gefragt

	if frei, fehler := Fassungswache(db, func() { t.Error("beendet") })(t.Context()); fehler == nil || frei != nil {
		t.Errorf("ein Lesefehler gab frei: %v", fehler)
	}
}
