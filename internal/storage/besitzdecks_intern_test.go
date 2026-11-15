package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Deck, ein Deckplatz, eine Zuordnung oder ein Exemplar aus einem
// fremden Profil trifft nichts** —
// #1815.
//
// Deckplatz, Zuordnung und Herkunft sind Abkömmlinge ihres Decks oder
// Exemplars (0035). Gerufen aus Profil 1 mit den Kennungen von Profil 2 —
// lesend und schreibend. Erwartet: „nicht gefunden", und Profil 2 steht
// danach wie vorher.
func TestEinFremdesDeckTrifftNichts(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)

	ausFremdemProfil(t, db, deckrufe(db, fremd, MitProfil(t.Context(), 1)))
}

// deckrufe sind die Rufe der Deck- und Sammlungsablagen, nach `Typ.Methode`.
func deckrufe(db *sql.DB, fremd fremdeKennungen, eigen context.Context) map[string]func() error {
	decks := NeueDeckablage(db)
	zuordnungen := NeueDeckzuordnungen(db)
	herkunft := NeueHerkunftsablage(db)
	sammlung := NeuerSammlungsspeicher(db)
	eingeholt := domain.Herkunft{Art: domain.HerkunftEinfuhr, Quelle: "fremd"}

	return map[string]func() error{
		"Deckablage.Deck":     func() error { _, f := decks.Deck(eigen, fremd.deck); return f },
		"Deckablage.Loeschen": func() error { return decks.Loeschen(eigen, fremd.deck) },
		"Deckablage.Aendern": func() error {
			_, f := decks.Aendern(eigen, domain.Deck{ID: fremd.deck, Name: "X", Spiel: "magic", Fassung: 1})
			return f
		},
		"Deckablage.PlatzSetzen": func() error {
			_, f := decks.PlatzSetzen(eigen, fremd.deck, domain.Deckplatz{
				KarteID: fremd.karte, Menge: 4, Bereich: domain.BereichHauptdeck,
			})
			return f
		},
		"Deckablage.Uebernehmen": func() error {
			_, f := decks.Uebernehmen(eigen, domain.Deck{Name: "Geholt", Spiel: "magic"},
				[]domain.Deckplatz{{KarteID: fremd.karte, Menge: 1, Bereich: domain.BereichHauptdeck}},
				eingeholt, 0)
			return f
		},
		"Deckablage.PlatzVerschieben": func() error {
			_, f := decks.PlatzVerschieben(eigen, fremd.deck, fremd.platz, domain.BereichSideboard)
			return f
		},
		"Deckablage.PlatzLoeschen": func() error { return decks.PlatzLoeschen(eigen, fremd.deck, fremd.platz) },
		// **Die Kopie eines fremden Decks** (DOM-03.AK-02, #1698) holte sonst
		// seine Liste in das eigene Profil.
		"Deckablage.Kopieren": func() error { _, f := decks.Kopieren(eigen, fremd.deck, "Beute"); return f },
		"Deckzuordnungen.Setzen": func() error {
			_, f := zuordnungen.Setzen(eigen, domain.Deckzuordnung{
				DeckID: fremd.deck, PlatzID: fremd.platz, ExemplarID: fremd.exemplar, Menge: 2,
			})
			return f
		},
		"Deckzuordnungen.Loesen":     func() error { return zuordnungen.Loesen(eigen, fremd.deck, fremd.zuordnung) },
		"Deckzuordnungen.JeDeck":     func() error { _, f := zuordnungen.JeDeck(eigen, fremd.deck); return f },
		"Deckzuordnungen.JeExemplar": func() error { _, f := zuordnungen.JeExemplar(eigen, fremd.exemplar); return f },
		"Herkunftsablage.Deckherkunft": func() error {
			_, f := herkunft.Deckherkunft(eigen, fremd.deck)
			return f
		},
		"Herkunftsablage.Exemplarherkunft": func() error {
			_, f := herkunft.Exemplarherkunft(eigen, fremd.exemplar)
			return f
		},
		"Herkunftsablage.DeckherkunftSetzen": func() error {
			return herkunft.DeckherkunftSetzen(eigen, fremd.deck, eingeholt, 0)
		},
		"Herkunftsablage.ExemplarherkunftSetzen": func() error {
			return herkunft.ExemplarherkunftSetzen(eigen, fremd.exemplar, eingeholt)
		},
		"Sammlungsspeicher.Einzeln": func() error { _, f := sammlung.Einzeln(eigen, fremd.exemplar); return f },
		"Sammlungsspeicher.Aendern": func() error {
			return sammlung.Aendern(eigen, &domain.Exemplar{
				ID: fremd.exemplar, Printing: "druck-fremd", Menge: 9, Zustand: domain.ZustandNearMint,
				Sprache: "en", Finish: domain.FinishNonfoil, Fassung: 1,
			})
		},
		"Sammlungsspeicher.Loeschen": func() error { return sammlung.Loeschen(eigen, fremd.exemplar, 1) },
	}
}

