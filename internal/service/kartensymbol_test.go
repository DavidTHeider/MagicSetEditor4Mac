package service_test

// **Ein Symbolfeld an der Karte zeichnet** — Werkstück A von Phase 4.11
// (#1339).
//
// # Was gemessen wurde, bevor gebaut wurde
//
// Das Issue verlangt „die Rasterung des Feldwerts (`render`)" und liest sich,
// als fehlte ein Zeichenweg. **Er fehlt nicht.**
//
// Nachgesehen am 2026-08-31 in `full-magic-pack`: die Vorlagen zeichnen ihr
// Setsymbol über das **Skript**, nicht über die Feldart —
//
//	symbol_variation(symbol: set.symbol, variation: "common")
//
// (`magic-m15-duplex.mse-style`, Zeile 221). `symbol_variation` nimmt einen
// **Dateinamen** entgegen und weiß nicht, aus welchem Feld er kommt. Damit
// zeichnet `card.<feld>` genauso wie `set.symbol` — vorausgesetzt, der
// Kartenwert steht in der Skriptumgebung, und das tut er für jedes Feld.
//
// **Dieser Lauf hat die Vermutung geprüft, nicht bestätigt bekommen.** Er war
// beim ersten Anlauf grün — die Rasterung fehlte nicht.
//
// # Was der Karte trotzdem fehlt, und was nicht
//
// Die **Fassungen** eines Symbolfeldes liest `msepaket` unter `set info style:`
// und `default set style:` — beides Setblöcke. Für ein Kartenfeld stünden sie
// unter `card style:`, und dort liest sie niemand.
//
// **Gebaut wird das nicht, und der Grund ist gezählt** (§10: die Reichweite
// einer Messung ist die Reichweite ihrer Aussage). Über `full-magic-pack`,
// 2026-08-31: **411 `variation:`-Blöcke**, davon 366 unter `set info style:`
// und 38 unter `default set style:` — **keiner** unter `card style:`. Eine
// dritte Lesung hätte im ganzen Bestand keinen Aufrufer, und §2 nennt so etwas
// eine Zusage ohne Einlösung.
//
// **Die erste Messung war falsch**, und das steht hier, damit niemand sie
// wiederholt: ein `grep` nach `variation:` fand 261 Vorkommen unter
// `card style:` — allesamt das **Argument** in
// `symbol_variation(symbol: …, variation: "common")`. Gezählt werden musste
// der Blockkopf, nicht die Zeichenfolge.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein `symbol`-Kartenfeld zeichnet wie ein Setsymbol.**
func TestEinSymbolfeldAnDerKarteZeichnet(t *testing.T) {
	t.Parallel()

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Symbolfassungen: []msepaket.Symbolfassung{
			{Feld: "kartensymbol", Name: "gewoehnlich", Fuellfarbe: "rgb(0,128,0)"},
		},
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "100", Hoehe: "20",
			Bild:    `{ symbol_variation(symbol: card.kartensymbol, variation: "gewoehnlich") }`,
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "x"},
		{Field: "kartensymbol", Text: "eigen.mse-symbol"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{
			Stil: stil,
			Kartenfelder: []domain.Field{
				{Name: "text", Kind: domain.KindText},
				{Name: "kartensymbol", Kind: domain.KindSymbol},
			},
		})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		t.Logf("Befund %s: %s", befund.Ort, befund.Meldung)
	}

	if gab := fuellfarbeIm(plan); !strings.Contains(gab, "rgb(0,128,0)") {
		t.Errorf("gezeichnet wird %q, erwartet war die Fassung des Kartenfeldes", gab)
	}
}
