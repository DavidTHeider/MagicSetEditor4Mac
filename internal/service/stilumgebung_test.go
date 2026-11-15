package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Umgebung steht, bevor ein Skript sie liest**
// (#240).
//
// # Was das gekostet hat
//
// `card` und `set` wurden **nach** dem Initskript gesetzt. Das Initskript von
// `magic.mse-game` liest `set` an Zeile 10580 — es brach dort ab, und alles
// danach fehlte: `module_card`, `card_style`, `color_background`,
// `module_identity`. Genau die vier Namen, an denen jede `image:`-Angabe
// scheiterte und die Karte weiß blieb.
//
// Dieselbe Falle hatte `exportvorlage.go` schon. Zweimal dieselbe heißt: es ist
// eine Regel.
func TestDasSpielskriptSiehtDieKarteUndDenStil(t *testing.T) {
	stil := msepaket.Stil{
		Spiel: "probe", Breite: 10, Hoehe: 10,
		Stiloptionen: []domain.StyleField{{Name: "rand oben", Initial: "12"}},
		Felder: []msepaket.Feldstil{
			{Feld: "name", Links: "5", Oben: "7", Breite: "40", Hoehe: "9"},
		},
	}

	// Jeder der vier Namen wird **im Initskript** gelesen — genau dort, wo es
	// vorher abbrach.
	vorlage := service.Vorlage{
		Stil: stil,
		Spielskript: `gesehen := card.name + "|" + styling.rand_oben + "|" +` +
			` to_string(card_style.name.content_width) + "|" + to_string(stylesheet.rand_oben)`,
	}

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "nicht bekannt") {
			t.Errorf("ein Name fehlte in der Umgebung: %s", befund.Meldung)
		}
	}
}

// **Ein gescheiterter Satz des Initskripts nimmt die übrigen nicht mit** (#240).
//
// # Warum das der Kern von #240 war
//
// `magic.mse-game` ruft an Zeile 14741 ein `expand_keywords`, das dieses
// Programm nicht kennt. Vorher verlor es damit die **4.000 Zeilen danach** —
// und in ihnen stand, was den Rahmen zeichnet.
func TestEinGescheiterterSatzNimmtDieUebrigenNichtMit(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{Spiel: "probe", Breite: 10, Hoehe: 10},
		Spielskript: "vorher := \"da\"\n" +
			"mitten := gibtesnicht()\n" +
			"nachher := \"auch da\"\n",
	}

	karte := &domain.CardRow{ID: 1}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// **Der ausgefallene Satz wird genannt** — verschwiegen wird er nicht.
	genannt := false

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "gibtesnicht") {
			genannt = true
		}
	}

	if !genannt {
		t.Error("der ausgefallene Satz steht in keinem Befund")
	}

	// **Und `nachher` ist trotzdem entstanden.** Die Gegenprobe geht über ein
	// Feldskript, das ihn liest: ohne die Regel wäre er unbekannt.
	vorlage.Stil.Felder = []msepaket.Feldstil{{Feld: "a", Links: "0", Oben: "0", Bild: "{ nachher }"}}

	_, weitere, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range weitere {
		if strings.Contains(befund.Meldung, `"nachher" ist hier nicht bekannt`) {
			t.Error("der Satz nach dem gescheiterten ist nicht gelaufen")
		}
	}
}

// **Eine Maßangabe, die ein Skript ist, wird null und bricht nicht ab.**
func TestEineUnlesbareMassangabeWirdNull(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 10, Hoehe: 10,
			Felder: []msepaket.Feldstil{{Feld: "a", Links: "{ gerechnet }", Oben: "0"}},
		},
		Spielskript: `gepruft := to_string(card_style.a.left)`,
	}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(),
		&domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "nicht bekannt") {
			t.Errorf("%s", befund.Meldung)
		}
	}
}

// **Ein Skript, das sich gar nicht zerteilen lässt, bleibt ein Abbruch.**
//
// Die Gegenprobe zur Regel darüber: satzweise weitergerechnet wird nur, wo es
// **Sätze** gibt. Ein Skript, das schon der Zerteiler abweist, hat keine.
func TestEinUnzerteilbaresInitskriptBleibtEinBefund(t *testing.T) {
	vorlage := service.Vorlage{
		Stil:        msepaket.Stil{Spiel: "probe", Breite: 10, Hoehe: 10},
		Spielskript: `x := (`,
	}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(),
		&domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	gemeldet := false

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "ließ sich nicht ausführen") {
			gemeldet = true
		}
	}

	if !gemeldet {
		t.Error("ein unzerteilbares Initskript ging stillschweigend durch")
	}
}

// **Ein einzelner Ausdruck ohne Folge geht weiterhin am Stück.**
func TestEinEinzelnerAusdruckAlsInitskript(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 10, Hoehe: 10,
			Felder: []msepaket.Feldstil{{Feld: "a", Links: "0", Oben: "0", Bild: "{ allein }"}},
		},
		Spielskript: `allein := "da"`,
	}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(),
		&domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "nicht bekannt") {
			t.Errorf("%s", befund.Meldung)
		}
	}
}

