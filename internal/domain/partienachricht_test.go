package domain_test

// **Was über die Leitung einer Partie geht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Vier Arten, und sie sind nicht dasselbe
//
// Das Issue verlangt die Trennung wörtlich: „bestätigte Events,
// Kontrollnachrichten, Chat und flüchtige Präsenz getrennt … keine
// Persistenz, Hashkette oder Replayaufnahme" — der letzte Halbsatz gilt der
// Präsenz.
//
// **Wer sie in eine Gestalt zwänge, bekäme an jeder Stelle Felder, die die
// Hälfte der Zeit nichts bedeuten** — und ein Mauszeiger läge in der
// Hashkette.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func eineNachricht(teile ...func(*domain.Partienachricht)) domain.Partienachricht {
	nachricht := domain.Partienachricht{
		Art:    domain.ArtEreignis,
		Partie: "p1",
		Epoche: 1,
		Folge:  1,
		Von:    "sitz-1",
		Inhalt: []byte(`{"was":"ziehen"}`),
		Kette:  strings.Repeat("a", 64),
	}

	for _, teil := range teile {
		teil(&nachricht)
	}

	return nachricht
}

// **Die kleinste Nachricht geht durch** — der Lauf steht zuerst da, damit
// jede Abweisung darunter eine Aussage ist.
func TestDieKleinsteNachrichtGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := eineNachricht().Pruefe(); fehler != nil {
		t.Errorf("die kleinste Nachricht wurde abgewiesen: %v", fehler)
	}
}

func TestWasKeineNachrichtIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Partienachricht){
		"ohne Art":        func(n *domain.Partienachricht) { n.Art = "" },
		"mit fremder Art": func(n *domain.Partienachricht) { n.Art = "geraune" },
		"ohne Partie":     func(n *domain.Partienachricht) { n.Partie = "" },
		"ohne Absender":   func(n *domain.Partienachricht) { n.Von = "  " },
		"ohne Epoche":     func(n *domain.Partienachricht) { n.Epoche = 0 },

		// **Ein Ereignis ohne Kettenglied ist kein bestätigtes Ereignis.**
		"als Ereignis ohne Kette": func(n *domain.Partienachricht) { n.Kette = "" },

		// **Und eines mit einem Glied falscher Länge auch nicht** — die Kette
		// ist ein SHA-256 in Hex, also vierundsechzig Zeichen.
		"als Ereignis mit kurzer Kette": func(n *domain.Partienachricht) {
			n.Kette = "abc"
		},

		"als Ereignis ohne Folge": func(n *domain.Partienachricht) { n.Folge = 0 },
	}

	for name, brich := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nachricht := eineNachricht(brich)
			if fehler := nachricht.Pruefe(); fehler == nil {
				t.Error("die Nachricht kam durch")
			}
		})
	}
}

// **Präsenz ist flüchtig** — sie trägt keine Folge und kein Kettenglied.
//
// Das Issue sagt es ausdrücklich: „keine Persistenz, Hashkette oder
// Replayaufnahme". Ein Mauszeiger in der Hashkette machte jede Bewegung zu
// einem Ereignis der Partie.
func TestPraesenzTraegtWederFolgeNochKette(t *testing.T) {
	t.Parallel()

	fluechtig := eineNachricht(func(n *domain.Partienachricht) {
		n.Art = domain.ArtPraesenz
		n.Folge = 0
		n.Kette = ""
	})

	if fehler := fluechtig.Pruefe(); fehler != nil {
		t.Errorf("flüchtige Präsenz wurde abgewiesen: %v", fehler)
	}
}

func TestPraesenzMitKetteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	falsch := eineNachricht(func(n *domain.Partienachricht) {
		n.Art = domain.ArtPraesenz
		n.Folge = 0
	})

	if fehler := falsch.Pruefe(); fehler == nil {
		t.Error("eine Präsenz mit Kettenglied kam durch")
	}
}

// **Chat und Kontrolle sind dauerhaft, aber nicht Teil der Hashkette.**
//
// Sie werden aufbewahrt — ein Chat gehört ins Protokoll einer Partie —, aber
// sie ändern den Spielstand nicht. Wer sie in die Kette nähme, machte aus
// „hallo" ein Spielereignis.
func TestChatUndKontrolleTragenFolgeAberKeineKette(t *testing.T) {
	t.Parallel()

	for _, art := range []domain.Nachrichtenart{domain.ArtChat, domain.ArtKontrolle} {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			nachricht := eineNachricht(func(n *domain.Partienachricht) {
				n.Art = art
				n.Kette = ""
			})

			if fehler := nachricht.Pruefe(); fehler != nil {
				t.Errorf("%s wurde abgewiesen: %v", art, fehler)
			}
		})
	}
}

// **Was aufbewahrt wird, sagt die Art** — und nicht der Aufrufer.
func TestWasAufbewahrtWird(t *testing.T) {
	t.Parallel()

	faelle := map[domain.Nachrichtenart]bool{
		domain.ArtEreignis:  true,
		domain.ArtKontrolle: true,
		domain.ArtChat:      true,
		domain.ArtPraesenz:  false,
	}

	for art, erwartet := range faelle {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			if art.WirdAufbewahrt() != erwartet {
				t.Errorf("%s: %v, erwartet %v", art, art.WirdAufbewahrt(), erwartet)
			}
		})
	}
}

// **Und was in die Hashkette geht, auch.**
func TestWasInDieKetteGeht(t *testing.T) {
	t.Parallel()

	if !domain.ArtEreignis.GehtInDieKette() {
		t.Error("ein Ereignis gehört in die Kette")
	}

	for _, art := range []domain.Nachrichtenart{
		domain.ArtKontrolle, domain.ArtChat, domain.ArtPraesenz,
	} {
		if art.GehtInDieKette() {
			t.Errorf("%s gehört nicht in die Kette", art)
		}
	}
}

// **Die vier stehen als Liste da**, damit eine Fläche sie aufzählen kann,
// ohne sie zu kennen.
func TestDieVierNachrichtenarten(t *testing.T) {
	t.Parallel()

	arten := domain.Nachrichtenarten()
	if len(arten) != 4 {
		t.Fatalf("%d Arten, erwartet 4", len(arten))
	}

	for _, art := range arten {
		if fehler := art.Pruefe(); fehler != nil {
			t.Errorf("%s ist keine gültige Art: %v", art, fehler)
		}
	}
}

// **Eine flüchtige Nachricht mit Folge wird abgewiesen** — sie hat keine
// Stelle in einer Reihenfolge, weil sie nicht aufbewahrt wird.
func TestPraesenzMitFolgeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	falsch := eineNachricht(func(n *domain.Partienachricht) {
		n.Art = domain.ArtPraesenz
		n.Kette = ""
	})

	if fehler := falsch.Pruefe(); fehler == nil {
		t.Error("eine Präsenz mit Folge kam durch")
	}
}

// **Und Chat mit Kettenglied auch** — es gehört nicht in die Hashkette.
func TestChatMitKetteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	falsch := eineNachricht(func(n *domain.Partienachricht) {
		n.Art = domain.ArtChat
	})

	if fehler := falsch.Pruefe(); fehler == nil {
		t.Error("ein Chat mit Kettenglied kam durch")
	}
}
