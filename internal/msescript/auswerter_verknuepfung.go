package msescript

import (
	"context"
	"math"
)

// binaer rechnet eine zweistellige Verknüpfung.
//
// `and` und `or` rechnen die rechte Seite **nur bei Bedarf**. In `magic-m15`
// steht eine and-Kette, deren spätere Glieder voraussetzen, dass die früheren
// zutrafen; wer beide Seiten immer rechnet, bekommt dort einen Fehler, den die
// Vorlage nicht meint. Belegt ist das nicht — es ist die Lesart, unter der die
// gesichteten Vorlagen laufen, und sie steht so in `docs/FORMATE.md`.
func (a *auswerter) binaer(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	// **`or else` fängt den Fehlerwert — und nur ihn.**
	//
	//	case I_OR_ELSE:
	//	  if (at == SCRIPT_ERROR) a = b;
	//
	// `src/script/context.cpp`, Zeilen 535–537, Fassung f9d9356 vom
	// 2021-07-06. `nil` ist **kein** Fehler und bleibt stehen
	// (#345).
	if knoten.Text == "or else" {
		links, err := a.werte(knoten.Kinder[0], umgebung)
		if err != nil {
			return Nichts, err
		}

		// **Die rechte Seite wird nur bei Bedarf gerechnet, und das weicht
		// ab.** `I_OR_ELSE` ist bei MSE eine Binäroperation: beide Seiten
		// stehen fertig da, bevor sie läuft. Sichtbar wird der Unterschied
		// allein, wenn die rechte Seite **wirft** — dann wirft MSE auch bei
		// heiler linker Seite.
		//
		// **Gemessen über den Packlauf** — am 2026-08-20 neu, über **376
		// Stile und 8.915 Feldstile**: **370.807-mal** wird die rechte Seite
		// übersprungen. Wertet man sie trotzdem aus, ergibt sie **0-mal**
		// einen Abbruch und 905-mal einen **Fehlerwert**.
		//
		// **Und ein Fehlerwert ist auch bei MSE ein Wert**, kein Abbruch:
		// `delay_error` legt ihn auf den Stapel
		// (`src/script/functions/basic.cpp`, Zeile 123, Fassung f9d9356 vom
		// 2021-07-06), und `I_OR_ELSE` sieht ihn erst dort. Die 905 sind auf beiden Seiten harmlos; die
		// Zahl, die zählt, ist die **Null**. Damit bleibt die Abweichung
		// stehen, **gemessen folgenlos**
		// (#271).
		if !links.IstFehler() {
			return links, nil
		}

		return a.werte(knoten.Kinder[1], umgebung)
	}

	if knoten.Text == "and" || knoten.Text == "or" {
		return a.verknuepfung(knoten, umgebung)
	}

	links, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	rechts, err := a.werte(knoten.Kinder[1], umgebung)
	if err != nil {
		return Nichts, err
	}

	// **Wer mit einem Fehlerwert rechnet, bekommt den Fehler.** Bei MSE wirft
	// jede Umwandlung eines `ScriptDelayedError`; hier ist der Vergleich die
	// Stelle, an der die Rechnung ihn anfasst (siehe [Fehlerwert]).
	for _, wert := range []Wert{links, rechts} {
		if wert.IstFehler() {
			return Nichts, a.fehler(knoten, SchluesselDurchgereicht, wert.Text)
		}
	}

	switch knoten.Text {
	case "xor":
		return a.entwederOder(knoten, links, rechts)

	case "==":
		return Wahrheit(gleich(links, rechts)), nil

	case "!=":
		return Wahrheit(!gleich(links, rechts)), nil

	case "+":
		return a.plus(knoten, links, rechts, umgebung)

	case "-", "*", "/", "mod":
		return a.rechnung(knoten, links, rechts)
	}

	return a.vergleich(knoten, links, rechts)
}

// entwederOder rechnet `xor` — und **beide** Seiten.
//
// Anders als `and` und `or`: ein „entweder oder" lässt sich nicht abkürzen. Wer
// es wie `or` mit Abkürzung baute, wertete die rechte Seite nie aus, wenn die
// linke zutrifft — und das Ergebnis wäre für die Hälfte aller Fälle falsch
// (#214).
func (a *auswerter) entwederOder(knoten *Knoten, links, rechts Wert) (Wert, error) {
	einer, brauchbar := links.AlsWahrheit()
	if !brauchbar {
		return Nichts, a.fehler(knoten,
			SchluesselWahrheitLinks, "xor", string(links.Art))
	}

	anderer, brauchbar := rechts.AlsWahrheit()
	if !brauchbar {
		return Nichts, a.fehler(knoten,
			SchluesselWahrheitRechts, "xor", string(rechts.Art))
	}

	return Wahrheit(einer != anderer), nil
}

