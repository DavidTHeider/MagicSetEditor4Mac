//go:build fremdpakete

// **Die Feldreihen an einem echten Spielpaket** — Werkstück 3 von Phase 3.6
// (#1167).
//
// # Was dieser Lauf hält, das kein Doppel halten kann
//
// **Dass die Regel an echten Feldnamen dasselbe ergibt wie an gebauten.** Die
// Zahlen in `internal/domain/feldreihe.go` sind an `magic.mse-game` erhoben;
// eine Messung, die niemand nachprüft, veraltet still (§10).
//
//	MSE_FREMDPAKETE=~/full-magic-pack/data \
//	  go test -tags fremdpakete ./internal/msepaket/ -run Feldreihen -v
package msepaket_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDieFeldreihenVonMagic(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	paket, fehler := msepaket.Oeffne(filepath.Join(wurzel, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	// **Die Module werden eingebunden**, bevor gelesen wird: `magic.mse-game`
	// führt seine Kartenfelder über `include file:` — ohne diesen Schritt
	// stünde ein Spiel mit null Feldern da, und der Lauf prüfte nichts.
	// **Ein Fehler dabei hält den Lauf nicht an**, wie in `rahmenfremd_test.go`
	// und aus demselben Grund: `magic.mse-game/game` bindet `statistics` und
	// `keywords` ohne Sprachkürzel ein, und im Pack liegen nur
	// `statistics_de`, `keywords_en` und so fort. Die **Kartenfelder** kommen
	// trotzdem — und um die geht es hier. Die Prüfung darunter merkt es,
	// wenn nicht: null Felder wären null Reihen.
	_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // siehe oben

	spiel, _ := paket.Spielmodell()

	// **Die drei Zahlen, die im Kommentar stehen** — erhoben am 2026-08-27.
	//
	// Sie stehen hier als **Gleichheit** und nicht als Untergrenze: eine
	// Untergrenze bliebe grün, wenn ein Paketwechsel die Hälfte der Reihen
	// wegnähme, und genau das wäre die Nachricht.
	if len(spiel.Fields) != 249 {
		t.Errorf("%d Kartenfelder — der Kommentar nennt 249", len(spiel.Fields))
	}

	reihen := spiel.Feldreihen()
	if len(reihen) != 43 {
		t.Errorf("%d Reihen — der Kommentar nennt 43", len(reihen))
	}

	// **Und die Gegenprobe: wie viele Felder eine Zahl tragen, ohne eine Reihe
	// zu bilden.** Ohne sie sagte die Zahl darüber nur, dass die Regel
	// *irgendetwas* trifft — nicht, dass sie das Richtige auslässt.
	muster := regexp.MustCompile(`^(.+) ([1-9]\d*)$`)

	vorhanden := make(map[string]bool, len(spiel.Fields))
	for i := range spiel.Fields {
		vorhanden[spiel.Fields[i].Name] = true
	}

	ohneGrundfeld := 0

	for i := range spiel.Fields {
		teile := muster.FindStringSubmatch(spiel.Fields[i].Name)
		if len(teile) > 0 && !vorhanden[teile[1]] {
			ohneGrundfeld++
		}
	}

	if ohneGrundfeld != 40 {
		t.Errorf("%d Felder mit Zahl ohne Grundfeld — der Kommentar nennt 40", ohneGrundfeld)
	}

	// **Und die höchste Stelle** — Befund 12 des Zweitblicks über Phase 3.6.
	//
	// Der Kommentar nennt sie („die höchste Stelle ist 8"), und der Wächter
	// prüfte sie nicht: eine Änderung von 8 auf 9 wäre bei unveränderten
	// Anzahlen grün geblieben. Ein Wächter, der die halbe Zusage hält, macht
	// die andere unsichtbar (§10).
	hoechste := 0

	for _, reihe := range reihen {
		for _, stelle := range reihe.Stellen {
			if stelle.Nummer > hoechste {
				hoechste = stelle.Nummer
			}
		}
	}

	if hoechste != 8 {
		t.Errorf("die höchste Stelle ist %d — der Kommentar nennt 8", hoechste)
	}

	// **`name` ist eine Reihe, `loyalty cost` keine.** Die beiden Fälle, an
	// denen die Regel hängt — namentlich geprüft, damit ein Zahlendreher
	// nicht als Erfolg durchgeht.
	nachGrundfeld := map[string][]int{}
	for _, reihe := range reihen {
		for _, stelle := range reihe.Stellen {
			nachGrundfeld[reihe.Grundfeld] = append(nachGrundfeld[reihe.Grundfeld], stelle.Nummer)
		}
	}

	if len(nachGrundfeld["name"]) != 3 {
		t.Errorf("die Reihe »name« hat %v — erwartet sind die Stellen 2, 3 und 4",
			nachGrundfeld["name"])
	}

	if _, da := nachGrundfeld["loyalty cost"]; da {
		t.Error("»loyalty cost« steht als Reihe da — es gibt kein Feld dieses Namens")
	}
}
