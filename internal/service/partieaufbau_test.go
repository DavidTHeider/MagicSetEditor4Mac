package service_test

// **Aus einem Deck wird ein Tisch** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Die vierte Abnahme des Werkstücks
//
// „Deckänderung nach Spielbeginn verändert den eingefrorenen Spielinhalt
// nicht. Karten gleichen Namens/gleicher Textur bleiben verschiedene
// Objekte."
//
// **Beide Hälften stehen hier**, und die zweite wird gezählt: vier Wälder
// sind vier Objekte, nicht eines mit der Menge vier.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// deckprobe gibt vorbereitete Decks — und zählt, wie oft sie gelesen wurden.
type deckprobe struct {
	decks   map[int64]domain.Deck
	gelesen int
	klagt   error
}

func (d *deckprobe) Deck(_ context.Context, kennung int64) (domain.Deck, error) {
	d.gelesen++

	if d.klagt != nil {
		return domain.Deck{}, d.klagt
	}

	deck, da := d.decks[kennung]
	if !da {
		return domain.Deck{}, domain.ErrNotFound
	}

	return deck, nil
}

// einDeck ist ein kleines, spielbares Deck.
func einDeck(kennung int64, spiel string) domain.Deck {
	return domain.Deck{
		ID: kennung, Name: "Probe", Spiel: spiel,
		Plaetze: []domain.Deckplatz{
			{ID: 1, Printing: "wald", Bezeichnung: "Wald", Menge: 4},
			{ID: 2, Printing: "baer", Bezeichnung: "Bär", Menge: 2},
			{
				ID: 3, Printing: "gift", Bezeichnung: "Gift", Menge: 3,
				Bereich: domain.BereichSideboard,
			},
		},
	}
}

// einPartieaufbau baut den Aufbau über zwei gleichen Decks.
func einPartieaufbau() *service.Partieaufbau {
	return service.NeuerPartieaufbau(&deckprobe{decks: map[int64]domain.Deck{
		1: einDeck(1, "magic"),
		2: einDeck(2, "magic"),
	}})
}

// eineWerkstatt ist das Regelwerk, unter dem aufgebaut wird.
func einRegelwerk() domain.Regelwerk {
	return domain.Regelwerk{
		Kennung: "sandkasten", Fassung: domain.Grundprofilfassung, Strenge: domain.StrengeSandkasten,
	}
}

