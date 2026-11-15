package msepaket_test

import "testing"

// **Die Schlüsselwörter eines Spiels werden gelesen** —
// #1060.
//
// # Was gemessen ist
//
// Am 2026-08-24 am bezogenen `M15-Magic-Pack`: `magic.mse-game/keywords`
// führt **334** `keyword:`-Blöcke, **134** davon mit `<atom-param>`, dazu
// **sechs** `keyword mode:` und mehrere `keyword parameter type:`. Eingebunden
// wird die Datei über `include localized file: keywords` (`game`, Zeile 28).
//
// Bis hierher erreichte keiner davon die Anwendung: ein Set mit
// Schlüsselwörtern überlebte den Rundlauf als Rohblock, sichtbar war keines.
//
// # Die Felder
//
// `keyword`, `match`, `reminder`, `rules`, `mode` — nachgelesen an
// `src/data/keyword.cpp:86–93` (`IMPLEMENT_REFLECTION(Keyword)`), Fassung
// f9d9356 vom 2021-07-06.
func TestDieSchluesselwoerterEinesSpielsWerdenGelesen(t *testing.T) {
	paket := spielMitBlock(t, `keyword:
	keyword: Fliegend
	match: Fliegend
	mode: core
	reminder: Diese Kreatur kann nur von fliegenden Kreaturen geblockt werden.
keyword:
	keyword: Schutz vor
	match: Schutz vor <atom-param>name</atom-param>
	mode: expert
	rules: Schutz ist eine Zusammenfassung.
	reminder: {schutzcode(param1.value)}.
`)

	woerter := paket.Schluesselwoerter()

	if len(woerter) != 2 {
		t.Fatalf("%d Schlüsselwörter, erwartet zwei: %+v", len(woerter), woerter)
	}

	erstes := woerter[0]
	if erstes.Wort != "Fliegend" || erstes.Muster != "Fliegend" || erstes.Modus != "core" {
		t.Errorf("das erste ist %+v", erstes)
	}

	if erstes.Erinnerung == "" {
		t.Error("der Erinnerungstext fehlt")
	}

	// **Ein Schlüsselwort aus dem Spiel ist fest** — `Keyword::fixed`,
	// `src/data/keyword.hpp:104`: „true for game keywods, false for set (Fassung f9d9356 vom 2021-07-06)
	// keywords". Wer es ändern wollte, änderte das Paket.
	if !erstes.Fest {
		t.Error("ein Schlüsselwort aus dem Spiel muss fest sein")
	}

	if woerter[1].Regeln != "Schutz ist eine Zusammenfassung." {
		t.Errorf("»rules« wurde nicht gelesen: %+v", woerter[1])
	}
}

// **Die Modi werden gelesen, samt der Vorgabe.**
//
// `KeywordMode` hat drei Felder — `name`, `description`, `is_default`
// (`src/data/keyword.cpp:50–54`). Die Vorgabe entscheidet, welchen Modus ein (Fassung f9d9356 vom 2021-07-06)
// **neues** Schlüsselwort bekommt (`keyword.hpp:81`).
func TestDieSchluesselwortmodiWerdenGelesen(t *testing.T) {
	paket := spielMitBlock(t, `keyword mode:
	name: core
	description: Kernsatz
keyword mode:
	is default: true
	name: eigen
	description: Selbst gemacht
`)

	modi := paket.Schluesselwortmodi()

	if len(modi) != 2 {
		t.Fatalf("%d Modi, erwartet zwei: %+v", len(modi), modi)
	}

	if modi[0].Name != "core" || modi[0].Vorgabe {
		t.Errorf("der erste Modus ist %+v", modi[0])
	}

	if !modi[1].Vorgabe {
		t.Error("»is default: true« wurde nicht gelesen")
	}
}

