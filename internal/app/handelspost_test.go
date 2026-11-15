package app_test

// **Der Poststand an der Naht** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass die Naht zählt und nichts zeigt.** Was ein Poststand ist, steht im
// Bestand; hier steht, dass die Zahlen heil durchkommen — und dass ein Nein
// ein Nein bleibt, statt als leerer Stand anzukommen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerPoststandKommtDurchDieNaht(t *testing.T) {
	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)

	ablage := storage.NeueAusgangsablage(db)

	for _, kennung := range []string{"n-1", "n-2"} {
		if fehler := ablage.Einreihen(probekontext(t), storage.Ausgangszeile{
			NachrichtID: kennung, Vorgang: "v-1",
			Empfaenger: "bgegenueber", Nachricht: "{}",
		}); fehler != nil {
			t.Fatalf("Einreihen: %v", fehler)
		}
	}

	if fehler := ablage.Versucht(probekontext(t), "n-2"); fehler != nil {
		t.Fatalf("Versucht: %v", fehler)
	}

	antwort := flaeche.Handelspoststand(probekontext(t), "bgegenueber")
	if antwort.Fehler != nil {
		t.Fatalf("Handelspoststand: %+v", antwort.Fehler)
	}

	if antwort.Stand.Offen != 2 || antwort.Stand.HoechsteVersuche != 1 {
		t.Errorf("der Stand ist %+v", antwort.Stand)
	}

	if antwort.Stand.Kennung != "bgegenueber" {
		t.Errorf("der Stand nennt %q", antwort.Stand.Kennung)
	}
}

// **Ein Fehler kommt als Fehler an** — und der Stand nennt trotzdem, zu wem
// er gehört.
func TestEinPoststandfehlerKommtAlsFehlerAn(t *testing.T) {
	antwort := app.PeerflaecheUeber(frischerBestand(t)).
		Handelspoststand(t.Context(), "bgegenueber")

	if antwort.Fehler == nil {
		t.Fatal("es kam kein Fehler")
	}

	if antwort.Stand.Kennung != "bgegenueber" {
		t.Errorf("der Stand nennt %q", antwort.Stand.Kennung)
	}
}
