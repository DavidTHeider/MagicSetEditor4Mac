package msescript_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func rechne(t *testing.T, quelle string, namen map[string]msescript.Wert) msescript.Wert {
	t.Helper()

	wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.NeueUmgebung(namen))
	if err != nil {
		t.Fatalf("%q abgewiesen: %v", quelle, err)
	}

	return wert
}

func TestAuswertenRechnetDieFormenAusDenPaketen(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   string
		erwartet string
	}{
		{name: "eine Zahl", quelle: `45.5`, erwartet: "45.5"},
		{name: "eine Zeichenkette", quelle: `"no"`, erwartet: `"no"`},
		{name: "nichts", quelle: `nil`, erwartet: "nil"},
		{name: "true ist ein Name, kein Schlüsselwort", quelle: `true`, erwartet: "true"},
		{
			// `length(…)*17 + 2` aus yugioh Zeile 732, ohne die Bibliothek.
			name: "eine Rechnung nach Vorrang", quelle: `3*17 + 2`, erwartet: "53",
		},
		{
			// `"attribute/earth_" + get_lang_ex() + ".png"` aus yugioh Zeile 275.
			name:     "Zeichenketten verketten",
			quelle:   `"attribute/earth_" + "en" + ".png"`,
			erwartet: `"attribute/earth_en.png"`,
		},
		{
			// **Eine Zahl in einem Pfad bekommt keine Nachkommastellen.**
			name:     "eine Zahl in einer Zeichenkette",
			quelle:   `"breite_" + 17`,
			erwartet: `"breite_17"`,
		},
		{
			name:     "eine Zuweisung ergibt ihren Wert und legt den Namen an",
			quelle:   "a := 2\na + 1",
			erwartet: "3",
		},
		{
			// `if styling.overlay == "" then nil else …` aus magic-m15 Zeile 363.
			name:     "eine Bedingung nimmt nur einen Zweig",
			quelle:   `if "" == "" then nil else fehlt_hier`,
			erwartet: "nil",
		},
		{
			// `card.rarity=="a" or card.rarity=="b"` aus yugioh Zeile 116.
			name:     "eine oder-Kette",
			quelle:   `"b"=="a" or "b"=="b" or "b"=="c"`,
			erwartet: "true",
		},
		{
			name:     "und rechnet die rechte Seite nicht, wenn links schon falsch ist",
			quelle:   `false and fehlt_hier`,
			erwartet: "false",
		},
		{
			name:     "oder rechnet die rechte Seite nicht, wenn links schon wahr ist",
			quelle:   `true or fehlt_hier`,
			erwartet: "true",
		},
		{
			name:     "nicht kehrt um",
			quelle:   `not false`,
			erwartet: "true",
		},
		{
			// `card_style.text.content_lines <= 2` aus magic-m15 Zeile 291.
			name: "ein Größenvergleich", quelle: `2 <= 2`, erwartet: "true",
		},
		{
			// **Eine Zahl und eine Zeichenkette sind nicht gleich.**
			name:   "Arten vergleichen sich nicht über die Darstellung",
			quelle: `1 == "1"`, erwartet: "false",
		},
		{
			// `template_prefix[type]` aus magic-m15 Zeile 54.
			name:     "eine Abbildung, indiziert",
			quelle:   `[card: "card.jpg", pt: "pt.png"]["pt"]`,
			erwartet: `"pt.png"`,
		},
		{
			// **Ein fehlender Schlüssel ergibt einen Fehlerwert**, und die
			// Vorlage fängt ihn mit `or else` — so steht es in `game_script`
			// Zeile 62 (`map["stylesheet"] or else "m15-altered"`).
			name:     "ein fehlender Schlüssel wird aufgefangen",
			quelle:   `[card: "c"]["gibtsnicht"] or else "nichts da"`,
			erwartet: `"nichts da"`,
		},
		{
			// `map := map + [stylesheet: "…"]` aus game_script — und es
			// **überschreibt nicht**, es hängt an: `getMember` fragt die linke
			// Seite zuerst (#271).
			name:     "zwei Abbildungen verketten",
			quelle:   `[a: 1, b: 2] + [b: 9, c: 3]`,
			erwartet: `[a: 1, b: 2, b: 9, c: 3]`,
		},
		{
			name:     "eine Folge ergibt den letzten Wert",
			quelle:   "a := 1\nb := 2\na + b",
			erwartet: "3",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if wert := rechne(t, fall.quelle, nil); wert.String() != fall.erwartet {
				t.Errorf("%s, erwartet war %s", wert.String(), fall.erwartet)
			}
		})
	}
}

