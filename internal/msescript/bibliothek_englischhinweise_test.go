package msescript_test

import "testing"

// **`process_english_hints` löst Einzahl und Mehrzahl im Erinnerungstext auf**
// — #1068.
//
// # Was die Auszeichnungen bedeuten
//
// Der Quelltext sagt es selbst (`src/script/functions/english.cpp`, Zeilen
// 224–236, Fassung f9d9356 vom 2021-07-06):
//
//   - `"<hint-1>xxx(yyy)zzz"` → `"xxxzzz"` (Einzahl)
//   - `"<hint-2>xxx(yyy)zzz"` → `"xxxyyyzzz"` (Mehrzahl)
//   - `"[^., ]a <param-..>[aeiou]"` → `"\1 an \2"` (der Artikel `an`)
//   - `"<hint-?>"` → `""` (der Hinweis selbst fällt danach weg)
func TestDieEnglischenHinweiseWerdenAufgeloest(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, hinein, heraus string }{
		{
			"Einzahl lässt die Klammer fallen",
			"<hint-1>Draw a card(s).", "Draw a card.",
		},
		{
			"Mehrzahl behält ihren Inhalt",
			"<hint-2>Draw two card(s).", "Draw two cards.",
		},
		{
			"ohne Hinweis bleibt die Klammer stehen",
			"Draw a card(s).", "Draw a card(s).",
		},
		{
			"der Hinweis selbst verschwindet",
			"<hint-1>nichts", "nichts",
		},
		{
			"<singular> bleibt bei Einzahl",
			"<hint-1><singular>eine</singular>", "eine",
		},
		{
			"<plural> bleibt bei Mehrzahl",
			"<hint-2><plural>viele</plural>", "viele",
		},
		{
			// **Der Hinweis gilt für genau einen Zweig.** `<singular>` setzt
			// ihn danach zurück (`english.cpp`, Zeile 301) — der zweite Zweig
			// steht dann ohne Hinweis da und fällt weg. Beide hintereinander
			// ergeben deshalb **nichts**, nicht den zweiten.
			//
			// Gemessen am Original, nicht überlegt: der erste Anlauf erwartete
			// hier »viele«.
			"zwei Zweige hintereinander ergeben nichts",
			"<hint-2><singular>eine</singular><plural>viele</plural>", "",
		},
		{
			// **Der Artikel.** `a` vor einem Vokal wird `an`.
			"a wird an vor einem Vokal",
			"gains a <param-x>artifact</param-x>", "gains an <param-x>artifact</param-x>",
		},
		{
			"a bleibt vor einem Mitlaut",
			"gains a <param-x>creature</param-x>", "gains a <param-x>creature</param-x>",
		},
		{
			// **Ein leerer Parameter nimmt das Leerzeichen davor mit**
			// („empty param, drop space before it", Zeilen 284–288). Sonst
			// stünde in „two ⟨leer⟩cards" ein doppelter Abstand.
			"ein leerer Parameter frisst sein Leerzeichen",
			"zwei <param-x></param-x>Karten", "zwei<param-x></param-x>Karten",
		},
		{
			"gewöhnliche Auszeichnungen bleiben",
			"<i>kursiv</i>", "<i>kursiv</i>",
		},
		{
			"eine Klammer ohne Hinweis bleibt",
			"(nur Text)", "(nur Text)",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			pruefeEnglisch(t, "process_english_hints", fall.hinein, fall.heraus)
		})
	}
}

// **Die Randfälle der Hinweisauflösung.**
func TestDieRandfaelleDerEnglischenHinweise(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, hinein, heraus string }{
		// **Eine Klammer ohne Ende bleibt ein Zeichen.**
		{"Klammer ohne Ende", "<hint-1>a(b", "a(b"},
		// **Ein Zweig ohne Ende bricht ab, statt zu raten.**
		{"Zweig ohne Ende", "<hint-1><singular>eine", "eine"},
		// **`<hint-` am Textende, ohne Zahl.**
		{"Hinweis ohne Zahl", "<hint-", ""},
		// **Eine Auszeichnung ohne schließende Klammer.**
		{"Auszeichnung ohne Ende", "a<i", "a<i"},
		// **Der Artikel wächst nicht mitten im Wort**: „Alpha" endet auf `a`,
		// aber davor steht ein Buchstabe.
		{"kein Artikel mitten im Wort", "Alpha <param-x>artifact</param-x>", "Alpha <param-x>artifact</param-x>"},
		// **Zu kurz für einen Artikel** — vor dem Parameter steht gar nichts.
		{"nichts vor dem Parameter", "<param-x>artifact</param-x>", "<param-x>artifact</param-x>"},
		// **Und ein anderes Wortende als `a `.**
		{"kein a davor", "the <param-x>artifact</param-x>", "the <param-x>artifact</param-x>"},
		// **Genau zwei Zeichen davor: der Artikel wächst.**
		{"nur der Artikel davor", "a <param-x>artifact</param-x>", "an <param-x>artifact</param-x>"},
		// **Und der Sonderfall aus `english_number_a(1)`**: der Parameter ist
		// selbst das `a`.
		{"der Parameter ist das a", "<hint-1><param-x>a</param-x> artifact", "<param-x>an</param-x> artifact"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			pruefeEnglisch(t, "process_english_hints", fall.hinein, fall.heraus)
		})
	}
}