// **Und über die Grenze hinweg verbinden geht auch nicht**: ein eigenes Deck
// nimmt keine fremde Karte und keinen fremden Platz auf, und ein fremdes
// Exemplar bedient keinen eigenen Platz.
//
// Ohne diese Fälle könnte ein Deck auf eine fremde Karte zeigen, und wer es
// liest, läse deren Namen mit.
func TestEinEigenesDeckNimmtNichtsFremdes(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)
	eigen := MitProfil(t.Context(), 1)

	decks := NeueDeckablage(db)

	eigenesDeck, fehler := decks.Anlegen(eigen, domain.Deck{Name: "Eigen", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("das eigene Deck: %v", fehler)
	}

	eigenerPlatz, fehler := decks.PlatzSetzen(eigen, eigenesDeck.ID, domain.Deckplatz{
		Printing: "druck-eigen", Menge: 1, Bereich: domain.BereichHauptdeck,
	})
	if fehler != nil {
		t.Fatalf("der eigene Platz: %v", fehler)
	}

	eigenesStueck, fehler := NeuerSammlungsspeicher(db).Anlegen(eigen, &domain.Exemplar{
		Printing: "druck-eigen", Menge: 1, Zustand: domain.ZustandNearMint,
		Sprache: "en", Finish: domain.FinishNonfoil,
	})
	if fehler != nil {
		t.Fatalf("das eigene Exemplar: %v", fehler)
	}

	zuordnungen := NeueDeckzuordnungen(db)

	ausFremdemProfil(t, db, map[string]func() error{
		"fremde Karte in eigenes Deck": func() error {
			_, f := decks.PlatzSetzen(eigen, eigenesDeck.ID, domain.Deckplatz{
				KarteID: fremd.karte, Menge: 1, Bereich: domain.BereichHauptdeck,
			})
			return f
		},
		"fremder Platz unter eigenem Deck": func() error {
			_, f := zuordnungen.Setzen(eigen, domain.Deckzuordnung{
				DeckID: eigenesDeck.ID, PlatzID: fremd.platz, ExemplarID: eigenesStueck, Menge: 1,
			})
			return f
		},
		"fremdes Exemplar an eigenem Platz": func() error {
			_, f := zuordnungen.Setzen(eigen, domain.Deckzuordnung{
				DeckID: eigenesDeck.ID, PlatzID: eigenerPlatz.ID, ExemplarID: fremd.exemplar, Menge: 1,
			})
			return f
		},
	})
}

// **Die Gegenprobe: der Eigner erreicht sein Deck** — samt Platz, Zuordnung
// und Herkunft. Ohne sie hieße ein „nicht gefunden" oben nur, dass der Ruf
// gar nicht ging.
func TestDerEignerErreichtSeinDeck(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)
	eigner := MitProfil(t.Context(), 2)

	verschoben, fehler := NeueDeckablage(db).PlatzVerschieben(eigner, fremd.deck, fremd.platz, domain.BereichSideboard)
	if fehler != nil || verschoben.Bereich != domain.BereichSideboard {
		t.Fatalf("der Eigner verschiebt seinen Platz nicht: %+v (%v)", verschoben, fehler)
	}

	zuordnungen, fehler := NeueDeckzuordnungen(db).JeExemplar(eigner, fremd.exemplar)
	if fehler != nil || len(zuordnungen) != 0 {
		// Der Platz ist verschoben und damit ein neuer: seine Zuordnung ging
		// mit dem alten (`ON DELETE CASCADE`).
		t.Errorf("Zuordnungen des Eigners: %+v (%v)", zuordnungen, fehler)
	}

	herkunft, fehler := NeueHerkunftsablage(db).Deckherkunft(eigner, fremd.deck)
	if fehler != nil || herkunft.Art != domain.HerkunftEinfuhr {
		t.Errorf("die Herkunft des Eigners: %+v (%v)", herkunft, fehler)
	}
}
