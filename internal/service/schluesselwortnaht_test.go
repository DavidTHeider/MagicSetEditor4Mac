// Die Naht: `expand_keywords` über die Umgebung
// (#1061).
//
// **Von `schluesselwoerter_test.go` getrennt**, als die Datei über 400 Zeilen
// wuchs (§2). Der Schnitt folgt der Frage: dort der **Motor** — was trifft,
// wie wird geordnet, was wird ersetzt —, hier die **Naht**: sieht die Sprache
// die Wortliste, und findet `combine:` die Namen, die MSE dort setzt?
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Weg durch die Umgebung: `expand_keywords` setzt wirklich ein.**
//
// # Warum dieser Lauf über die Sprache geht
//
// Die Läufe darüber prüfen den Motor. Dieser prüft die **Naht**: dass
// `expand_keywords` in der Umgebung des Planbauers die Wortliste sieht, dass
// der Erinnerungstext als **Textskript** ausgewertet wird und dass `combine:`
// die Namen findet, die MSE dort setzt.
//
// Ohne ihn stünde der Motor da und niemand riefe ihn — dieselbe Gestalt wie
// #852.
func TestExpandKeywordsSetztUeberDieUmgebungEin(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Schutz vor", Muster: "Schutz vor <atom-param>name</atom-param>",
		Modus: "expert", Erinnerung: "schützt vor {param1}",
	}}, []msepaket.Schluesselwortparameterart{{Name: "name", Muster: "[A-Za-zÄÖÜäöü]+"}})

	umgebung := service.UmgebungMitSchluesselwoertern(liste)

	wert, fehler := msescript.AuswertenQuelle(t.Context(),
		`expand_keywords(input: "Schutz vor Rot.", combine: { keyword + " (" + reminder + ")" })`,
		umgebung)
	if fehler != nil {
		t.Fatalf("auswerten: %v", fehler)
	}

	const erwartet = "Schutz vor Rot (schützt vor Rot)."
	if wert.AlsText() != erwartet {
		t.Errorf("gesetzt %q, erwartet %q", wert.AlsText(), erwartet)
	}
}

// **`mode` steht dem `combine:` zur Verfügung.**
//
// Die M15-Vorlage entscheidet daran, wie ein Erinnerungstext aussieht:
// `if mode == "pseudo" then "<i-auto>{keyword}</i-auto>" else …`.
func TestDerModusStehtDemZusammensetzenZurVerfuegung(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Eile", Muster: "Eile", Modus: "pseudo", Erinnerung: "schnell",
	}}, nil)

	wert, fehler := msescript.AuswertenQuelle(t.Context(),
		`expand_keywords(input: "Eile", combine: { mode + ":" + keyword })`,
		service.UmgebungMitSchluesselwoertern(liste))
	if fehler != nil {
		t.Fatalf("auswerten: %v", fehler)
	}

	if wert.AlsText() != "pseudo:Eile" {
		t.Errorf("gesetzt %q", wert.AlsText())
	}
}

// **Ein Fehler im Erinnerungstext kostet seinen Treffer, nicht die Karte.**
//
//	catch (const Error& e) {
//	  handle_error(_ERROR_2_("in keyword reminder", e.what(), keyword.keyword));
//	}
//
// `src/data/keyword.cpp`, Zeilen 733–736, Fassung f9d9356 vom 2021-07-06.
func TestEinFehlerImErinnerungstextKostetNurSeinenTreffer(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Kaputt", Muster: "Kaputt", Erinnerung: "{gibtsnichtundistkeinname()}",
	}}, nil)

	wert, fehler := msescript.AuswertenQuelle(t.Context(),
		`expand_keywords(input: "Kaputt bleibt", combine: { keyword + " (" + reminder + ")" })`,
		service.UmgebungMitSchluesselwoertern(liste))
	if fehler != nil {
		t.Fatalf("der Fehler kam bis zur Karte durch: %v", fehler)
	}

	if wert.AlsText() != "Kaputt bleibt" {
		t.Errorf("gesetzt %q — erwartet war der unveränderte Text", wert.AlsText())
	}
}

// **Und die Fehlerwege der Naht.**
func TestDieFehlerwegeDerSchluesselwortnaht(t *testing.T) {
	t.Parallel()

	umgebung := service.UmgebungMitSchluesselwoertern(service.Wortliste{})

	for name, quelle := range map[string]string{
		"ohne Text":     `expand_keywords(combine: { keyword })`,
		"ohne combine:": `expand_keywords(input: "x")`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := msescript.AuswertenQuelle(t.Context(), quelle, umgebung); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein `combine:`, das selbst scheitert, kostet seinen Treffer.**
func TestEinScheiterndesZusammensetzenKostetSeinenTreffer(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Eile", Muster: "Eile", Erinnerung: "schnell",
	}}, nil)

	wert, fehler := msescript.AuswertenQuelle(t.Context(),
		`expand_keywords(input: "Eile bleibt", combine: { gibtsnichtundistkeinname() })`,
		service.UmgebungMitSchluesselwoertern(liste))
	if fehler != nil {
		t.Fatalf("der Fehler kam bis zur Karte durch: %v", fehler)
	}

	if wert.AlsText() != "Eile bleibt" {
		t.Errorf("gesetzt %q", wert.AlsText())
	}
}

