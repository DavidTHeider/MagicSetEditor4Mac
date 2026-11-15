package service_test

// **Die Parameterhilfen der Schlüsselwörter** — Werkstück D von Phase 4.11
// (#1342).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `KeywordsPanel::onCommand` und `runRefScript`
// (`src/gui/set/keywords_panel.cpp`, Zeilen 190–261, Fassung f9d9356 vom
// 2021-07-06). Zwei Hilfen, und sie zeigen auf **verschiedene Felder**:
//
//   - `insert parameter` bietet die Parameterarten des **Spiels** an und
//     schreibt `<atom-param>NAME</atom-param>` in `match`;
//   - `refer parameter` bietet die Parameter **dieses Schlüsselworts** an und
//     schreibt in `reminder` — `{paramN}`, oder das Ergebnis eines
//     `refer script:` mit `input = "paramN"`.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probearten sind zwei Parameterarten — eine mit Verweisskripten, eine ohne.
func probearten() []msepaket.Schluesselwortparameterart {
	return []msepaket.Schluesselwortparameterart{
		{
			Name: "zahl", Beschreibung: "eine Zahl", Muster: "[0-9]+",
			Verweisskripte: []msepaket.Schluesselwortverweis{},
		},
		{
			Name: "mana", Beschreibung: "Manasymbole", Muster: "[WUBRG]+",
			// **In MSEs eigener Schreibweise**, nicht in einer bequemeren.
			// `\{` ist eine geschweifte Klammer im Text, `{input}` ein
			// eingebetteter Ausdruck darin — genau so steht es in
			// `magic.mse-game/keywords_en`. Der erste Anlauf schrieb
			// `"{" + input + "}"` und bekam `+ input +` heraus: die Klammer
			// im Text hatte den Rest als Ausdruck geöffnet.
			Verweisskripte: []msepaket.Schluesselwortverweis{
				{Name: "normal", Beschreibung: "ohne Änderung", Skript: `\{{input}\}`},
				{Name: "gezählt", Beschreibung: "als Zahl", Skript: `\{zahl({input})\}`},
			},
		},
	}
}

// **Das Einfügeangebot nennt jede Parameterart des Spiels — samt ihrem Text.**
//
// Der Text ist das, was MSE einfügt: `<atom-param>` + Name + `</atom-param>`
// (Zeile 227). Wer nur den Namen gäbe, ließe jede Fläche die Marke selbst
// bauen — und die zweite Fassung liefe irgendwann auseinander.
func TestDasEinfuegeangebotNenntJedeParameterart(t *testing.T) {
	t.Parallel()

	angebot := service.Parameterangebot(probearten())

	if len(angebot) != 2 {
		t.Fatalf("%d Einträge: %+v", len(angebot), angebot)
	}

	if angebot[0].Name != "zahl" || angebot[0].Beschreibung != "eine Zahl" {
		t.Errorf("der erste lautet %+v", angebot[0])
	}

	if angebot[0].Einfuegen != "<atom-param>zahl</atom-param>" {
		t.Errorf("der Einfügetext lautet %q", angebot[0].Einfuegen)
	}
}

// **Ein Parameter ohne Verweisskripte bekommt die schlichte Form.**
//
//	if (p->refer_scripts.empty()) { … return "{" + param_s + "}"; }
func TestEinParameterOhneVerweisskripteBekommtDieSchlichteForm(t *testing.T) {
	t.Parallel()

	verweise, befunde := service.Parameterverweise(t.Context(),
		"Beben <atom-param>zahl</atom-param>", probearten(), "")
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(verweise) != 1 {
		t.Fatalf("%d Verweise: %+v", len(verweise), verweise)
	}

	if verweise[0].Nummer != 1 || verweise[0].Art != "zahl" ||
		verweise[0].Einfuegen != "{param1}" {
		t.Errorf("der Verweis lautet %+v", verweise[0])
	}
}

// **Ein Parameter mit Verweisskripten bekommt je Skript einen Eintrag.**
//
// Und jeder rechnet mit `input = "paramN"` — das ist die Zeile, die MSE
// ausdrücklich setzt (`ctx.setVariable(SCRIPT_VAR_input, to_script(param_s))`).
func TestEinParameterMitVerweisskriptenBekommtJeSkriptEinenEintrag(t *testing.T) {
	t.Parallel()

	verweise, befunde := service.Parameterverweise(t.Context(),
		"Kicker <atom-param>mana</atom-param>", probearten(), "")
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(verweise) != 2 {
		t.Fatalf("%d Verweise: %+v", len(verweise), verweise)
	}

	if verweise[0].Name != "normal" || verweise[0].Einfuegen != "{param1}" {
		t.Errorf("der erste lautet %+v", verweise[0])
	}

	if verweise[1].Einfuegen != "{zahl(param1)}" {
		t.Errorf("der zweite lautet %+v", verweise[1])
	}
}

