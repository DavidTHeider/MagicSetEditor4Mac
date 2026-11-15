package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fassungsvorlage baut eine Vorlage, deren Symbolschrift `T` **dreimal** führt.
//
// **So wie das echte Paket es tut.** `magic-mana-small` trennt die drei
// Zapfsymbole über `enabled: { mana_t() == … }`, und die Reihenfolge in der
// Datei setzt die ältere Fassung nach vorn.
func fassungsvorlage(initskript string) service.Vorlage {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	vorlage := planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{
		"mana": {Name: "mana", Symbole: []msepaket.Symbol{
			{Code: "T", Bild: "t_older.png", Bedingung: `{ mana_t() == "older" }`},
			{Code: "T", Bild: "t_old.png", Bedingung: `{ mana_t() == "old" }`},
			{Code: "T", Bild: "t.png"},
		}},
	})

	vorlage.Stil.Initskript = initskript

	return vorlage
}

// symbolbilder liest die Bildpfade des Symbolzugs eines Plans.
func symbolpfadeAus(t *testing.T, plan *render.Plan) []string {
	t.Helper()

	var pfade []string

	for stelle := range plan.Ebenen {
		for _, abschnitt := range plan.Ebenen[stelle].Abschnitte {
			for i := range abschnitt.Symbole {
				symbol := abschnitt.Symbole[i]

				pfade = append(pfade, symbol.Bild.Beschreibung())
			}
		}
	}

	return pfade
}

// **Die Fassung, deren Bedingung erfüllt ist, gewinnt — nicht die erste.**
//
// `SymbolFont::split` geht die Symbole in **Dateireihenfolge** durch und
// überspringt jedes, dessen `enabled:` nicht gilt (Zeilen 191–230, Fassung
// f9d9356 vom 2021-07-06). Bis zum 2026-08-17 nahm dieses Programm den ersten
// Eintrag und zeichnete damit `mana_t_older.png`, wo MSE `mana_t.png` zeichnet
// (#293).
//
// **Die Gegenprobe** steht im Fall „ohne Bedingung gewinnt die erste Fassung":
// dieselbe Schrift, dieselbe Karte, nur eine andere Antwort auf `mana_t()` —
// und ein anderes Bild. Ein Wächter, der immer dasselbe Bild fände, prüfte die
// Reihenfolge und nicht die Bedingung.
func TestDieFassungMitDerErfuelltenBedingungGewinnt(t *testing.T) {
	faelle := []struct {
		name     string
		skript   string
		erwartet string
	}{
		{"neu — keine Bedingung trifft", `mana_t := { "new" }`, "/mana.mse-symbol-font/t.png"},
		{"alt", `mana_t := { "old" }`, "/mana.mse-symbol-font/t_old.png"},
		{"älter — die erste Fassung", `mana_t := { "older" }`, "/mana.mse-symbol-font/t_older.png"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan, befunde := gedeutetMitVorlage(t,
				kartemit("casting cost", "T"), fassungsvorlage(fall.skript))
			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			pfade := symbolpfadeAus(t, plan)
			if len(pfade) != 1 || pfade[0] != fall.erwartet {
				t.Errorf("gezeichnet wird %v, erwartet %q", pfade, fall.erwartet)
			}
		})
	}
}

// **Eine Bedingung, die sich nicht ausrechnen lässt, gilt als erfüllt.**
//
// Das ist MSEs Verhalten: `enabled` ist ein `Scriptable<bool>` mit dem Vorgabe­
// wert `true`; scheitert die Auswertung, bleibt der alte Wert stehen
// (`SymbolInFont::SymbolInFont`, Zeile 103, und `Scriptable::update`).
//
// **Und sie wird benannt.** Ein Symbol, das nur deshalb gezeichnet wird, weil
// niemand seine Bedingung ausrechnen konnte, ist eine Auskunft wert (§8).
func TestEineUnausrechenbareBedingungGiltAlsErfuellt(t *testing.T) {
	plan, befunde := gedeutetMitVorlage(t,
		kartemit("casting cost", "T"), fassungsvorlage(""))

	pfade := symbolpfadeAus(t, plan)
	if len(pfade) != 1 || pfade[0] != "/mana.mse-symbol-font/t_older.png" {
		t.Errorf("gezeichnet wird %v, erwartet die erste Fassung", pfade)
	}

	gefunden := false

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "mana_t") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("kein Befund über die unausrechenbare Bedingung: %v", befunde)
	}
}

