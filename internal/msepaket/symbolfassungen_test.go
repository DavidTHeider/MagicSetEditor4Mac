package msepaket_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Symbolschrift führt denselben Code mehrfach und trennt die Fassungen
// über `enabled:`.**
//
// `magic-mana-large` führt `T` und `Q` je **dreimal** und trägt insgesamt 54
// `enabled:`-Bedingungen. MSE geht die Liste **in Dateireihenfolge** durch und
// nimmt den ersten Eintrag, dessen Bedingung erfüllt ist
// (`SymbolFont::split`, Zeilen 191–230, Fassung f9d9356 vom 2021-07-06).
//
// Ohne die Bedingung im Modell gewinnt bei uns der erste Eintrag — und das ist
// bei `T` die Fassung `mana_t_older.png`, während MSE `mana_t.png` zeichnet
// (#293).
func TestEinSymbolNenntSeineBedingung(t *testing.T) {
	pfad := probesymbolschrift(t, `mse version: 2.0.0
symbol:
	code: T
	image: mana_t_older.png
	enabled: { mana_t() == "older" }
symbol:
	code: T
	image: mana_t.png
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	schrift, befunde := paket.Symbolmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(schrift.Symbole) != 2 {
		t.Fatalf("%d Symbole, erwartet zwei: %+v", len(schrift.Symbole), schrift.Symbole)
	}

	if schrift.Symbole[0].Bedingung != `{ mana_t() == "older" }` {
		t.Errorf("die Bedingung ist %q", schrift.Symbole[0].Bedingung)
	}

	// **Ohne Angabe keine Bedingung**, nicht `true` als Zeichenkette: der
	// Unterschied zwischen „gilt immer" und „gilt, wenn dieser Ausdruck wahr
	// ist" gehört ins Modell, sonst wertet der Aufrufer eine Bedingung aus, die
	// niemand geschrieben hat.
	if schrift.Symbole[1].Bedingung != "" {
		t.Errorf("das zweite Symbol nennt %q, erwartet keine Bedingung",
			schrift.Symbole[1].Bedingung)
	}
}

// **Ein Code darf ein regulärer Ausdruck sein, und das steht am Symbol.**
//
// `magic-mana-large` setzt `regex: yes` an 32 Symbolen — `([XYZ0-9])/W`,
// `10(?!/)`, und als letzten Eintrag den Auffangausdruck
// `[0-9](?!/[WUBRGCSTQ2])|.`. Wer die Angabe nicht liest, vergleicht solch
// einen Code **wörtlich** und trifft nie
// (`SymbolFont::split`, Zeilen 196–216).
func TestEinSymbolcodeDarfEinAusdruckSein(t *testing.T) {
	pfad := probesymbolschrift(t, `mse version: 2.0.0
symbol:
	code: ([XYZ0-9])/W
	image: hybrid.png
	regex: yes
symbol:
	code: W
	image: w.png
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	schrift, _ := paket.Symbolmodell()

	if len(schrift.Symbole) != 2 {
		t.Fatalf("%d Symbole, erwartet zwei", len(schrift.Symbole))
	}

	if !schrift.Symbole[0].Ausdruck {
		t.Error("der erste Code ist kein Ausdruck — `regex: yes` wurde nicht gelesen")
	}

	if schrift.Symbole[1].Ausdruck {
		t.Error("der zweite Code ist ein Ausdruck, obwohl `regex:` fehlt")
	}
}

// **`yes`, `true` und `1` heißen ja — alles andere nein.**
//
// MSE schreibt `yes`; die anderen beiden stehen in seinem Leser daneben. Eine
// unbekannte Schreibweise als „ja" zu deuten machte aus einem gewöhnlichen
// Code stillschweigend einen Ausdruck, und der träfe dann anders.
func TestNurBekannteSchreibweisenHeissenJa(t *testing.T) {
	faelle := []struct {
		angabe   string
		erwartet bool
	}{
		{"yes", true},
		{"true", true},
		{"1", true},
		{"YES", true},
		{"no", false},
		{"false", false},
		{"0", false},
		{"", false},
	}

	for _, fall := range faelle {
		t.Run("regex: "+fall.angabe, func(t *testing.T) {
			pfad := probesymbolschrift(t, "mse version: 2.0.0\nsymbol:\n\tcode: W\n"+
				"\timage: w.png\n\tregex: "+fall.angabe+"\n")

			paket, err := msepaket.Oeffne(pfad)
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() {
				if fehler := paket.Schliessen(); fehler != nil {
					t.Errorf("schließen: %v", fehler)
				}
			})

			schrift, _ := paket.Symbolmodell()
			if len(schrift.Symbole) != 1 {
				t.Fatalf("%d Symbole, erwartet eines", len(schrift.Symbole))
			}

			if schrift.Symbole[0].Ausdruck != fall.erwartet {
				t.Errorf("%q ergibt %v, erwartet %v",
					fall.angabe, schrift.Symbole[0].Ausdruck, fall.erwartet)
			}
		})
	}
}

