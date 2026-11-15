package render_test

// **Ein einzeiliges Feld bricht nicht um** — Befund 2 des nachgeholten
// Zweitblicks über die Folgepakete (J05, #2269, zu #2299).
//
// MSE prüft vor jedem Umbruch, ob das Feld mehrzeilig ist: bei
// `multi_line == false` geschieht weder der weiche noch der erzwungene
// Umbruch — `if (!style.field().multi_line) { if (stop_if_too_long) return
// false; }`, und `stop_if_too_long` ist an jedem Aufruf `false`
// (`TextViewer::prepareLinesAtScale`, `src/render/text/viewer.cpp`, Zeilen
// 614–621 und 453–541, Fassung f9d9356 vom 2021-07-06). Die Zeile läuft
// über. **Ein harter Umbruch bricht trotzdem** (ebd. Zeile 590).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinEinzeiligesFeldBrichtNurAmHartenUmbruch(t *testing.T) {
	for _, fall := range []struct {
		name, text string
		einzeilig  bool
		erwartet   []string
	}{
		// Je Zeichen 5 Punkte, acht passen in 40.
		{name: "Wörter laufen über", text: "aaaa bbbb cccc", einzeilig: true, erwartet: []string{"aaaa bbbb cccc"}},
		{name: "ein zu langes Wort bleibt ganz", text: "aaaaaaaaaaaa", einzeilig: true, erwartet: []string{"aaaaaaaaaaaa"}},
		{name: "am harten Umbruch bricht es", text: "aaaa bbbb\ncccc", einzeilig: true, erwartet: []string{"aaaa bbbb", "cccc"}},
		// Die Gegenprobe: mehrzeilig bricht dasselbe um.
		{name: "mehrzeilig bricht um", text: "aaaa bbbb cccc", einzeilig: false, erwartet: []string{"aaaa", "bbbb", "cccc"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			gesetzt := render.SetzeAbschnitte(
				render.Satz{
					Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
					Einzeilig: fall.einzeilig,
				},
				render.Rechteck{Breite: 40, Hoehe: 100},
				[]render.Abschnitt{{Text: fall.text}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

			zeilen := make([]string, 0, len(gesetzt.Zeilen))
			for _, zeile := range gesetzt.Zeilen {
				zeilen = append(zeilen, zeile.Text)
			}

			if !slices.Equal(zeilen, fall.erwartet) {
				t.Errorf("Zeilen %q, erwartet %q", zeilen, fall.erwartet)
			}
		})
	}
}
