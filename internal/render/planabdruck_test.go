package render_test

// **Der Abdruck eines Zeichenplans** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass zwei Pläne genau dann denselben Abdruck haben, wenn sie dasselbe
// zeichnen.** Er ist die halbe Antwort auf die Frage, ob ein gerastertes Bild
// noch gilt; die andere Hälfte sind die Dateien, die er nur **nennt**.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// einPlan ist der kleinste, der etwas zeichnet.
func einPlan() render.Plan {
	return render.Plan{
		Breite: 375, Hoehe: 523, DPI: 150,
		Ebenen: []render.Ebene{
			{
				Art:    render.ArtBild,
				Rahmen: render.Rechteck{Links: 0, Oben: 0, Breite: 375, Hoehe: 523},
				Bild:   render.Bildbaum{Knotenart: "datei", Pfad: "rahmen.png"},
			},
			{
				Art:    render.ArtText,
				Feld:   "name",
				Rahmen: render.Rechteck{Links: 30, Oben: 30, Breite: 300, Hoehe: 20},
				Satz:   render.Satz{Schrift: "MPlantin", Kursivschrift: "MPlantin-Italic"},
			},
		},
	}
}

// **Derselbe Plan gibt denselben Abdruck.**
func TestDerselbePlanGibtDenselbenAbdruck(t *testing.T) {
	t.Parallel()

	// **Zwei getrennt gebaute Pläne** und nicht zweimal derselbe Ausdruck:
	// der Abdruck soll über den **Inhalt** gehen und nicht über die Adresse
	// im Arbeitsspeicher.
	erster, zweiter := einPlan(), einPlan()

	if render.Planabdruck(erster) != render.Planabdruck(zweiter) {
		t.Error("zweimal derselbe Plan, zwei Abdrücke")
	}

	// **Und er ist hexadezimal und lang genug, um zu unterscheiden.**
	abdruck := render.Planabdruck(einPlan())
	if len(abdruck) != 64 || strings.Trim(abdruck, "0123456789abcdef") != "" {
		t.Errorf("der Abdruck %q sieht nicht aus wie einer", abdruck)
	}
}

// **Jede zeichnende Änderung ändert ihn.**
//
// **Tabellengetrieben und nicht als ein Fall** (§3): jede Klasse für sich, mit
// sprechendem Namen — sonst sagte ein roter Lauf nur, dass „irgendetwas"
// durchkommt.
func TestJedeZeichnendeAenderungAendertDenAbdruck(t *testing.T) {
	t.Parallel()

	grund := render.Planabdruck(einPlan())

	for name, aendern := range map[string]func(*render.Plan){
		"die Breite": func(p *render.Plan) { p.Breite = 376 },
		"die Höhe":   func(p *render.Plan) { p.Hoehe = 524 },
		"die DPI":    func(p *render.Plan) { p.DPI = 300 },
		"ein Bildpfad": func(p *render.Plan) {
			p.Ebenen[0].Bild.Pfad = "anderer.png"
		},
		"eine Schrift": func(p *render.Plan) {
			p.Ebenen[1].Satz.Schrift = "Beleren"
		},
		"ein Rahmen": func(p *render.Plan) {
			p.Ebenen[1].Rahmen.Links = 31
		},
		"eine Ebene weniger": func(p *render.Plan) {
			p.Ebenen = p.Ebenen[:1]
		},
		"die Reihenfolge der Ebenen": func(p *render.Plan) {
			p.Ebenen[0], p.Ebenen[1] = p.Ebenen[1], p.Ebenen[0]
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			geaendert := einPlan()
			aendern(&geaendert)

			if render.Planabdruck(geaendert) == grund {
				t.Errorf("%s ändert den Abdruck nicht", name)
			}
		})
	}
}

// **Und was nicht gezeichnet wird, ändert ihn nicht.**
//
// # Warum das eine Zusage ist und kein Versehen
//
// **Feldflächen und Vorgaben sind die Auskunft an die Oberfläche** — der
// Rasterer liest sie nicht. Stünden sie im Abdruck, verwürfe eine geänderte
// Herleitung ein Bild, das Pixel für Pixel dasselbe wäre; die Karte flackerte,
// und der Cache träfe nie.
func TestWasNichtGezeichnetWirdAendertDenAbdruckNicht(t *testing.T) {
	t.Parallel()

	grund := render.Planabdruck(einPlan())

	mitAuskunft := einPlan()
	mitAuskunft.Feldflaechen = []render.Feldflaeche{
		{Feld: "name", Rahmen: render.Rechteck{Breite: 10, Hoehe: 10}},
	}
	mitAuskunft.Vorgaben = []render.Vorgabe{{Feld: "typ", Wert: "Creature"}}

	if render.Planabdruck(mitAuskunft) != grund {
		t.Error("Feldflächen oder Vorgaben ändern den Abdruck")
	}
}

// **Die Fassung des Rasterers ist eine Zahl, die sich nennen lässt.**
func TestDieFassungDesRasterersIstBenannt(t *testing.T) {
	t.Parallel()

	if strings.TrimSpace(render.Fassung) == "" {
		t.Error("der Rasterer nennt keine Fassung")
	}
}

// **Ein Plan nennt seine Bilder und seine Schriften.**
//
// **Geordnet und ohne Dubletten**, denn wer daraus einen Schlüssel baut,
// braucht bei gleichem Plan dieselbe Reihenfolge.
func TestEinPlanNenntSeineBilderUndSchriften(t *testing.T) {
	t.Parallel()

	plan := einPlan()
	plan.Ebenen = append(plan.Ebenen, render.Ebene{
		Art:  render.ArtBild,
		Bild: render.Bildbaum{Knotenart: "datei", Pfad: "rahmen.png"},
		Maske: render.Bildbaum{
			Knotenart: "verknuepfung",
			Kinder: []render.Bildkind{{
				Name: "image",
				Baum: render.Bildbaum{Knotenart: "datei", Pfad: "maske.png"},
			}},
		},
	})

	bilder := render.Bildpfade(plan)
	if len(bilder) != 2 || bilder[0] != "maske.png" || bilder[1] != "rahmen.png" {
		t.Errorf("Bilder: %v", bilder)
	}

	schriften := render.Schriftnamen(plan)
	if len(schriften) != 2 ||
		schriften[0] != "MPlantin" || schriften[1] != "MPlantin-Italic" {
		t.Errorf("Schriften: %v", schriften)
	}
}

// **Ein Plan ohne Bilder und Schriften gibt leere Listen und kein `nil`.**
func TestEinLeererPlanGibtLeereListen(t *testing.T) {
	t.Parallel()

	if bilder := render.Bildpfade(render.Plan{}); bilder == nil {
		t.Error("Bildpfade gibt nil")
	}

	if schriften := render.Schriftnamen(render.Plan{}); schriften == nil {
		t.Error("Schriftnamen gibt nil")
	}
}