// **`card_style` führt jedes Kartenfeld des Spiels — auch das ungestylte.**
//
// MSE legt die Abbildung aus den Feldern des **Spiels** an, nicht aus denen der
// Vorlage:
//
//	REFLECT_IF_READING {
//	  card_style.init(game->card_fields);
//	  …
//	}
//	REFLECT(card_style);
//
// `src/data/stylesheet.cpp`, Zeilen 106–113, Fassung f9d9356 vom 2021-07-06.
// `IndexMap::init` legt für **jeden** Schlüssel einen Eintrag an
// (`src/util/index_map.hpp`, Zeilen 48–61); was die Vorlage nicht beschreibt,
// bekommt einen Stil mit Nullmaßen. Danach überschreibt `REFLECT` nur, was
// wirklich dasteht.
//
// **Ohne das ist `card_style.<feld>` nichts**, und `content_width` daran ein
// Fehler statt einer Null — dreimal über die 190 Stile von `M15-Magic-Pack`
// (#291).
func TestCardStyleFuehrtJedesKartenfeldDesSpiels(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 10, Hoehe: 10,
			Felder: []msepaket.Feldstil{
				{Feld: "name", Links: "5", Oben: "7", Breite: "40", Hoehe: "9"},
			},
		},

		// **Das Spiel führt zwei Felder, die Vorlage stylt eines.**
		Kartenfelder: []domain.Field{{Name: "name"}, {Name: "pt"}},

		Spielskript: `gesehen := to_string(card_style.pt.content_width) + "|" +` +
			` to_string(card_style.pt.width)`,
	}

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "content_width") || strings.Contains(befund.Meldung, "width") {
			t.Errorf("das ungestylte Feld fehlte in card_style: %s", befund.Meldung)
		}
	}
}

// **Ein Feldstil hat sechs Maße, und drei davon rechnet MSE aus.**
//
//	if      (automatic_side & AUTO_LEFT)   left   = right - width;
//	else if (automatic_side & AUTO_WIDTH)  width  = right - left;
//	else if (automatic_side & AUTO_RIGHT)  right  = left + width;
//	else  { … left und right um die Mitte … }
//
// `Style::update`, `src/data/field.cpp`, Zeilen 152–172, Fassung f9d9356 vom
// 2021-07-06. Welches der drei fehlt, entscheidet `automatic_side`: was in der
// Vorlage nicht dasteht, trägt bei MSE 1.000.000 und gilt als **nicht gesetzt**.
//
// **Sichtbar wird das an den Saga-Vorlagen**: sie stapeln ihre Stufen über
// `card_style.level_1_text.bottom`. Ohne `right` und `bottom` in der Umgebung
// ergab jede dieser Angaben „keine Zahl, sondern nichts" — zwölfmal über die
// 190 Stile von `M15-Magic-Pack`
// (#291).
func TestEinFeldstilRechnetSeineFehlendenMasseAus(t *testing.T) {
	faelle := []struct {
		name          string
		stil          msepaket.Feldstil
		rechts, unten float64
	}{
		{
			name:   "ohne rechts wird rechts gerechnet",
			stil:   msepaket.Feldstil{Feld: "f", Links: "10", Breite: "30", Oben: "5", Hoehe: "7"},
			rechts: 40, unten: 12,
		},
		{
			name:   "ohne breite wird die breite gerechnet",
			stil:   msepaket.Feldstil{Feld: "f", Links: "10", Rechts: "40", Oben: "5", Unten: "12"},
			rechts: 40, unten: 12,
		},
		{
			// **Ohne links steht rechts schon da** — gerechnet wird die linke
			// Kante, und die Gegenkante bleibt, wie sie ist.
			name:   "ohne links bleibt rechts stehen",
			stil:   msepaket.Feldstil{Feld: "f", Breite: "30", Rechts: "40", Hoehe: "7", Unten: "12"},
			rechts: 40, unten: 12,
		},
		{
			// **Alle drei gesetzt**: die Breite gilt und steht mittig zwischen
			// den Kanten — aus 0/10/40 wird 15…25.
			name:   "mit allen dreien steht die breite mittig",
			stil:   msepaket.Feldstil{Feld: "f", Links: "0", Breite: "10", Rechts: "40", Oben: "0", Hoehe: "10", Unten: "40"},
			rechts: 25, unten: 25,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			// **Das zweite Feld liest die gerechneten Kanten des ersten** — so
			// wie die Saga-Vorlagen ihre Stufen stapeln.
			zweites := msepaket.Feldstil{
				Feld:  "g",
				Links: "{ card_style.f.right }", Oben: "{ card_style.f.bottom }",
				Breite: "10", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}

			vorlage := service.Vorlage{Stil: msepaket.Stil{
				Spiel: "probe", Breite: 100, Hoehe: 100,
				Felder: []msepaket.Feldstil{fall.stil, zweites},
			}}

			karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "g", Text: "x"}}}

			plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, befund := range befunde {
				t.Errorf("Befund: %s — %s", befund.Ort, befund.Meldung)
			}

			if len(plan.Ebenen) < 2 {
				t.Fatalf("%d Ebenen, erwartet mindestens zwei", len(plan.Ebenen))
			}

			rahmen := plan.Ebenen[len(plan.Ebenen)-1].Rahmen
			if rahmen.Links != fall.rechts || rahmen.Oben != fall.unten {
				t.Errorf("das zweite Feld sitzt bei (%g,%g), erwartet (%g,%g)",
					rahmen.Links, rahmen.Oben, fall.rechts, fall.unten)
			}
		})
	}
}
