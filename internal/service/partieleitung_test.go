package service_test

// **Die Partieleitung** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Die erste Abnahme des Werkstücks
//
// „Derselbe Startsnapshot plus dieselbe Folge akzeptierter Commands ergibt
// gleichen Fachzustand und gleiche autorisierte Projektionen; Einspielen
// doppelt/zu spät/falsch geordneter Commands erzeugt keinen zweiten Zug."
//
// **Drei Fälle, und sie sind nicht derselbe**: *doppelt* ist derselbe Befehl
// noch einmal, *zu spät* ist einer gegen eine Fassung, die schon weiter ist,
// und *falsch geordnet* ist einer, der gegen eine Fassung rechnet, die es nie
// gab. Jeder bekommt hier seinen eigenen Lauf.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einePartie ist die kleinste, die eine Leitung führen kann.
func einePartie() domain.Partie {
	abzug := domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{
			{Kennung: "k1", Name: "Wald", Menge: 40},
		},
	}

	return domain.Partie{
		ID:        "p1",
		Spiel:     "magic",
		Regelwerk: domain.Regelwerk{Kennung: "sandkasten", Fassung: domain.Grundprofilfassung, Strenge: domain.StrengeSandkasten},
		Profil:    domain.Grundprofil("magic"),
		Wirt:      "sitz-1",
		Stand:     domain.StandLaeuft,
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "David", Deck: abzug},
			{Kennung: "sitz-2", Name: "Gegenüber", Deck: abzug},
		},
		Objekte: []domain.Spielobjekt{
			{
				Kennung: "o1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
				Eigner: "sitz-1", Zone: domain.ZoneHand, Karte: "k1",
			},
		},

		// **Mit Startleben** — wie jede Partie, die durch den Aufbau geht
		// (`service.Partieaufbau`). Ohne diese Zeile stünden beide Sitze
		// bei null, und CR 704.5a beendete die Partie beim ersten Befehl
		// (M5, #1607).
		//
		// **Eine Vorrichtung, die den Betrieb nicht nachbaut, prüft sich
		// selbst** (§3) — dieselbe Stelle wie in `internal/domain`.
	}.MitStartleben()
}

// eineLeitung baut eine Leitung über einem Speicher im Arbeitsspeicher.
func eineLeitung(t *testing.T) (*service.Partieleitung, *partieprobe) {
	t.Helper()

	speicher := &partieprobe{partien: map[string]domain.Partie{}}

	return service.NeuePartieleitung(speicher, service.NeuerSandkasten(service.NeuerZufall())), speicher
}

// **Eine Partie entsteht, und ihr Deck ist eingefroren.**
func TestEinePartieEntstehtMitEingefrorenemDeck(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)

	angelegt, fehler := leitung.Anlegen(context.Background(), einePartie())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if angelegt.Fassung != 1 {
		t.Errorf("die neue Partie hat die Fassung %d", angelegt.Fassung)
	}

	if angelegt.Sitze[0].Deck.Abdruck != "abcd" {
		t.Errorf("der Deckabzug fehlt: %+v", angelegt.Sitze[0].Deck)
	}
}

// **Eine Partie, die nicht prüft, entsteht nicht.**
func TestEineUngueltigePartieEntstehtNicht(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)

	kaputt := einePartie()
	kaputt.Wirt = "niemand"

	if _, fehler := leitung.Anlegen(
		context.Background(), kaputt); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if speicher.angelegt != 0 {
		t.Errorf("der Speicher wurde %dmal gerufen", speicher.angelegt)
	}
}

// **Derselbe Befehl zweimal erzeugt keinen zweiten Zug.**
func TestDerselbeBefehlZweimalErzeugtKeinenZweitenZug(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	befehl := domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
	}

	erste, fehler := leitung.Befehl(ctx, befehl)
	if fehler != nil {
		t.Fatalf("der erste Befehl: %v", fehler)
	}

	zweite, fehler := leitung.Befehl(ctx, befehl)
	if fehler != nil {
		t.Fatalf("der zweite Befehl: %v", fehler)
	}

	// **Dieselben Ereignisse, keine neuen.** Ein zweiter Ruf ist kein
	// Fehler — er ist derselbe Wunsch, und er bekommt dieselbe Antwort.
	if len(erste.Ereignisse) != len(zweite.Ereignisse) {
		t.Errorf("der zweite Ruf gab %d statt %d Ereignisse",
			len(zweite.Ereignisse), len(erste.Ereignisse))
	}

	if speicher.angewendet != 1 {
		t.Errorf("geschrieben wurde %dmal", speicher.angewendet)
	}

	if erste.Fassung != zweite.Fassung {
		t.Errorf("die Fassung stieg von %d auf %d", erste.Fassung, zweite.Fassung)
	}
}

// **Ein Befehl gegen eine überholte Fassung wird abgewiesen.**
func TestEinBefehlGegenEineUeberholteFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
	}); fehler != nil {
		t.Fatalf("der erste Befehl: %v", fehler)
	}

	faelle := map[string]int64{
		"zu spät":      1,
		"aus der Luft": 99,
	}

	for name, erwartet := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
				Kennung: "b-" + name, PartieID: "p1", Absender: "sitz-1",
				Art: service.BefehlEnttappen, Erwartet: erwartet,
				Nutzlast: map[string]any{"objekt": "o1"},
			})

			if !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Ein Befehl von einem, der nicht am Tisch sitzt, kommt nicht durch.**
func TestEinBefehlVonAusserhalbKommtNichtDurch(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	vorher := speicher.angewendet

	_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "wer-auch-immer",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if speicher.angewendet != vorher {
		t.Error("trotzdem geschrieben")
	}
}

// **Eine beendete Partie nimmt nichts mehr an.**
func TestEineBeendetePartieNimmtNichtsMehrAn(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if fehler := leitung.Beenden(ctx, "p1"); fehler != nil {
		t.Fatalf("Beenden: %v", fehler)
	}

	_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
	})

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Und eine pausierte ebenso** — sie wartet auf ihren Wirt.
	speicher.partien["p1"] = mitStand(speicher.partien["p1"], domain.StandPausiert)

	if _, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b2", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
	}); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die pausierte Partie gab %v", fehler)
	}
}

// mitStand gibt dieselbe Partie mit einem anderen Stand.
func mitStand(partie domain.Partie, stand domain.Partiestand) domain.Partie {
	partie.Stand = stand

	return partie
}
