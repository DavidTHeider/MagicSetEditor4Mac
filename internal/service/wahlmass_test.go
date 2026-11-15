package service_test

// **Eine Wahl meldet das Maß ihres vorbereiteten Bilds**
// (#2303).
//
// `prepare_choice_viewer` erzeugt das Stilbild für den Wert und übernimmt die
// Größe, die `generateCached` meldet (`src/render/value/choice.cpp`, Zeilen
// 35–58); wie, sagt `get_options` (Zeilen 111–124): bei `render style:
// image` der Kasten in ganzen Punkten, eingepasst, wenn die Ausrichtung
// nicht `stretch` sagt, bei `both` die Originalgröße. Ohne Bildanteil
// geschieht nichts, und es bleibt `content_width(0.0)`
// (`src/data/field/choice.cpp`, Zeile 168). Fassung f9d9356 vom 2021-07-06.
//
// **Die Zahlen sind gefragt**, am 2026-10-11 über einen Orakelstil mit
// `rarity` in 44 × 22 und `alignment: middle right`: eine Datei von 30 × 10
// meldet 44 × 14, ein Wert ohne Wahlbild und der leere Wert 44 × 22; mit
// `both` 30 × 10 und für den leeren Wert 1 × 1.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineWahlMeldetDasMassIhresBilds(t *testing.T) {
	quelle := massquelle{"/probe.mse-style/breit.png": {X: 30, Y: 10}}

	for _, fall := range []struct {
		name, darstellung, bild, wert string
		art                           domain.FieldKind
		ohneWahlbilder                bool
		breite, hoehe                 float64
	}{
		{name: "eingepasst", darstellung: "image", wert: "common", breite: 44, hoehe: 14},
		{name: "ein Wert ohne Wahlbild", darstellung: "image", wert: "special", breite: 44, hoehe: 22},
		{name: "der leere Wert", darstellung: "image", wert: "", breite: 44, hoehe: 22},
		{name: "mit Text in Originalgröße", darstellung: "both", wert: "common", breite: 30, hoehe: 10},
		{name: "mit Text und leer", darstellung: "both", wert: "", breite: 1, hoehe: 1},
		// **Ohne Bildanteil kein Maß** — auch nicht das des Texts.
		{name: "nur Text", darstellung: "text", wert: "common", breite: 0, hoehe: 0},
		// **Ein `image:` ohne Skript wird nie neu gerechnet**: `update`
		// meldet für ein festes Bild keine Änderung (`src/script/image.cpp`,
		// Zeilen 44–46 und 67–78, Fassung f9d9356 vom 2021-07-06).
		{name: "ein festes Bild", darstellung: "image", bild: "breit.png", wert: "common", breite: 0, hoehe: 0},
		{name: "ein gerechnetes Bild", darstellung: "image", bild: `{ "breit.png" }`, wert: "common", breite: 44, hoehe: 14},
		// **Ein Ja/Nein-Feld ist eine Wahl** (`BooleanStyle : ChoiceStyle`).
		{name: "ein Ja/Nein-Feld", darstellung: "image", wert: "yes", art: domain.KindBoolean, breite: 44, hoehe: 22},
		// **Eine Mehrfachwahl als Liste misst nichts** —
		// `MultipleChoiceValueViewer::prepare`, `src/render/value/multiple_choice.cpp`,
		// Zeilen 19–22, Fassung f9d9356 vom 2021-07-06.
		{name: "eine Mehrfachwahl als Liste", darstellung: "image list", wert: "common", art: domain.KindMultipleChoice, breite: 0, hoehe: 0},
		{name: "eine Mehrfachwahl als Liste mit Text", darstellung: "both list", wert: "common", art: domain.KindMultipleChoice, breite: 0, hoehe: 0},
		// **Eine einfache Wahl kennt keine Liste** — nur der Betrachter der
		// Mehrfachwahl fragt danach; sie bereitet ihr Bild vor wie bei `both`.
		{name: "eine einfache Wahl als Liste", darstellung: "both list", wert: "common", breite: 30, hoehe: 10},
		// **Ohne `image:` und ohne Wahlbilder gibt es nichts zu rechnen**
		// (`ChoiceStyle::initImage`, Zeile 172).
		{name: "ohne Bild und Wahlbilder", darstellung: "image", wert: "common", ohneWahlbilder: true, breite: 0, hoehe: 0},
	} {
		t.Run(fall.name, func(t *testing.T) {
			art := fall.art
			if art == "" {
				art = domain.KindChoice
			}

			wahlbilder := []msepaket.Wahlbild{{Wahl: "common", Bild: "breit.png"}}
			if fall.ohneWahlbilder {
				wahlbilder = nil
			}

			breite, hoehe := wahlmass(t, quelle, art, fall.darstellung, fall.bild, fall.wert, wahlbilder)
			if breite != fall.breite || hoehe != fall.hoehe {
				t.Errorf("%.0f × %.0f, erwartet %.0f × %.0f", breite, hoehe, fall.breite, fall.hoehe)
			}
		})
	}
}

// wahlmass setzt eine Wahl in einen Kasten von 44 × 22 und liest ihr Maß an
// der Lage zweier Felder ab.
func wahlmass(
	t *testing.T, quelle render.Bildquelle, art domain.FieldKind, darstellung, bild, wert string,
	wahlbilder []msepaket.Wahlbild,
) (breite, hoehe float64) {
	t.Helper()

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{
			{
				Feld: "rarity", Links: "100", Oben: "40", Breite: "44", Hoehe: "22",
				Ausrichtung: "middle right", Darstellung: darstellung, Bild: bild,
				Schrift:    msepaket.Schrift{Name: "Probe", Groesse: "10"},
				Wahlbilder: wahlbilder,
			},
			{
				Feld: "breite", Oben: "100", Hoehe: "5", Breite: "5",
				Links:   "{ 10 + card_style.rarity.content_width }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				Feld: "hoehe", Oben: "120", Hoehe: "5", Breite: "5",
				Links:   "{ 10 + card_style.rarity.content_height }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "rarity", Text: wert}, {Field: "breite", Text: "b"}, {Field: "hoehe", Text: "h"},
	}}

	plan, _, err := service.NeuerPlanbauer().MitBildern(quelle).
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{
			Stil: stil,
			Kartenfelder: []domain.Field{
				{Name: "rarity", Kind: art, Choices: []string{"common", "special"}},
				{Name: "breite", Kind: domain.KindText},
				{Name: "hoehe", Kind: domain.KindText},
			},
		})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	lage := map[string]float64{}
	for _, ebene := range plan.Ebenen {
		lage[ebene.Klartext()] = ebene.Rahmen.Links - 10
	}

	return lage["b"], lage["h"]
}