// **Ein Ausdruck, der sich nicht lesen lässt, trifft nicht — statt zu viel zu
// treffen.**
//
// Zerlegt wird nur die Bauform, die MSEs Schriften benutzen: eine Alternative,
// deren Zweige auf eine negative Vorausschau enden dürfen. Steht die
// Vorausschau **mitten** im Zweig, verlöre ein Abschneiden den Rest — `a(?!b)c`
// träfe dann jedes `a`. Solch ein Code wird übersprungen, und das Symbol
// dahinter kommt zum Zug
// (#293).
func TestEinUnlesbarerAusdruckTrifftNicht(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	vorlage := planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{
		"mana": {Name: "mana", Symbole: []msepaket.Symbol{
			{Code: `R(?!x)y`, Bild: "ausdruck.png", Ausdruck: true},
			{Code: "R", Bild: "r.png"},
		}},
	})

	plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "R"), vorlage)
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	pfade := symbolpfadeAus(t, plan)
	if len(pfade) != 1 || pfade[0] != "/mana.mse-symbol-font/r.png" {
		t.Errorf("gezeichnet wird %v, erwartet das gewöhnliche Symbol", pfade)
	}
}

// **Die Reihenfolge in der Datei entscheidet, nicht die Länge des Codes.**
//
// `SymbolFont::split` nimmt den **ersten** Eintrag, dessen Code am Anfang des
// Restes steht — die Liste steht in Dateireihenfolge (Zeilen 191–230). Bis zum
// 2026-08-17 sortierte dieses Programm nach Länge, längster zuerst. In
// `magic-mana-large` stehen die zusammengesetzten Codes ohnehin vorn, und
// deshalb fiel es nie auf; eine Schrift, die es andersherum hält, bekäme bei
// uns eine andere Zerlegung als bei MSE.
func TestDieDateireihenfolgeEntscheidetUeberDieZerlegung(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	// `2` steht **vor** `2R`: MSE zerlegt `2R` deshalb in zwei Symbole, nicht
	// in das eine, dessen Code länger ist.
	vorlage := planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{
		"mana": {Name: "mana", Symbole: []msepaket.Symbol{
			{Code: "2", Bild: "2.png"},
			{Code: "2R", Bild: "zweirot.png"},
			{Code: "R", Bild: "r.png"},
		}},
	})

	plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "2R"), vorlage)
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	pfade := symbolpfadeAus(t, plan)
	if len(pfade) != 2 ||
		pfade[0] != "/mana.mse-symbol-font/2.png" ||
		pfade[1] != "/mana.mse-symbol-font/r.png" {
		t.Errorf("gezeichnet wird %v, erwartet 2.png und r.png", pfade)
	}
}

// **Ohne `always symbol:` bleibt ein Feld Text, auch wenn alles passt.**
//
// MSE zerlegt ein ganzes Feld nur dann in Symbole, wenn der Stil es sagt
// (`src/render/text/element.cpp`, Zeile 244, Fassung f9d9356 vom 2021-07-06).
// Das Feld `text` von `magic-m15` nennt eine Symbolschrift und **keine** solche
// Ansage: sein Wert ist Regeltext, in dem nur `<sym>`-Stellen Symbole sind.
//
// **Die Gegenprobe ist der Auffangcode.** `magic-mana-large` endet auf
// `[0-9](?!/[WUBRGCSTQ2])|.` — ein Ausdruck, der jedes Zeichen deckt. Ohne die
// Ansage wäre damit jeder Regeltext eine Reihe von Bildern
// (#293).
func TestOhneAnsageBleibtEinFeldText(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		Schrift:       msepaket.Schrift{Name: "MPlantin", Groesse: "15"},
	}

	vorlage := planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{
		"mana": {Name: "mana", Symbole: []msepaket.Symbol{{Code: "R", Bild: "r.png"}}},
	})

	plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "R"), vorlage)
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if pfade := symbolpfadeAus(t, plan); len(pfade) != 0 {
		t.Errorf("gezeichnet werden Symbole (%v) — die Ansage `always symbol:` fehlt aber", pfade)
	}
}

