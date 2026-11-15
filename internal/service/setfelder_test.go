package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Set nennt seine Angabenfelder** —
// #567.
//
// # Wofür
//
// MSE hat dafür eine eigene Lasche:
//
//	void SetInfoEditor::onChangeSet() {
//	  setStyles(set->stylesheet, set->stylesheet->set_info_style);
//	  setData(set->data);
//	}
//
// `src/gui/control/native_look_editor.cpp`, Zeilen 187–190, Fassung f9d9356
// vom 2021-07-06. Die **Felder** kommen aus dem Spiel
// (`set_info_style.cloneFrom(game->default_set_style)`,
// `src/data/stylesheet.cpp`, Zeile 110), die **Werte** aus `set->data`.
//
// # Ohne Sichtbarkeitsfilter
//
// `NativeLookEditor::resizeViewers` legt **jede** Ansicht untereinander und
// setzt ihren Kasten selbst (Zeilen 48–93). Anders als beim Karteneditor
// entscheidet der Stil hier also nicht, was zu sehen ist — gezeigt wird alles,
// was das Spiel erklärt. Über `magic.mse-game` sind das 89 Felder.
func TestEinSetNenntSeineAngabenfelder(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 1, Game: "probe", Style: "probe-schlicht"},
		}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	felder, err := waehler.Setfelder(t.Context(), 1)
	if err != nil {
		t.Fatalf("Setfelder: %v", err)
	}

	if len(felder) == 0 {
		t.Fatal("kein einziges Setfeld")
	}

	// **Die Felder des Sets, nicht die der Karte.** `probe.mse-game` erklärt
	// beides; wer die falsche Liste nähme, böte im Setfenster `name` und
	// `text` an — die Felder einer Karte.
	nach := map[string]domain.Field{}
	for i := range felder {
		nach[felder[i].Name] = felder[i]
	}

	if _, da := nach["title"]; !da {
		t.Errorf("`title` fehlt; gefunden: %v", namenAus(felder))
	}

	if _, da := nach["name"]; da {
		t.Errorf("`name` ist ein Kartenfeld und steht in der Setliste: %v", namenAus(felder))
	}
}

// **Ein Fehler des Bestands kommt durch** (§8) — er ist etwas anderes als „das
// Spiel erklärt keine Setfelder".
func TestSetfelderMeldenDenBestandsfehler(t *testing.T) {
	kaputt := errors.New("der Bestand schweigt")

	waehler := service.NeuerStilwaehler(
		&koepfequelle{fehler: kaputt},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := waehler.Setfelder(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("%v", err)
	}
}

// **Und ein Spiel, das nicht dasteht, ist `ErrNotFound`.**
//
// Der Fall ist im Betrieb erreichbar: ein Set nennt sein Spiel, und das Paket
// ist nicht installiert.
func TestSetfelderMeldenEinFehlendesSpiel(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "gibtsnicht"}}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := waehler.Setfelder(t.Context(), 1); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("%v", err)
	}
}

// **Ein Set, das es nicht gibt, ist ein Befund.**
func TestSetfelderEinesFremdenSets(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := waehler.Setfelder(t.Context(), 4711); err == nil {
		t.Error("ein fremdes Set kam durch")
	}
}

func namenAus(felder []domain.Field) []string {
	namen := make([]string, 0, len(felder))
	for i := range felder {
		namen = append(namen, felder[i].Name)
	}

	return namen
}
