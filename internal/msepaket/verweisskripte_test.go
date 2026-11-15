package msepaket_test

// **`refer script:` — womit ein Erinnerungstext auf einen Parameter zeigt** —
// Werkstück D von Phase 4.11
// (#1342,
// #1349).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `KeywordsPanel::runRefScript` (`src/gui/set/keywords_panel.cpp`, Zeilen
// 236–261, Fassung f9d9356 vom 2021-07-06):
//
//	String param_s = String("param") << ++param;
//	if (p->refer_scripts.empty()) {
//	  … return "{" + param_s + "}";
//	} else {
//	  … ctx.setVariable(SCRIPT_VAR_input, to_script(param_s));
//	     return r->script.invoke(ctx)->toString();
//	}
//
// Eine Parameterart **ohne** Verweisskripte gibt `{paramN}`; eine **mit**
// bietet je Skript einen Eintrag, und das Skript rechnet mit `input =
// "paramN"`.
//
// # Und was im Pack steht — gemessen am 2026-08-31
//
// **Zwei Zahlen, und sie messen Verschiedenes.** Über alle Dateien des Packs
// stehen **161** `keyword parameter type:`-Blöcke, **46** davon mit
// `refer script:` — das zählt jede Sprachfassung mit. Was `magic.mse-game`
// **wirksam** mitbringt, nachdem `include localized file:` aufgelöst ist,
// sind **15** Arten, **6** mit Verweisen und **19** Skripte (beides gemessen
// am 2026-08-31). Die zweite Zahl ist die, an der eine Oberfläche hängt.
//
// Die Art `mana` bringt drei mit: `\{{input}\}`, `\{cmc({input})\}` und
// `\{colored_mana({input})\}`.

import "testing"

// **Ein `refer script:` wird mit allen drei Angaben gelesen.**
func TestEinVerweisskriptWirdGelesen(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `keyword parameter type:
	name: mana
	match: [WUBRG]+
	refer script:
		name: normal
		description: Ohne Änderung
		script: \{{input}\}
	refer script:
		name: manawert
		description: Rechnet Mana in eine Zahl
		script: \{cmc({input})\}
`)

	arten := paket.Schluesselwortparameterarten()
	if len(arten) != 1 {
		t.Fatalf("%d Parameterarten", len(arten))
	}

	verweise := arten[0].Verweisskripte
	if len(verweise) != 2 {
		t.Fatalf("%d Verweisskripte: %+v", len(verweise), verweise)
	}

	if verweise[0].Name != "normal" || verweise[0].Beschreibung != "Ohne Änderung" {
		t.Errorf("das erste lautet %+v", verweise[0])
	}

	// **Der Skripttext bleibt roh.** `\{` ist MSEs Maskierung einer
	// geschweiften Klammer im Text; wer sie hier auflöste, machte aus einem
	// Ergebnis ein Skript.
	if verweise[1].Skript != `\{cmc({input})\}` {
		t.Errorf("das zweite Skript lautet %q", verweise[1].Skript)
	}
}

// **Eine Parameterart ohne `refer script:` trägt eine leere Liste.**
//
// **Und keine `nil`** — dieselbe Zusage wie überall sonst (§3): wer sie liest,
// prüft auf eines statt auf zwei. In `magic.mse-game` sind das 9 von 15.
func TestEineParameterartOhneVerweisskripteTraegtEineLeereListe(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `keyword parameter type:
	name: zahl
	match: [0-9]+
`)

	arten := paket.Schluesselwortparameterarten()
	if len(arten) != 1 || arten[0].Verweisskripte == nil {
		t.Errorf("die Arten lauten %+v", arten)
	}

	if len(arten[0].Verweisskripte) != 0 {
		t.Errorf("die Verweisskripte lauten %+v", arten[0].Verweisskripte)
	}
}
