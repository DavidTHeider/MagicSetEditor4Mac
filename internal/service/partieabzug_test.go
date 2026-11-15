package service_test

// **Was einen Abzug selbsttragend macht** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Neben `partieaufbau_test.go` und nicht darin**: dort steht, wie aus
// Wünschen eine Partie wird, hier, was eine Karte zum Zeichnen braucht (§2).

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Abzug trägt die Feldwerte und den Stil** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Wogegen hier geprüft wird
//
// **Dass ein verliehener Abzug auf einem fremden Rechner zeichnet.** Die
// Kennung einer Karte macht das Bild auf **diesem** Rechner erreichbar; wer
// das Deck verleiht, verleiht es an jemanden, der diesen Bestand nicht hat.
// Erst die eingefrorenen Feldwerte machen den Abzug selbsttragend.
func TestEinAbzugTraegtDieFeldwerteUndDenStil(t *testing.T) {
	t.Parallel()

	decks := &deckprobe{decks: map[int64]domain.Deck{7: {
		ID: 7, Name: "Wald", Spiel: "magic",
		Plaetze: []domain.Deckplatz{{
			ID: 1, KarteID: 3, Bezeichnung: "Wald", Menge: 40,
		}},
	}}}

	blaetter := &blattprobe{
		karten: map[int64]*domain.CardRow{3: {
			ID: 3, SetID: 9,
			Values: []domain.Value{
				{Field: "name", Text: "Wald"},
				{Field: "type", Text: "Basic Land — Forest"},
			},
		}},
		sets: map[int64]domain.SetHead{9: {ID: 9, Style: "m15-mainframe"}},
	}

	partie, fehler := service.NeuerPartieaufbau(decks).MitBlaettern(blaetter).
		Aufbauen(t.Context(), "p1", "magic", domain.Regelwerk{Strenge: domain.StrengeSandkasten},
			"anna", []service.Sitzwunsch{{Kennung: "anna", Name: "Anna", DeckID: 7}})
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	seite := partie.Sitze[0].Deck.Karten[0].Seiten[0]

	switch {
	case seite.Felder["type"] != "Basic Land — Forest":
		t.Errorf("die Feldwerte sind %v", seite.Felder)
	case seite.Stil != "m15-mainframe":
		t.Errorf("der Stil ist %q", seite.Stil)
	}
}

// **Ohne Blattquelle bleibt der Abzug, was er war** — der Weg ohne sie ist
// der Normalfall in einem Lauf und darf nicht scheitern (§3).
func TestOhneBlattquelleBleibtDerAbzugSchmal(t *testing.T) {
	t.Parallel()

	decks := &deckprobe{decks: map[int64]domain.Deck{7: {
		ID: 7, Name: "Wald", Spiel: "magic",
		Plaetze: []domain.Deckplatz{{
			ID: 1, KarteID: 3, Bezeichnung: "Wald", Menge: 40,
		}},
	}}}

	partie, fehler := service.NeuerPartieaufbau(decks).Aufbauen(
		t.Context(), "p1", "magic", domain.Regelwerk{Strenge: domain.StrengeSandkasten},
		"anna", []service.Sitzwunsch{{Kennung: "anna", Name: "Anna", DeckID: 7}})
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	seite := partie.Sitze[0].Deck.Karten[0].Seiten[0]

	if len(seite.Felder) != 0 || seite.Stil != "" {
		t.Fatalf("die Seite trägt %v und %q", seite.Felder, seite.Stil)
	}
}

// **Ein echter Druck hat keine eigene Karte** — und wird nicht nachgeschlagen.
func TestEinEchterDruckWirdNichtNachgeschlagen(t *testing.T) {
	t.Parallel()

	decks := &deckprobe{decks: map[int64]domain.Deck{7: {
		ID: 7, Name: "Wald", Spiel: "magic",
		Plaetze: []domain.Deckplatz{{
			ID: 1, Printing: "m21-250", Bezeichnung: "Forest", Menge: 40,
		}},
	}}}

	blaetter := &blattprobe{karten: map[int64]*domain.CardRow{}}

	if _, fehler := service.NeuerPartieaufbau(decks).MitBlaettern(blaetter).
		Aufbauen(t.Context(), "p1", "magic", domain.Regelwerk{Strenge: domain.StrengeSandkasten},
			"anna", []service.Sitzwunsch{{Kennung: "anna", Name: "Anna", DeckID: 7}}); fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	if blaetter.gelesen != 0 {
		t.Fatalf("es wurde %dmal nachgeschlagen", blaetter.gelesen)
	}
}

