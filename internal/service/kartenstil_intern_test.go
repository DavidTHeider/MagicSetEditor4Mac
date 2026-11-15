package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine Karte bringt eigene Stiloptionen mit**
// (#547).
//
//	IndexMap<FieldP,ValueP>& Set::stylingDataFor(const CardP& card) {
//	  if (card && card->has_styling) return card->styling_data;
//	  else                           return stylingDataFor(stylesheetFor(card));
//	}
//
// `src/data/set.cpp`, Zeilen 102–105, Fassung f9d9356 vom 2021-07-06.
//
// # Warum es eine Überlagerung ist und kein Ersetzen
//
// MSE gibt die **ganze** Abbildung der Karte zurück — sie ist aber schon aus
// den `styling_fields` des Stils vorbelegt: `styling_data.init(stylesheet->styling_fields)`
// steht **vor** `REFLECT(styling_data)` (`src/data/card.cpp`, Zeilen 76–90).
// Was die Karte nicht nennt, trägt damit die Vorgabe des Stils.
//
// # Umfang
//
// **Alle 25.148 kanonischen Karten** tragen `has_styling: true`. Wirkung haben
// 2.987 × `overlay`, 718 × `other_options`, 471 × `frames`, 322 × `front_style`,
// 322 × `back_style`, 312 × `dfc_type`, 145 × `legend_crown`.
func TestKartenstilwerte(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		karte    *domain.CardRow
		erwartet map[string]string
	}{
		{
			name: "eine Karte mit eigenen Stiloptionen",
			karte: karteMitWerten(
				"has styling", "true",
				"styling data", "dfc_type: transforming\nfront_style: 3 box saga\n"+
					"snap_chapter_to_box: 1,2,3,"),
			erwartet: map[string]string{
				"dfc_type":            "transforming",
				"front_style":         "3 box saga",
				"snap_chapter_to_box": "1,2,3,",
			},
		},
		{
			// **`has_styling: false` heißt: der Stil gilt.** MSE fragt es
			// ausdrücklich ab, statt einfach zu schauen, ob Daten dastehen.
			name: "ohne has styling gilt nichts davon",
			karte: karteMitWerten(
				"has styling", "false",
				"styling data", "front_style: 3 box saga"),
			erwartet: map[string]string{},
		},
		{
			name:     "eine Karte ganz ohne Stiloptionen",
			karte:    karteMitWerten("name", "Probe"),
			erwartet: map[string]string{},
		},
		{
			// **Ein Wert darf einen Doppelpunkt tragen.** `other_options`
			// führt Listen, und `snap_chapter_to_box: 1,2,3,` endet auf ein
			// Komma — getrennt wird am **ersten** Doppelpunkt.
			name: "ein Wert mit Doppelpunkt",
			karte: karteMitWerten(
				"has styling", "true",
				"styling data", "muster: a: b: c"),
			erwartet: map[string]string{"muster": "a: b: c"},
		},
		{
			// **Eine Zeile ohne Doppelpunkt und eine ohne Namen fallen weg.**
			// Fremde Daten sind der Normalfall, nicht der Ausnahmefall (§8) —
			// ein halb geschriebener Block darf die übrigen Werte nicht
			// mitnehmen.
			name: "unbrauchbare Zeilen fallen weg",
			karte: karteMitWerten(
				"has styling", "true",
				"styling data", "ohne Doppelpunkt\n: ohne Namen\nfront_style: normal"),
			erwartet: map[string]string{"front_style": "normal"},
		},
		{
			name:     "keine Karte",
			karte:    nil,
			erwartet: map[string]string{},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			werte := kartenstilwerte(fall.karte)
			if len(werte) != len(fall.erwartet) {
				t.Fatalf("%d Werte, erwartet %d: %v", len(werte), len(fall.erwartet), werte)
			}

			for name, erwartet := range fall.erwartet {
				if werte[name] != erwartet {
					t.Errorf("%q ist %q, erwartet %q", name, werte[name], erwartet)
				}
			}
		})
	}
}

// karteMitWerten baut eine Karte aus Paaren.
func karteMitWerten(paare ...string) *domain.CardRow {
	karte := &domain.CardRow{}
	for stelle := 0; stelle+1 < len(paare); stelle += 2 {
		karte.Values = append(karte.Values,
			domain.Value{Field: paare[stelle], Text: paare[stelle+1]})
	}

	return karte
}

// TestMitKartenstilLegtDieWerteDerKarteDarueber prüft die Überlagerung.
func TestMitKartenstilLegtDieWerteDerKarteDarueber(t *testing.T) {
	t.Parallel()

	vorgaben := []msescript.Eintrag{
		{Name: "front_style", Wert: msescript.Text("normal")},
		{Name: "overlay", Wert: msescript.Text("")},
		{Name: "text_color", Wert: msescript.Text("black")},
	}

	farben := map[string]farbfeld{
		"text_color": {wahlfarben: []string{"rgb(0,0,0)"}},
	}

	ergebnis := mitKartenstil(vorgaben, map[string]string{
		"front_style": "3 box saga",
		"text_color":  "rgb(255,255,255)",
		// **Ein Name ohne erklärtes `styling field:` fällt weg** — er gehört
		// zu einem anderen Stilpaket.
		"unbekannt": "egal",
	}, farben)

	if len(ergebnis) != 3 {
		t.Fatalf("%d Einträge, erwartet 3", len(ergebnis))
	}

	if ergebnis[0].Wert.AlsText() != "3 box saga" {
		t.Errorf("front_style ist %q, erwartet \"3 box saga\"", ergebnis[0].Wert.AlsText())
	}

	// **Die Vorgabe bleibt, wo die Karte nichts sagt.**
	if ergebnis[1].Wert.AlsText() != "" {
		t.Errorf("overlay ist %q, erwartet leer", ergebnis[1].Wert.AlsText())
	}

	// **Ein Farbfeld trägt eine Farbe, auch wenn die Karte sie nennt.**
	if ergebnis[2].Wert.Art != msescript.ArtFarbe {
		t.Errorf("text_color hat die Art %v, erwartet eine Farbe", ergebnis[2].Wert.Art)
	}
}

// TestMitKartenstilOhneWerteAendertNichts prüft den Normalfall.
//
// **Er ist der häufigste**: eine Karte ohne `has_styling` lässt jede Vorgabe
// stehen.
func TestMitKartenstilOhneWerteAendertNichts(t *testing.T) {
	t.Parallel()

	vorgaben := []msescript.Eintrag{{Name: "a", Wert: msescript.Text("eins")}}

	ergebnis := mitKartenstil(vorgaben, map[string]string{}, nil)
	if len(ergebnis) != 1 || ergebnis[0].Wert.AlsText() != "eins" {
		t.Errorf("Ergebnis %v, erwartet unverändert", ergebnis)
	}
}
