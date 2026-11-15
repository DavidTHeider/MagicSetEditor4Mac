package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// `cycle(…)` und seine Kurzform `(…)` — die Ordnung, die
// #98 offen
// gelassen hat.
//
// # Sie ist jetzt nachgelesen statt geraten
//
// `src/util/spec_sort.hpp` von twanvl/MagicSetEditor2, Stand 2026-08-16,
// beschreibt sie im Kopfkommentar, und `cycle_sort` in `spec_sort.cpp:135`
// setzt sie um. Die genannten Zeichen bilden einen **Kreis im Uhrzeigersinn**;
// ausgegeben wird die Drehung, die
//
//  1. den kürzesten Weg über den Kreis nimmt,
//  2. ihre **Lücken früh** hat — eine Lücke ist ein Zeichen der Ordnung, das in
//     der Eingabe fehlt,
//  3. bei Gleichstand am weitesten vorn im Ausdruck beginnt.
//
// Das Beispiel im Quelltext ist das von Magic:
// `sort_text(order: "XYZ<0123456789>cycle(WUBRG)")`.
//
// # Warum das der Rahmen war
//
// `magic-m15.mse-style` kommt über `card color → image` und drei Ecken auf
// `sort_text` mit `"(wubrg)"`. Ohne diese Ordnung meldete die Angabe einen
// Befund, und das Feld, das den **Rahmen** zeichnet, fiel aus
// (#240).
//
// Fassung f9d9356 vom 2021-07-06.
func TestEinZyklusOrdnetNachDemKuerzestenWeg(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt string
	}{
		// **Die drei Beispiele aus dem Quelltext von MSE**, wörtlich.
		{
			"eine Lücke ist billiger am Ende",
			`sort_text("W1G", order: "XYZ<0123456789>cycle(WUBRG)")`, "1GW",
		},
		{
			"alle fünf Farben",
			`sort_text("GRBUWWUG", order: "cycle(WUBRG)")`, "WWUUBRGG",
		},
		{
			"die frühere Lücke gewinnt",
			`sort_text("WUR", order: "cycle(WUBRG)")`, "RWU",
		},

		// **Die Kurzform ist dasselbe** — `spec_sort.cpp:294` behandelt `(`
		// und `cycle(` im selben Zweig.
		{
			"die runde Klammer allein",
			`sort_text("WUR", order: "(WUBRG)")`, "RWU",
		},

		// So steht es in `magic.mse-game` — klein geschrieben.
		//
		// **Nachgerechnet, nicht geraten.** Der erste Anlauf erwartete hier
		// `"ru"`, weil `r` in der Eingabe vorn steht — das ist aber die Regel
		// von `mixed(…)`, nicht die des Zyklus. Über `w,u,b,r,g` mit `u` und
		// `r` belegt ergeben die Startpunkte `w`, `u` und `g` je den Wert 5
		// und die übrigen 9; bei Gleichstand gewinnt der früheste, also `w`.
		// Damit steht `u` vor `r`.
		{"klein geschrieben", `sort_text("ru", order: "(wubrg)")`, "ur"},

		// Ein einzelnes Zeichen dreht nicht.
		{"ein einziges Zeichen", `sort_text("R", order: "cycle(WUBRG)")`, "R"},

		// Was die Ordnung nicht kennt, bleibt hinten — wie bei jeder anderen
		// Ordnung dieses Programms auch.
		//
		// **Und `R` steht vor `W`**, aus demselben Grund wie beim Beispiel
		// `"WUR" → "RWU"` aus dem Quelltext: mit nur `W` und `R` belegt
		// gewinnen die Startpunkte `U`, `B` und `R` mit Wert 5, der früheste
		// ist `U` — und von `U` aus kommt `R` vor `W`.
		// **Unbekanntes fällt weg — nachgelesen, nicht erfunden.**
		//
		// Hier stand bis zum 2026-08-18 „Unbekanntes bleibt hinten" mit dem
		// Ergebnis `"RWx"`. Das war eine Annahme dieses Projekts; MSEs
		// `spec_sort` hängt an sein Ergebnis **nur** an, was die
		// Spezifikation nennt, und wirft den Rest der Eingabe weg
		// (`src/util/spec_sort.cpp`, Zeilen 266–352 und 355, Fassung f9d9356
		// vom 2021-07-06).
		//
		// Sichtbar wurde der Unterschied an den kanonischen Sets: sie
		// speichern Manakosten als `[3][W]`, und die Klammern standen als
		// leere Kästen auf jeder Karte
		// (#411).
		{
			"Unbekanntes fällt weg",
			`sort_text("RxW", order: "cycle(WUBRG)")`, "RW",
		},

		// **Ein Zyklus neben festen Zeichen.** Die Reihenfolge der Ordnung
		// gilt weiter: `X` vor dem Zyklus, `Y` danach.
		{
			"eingebettet in eine längere Ordnung",
			`sort_text("YWX", order: "Xcycle(WUBRG)Y")`, "XWY",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartt {
				t.Errorf("%q, erwartet war %q", wert.AlsText(), fall.erwartt)
			}
		})
	}
}

// **Eine offene runde Klammer bleibt ein Fehler.**
//
// Die Gegenprobe: `(` heißt jetzt etwas, und genau deshalb muss ein `(` ohne
// `)` gemeldet werden — sonst verschluckte die Ordnung den Rest still.
func TestEineOffeneRundeKlammerWirdGemeldet(t *testing.T) {
	for _, quelle := range []string{
		`sort_text("a", order: "(WUBRG")`,
		`sort_text("a", order: "cycle(WUBRG")`,
	} {
		t.Run(quelle, func(t *testing.T) {
			_, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
			if err == nil {
				t.Fatal("kein Fehler")
			}
		})
	}
}
