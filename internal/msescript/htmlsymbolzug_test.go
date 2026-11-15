package msescript_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Der Symbolzug in `to_html` und die Randfälle drumherum** —
// #723.
//
// `bibliothek_html_test.go` daneben führt die Tabelle des Orakels und den
// Auszeichnungsstapel. Hier steht, was den **Rückruf** betrifft: woher die
// Symbolschrift kommt, was er zu sehen bekommt, und was ohne ihn geschieht.

// **Die Symbolschrift kommt auch aus der Umgebung.**
//
//	symbol_font      := "magic-mana-small"
//	symbol_font_size := 12
//
// `magic-planesculptors.mse-export-template/export-template`, Zeilen 27–28:
// die Vorlage setzt sie **als Variable** und ruft `to_html(card.name)` ohne
// benanntes Argument. MSEs `SCRIPT_OPTIONAL_PARAM_N` liest aus dem Kontext,
// nicht aus der Argumentliste (`src/script/functions/util.hpp`, Zeilen
// 98–101, Fassung f9d9356 vom 2021-07-06) — dieselbe Regel wie bei den
// Bildfunktionen.
func TestDieSymbolschriftKommtAuchAusDerUmgebung(t *testing.T) {
	t.Parallel()

	gerufen := ""

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("symbol_font", msescript.Text("magic-mana-small"))
	umgebung.Setze("to_html", msescript.NeueEingebaute(msescript.HTMLFunktion(
		func(_ context.Context, zug, schrift string, groesse float64) (string, error) {
			gerufen = schrift + "|" + zug + "|" + msescript.Zahl(groesse).AlsText()

			return "<img>", nil
		})))

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`to_html("Kosten: <sym>2WW</sym>!")`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != `Kosten: <span class="symbol"><img></span>!` {
		t.Errorf("%q", wert.AlsText())
	}

	if gerufen != "magic-mana-small|2WW|12" {
		t.Errorf("gerufen: %q", gerufen)
	}
}