// **Vier Wälder sind vier Objekte.**
func TestVierWaelderSindVierObjekte(t *testing.T) {
	t.Parallel()

	aufbau := einPartieaufbau()

	partie, fehler := aufbau.Aufbauen(context.Background(), "p1", "magic",
		einRegelwerk(), "sitz-1", []service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	// **Sechs je Sitz** — vier Wälder und zwei Bären; das Seitendeck bleibt
	// draußen.
	if len(partie.Objekte) != 12 {
		t.Fatalf("es liegen %d Objekte", len(partie.Objekte))
	}

	gesehen := map[string]bool{}

	for _, objekt := range partie.Objekte {
		if gesehen[objekt.Kennung] {
			t.Errorf("die Kennung %q kommt zweimal vor", objekt.Kennung)
		}

		gesehen[objekt.Kennung] = true

		if objekt.Zone != domain.ZoneBibliothek {
			t.Errorf("%q liegt in %q", objekt.Kennung, objekt.Zone)
		}
	}

	// **Und die Partie prüft** — sie ist spielbar, nicht nur gebaut.
	if fehler := partie.Pruefe(); fehler != nil {
		t.Errorf("die aufgebaute Partie: %v", fehler)
	}
}

// **Das Seitendeck liegt im Abzug und nicht in der Bibliothek.**
//
// **Beides zusammen ist die Aussage**: wer es ganz wegließe, könnte nachher
// nicht wechseln; wer es hineinlegte, hätte fünfzehn Karten zu viel im Deck.
func TestDasSeitendeckLiegtImAbzugUndNichtInDerBibliothek(t *testing.T) {
	t.Parallel()

	aufbau := einPartieaufbau()

	partie, fehler := aufbau.Aufbauen(context.Background(), "p1", "magic",
		einRegelwerk(), "sitz-1", []service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	abzug := partie.Sitze[0].Deck

	if abzug.Zaehle() != 9 {
		t.Errorf("der Abzug zählt %d Karten", abzug.Zaehle())
	}

	seitendeck := 0

	for _, karte := range abzug.Karten {
		if karte.Bereich == domain.BereichSideboard {
			seitendeck += karte.Menge
		}
	}

	if seitendeck != 3 {
		t.Errorf("im Abzug liegen %d Seitendeckkarten", seitendeck)
	}

	for _, objekt := range partie.Objekte {
		if objekt.Karte == "p2-3" {
			t.Error("eine Seitendeckkarte liegt in der Bibliothek")
		}
	}
}

// **Zwei gleiche Decks geben zwei gleiche Abdrücke — und zwei Sitze.**
//
// **Der Abdruck hängt am Inhalt und nicht an der Kennung des Decks.** Zwei
// Seiten meinen dasselbe, wenn dieselben Karten darin liegen; wer die
// Deckkennung hineinrechnete, bekäme für zwei Kopien desselben Decks zwei
// verschiedene Abdrücke.
func TestDerAbdruckHaengtAmInhalt(t *testing.T) {
	t.Parallel()

	aufbau := einPartieaufbau()

	partie, fehler := aufbau.Aufbauen(context.Background(), "p1", "magic",
		einRegelwerk(), "sitz-1", []service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if partie.Sitze[0].Deck.Abdruck != partie.Sitze[1].Deck.Abdruck {
		t.Error("zwei gleiche Decks haben zwei Abdrücke")
	}

	// **Die Gegenprobe**: ein anderes Deck, ein anderer Abdruck.
	quelle := &deckprobe{decks: map[int64]domain.Deck{
		1: einDeck(1, "magic"),
		2: func() domain.Deck {
			deck := einDeck(2, "magic")
			deck.Plaetze[0].Menge = 5

			return deck
		}(),
	}}

	andere, fehler := service.NeuerPartieaufbau(quelle).Aufbauen(
		context.Background(), "p2", "magic", einRegelwerk(), "sitz-1",
		[]service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if andere.Sitze[0].Deck.Abdruck == andere.Sitze[1].Deck.Abdruck {
		t.Error("zwei verschiedene Decks haben denselben Abdruck")
	}

	// **Und die Herkunft steht dabei** — eine Spur, kein Verweis.
	if andere.Sitze[1].Deck.Herkunft != "2" {
		t.Errorf("die Herkunft ist %q", andere.Sitze[1].Deck.Herkunft)
	}
}

// **Was sich nicht aufbauen lässt.**
func TestWasSichNichtAufbauenLaesst(t *testing.T) {
	t.Parallel()

	t.Run("ein Deck aus einem anderen Spiel", func(t *testing.T) {
		t.Parallel()

		quelle := &deckprobe{decks: map[int64]domain.Deck{
			1: einDeck(1, "magic"),
			2: einDeck(2, "space"),
		}}

		_, fehler := service.NeuerPartieaufbau(quelle).Aufbauen(
			context.Background(), "p1", "magic", einRegelwerk(), "sitz-1",
			[]service.Sitzwunsch{
				{Kennung: "sitz-1", Name: "David", DeckID: 1},
				{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
			})

		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("es kam %v", fehler)
		}
	})

	// **Ein Deck über der Grenze wird nicht gelegt** (Nachlauf über Phase 7,
	// R1) — läge es, liefe dieser Lauf über Minuten und Gigabytes.
	t.Run("ein Deck über der Grenze des Abzugs", func(t *testing.T) {
		t.Parallel()

		riesig := einDeck(1, "magic")
		riesig.Plaetze[0].Menge = 1 << 40
		quelle := &deckprobe{decks: map[int64]domain.Deck{1: riesig, 2: einDeck(2, "magic")}}

		_, fehler := service.NeuerPartieaufbau(quelle).Aufbauen(
			context.Background(), "p1", "magic", einRegelwerk(), "sitz-1",
			[]service.Sitzwunsch{
				{Kennung: "sitz-1", Name: "David", DeckID: 1},
				{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
			})

		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("es kam %v", fehler)
		}
	})

	t.Run("ein Deck, das es nicht gibt", func(t *testing.T) {
		t.Parallel()

		aufbau := einPartieaufbau()

		_, fehler := aufbau.Aufbauen(
			context.Background(), "p1", "magic", einRegelwerk(), "sitz-1",
			[]service.Sitzwunsch{
				{Kennung: "sitz-1", Name: "David", DeckID: 9},
				{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
			})

		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("es kam %v", fehler)
		}
	})

	t.Run("ein Tisch ohne jeden Sitz", func(t *testing.T) {
		t.Parallel()

		aufbau := einPartieaufbau()

		_, fehler := aufbau.Aufbauen(
			context.Background(), "p1", "magic", einRegelwerk(), "sitz-1",
			[]service.Sitzwunsch{})

		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("es kam %v", fehler)
		}
	})

	t.Run("wenn die Deckquelle klagt", func(t *testing.T) {
		t.Parallel()

		kaputt := errors.New("der Bestand ist zu")
		quelle := &deckprobe{klagt: kaputt}

		_, fehler := service.NeuerPartieaufbau(quelle).Aufbauen(
			context.Background(), "p1", "magic", einRegelwerk(), "sitz-1",
			[]service.Sitzwunsch{
				{Kennung: "sitz-1", Name: "David", DeckID: 1},
				{Kennung: "sitz-2", Name: "Gegenüber", DeckID: 2},
			})

		if !errors.Is(fehler, kaputt) {
			t.Fatalf("es kam %v", fehler)
		}
	})
}

// **Jede Abzugskarte trägt ihre auflösbare Abhängigkeit.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Ohne sie hätte der Spieltisch kein Bild.** Ein Spielobjekt zeigt auf eine
// Abzugskarte, und die Abzugskarte muss sagen, **welche Karte** das im
// Bestand ist — sonst ist der Weg zum Kartenbild abgeschnitten (G2,
// #1459).
//
// **Ein echter Druck trägt keine**, und das ist kein Mangel: dort gibt es
// keine eigene Karte, die dieses Programm zeichnen könnte.
func TestJedeAbzugskarteTraegtIhreKartenkennung(t *testing.T) {
	t.Parallel()

	decks := &deckprobe{decks: map[int64]domain.Deck{
		1: {
			ID: 1, Name: "Aggro", Spiel: "magic",
			Plaetze: []domain.Deckplatz{
				{ID: 10, KarteID: 77, Bezeichnung: "Wald", Menge: 2},
				{ID: 11, Printing: "abc-1", Bezeichnung: "Fremd", Menge: 1},
			},
		},
	}}

	partie, fehler := service.NeuerPartieaufbau(decks).Aufbauen(
		t.Context(), "p1", "magic",
		domain.Regelwerk{Kennung: "sandkasten", Fassung: domain.Grundprofilfassung, Strenge: domain.StrengeSandkasten},
		"anna", []service.Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: 1},
			{Kennung: "bert", Name: "Bert", DeckID: 1},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	karten := partie.Sitze[0].Deck.Karten
	if len(karten) != 2 {
		t.Fatalf("%d Abzugskarten", len(karten))
	}

	if len(karten[0].Seiten) != 1 || karten[0].Seiten[0].KarteID != 77 {
		t.Errorf("die eigene Karte: %+v", karten[0].Seiten)
	}

	if karten[0].Seiten[0].Name != "Wald" {
		t.Errorf("Name %q", karten[0].Seiten[0].Name)
	}

	// **Ein echter Druck trägt null** — es gibt keine eigene Karte.
	if len(karten[1].Seiten) != 1 || karten[1].Seiten[0].KarteID != 0 {
		t.Errorf("der echte Druck: %+v", karten[1].Seiten)
	}
}

// **Ein Tisch für einen baut sich auf** — Werkstück G3 von Phase 5.5
// (#1460):
// „Solo Goldfish sowie 2–4 lokale Self-Play-Sitze".
//
// **CR 100.1** (Fassung 2026-08-19): die Regeln gelten für Spiele mit zwei
// oder mehr Spielern. Ein Goldfish hat einen — er ist damit ein **Aufbau** und
// kein Spiel, und dieses Programm setzt an ihm auch keine Regel durch.
func TestEinTischFuerEinenBautSichAuf(t *testing.T) {
	t.Parallel()

	aufbau := einPartieaufbau()

	partie, fehler := aufbau.Aufbauen(
		context.Background(), "p1", "magic", einRegelwerk(), "sitz-1",
		[]service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: 1},
		})

	switch {
	case fehler != nil:
		t.Fatalf("ein Tisch für einen wurde abgewiesen: %v", fehler)
	case len(partie.Sitze) != 1:
		t.Errorf("%d Sitze, erwartet 1", len(partie.Sitze))
	}
}