// Der Kern der Sprache: eine geschweifte Klammer bekommt ihre Umgebung gereicht.
//
// **Das ist die Frage, für die #23 da war**, und der Beleg steht in
// `magic-m15.mse-style/style` Zeile 71:
//
//	mana_sort := {
//	    if styling.unsorted_casting_cost then input
//	    else sort_text@(order: "…")()
//	}
//
// `input` steht in keiner Zuweisung und in keiner Argumentliste. Es kann nur von
// dort kommen, wo die Klammer **benutzt** wird — MSE setzt es je Feld. Ein
// Auswerter, der die Umgebung der Definition einschlösse, fände es nie.
func TestEineKlammerSiehtDieUmgebungDesAufrufs(t *testing.T) {
	// **Der erste Anlauf dieses Tests belegte nichts.** Er setzte `input` auf
	// der obersten Ebene und rief den Block von dort — und die oberste Ebene
	// ist zugleich die, in der er geschrieben steht. Beide Lesarten ergaben
	// dasselbe; der Gegenlauf mit eingeschlossener Umgebung blieb grün.
	//
	// Unterscheidbar wird es erst, wenn der Block an einer Stelle **steht**, an
	// der `input` fehlt, und an einer anderen **gerufen** wird, an der es da
	// ist. Unter „Umgebung des Aufrufs" ergibt das "x"; unter „Umgebung der
	// Definition" ist `input` unbekannt und die Auswertung scheitert.
	quelle := "f := {input}\n" +
		"g := {input := \"x\"\n f()}\n" +
		"g()"

	if wert := rechne(t, quelle, nil); wert.String() != `"x"` {
		t.Errorf("%s, erwartet war \"x\"", wert.String())
	}
}

// Und eine Zuweisung im Block sickert nicht nach außen.
//
// Zwei Felder, die dieselbe Vorlage benutzen, dürfen einander nicht sehen. Ein
// solcher Einfluss wäre nicht zu finden: er zeigte sich nur in einer bestimmten
// Reihenfolge.
func TestEineZuweisungImBlockBleibtDrinnen(t *testing.T) {
	quelle := "a := 1\nf := {a := 99\n a}\nf()\na"

	if wert := rechne(t, quelle, nil); wert.String() != "1" {
		t.Errorf("%s, erwartet war 1 — die Zuweisung im Block ist ausgetreten", wert.String())
	}
}

// `@` belegt vor, es nagelt nicht fest.
func TestGebundeneArgumenteWerdenVomAufrufGeschlagen(t *testing.T) {
	quelle := "f := {ordnung}\ng := f@(ordnung: \"vorab\")\ng() + \" \" + g(ordnung: \"beim Aufruf\")"

	if wert := rechne(t, quelle, nil); wert.String() != `"vorab beim Aufruf"` {
		t.Errorf("%s", wert.String())
	}
}

// Ein Kartenkontext kommt als Abbildung herein — so wird `card.rarity` gelesen.
func TestDerKartenkontextKommtAlsAbbildung(t *testing.T) {
	umgebung := map[string]msescript.Wert{
		"card": {Art: msescript.ArtAbbildung, Eintraege: []msescript.Eintrag{
			{Name: "rarity", Wert: msescript.Text("star rare")},
		}},
	}

	quelle := `card.rarity=="rare" or card.rarity=="star rare"`
	if wert := rechne(t, quelle, umgebung); wert.String() != "true" {
		t.Errorf("%s, erwartet war true", wert.String())
	}
}

