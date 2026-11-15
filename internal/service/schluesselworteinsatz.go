// `expand_keywords` mit einer echten Wortliste
// (#1061).
//
// # Warum die Funktion hier überschrieben wird
//
// `internal/msescript` bringt sie als Stummel mit: eine Vorlage ohne den Namen
// scheitert als Ganzes, und bis zum 2026-08-24 gab es keine Wortliste
// (#345). Die Sprache kennt keine Kartenbegriffe, und ein Sprachwerkzeug, das
// Schlüsselwörter läse, wäre die Vermischung, die §2 verhindert.
//
// **Also setzt `service` den Namen in die Umgebung** — dieselbe Naht wie bei
// `exists_in_package` und `symbol_variation`.

package service

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// schluesselworteinsatz baut die Funktion `expand_keywords` über einer Liste.
//
// # Was von MSE2 übernommen ist
//
// `expand_keyword`, `src/data/keyword.cpp`, Zeilen 704–752, Fassung f9d9356
// vom 2021-07-06 — die Namen, die der Kontext beim Zusammensetzen führt:
//
//	ctx.setVariable("keyword", …)   ctx.setVariable("mode", …)
//	ctx.setVariable("expand", …)    ctx.setVariable("reminder", …)
//	ctx.setVariable("correct_case", …)
//	ctx.setVariable("used_placeholders", …)
//
// und je Parameter `param1`, `param2` … (`keyword_match_detail`, Zeile 651).
//
// # Was **nicht** übernommen ist, und das gehört daneben
//
//   - **Der `<kw-X>`-Zustand** — MSE merkt sich je Stelle im Text, ob ein
//     Erinnerungstext ausdrücklich an- oder abgeschaltet ist (`<kw-0>`,
//     `<kw-1>`, Zeilen 494–500). Hier gilt immer der Vorgabezustand, und
//     damit entscheidet `default_expand:` allein.
//   - **`used_placeholders`** steht als `false` da: Platzhalter für fehlende
//     Parameter gibt es noch nicht.
//
// **`condition:` und `correct_case` sind seit dem 2026-08-24 da**, und der
// Grund steht auf einer echten Karte: ohne sie stand der Erinnerungstext von
// „Flying" zweimal darauf — einmal am Zeilenanfang und einmal mitten im Satz
// hinter „gains flying". Die M15-Vorlage weist den zweiten ab, weil die
// Schreibung nicht stimmt.
func schluesselworteinsatz(
	liste Wortliste, vermerk *Schluesselwortverwendung,
) msescript.Wert {
	return msescript.NeueEingebaute(func(ctx context.Context, ruf msescript.Ruf) (
		msescript.Wert, error,
	) {
		if len(ruf.Stellung) == 0 {
			return msescript.Nichts, ruf.FehlerAn("expand_keywords braucht einen Text")
		}

		zusammen, da := ruf.Benannter("combine")
		if !da {
			return msescript.Nichts, ruf.FehlerAn("expand_keywords braucht ein Argument `combine:`")
		}

		text := ruf.Stellung[0].AlsText()

		bedingung, mitBedingung := ruf.Benannter("condition")
		vorgabe, mitVorgabe := ruf.Benannter("default_expand")

		gesetzt := liste.Einsetzen(text, func(treffer Schluesselworttreffer) string {
			// **Die Bedingung entscheidet, ob ein Treffer zählt**
			// (`options.match_condition`, Zeile 714). Sie sieht dieselben
			// Namen wie `combine:`.
			if mitBedingung && !erfuellt(ctx, ruf, bedingung, treffer, true) {
				return treffer.Text
			}

			// **Ob der Erinnerungstext sichtbar ist, entscheidet
			// `default_expand:`** (Zeile 721). Ohne die Angabe: ja.
			sichtbar := true
			if mitVorgabe {
				sichtbar = erfuellt(ctx, ruf, vorgabe, treffer, true)
			}

			gebaut, fehler := zusammengesetzt(ctx, ruf, zusammen, treffer, sichtbar)
			if fehler != nil {
				// **Ein Fehler im Erinnerungstext kostet seinen Treffer, nicht
				// die Karte** — MSE fängt ihn ebenso ab und macht weiter
				// („in keyword reminder", Zeilen 733–736).
				return treffer.Text
			}

			// **Vermerkt wird hier und nicht früher** — MSE trägt erst ein,
			// **nachdem** der Erinnerungstext zusammengesetzt ist
			// (`src/data/keyword.cpp` Zeilen 749–751, Fassung `f9d9356` vom
			// 2021-07-06). Eine Bedingung, die
			// nicht zutrifft, kehrt vorher um und trägt nichts ein.
			//
			// **Und ein gescheitertes Zusammensetzen ebenso wenig.** MSE
			// käme hier nicht vorbei — es lässt den Fehler bis zum Aufrufer
			// von `expand_keywords` durch; dieses Programm behält den Text
			// und macht weiter (die Zeile darüber). Eine Verwendung zu
			// vermerken, wo nichts eingesetzt wurde, wäre eine Behauptung.
			vermerk.vermerkeAmLaufendenFeld(treffer.Wort)

			return gebaut
		})

		return msescript.Text(gesetzt), nil
	})
}

