package service

// **Gegenstandsidentität außerhalb des Auszählungswegs** —
// #1366.
//
// # Was MSE tut
//
// `SetScriptContext::getContext` bindet `set`, `stylesheet`, `styling`,
// `card_style` und `card` als **Gegenstände** (`src/script/script_manager.cpp`,
// Zeilen 30–60, Fassung `f9d9356` vom 2021-07-06), und ihr Vergleich läuft
// über `COMPARE_AS_POINTER` (`src/script/value.cpp`, Zeilen 119–129).
//
// # Was das Orakel dazu sagt
//
// Am 2026-09-01 über `White.mse-set` mit
// `magicseteditor --cli … -q --raw` gemessen:
//
//	set.cards.0 == set.cards.0   → true
//	set.cards.0 == set.cards.1   → false
//	set == set                   → true
//	stylesheet == stylesheet     → true
//	styling == styling           → true
//	set == stylesheet            → false
//
// **Ein interner Lauf**, weil die Umgebungsbauer paketintern sind.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Zwei strukturgleiche Karten sind in der Listenumgebung verschieden.**
//
// Der Vergleich läuft über die Kennung, nicht über die Felder — sonst gäbe
// jeder Ausdruck der Art `card != set.cards.0` für die zweite Karte dasselbe
// wie für die erste.
func TestZweiGleicheKartenSindInDerListeVerschieden(t *testing.T) {
	felder := []domain.Field{{Name: "name"}}

	eine := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: "Gleich"}}}
	andere := &domain.CardRow{ID: 2, Values: []domain.Value{{Field: "name", Text: "Gleich"}}}

	spiel := msescript.NeueUmgebung(nil)

	ersteKarte, da := listenumgebung(spiel, eine, felder).Suche("card")
	if !da {
		t.Fatal("die Listenumgebung kennt kein card")
	}

	zweiteKarte, _ := listenumgebung(spiel, andere, felder).Suche("card")

	pruefeGegenstaende(t, ersteKarte, zweiteKarte)
}

// **Und in den Exportdaten ebenso.**
func TestZweiGleicheKartenSindImExportVerschieden(t *testing.T) {
	felder := []domain.Field{{Name: "name"}}

	karten := []*domain.CardRow{
		{ID: 1, Values: []domain.Value{{Field: "name", Text: "Gleich"}}},
		{ID: 2, Values: []domain.Value{{Field: "name", Text: "Gleich"}}},
	}

	liste := kartenliste(karten, felder)

	if len(liste.Eintraege) != 2 {
		t.Fatalf("%d Einträge, erwartet 2", len(liste.Eintraege))
	}

	pruefeGegenstaende(t, liste.Eintraege[0].Wert, liste.Eintraege[1].Wert)
}

// pruefeGegenstaende sichert zu, was MSE für zwei Gegenstände zusagt.
func pruefeGegenstaende(t *testing.T, eine, andere msescript.Wert) {
	t.Helper()

	if eine.Kennung == "" {
		t.Fatal("der erste Gegenstand trägt keine Kennung — dann entscheidet die Gestalt")
	}

	if eine.Kennung == andere.Kennung {
		t.Errorf("beide tragen %q — zwei verschiedene Gegenstände wären gleich",
			eine.Kennung)
	}

	// **Und derselbe Gegenstand bleibt gleich**, auch nachdem sich sein
	// Inhalt geändert hat — die andere Hälfte von `COMPARE_AS_POINTER`.
	//
	// **Durch den Auswerter und nicht über die Kennung.** Der erste Anlauf
	// verglich `geaendert.Kennung != eine.Kennung`, und das trifft zu, bevor
	// irgendetwas geschieht: eine Zusicherung, die schon vor der Antwort
	// gilt (§3). `govet` hat sie als toten Schreibvorgang angestrichen, und
	// das war sie auch.
	geaendert := eine
	geaendert.Eintraege = nil

	umgebung := msescript.NeueUmgebung(map[string]msescript.Wert{
		"vorher":  eine,
		"nachher": geaendert,
	})

	wert, fehler := msescript.AuswertenQuelle(t.Context(), "vorher == nachher", umgebung)
	if fehler != nil {
		t.Fatalf("auswerten: %v", fehler)
	}

	if wert.AlsText() != "true" {
		t.Errorf("vorher == nachher ergibt %q — eine Inhaltsänderung hat den "+
			"Gegenstand zu einem anderen gemacht", wert.AlsText())
	}
}

