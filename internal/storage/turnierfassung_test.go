package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fassung zählt jede Änderung am Turnier.**
//
// Sie stand in der Tabelle, im Verbund und im DTO — und wurde von niemandem
// erhöht und von niemandem geprüft (Befund 7 des Nachlaufs über Phase 4.7).
// Ein Feld, das eine Zusage macht, die es nicht einlöst, ist schlimmer als
// keines: das Fenster bekam eine Zahl, die immer 1 war, und hätte darauf
// gebaut.
func TestJedeAenderungErhoehtDieFassung(t *testing.T) {
	ablage := turnierablage(t)
	hintergrund := probekontext(t)

	turnier, err := ablage.Anlegen(hintergrund, domain.Turnier{
		Name: "Küchentisch", Spiel: "Magic",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if turnier.Fassung != 1 {
		t.Fatalf("frisch angelegt → Fassung %d, erwartet 1", turnier.Fassung)
	}

	vorher := turnier.Fassung

	for _, schritt := range []struct {
		name string
		tun  func() error
	}{
		{"anmelden", func() error {
			_, meldefehler := ablage.Anmelden(hintergrund, turnier.ID,
				domain.Teilnehmer{Name: "Ada"})

			return meldefehler
		}},
		{"noch einer", func() error {
			_, meldefehler := ablage.Anmelden(hintergrund, turnier.ID,
				domain.Teilnehmer{Name: "Bert"})

			return meldefehler
		}},
		{"aussteigen", func() error {
			gelesen, lesefehler := ablage.Turnier(hintergrund, turnier.ID)
			if lesefehler != nil {
				return lesefehler
			}

			return ablage.Aussteigen(hintergrund, gelesen.Teilnehmer[0].ID, true)
		}},
	} {
		if fehler := schritt.tun(); fehler != nil {
			t.Fatalf("%s: %v", schritt.name, fehler)
		}

		gelesen, lesefehler := ablage.Turnier(hintergrund, turnier.ID)
		if lesefehler != nil {
			t.Fatalf("%s nachlesen: %v", schritt.name, lesefehler)
		}

		if gelesen.Fassung <= vorher {
			t.Errorf("nach %q → Fassung %d, erwartet mehr als %d",
				schritt.name, gelesen.Fassung, vorher)
		}

		vorher = gelesen.Fassung
	}
}

// **Und ein Ergebnis wird gegen die Fassung eingetragen.**
//
// Zwei Leute am selben Tisch, beide haben die Runde offen: wer als Zweiter
// einträgt, überschriebe sonst die Berichtigung des Ersten, ohne dass jemand
// es merkte (§7, Regel 2).
func TestEinErgebnisGegenEineVeralteteFassungWirdAbgewiesen(t *testing.T) {
	ablage := turnierablage(t)
	hintergrund := probekontext(t)

	turnier := gelegtesTurnier(t, ablage)
	paarung := turnier.Runden[0].Paarungen[0]

	// **Der Erste trägt ein.**
	if err := ablage.ErgebnisSetzen(hintergrund, turnier.Fassung, paarung.ID,
		&domain.Ergebnis{SpieleA: 2}); err != nil {
		t.Fatalf("erstes Ergebnis: %v", err)
	}

	// **Der Zweite hat denselben Stand gelesen** — und kommt zu spät.
	err := ablage.ErgebnisSetzen(hintergrund, turnier.Fassung, paarung.ID,
		&domain.Ergebnis{SpieleB: 2})

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("zweites Ergebnis → %v, erwartet ErrConflict", err)
	}

	// **Und der Stand des Ersten steht noch.**
	gelesen, lesefehler := ablage.Turnier(hintergrund, turnier.ID)
	if lesefehler != nil {
		t.Fatalf("nachlesen: %v", lesefehler)
	}

	stand := gelesen.Runden[0].Paarungen[0].Ergebnis
	if stand == nil || stand.SpieleA != 2 {
		t.Errorf("Ergebnis nach dem Zusammenstoß → %+v, erwartet SpieleA 2", stand)
	}
}

// gelegtesTurnier gibt ein Turnier mit zwei Teilnehmern und einer Runde.
func gelegtesTurnier(t *testing.T, ablage *storage.Turnierablage) domain.Turnier {
	t.Helper()

	turnier, err := ablage.Anlegen(probekontext(t), domain.Turnier{
		Name: "Küchentisch", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ada, adafehler := ablage.Anmelden(probekontext(t), turnier.ID,
		domain.Teilnehmer{Name: "Ada"})
	if adafehler != nil {
		t.Fatalf("Ada anmelden: %v", adafehler)
	}

	bert, bertfehler := ablage.Anmelden(probekontext(t), turnier.ID,
		domain.Teilnehmer{Name: "Bert"})
	if bertfehler != nil {
		t.Fatalf("Bert anmelden: %v", bertfehler)
	}

	if _, legefehler := ablage.RundeLegen(probekontext(t), turnier.ID, 1,
		[]domain.Paarung{{A: ada.ID, B: bert.ID}},
	); legefehler != nil {
		t.Fatalf("Runde legen: %v", legefehler)
	}

	gelesen, lesefehler := ablage.Turnier(probekontext(t), turnier.ID)
	if lesefehler != nil {
		t.Fatalf("nachlesen: %v", lesefehler)
	}

	return gelesen
}