// **Eine Karte, die es nicht mehr gibt, hält den Aufbau nicht auf.**
//
// **Weil ein Deck auf eine gelöschte Karte zeigen darf** (§8: fremde und alte
// Daten sind der Normalfall). Der Abzug bleibt dann schmal — mit Name und
// Menge, ohne Feldwerte —, und das ist mehr als kein Tisch.
func TestEineFehlendeKarteHaeltDenAufbauNichtAuf(t *testing.T) {
	t.Parallel()

	decks := &deckprobe{decks: map[int64]domain.Deck{7: {
		ID: 7, Name: "Wald", Spiel: "magic",
		Plaetze: []domain.Deckplatz{{
			ID: 1, KarteID: 99, Bezeichnung: "Wald", Menge: 40,
		}},
	}}}

	partie, fehler := service.NeuerPartieaufbau(decks).
		MitBlaettern(&blattprobe{karten: map[int64]*domain.CardRow{}}).
		Aufbauen(t.Context(), "p1", "magic", domain.Regelwerk{Strenge: domain.StrengeSandkasten},
			"anna", []service.Sitzwunsch{{Kennung: "anna", Name: "Anna", DeckID: 7}})
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	if len(partie.Sitze[0].Deck.Karten) != 1 {
		t.Fatalf("der Abzug hat %d Karten",
			len(partie.Sitze[0].Deck.Karten))
	}
}

// blattprobe gibt Karten und ihre Sets her.
type blattprobe struct {
	karten  map[int64]*domain.CardRow
	sets    map[int64]domain.SetHead
	gelesen int
}

func (p *blattprobe) KarteLesen(
	_ context.Context, karteID int64,
) (*domain.CardRow, error) {
	p.gelesen++

	karte, da := p.karten[karteID]
	if !da {
		return nil, fmt.Errorf("%w: die Karte %d gibt es nicht",
			domain.ErrNotFound, karteID)
	}

	return karte, nil
}

func (p *blattprobe) Setstil(
	_ context.Context, setID int64,
) (string, error) {
	kopf, da := p.sets[setID]
	if !da {
		return "", fmt.Errorf("%w: das Set %d gibt es nicht",
			domain.ErrNotFound, setID)
	}

	return kopf.Style, nil
}

// **Ein Sitz bekommt sein Profil mit** — Werkstück G8 von Phase 5.5
// (#1465).
//
// **Und ohne Profil bleibt er ein Sitz**: die Bindung ist freiwillig.
func TestEinSitzBekommtSeinProfilMit(t *testing.T) {
	t.Parallel()

	decks := &deckprobe{decks: map[int64]domain.Deck{7: {
		ID: 7, Name: "Wald", Spiel: "magic",
		Plaetze: []domain.Deckplatz{{
			ID: 1, KarteID: 3, Bezeichnung: "Wald", Menge: 40,
		}},
	}}}

	partie, fehler := service.NeuerPartieaufbau(decks).Aufbauen(
		t.Context(), "p1", "magic",
		domain.Regelwerk{Strenge: domain.StrengeSandkasten}, "anna",
		[]service.Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: 7, ProfilID: 2},
			{Kennung: "bert", Name: "Bert", DeckID: 7},
		})
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	switch {
	case partie.Sitze[0].ProfilID != 2:
		t.Errorf("das Profil ist %d", partie.Sitze[0].ProfilID)
	case partie.Sitze[1].ProfilID != 0:
		t.Errorf("der ungebundene Sitz trägt %d", partie.Sitze[1].ProfilID)
	}
}