// **Ein Auffangcode deckt jedes Zeichen — mit Vorausschau.**
//
// `magic-mana-large` schließt mit `[0-9](?!/[WUBRGCSTQ2])|.`: eine Ziffer, die
// **nicht** Teil eines Hybridcodes ist, sonst irgendein Zeichen. Ohne diesen
// Eintrag hat die Schrift für `1` gar kein Symbol — die Ziffern selbst hängen
// dort an `enabled: {ancestral_mana()}` und fallen bei `false` weg.
//
// Gos `regexp` kennt keine Vorausschau; das Muster wird deshalb an seiner
// `(?!…)` zerlegt (`symbolmuster.go`).
func TestEinAuffangcodeMitVorausschauTrifft(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
		Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
		ImmerSymbole:  true,
	}

	vorlage := planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{
		"mana": {Name: "mana", Symbole: []msepaket.Symbol{
			{Code: "1/W", Bild: "hybrid.png"},
			{Code: "W", Bild: "w.png"},
			{Code: `[0-9](?!/[W])|.`, Bild: "kreis.png", Ausdruck: true},
		}},
	})

	faelle := []struct {
		name, wert string
		erwartet   string
	}{
		// Die `1` ist keine Hybridziffer: der Auffangcode greift.
		{"eine Ziffer", "1", "/mana.mse-symbol-font/kreis.png"},
		// Vor `/W` greift er **nicht** — dort gewinnt der Hybridcode.
		{"eine Hybridziffer", "1/W", "/mana.mse-symbol-font/hybrid.png"},
		// Und der zweite Zweig deckt, was sonst niemand deckt.
		{"ein fremdes Zeichen", "&", "/mana.mse-symbol-font/kreis.png"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", fall.wert), vorlage)
			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			pfade := symbolpfadeAus(t, plan)
			if len(pfade) != 1 || pfade[0] != fall.erwartet {
				t.Errorf("gezeichnet wird %v, erwartet %q", pfade, fall.erwartet)
			}
		})
	}
}

// **Eine Bedingung darf ein schlichter Wert sein, kein Ausdruck.**
//
// MSE liest `enabled: false` genauso wie `enabled: { false }` — beides ist ein
// `Scriptable<bool>`. Und ein Wert, der **kein** Wahrheitswert ist, lässt die
// Fassung gelten und wird benannt: eine Zahl als Bedingung ist ein Fehler in
// der Vorlage, kein Grund, ein Symbol verschwinden zu lassen (§8).
func TestEineBedingungOhneKlammernUndEineOhneWahrheitswert(t *testing.T) {
	faelle := []struct {
		name, bedingung, erwartet string
		befund                    bool
	}{
		{"ohne Klammern, falsch", "false", "/mana.mse-symbol-font/zweit.png", false},
		{"ohne Klammern, wahr", "true", "/mana.mse-symbol-font/erst.png", false},
		{"kein Wahrheitswert", "{ 5 }", "/mana.mse-symbol-font/erst.png", true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			feld := msepaket.Feldstil{
				Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
				Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
				ImmerSymbole:  true,
			}

			vorlage := planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{
				"mana": {Name: "mana", Symbole: []msepaket.Symbol{
					{Code: "R", Bild: "erst.png", Bedingung: fall.bedingung},
					{Code: "R", Bild: "zweit.png"},
				}},
			})

			plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "R"), vorlage)

			if pfade := symbolpfadeAus(t, plan); len(pfade) != 1 || pfade[0] != fall.erwartet {
				t.Errorf("gezeichnet wird %v, erwartet %q", pfade, fall.erwartet)
			}

			if (len(befunde) > 0) != fall.befund {
				t.Errorf("Befunde %v, erwartet welche: %v", befunde, fall.befund)
			}
		})
	}
}
