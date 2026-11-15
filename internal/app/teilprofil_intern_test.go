package app

// **Eine Teilung gehört ihrem Profil** — Befund Z8 des Zweitblicks über
// Phase 6.7 (NET-Teil).
//
// Eine vorbereitete Kopie trägt Inhalt und Pfad; ohne ihr Profil konnte
// dieselbe Kennung nach einem Wechsel die Einladung, Sammlung oder das Set
// des vorigen Menschen anbieten — und der Wechsel räumte sie nicht weg.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineTeilungGehoertIhremProfil(t *testing.T) {
	t.Parallel()

	anwendung, gezeigt, _ := eineTeilende(t)
	anwendung.profil.Store(1)

	gut := anwendung.TeilenVorbereiten(t.Context(), "einladung", "code").Teilgut
	if gut == nil {
		t.Fatal("die Teilung kam nicht zustande")
	}

	// **Ein anderes Profil sieht sie nicht.**
	anwendung.profil.Store(2)

	angeboten := anwendung.TeilenAnbieten(t.Context(), gut.Kennung)
	if angeboten.Fehler == nil || angeboten.Fehler.Code != CodeNichtGefunden || len(*gezeigt) != 0 {
		t.Errorf("im fremden Profil angeboten: %+v, gezeigt %v", angeboten.Fehler, *gezeigt)
	}

	if verworfen := anwendung.TeilenVerwerfen(t.Context(), gut.Kennung); verworfen.Fehler == nil {
		t.Error("im fremden Profil verworfen")
	}

	// **Im eigenen geht es weiter.**
	anwendung.profil.Store(1)

	if angeboten := anwendung.TeilenAnbieten(t.Context(), gut.Kennung); angeboten.Fehler != nil {
		t.Errorf("im eigenen Profil: %+v", angeboten.Fehler)
	}
}

// **Der Wechsel räumt die Teilungen weg** — samt ihrer Kopie.
func TestDerWechselRaeumtDieTeilungen(t *testing.T) {
	t.Parallel()

	anwendung, _, wurzel := eineTeilende(t)
	anwendung.profil.Store(1)

	gut := anwendung.TeilenVorbereiten(t.Context(), "einladung", "code").Teilgut
	if gut == nil {
		t.Fatal("die Teilung kam nicht zustande")
	}

	anwendung.teilungenRaeumen()

	if _, fehler := os.Stat(filepath.Join(wurzel, gut.Kennung)); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die Kopie liegt noch: %v", fehler)
	}

	if angeboten := anwendung.TeilenAnbieten(t.Context(), gut.Kennung); angeboten.Fehler == nil ||
		angeboten.Fehler.Code != CodeNichtGefunden {
		t.Errorf("nach dem Räumen: %+v", angeboten.Fehler)
	}
}

// **Eine Vorbereitung über den Wechsel hinweg trägt nichts ein** (Nachlauf
// über Phase 6.7, N1) — sie begann unter dem einen Profil und endet unter
// dem anderen; ihre Kopie geht, statt dem nächsten Menschen zu gehören.
func TestEineVorbereitungUeberDenWechselTraegtNichtsEin(t *testing.T) {
	t.Parallel()

	anwendung, _, wurzel := eineTeilende(t)
	anwendung.profil.Store(1)

	// **Der Wechsel geschieht mitten im Bauen** — die Deckausgabe ist der
	// Ort, an dem gebaut wird, und dort räumt der Lauf.
	anwendung.MitDecknaht(wechselndeDecks{anwendung: anwendung})

	if antwort := anwendung.TeilenVorbereiten(t.Context(), "deck", "1"); antwort.Fehler == nil ||
		antwort.Teilgut != nil {
		t.Fatalf("über den Wechsel hinweg vorbereitet: %+v", antwort)
	}

	if eintraege := ordnerinhalt(t, wurzel); len(eintraege) != 0 {
		t.Errorf("im Teilordner liegt %v", eintraege)
	}
}

// wechselndeDecks gibt ein Deck — und wechselt dabei das Profil, wie ein
// Mensch es täte, während die Kopie entsteht.
type wechselndeDecks struct {
	Decknaht

	anwendung *App
}

func (w wechselndeDecks) Deck(context.Context, int64) (domain.Deck, error) {
	w.anwendung.profil.Store(2)
	w.anwendung.teilungenRaeumen()

	return domain.Deck{Name: "Rot"}, nil
}
