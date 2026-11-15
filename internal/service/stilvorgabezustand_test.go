package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Stiloption weiß, ob ihr Wert gesetzt ist oder aus der Vorgabe kommt**
// (#764).
//
// # Warum das eine eigene Angabe braucht
//
// Bis hierher trug eine Option **einen** Wert, und „ausdrücklich gesetzt" und
// „nichts gesetzt" waren dasselbe. MSE unterscheidet beides: der Wert ist ein
// `Defaultable`, und `change(Defaultable<T>&, …)` ruft
// `assignDontChangeDefault` (`src/script/scriptable.hpp`, Zeile 40, Fassung
// f9d9356 vom 2021-07-06) — die Markierung überlebt sogar das Vorgabeskript.
//
// Daran hängen zwei Dinge, die es ohne sie nicht geben kann: der
// Vorgabeeintrag im Farbeditor und die Handlung *zurück auf die Vorgabe*.
func TestEineStiloptionKenntIhrenVorgabezustand(t *testing.T) {
	t.Parallel()

	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		&stilablage{werte: []domain.Value{{Field: "hintergrund", Text: "dunkel"}}},
		&feldliste{})

	optionen, err := waehle.Optionen(t.Context(), 1)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	if len(optionen) != 3 {
		t.Fatalf("%d Optionen, erwartet drei", len(optionen))
	}

	// `hintergrund` steht im Bestand, `rahmen zeigen` nicht.
	if !optionen[0].Gesetzt {
		t.Errorf("%q gilt als Vorgabe, obwohl ein Wert im Bestand steht", optionen[0].Name)
	}

	if optionen[1].Gesetzt {
		t.Errorf("%q gilt als gesetzt, obwohl nichts im Bestand steht", optionen[1].Name)
	}
}

// **Und sie lässt sich zurückstellen.**
//
//	void DropDownColorList::select(size_t item) {
//	  if (isDefault(item)) {
//	    cve.change( Defaultable<Color>());
//
// `src/gui/value/color.cpp`, Zeilen 118–120: die Wahl des Vorgabeeintrags
// setzt einen **leeren** `Defaultable` ein — der gespeicherte Wert
// verschwindet, und es gilt wieder, was die Vorlage rechnet.
//
// **Hier wird nur entfernt, nicht entschieden.** Ob eine Option den Eintrag
// überhaupt anbietet, hängt bei MSE an `hasDefault()` und damit am
// `default_script` (`src/gui/value/color.cpp`, Zeile 38, Fassung f9d9356 vom
// 2021-07-06); das ist eine Frage der Oberfläche.
// Der Dienst räumt den Bestand.
func TestEineStiloptionLaesstSichZurueckstellen(t *testing.T) {
	t.Parallel()

	ablage := &stilablage{werte: []domain.Value{
		{Field: "hintergrund", Text: "dunkel"},
		{Field: "rahmen zeigen", Text: "yes"},
	}}

	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		ablage, &feldliste{})

	if err := waehle.OptionZuruecksetzen(t.Context(), 1, "hintergrund"); err != nil {
		t.Fatalf("zurücksetzen: %v", err)
	}

	// **Der Eintrag ist weg, und nur er.**
	for _, wert := range ablage.werte {
		if wert.Field == "hintergrund" {
			t.Errorf("`hintergrund` steht noch im Bestand: %+v", ablage.werte)
		}
	}

	if len(ablage.werte) != 1 || ablage.werte[0].Field != "rahmen zeigen" {
		t.Errorf("die übrigen Werte sind nicht unberührt: %+v", ablage.werte)
	}
}

// **Eine Option, die der Stil nicht kennt, wird abgewiesen** — dieselbe Prüfung
// wie beim Setzen. Ohne sie räumte ein Tippfehler still nichts.
func TestZuruecksetzenWeistEineFremdeOptionAb(t *testing.T) {
	t.Parallel()

	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		&stilablage{}, &feldliste{})

	if err := waehle.OptionZuruecksetzen(t.Context(), 1, "gibtesnicht"); err == nil {
		t.Error("eine unbekannte Option wurde stillschweigend zurückgesetzt")
	}
}