func (a *auswerter) verknuepfung(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	links, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	trifft, brauchbar := links.AlsWahrheit()
	if !brauchbar {
		return Nichts, a.fehler(knoten.Kinder[0],
			SchluesselWahrheitLinks, knoten.Text, string(links.Art))
	}

	if (knoten.Text == "and" && !trifft) || (knoten.Text == "or" && trifft) {
		return Wahrheit(trifft), nil
	}

	rechts, err := a.werte(knoten.Kinder[1], umgebung)
	if err != nil {
		return Nichts, err
	}

	rechtsTrifft, brauchbar := rechts.AlsWahrheit()
	if !brauchbar {
		return Nichts, a.fehler(knoten.Kinder[1],
			SchluesselWahrheitRechts, knoten.Text, string(rechts.Art))
	}

	return Wahrheit(rechtsTrifft), nil
}

// plus ist dreifach belegt: es rechnet, verkettet und vereinigt.
//
// Welche Bedeutung gilt, entscheiden die **Werte**. Zwei Zahlen rechnen; ist
// eine Seite eine Abbildung, wird vereinigt; sonst wird verkettet. Die
// Verkettung ist der Auffangfall, weil `"attribute/earth_" + get_lang_ex()`
// genau so aussieht — eine Zeichenkette und ein Wert, dessen Art man erst zur
// Laufzeit kennt.
func (a *auswerter) plus(
	knoten *Knoten, links, rechts Wert, umgebung *Umgebung,
) (Wert, error) {
	// **`nil` ist das neutrale Element, und zwar vor allem anderen.**
	//
	// Herkunft: `context.cpp:465`, der erste Zweig von `I_ADD` —
	// `if (at == SCRIPT_NIL) { a = b; } else if (bt == SCRIPT_NIL) { }`.
	//
	// Der Unterschied ist nicht nur der Wert, sondern die **Art**: `nil + 5`
	// ist die Zahl fünf. Fiele es in die Verkettung darunter, käme die
	// Zeichenkette `"5"` heraus, und die nächste Rechnung darüber wäre der
	// nächste Befund. Genau so entstanden Ketten von Folgefehlern, deren erster
	// weit oben stand.
	if links.Art == ArtNichts {
		return rechts, nil
	}

	if rechts.Art == ArtNichts {
		return links, nil
	}

	if links.Art == ArtZahlwert && rechts.Art == ArtZahlwert {
		return Zahl(links.Zahl + rechts.Zahl), nil
	}

	if links.Art == ArtAbbildung || rechts.Art == ArtAbbildung {
		// **Abbildung und Text ergeben Text** (#216).
		//
		// `magic.mse-game` Zeile 8411 verbindet eine Zeichenkette mit einer
		// Liste; ohne diesen Zweig bricht das Initskript dort ab, und mit ihm
		// fünf Exportvorlagen, die `cmc` rufen.
		//
		// Das ist keine Erfindung, sondern die schon getroffene Entscheidung
		// zu Ende geführt: [Wert.AlsText] einer Abbildung ist die Folge ihrer
		// Werte. Wo ein Text erwartet wird, *ist* eine Liste ihr Text.
		//
		// **Mit allem anderen wird aneinandergesetzt** — MSEs letzter Zweig
		// fängt **jede** Kombination:
		//
		//	else { a = to_script(a->toString() + b->toString()); }
		//
		// `src/script/context.cpp`, Zeile 512
		// (#270).
		//
		// Fassung f9d9356 vom 2021-07-06.
		if links.Art == ArtAbbildung && rechts.Art == ArtAbbildung {
			return a.verketten(links, rechts), nil
		}

		return Text(links.AlsText() + rechts.AlsText()), nil
	}

	if links.Art == ArtFunktion && rechts.Art == ArtFunktion {
		return verkettet(links, rechts, umgebung), nil
	}

	if links.Art == ArtFunktion || rechts.Art == ArtFunktion {
		return Nichts, a.fehler(knoten, SchluesselFunktionVerbinden,
			string(links.Art), string(rechts.Art))
	}

	return Text(links.AlsText() + rechts.AlsText()), nil
}

// gleich vergleicht über die Arten hinweg.
//
// **Eine Zahl und eine Zeichenkette sind nicht gleich**, auch wenn sie sich
// gleich schreiben. Die Vorlagen vergleichen Feldwerte gegen feste
// Zeichenketten (`card.rarity == "rare"`); eine Gleichheit über die Darstellung
// machte `1 == "1"` wahr, und das ist in keiner Vorlage gemeint.
// ungefaehrGleich ist `approx_equal` aus `value.cpp:92` von MSE2:
//
//	return a == b || fabs(a - b) < 1e-14;
//
// **In den Vorlagen wird gerechnet, bevor verglichen wird** —
// `card_style.name.width * 0.1 == 3.75`. Gleitkomma trifft solche Gleichungen
// nicht auf das letzte Bit, und ein Vergleich ohne Toleranz ist dann
// *manchmal* falsch: die unangenehmste Sorte, weil sie von den Eingabewerten
// abhängt.
func ungefaehrGleich(links, rechts float64) bool {
	return links == rechts || math.Abs(links-rechts) < 1e-14
}

