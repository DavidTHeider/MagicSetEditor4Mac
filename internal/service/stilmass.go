package service

// Maße eines Feldstils: aus einer Angabe wird eine Zahl, aus vier Zahlen ein
// Rahmen.
//
// **Getrennt von `stilplan.go`**, weil es eine andere Frage beantwortet: dort
// steht, *welche* Ebene ein Feld ergibt, hier, *wo* sie liegt.

import (
	"context"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// achse rechnet Anfang und Länge einer Richtung aus drei Angaben.
//
// # Zwei von dreien bestimmen die dritte
//
// Ein Feldstil nennt `left`, `right` und `width` — aber selten alle drei.
// `magic-m15.mse-style` führt für `name:` die Kombination `left`+`right`, für
// `casting cost:` dagegen **`right`+`width` ohne `left`**:
//
//	casting cost:
//	    right: 346
//	    width: { max(30, card_style.casting_cost.content_width) + 5 }
//
// Bis #240 nahm dieses Programm den Anfang immer als `left` und damit null —
// die Manakosten landeten am linken Kartenrand statt rechts oben, wo sie auf
// jeder Magic-Karte stehen. Am gerenderten Bild war es zu sehen, und in keinem
// Test: eine Vorrichtung, die `left` immer mitgibt, prüft den Fall nie.
//
// **Eine ausdrückliche Null ist keine fehlende Angabe.** Wer `left: 0` schreibt,
// meint den linken Rand — unterschieden wird über die leere Zeichenkette, nicht
// über den Wert.
func achse(
	ctx context.Context, feld, kante, gegenkante, laenge string,
	karte float64, umgebung *msescript.Umgebung,
) (anfang, spanne float64, befunde []domain.Befund) {
	anfang, spanne, befunde = roheAchse(ctx, feld, kante, gegenkante, laenge, karte, umgebung)

	// **Eine negative Länge wird zum Betrag** — MSEs letzte zwei Zeilen in
	// `Style::update`:
	//
	//	if (width  < 0) width  = -width;
	//	if (height < 0) height = -height;
	//
	// `src/data/field.cpp`, Zeilen 185–186, Fassung f9d9356 vom 2021-07-06.
	//
	// **Der Fall steht im Pack.** Das Namensfeld von `magic-m15` rechnet
	//
	//	left:  { if card.card_symbol == "none" then 32 else 50 }
	//	right: { 341 - card_style.casting_cost.content_width }
	//
	// — bei sehr breiten Manakosten liegt `right` links von `left`. Über alle
	// 23.168 kanonischen Karten meldete dieses Programm dafür **20-mal** „ein
	// Textrahmen von −309×23" und ließ die Ebene fallen; MSE dreht sie um und
	// zeichnet (#504).
	//
	// **Die Kante bleibt, wo sie ist.** MSE dreht allein die Länge um, nicht
	// die Stelle — der Kasten wächst nach rechts beziehungsweise nach unten.
	if spanne < 0 {
		spanne = -spanne
	}

	return anfang, spanne, befunde
}

// roheAchse rechnet die Achse aus Kante, Gegenkante und Länge — **ohne** den
// Betrag; den nimmt [achse].
func roheAchse(
	ctx context.Context, feld, kante, gegenkante, laenge string,
	karte float64, umgebung *msescript.Umgebung,
) (anfang, spanne float64, befunde []domain.Befund) {
	if kante != "" {
		anfang, befunde = masszahl(ctx, feld, "kante", kante, Kantenvorgabe, umgebung)

		if laenge != "" {
			gemessen, weitere := masszahl(ctx, feld, "länge", laenge, Laengenvorgabe, umgebung)

			return anfang, gemessen, append(befunde, weitere...)
		}

		if gegenkante == "" {
			// Ohne Länge und ohne Gegenkante reicht das Feld bis zum Rand — die
			// gutmütigste Deutung, und sichtbar, statt eine Ebene der Länge null
			// zu bauen, die nichts zeichnet und niemandem auffällt.
			return anfang, karte - anfang, befunde
		}

		ende, weitere := masszahl(ctx, feld, "gegenkante", gegenkante, Kantenvorgabe, umgebung)

		return anfang, ende - anfang, append(befunde, weitere...)
	}

	// **Keine Kante: die Gegenkante verankert.**
	if gegenkante != "" && laenge != "" {
		ende, befunde := masszahl(ctx, feld, "gegenkante", gegenkante, Kantenvorgabe, umgebung)
		gemessen, weitere := masszahl(ctx, feld, "länge", laenge, Laengenvorgabe, umgebung)

		return ende - gemessen, gemessen, append(befunde, weitere...)
	}

	if gegenkante != "" {
		ende, befunde := masszahl(ctx, feld, "gegenkante", gegenkante, Kantenvorgabe, umgebung)

		return 0, ende, befunde
	}

	if laenge != "" {
		gemessen, befunde := masszahl(ctx, feld, "länge", laenge, Laengenvorgabe, umgebung)

		return 0, gemessen, befunde
	}

	return 0, karte, nil
}

// masszahl wertet eine Angabe aus, die eine Zahl **oder** ein Skript sein darf.
//
// **`vorwert` ist der bisherige Wert dieser Angabe** und steht dem Skript als
// `value` zur Verfügung — MSE bindet ihn in `Scriptable<T>::update`:
//
//	inline bool update(Context& ctx) {
//	  return script.invokeOn(ctx, value);
//	}
//
// `src/script/scriptable.hpp`, Zeilen 134–136, Fassung f9d9356 vom 2021-07-06.
func masszahl(
	ctx context.Context, feld, angabe, quelle string, vorwert float64,
	umgebung *msescript.Umgebung,
) (float64, []domain.Befund) {
	// **Kein Zweig für die leere Angabe.** [achse] fragt vorher, ob eine
	// Angabe dasteht, und ruft nur dann hierher. Ein Zweig, den kein Aufruf
	// betreten kann, ist eine Behauptung darüber, was im Ernstfall geschieht
	// (§3).
	roh := strings.TrimSpace(quelle)

	if zahl, err := strconv.ParseFloat(roh, 64); err == nil {
		return zahl, nil
	}

	// Die geschweifte Klammer wird abgestreift — der Grund steht bei
	// [skriptrumpf].
	roh, _ = skriptrumpf(roh)

	// **Eine eigene Ebene je Angabe.** `value` gehört zu *dieser* Angabe; auf
	// die geteilte Umgebung gesetzt sähe die nächste den Wert der vorigen.
	eigene := umgebung.Darueber()
	eigene.Setze("value", msescript.Zahl(vorwert))

	wert, err := msescript.AuswertenQuelle(ctx, roh, eigene)
	if err != nil {
		return 0, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ließ sich nicht ausrechnen: " + err.Error(),
		}}
	}

	// **Ein Fehlerwert nennt seinen eigenen Grund.** Er sagt, **welcher**
	// Zugriff ins Leere ging; „ergab keine Zahl, sondern Fehler" sagte nur,
	// dass etwas schieflief (siehe [msescript.Fehlerwert]).
	if wert.IstFehler() {
		return 0, []domain.Befund{{
			Ort:     feld + " → " + angabe,
			Meldung: "die Angabe ergab keine Zahl: " + wert.Text,
		}}
	}

	if wert.Art != msescript.ArtZahlwert {
		zahl, fehler := strconv.ParseFloat(strings.TrimSpace(wert.AlsText()), 64)
		if fehler != nil {
			return 0, []domain.Befund{{
				Ort: feld + " → " + angabe,
				// **Die Art dazu, und den Wert in Anführungszeichen.**
				//
				// Ein Befund, der seinen Gegenstand nicht nennt, ist eine
				// Zeile im Protokoll und keine Auskunft (§8).
				Meldung: "die Angabe ergab keine Zahl, sondern " +
					artwort(wert) + ": " + roh,
			}}
		}

		return zahl, nil
	}

	return wert.Zahl, nil
}

// artwort benennt, was eine Maßangabe stattdessen ergeben hat.
//
// **`nichts` bekommt ein eigenes Wort**, denn es ist der häufigste und der
// undurchsichtigste Fall: ein Name, den es gibt, mit einem Schlüssel, den es
// nicht gibt. Ein leerer Text sieht genauso aus und ist etwas anderes.
func artwort(wert msescript.Wert) string {
	if wert.Art == msescript.ArtNichts {
		return "nichts"
	}

	return string(wert.Art) + " " + strconv.Quote(wert.AlsText())
}
