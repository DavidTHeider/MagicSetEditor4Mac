package domain_test

// **Was neben dem Spiel läuft** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Was diese Datei nicht prüft**: den **Ablauf**. Er steht im Fenster, weil
// er eine Frage an die Wanduhr ist und an das, was jemand sieht — geprüft
// wird er in `frontend/src/haken/partienebenbei.test.ts`. Hier steht die
// andere Hälfte der Zusage aus dem Issue: der **monotone Empfängerstand**.

import (
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestChatPruefe(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name  string
		chat  domain.Chat
		nimmt bool
	}{
		{"ein Satz", domain.Chat{Von: "anna", Text: "gut gezogen"}, true},
		{"ohne Sprecher", domain.Chat{Text: "hallo"}, false},
		{"Sprecher nur Leerzeichen", domain.Chat{Von: "  ", Text: "hallo"}, false},
		{"leerer Satz", domain.Chat{Von: "anna"}, false},
		{"nur Leerzeichen", domain.Chat{Von: "anna", Text: "   "}, false},
		{
			"genau die Grenze",
			domain.Chat{Von: "anna", Text: strings.Repeat("a", domain.Chatlaenge)},
			true,
		},
		{
			"eines zu viel",
			domain.Chat{Von: "anna", Text: strings.Repeat("a", domain.Chatlaenge+1)},
			false,
		},
		{
			// **Die Gegenprobe zur Zählweise.** In Bytes gezählt wären das
			// 1.000 und damit zu viel — in Zeichen sind es 500.
			"die Grenze in Umlauten",
			domain.Chat{Von: "anna", Text: strings.Repeat("ä", domain.Chatlaenge)},
			true,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.chat.Pruefe()

			if fall.nimmt && fehler != nil {
				t.Fatalf("der Satz sollte durchgehen, kam aber zurück: %v", fehler)
			}

			if !fall.nimmt && fehler == nil {
				t.Fatal("der Satz ging durch, obwohl er hätte anhalten sollen")
			}
		})
	}
}

func TestPraesenzartPruefe(t *testing.T) {
	t.Parallel()

	arten := domain.Praesenzarten()
	if len(arten) != 4 {
		t.Fatalf("es sollten vier Präsenzarten sein, es sind %d", len(arten))
	}

	for _, art := range arten {
		if fehler := art.Pruefe(); fehler != nil {
			t.Fatalf("%q ist eine der vier und wurde abgewiesen: %v", art, fehler)
		}
	}

	if domain.Praesenzart("winken").Pruefe() == nil {
		t.Fatal("eine erfundene Präsenzart ging durch")
	}
}

func TestPraesenzPruefe(t *testing.T) {
	t.Parallel()

	gut := domain.Praesenz{
		Von: "anna", Art: domain.PraesenzZeiger, Stand: 1, X: 612.5, Y: -140,
	}

	if fehler := gut.Pruefe(); fehler != nil {
		t.Fatalf("eine gültige Präsenz wurde abgewiesen: %v", fehler)
	}

	faelle := []struct {
		name     string
		praesenz domain.Praesenz
	}{
		{"ohne Sitz", domain.Praesenz{
			Art: domain.PraesenzZeiger, Stand: 1,
		}},
		{"unbekannte Art", domain.Praesenz{
			Von: "anna", Art: "winken", Stand: 1,
		}},
		{"ohne Stand", domain.Praesenz{
			Von: "anna", Art: domain.PraesenzZeiger,
		}},
		{"weit links von jedem Tisch", domain.Praesenz{
			Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
			X: -domain.Tischweite - 1,
		}},
		{"weit rechts von jedem Tisch", domain.Praesenz{
			Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
			X: domain.Tischweite + 1,
		}},
		{"gar keine Zahl", domain.Praesenz{
			Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
			X: math.NaN(),
		}},
		{"unendlich weit weg", domain.Praesenz{
			Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
			Y: math.Inf(1),
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if fall.praesenz.Pruefe() == nil {
				t.Fatal("die Präsenz ging durch, obwohl sie hätte anhalten sollen")
			}
		})
	}
}

func TestPraesenzkreisNimmtNurNeues(t *testing.T) {
	t.Parallel()

	kreis := domain.NeuerPraesenzkreis()

	zeige := func(sitz string, stand uint64) bool {
		return kreis.Nimm(domain.Praesenz{
			Von: sitz, Art: domain.PraesenzZeiger, Stand: stand,
		})
	}

	faelle := []struct {
		name  string
		sitz  string
		stand uint64
		gilt  bool
	}{
		{"die erste gilt", "anna", 5, true},
		{"dieselbe noch einmal nicht", "anna", 5, false},
		{"eine ältere nicht", "anna", 4, false},
		{"eine neuere schon", "anna", 6, true},
		{"ein anderer Sitz zählt für sich", "bert", 1, true},
		{"und anna bleibt, wo sie war", "anna", 6, false},
	}

	for _, fall := range faelle {
		if gilt := zeige(fall.sitz, fall.stand); gilt != fall.gilt {
			t.Fatalf("%s: erwartet %v, bekommen %v",
				fall.name, fall.gilt, gilt)
		}
	}
}

func TestPraesenzkreisVergisstEinenSitz(t *testing.T) {
	t.Parallel()

	kreis := domain.NeuerPraesenzkreis()

	spaet := domain.Praesenz{
		Von: "anna", Art: domain.PraesenzZeiger, Stand: 99,
	}

	if !kreis.Nimm(spaet) {
		t.Fatal("die erste Präsenz galt nicht")
	}

	frisch := domain.Praesenz{
		Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
	}

	if kreis.Nimm(frisch) {
		t.Fatal("ein Rückschritt galt, obwohl anna noch am Tisch saß")
	}

	// **Der Grund, aus dem es `Geht` gibt**: ein Gast, der neu verbindet,
	// fängt wieder bei eins an.
	kreis.Geht("anna")

	if !kreis.Nimm(frisch) {
		t.Fatal("nach dem Aufstehen galt der frische Anfang nicht")
	}
}

// **Vier Sitze zugleich, und jeder zählt für sich.**
//
// # Was dieser Lauf prüft und was der Detektor prüft
//
// **Er prüft die Zusage, nicht nur die Abwesenheit eines Rennens.** Jeder Sitz
// schickt seine zweihundert Stände aufsteigend; **alle** müssen gelten, denn
// ein Sitz überholt sich nicht selbst. Ein Kreis ohne Schloss verlöre hier
// Einträge — und der Race-Detektor sagt daneben, dass es keines gab.
func TestPraesenzkreisZaehltJedenSitzFuerSich(t *testing.T) {
	t.Parallel()

	const sitze, staende = 4, 200

	kreis := domain.NeuerPraesenzkreis()
	gilt := make([]int64, sitze)

	var warte sync.WaitGroup

	for sitz := range sitze {
		warte.Add(1)

		go func() {
			defer warte.Done()

			for stand := uint64(1); stand <= staende; stand++ {
				if kreis.Nimm(domain.Praesenz{
					Von:   string(rune('a' + sitz)),
					Art:   domain.PraesenzZeiger,
					Stand: stand,
				}) {
					atomic.AddInt64(&gilt[sitz], 1)
				}
			}
		}()
	}

	warte.Wait()

	for sitz := range sitze {
		if gezaehlt := atomic.LoadInt64(&gilt[sitz]); gezaehlt != staende {
			t.Errorf("Sitz %d: %d von %d Ständen galten",
				sitz, gezaehlt, staende)
		}
	}
}