func gleich(links, rechts Wert) bool {
	if links.Art != rechts.Art {
		return false
	}

	switch links.Art {
	case ArtZahlwert:
		return ungefaehrGleich(links.Zahl, rechts.Zahl)

	case ArtTextwert:
		return links.Text == rechts.Text

	case ArtWahrheitswert:
		return links.Wahrheit == rechts.Wahrheit

	case ArtNichts:
		return true

	case ArtAbbildung, ArtFarbe:
		// **Trägt eine Seite eine Kennung, entscheidet sie** — MSEs
		// `COMPARE_AS_POINTER` (siehe [Wert.Kennung]). Derselbe Gegenstand
		// bleibt derselbe, auch nachdem sich etwas an ihm geändert hat; zwei
		// verschiedene bleiben verschieden, auch wenn sie gleich aussehen.
		//
		// **Eine, nicht beide**:
		//
		//	} else if (aw == COMPARE_AS_POINTER || bw == COMPARE_AS_POINTER) {
		//	  return ap == bp;
		//	}
		//
		// `src/script/value.cpp`, Zeilen 125–127, Fassung `f9d9356` vom
		// 2021-07-06 — ein `||`. Der Zeichenkettenvergleich steht **davor** und
		// greift nur, wenn eine Seite `COMPARE_AS_STRING` meldet — Farben und
		// Kartenwerte, keine Gegenstände
		// (#1366).
		//
		// **Ein Nachbau ist damit nicht der Gegenstand**: dieselbe Gestalt,
		// keine Kennung, anderer Zeiger.
		if links.Kennung != "" || rechts.Kennung != "" {
			return links.Kennung == rechts.Kennung
		}

		// **MSE vergleicht Anzahl und Werte, nicht die Schlüssel**
		// (`src/script/value.cpp`, Zeile 107, Fassung f9d9356 vom
		// 2021-07-06). Hier zählen die Namen mit, und zwei Abbildungen mit
		// anderen Namen und derselben Wertefolge sind damit ungleich.
		//
		// **Gemessen über den Packlauf: null.** Keine einzige der 5.950
		// Feldstilrechnungen vergleicht zwei Abbildungen
		// (#271).
		return gleicheEintraege(links, rechts)

	case ArtFunktion:
		// **MSE erklärt zwei identische Zeiger für gleich**; hier ist jeder
		// Funktionsvergleich falsch. **Gemessen über den Packlauf: null.**
	}

	return false
}

func gleicheEintraege(links, rechts Wert) bool {
	if len(links.Eintraege) != len(rechts.Eintraege) {
		return false
	}

	for i := range links.Eintraege {
		if links.Eintraege[i].Name != rechts.Eintraege[i].Name ||
			!gleich(links.Eintraege[i].Wert, rechts.Eintraege[i].Wert) {
			return false
		}
	}

	return true
}

// verkettet legt zwei Funktionen hintereinander: erst die linke, dann die
// rechte auf deren Ergebnis.
//
// # Erschlossen, nicht erfunden
//
// `magic-mtgsalvation-forum.mse-export-template` reiht zehn `replace@(…)` mit
// `+` aneinander, `magic-m15-godzilla-planeswalker` verbindet `replace@(…)` mit
// `filter_text@(…)`. Beides ergibt nur einen Sinn, wenn die Anwendungen
// **nacheinander** laufen — sieben der neunzehn bezogenen Vorlagen hingen daran
// (#202).
//
// # Wie der Wert weitergereicht wird
//
// **Auf beiden Plätzen, und das hat einen Grund.** `magic-mtgsalvation-forum`
// ruft die verkettete Funktion als `paragraph_count(card.rule_text)`, und
// `replace` wie `filter_text` lesen ihr Werkstück **an Stelle eins**. Ein Block
// `{ … }` hat keine Parameterliste und liest es als **`input`**. Welche Form
// die rechte Seite hat, weiß die Verkettung nicht — also bekommt der
// Zwischenwert beide Plätze. Was eine Seite nicht liest, stört sie nicht.
//
// **Die benannten Argumente des Aufrufs reisen an beide Seiten weiter.**
func verkettet(links, rechts Wert, umgebung *Umgebung) Wert {
	linke, rechte := links, rechts

	return Wert{Art: ArtFunktion, Funktion: &Funktionsteil{
		Eingebaut: func(ctx context.Context, ruf Ruf) (Wert, error) {
			zwischen, err := ruferufen(ctx, linke, ruf, umgebung)
			if err != nil {
				return Nichts, err
			}

			// **Eine Zuteilung für die Liste, nicht drei, und so lang wie
			// nötig** (E03): ein `append` an die Liste des Rufs ließe sie
			// nachwachsen und verdoppeln, ein Literal für `input` entkäme
			// unter dem Race-Detektor als eigene Zuteilung.
			benannt := make([]Eintrag, len(ruf.Benannt), len(ruf.Benannt)+1)
			copy(benannt, ruf.Benannt)

			weiter := Ruf{
				Stellung: []Wert{zwischen},
				Benannt:  append(benannt, Eintrag{Name: "input", Wert: zwischen}),
				Zeile:    ruf.Zeile,
				Spalte:   ruf.Spalte,
			}

			return ruferufen(ctx, rechte, weiter, umgebung)
		},
	}}
}
