// **Die Testkarten der Werkstatt an der Naht** — Werkstück U7 von Phase 5.1
// (#1418).
// Eine eigene Datei, weil `vorlagenzeichnernaht_test.go` an der Zeilengrenze
// aus §2 steht; die Vorrichtungen kommen von dort.
package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine genannte Karte kommt, wie sie ist** — auch aus einem Set, das den
// Stil nicht führt: die Werkstatt fragt ausdrücklich nach ihr (U7, #1418).
func TestEineGenannteKarteKommtAlsProbekarte(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Fremd", Game: "magic", Style: "anderer"}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	karteID, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	naht, nahtfehler := app.VorlagennahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	karte, lesefehler := naht.Probekarte(probekontext(t), "/pakete/magic-m15.mse-style", karteID)
	if lesefehler != nil {
		t.Fatalf("Probekarte: %v", lesefehler)
	}

	if karte.ID != karteID {
		t.Errorf("die Karte ist %d — erwartet war %d", karte.ID, karteID)
	}

	if _, unbekannt := naht.Probekarte(probekontext(t), "/pakete/magic-m15.mse-style", 4711); !errors.Is(unbekannt, domain.ErrNotFound) {
		t.Errorf("eine unbekannte Karte gab %v — will ErrNotFound", unbekannt)
	}
}

// **Die Probekarten sind alle Karten des ersten Sets, das den Stil führt** —
// dieselbe Suche wie bei der Probekarte, damit die Auswahl im Inspector und
// die Vorgabe der Vorschau nie auseinanderlaufen (U7, #1418). Ohne passendes
// Set ist die Liste leer und kein Fehler: die Werkstatt zeigt dann den Weg.
func TestDieProbekartenSindDieDesPassendenSets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	schreiber := storage.NeuerKartenschreiber(db)

	fremd, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Fremd", Game: "magic", Style: "anderer"}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	if _, fehler := schreiber.KarteAnlegen(probekontext(t), fremd); fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	passend, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Passend", Game: "magic", Style: "m15"}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	for range 2 {
		if _, fehler := schreiber.KarteAnlegen(probekontext(t), passend); fehler != nil {
			t.Fatalf("Karte anlegen: %v", fehler)
		}
	}

	naht, nahtfehler := app.VorlagennahtUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	karten, lesefehler := naht.Probekarten(probekontext(t), "/pakete/magic-m15.mse-style")
	if lesefehler != nil {
		t.Fatalf("Probekarten: %v", lesefehler)
	}

	if len(karten) != 2 || karten[0].SetID != passend || karten[1].SetID != passend {
		t.Errorf("gefunden %+v", karten)
	}

	leer, leerfehler := naht.Probekarten(probekontext(t), "/pakete/magic-keiner.mse-style")
	if leerfehler != nil || len(leer) != 0 {
		t.Errorf("ohne Set kam %+v, %v", leer, leerfehler)
	}

	if _, kaputt := naht.MitLeser(leserohneKarten{fehler: errors.New("weg")}).
		Probekarten(probekontext(t), "/pakete/magic-m15.mse-style"); kaputt == nil {
		t.Error("ein Lesefehler kam nicht durch")
	}
}
