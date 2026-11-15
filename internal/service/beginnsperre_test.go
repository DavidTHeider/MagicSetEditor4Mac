package service_test

// **Vor dem ersten Zug steht der Beginn** — CR 103.8 (Fassung 2026-09-25),
// Befund 2 des Zweitblicks über Phase 6.9
// (#1712).
//
// # Der Befund
//
// Der Aufbau ließ den Beginn ausstehen und begann trotzdem den ersten Zug —
// und der Sandkasten nahm darin jeden Befehl an. Angeboten wurde der Beginn,
// gesperrt war nichts: ein Sitz konnte mitten im Mulligan des anderen
// tappen, passen und ziehen.
//
// # Was hier gilt
//
// **Der Beginn läuft vom Aufbau an, bis jeder behalten hat**
// (`domain.Partie.BeginnLaeuft`) — seit
// #1851
// auch dann, wenn noch niemand eine Starthand gezogen hat. Bis dahin galt er
// erst ab der ersten, und eine Partie, in der niemand anfing, spielte ohne
// ihn aus der ungemischten Bibliothek.

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Auch wenn noch niemand angefangen hat** — CR 103.8 (Fassung 2026-09-25),
// #1851. Bis dahin nahm der Sandkasten in einer solchen Partie alles an: wer
// die Priorität des ersten Zugs hatte, passte, tappte und zog aus einer
// Bibliothek, die nie gemischt worden war.
//
// **Zwei Befehle aus zwei Klassen**: einer, der Priorität braucht, und
// einer, der als Buchführung ohne sie geht (CR 117.1). Die Sperre hängt an
// keinem von beiden, sondern am Beginn.
func TestDerBeginnHaeltAuchWennNiemandAngefangenHat(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})
	partie := eineAufgebautePartie(t)

	for _, befehl := range []domain.Spielbefehl{
		{Absender: "sitz-1", Art: service.BefehlPassen},
		{
			Absender: "sitz-2", Art: service.BefehlLeben,
			Nutzlast: map[string]any{"um": float64(-1)},
		},
	} {
		befehl.PartieID, befehl.Kennung = "p1", "b1"

		_, _, fehler := kasten.Auslegen(partie, befehl)
		if !errors.Is(fehler, domain.ErrValidation) ||
			!strings.Contains(fehler.Error(), "CR 103.8") {
			t.Errorf("%s vor jeder Starthand: %v", befehl.Art, fehler)
		}
	}

	// **Angeboten wird dasselbe wie mitten im Beginn** — jedem Sitz, auch
	// dem mit der Priorität des ersten Zugs.
	for _, wer := range []string{"sitz-1", "sitz-2"} {
		if angeboten := kasten.Erlaubt(partie, wer); !slices.Equal(
			angeboten, []string{
				service.BefehlStarthand, service.BefehlAufgeben,
				service.BefehlMuenzwurf, service.BefehlWuerfeln,
			}) {
			t.Errorf("%s: vor jeder Starthand angeboten %v", wer, angeboten)
		}
	}
}

func TestWerDenBeginnAnfaengtBringtIhnZuEnde(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})

	// **Sitz 2 fängt an**, und Sitz 1 hat die Priorität des ersten Zugs.
	_, partie := schritt(t, kasten, eineAufgebautePartie(t), "sitz-2",
		service.BefehlStarthand, nil)

	_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlPassen,
	})
	if !errors.Is(fehler, domain.ErrValidation) ||
		!strings.Contains(fehler.Error(), "CR 103.8") {
		t.Fatalf("passen mitten im Beginn: %v", fehler)
	}

	// **Angeboten wird der Beginn und was jederzeit geht** — aufgeben
	// (CR 104.3a) und die zwei Würfe, mit denen CR 103.1 den Startspieler
	// bestimmen lässt.
	if angeboten := kasten.Erlaubt(partie, "sitz-1"); !slices.Equal(
		angeboten, []string{
			service.BefehlStarthand, service.BefehlAufgeben,
			service.BefehlMuenzwurf, service.BefehlWuerfeln,
		}) {
		t.Errorf("mitten im Beginn angeboten: %v", angeboten)
	}

	// **Aufgeben geht trotzdem** — CR 104.3a.
	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b2",
		Art: service.BefehlAufgeben,
	}); fehler != nil {
		t.Errorf("aufgeben mitten im Beginn: %v", fehler)
	}

	// **Und sobald beide behalten haben, fängt der Zug an.**
	_, partie = schritt(t, kasten, partie, "sitz-2", service.BefehlBehalten, nil)
	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlStarthand, nil)
	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlBehalten, nil)

	if !slices.Contains(kasten.Erlaubt(partie, "sitz-1"), service.BefehlPassen) {
		t.Errorf("nach dem Beginn angeboten: %v", kasten.Erlaubt(partie, "sitz-1"))
	}

	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b3",
		Art: service.BefehlPassen,
	}); fehler != nil {
		t.Errorf("passen nach dem Beginn: %v", fehler)
	}
}

// **Ohne Zugstruktur sperrt der Beginn nichts** — wer sie für seine Runde
// abwählt, hat einen Sandkasten, in dem jeder jederzeit alles darf (§15,
// Punkt 4).
func TestOhneZugstrukturSperrtDerBeginnNichts(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})
	partie := eineAufgebautePartie(t)
	partie.Profil.Matrix = domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichZugstruktur},
	}.Wirksam(partie.Profil.Matrix)

	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlStarthand, nil)

	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art: service.BefehlLeben, Nutzlast: map[string]any{"um": float64(-1)},
	}); fehler != nil {
		t.Errorf("im Sandkasten abgewiesen: %v", fehler)
	}
}

// **Wer mitten im Beginn aufgibt, hält die anderen nicht auf** — Befund 2
// des Nachlaufs über Phase 6.9, der ganze Weg zu dritt bis zum ersten
// gewöhnlichen Befehl. Der Lauf darüber prüfte nur, dass Aufgeben
// angenommen wird, und benutzte den Stand danach nicht.
func TestWerMittenImBeginnAufgibtHaeltDieAnderenNichtAuf(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})

	partie, fehler := einPartieaufbau().Aufbauen(t.Context(), "p1", "magic",
		einRegelwerk(), "sitz-1", append(zweiWuensche(),
			service.Sitzwunsch{Kennung: "sitz-3", Name: "Cleo", DeckID: 1}))
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	// **Cleo geht, bevor sie zieht** — und zwar vor Annas Starthand: die
	// Decks dieser Vorrichtung haben sechs Karten, eine Starthand von sieben
	// greift ins Leere, und die Prüfung nach dem Aufgeben (CR 704.3) nähme
	// Anna sonst gleich mit (CR 104.3c).
	_, partie = schritt(t, kasten, partie, "sitz-3", service.BefehlAufgeben, nil)

	if partie.Stand != domain.StandLaeuft {
		t.Fatalf("nach dem Aufgeben: Stand %q — zwei sind noch dabei", partie.Stand)
	}

	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlStarthand, nil)
	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlBehalten, nil)
	_, partie = schritt(t, kasten, partie, "sitz-2", service.BefehlStarthand, nil)
	_, partie = schritt(t, kasten, partie, "sitz-2", service.BefehlBehalten, nil)

	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlPassen,
	}); fehler != nil {
		t.Errorf("passen nach dem Beginn der zwei, die blieben: %v", fehler)
	}
}