// **`condition:` entscheidet, ob ein Treffer zählt.**
//
//	if (options.match_condition && options.match_condition->eval(ctx)->toBool() == false) {
//	  return {false,it};
//	}
//
// `src/data/keyword.cpp`, Zeilen 713–716, Fassung f9d9356 vom 2021-07-06.
//
// # Warum das sichtbar ist
//
// Die M15-Vorlage schreibt `correct_case or (mode != "pseudo" and not
// used_placeholders)`. Ohne die Bedingung stand auf `White.mse-set`, Karte 3,
// **zweimal** derselbe Erinnerungstext — einmal hinter „Flying" und einmal
// mitten im Satz hinter „gains flying".
func TestDieBedingungEntscheidetUeberEinenTreffer(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Fliegend", Muster: "Fliegend", Erinnerung: "erinnert",
	}}, nil)

	umgebung := service.UmgebungMitSchluesselwoertern(liste)

	wert, fehler := msescript.AuswertenQuelle(t.Context(),
		`expand_keywords(input: "Fliegend und fliegend", condition: { correct_case }, `+
			`combine: { keyword + " (" + reminder + ")" })`, umgebung)
	if fehler != nil {
		t.Fatalf("auswerten: %v", fehler)
	}

	const erwartet = "Fliegend (erinnert) und fliegend"
	if wert.AlsText() != erwartet {
		t.Errorf("gesetzt %q, erwartet %q", wert.AlsText(), erwartet)
	}
}

// **Eine Bedingung, die sich nicht ausrechnen lässt, weist ab.**
//
// Dieselbe Vorsicht wie beim Erinnerungstext: eine unklare Bedingung ist kein
// Grund, einen Erinnerungstext zu setzen, den niemand angeordnet hat.
func TestEineUnbrauchbareBedingungWeistAb(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Eile", Muster: "Eile", Erinnerung: "erinnert",
	}}, nil)

	for name, quelle := range map[string]string{
		"ein Fehler":      `expand_keywords(input: "Eile", condition: { gibtsnichtundistkeinname() }, combine: { keyword })`,
		"keine Wahrheit":  `expand_keywords(input: "Eile", condition: { "vielleicht" }, combine: { keyword })`,
		"schlicht falsch": `expand_keywords(input: "Eile", condition: { false }, combine: { keyword })`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wert, fehler := msescript.AuswertenQuelle(t.Context(), quelle,
				service.UmgebungMitSchluesselwoertern(liste))
			if fehler != nil {
				t.Fatalf("auswerten: %v", fehler)
			}

			if wert.AlsText() != "Eile" {
				t.Errorf("gesetzt %q — erwartet war der unveränderte Text", wert.AlsText())
			}
		})
	}
}

// **`default_expand:` entscheidet, ob der Erinnerungstext sichtbar ist.**
//
//	expand = options.expand_default ? options.expand_default->eval(ctx)->toBool() : true;
//
// `src/data/keyword.cpp`, Zeile 721, Fassung f9d9356 vom 2021-07-06.
//
// # Warum das mehr ist als ein Schalter
//
// Die M15-Vorlage liest darin die Setangabe `automatic_reminder_text` — eine
// Mehrfachwahl, in der steht, für welche Modi ein Erinnerungstext erscheinen
// soll. **In allen zehn kanonischen Sets ist sie leer**, gemessen am
// 2026-08-24: MSE zeigt dort also keinen einzigen Erinnerungstext, und dieses
// Programm tut jetzt dasselbe.
//
// Wer sie auf `core, expert` setzt, bekommt sie — nachgesehen an
// `White.mse-set`, Karte 3: der Erinnerungstext steht **einmal** hinter
// „Flying" und nicht hinter „gains flying".
func TestDieVorgabeEntscheidetUeberDieSichtbarkeit(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Eile", Muster: "Eile", Erinnerung: "erinnert",
	}}, nil)

	umgebung := service.UmgebungMitSchluesselwoertern(liste)

	for name, fall := range map[string]struct{ vorgabe, erwartet string }{
		"sichtbar":    {"true", "Eile [erinnert]"},
		"nicht":       {"false", "Eile []"},
		"unbrauchbar": {`"vielleicht"`, "Eile []"},
		"mit Fehler":  {"gibtsnichtundistkeinname()", "Eile []"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wert, fehler := msescript.AuswertenQuelle(t.Context(),
				`expand_keywords(input: "Eile", default_expand: { `+fall.vorgabe+` }, `+
					`combine: { keyword + " [" + (if expand then reminder else "") + "]" })`,
				umgebung)
			if fehler != nil {
				t.Fatalf("auswerten: %v", fehler)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("gesetzt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
