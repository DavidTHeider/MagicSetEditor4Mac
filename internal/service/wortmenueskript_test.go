package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zeigeMenue macht aus einem Menü eine lesbare Zeile: `Name` und `Name_` für
// einen Eintrag mit Strich darunter.
func zeigeMenue(menue service.Wortmenue) string {
	teile := make([]string, 0, len(menue.Eintraege))

	for _, eintrag := range menue.Eintraege {
		name := eintrag.Name
		if eintrag.ZeileDarunter {
			name += "_"
		}

		teile = append(teile, name)
	}

	return strings.Join(teile, "|")
}

// **Ein Skript liefert seine Wörter als eine Zeichenkette mit Kommas** —
// `DropDownWordList::addWordsFromScript`, `src/gui/value/text.cpp`, Zeilen
// 181–213, Fassung f9d9356 vom 2021-07-06.
//
// Im echten Pack steht dort `all_sub_types(setting: "is_land")`, und das gibt
// `",Forest,Island,…"` — mit führendem Komma, ohne Leerzeichen. 77 der 432
// Wörter haben ein Skript (gemessen am 2026-08-24 am `M15-Magic-Pack`).
func TestEinSkriptLiefertDieWoerter(t *testing.T) {
	faelle := []struct {
		name     string
		ergebnis string
		darunter bool
		erwartet string
	}{
		{
			name:     "drei Wörter",
			ergebnis: "Forest,Island,Mountain",
			erwartet: "Forest|Island|Mountain",
		},
		{
			// **Sortiert wird immer** — `sort(strings.begin(), strings.end())`,
			// Zeile 201. Eine Liste aus dem Set stünde sonst in
			// Kartenreihenfolge.
			name:     "unsortiert kommt sortiert an",
			ergebnis: "Mountain,Forest,Island",
			erwartet: "Forest|Island|Mountain",
		},
		{
			// **Doppelte fallen heraus** — `if (prev >= items.size() || s !=
			// items[prev].name)`, Zeile 206. Das Sortieren stellt sie
			// nebeneinander; verglichen wird nur mit dem Vorgänger.
			name:     "doppelte fallen heraus",
			ergebnis: "Forest,Island,Forest",
			erwartet: "Forest|Island",
		},
		{
			// **Das führende Komma gibt einen leeren Eintrag** — und der ist
			// die Wahl »nichts davon«: wer ihn nimmt, leert das Feld.
			name:     "das führende Komma",
			ergebnis: ",Forest,Island",
			erwartet: "|Forest|Island",
		},
		{
			name:     "ein einzelnes Wort ohne Komma",
			ergebnis: "Forest",
			erwartet: "Forest",
		},
		{
			// **Der Strich steht unter dem letzten** — `items[prev].flags = 0`
			// nimmt ihn dem Vorgänger wieder weg (Zeile 208). Ein Strich unter
			// jedem Eintrag wäre kein Trenner mehr.
			name:     "der Strich bleibt beim letzten",
			ergebnis: "Forest,Island,Mountain",
			darunter: true,
			erwartet: "Forest|Island|Mountain_",
		},
		{
			// **Und der leere Eintrag behält seinen.** Der Vorgänger wird nur
			// dann entstrichen, wenn er einen Namen hat: `!items[prev].name
			// .empty()`. So steht der Trenner zwischen »nichts« und den
			// Wörtern **und** unter der Gruppe.
			name:     "der leere Eintrag behält seinen Strich",
			ergebnis: ",Forest,Island",
			darunter: true,
			erwartet: "_|Forest|Island_",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			woerter := []msepaket.Wortlisteneintrag{
				{Skript: "all_sub_types()", ZeileDarunter: fall.darunter},
			}

			menue := service.NeuesWortmenue(woerter, "", func(skript string) string {
				if skript != "all_sub_types()" {
					t.Errorf("gefragt wurde nach %q", skript)
				}

				return fall.ergebnis
			})

			if zeigeMenue(menue) != fall.erwartet {
				t.Errorf("das Menü ist %q, erwartet %q", zeigeMenue(menue), fall.erwartet)
			}
		})
	}
}

// **Ein Skriptwort gibt seine Art an die Wörter weiter.**
//
// `DropDownWordListItem(w, s, …)` behält `w` (`src/gui/value/text.cpp`, Zeilen
// 94–99, Fassung f9d9356 vom 2021-07-06); gefragt wird später
// `item.word->is_prefix` (Zeilen 217, 256, 283).
// Im Pack stehen so sieben spanische Präfixe als `script: " tribal"`.
func TestEinSkriptwortGibtSeineArtWeiter(t *testing.T) {
	woerter := []msepaket.Wortlisteneintrag{
		{Skript: "s", IstPraefix: true},
	}

	menue := service.NeuesWortmenue(woerter, "", func(string) string {
		return "Token ,Legendary "
	})

	if !menue.Kaestchen {
		t.Error("ein Skriptwort mit »is prefix« braucht Kästchen")
	}

	for _, eintrag := range menue.Eintraege {
		if !eintrag.IstPraefix {
			t.Errorf("%q ist kein Präfix geworden", eintrag.Name)
		}
	}
}

// **Das Leerzeichen nach einem Komma bleibt stehen** — und das ist ein Fehler
// in MSE, der hier mitportiert wird (§15.4).
//
//	prev = comma + 1;
//	if (prev + 1 < str.size() && str.GetChar(prev + 1) == _(' ')) ++prev;
//
// `src/gui/value/text.cpp`, Zeilen 195–196, Fassung f9d9356 vom 2021-07-06.
// Geprüft wird das Zeichen **hinter** dem ersten nach dem Komma; bei `"a, b"`
// steht dort `b`, also wird nichts übersprungen und der Eintrag heißt `" b"`.
// Bei zwei Leerzeichen greift die Zeile — und lässt eines stehen.
//
// **Warum nicht behoben**: das Ergebnis ist der Name, der eingesetzt wird. Wer
// hier ein Leerzeichen wegnimmt, ändert den Wert, den ein Set bekommt, und die
// Karte sähe in MSE2 anders aus als hier. Im echten Pack tritt der Fall nicht
// auf: `all_sub_types` trennt ohne Leerzeichen.
func TestDasLeerzeichenNachDemKommaBleibt(t *testing.T) {
	// **Und danach wird sortiert** — ein Leerzeichen steht vor jedem
	// Buchstaben, also rutscht der Eintrag mit ihm nach vorn.
	faelle := map[string]string{
		"a, b":  " b|a",
		"a,  b": " b|a",
		"a,b":   "a|b",
	}

	for ergebnis, erwartet := range faelle {
		t.Run(ergebnis, func(t *testing.T) {
			woerter := []msepaket.Wortlisteneintrag{{Skript: "s"}}

			menue := service.NeuesWortmenue(woerter, "", func(string) string {
				return ergebnis
			})

			if zeigeMenue(menue) != erwartet {
				t.Errorf("aus %q wird %q, erwartet %q", ergebnis, zeigeMenue(menue), erwartet)
			}
		})
	}
}