// **Ein Feldstil sagt, ob sein Wert *immer* aus Symbolen besteht.**
//
// `always symbol: true` steht **133-mal** in den 190 Stilen von
// `M15-Magic-Pack` — an den Manakosten, nicht am Regeltext. MSE entscheidet
// daran, ob ein Feld überhaupt in Symbole zerlegt wird
// (`src/render/text/element.cpp`, Zeile 244, Fassung f9d9356 vom 2021-07-06);
// ohne die Angabe ist der Wert Text, in dem nur `<sym>`-Stellen Symbole sind.
//
// **Ohne sie muss man raten.** Dieses Programm hat bis zum 2026-08-17 gefragt,
// ob sich der ganze Wert in Symbole zerlegen lässt — und das trifft, sobald
// eine Schrift einen Auffangcode führt, auf **jeden** Text zu
// (#293).
func TestEinFeldstilSagtObSeinWertImmerAusSymbolenBesteht(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
card style:
	casting cost:
		left: 10
		symbol font:
			name: mana
			size: 15
		always symbol: true
	text:
		left: 10
		symbol font:
			name: mana
			size: 14
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()

	if len(stil.Felder) != 2 {
		t.Fatalf("%d Feldstile, erwartet zwei", len(stil.Felder))
	}

	if !stil.Felder[0].ImmerSymbole {
		t.Error("die Manakosten sind nicht als reines Symbolfeld gelesen")
	}

	if stil.Felder[1].ImmerSymbole {
		t.Error("der Regeltext gilt als reines Symbolfeld, obwohl er es nicht sagt")
	}
}

// probestilpaket legt ein `.mse-style` mit dem gegebenen Inhalt an.
func probestilpaket(t *testing.T, inhalt string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pfad, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}

// **Polster und Zeilenhöhen werden roh gelesen.**
//
// Sie dürfen Skripte sein; ausgerechnet wird in `internal/service`, wo eine
// Umgebung dasteht (§2). Im Pack stehen `padding top:` 279-mal und die drei
// Zeilenhöhen je rund 290-mal
// (#345).
func TestEinFeldstilNenntPolsterUndZeilenhoehen(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
card style:
	text:
		left: 29
		padding left: 6
		padding right: 4
		padding top: 3
		padding bottom: { 2 }
		line height soft: 0.9
		line height hard: 1.2
		line height line: 1.5
		line height soft max: 1.0
		line height hard max: 1.3
		line height line max: 1.6
	ohne:
		left: 0
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()

	if len(stil.Felder) != 2 {
		t.Fatalf("%d Feldstile, erwartet zwei", len(stil.Felder))
	}

	polster := stil.Felder[0].Polster
	if polster != (msepaket.Polsterangabe{
		Links: "6", Rechts: "4", Oben: "3", Unten: "{ 2 }",
	}) {
		t.Errorf("das Polster ist %+v", polster)
	}

	hoehen := stil.Felder[0].Zeilenhoehen
	if hoehen != (msepaket.Zeilenhoehenangabe{
		Weich: "0.9", Hart: "1.2", Linie: "1.5",
		WeichMax: "1.0", HartMax: "1.3", LinieMax: "1.6",
	}) {
		t.Errorf("die Zeilenhöhen sind %+v", hoehen)
	}

	// **Ohne Angabe bleibt beides leer** — und nicht etwa null als Zeichenkette.
	if stil.Felder[1].Polster != (msepaket.Polsterangabe{}) ||
		stil.Felder[1].Zeilenhoehen != (msepaket.Zeilenhoehenangabe{}) {
		t.Errorf("das zweite Feld nennt %+v und %+v",
			stil.Felder[1].Polster, stil.Felder[1].Zeilenhoehen)
	}
}

// **`alignment:` darf ein `script:`-Block sein.**
//
// 216 der 1.597 Ausrichtungen im Pack stehen so; der Wert des Knotens ist dann
// leer (#345).
func TestEineAusrichtungAlsSkriptblockWirdGelesen(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
card style:
	text:
		alignment:
			script:
				if x then "middle center" else "middle left"
	schlicht:
		alignment: bottom right
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()

	if len(stil.Felder) != 2 {
		t.Fatalf("%d Feldstile, erwartet zwei", len(stil.Felder))
	}

	if !strings.Contains(stil.Felder[0].AusrichtungSkript, "middle center") {
		t.Errorf("das Skript ist %q", stil.Felder[0].AusrichtungSkript)
	}

	// **Eine schlichte Angabe bleibt der Wert** — und trägt kein Skript.
	if stil.Felder[1].Ausrichtung != "bottom right" ||
		stil.Felder[1].AusrichtungSkript != "" {
		t.Errorf("das zweite Feld ist %q / %q",
			stil.Felder[1].Ausrichtung, stil.Felder[1].AusrichtungSkript)
	}
}