// **Und sie trägt den gerechneten Vorgabewert mit** — MSE zeichnet den
// Farbtupfer des Eintrags aus `default_color`, das `selection()` durch
// Auswerten des Vorgabeskripts gewinnt (`src/gui/value/color.cpp`, Zeilen
// 107–112, Fassung f9d9356 vom 2021-07-06).
//
// # Warum das nicht dasselbe ist wie [Stiloption.Wert]
//
// **`Wert` trägt den Rückfall, `Vorgabe` nur das ausdrückliche `default:`.**
// Ohne `initial:` und ohne `default:` gibt `Wert` die erste Wahl — auch dann,
// wenn die Vorlage gar keine Vorgabe nennt. Hinge der Eintrag daran, stünde er
// bei fast jeder Option, und MSE zeigt ihn nur bei `hasDefault()`
// (`src/gui/value/color.cpp`, Zeile 38, Fassung f9d9356 vom 2021-07-06).
func TestEineStiloptionTraegtIhrenVorgabewert(t *testing.T) {
	t.Parallel()

	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
		&stilablage{}, &feldliste{})

	optionen, err := waehle.Optionen(t.Context(), 1)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	// **Zwei ohne `default:`, eine mit** — die Saat trägt beide Fälle, damit
	// die Zusicherung nicht schon dadurch zutrifft, dass es keinen Fall gibt.
	for i := range optionen {
		hat := optionen[i].Vorgabe != ""
		soll := optionen[i].Name == "titelfarbe"

		if hat != soll {
			t.Errorf("%q: Vorgabe %q, erwartet %s", optionen[i].Name,
				optionen[i].Vorgabe, map[bool]string{true: "eine", false: "keine"}[soll])
		}

		// **Der Rückfall bleibt daneben stehen.** `Wert` trägt weiter, was
		// gälte — auch dort, wo die Vorlage keine Vorgabe nennt.
		if optionen[i].Wert == "" {
			t.Errorf("%q hat gar keinen Wert — der Rückfall greift nicht mehr", optionen[i].Name)
		}
	}
}

// **Und die beiden Fehlerwege des Zurücksetzens sind betreten** (§3).
//
// **Sie sind nicht dieselben**: der erste scheitert am Setkopf, bevor irgendein
// Wert gelesen wurde; der zweite am Lesen der Werte, also **nach** der Prüfung,
// dass es die Option gibt. Ein Test für einen von beiden ließe den anderen
// unbetreten.
func TestZuruecksetzenReichtSeineFehlerDurch(t *testing.T) {
	t.Parallel()

	t.Run("der Setkopf fehlt", func(t *testing.T) {
		t.Parallel()

		waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
			&stilablage{}, &feldliste{})

		if err := waehle.OptionZuruecksetzen(t.Context(), 99, "titelfarbe"); err == nil {
			t.Error("ein unbekanntes Set wurde stillschweigend zurückgesetzt")
		}
	})

	// **Der zweite ist seit dem Zweitblick über Phase L der Schreibfehler**,
	// nicht mehr der Lesefehler: die Option wird in **einer** Anweisung
	// entfernt, ohne die Liste vorher zu holen.
	t.Run("die Ablage nimmt die Änderung nicht an", func(t *testing.T) {
		t.Parallel()

		waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
			&stilablage{setzfehler: errLeseprobe}, &feldliste{})

		if err := waehle.OptionZuruecksetzen(t.Context(), 1, "titelfarbe"); err == nil {
			t.Error("ein Schreibfehler ging verloren")
		}
	})

	// **Und derselbe Weg beim Setzen** — er hatte bis hierher keinen eigenen
	// Test, weil der Lesefehler ihn deckte (`stilfehler_test.go`).
	t.Run("die Ablage nimmt das Setzen nicht an", func(t *testing.T) {
		t.Parallel()

		waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-schlicht"},
			&stilablage{setzfehler: errLeseprobe}, &feldliste{})

		if err := waehle.OptionSetzen(t.Context(), 1, "hintergrund", "hell"); err == nil {
			t.Error("ein Schreibfehler ging verloren")
		}
	})
}

// errLeseprobe ist der Fehler, den die Ablage im zweiten Fall gibt.
var errLeseprobe = errors.New("die Ablage antwortet nicht")
