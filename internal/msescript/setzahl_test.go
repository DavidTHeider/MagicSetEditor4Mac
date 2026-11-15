package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`length` und `number_of_items` zählen die Karten des Sets, auch durch
// einen Filter** — Folge aus
// #2006.
//
//	if (ScriptObject<Set*>* setobj = dynamic_cast<ScriptObject<Set*>*>(collection.get())) {
//	  SCRIPT_OPTIONAL_PARAM_C_(ScriptValueP, filter);
//	  return set->numberOfCards(filter);
//
// `src/script/functions/basic.cpp`, Zeilen 555–563, und `Set::numberOfCards`,
// `src/data/set.cpp`, Zeilen 301–314, Fassung f9d9356 vom 2021-07-06.
// `card_count()` aus `magic.mse-game` zählt so — bei einer Kartennummer mit
// Gesamtzahl (`0001/0099`) stand hier die Zahl der Glieder des Sets, nicht der
// Karten, die der Filter behält.

// setMitFeld ist ein Set mit drei Karten und einem Feld daneben — das Feld ist
// keine Karte und zählt nicht.
func setMitFeld(t *testing.T) *msescript.Umgebung {
	t.Helper()

	set, _ := satzmitKarten("Anton", "Berta", "Adam")
	set.Eintraege = append(set.Eintraege,
		msescript.Eintrag{Name: "title", Wert: msescript.Text("Probe")})

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("set", set)

	return umgebung
}

func TestDieKartenDesSetsWerdenGezaehlt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ quelle, erwartet string }{
		{`length(set)`, "3"},
		{`length(set, filter: {card.art == "A"})`, "2"},
		{`number_of_items(in: set)`, "3"},
		{`number_of_items(in: set, filter: {card.art == "B"})`, "1"},
		{`length(set, filter: {false})`, "0"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, setMitFeld(t))
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%s ergab %s, erwartet %s", fall.quelle, wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Ein Filter, der nicht auswertbar ist, bricht ab** — wie bei `position`.
func TestEinUnbrauchbarerFilterBeimZaehlenIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, err := msescript.AuswertenQuelle(t.Context(),
		`length(set, filter: {unbekannt()})`, setMitFeld(t)); err == nil {
		t.Error("kein Fehler")
	}
}

// **Gezählt wird einmal je Filter, wenn der Speicher da ist** — MSEs
// `filter_cache`. Ohne ihn liefe `card_count()` je Karte über das ganze Set:
// bei 3.788 Karten vierzehn Millionen Filterläufe.
func TestDieZahlJeFilterWirdGemerkt(t *testing.T) {
	t.Parallel()

	umgebung := setMitFeld(t)
	umgebung.OrdnungenMerken()

	if err := msescript.Vorbelegen(t.Context(),
		`gezaehlt := { length(set, filter: {trace(card.art) == "A"}) }`, umgebung); err != nil {
		t.Fatalf("erklären: %v", err)
	}

	for range 4 {
		wert, err := msescript.AuswertenQuelle(t.Context(), `gezaehlt()`, umgebung.Darueber())
		if err != nil || wert.AlsText() != "2" {
			t.Fatalf("gezählt %v (%v), erwartet 2", wert.AlsText(), err)
		}
	}

	if gelaufen := len(umgebung.Meldungen()); gelaufen != 3 {
		t.Errorf("der Filter lief %d-mal, erwartet 3 (einmal je Karte)", gelaufen)
	}
}