// **In den Symbolzug geht das Zeichen roh.**
//
//	if (symbol.opened > 0 && symbol_font) {
//	  symbols += c; // write as symbols instead
//	} else {
//	  c = untag_char(c);
//
// `export.cpp`, Zeilen 244–247, Fassung f9d9356 vom 2021-07-06: `untag_char`
// steht im **anderen** Zweig. Ein geschütztes Kleinerzeichen im Symbolblock
// erreicht die Symbolschrift deshalb, wie es dasteht.
//
// **Dieser Test steht hier, weil die Gegenprobe grün war.** Die erste Fassung
// entschützte in beiden Zweigen; kein Fall der Tabelle nebenan betrat den
// Unterschied, und das Lesen des Originals hatte ihn übersehen (§10).
func TestEinSymbolzugBekommtDasRoheZeichen(t *testing.T) {
	t.Parallel()

	gesehen := ""

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("to_html", msescript.NeueEingebaute(msescript.HTMLFunktion(
		func(_ context.Context, zug, _ string, _ float64) (string, error) {
			gesehen = zug

			return "", nil
		})))

	_, err := msescript.AuswertenQuelle(t.Context(),
		"to_html(\"<sym>a\x01b\ueb00c</sym>\", symbol_font: \"magic-mana-small\")",
		umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if gesehen != "a\x01b\ueb00c" {
		t.Errorf("der Symbolzug kam als %q", gesehen)
	}
}

// **Ohne Wandler ist eine gesetzte Symbolschrift ein Fehler.**
//
//	guard_export_info(_("symbols_to_html"));
//
// `export.cpp`, Zeile 184: außerhalb einer Ausfuhr gibt es keinen Ort für die
// Bilder, und MSE bricht ab. **Nur wenn wirklich ein Symbolzug anfällt** — ein
// Text ohne `<sym>` läuft durch, gleich was `symbol_font` sagt.
func TestOhneWandlerIstEineSymbolschriftEinFehler(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("symbol_font", msescript.Text("magic-mana-small"))

	_, err := msescript.AuswertenQuelle(t.Context(),
		`to_html("Kosten: <sym>2WW</sym>")`, umgebung)
	if err == nil {
		t.Fatal("der Aufruf ging durch")
	}

	if !strings.Contains(err.Error(), "symbols_to_html") {
		t.Errorf("Meldung: %v", err)
	}

	// **Ohne Symbolzug bleibt es still.**
	wert, err := msescript.AuswertenQuelle(t.Context(), `to_html("<b>x</b>")`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "<b>x</b>" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Die Wege, die nur ein Randfall betritt.**
//
// Sie stehen zusammen, weil sie dasselbe belegen: dass jede Verzweigung in
// `to_html` erreichbar ist. Ohne sie hielte die Abdeckung von 100 % in diesem
// Paket eine Zusage, die niemand geprüft hat (§3).
func TestDieRandfaelleVonToHTML(t *testing.T) {
	t.Parallel()

	t.Run("ohne Eingabe wird der Name genannt", func(t *testing.T) {
		t.Parallel()

		_, err := msescript.AuswertenQuelle(t.Context(), `to_html()`, msescript.Bibliothek())
		if err == nil || !strings.Contains(err.Error(), "to_html") {
			t.Fatalf("Fehler = %v", err)
		}
	})

	// **Ein Symbolzug ohne schließende Marke wird trotzdem gewandelt.**
	//
	//	// end of input
	//	if (!symbols.empty()) { … ret += symbols_to_html(symbols, …); }
	//
	// `export.cpp`, Zeilen 262–267, Fassung f9d9356 vom 2021-07-06.
	t.Run("ein offener Symbolzug wird am Ende gewandelt", func(t *testing.T) {
		t.Parallel()

		gesehen := ""
		umgebung := msescript.Bibliothek().Darueber()
		umgebung.Setze("to_html", msescript.NeueEingebaute(msescript.HTMLFunktion(
			func(_ context.Context, zug, _ string, _ float64) (string, error) {
				gesehen = zug

				return "<img>", nil
			})))

		wert, err := msescript.AuswertenQuelle(t.Context(),
			`to_html("<sym>2WW", symbol_font: "magic-mana-small")`, umgebung)
		if err != nil {
			t.Fatalf("auswerten: %v", err)
		}

		if gesehen != "2WW" || wert.AlsText() != `<span class="symbol"><img></span>` {
			t.Errorf("gesehen %q, Ausgabe %q", gesehen, wert.AlsText())
		}
	})

	// **Ein Fehler des Wandlers kommt durch** — und behält seine eigene
	// Meldung, statt von der Wache überschrieben zu werden.
	t.Run("ein Fehler des Wandlers kommt durch", func(t *testing.T) {
		t.Parallel()

		umgebung := msescript.Bibliothek().Darueber()
		umgebung.Setze("to_html", msescript.NeueEingebaute(msescript.HTMLFunktion(
			func(_ context.Context, _, _ string, _ float64) (string, error) {
				return "", errors.New("die Platte ist voll")
			})))

		_, err := msescript.AuswertenQuelle(t.Context(),
			`to_html("<sym>W</sym>", symbol_font: "magic-mana-small")`, umgebung)
		if err == nil || !strings.Contains(err.Error(), "die Platte ist voll") {
			t.Fatalf("Fehler = %v", err)
		}

		// **Und ebenso am Ende der Eingabe.** Das ist ein zweiter Aufruf im
		// Quelltext, und ein Fehlerweg, den kein anderer Fall betritt.
		_, err = msescript.AuswertenQuelle(t.Context(),
			`to_html("<sym>W", symbol_font: "magic-mana-small")`, umgebung)
		if err == nil || !strings.Contains(err.Error(), "die Platte ist voll") {
			t.Fatalf("am Ende: Fehler = %v", err)
		}
	})

	// **Zweimal dieselbe offene Marke bleibt zweimal vorgemerkt.** Sie hebt
	// sich nicht auf — nur eine Gegenrichtung tut das.
	t.Run("zwei gleiche Öffnungen heben sich nicht auf", func(t *testing.T) {
		t.Parallel()

		wert, err := msescript.AuswertenQuelle(t.Context(),
			`to_html("<b><b>x</b></b>")`, msescript.Bibliothek())
		if err != nil {
			t.Fatalf("auswerten: %v", err)
		}

		if wert.AlsText() != "<b>x</b>" {
			t.Errorf("%q", wert.AlsText())
		}
	})

	// **Eine Größe, die keine Zahl ist, ist dasselbe wie keine.**
	//
	//	if (symbol_font_size <= 0) symbol_font_size = 12; // a default
	//
	// `export.cpp`, Zeile 282: MSE fängt die fehlende Variable ab und
	// vergleicht danach gegen null. Ein unlesbarer Wert landet auf demselben
	// Weg bei der Vorgabe — und nicht in einem Abbruch.
	t.Run("eine unlesbare Größe fällt auf die Vorgabe zurück", func(t *testing.T) {
		t.Parallel()

		gesehen := 0.0
		umgebung := msescript.Bibliothek().Darueber()
		umgebung.Setze("to_html", msescript.NeueEingebaute(msescript.HTMLFunktion(
			func(_ context.Context, _, _ string, groesse float64) (string, error) {
				gesehen = groesse

				return "", nil
			})))

		_, err := msescript.AuswertenQuelle(t.Context(),
			`to_html("<sym>W</sym>", symbol_font: "m", symbol_font_size: "keine Zahl")`,
			umgebung)
		if err != nil {
			t.Fatalf("auswerten: %v", err)
		}

		if gesehen != msescript.HTMLVorgabegroesse {
			t.Errorf("Größe %v, erwartet %v", gesehen, msescript.HTMLVorgabegroesse)
		}
	})

	// **Ein weicher Trenner am Ende der Eingabe.**
	t.Run("ein weicher Trenner ganz am Ende", func(t *testing.T) {
		t.Parallel()

		wert, err := msescript.AuswertenQuelle(t.Context(),
			`to_html("a<sep-soft>X</sep-soft>")`, msescript.Bibliothek())
		if err != nil {
			t.Fatalf("auswerten: %v", err)
		}

		if wert.AlsText() != "a" {
			t.Errorf("%q", wert.AlsText())
		}
	})

	// **Und zwei nacheinander, mit Text dahinter.**
	t.Run("zwei weiche Trenner nacheinander", func(t *testing.T) {
		t.Parallel()

		wert, err := msescript.AuswertenQuelle(t.Context(),
			`to_html("a<sep-soft>X</sep-soft>b<sep-soft>Y</sep-soft>c")`,
			msescript.Bibliothek())
		if err != nil {
			t.Fatalf("auswerten: %v", err)
		}

		if wert.AlsText() != "abc" {
			t.Errorf("%q", wert.AlsText())
		}
	})
}

// **Die beiden ausgeführten Helfer eines Rufs.**
//
// `internal/service` setzt eigene eingebaute Funktionen in die Umgebung —
// `symbols_to_html` liest dieselben zwei Angaben wie `to_html` und meldet ein
// fehlendes Argument mit demselben Satz (#658). Beide Wege stehen deshalb
// exportiert da; geprüft werden sie hier, wo sie herkommen.
func TestDieAusgefuehrtenHelferEinesRufs(t *testing.T) {
	t.Parallel()

	t.Run("EineEingabe nennt den Namen", func(t *testing.T) {
		t.Parallel()

		err := msescript.Ruf{Zeile: 3}.EineEingabe("meine_funktion")
		if err == nil || !strings.Contains(err.Error(), "meine_funktion") {
			t.Fatalf("Fehler = %v", err)
		}
	})

	t.Run("AusRufOderUmgebung liest beide Quellen", func(t *testing.T) {
		t.Parallel()

		umgebung := msescript.NeueUmgebung(map[string]msescript.Wert{
			"schrift": msescript.Text("aus der Umgebung"),
		})

		ruf := msescript.Ruf{
			Benannt:  []msescript.Eintrag{{Name: "eigen", Wert: msescript.Text("benannt")}},
			Umgebung: umgebung,
		}

		if wert := ruf.AusRufOderUmgebung("eigen"); wert.AlsText() != "benannt" {
			t.Errorf("benannt: %q", wert.AlsText())
		}

		if wert := ruf.AusRufOderUmgebung("schrift"); wert.AlsText() != "aus der Umgebung" {
			t.Errorf("aus der Umgebung: %q", wert.AlsText())
		}

		if wert := ruf.AusRufOderUmgebung("gibtsnicht"); wert.Art != msescript.ArtNichts {
			t.Errorf("unbekannt: %v", wert.Art)
		}
	})
}
