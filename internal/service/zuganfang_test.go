package service_test

// **Eine Partie fängt mit Leben und mit einem ersten Zug an** — Werkstück G9
// von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
// **Dass das eingefrorene Regelprofil entscheidet, ob es zugweise zugeht.**
// Nicht ein Schalter und nicht eine Vorgabe im Aufbau: die Regelmatrix aus
// G4 sagt, was dieses Programm durchsetzt, und seit G9 steht `zugstruktur`
// dort auf `automatisch`.
//
// **Und dass alte Partien davon nichts merken.** Ihr Profil ist eingefroren
// (G4, #1461) und führt `zugstruktur` weiter als `manuell` — sie bleiben
// Sandkästen (§15, Punkt 4: keine stillschweigende Änderung bestehender
// Semantik).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jeder Sitz fängt mit dem Leben seines Profils an** — CR 119.1.
func TestEinAufbauSetztDasStartleben(t *testing.T) {
	t.Parallel()

	partie, fehler := einPartieaufbau().Aufbauen(t.Context(), "p1", "magic",
		einRegelwerk(), "sitz-1", zweiWuensche())
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	for _, sitz := range partie.Sitze {
		if sitz.Leben != 20 {
			t.Errorf("der Sitz %q hat %d Leben", sitz.Kennung, sitz.Leben)
		}
	}
}

// **Und der erste Zug läuft, weil das Profil es sagt.**
//
// **Der Wirt fängt an** — CR 103.1 (Fassung 2026-09-25) lässt den
// Startspieler bestimmen; wer das
// hier ist, steht fest, weil der Wirt die Partie hält. Eine Wahl darüber
// gehört zum Spielbeginn und nicht zum Aufbau.
func TestEinAufbauBeginntDenErstenZug(t *testing.T) {
	t.Parallel()

	partie, fehler := einPartieaufbau().Aufbauen(t.Context(), "p1", "magic",
		einRegelwerk(), "sitz-1", zweiWuensche())
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	switch {
	case partie.Zug.Nummer != 1:
		t.Errorf("es ist Zug %d", partie.Zug.Nummer)
	case partie.Zug.Aktiv != "sitz-1":
		t.Errorf("aktiv ist %q", partie.Zug.Aktiv)
	case partie.Zug.Starter != "sitz-1":
		t.Errorf("angefangen hat %q", partie.Zug.Starter)
	case partie.Zug.Abschnitt.Schritt != domain.SchrittVersorgung:
		t.Errorf("es steht bei %+v", partie.Zug.Abschnitt)
	case partie.Zug.Prioritaet != "sitz-1":
		t.Errorf("Priorität hat %q", partie.Zug.Prioritaet)
	}
}

// **Wer anfängt, zieht in seinem ersten Zug nicht** — CR 103.8a.
//
// **An der Kartenzahl nachgewiesen und nicht an einer Behauptung**, wie die
// Abnahme des Issues es verlangt.
func TestDerStarterZiehtImErstenZugNicht(t *testing.T) {
	t.Parallel()

	partie, fehler := einPartieaufbau().Aufbauen(t.Context(), "p1", "magic",
		einRegelwerk(), "sitz-1", zweiWuensche())
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	for _, objekt := range partie.Objekte {
		if objekt.Zone == domain.ZoneHand {
			t.Fatalf("%q liegt schon auf der Hand", objekt.Kennung)
		}
	}
}

// **Und der Aufbau fragt die Matrix nicht** — er baut immer mit Zug.
//
// # Warum hier kein Zweig steht
//
// **Weil er unbetretbar wäre** (§10). Der Aufbau nimmt sein Profil aus
// [domain.Grundprofil], und das führt `zugstruktur` seit G9 auf
// `automatisch`; ein `if` darüber hätte einen Ast, den kein Lauf und kein
// Betrieb erreicht.
//
// **Die Zusage an die alten Partien hängt woanders**: sie werden nicht
// aufgebaut, sondern **gelesen**, und ihr eingefrorenes Profil sagt weiter
// `manuell`. Was daraus folgt, entscheidet der Sandkasten beim Auslegen
// eines Befehls — dort ist der Zweig erreichbar, und dort steht er.
func TestDerAufbauBautImmerMitZug(t *testing.T) {
	t.Parallel()

	partie, fehler := einPartieaufbau().Aufbauen(t.Context(), "p1", "magic",
		einRegelwerk(), "sitz-1", zweiWuensche())
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	if grad := partie.Profil.Matrix.Grad(domain.BereichZugstruktur); grad !=
		domain.UnterstuetzungAutomatisch {
		t.Fatalf("das eingefrorene Profil führt die Zugstruktur als %q", grad)
	}

	if partie.Zug.Nummer != 1 {
		t.Errorf("es ist Zug %d", partie.Zug.Nummer)
	}
}

// zweiWuensche sind die zwei Sitze, mit denen hier aufgebaut wird.
func zweiWuensche() []service.Sitzwunsch {
	return []service.Sitzwunsch{
		{Kennung: "sitz-1", Name: "Anna", DeckID: 1},
		{Kennung: "sitz-2", Name: "Bert", DeckID: 2},
	}
}
