package render_test

// **Umbruch und Stauchung sind zwei Schritte** — #2308, Folge aus dem
// Nachlauf über J05 (#2269), Befund 6.
//
// Bei MSE2 hängt der Umbruch allein an `multi line` und am harten Umbruch
// (`TextViewer::prepareLinesAtScale`, `src/render/text/viewer.cpp`, Zeilen
// 590–618); die Stauchung kommt danach und rechnet aus der Breite des
// fertigen Layouts (`TextStyle::getStretch`, `src/data/field/text.cpp`,
// Zeilen 73–81) — Fassung f9d9356 vom 2021-07-06. Hier ersetzte ein
// gestauchter Satz bis dahin den gebrochenen durch **eine** Zeile: ein
// mehrzeiliges Feld mit `shrink` brach gar nicht um (22 Feldstile in 17
// `magic-*`-Stilen des Pack — Embleme, Zählerbögen, Battle), und ein harter
// Umbruch in einem gestauchten Feld wurde zusammengefügt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinMehrzeiligesGestauchtesFeldBrichtUm(t *testing.T) {
	schrift := gleichbreit{breite: 0.5}
	satz := render.Satz{Groesse: 16, Stauchen: true, NurBeiUeberlauf: true}

	ergebnis := render.Setze(satz, render.Rechteck{Breite: 40, Hoehe: 60}, "abcde fghij", schrift)

	if len(ergebnis.Zeilen) < 2 {
		t.Errorf("%d Zeilen, erwartet mindestens zwei", len(ergebnis.Zeilen))
	}

	// **Nach dem Umbruch passt jede Zeile** — gestaucht wird nicht.
	if ergebnis.Stauchung != 1 {
		t.Errorf("die Stauchung ist %g, erwartet 1", ergebnis.Stauchung)
	}
}

func TestEinEinzeiligesGestauchtesFeldBrichtAmHartenUmbruch(t *testing.T) {
	schrift := gleichbreit{breite: 0.5}
	satz := render.Satz{Groesse: 16, Stauchen: true, NurBeiUeberlauf: true, Einzeilig: true}

	// „abcdefghij" ist 80 breit, „ab" 16 — in einem Feld von 40.
	ergebnis := render.Setze(satz, render.Rechteck{Breite: 40, Hoehe: 60}, "abcdefghij\nab", schrift)

	if len(ergebnis.Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet zwei — eine je harten Umbruch", len(ergebnis.Zeilen))
	}

	// **Der Faktor kommt aus der breitesten Zeile**, nicht aus dem ganzen Text
	// in einer: 40 / 80, nicht 40 / 96.
	if ergebnis.Stauchung != 0.5 {
		t.Errorf("die Stauchung ist %g, erwartet 0,5", ergebnis.Stauchung)
	}
}

// **Gestaucht wird nach der harten Breite** — ein weicher Schwanz zählt nicht
// mit, wie bei MSE2 (`Line::width` bis `end_or_soft`, `src/render/text/
// viewer.cpp`, Zeile 35, Fassung f9d9356 vom 2021-07-06; Zweitblick über die
// Folgen vor 0.8.0, Befund 10).
func TestEinWeicherSchwanzLoestKeineStauchungAus(t *testing.T) {
	satz := render.Satz{Groesse: 16, Stauchen: true, NurBeiUeberlauf: true, Einzeilig: true}

	// „abcde" ist 40 breit und passt genau; „xy" ist weich.
	ergebnis := render.SetzeAbschnitte(satz, render.Rechteck{Breite: 40, Hoehe: 60},
		[]render.Abschnitt{{Text: "abcde"}, {Text: "xy", Weich: true}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if ergebnis.Stauchung != 1 {
		t.Errorf("die Stauchung ist %g, erwartet 1 — der weiche Schwanz zählte mit", ergebnis.Stauchung)
	}
}

// **Ob eine Zeile füllt, entscheidet jede für sich** — `should_fill` in
// `Line::alignHorizontal` (`src/render/text/viewer.cpp`, Zeilen 834–869,
// Fassung f9d9356 vom 2021-07-06; Zweitblick über die Folgen vor 0.8.0,
// Befund 11). Mit `shrink-overflow` beginnt nur die Zeile links, die
// überläuft; eine kurze bleibt in ihrer Ausrichtung.
func TestNurDieUeberlaufendeZeileBeginntLinks(t *testing.T) {
	satz := render.Satz{
		Groesse: 16, Stauchen: true, NurBeiUeberlauf: true, Einzeilig: true,
		Ausrichtung: render.MitteOben,
	}

	// „abcdefghij" ist 80 breit und läuft über, „ab" ist 16 breit.
	ergebnis := render.Setze(satz, render.Rechteck{Breite: 40, Hoehe: 60}, "abcdefghij\nab", gleichbreit{breite: 0.5})

	if len(ergebnis.Zeilen) != 2 {
		t.Fatalf("%d Zeilen", len(ergebnis.Zeilen))
	}

	if ergebnis.Zeilen[0].Links != 0 {
		t.Errorf("die überlaufende Zeile beginnt bei %g, erwartet 0", ergebnis.Zeilen[0].Links)
	}

	if ergebnis.Zeilen[1].Links != 12 {
		t.Errorf("die kurze Zeile beginnt bei %g, erwartet 12 — mittig in 40", ergebnis.Zeilen[1].Links)
	}
}
