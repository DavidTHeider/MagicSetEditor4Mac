package service

import (
	"context"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Stil darf eigene Felder mitbringen — `extra card field:`.**
//
//	REFLECT(extra_card_fields);
//	REFLECT_IF_READING { extra_card_style.init(extra_card_fields); }
//	REFLECT(extra_card_style);
//
// `src/data/stylesheet.cpp`, Zeilen 116–125, Fassung f9d9356 vom 2021-07-06.
//
// **Daran hängt der Kasten für Stärke und Widerstandskraft.** `magic-m15`
// zeichnet ihn nicht über ein Kartenfeld, sondern über `pt box` — ein Feld, das
// es im Spiel nicht gibt und das die Vorlage aus `card.card_color` rechnet
// (#270).

// zusatzfelder rechnet die Werte der `extra card field:`-Blöcke.
//
// **Nach den Kartenfeldern**, denn sie lesen sie: `script: card.card_color`.
// Umgekehrt liest kein Kartenfeld sie — bis auf `vorthos_box`, das
// `extra_card.vorthos_box or else ""` schreibt und den Fall damit selbst
// abfängt.
func zusatzfelder(
	ctx context.Context, felder []domain.Field, gespeichert map[string]string,
	umgebung *msescript.Umgebung,
) ([]msescript.Eintrag, []domain.Befund) {
	eintraege := make([]msescript.Eintrag, 0, len(felder))

	var befunde []domain.Befund

	for i := range felder {
		feld := felder[i]

		if ctx.Err() != nil {
			return eintraege, befunde
		}

		name := strings.ReplaceAll(feld.Name, " ", "_")

		// **Der gespeicherte Wert vor der Vorgabe** (#2298): MSE rechnet das
		// `default:`-Skript nur über einem Wert, der noch Vorgabe ist
		// (`invokeOnDefault`, `ChoiceValue::update`,
		// `src/data/field/choice.cpp`, Zeilen 290–295, Fassung f9d9356 vom
		// 2021-07-06); das `script:` läuft darunter über beiden.
		wert, da := gespeichert[name]
		if !da {
			//nolint:errcheck // ein Zusatzfeld ohne Vorgabe ist der Normalfall
			wert, _ = vorgabewert(ctx, feld, umgebung, "")
		}

		if roh := strings.TrimSpace(feld.Skript); roh != "" {
			eigene := umgebung.Darueber()
			eigene.Setze("value", msescript.Text(wert))

			gerechnet, err := msescript.AuswertenQuelle(ctx, skriptrumpfOder(roh), eigene)

			// **Ein Fehlerwert wiegt so schwer wie ein Fehler** — beide
			// bekommen denselben Befund, und beide lassen den Wert stehen.
			switch {
			case err != nil:
				befunde = append(befunde, zusatzbefund(feld.Name, err.Error()))

			case gerechnet.IstFehler():
				befunde = append(befunde, zusatzbefund(feld.Name, gerechnet.Text))

			default:
				wert = gerechnet.AlsText()
			}
		}

		eintraege = append(eintraege, msescript.Eintrag{Name: name, Wert: msescript.Text(wert)})
	}

	return eintraege, befunde
}

// zusatzwerte holt die gerechneten Zusatzfelder wieder aus der Umgebung.
//
// **Aus der Umgebung und nicht aus einem zweiten Rückgabewert**: dort stehen
// sie ohnehin, weil die Bildskripte sie über `extra_card` lesen, und zwei
// Quellen für dieselbe Auskunft laufen auseinander.
//
// **Ohne Wächter auf den Namen**: `extra_card` wird in [umgebungAus] gesetzt,
// bevor irgendetwas es liest — erst leer, dann gerechnet. Ein Zweig für „gibt
// es nicht" wäre nie betreten worden (§3).
func zusatzwerte(umgebung *msescript.Umgebung) []msescript.Eintrag {
	wert, _ := umgebung.Suche("extra_card")

	return wert.Eintraege
}

// alleFeldstile bringt Karten- und Zusatzfelder in **eine** Zeichenreihenfolge.
//
//	setStyles(stylesheet, stylesheet->card_style, &stylesheet->extra_card_style);
//
// `CardViewer::setCard`, `src/render/card/viewer.cpp`, Zeile 138, Fassung
// f9d9356 vom 2021-07-06 — MSE führt beide Listen in einer Ansicht zusammen,
// und dort entscheidet allein `z index`. Wer sie hintereinander hängte, legte
// den Kasten für Stärke und Widerstandskraft über den Rahmen statt darunter.
//
// **Stabil sortiert**: zwei Ebenen mit derselben Zahl behalten ihre
// Reihenfolge, und die Kartenfelder stehen dabei vorn.
func alleFeldstile(karte, zusatz []msepaket.Feldstil) []zeichenfeld {
	alle := make([]zeichenfeld, 0, len(karte)+len(zusatz))

	for stelle := range karte {
		alle = append(alle, zeichenfeld{stil: &karte[stelle]})
	}

	for stelle := range zusatz {
		alle = append(alle, zeichenfeld{stil: &zusatz[stelle], zusatz: true})
	}

	sort.SliceStable(alle, func(a, b int) bool {
		return alle[a].stil.ZIndex < alle[b].stil.ZIndex
	})

	return alle
}

// zeichenfeld ist ein Feldstil samt der Auskunft, woher seine Werte kommen.
type zeichenfeld struct {
	stil   *msepaket.Feldstil
	zusatz bool
}

// zusatzbefund baut die Meldung eines gescheiterten Zusatzfeldes.
func zusatzbefund(feld, grund string) domain.Befund {
	return domain.Befund{
		Ort:     feld,
		Meldung: "das Skript des Zusatzfeldes ließ sich nicht ausrechnen: " + grund,
	}
}