// **Ein unbekanntes Feld ergibt einen Fehlerwert, nicht nichts.**
//
//	ScriptValueP ScriptValue::getMember(const String& name) const {
//	  return delay_error(ScriptErrorNoMember(typeName(), name));
//	}
//
// `src/script/value.cpp`, Zeilen 67–74, Fassung f9d9356 vom 2021-07-06. Hier
// stand bis zum 2026-08-17 „ergibt nichts" — und damit fing `or else` den
// falschen Fall (siehe [msescript.Fehlerwert]).
func TestEinUnbekanntesFeldErgibtEinenFehlerwert(t *testing.T) {
	umgebung := map[string]msescript.Wert{"card": {Art: msescript.ArtAbbildung}}

	if wert := rechne(t, `card.gibtsnicht`, umgebung); !wert.IstFehler() {
		t.Errorf("%s, erwartet war ein Fehlerwert", wert.String())
	}
}

// Eine Vorlage darf die Anwendung nicht beenden.
//
// `a := {a()}` gefolgt von `a()` ist in zwei Zeilen geschrieben, und fremde
// Daten sind der Normalfall (AGENTS.md §8). Ohne Grenze frisst das den Stapel,
// und Go beendet den Prozess — nicht als Fehler, den jemand anzeigen könnte.
func TestEineEndloseRekursionWirdAbgewiesenStattAbzustuerzen(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(), "a := {a()}\na()", nil)

	var befund *msescript.Fehler
	if !errors.As(err, &befund) {
		t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
	}
}

// Eine laufende Auswertung lässt sich abbrechen (AGENTS.md §7).
//
// Eine Kartenvorschau wird abgebrochen, nicht abgewartet: wer während des
// Tippens vier Läufe anstößt, zeigt sonst den drittletzten Stand.
func TestEineAuswertungLaesstSichAbbrechen(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	_, err := msescript.AuswertenQuelle(ctx, `1 + 1`, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Fehler %v, erwartet war der Abbruch", err)
	}
}

// **Und mitten im Lauf**, durch eine Kette von Kontexten mit Werten — wie
// ihn Dienst und Ausfuhr hereinreichen. E04 (#2241): der Auswerter fragt
// den Abbruch seitdem über den `Done`-Kanal statt bei jedem Schritt über
// `Err()` die ganze Kette hinauf.
func TestEineLaufendeAuswertungLaesstSichAbbrechen(t *testing.T) {
	type schluessel int

	ctx, abbrechen := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer abbrechen()

	for i := range 8 {
		ctx = context.WithValue(ctx, schluessel(i), i)
	}

	beginn := time.Now()

	_, err := msescript.AuswertenQuelle(ctx, `for x from 1 to 100000000 do x`, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Fehler %v, erwartet war der Ablauf der Frist", err)
	}

	if dauer := time.Since(beginn); dauer > 2*time.Second {
		t.Errorf("der Abbruch kam nach %v", dauer)
	}
}

// Ohne Umgebung geht es auch.
func TestAuswertenOhneUmgebung(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(), `1 + 1`, nil)
	if err != nil || wert.String() != "2" {
		t.Errorf("%v / %v", wert, err)
	}
}

// Ein Zerteilfehler kommt unverändert durch.
func TestAuswertenQuelleReichtDenZerteilfehlerDurch(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(), `a :=`, nil)

	var befund *msescript.Fehler
	if !errors.As(err, &befund) || befund.Spalte != 5 {
		t.Errorf("Fehler %v, erwartet war die Stelle 1:5", err)
	}
}

