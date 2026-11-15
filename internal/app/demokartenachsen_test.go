//go:build fremdpakete

// Die **Achsen** des Demokartensatzes
// (#476).
//
// # Warum es diese Datei gibt
//
// Die sieben Demokarten waren **historisch, nicht systematisch**: jede ist
// entstanden, weil ein Befund sie verlangt hat — „Illustriert" wegen #380,
// „Artefakt" wegen #390, der Fuß wegen #382. Was nie vorgekommen ist, fehlte,
// und niemand sah, dass es fehlte.
//
// Hier stehen die Achsen und ein Wächter darüber. Er sagt nicht „es sind genug
// Karten", sondern: **jeder Achsenwert ist belegt** — und was nicht belegt ist,
// steht namentlich mit Grund daneben (§10).
package app_test

import (
	"sort"
	"strings"
	"testing"
)

// rahmenfarben sind die Wahlen des Feldes `card color`.
//
// **Aus dem Spielpaket gelesen, nicht erdacht** (§15):
// `magic.mse-game/card_fields`, Zeilen 104–147. `colorless` steht dort als
// `empty choice`, die übrigen als `choice:`.
//
// **Fünf davon sind Untermarken und keine eigenen Rahmen**: `horizontal`,
// `vertical`, `radial`, `overlay` und `reversed` sind `enabled`, sobald zwei
// Farben gewählt sind — sie sagen, **wie** ein mehrfarbiger Rahmen geteilt
// wird, nicht welche Farbe er hat. Sie stehen hier trotzdem, damit die
// Aufzählung vollständig ist und die Lücke sichtbar bleibt.
var rahmenfarben = []string{
	"colorless", "white", "blue", "black", "red", "green",
	"artifact", "land", "multicolor", "hybrid",
	"horizontal", "vertical", "radial", "overlay", "reversed",
}

// vorlagen sind die Grundformen, die `M15-Magic-Pack` als eigene Stile führt.
//
// **Gezählt über `~/Library/Application Support/MagicSetEditor4Mac/pakete/M15-Magic-Pack/data`**:
// 190 Stilpakete, davon 149 mit dem Präfix `magic-m15-`. Diese neun sind die
// Formen, die eine Karte grundlegend anders zeichnen — nicht die 140
// Umfärbungen darüber.
var vorlagen = []string{
	"magic-m15",
	"magic-m15-saga",
	"magic-m15-planeswalker",
	"magic-m15-token",
	"magic-m15-leveler",
	"magic-m15-flip",
	"magic-m15-textless",
	"magic-m15-textless-land",
	"magic-m15-split-fusable",
}

// textformen sind die Gestalten, die ein Regeltext annimmt.
//
// **Sie sind der Grund für die zweistufige Setzung** (#283): eine Aufzählung
// hängt anders um als ein Absatz, und eine Auszeichnung ändert die Breite.
var textformen = []string{
	"leer", "einzeilig", "mehrzeilig", "aufzaehlung", "auszeichnung",
}

// ungedeckt nennt die Achsenwerte, für die es **keine** Demokarte gibt — und
// warum.
//
// **Eine Ausnahme ohne Grund ist eine Ausrede** (§10). Wer hier etwas einträgt,
// schreibt dazu, was ihn hindert; wer eine Karte baut, streicht die Zeile.
var ungedeckt = map[string]string{}

// TestJederAchsenwertIstBelegt hält den Demokartensatz gegen seine Achsen.
//
// **Die Gegenprobe steht in `TestDerAchsenwaechterWirdRot`**: ohne sie wäre
// nicht belegt, dass dieser Wächter überhaupt greift.
func TestJederAchsenwertIstBelegt(t *testing.T) {
	achsen := []struct {
		name   string
		werte  []string
		belegt func([]demokarte) map[string]bool
	}{
		{"Rahmenfarbe", rahmenfarben, belegteRahmenfarben},
		{"Vorlage", vorlagen, belegteVorlagen},
		{"Textform", textformen, belegteTextformen},
	}

	for _, achse := range achsen {
		t.Run(achse.name, func(t *testing.T) {
			belegt := achse.belegt(demokarten())

			var fehlend []string

			for _, wert := range achse.werte {
				if belegt[wert] {
					continue
				}

				if grund, bekannt := ungedeckt[wert]; bekannt {
					t.Logf("  %s: nicht belegt — %s", wert, grund)

					continue
				}

				fehlend = append(fehlend, wert)
			}

			sort.Strings(fehlend)

			if len(fehlend) > 0 {
				t.Errorf("%d Werte der Achse %s sind weder belegt noch begründet: %s",
					len(fehlend), achse.name, strings.Join(fehlend, ", "))
			}
		})
	}
}

// belegteRahmenfarben sammelt, welche Rahmenfarben vorkommen.
//
// **Ein Feld `card color` trägt eine Liste**: `black, red, multicolor` nennt
// drei Werte auf einmal.
func belegteRahmenfarben(karten []demokarte) map[string]bool {
	belegt := map[string]bool{}

	for stelle := range karten {
		farbe := strings.TrimSpace(karten[stelle].felder["card color"])
		if farbe == "" {
			belegt["colorless"] = true

			continue
		}

		for _, teil := range strings.Split(farbe, ",") {
			belegt[strings.TrimSpace(teil)] = true
		}
	}

	return belegt
}

// belegteVorlagen sammelt die benutzten Stile.
func belegteVorlagen(karten []demokarte) map[string]bool {
	belegt := map[string]bool{}
	for stelle := range karten {
		belegt[karten[stelle].stil] = true
	}

	return belegt
}

// belegteTextformen sammelt die Gestalten des Regeltextes.
func belegteTextformen(karten []demokarte) map[string]bool {
	belegt := map[string]bool{}

	for stelle := range karten {
		belegt[textform(karten[stelle].felder["rule text"])] = true
	}

	return belegt
}

// textform sagt, welche Gestalt ein Regeltext hat.
//
// **Die Reihenfolge ist die Feinheit**: ein Text mit Aufzählung ist auch
// mehrzeilig, und die Aufzählung ist die seltenere Aussage.
func textform(text string) string {
	switch {
	case strings.TrimSpace(text) == "":
		return "leer"
	case strings.Contains(text, "<bullet>"):
		return "aufzaehlung"
	case strings.Contains(text, "<i>") || strings.Contains(text, "<sym>"):
		return "auszeichnung"
	case strings.Contains(text, "\n"):
		return "mehrzeilig"
	default:
		return "einzeilig"
	}
}