// **`set` und `stylesheet` sind zwei verschiedene Gegenstände.**
//
// # Zwei, obwohl das Issue vier nennt
//
// #1366
// führt `set`, `stylesheet`, `styling` und `card_style`. Nachgelesen am
// 2026-09-01 bleiben zwei:
//
//   - **`styling` ist keiner.** `to_script(&set.stylingDataFor(*stylesheet))`
//     gibt ein `ScriptMap` mit `type() == SCRIPT_COLLECTION`
//     (`src/script/to_value.hpp`, Zeilen 169–190, Fassung `f9d9356` vom
//     2021-07-06), und `equal` vergleicht zwei
//     Sammlungen **gliedweise**, bevor es überhaupt zu `compareAs` kommt
//     (`value.cpp`, Zeilen 107–117). Eine Kennung dort machte etwas anderes
//     als MSE.
//   - **`card_style` steht in dieser Umgebung gar nicht** — aus einem eigenen,
//     dokumentierten Grund: `card_style.<feld>.content_width` verlangt eine
//     Messung, die erst `internal/render` machen kann
//     (#195).
//     Eine Kennung für einen Namen zu setzen, den niemand binden kann, wäre
//     Buchführung statt Wirkung.
func TestDieStilbindungenSindVerschiedeneGegenstaende(t *testing.T) {
	umgebung := msescript.NeueUmgebung(nil)

	belegeVorlage(t.Context(), umgebung, Vorlage{
		Stil: msepaket.Stil{Paket: "magic-m15.mse-style"},
	}, nil, nil)

	kennungen := map[string]string{}

	for _, name := range []string{"set", "stylesheet"} {
		wert, da := umgebung.Suche(name)
		if !da {
			t.Fatalf("die Umgebung kennt kein %q", name)
		}

		if wert.Kennung == "" {
			t.Errorf("%q trägt keine Kennung — dann entscheidet die Gestalt", name)

			continue
		}

		if voriger, doppelt := kennungen[wert.Kennung]; doppelt {
			t.Errorf("%q und %q tragen beide %q", voriger, name, wert.Kennung)
		}

		kennungen[wert.Kennung] = name
	}
}

// **Zwei Stile geben zwei verschiedene `stylesheet`.**
//
// Ohne diesen Lauf genügte eine feste Zeichenkette — und die wäre für jeden
// Stil dieselbe.
func TestZweiStileGebenZweiVerschiedeneStylesheets(t *testing.T) {
	kennungen := make([]string, 0, 2)

	for _, paket := range []string{"magic-m15.mse-style", "magic-m15-saga.mse-style"} {
		umgebung := msescript.NeueUmgebung(nil)

		belegeVorlage(t.Context(), umgebung, Vorlage{
			Stil: msepaket.Stil{Paket: paket},
		}, nil, nil)

		wert, _ := umgebung.Suche("stylesheet")
		kennungen = append(kennungen, wert.Kennung)
	}

	if kennungen[0] == kennungen[1] {
		t.Errorf("beide Stile tragen %q", kennungen[0])
	}
}

