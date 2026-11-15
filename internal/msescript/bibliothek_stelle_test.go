package msescript_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Funktionen, die #216 aus den echten Vorlagen und dem Spielskript
// nachgezogen hat. Jede steht mit ihrer Fundstelle bei ihrer Umsetzung.
func TestDieNachgezogenenBibliotheksfunktionen(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		// `break_text` gibt die **Treffer**, `split_text` das dazwischen.
		{
			"break_text sammelt die Treffer",
			`to_text(break_text("2RG", match:"[WUBRG0-9]"))`, "2RG",
		},
		{
			"break_text zählt die Treffer",
			`length(break_text("2RG", match:"[WUBRG]"))`, "2",
		},
		{
			"break_text ohne Treffer",
			`length(break_text("xyz", match:"[WUBRG]"))`, "0",
		},

		// **`position` über Text sucht `in` in `of`** (#2006) — gefragt am
		// Orakel; die Fälle stehen in `stelletext_test.go`.
		{"position in einer Liste", `position(of:2, in:[1,2,3])`, "1"},
		{"position, was nicht dasteht", `position(of:9, in:[1,2])`, "-1"},
		{"position in einem Text", `position(of:"abc", in:"c")`, "2"},

		// `to_number` neben `to_int` — der leere Text ergibt null.
		{"to_number aus Text", `to_number("13")`, "13"},
		{"to_number aus dem leeren Text", `to_number("")`, "0"},

		// **Und es schneidet **nicht** ab.** Hier stand bis zum 2026-08-22
		// `13` — eine Erwartung aus dem Kopf, nicht vom Orakel: MSE probiert
		// erst `ToLong` und dann `ToDouble`, und die zweite gelingt
		// (#859).
		{"to_number schneidet nicht ab", `"" + to_number("13.9")`, "13.9"},

		// `remove_tag(input, tag: "<word-list")` — der Anfang zählt.
		{
			"remove_tag nimmt eine benannte Auszeichnung",
			`remove_tag("<word-list-land>Wald</word-list-land>", tag: "<word-list")`, "Wald",
		},
		{
			"remove_tag lässt andere stehen",
			`remove_tag("<i>Wald</i>", tag: "<sym")`, "<i>Wald</i>",
		},
		{
			"remove_tag mit vollem Namen",
			`remove_tag("<sym-auto>R</sym-auto>", tag:"<sym-auto>")`, "R",
		},

		// `regex_escape` — ein Kartenname wird zu einem Muster, das ihn trifft.
		{"regex_escape schützt Sonderzeichen", `regex_escape("A+B")`, `A\+B`},

		// `replace_text` ist `replace` unter einem zweiten Namen.
		{
			"replace_text ersetzt",
			`replace_text("Legendary Creature", match:"(?i)Legendary ?", replace:"")`, "Creature",
		},

		// **Eine Zahl ist eine Bedingung** — belegt in zwei Exportvorlagen.
		{"eine Zahl ungleich null ist wahr", `if 3 then "ja" else "nein"`, "ja"},
		{"null ist falsch", `if 0 then "ja" else "nein"`, "nein"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Eine Zeichenkette bleibt als Bedingung ein Fehler.**
//
// Die Gegenprobe zur Lockerung: für sie ist keine Fundstelle da, und die
// Lesarten gehen auseinander — „nicht leer" oder „nicht `no`"?
func TestEineZeichenketteBleibtKeineBedingung(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(), `if "ja" then 1 else 2`,
		msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler — eine Zeichenkette galt als Bedingung")
	}
}

// **Jede weist ihr fehlendes benanntes Argument ab.**
func TestDieNachgezogenenBrauchenIhreArgumente(t *testing.T) {
	for _, quelle := range []string{
		`break_text("a")`, `position(of:"a")`, `position(in:"a")`,
		`remove_tag("a")`, `remove_tag("a", tag: "")`,
	} {
		t.Run(quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())

			// `tag: ""` ist kein Fehler, sondern gibt den Text unverändert —
			// eine Auszeichnung ohne Namen gibt es nicht zu entfernen.
			if quelle == `remove_tag("a", tag: "")` {
				if err != nil || wert.AlsText() != "a" {
					t.Errorf("%q / %v", wert.AlsText(), err)
				}

				return
			}

			if err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein unbrauchbares Muster wird auch hier benannt.**
func TestBreakTextWeistEinUnbrauchbaresMusterAb(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(), `break_text("a", match:"[")`,
		msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein zweites bloßes Argument ist kein Fehler — es wird verworfen.**
//
// Hier stand bis
// #438 das
// Gegenteil, und es war eine Annahme: bei MSE bekommt **jedes** unbenannte
// Argument den Namen `input` (`src/script/parser.cpp`, Zeilen 890–894, Fassung
// f9d9356 vom 2021-07-06), und die Zuweisungsschleife lässt das **erste**
// stehen (`src/script/context.cpp`, Zeilen 152–155).
func TestEinZweitesBlossesArgumentWirdVerworfen(t *testing.T) {
	for _, fall := range []struct{ quelle, erwartet string }{
		{`regex_escape("a", "b")`, "a"},
		{`to_number("1", "2")`, "1"},
		{`to_int("1", "2")`, "1"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Die Zweige, die im Betrieb selten sind, aber nicht unmöglich.**
func TestDieNachgezogenenMeldenIhreGrenzfaelle(t *testing.T) {
	for _, fall := range []struct{ name, quelle string }{
		// **Ein Muster, das in jeder Sprache falsch ist** — anders als ein
		// Vorausblick, den die Rückfalllinie ausführt.
		{"split_text mit kaputtem Muster", `split_text("a", match:"[")`},

		// **`to_int` und `to_number` auf etwas Unzahligem melden erst beim
		// Benutzen** — MSE gibt dort einen verzögerten Fehlerwert
		// (`src/script/functions/basic.cpp`, Zeile 123, Fassung f9d9356 vom
		// 2021-07-06). Der `+ 1` ist die Stelle, an der er wirft.
		{"to_number auf Unbrauchbarem", `to_number("keine Zahl") + 1`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **`chosen` und `position` auf einer Zahl** — beide kommen ohne Einträge aus.
func TestChosenUndPositionAufEinfachenWerten(t *testing.T) {
	for _, fall := range []struct{ quelle, erwartet string }{
		{`chosen(3, choice:"3")`, "true"},
		{`position(of:"", in:"abc")`, "-1"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **`NeueEingebaute` macht aus einer Go-Funktion einen Funktionswert.**
//
// Sie ist die Naht für Funktionen, die dieses Paket nicht haben darf — der
// Grund steht bei ihr. Hier wird sie ohne eine solche geprüft.
func TestNeueEingebauteIstEineFunktion(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("gruss", msescript.NeueEingebaute(
		func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
			return msescript.Text("hallo " + ruf.Stellung[0].AlsText()), nil
		}))

	wert, err := msescript.AuswertenQuelle(t.Context(), `gruss("Welt")`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "hallo Welt" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Eine Abbildung und ein Text ergeben Text** (#216).
//
// `magic.mse-game` Zeile 8411 verbindet eine Zeichenkette mit einer Liste.
// Die Entscheidung ist nicht neu, sondern zu Ende geführt: `AlsText` einer
// Abbildung ist längst die Folge ihrer Werte.
func TestEineAbbildungUndEinTextErgebenText(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		{"Text links", `"vor " + split_text(match:",", "a,b")`, "vor ab"},
		{"Text rechts", `split_text(match:",", "a,b") + " nach"`, "ab nach"},

		// **Zwei Listen verketten sich** (#390):
		//
		//	} else if (at == SCRIPT_COLLECTION && bt == SCRIPT_COLLECTION) {
		//	    a = make_intrusive<ScriptConcatCollection>(a, b);
		//
		// `src/script/context.cpp`, Zeile 465 ff., Fassung f9d9356 vom
		// 2021-07-06; die Zahlenschlüssel werden dabei umnummeriert. Hier stand
		// **2** — das war die Vereinigung nach Namen, und die gilt nur für
		// benannte Schlüssel wie bei `map + [stylesheet: "…"]`.
		{"zwei Listen", `length(split_text(match:",", "a,b") + split_text(match:",", "c"))`, "3"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Eine Abbildung mit einer Zahl wird aneinandergesetzt.**
//
// Hier stand bis zum 2026-08-17 „bleibt ein Fehler", mit der Begründung „für
// sie ist keine Fundstelle da". Sie war da:
//
//	else { a = to_script(a->toString() + b->toString()); }
//
// `src/script/context.cpp`, Zeile 512, Fassung f9d9356 vom 2021-07-06 — MSEs
// letzter Zweig fängt **jede** Kombination. 47 Befunde über die 190 Stile
// (#270).
func TestEineAbbildungUndEineZahlWerdenAneinandergesetzt(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`split_text(match:",", "a,b") + 3`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "ab3" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "ab3")
	}
}

// `count_chosen` zählt, wie viele der **genannten** Wahlen ein Feld trägt (#240).
func TestCountChosenZaehltDieGenanntenWahlen(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		{
			"zwei von fünf Farben",
			`count_chosen("white, blue", choices: "white,blue,black,red,green")`, "2",
		},

		// **Kein stumpfes Zählen.** `card_color` trägt neben den Farben auch
		// `artifact`; wer die Einträge zählte, machte aus einem farblosen
		// Artefakt eine einfarbige Karte.
		{
			"artifact zählt nicht als Farbe",
			`count_chosen("artifact", choices: "white,blue,black,red,green")`, "0",
		},

		{"nichts gewählt", `count_chosen("", choices: "white,blue")`, "0"},
		{
			"ohne Rücksicht auf Groß und Klein",
			`count_chosen("White", choices: "white")`, "1",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **`count_chosen` braucht seine Liste und genau eine Stellung.**
func TestCountChosenBrauchtSeineArgumente(t *testing.T) {
	for _, quelle := range []string{`count_chosen("a")`} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(t.Context(), quelle,
				msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
