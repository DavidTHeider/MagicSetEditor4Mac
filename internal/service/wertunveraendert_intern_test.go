package service

// **Ein Wertskript, das nichts ändert, baut `card` nicht neu** — PERF-03
// (#2170).
//
// Gemessen am 2026-10-07 über die Ausfuhr von `Archenemy.mse-set` (70 Karten,
// `magic-simple-exporter`): 1,8 GB Zuteilung je Ausfuhr, und 728 MB davon
// über vier Läufe allein im Rückruf, der nach einem Wertskript `card` neu
// baut (`docs/LEISTUNGSSCHNITTE.md`). Hat das Skript den Wert gelassen, wie
// er war, steht in `card` schon, was ein späteres Skript lesen kann.

import (
	"context"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEinUnveraenderterWertBautCardNichtNeu(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, skript string
		neubauten    int
	}{
		{"das Skript lässt den Wert", "value", 0},
		{"das Skript ändert den Wert", `value + "!"`, 1},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			felder := []domain.Field{
				{Name: "eins", Skript: fall.skript},
				// **`zwei` liest `eins`** — erst dadurch ist `eins` ein Feld,
				// nach dem `card` neu gebaut wird.
				{Name: "zwei", Skript: `card.eins + "?"`},
			}
			gefuellt := []msescript.Eintrag{
				{Name: "eins", Wert: msescript.Text("alt")},
				{Name: "zwei", Wert: msescript.Text("")},
			}

			umgebung := msescript.Bibliothek().Darueber()
			umgebung.Setze("card", msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: gefuellt})

			neubauten := 0
			ergebnis, befunde := mitWertskripten(context.Background(), gefuellt, felder, umgebung, nil,
				func(stand []msescript.Eintrag) {
					neubauten++
					umgebung.Setze("card", msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: append([]msescript.Eintrag(nil), stand...)})
				})

			if len(befunde) != 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if neubauten != fall.neubauten {
				t.Errorf("%d Neubauten, erwartet %d", neubauten, fall.neubauten)
			}

			// **Und das spätere Skript sah den richtigen Wert** — gleich ob
			// neu gebaut wurde oder nicht.
			if zwei := ergebnis[1].Wert.AlsText(); zwei != ergebnis[0].Wert.AlsText()+"?" {
				t.Errorf("zwei ist %q, eins %q", zwei, ergebnis[0].Wert.AlsText())
			}
		})
	}
}

// **Die Gleichheit kennt jedes Feld von `Wert`** — bekommt er ein weiteres,
// wird dieser Test rot, statt dass ein geänderter Wert still als unverändert
// gilt und `card` alt bliebe.
func TestUnveraendertKenntJedesFeldVonWert(t *testing.T) {
	t.Parallel()

	if felder := reflect.TypeFor[msescript.Wert]().NumField(); felder != wertfelder {
		t.Errorf("Wert hat %d Felder, wertBleibt prüft %d — jedes neue gehört in die Gleichheit", felder, wertfelder)
	}

	basis := msescript.Text("x")
	for name, anders := range map[string]msescript.Wert{
		"Art":       {Art: msescript.ArtZahlwert, Text: "x"},
		"Zahl":      {Art: basis.Art, Text: "x", Zahl: 1},
		"Text":      msescript.Text("y"),
		"Wahrheit":  {Art: basis.Art, Text: "x", Wahrheit: true},
		"Feld":      {Art: basis.Art, Text: "x", Feld: "f"},
		"Eintraege": {Art: basis.Art, Text: "x", Eintraege: []msescript.Eintrag{{Name: "e"}}},
		"Glieder":   {Art: basis.Art, Text: "x", Glieder: []msescript.Eintrag{{Name: "g"}}},
		"Verbund":   {Art: basis.Art, Text: "x", Verbund: true},
		"Kennung":   {Art: basis.Art, Text: "x", Kennung: "k"},
		"Funktion":  {Art: basis.Art, Text: "x", Funktion: &msescript.Funktionsteil{}},
		"Marke":     {Art: basis.Art, Text: "x", Marke: "m"},
	} {
		if wertBleibt(basis, anders) {
			t.Errorf("ein anderes %s gilt als unverändert", name)
		}
	}

	if !wertBleibt(basis, msescript.Text("x")) {
		t.Error("derselbe Text gilt als verändert")
	}
}
