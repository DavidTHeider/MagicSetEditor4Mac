package app

// **Die fünf Zahlen des Sitzes gehen über die Naht** — PLAY-03
// (#1711).
//
// # Warum dieser Lauf durch `json` geht
//
// **Weil §3 genau diese Lücke benennt**: „Frontend-Tests gehen über
// `test/api.ts` und laufen nie durch die JSON-Serialisierung; Go-Tests sehen
// die Struct-Marken nicht." Eine `map` ohne Eintrag wird zu `null`, sobald
// sie `nil` ist — und dann liest das Fenster etwas anderes als eine leere
// Tabelle, bei grüner Pipeline.
//
// **Was er nicht prüft**: dass eine Fläche die Zahlen zeigt. Das hält der
// Lauf über `Spielerleiste` im Fenster.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineSichtMitZahlen gibt eine Sicht, deren einziger Sitz alle fünf trägt.
func eineSichtMitZahlen() domain.Partiesicht {
	return domain.Partiesicht{
		PartieID: "p1",
		Sitze: []domain.Sichtsitz{
			{
				Kennung:         "sitz-1",
				Name:            "David",
				Strahlung:       3,
				Commander:       "sitz-1/held/0",
				Commandersteuer: 2,
				Leergezogen:     true,
				Eroeffnung:      domain.EroeffnungOffen,
				Mulligans:       2,
				Unterlegen:      1,
				Commanderschaden: map[string]int{
					"sitz-2/drache/0": 7,
				},
			},
		},
	}
}

func TestDasSitzDTOTraegtDieFuenfZahlen(t *testing.T) {
	dto := partieAus(eineSichtMitZahlen())

	if len(dto.Sitze) != 1 {
		t.Fatalf("die Antwort hat %d Sitze", len(dto.Sitze))
	}

	sitz := dto.Sitze[0]

	if sitz.Strahlung != 3 || sitz.Commandersteuer != 2 ||
		sitz.Commander != "sitz-1/held/0" || !sitz.Leergezogen {
		t.Errorf("eine der vier Zahlen fehlt: %+v", sitz)
	}

	if sitz.Commanderschaden["sitz-2/drache/0"] != 7 {
		t.Errorf("der Commanderschaden steht bei %d",
			sitz.Commanderschaden["sitz-2/drache/0"])
	}
}

// **Eine leere Tabelle ist `{}` und nicht `null`.**
//
// Ohne diesen Lauf verschwände der Unterschied zwischen „kein Schaden" und
// „keine Auskunft" — und zwar genau dort, wo ihn kein Test sieht: das
// Fenster prüft `sitz.commanderschaden['x']` und bekäme einen Fehler statt
// einer Null.
func TestEineLeereSchadenstabelleWirdNichtZuNull(t *testing.T) {
	sicht := eineSichtMitZahlen()
	sicht.Sitze[0].Commanderschaden = map[string]int{}

	roh, fehler := json.Marshal(partieAus(sicht).Sitze[0])
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	if !strings.Contains(string(roh), `"commanderschaden":{}`) {
		t.Errorf("die leere Tabelle steht nicht als {} da: %s", roh)
	}
}

// **Und jede der fünf steht unter ihrem eigenen Namen im JSON.** Ein
// vergessenes `json:"…"` gäbe den Go-Namen mit großem Anfangsbuchstaben, und
// das Fenster suchte vergeblich.
func TestDieFuenfZahlenTragenIhreNamenImJSON(t *testing.T) {
	roh, fehler := json.Marshal(partieAus(eineSichtMitZahlen()).Sitze[0])
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	for _, name := range []string{
		`"strahlung":3`,
		`"commander":"sitz-1/held/0"`,
		`"commandersteuer":2`,
		`"leergezogen":true`,
		`"eroeffnung":"offen"`,
		`"mulligans":2`,
		`"unterlegen":1`,
		`"commanderschaden":{"sitz-2/drache/0":7}`,
	} {
		if !strings.Contains(string(roh), name) {
			t.Errorf("%s fehlt in %s", name, roh)
		}
	}
}
