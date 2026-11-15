package service_test

// **Eine Partie, in der gespielt wird — und eine, in der der Beginn noch
// aussteht** (#1851).
//
// # Warum zwei Vorrichtungen und nicht eine
//
// **Weil der Beginn vom Aufbau an läuft** — CR 103.8 (Fassung 2026-09-25):
// bis jeder Sitz seine Starthand behalten hat, nimmt der Sandkasten nur ihn
// an. Wer den Beginn prüft, braucht die Partie davor ([eineAufgebautePartie]);
// wer einen Befehl im Spiel prüft, die danach ([eineZugpartie]). Bis #1851
// war beides dieselbe, denn eine Partie, in der niemand anfing, spielte ohne
// Beginn.
//
// # Und warum der Beginn über den Sandkasten geht
//
// **Weil eine Vorrichtung, die den Betrieb nicht nachbaut, sich selbst
// prüft** (§3): ein Sitz, dem der Lauf „behalten" in den Stand schriebe,
// hätte keine Hand, und seine Bibliothek läge in der Reihenfolge der
// Deckliste — ein Zustand, den kein Befehl erklärt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineAufgebautePartie gibt eine Partie, wie der Aufbau sie hinterlässt: der
// erste Zug ist begonnen, und der Beginn steht für jeden Sitz aus.
func eineAufgebautePartie(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := einPartieaufbau().Aufbauen(t.Context(), "p1", "magic",
		einRegelwerk(), "sitz-1", zweiWuensche())
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	return partie
}

// eineZugpartie gibt eine Partie mit laufendem Zug — nach dem Beginn.
//
// **Über Decks, aus denen eine Starthand wird** ([einSpielaufbau]): die aus
// [einPartieaufbau] haben sechs Karten, der siebte Griff ginge ins Leere
// (CR 104.3c), und die Partie endete beim ersten Befehl danach.
func eineZugpartie(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := einSpielaufbau().Aufbauen(t.Context(), "p1", "magic",
		einRegelwerk(), "sitz-1", zweiWuensche())
	if fehler != nil {
		t.Fatalf("aufbauen: %v", fehler)
	}

	return beginnAbschliessen(t, partie)
}

// einSpielaufbau baut den Aufbau über zwei gleichen Decks mit zehn Karten:
// sieben für die Starthand (CR 103.5) und drei, damit auch der nächste Zug
// noch ziehen kann.
//
// **Dieselben Decks wie in [einPartieaufbau] und vier Berge dazu** — die
// Läufe über den Aufbau zählen deren Karten, und sie sollen weiter zählen,
// was sie zählen.
func einSpielaufbau() *service.Partieaufbau {
	decks := map[int64]domain.Deck{}

	for _, kennung := range []int64{1, 2} {
		deck := einDeck(kennung, "magic")
		deck.Plaetze = append(deck.Plaetze, domain.Deckplatz{
			ID: 4, Printing: "berg", Bezeichnung: "Berg", Menge: 4,
		})
		decks[kennung] = deck
	}

	return service.NeuerPartieaufbau(&deckprobe{decks: decks})
}

// beginnAbschliessen bringt den Beginn zu Ende: für jeden Sitz die Starthand,
// dann behalten — CR 103.5 und CR 103.8, über den Sandkasten wie jeder
// andere Befehl.
//
// **Mit einer Mischung, die nichts umordnet**, über die ganze Bibliothek des
// Sitzes: dieselbe Vorrichtung gibt so bei jedem Lauf dieselbe Hand.
func beginnAbschliessen(t *testing.T, partie domain.Partie) domain.Partie {
	t.Helper()

	for _, sitz := range partie.Sitze {
		kasten := service.NeuerSandkasten(festeMischung{
			reihe: ungemischt(len(bibliothekDes(partie, sitz.Kennung))),
		})

		_, partie = schritt(t, kasten, partie, sitz.Kennung,
			service.BefehlStarthand, nil)
		_, partie = schritt(t, kasten, partie, sitz.Kennung,
			service.BefehlBehalten, nil)
	}

	// **Dass der Beginn wirklich vorbei ist** — sonst prüfte jeder Lauf
	// danach die Sperre statt seines Befehls.
	if partie.BeginnLaeuft() {
		t.Fatal("nach Starthand und Behalten läuft der Beginn noch")
	}

	return partie
}

// ungemischt ist die Mischung, die nichts umordnet — über so viele Karten,
// wie genannt.
func ungemischt(anzahl int) []int {
	folge := make([]int, anzahl)
	for stelle := range folge {
		folge[stelle] = stelle
	}

	return folge
}

// bibliothekDes gibt die Kennungen in der Bibliothek eines Sitzes.
func bibliothekDes(partie domain.Partie, wer string) []string {
	bibliothek := []string{}

	for _, objekt := range partie.Objekte {
		if objekt.Beherrscher == wer && objekt.Zone == domain.ZoneBibliothek {
			bibliothek = append(bibliothek, objekt.Kennung)
		}
	}

	return bibliothek
}
