//go:build fremdpakete

package service_test

// **Die Dimension `keywords` über die ganze Kette** —
// #1354.
//
// # Warum das Orakel hier nicht antwortet
//
// Das Issue verlangt „dieselbe Liste wie MSE2, gefragt über
// `magicseteditor --cli`". **Das geht nicht**, und der Grund ist ein anderer
// als bei #1355:
// `card->keyword_usage` wird als **Nebenwirkung** von `expand_keywords`
// gefüllt, und das läuft nur, wenn MSE die Feldwerte einer Karte
// aktualisiert. Die Kommandozeile tut das nicht.
//
// **Gemessen am 2026-08-31** über `White.mse-set`:
//
//	n := length(set.cards)-1
//	for x from 0 to n do [ keyword_usage(card: set.cards[x], unique: true) ]
//	→ 3761 Karten, davon 0 mit Schlüsselwörtern
//
// # Und die kanonischen Sets helfen auch nicht
//
// **Gemessen am selben Tag**: in `White`, `Green`, `Blue` und `Colorless`
// beginnt **kein einziger** Regeltext mit einem der elf Kernschlüsselwörter.
// Sie kommen nur mitten im Satz vor („gains flying"), und dort weist die
// M15-Vorlage sie über `correct_case` ab — MSE täte dasselbe.
//
// # Was stattdessen geprüft wird
//
// **Der Vermerk, an der echten Wortliste aus `magic.mse-game`**: 368
// Schlüsselwörter, echte Parametertypen, echtes Spielskript. Was hier nicht
// steht, ist ein Vergleich gegen MSE2 — und dieser Absatz steht da, damit
// niemand mehr hineinliest (§10).
//
// # Und über echte Karten
//
// **Gemessen am 2026-08-31** über `White.mse-set`, nach der Behebung von
// #1363:
// **292 von 3.761 Karten** tragen Schlüsselwörter — `Equip`, `Vigilance`,
// `Outlast`, `Indestructible`, `Eternalize`.
//
// **Vorher waren es null.** `text_filter` fror die Schlüsselwortfunktion beim
// Initialisieren des Spiels ein, und die echte stand erst danach da; die
// Kette rief für immer den Stummel. Das war Befund 5 des Zweitblicks über
// Phase 4.11 — und es betraf nicht nur die Statistik, sondern **jeden
// Erinnerungstext auf jeder Karte**.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// TestDieDimensionKeywordsNenntDieBenutztenWoerter ist die Abnahme.
func TestDieDimensionKeywordsNenntDieBenutztenWoerter(t *testing.T) {
	paket, fehler := msepaket.Oeffne(filepath.Join(packort, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(
		msepaket.NeuerPaketleser(packort)); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	spiel, _ := paket.Spielmodell()

	if anzahl := len(paket.Schluesselwoerter()); anzahl < 300 {
		t.Fatalf("%d Schlüsselwörter — die Wortliste ist nicht die echte", anzahl)
	}

	vorlage := mitWortliste(t, magicvorlage(spiel), paket)

	for _, fall := range []struct {
		name     string
		text     string
		erwartet string
	}{
		// **Ohne Parameter** — `match: Flying`.
		{"ein Kernschlüsselwort", "Flying", "Flying"},

		// **Mit Parameter** — `match: Equip <atom-param>cost</atom-param>`,
		// und der Parametertyp `cost` verlangt ein Manasymbol. „Equip 3" ohne
		// `<sym>` trifft **nicht**, und das ist richtig: der erste Anlauf
		// hier prüfte genau das und hielt das Nein für einen Fehler.
		{"eines mit Parameter", "Equip <sym>3</sym>", "Equip"},

		// **Zwei in einem Text** — in der Reihenfolge, in der sie stehen.
		{"zwei hintereinander", "Flying\nTrample", "Flying, Trample"},

		// **Und keines** — ein Wert und kein Fehler.
		{"ohne Schlüsselwort", "Draw a card.", ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			karte := &domain.CardRow{
				ID: 1,
				Values: []domain.Value{
					{Field: "name", Text: "Prüfkarte"},
					{Field: "rule text", Text: fall.text},
				},
			}

			// **Das Skript setzt selbst ein und fragt dann.**
			//
			// So läuft es bei MSE: `keyword_usage` liest, was
			// `expand_keywords` hinterlassen hat. Über das Wertskript der
			// Karte käme heute nichts an — siehe den Kopf dieser Datei.
			werte, befunde := service.Auszaehlwerte(
				t.Context(), []*domain.CardRow{karte}, vorlage, nil,
				service.Auszaehlangabe{
					Name: "keywords",
					// **Nur fragen, nicht selbst einsetzen.**
					//
					// Der erste Anlauf setzte im Skript selbst ein, weil über
					// das Wertskript der Karte nichts ankam. Seit Befund 5 des
					// Zweitblicks über Phase 4.11 kommt es an: `text_filter`
					// fror die Funktion beim Initialisieren ein, und die
					// echte stand erst danach da.
					Skript:      `keyword_usage(unique: true)`,
					LeereZeigen: true,
				})

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %+v", befunde)
			}

			if len(werte) != 1 {
				t.Fatalf("%d Werte", len(werte))
			}

			if werte[0] != fall.erwartet {
				t.Errorf("%q, erwartet %q", werte[0], fall.erwartet)
			}
		})
	}
}

// **Und die Dimension des Pakets rechnet, ohne zu scheitern.**
//
// Ihr Skript ist nicht bloß `keyword_usage(unique: true)`, sondern
//
//	if filter_from_statistic() == "Filtered" or trim_from_statistic() then ""
//	else keyword_usage(unique: true)
//
// — es liest also zwei weitere Funktionen aus `statistics_script`. **Dass sie
// heute leer antwortet, ist der Befund im Kopf dieser Datei** und keine
// Eigenschaft dieses Laufs: geprüft wird, dass sie **ohne Fehler** rechnet.
func TestDieEchteDimensionKeywordsRechnet(t *testing.T) {
	paket, fehler := msepaket.Oeffne(filepath.Join(packort, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(
		msepaket.NeuerPaketleser(packort)); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	spiel, _ := paket.Spielmodell()

	var dimension service.Auszaehlangabe

	for _, eine := range service.Auszaehlangaben(paket.Auszaehldimensionen()) {
		if eine.Name == "keywords" {
			dimension = eine
		}
	}

	if !strings.Contains(dimension.Skript, "keyword_usage") {
		t.Fatalf("das Skript der Dimension ist %q", dimension.Skript)
	}

	dimension.LeereZeigen = true
	dimension.Zerlegen = false

	karte := &domain.CardRow{
		ID: 1,
		Values: []domain.Value{
			{Field: "name", Text: "Prüfkarte"},
			{Field: "rule text", Text: "Flying"},
		},
	}

	_, befunde := service.Auszaehlwerte(
		t.Context(), []*domain.CardRow{karte},
		mitWortliste(t, magicvorlage(spiel), paket), nil, dimension)

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}
}