// **Die Nummer zählt die Parameter, nicht die Einträge.**
//
// Zwei Parameter mit je zwei Verweisskripten geben vier Einträge — aber die
// Nummern sind 1, 1, 2, 2. MSEs `++param` steht **außerhalb** der inneren
// Schleife, und `param_s` heißt für beide Skripte desselben Parameters
// dasselbe.
func TestDieNummerZaehltDieParameterUndNichtDieEintraege(t *testing.T) {
	t.Parallel()

	verweise, _ := service.Parameterverweise(t.Context(),
		"X <atom-param>mana</atom-param> Y <atom-param>zahl</atom-param>",
		probearten(), "")

	if len(verweise) != 3 {
		t.Fatalf("%d Verweise: %+v", len(verweise), verweise)
	}

	erwartet := []int{1, 1, 2}
	for stelle, nummer := range erwartet {
		if verweise[stelle].Nummer != nummer {
			t.Errorf("Eintrag %d trägt Nummer %d, erwartet %d",
				stelle, verweise[stelle].Nummer, nummer)
		}
	}

	if verweise[2].Einfuegen != "{param2}" {
		t.Errorf("der dritte lautet %+v", verweise[2])
	}
}

// **Eine unbekannte Parameterart ist ein Befund und kein Eintrag.**
//
// MSE meldet sie beim Vorbereiten des Schlüsselworts; hier gilt dasselbe (§8):
// wer sie stillschweigend überginge, zeigte ein Menü, das den Parameter im
// Muster nicht kennt.
func TestEineUnbekannteParameterartIstEinBefund(t *testing.T) {
	t.Parallel()

	verweise, befunde := service.Parameterverweise(t.Context(),
		"Beben <atom-param>gibtsnicht</atom-param>", probearten(), "")

	if len(verweise) != 0 {
		t.Errorf("die Verweise lauten %+v", verweise)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	}
}

// **Ein Muster ohne Parameter gibt eine leere Liste und keinen Befund.**
func TestEinMusterOhneParameterGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	verweise, befunde := service.Parameterverweise(t.Context(),
		"Fliegend", probearten(), "")

	if verweise == nil || len(verweise) != 0 || len(befunde) != 0 {
		t.Errorf("Verweise %+v, Befunde %+v", verweise, befunde)
	}
}

// **Ein Verweisskript, das scheitert, nimmt nur seinen Eintrag mit.**
//
// MSE fängt den Fehler je Skript (`catch (ScriptError const& e)`) und gibt für
// diesen Eintrag nichts zurück. Ein Menü, das an einem kaputten Skript ganz
// ausfiele, nähme die guten mit.
func TestEinKaputtesVerweisskriptNimmtNurSeinenEintragMit(t *testing.T) {
	t.Parallel()

	arten := probearten()
	// **Ein Textskript scheitert an seiner Einbettung**, nicht am Text
	// darum: `\{…\}` ist Text, `{…}` ein Ausdruck.
	arten[1].Verweisskripte[0].Skript = `\{{gibtesnicht(input)}\}`

	verweise, befunde := service.Parameterverweise(t.Context(),
		"Kicker <atom-param>mana</atom-param>", arten, "")

	if len(verweise) != 1 || verweise[0].Name != "gezählt" {
		t.Errorf("die Verweise lauten %+v", verweise)
	}

	if len(befunde) != 1 {
		t.Errorf("die Befunde lauten %+v", befunde)
	}
}

// **Eine Marke ohne Ende bleibt Text.**
//
// # Warum das kein Fehler ist
//
// Weil das Muster ein Feld ist, in das jemand gerade tippt. Wer
// `<atom-param>` schreibt und noch nicht `</atom-param>`, hat einen halben
// Einschub dastehen — und ein Menü, das darauf mit einer Fehlermeldung
// antwortete, meldete sich bei jedem zweiten Tastendruck.
//
// **Dieselbe Duldung wie in [ausdruckFuer]** (§8): der Rest wird nicht mehr
// nach Parametern durchsucht, und was schon gefunden war, bleibt.
func TestEineMarkeOhneEndeBleibtText(t *testing.T) {
	t.Parallel()

	verweise, befunde := service.Parameterverweise(t.Context(),
		"Beben <atom-param>zahl</atom-param> und <atom-param>mana",
		probearten(), "")

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	// **Der erste zählt, der halbe nicht.**
	if len(verweise) != 1 || verweise[0].Art != "zahl" {
		t.Errorf("die Verweise lauten %+v", verweise)
	}
}