// erfuellt sagt, ob die Bedingung eines Treffers zutrifft.
//
// **Ein Fehler heißt: der Treffer zählt nicht.** MSE wirft ihn bis zum
// Aufrufer von `expand_keywords` (Zeilen 713–716); hier ist derselbe Ausgang
// die stille Ablehnung — ein Erinnerungstext, dessen Bedingung sich nicht
// ausrechnen lässt, ist kein Grund, den Kartentext zu verlieren.
func erfuellt(
	ctx context.Context, ruf msescript.Ruf, bedingung msescript.Wert,
	treffer Schluesselworttreffer, sichtbar bool,
) bool {
	umgebung := trefferumgebung(ruf, treffer, sichtbar)

	wert, fehler := msescript.RufeFunktion(ctx, bedingung, umgebung)
	if fehler != nil {
		return false
	}

	trifftZu, brauchbar := wert.AlsWahrheit()

	// **Was sich nicht als Bedingung lesen lässt, zählt nicht.** Dieselbe
	// Vorsicht wie beim Fehler darüber: eine unklare Bedingung ist kein Grund,
	// einen Erinnerungstext zu setzen, den niemand angeordnet hat.
	return brauchbar && trifftZu
}

// trefferumgebung legt die Namen an, die MSE beim Zusammensetzen setzt.
func trefferumgebung(
	ruf msescript.Ruf, treffer Schluesselworttreffer, sichtbar bool,
) *msescript.Umgebung {
	umgebung := ruf.Umgebung.Darueber()

	umgebung.Setze("keyword", msescript.Text(treffer.Text))
	umgebung.Setze("mode", msescript.Text(treffer.Modus))
	umgebung.Setze("correct_case", msescript.Wahrheit(treffer.SchreibungStimmt))
	umgebung.Setze("used_placeholders", msescript.Wahrheit(false))
	umgebung.Setze("expand", msescript.Wahrheit(sichtbar))

	for stelle, wert := range treffer.Parameter {
		umgebung.Setze("param"+strconv.Itoa(stelle+1), msescript.Text(wert))
	}

	return umgebung
}

// zusammengesetzt wertet Erinnerungstext und `combine:` für einen Treffer aus.
func zusammengesetzt(
	ctx context.Context, ruf msescript.Ruf, zusammen msescript.Wert,
	treffer Schluesselworttreffer, sichtbar bool,
) (string, error) {
	umgebung := trefferumgebung(ruf, treffer, sichtbar)

	erinnerung, fehler := msescript.AuswertenTextskript(ctx, treffer.Erinnerung, umgebung)
	if fehler != nil {
		return "", fehler
	}

	umgebung.Setze("reminder", erinnerung)

	gebaut, fehler := msescript.RufeFunktion(ctx, zusammen, umgebung)
	if fehler != nil {
		return "", fehler
	}

	return gebaut.AlsText(), nil
}

// UmgebungMitSchluesselwoertern gibt eine Bibliothek, in der
// `expand_keywords` die gegebene Wortliste benutzt.
//
// **Für Läufe über die Naht.** Im Betrieb setzt `stilumgebung.go` den Namen
// in die Umgebung des Planbauers; ein Test, der den Weg prüfen will, braucht
// dieselbe Naht ohne den ganzen Planbauer.
func UmgebungMitSchluesselwoertern(liste Wortliste) *msescript.Umgebung {
	umgebung := msescript.Bibliothek()
	umgebung.Setze("expand_keywords",
		schluesselworteinsatz(liste, &Schluesselwortverwendung{}))

	return umgebung
}