// Die Formen, die keine Paketzeile vorgibt — aber jede Sprache hat.
func TestAuswertenRechnetDieUebrigenFormen(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"ein Minuszeichen", `-5`, "-5"},
		{"abziehen", `7 - 2`, "5"},
		{"teilen", `7 / 2`, "3.5"},
		{"ungleich", `1 != 2`, "true"},
		{"kleiner", `1 < 2`, "true"},
		{"größer", `2 > 1`, "true"},
		{"größergleich", `2 >= 2`, "true"},
		{"Zeichenketten der Größe nach", `"a" < "b"`, "true"},
		{"nichts ist gleich nichts", `nil == nil`, "true"},
		{"zwei gleiche Abbildungen", `[a: 1] == [a: 1]`, "true"},
		{"zwei ungleiche Abbildungen", `[a: 1] == [a: 2]`, "false"},
		{"verschieden lange Abbildungen", `[a: 1] == [a: 1, b: 2]`, "false"},
		{"nichts ist falsch", `if nil then 1 else 2`, "2"},
		{"eine Funktion ist nicht gleich", `{1} == {1}`, "false"},
		{"und mit zwei wahren", `true and true`, "true"},
		{"zwei gleiche Wahrheitswerte", `true == true`, "true"},
		{"ein Stellungsargument wird übergangen", `{1}(5)`, "1"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if wert := rechne(t, fall.quelle, nil); wert.String() != fall.erwartet {
				t.Errorf("%s, erwartet war %s", wert.String(), fall.erwartet)
			}
		})
	}
}

// Eine Form, die der Auswerter nicht kennt, wird gemeldet statt übergangen.
//
// Der Zerteiler erzeugt sie nicht — dieser Test baut sie von Hand. Der Zweck ist
// die Zukunft: kommt eine Form dazu und vergisst jemand den Zweig, ergibt der
// Ausdruck sonst still `nil`, und die Karte hat ein leeres Feld.
func TestEineUnbekannteFormWirdGemeldet(t *testing.T) {
	knoten := &msescript.Knoten{Form: "Unfug", Zeile: 3, Spalte: 7}

	_, err := msescript.Auswerten(t.Context(), knoten, nil)

	var befund *msescript.Fehler
	if !errors.As(err, &befund) || befund.Zeile != 3 {
		t.Errorf("Fehler %v, erwartet war die Stelle 3:7", err)
	}
}

// Die Darstellung eines Werts zeigt seine Art.
func TestDieDarstellungEinesWerts(t *testing.T) {
	if msescript.Wahrheit(false).AlsText() != "false" {
		t.Error("false")
	}
	if msescript.Nichts.AlsText() != "" {
		t.Error("nil als Text")
	}
	if (msescript.Wert{Art: msescript.ArtFunktion}).AlsText() != "Funktion" {
		t.Error("Funktion als Text")
	}
	if (msescript.Wert{Art: msescript.ArtFunktion}).String() != "Funktion" {
		t.Error("Funktion")
	}
	// **Eine Abbildung wird aneinandergesetzt, nicht benannt** (#51).
	//
	// Hier stand „Abbildung"; das war die Gestalt, nicht der Inhalt. Die
	// Exportvorlagen brauchen den Inhalt: `to_text(for each card in cards do
	// zeile())` ergibt den Text aller Zeilen hintereinander. Eine leere
	// Abbildung ergibt folgerichtig einen leeren Text.
	if (msescript.Wert{Art: msescript.ArtAbbildung}).AlsText() != "" {
		t.Error("leere Abbildung als Text")
	}

	gefuellt := msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: []msescript.Eintrag{
		{Name: "0", Wert: msescript.Text("a")}, {Name: "1", Wert: msescript.Zahl(2)},
	}}
	if gefuellt.AlsText() != "a2" {
		t.Errorf("gefüllte Abbildung als Text: %q", gefuellt.AlsText())
	}
}

// Ein Zahlknoten, dessen Text keine Zahl ist.
//
// Der Lexer erzeugt ihn nicht — dieser Test baut ihn von Hand. Er steht hier
// aus demselben Grund wie die unbekannte Form: der Zweig ist gebaut, also
// gehört gezeigt, was er tut.
func TestEinZahlknotenOhneZahl(t *testing.T) {
	knoten := &msescript.Knoten{Form: msescript.FormZahl, Text: "keine", Zeile: 2, Spalte: 4}

	_, err := msescript.Auswerten(t.Context(), knoten, nil)

	var befund *msescript.Fehler
	if !errors.As(err, &befund) || befund.Spalte != 4 {
		t.Errorf("Fehler %v, erwartet war die Stelle 2:4", err)
	}
}
