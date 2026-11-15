package msescript

import (
	"context"
	"strings"
)

// Drei Textfunktionen, die die Exportvorlagen rufen.
//
// # Erschlossen, nicht erfunden
//
// Jede steht mit ihrer Fundstelle bei ihrer Funktion. Was hier **nicht** steht,
// ist der Rest der Lücke: `card_name`, `cmc` und `booster_sort` kommen aus dem
// Spielpaket und nicht aus der Bibliothek, `write_image_file` schreibt Dateien
// und ist ein Werkstück für sich
// (#202).
func init() {
	eintragen("to_string", zuZeichenkette)
	eintragen("trim", beschneiden)
	eintragen("substring", teilkette)
	eintragen("to_upper", zuGross)
}

// zuZeichenkette — `to_string("Card URLs\n"+write_links)` aus
// `magic-toexp.mse-export-template`.
//
// **Sie ist der Abschluss einer Exportvorlage**: jede der neunzehn endet damit,
// ihren zusammengesetzten Text zurückzugeben. Was hereinkommt, ist bereits
// Text; die Funktion macht aus einer Liste oder Zahl trotzdem einen.
func zuZeichenkette(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_string"); err != nil {
		return Nichts, err
	}

	// **`format:` wird gelesen, nicht verworfen** (`basic.cpp`, Zeilen 89–99,
	// Fassung f9d9356). Sieben Stellen im vollen Pack rufen
	// `to_string(card.time_created, format: "%y-%m-%d, %H:%M:%S")`; bis zum
	// 2026-08-22 gab diese Funktion das Argument stillschweigend auf
	// (#843).
	if wort, da := ruf.Benannter("format"); da && wort.Art == ArtTextwert {
		gesetzt, err := formatiere(wort.Text, ruf.Stellung[0])
		if err != nil {
			return Nichts, ruf.FehlerAn(err.Error())
		}

		return Text(gesetzt), nil
	}

	return Text(ruf.Stellung[0].AlsText()), nil
}

// zuGross — `to_upper(card.card_code_text)` aus `magic-lackeybot`.
//
// **Das Geschwister von `to_lower`**, das seit Phase 3 dasteht; dass nur eine
// von beiden gebaut war, lag daran, dass die Exportvorlagen erst mit #51
// überhaupt gelesen wurden.
func zuGross(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_upper"); err != nil {
		return Nichts, err
	}

	return Text(strings.ToUpper(ruf.Stellung[0].AlsText())), nil
}

// beschneiden — `trim(card.super_type)` aus `magic-cockatrice-v2`.
//
// **Beidseitig und nur Leerraum.** Die Vorlagen rufen sie auf Kartenfelder, die
// mit einem Leerzeichen enden können, weil ein Skript sie zusammengesetzt hat.
func beschneiden(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("trim"); err != nil {
		return Nichts, err
	}

	return Text(strings.TrimSpace(ruf.Stellung[0].AlsText())), nil
}

// teilkette — `substring(card.illustrator, end:7)` und `substring(_2, begin: 1)`
// aus `magic-cockatrice-v2` und `magic-lackeybot`.
//
// **`begin` und `end` sind Stellen, keine Länge**, und beide sind einzeln
// weglassbar: die Vorlagen benutzen jede Form.
//
// **Gezählt wird in Runen, nicht in Bytes.** Ein Kartenname mit Umlaut käme
// sonst mitten in einem Zeichen auseinander — und der Name eines Illustrators
// ist genau das Feld, auf das die Vorlagen sie rufen.
//
// **Stellen außerhalb werden eingefangen statt gemeldet.** `end: 7` auf einem
// kürzeren Text ist in den Vorlagen der Normalfall — dort steht es, um lange
// Werte zu kürzen, nicht um kurze abzuweisen.
func teilkette(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("substring"); err != nil {
		return Nichts, err
	}

	runen := []rune(ruf.Stellung[0].AlsText())

	von := stelle(ruf, "begin", 0, len(runen))
	bis := stelle(ruf, "end", len(runen), len(runen))

	if von > bis {
		von = bis
	}

	return Text(string(runen[von:bis])), nil
}

// stelle liest eine benannte Stelle und hält sie im Text.
func stelle(ruf Ruf, name string, vorgabe, laenge int) int {
	wert, da := ruf.Benannter(name)
	if !da {
		return vorgabe
	}

	gewaehlt := int(wert.Zahl)

	return max(0, min(gewaehlt, laenge))
}
