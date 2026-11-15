package service

// **Die Angaben des Sets für die Skripte** — `set.<feld>`.
//
// Eigene Datei, weil `stilumgebung.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **welche** Umgebung ein Stil bekommt,
// hier, was **das Set** darin beisteuert.

import (
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// setangaben macht aus den Feldern des `set info:`-Blocks die Abbildung `set`.
//
// **Der Unterstrich ist dieselbe Regel wie bei der Karte**: in der Datei steht
// `border color`, die Skripte lesen `set.border_color`.
//
// # Was hier roh durchgereicht wird
//
// `set.symbol` ist bei MSE ein **Symbolwert** — der Name einer
// `.mse-symbol`-Vektordatei im Set —, und `symbol_variation` rastert ihn
// (`src/script/functions/image.cpp:142`, Fassung f9d9356 vom 2021-07-06).
// Einen Vektorrasterer gibt es hier nicht; ankommen tut der Name, wie er in der
// Datei steht
// (#323).
//
// # Was hier **nicht** drinsteht: die Karten
//
// In MSE ist `set` beides — durchlaufen gibt die Karten, `set.<name>` die
// Angabe (`src/data/set.cpp` 251–253, Fassung f9d9356 vom 2021-07-06). Die
// Ausfuhr baut es so (#653);
// beim **Zeichnen** trägt `set` allein seine Angaben.
//
// **Weil hier eine Karte gezeichnet wird und nicht ein Set.** Die Liste läge
// nicht bereit, und sie je Karte zu bauen kostete bei jedem Strich.
//
// **Und weil kein Stil sie liest**: `for each … in set` steht im ganzen Pack
// in `magic.mse-game/script` 2373–2385 und `space.mse-game/script` 414–423 —
// `all_sub_types`, `all_races`, `all_classes`. Gerufen werden sie
// ausschließlich aus `word_lists_*`, den Vorschlagslisten des Editors. Kommt
// je ein Stil hinzu, der die Karten liest, ist das hier die Stelle.
func setangaben(
	ctx context.Context, umgebung *msescript.Umgebung,
	felder []domain.Field, werte []domain.Value,
) msescript.Wert {
	// **Erst die Erklärung, dann das Ausgefüllte.** Ein Setfeld, das das Set
	// nicht ausgefüllt hat, steht mit der Vorgabe des Spiels da — und wenn es
	// keine gibt, leer. Leer und nicht *nichts*: `set.copyright + " x"` rechnet
	// bei MSE auch dann weiter, und ein fehlender Name bräche den ganzen
	// Ausdruck ab. Dieselbe Klasse wie bei den Kartenfeldern (#278).
	eintraege := make([]msescript.Eintrag, 0, len(felder)+len(werte))
	stelle := make(map[string]int, len(felder)+len(werte))

	// **Der Wert trägt sein Feld** — MSEs `value->fieldP`
	// (#548).
	// `symbol_variation` sucht seine Fassung am Stil des Feldes, aus dem der
	// Wert stammt, und ohne diese Angabe fände es die eines anderen.
	//
	// **Der ursprüngliche Name, nicht der mit Unterstrichen**: die Fassungen
	// stehen unter `masterpiece symbol:` in der Datei, mit Leerzeichen.
	setze := func(name, text string) {
		feld := name
		name = strings.ReplaceAll(name, " ", "_")

		wert := msescript.Text(text)
		wert.Feld = feld

		if i, da := stelle[name]; da {
			eintraege[i].Wert = wert

			return
		}

		stelle[name] = len(eintraege)
		eintraege = append(eintraege, msescript.Eintrag{Name: name, Wert: wert})
	}

	// **Die Vorgabe ist ein Skript**, wie bei den Kartenfeldern:
	// `magic.mse-game/set_fields` schreibt `default: "Type custom types here;
	// …"` — eine Zeichenkette in **Skriptsyntax**. Roh durchgereicht stünden
	// die Anführungszeichen im Wert (#298). [vorgabewert] ist derselbe Weg.
	//
	// **`set` ist dabei noch nicht gesetzt**: eine Vorgabe, die `set` liest,
	// scheitert und wird leer. Das ist die Reihenfolge, die MSE selbst hat —
	// ein Feld, das sich über die Abbildung definiert, in der es steht, hätte
	// keinen Anfang.
	for i := range felder {
		feld := felder[i]

		// **Der Fehler fällt hier weg, und das ist kein Verschweigen.**
		// Dieselbe Rechnung läuft in [mitVorgaben] und meldet dort; hier geht
		// es nur darum, dass `card.<feld>` einen Wert hat, während das
		// Initskript läuft.
		//nolint:errcheck // dieselbe Rechnung meldet in [mitVorgaben]
		wert, _ := vorgabewert(ctx, feld, umgebung, "")
		setze(feld.Name, wert)
	}

	for _, wert := range werte {
		setze(wert.Field, wert.Text)
	}

	// **Und die Farbfelder als Farben** — `set.border_color` und
	// `set.copyright_text_color` werden gelesen, wo eine Farbe hingehört
	// (`farbfelder.go`).
	gefaerbt := mitFarbwerten(eintraege, farbkartenfelder(felder))

	// **Dieselben Felder ein zweites Mal unter `set_info`.**
	//
	//	REFLECT_N("set_info", data);
	//	…
	//	reflect_set_info_get_member(handler, data);
	//
	// `src/data/set.cpp`, Zeilen 196 und 206, Fassung f9d9356 vom 2021-07-06.
	// Die erste Zeile steht außerhalb jeder `REFLECT_IF_READING`-Klammer und
	// gilt deshalb auch für den `GetMember`-Handler; die zweite legt dieselbe
	// Abbildung zusätzlich **namenlos** auf `set`. Der Kommentar im Altcode
	// sagt es: „in scripts, set.something is read from the set_info".
	//
	// Über `Full-Magic-Pack` steht `set.set_info.` **38-mal** da — unter
	// anderem viermal in `magic.mse-game/script`, wo sich entscheidet, ob eine
	// mehrfarbige Karte einen Verlauf bekommt
	// (#424).
	//
	// **Nicht rekursiv**: bei MSE ist `set_info` ein Glied von `Set`, nicht
	// von `data`. `set.set_info.set_info` gibt es dort nicht.
	innen := msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: gefaerbt}

	return msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: append(slices.Clone(gefaerbt),
			msescript.Eintrag{Name: "set_info", Wert: innen}),
	}
}
