package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`sort_list` vergleicht die Schlüssel als Text, und zwar klug** —
// #2006.
//
//	values.push_back(make_pair(order_by.eval(ctx)->toString(), v));
//	…
//	sort(values.begin(), values.end(), smart_less_first);
//
// `src/script/functions/basic.cpp`, Zeilen 489–520, Fassung f9d9356 vom
// 2021-07-06. Gemessen am Leistungsvergleich: `magic-text` ordnet nach
// `card.card_number`, und das Feld ist **Text** wie `2/99` — Byte für Byte
// verglichen stand `10/99` vor `2/99`, und die Ausfuhr schrieb die Karten in
// anderer Reihenfolge als MSE2.
func TestSortListeOrdnetKartennummernAlsText(t *testing.T) {
	t.Parallel()

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list(["10/99", "2/99", "1/99"], order_by: {input})`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	gesehen := make([]string, 0, len(wert.Eintraege))
	for _, eintrag := range wert.Eintraege {
		gesehen = append(gesehen, eintrag.Wert.Text)
	}

	if len(gesehen) != 3 || gesehen[0] != "1/99" || gesehen[1] != "2/99" || gesehen[2] != "10/99" {
		t.Errorf("%v — erwartet 1/99, 2/99, 10/99", gesehen)
	}
}

// **Groß und klein sind gleich, und gleiche Schlüssel bleiben, wo sie waren.**
//
// `Apfel` und `apfel` sind für `smart_less` gleich; die Ordnung lässt sie in
// der Reihenfolge der Eingabe (`TestSortListeIstStabil`).
func TestSortListeOrdnetOhneAufGrossUndKleinZuSehen(t *testing.T) {
	t.Parallel()

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`sort_list(["banane", "apfel", "Apfel"], order_by: {input})`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if len(wert.Eintraege) != 3 ||
		wert.Eintraege[0].Wert.Text != "apfel" ||
		wert.Eintraege[1].Wert.Text != "Apfel" ||
		wert.Eintraege[2].Wert.Text != "banane" {
		t.Errorf("%+v", wert.Eintraege)
	}
}
