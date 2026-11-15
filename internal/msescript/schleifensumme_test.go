package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine Schleife addiert ihre Ergebnisse — sie sammelt sie nicht.**
//
//	// the loop body should have a net stack effect of 0, but the entire expression of +1
//	// solution: add all results from the body, start with nil
//	script.addInstruction(I_PUSH_CONST, script_nil);      //    push nil
//	…
//	parseOper(input, script, PREC_SET, I_BINARY, I_ADD);  // EEE;  add
//
// `src/script/parser.cpp`, Zeilen 554–608, Fassung f9d9356 vom 2021-07-06.
//
// Das ist keine Feinheit: `for x from 1 to 9 do [wert]` ist die Art, in der die
// Vorlagen eine **flache** Liste bauen — `nil + [a]` ist `[a]`, `[a] + [b]` ist
// `[a,b]`. Wer stattdessen sammelt, bekommt eine Liste von Listen, und der
// nächste Zugriff darauf rechnet mit einer Liste statt mit einer Zahl
// (#390).
func TestEineSchleifeAddiertIhreErgebnisse(t *testing.T) {
	faelle := []struct {
		name, quelle, erwartet string
		art                    msescript.WertArt
		glieder                int
	}{
		{
			name: "Listen werden verkettet", quelle: `for x from 1 to 3 do [x*10]`,
			art: msescript.ArtAbbildung, glieder: 3, erwartet: "102030",
		},
		{
			name: "Zahlen werden addiert", quelle: `for x from 1 to 3 do x`,
			art: msescript.ArtZahlwert, erwartet: "6",
		},
		{
			name: "Texte werden aneinandergesetzt", quelle: `for x from 1 to 3 do "a"`,
			art: msescript.ArtTextwert, erwartet: "aaa",
		},
		{
			name: "auch `for each`", quelle: `for each x in [1,2] do [x]`,
			art: msescript.ArtAbbildung, glieder: 2, erwartet: "12",
		},
		{
			name: "eine leere Spanne bleibt nichts", quelle: `for x from 1 to 0 do [x]`,
			art: msescript.ArtNichts, erwartet: "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.NeueUmgebung(nil))
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Art != fall.art {
				t.Errorf("Art %v, erwartet %v", wert.Art, fall.art)
			}

			if fall.glieder > 0 && len(wert.Eintraege) != fall.glieder {
				t.Errorf("%d Glieder, erwartet %d", len(wert.Eintraege), fall.glieder)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("Text %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Und ein Glied der verketteten Liste ist wieder eine Zahl.**
//
// Das ist der Zugriff, an dem die 43 Befunde hingen: `excess_height()[input-1]`
// steht in einer Rechnung, und eine Liste an dieser Stelle macht daraus eine
// Zeichenkette.
func TestEinGliedDerVerkettetenListeIstEineZahl(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`(for x from 1 to 3 do [x*10])[1]`, msescript.NeueUmgebung(nil))
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Art != msescript.ArtZahlwert || wert.Zahl != 20 {
		t.Errorf("Art %v, Zahl %v — erwartet die Zahl 20", wert.Art, wert.Zahl)
	}
}

// **`+` verkettet zwei Sammlungen.**
//
//	} else if (at == SCRIPT_COLLECTION && bt == SCRIPT_COLLECTION) {
//	    a = make_intrusive<ScriptConcatCollection>(a, b);
//
// `src/script/context.cpp`, Zeile 465 ff.; die Zahlenschlüssel werden dabei
// umnummeriert (`ScriptConcatCollection::getMember`, `src/script/value.cpp`,
// Zeilen 551–562).
//
// Fassung f9d9356 vom 2021-07-06.
func TestZweiListenWerdenVerkettet(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(), `[1] + [2]`,
		msescript.NeueUmgebung(nil))
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if len(wert.Eintraege) != 2 {
		t.Fatalf("%d Glieder, erwartet zwei: %+v", len(wert.Eintraege), wert.Eintraege)
	}

	if wert.AlsText() != "12" {
		t.Errorf("Text %q, erwartet %q", wert.AlsText(), "12")
	}
}

// **Benannte Schlüssel bleiben, wie sie waren.**
//
// Eine Abbildung mit Namen ist keine Liste; sie wird weiter verschmolzen. Dass
// MSEs `getMember` dabei die **linke** Seite zuerst fragt, ist eine zweite
// Abweichung und ein eigenes Werkstück (#390).
func TestBenannteSchluesselBleibenVerschmolzen(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(), `[a: 1] + [b: 2]`,
		msescript.NeueUmgebung(nil))
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if len(wert.Eintraege) != 2 {
		t.Errorf("%d Glieder, erwartet zwei", len(wert.Eintraege))
	}
}

// **Scheitert die Addition, scheitert die Schleife — mit Stelle.**
//
// Eine Funktion lässt sich nur mit einer Funktion verbinden. Trifft im zweiten
// Durchgang eine auf die Zahl aus dem ersten, ist das ein Fehler der Vorlage,
// und er gehört gemeldet statt verschluckt (§8).
func TestEineSchleifeMeldetEineGescheiterteAddition(t *testing.T) {
	faelle := map[string]string{
		"die Zählschleife": `for x from 1 to 2 do ` +
			`(if x == 1 then 5 else replace@(match: "a", replace: "b"))`,
		"die Listenschleife": `for each x in [1,2] do ` +
			`(if x == 1 then 5 else replace@(match: "a", replace: "b"))`,
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
			if err == nil {
				t.Fatal("kein Fehler, obwohl Zahl und Funktion nicht zusammengehen")
			}
		})
	}
}