// **Und jetzt das Verhalten, nicht die Buchführung.**
//
// Die Läufe darüber prüfen **Kennungen**. Ob daraus ein Vergleich wird, den
// MSE genauso beantwortet, sagt erst der Auswerter — §10: „Die Reichweite
// einer Messung ist die Reichweite ihrer Aussage."
//
// Die erwarteten Werte kommen vom Orakel (siehe Dateikopf).
func TestDerVergleichVerhaeltSichWieBeiMSE(t *testing.T) {
	felder := []domain.Field{{Name: "name"}}

	eine := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: "Gleich"}}}
	andere := &domain.CardRow{ID: 2, Values: []domain.Value{{Field: "name", Text: "Gleich"}}}

	spiel := msescript.NeueUmgebung(nil)

	// **Die Listenumgebung der einen Karte, und die andere daneben** — so
	// begegnen sie einander in `card != set.cards.0`.
	umgebung := listenumgebung(spiel, eine, felder)

	zweite, _ := listenumgebung(spiel, andere, felder).Suche("card")
	umgebung.Setze("andere", zweite)

	belegeVorlage(t.Context(), umgebung, Vorlage{
		Stil: msepaket.Stil{Paket: "magic-m15.mse-style"},
	}, nil, eine)

	for _, fall := range []struct {
		quelle   string
		erwartet string
	}{
		{"card == card", "true"},
		{"card == andere", "false"},
		{"card != andere", "true"},
		{"set == set", "true"},
		{"stylesheet == stylesheet", "true"},
		{"styling == styling", "true"},
		{"set == stylesheet", "false"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			wert, fehler := msescript.AuswertenQuelle(t.Context(), fall.quelle, umgebung)
			if fehler != nil {
				t.Fatalf("auswerten: %v", fehler)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%s ergibt %q, MSE gibt %q",
					fall.quelle, wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Und der Fall, der die Kennung wirklich verlangt: der Nachbau.**
//
// # Warum die Läufe darüber nicht genügen
//
// `set == stylesheet` ist auch **ohne** Kennung `false`. Die Gegenprobe blieb
// deshalb grün, als die Stilkennungen entfernt wurden — der Lauf enthielt
// seinen Fall nicht (§10, „eine Gegenprobe, die nichts ändert, ist keine").
//
// # Und warum kein Vergleich der drei untereinander ihn enthält
//
// **Gemessen am 2026-09-01**: die drei sind paarweise **strukturverschieden,
// von Bauart wegen**. `set` trägt immer `set_info`, und zwar als geschachtelte
// Abbildung; `stylesheet` trägt immer vier Einträge mehr als `styling`
// (`card_width`, `card_height`, `short_name`, `full_name`). Auch ein Stil mit
// einer Option namens `set_info` macht sie nicht gleich: dort ist es eine
// Zeichenkette, im Set eine Abbildung.
//
// Das ist keine Zusage, sondern ein Zustand — er hängt daran, dass
// `setangaben` `set_info` führt. Ginge das weg, glichen sich zwei Bindungen
// **still**.
//
// # Der Fall, der ihn enthält
//
// Ein **Nachbau**: dieselbe Gestalt, anderer Gegenstand. MSE sagt dazu
// „verschieden", denn `COMPARE_AS_POINTER` fragt nach dem Gegenstand und nicht
// nach dem Inhalt (`src/script/value.cpp`, Zeilen 119–129, Fassung `f9d9356`
// vom 2021-07-06). Ohne Kennung entschiede hier die Gestalt und sagte
// „gleich".
func TestEinNachbauIstNichtDerGegenstand(t *testing.T) {
	for _, name := range []string{"set", "stylesheet"} {
		t.Run(name, func(t *testing.T) {
			umgebung := msescript.NeueUmgebung(nil)

			belegeVorlage(t.Context(), umgebung, Vorlage{
				Stil: msepaket.Stil{
					Paket: "leer.mse-style",
					Stiloptionen: []domain.StyleField{
						{Name: "rand", Kind: domain.KindText},
					},
				},
			}, nil, nil)

			gegenstand, da := umgebung.Suche(name)
			if !da {
				t.Fatalf("die Umgebung kennt kein %q", name)
			}

			// **Zeichengleich bis auf die Kennung** — genau das, was ein
			// Vergleich über die Gestalt nicht auseinanderhalten kann.
			nachbau := gegenstand
			nachbau.Kennung = ""

			umgebung.Setze("nachbau", nachbau)

			wert, fehler := msescript.AuswertenQuelle(
				t.Context(), name+" == nachbau", umgebung)
			if fehler != nil {
				t.Fatalf("auswerten: %v", fehler)
			}

			if wert.AlsText() != "false" {
				t.Errorf("%s == nachbau ergibt %q — ein Nachbau galt als der "+
					"Gegenstand, MSE gibt \"false\"", name, wert.AlsText())
			}
		})
	}
}
