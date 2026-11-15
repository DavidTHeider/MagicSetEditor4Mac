package xai

// **Das Schemasubset, das xAI annimmt** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Warum das hier steht und nicht beim Anbieter bleibt
//
// **Das Issue verlangt es**: „Unterstützte Schemasubsets lokal
// nachvalidieren." Der Grund ist derselbe wie bei [Frage.pruefe]: ein Schema,
// das der Dienst abweist, **kostet den Ruf trotzdem**.
//
// **Und das Schema kommt nicht von hier.** Es entsteht in `service` aus den
// Feldern einer **Kartenvorlage** — aus einem fremden Paket also, das
// irgendwer irgendwann geschrieben hat. Was darin steht, ist keine
// Entscheidung dieses Programms.
//
// # Vier Fälle, nachgelesen und nicht ausgedacht
//
// Am 2026-09-12 unter
// <https://docs.x.ai/developers/model-capabilities/text/structured-outputs>:
// xAI nennt ausdrücklich, was es mit **400** beantwortet.
//
//  1. `enum` oder `anyOf` **ohne einen einzigen** Wert
//  2. eine Eigenschaft, deren Schema `true` oder `false` ist
//  3. `maxContains` und `minContains`
//  4. `items` als **Liste** — dafür gibt es `prefixItems`
//
// # Und was hier **nicht** abgewiesen wird
//
// **Alles, was xAI »best effort« nennt**: `not`, `if`/`then`/`else`,
// mehrteiliges `allOf`, Längen über den garantierten Grenzen. Es wird
// **angenommen** und nur nicht erzwungen. Ein Prüfer, der mehr abwiese als
// der Dienst, sähe sorgfältig aus und verböte dem Aufrufer etwas, das geht
// (§10).

import (
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dieVerboteneWorte sind die Schlüssel, die xAI mit 400 beantwortet.
var dieVerboteneWorte = []string{"maxContains", "minContains"}

// dieSchemafelder sind die Stellen, an denen ein **Schema** stehen kann.
//
// **Sie stehen hier, weil der Wahrheitswert nur dort einer ist.**
// `additionalProperties: false` ist der Normalfall und kein Schema an einer
// Eigenschaft; wer beide über einen Kamm schöre, wiese jedes übliche Schema
// ab.
var dieSchemafelder = []string{
	"properties", "$defs", "definitions",
	"patternProperties",
}

// pruefeSchema geht ein Schema durch und weist ab, was xAI abweist.
func pruefeSchema(schema map[string]any) error {
	return pruefeStelle(schema, "")
}

// pruefeStelle prüft eine Stelle im Baum — `wo` nennt sie beim Namen.
func pruefeStelle(schema map[string]any, wo string) error {
	for _, wort := range dieVerboteneWorte {
		if _, steht := schema[wort]; steht {
			return schemafehler(wo,
				"xAI nimmt %s nicht an und antwortet darauf mit 400", wort)
		}
	}

	if fehler := pruefeLeereWahl(schema, wo); fehler != nil {
		return fehler
	}

	if _, istListe := schema["items"].([]any); istListe {
		return schemafehler(wo,
			"»items« steht als Liste da — xAI nimmt dafür »prefixItems«")
	}

	return pruefeTiefer(schema, wo)
}

// pruefeLeereWahl weist eine Wahl ohne eine einzige Möglichkeit ab.
//
// **Sie ist kein Schreibfehler, sondern eine Aussage**: „nichts ist gültig".
// Nur kann niemand sie erfüllen, und xAI sagt das mit 400.
func pruefeLeereWahl(schema map[string]any, wo string) error {
	for _, wort := range []string{"enum", "anyOf", "oneOf"} {
		wahl, istListe := schema[wort].([]any)
		if istListe && len(wahl) == 0 {
			return schemafehler(wo,
				"»%s« steht ohne einen einzigen Wert da — xAI antwortet "+
					"darauf mit 400", wort)
		}
	}

	return nil
}

// pruefeTiefer steigt in die Unterschemata ab.
//
// **Weil ein Kartenschema verschachtelt ist**: die Felder einer Karte stehen
// unter `properties`, und was darunter liegt, ist wieder ein Schema. Ein
// Prüfer, der nur die oberste Ebene ansähe, wäre grün und hielte nichts.
func pruefeTiefer(schema map[string]any, wo string) error {
	for schluessel, wert := range schema {
		if slices.Contains(dieSchemafelder, schluessel) {
			if fehler := pruefeUnterschemata(wert, wo); fehler != nil {
				return fehler
			}

			continue
		}

		if fehler := pruefeWert(wert, wo+"/"+schluessel); fehler != nil {
			return fehler
		}
	}

	return nil
}

// pruefeUnterschemata prüft die benannten Schemata unter einem Feld.
func pruefeUnterschemata(wert any, wo string) error {
	benannte, istKarte := wert.(map[string]any)
	if !istKarte {
		return nil
	}

	for name, unter := range benannte {
		// **Hier ist ein Wahrheitswert ein Schema** — und genau der ist
		// gemeint.
		if _, istWahrheit := unter.(bool); istWahrheit {
			return schemafehler(wo+"/"+name,
				"das Schema ist ein Wahrheitswert — xAI nimmt das nicht an")
		}

		if fehler := pruefeWert(unter, wo+"/"+name); fehler != nil {
			return fehler
		}
	}

	return nil
}

// pruefeWert steigt in einen Wert ab, wenn er ein Schema sein kann.
func pruefeWert(wert any, wo string) error {
	switch genauer := wert.(type) {
	case map[string]any:
		return pruefeStelle(genauer, wo)

	case []any:
		for nummer, stueck := range genauer {
			if fehler := pruefeWert(stueck,
				fmt.Sprintf("%s[%d]", wo, nummer)); fehler != nil {
				return fehler
			}
		}
	}

	return nil
}

// schemafehler baut die Meldung — mit der Stelle, wenn es eine gibt.
//
// **Die Stelle ist die halbe Auskunft.** Ein Kartenschema hat je Feld einen
// Eintrag; „irgendwo steht ein leeres enum" ließe jemanden suchen.
func schemafehler(wo, satz string, teile ...any) error {
	gesagt := fmt.Sprintf(satz, teile...)

	if wo == "" {
		return fmt.Errorf("%w: %s", domain.ErrValidation, gesagt)
	}

	return fmt.Errorf("%w: %s (bei %s)", domain.ErrValidation, gesagt, wo)
}