// **Die Parameterarten werden gelesen.**
//
// `KeywordParam` hat dreizehn Felder (`src/data/keyword.cpp:35–49`). Gelesen (Fassung f9d9356 vom 2021-07-06)
// werden die, für die es hier einen Verwender gibt — dieselbe Überlegung wie
// bei den Bulkeinträgen in §16: „hier stehen die, für die es einen Verwender
// gibt."
func TestDieSchluesselwortparameterartenWerdenGelesen(t *testing.T) {
	paket := spielMitBlock(t, `keyword parameter type:
	name: mana
	match: [WUBRG0-9]+
keyword parameter type:
	name: kosten
	match: [ ][WUBRG0-9]*
	separator before is: [ —-]
	separator after is: [.]
	optional: true
`)

	arten := paket.Schluesselwortparameterarten()

	if len(arten) != 2 {
		t.Fatalf("%d Arten, erwartet zwei: %+v", len(arten), arten)
	}

	if arten[0].Name != "mana" || arten[0].Muster != "[WUBRG0-9]+" {
		t.Errorf("die erste Art ist %+v", arten[0])
	}

	// **`optional` steht ohne Angabe auf `true`.**
	//
	//	KeywordParam::KeywordParam()
	//	  : optional(true)
	//	  , eat_separator(true)
	//
	// `src/data/keyword.cpp`, Zeilen 24–27, Fassung f9d9356 vom 2021-07-06.
	//
	// **Hier stand bis zum 2026-08-24 das Gegenteil** — mit derselben
	// Fundstelle daneben, die es widerlegt. Der Test war der Grund, warum
	// der Fehler geprüft aussah: ein Wächter, der die falsche Zusage hält,
	// ist schlimmer als keiner. Angestrichen vom Zweitblick durch Codex über
	// Phase W; von 15 Parameterarten im echten Pack tragen 13 kein
	// `optional:`.
	if !arten[0].Wahlfrei {
		t.Error("ohne Angabe ist eine Parameterart wahlfrei — MSE setzt optional(true)")
	}

	// **Und `eat separator` ebenso** (ebd.). Im echten Pack setzt es keine
	// einzige Art.
	if !arten[0].TrennerFressen {
		t.Error("ohne Angabe wird der Trenner gefressen")
	}

	if !arten[1].Wahlfrei || arten[1].TrennerDavor != "[ —-]" {
		t.Errorf("die zweite Art ist %+v", arten[1])
	}
}

// **Ein Schlüsselwort ohne Muster wird übergangen.**
//
// # Warum
//
// `Keyword::prepare` endet mit `valid = !match_re.matches("")` — ein
// Schlüsselwort, dessen Ausdruck auf den leeren Text passt, wird abgeschaltet,
// „that would make MSE hang" (`src/data/keyword.cpp:207–208`). Ein leeres (Fassung f9d9356 vom 2021-07-06)
// `match` ist genau dieser Fall, und zwar der einfachste.
func TestEinSchluesselwortOhneMusterWirdUebergangen(t *testing.T) {
	paket := spielMitBlock(t, `keyword:
	keyword: Ohne
	mode: core
keyword:
	keyword: Mit
	match: Mit
	mode: core
`)

	woerter := paket.Schluesselwoerter()

	if len(woerter) != 1 || woerter[0].Wort != "Mit" {
		t.Fatalf("erwartet war nur »Mit«: %+v", woerter)
	}
}

// **Ohne Schlüsselwörter kommt eine leere Reihe, kein Fehler.**
func TestEinSpielOhneSchluesselwoerter(t *testing.T) {
	paket := spielMitBlock(t, "card fields:\n")

	if woerter := paket.Schluesselwoerter(); len(woerter) != 0 {
		t.Errorf("%d Schlüsselwörter, erwartet keines", len(woerter))
	}

	if modi := paket.Schluesselwortmodi(); len(modi) != 0 {
		t.Errorf("%d Modi, erwartet keinen", len(modi))
	}

	if arten := paket.Schluesselwortparameterarten(); len(arten) != 0 {
		t.Errorf("%d Parameterarten, erwartet keine", len(arten))
	}
}
