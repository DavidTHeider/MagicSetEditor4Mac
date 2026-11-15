package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// typenliste ist eine Liste nach dem Muster von `type-en` im echten Pack:
// vier Präfixe mit Kästchen, danach die Grundtypen als Punkte.
func typenliste() []msepaket.Wortlisteneintrag {
	return []msepaket.Wortlisteneintrag{
		{Name: "Token ", IstPraefix: true},
		{Name: "Legendary ", IstPraefix: true, ZeileDarunter: true},
		{Name: "Creature"},
		{Name: "Instant"},
		{Name: "Land"},
	}
}

// ohneSkript ist ein Auswerter, den kein Eintrag dieser Liste braucht.
func ohneSkript(t *testing.T) func(string) string {
	t.Helper()

	return func(skript string) string {
		t.Errorf("kein Eintrag hat ein Skript, gefragt wurde nach %q", skript)

		return ""
	}
}

// **Kästchen gibt es, sobald ein Eintrag ein Präfix ist** —
// `if (w->is_prefix) has_checkboxes = true;`, `DropDownWordList::setWords`,
// `src/gui/value/text.cpp`, Zeile 160, Fassung f9d9356 vom 2021-07-06.
func TestEinPraefixMachtKaestchen(t *testing.T) {
	mit := service.NeuesWortmenue(typenliste(), "", ohneSkript(t))
	if !mit.Kaestchen {
		t.Error("eine Liste mit Präfixen braucht Kästchen")
	}

	ohne := service.NeuesWortmenue(
		[]msepaket.Wortlisteneintrag{{Name: "Creature"}}, "", ohneSkript(t))
	if ohne.Kaestchen {
		t.Error("eine Liste ohne Präfixe braucht keine")
	}
}

// **Der Zustand kommt aus dem Text** — `DropDownWordList::selection()`,
// `src/gui/value/text.cpp`, Zeilen 250–274, Fassung f9d9356 vom 2021-07-06.
//
// Ein Präfix ist aktiv, wenn der Text mit ihm **beginnt**; es wird dann
// abgeschnitten, und der Rest geht an den nächsten. Ein gewöhnlicher Eintrag
// ist aktiv, wenn der Rest ihm **gleicht**.
func TestDerZustandKommtAusDemText(t *testing.T) {
	faelle := []struct {
		name     string
		text     string
		aktiv    []string
		gewaehlt string
	}{
		{
			name:     "ein einfacher Typ",
			text:     "Creature",
			aktiv:    []string{"Creature"},
			gewaehlt: "Creature",
		},
		{
			name:     "ein Präfix davor",
			text:     "Legendary Creature",
			aktiv:    []string{"Legendary ", "Creature"},
			gewaehlt: "Creature",
		},
		{
			name:     "zwei Präfixe in der Reihenfolge der Liste",
			text:     "Token Legendary Creature",
			aktiv:    []string{"Token ", "Legendary ", "Creature"},
			gewaehlt: "Creature",
		},
		{
			// **Die Reihenfolge der Liste entscheidet**, nicht die des Textes:
			// `Legendary ` wird zuerst geprüft, wenn es zuerst steht — hier
			// steht `Token ` zuerst und passt nicht, also bleibt der ganze
			// Text stehen und nichts gleicht ihm.
			name:     "Präfixe in der falschen Reihenfolge",
			text:     "Legendary Token Creature",
			aktiv:    []string{"Legendary "},
			gewaehlt: "Legendary ",
		},
		{
			name:     "nichts davon",
			text:     "Artifact",
			aktiv:    nil,
			gewaehlt: "",
		},
		{
			name:     "leer",
			text:     "",
			aktiv:    nil,
			gewaehlt: "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			menue := service.NeuesWortmenue(typenliste(), fall.text, ohneSkript(t))

			var aktiv []string

			for _, eintrag := range menue.Eintraege {
				if eintrag.Aktiv {
					aktiv = append(aktiv, eintrag.Name)
				}
			}

			if strings.Join(aktiv, "|") != strings.Join(fall.aktiv, "|") {
				t.Errorf("aktiv sind %v, erwartet %v", aktiv, fall.aktiv)
			}

			gewaehlt := ""
			if menue.Gewaehlt >= 0 {
				gewaehlt = menue.Eintraege[menue.Gewaehlt].Name
			}

			if gewaehlt != fall.gewaehlt {
				t.Errorf("gewählt ist %q, erwartet %q", gewaehlt, fall.gewaehlt)
			}
		})
	}
}

// **Ein gewöhnlicher Eintrag ersetzt, ein Präfix schaltet um** —
// `DropDownWordList::select`, `src/gui/value/text.cpp`, Zeilen 276–297, Fassung
// f9d9356 vom 2021-07-06.
func TestEineWahlBautDenNeuenText(t *testing.T) {
	faelle := []struct {
		name     string
		text     string
		wahl     string
		erwartet string
	}{
		{
			name:     "ein anderer Typ ersetzt den bisherigen",
			text:     "Creature",
			wahl:     "Instant",
			erwartet: "Instant",
		},
		{
			// **Die Präfixe bleiben**: `toggling_prefix` ist falsch, also
			// zählt für sie allein ihr eigener Zustand.
			name:     "ein anderer Typ behält die Präfixe",
			text:     "Legendary Creature",
			wahl:     "Instant",
			erwartet: "Legendary Instant",
		},
		{
			name:     "ein Präfix kommt dazu",
			text:     "Creature",
			wahl:     "Legendary ",
			erwartet: "Legendary Creature",
		},
		{
			name:     "ein Präfix geht wieder weg",
			text:     "Legendary Creature",
			wahl:     "Legendary ",
			erwartet: "Creature",
		},
		{
			// **Die Reihenfolge ist die der Liste, nicht die des Klickens.**
			// `select` läuft einmal über `items` und hängt an, was aktiv ist.
			name:     "das zweite Präfix rückt an seinen Platz",
			text:     "Legendary Creature",
			wahl:     "Token ",
			erwartet: "Token Legendary Creature",
		},
		{
			name:     "dasselbe noch einmal ändert nichts",
			text:     "Creature",
			wahl:     "Creature",
			erwartet: "Creature",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			menue := service.NeuesWortmenue(typenliste(), fall.text, ohneSkript(t))

			wahl := -1

			for i, eintrag := range menue.Eintraege {
				if eintrag.Name == fall.wahl {
					wahl = i
				}
			}

			if wahl < 0 {
				t.Fatalf("%q steht nicht im Menü", fall.wahl)
			}

			if neu := menue.Waehlen(wahl); neu != fall.erwartet {
				t.Errorf("gewählt wurde %q, daraus wird %q, erwartet %q",
					fall.wahl, neu, fall.erwartet)
			}
		})
	}
}
