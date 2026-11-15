package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Wege durch die Symbolschriften, die eine Karte allein nicht betritt.
//
// **Jeder davon ist eine Entscheidung**, und eine Entscheidung ohne Testfall ist
// eine Behauptung darüber, was im Ernstfall geschieht (§3).
func TestDieSymbolwegeSindBetreten(t *testing.T) {
	symbolfeld := func(anpassen func(*msepaket.Feldstil)) msepaket.Feldstil {
		feld := msepaket.Feldstil{
			Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "40",
			Symbolschrift: msepaket.Symbolangabe{Name: "mana"},
			ImmerSymbole:  true,
		}
		anpassen(&feld)

		return feld
	}

	// **Ohne `size:` gilt die Höhe des Kastens.** Einige Stile lassen die
	// Angabe weg; eine erfundene Zahl wäre schlechter, und null zeichnete nichts.
	t.Run("ohne Größenangabe gilt die Kastenhöhe", func(t *testing.T) {
		plan, _ := gedeutetMitVorlage(t, kartemit("casting cost", "R"),
			symbolvorlage(symbolfeld(func(*msepaket.Feldstil) {})))

		if hoehe := ebenen(t, plan)[0].Rahmen.Hoehe; hoehe != 40 {
			t.Errorf("%v", hoehe)
		}
	})

	// **Eine unlesbare Größenangabe fällt auf dieselbe Regel zurück.**
	t.Run("eine unlesbare Größe", func(t *testing.T) {
		plan, _ := gedeutetMitVorlage(t, kartemit("casting cost", "R"),
			symbolvorlage(symbolfeld(func(f *msepaket.Feldstil) {
				f.Symbolschrift.Groesse = "{ gibtsnicht() }"
			})))

		if hoehe := ebenen(t, plan)[0].Rahmen.Hoehe; hoehe != 40 {
			t.Errorf("%v", hoehe)
		}
	})

	t.Run("mittig ausgerichtet", func(t *testing.T) {
		plan, _ := gedeutetMitVorlage(t, kartemit("casting cost", "R"),
			symbolvorlage(symbolfeld(func(f *msepaket.Feldstil) {
				f.Ausrichtung = "middle center"
				f.Symbolschrift.Groesse = "20"
			})))

		// **Die Ausrichtung reist mit**; ausgerechnet wird sie beim Rastern.
		if art := ebenen(t, plan)[0].Satz.Ausrichtung; art != render.MitteMitte {
			t.Errorf("%q", art)
		}
	})

	// **Ein Name, der ein Skript ist** — `text:` in `magic-m15.mse-style`
	// schreibt `symbol font: { styling.text_box_mana_symbols }`.
	t.Run("ein Name aus einem Skript", func(t *testing.T) {
		feld := symbolfeld(func(f *msepaket.Feldstil) {
			f.Symbolschrift.Name = `{ "ma" + "na" }`
			f.Symbolschrift.Groesse = "15"
		})

		plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "R"),
			symbolvorlage(feld))
		if len(befunde) > 0 {
			t.Fatalf("%v", befunde)
		}

		symbole := ebenen(t, plan)[0].Abschnitte[0].Symbole
		if len(symbole) != 1 || symbole[0].Bild.Beschreibung() != "/mana.mse-symbol-font/r.png" {
			t.Errorf("%+v", symbole)
		}
	})

	// **Ein Skript, das scheitert, macht das Feld zum Textfeld.** Ohne
	// Textschrift bleibt es ungedeutet und wird benannt.
	t.Run("ein Skript, das scheitert", func(t *testing.T) {
		feld := symbolfeld(func(f *msepaket.Feldstil) {
			f.Symbolschrift.Name = "{ gibtsnicht() }"
		})

		_, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "R"), symbolvorlage(feld))
		if len(befunde) != 1 {
			t.Fatalf("%d Befunde: %v", len(befunde), befunde)
		}
	})

	// **Ein leerer Wert ergibt keine Symbolebene.**
	t.Run("ein leeres Feld", func(t *testing.T) {
		plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", ""),
			symbolvorlage(symbolfeld(func(*msepaket.Feldstil) {})))
		if len(befunde) > 0 {
			t.Fatalf("%v", befunde)
		}

		if len(plan.Ebenen) != 1 {
			t.Errorf("%d Ebenen — erwartet war nur der Grund", len(plan.Ebenen))
		}
	})
}

// **Der Stilspeicher lädt die Symbolschriften, die ein Stil nennt.**
//
// Die Naht zwischen `msepaket` und `service`: der Planbauer öffnet keine
// Pakete, er bekommt ein Modell. Ohne diesen Test wäre nicht belegt, dass ihm
// jemand eines gibt.
func TestDerStilspeicherLaedtSymbolschriften(t *testing.T) {
	wurzel := t.TempDir()

	paket := filepath.Join(wurzel, "mana.mse-symbol-font")
	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(paket, "symbol-font"),
		[]byte("mse version: 0.3.9\nsymbol:\n\tcode: R\n\timage: r.png\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stil := msepaket.Stil{
		Paket: "probe.mse-style", Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
			Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
			ImmerSymbole:  true,
		}},
	}

	schriften, befunde := service.NeuerStilspeicher(nil, wurzel).Symbolschriften(stil)
	if len(befunde) > 0 {
		t.Fatalf("%v", befunde)
	}

	if _, da := schriften["mana"]; !da {
		t.Fatalf("die Schrift fehlt: %v", schriften)
	}

	// **Ein Stil ohne Symbolschrift lädt nichts** — und öffnet kein Paket.
	ohne := msepaket.Stil{Paket: "probe.mse-style", Felder: []msepaket.Feldstil{{Feld: "name"}}}

	if geladen, _ := service.NeuerStilspeicher(nil, wurzel).Symbolschriften(ohne); geladen != nil {
		t.Errorf("%v", geladen)
	}

	// **Eine Schrift, die es nicht gibt, ist hier kein Befund.** Ob das eine
	// Lücke ist, entscheidet die Karte — gemeldet wird an der Ebene.
	fehlt := stil
	fehlt.Felder = []msepaket.Feldstil{{
		Feld: "casting cost", Symbolschrift: msepaket.Symbolangabe{Name: "gibtsnicht"},
		ImmerSymbole: true,
	}}

	geladen, weitere := service.NeuerStilspeicher(nil, wurzel).Symbolschriften(fehlt)
	if len(weitere) > 0 || len(geladen) != 0 {
		t.Errorf("%v / %v", geladen, weitere)
	}
}

// **Die Wege, die eine gut gemachte Vorlage nie nimmt.**
//
// Sie sind trotzdem betreten, denn fremde Daten sind der Normalfall (§8): ein
// Paket von 2007 hält sich an keine Zusage, die wir uns wünschen.
func TestDieUnwahrscheinlichenSymbolwegeSindBetreten(t *testing.T) {
	// **Nur eine Gegenkante, keine Länge** — die Fläche beginnt am Rand.
	t.Run("nur eine Gegenkante", func(t *testing.T) {
		feld := msepaket.Feldstil{
			Feld: "name", Rechts: "100", Unten: "50",
			Schrift: msepaket.Schrift{Name: "Beleren", Groesse: "12"},
		}

		plan, befunde := gedeutetMit(t, kartemit("name", "x"), bildstil(feld))
		if len(befunde) > 0 {
			t.Fatalf("%v", befunde)
		}

		if rahmen := ebenen(t, plan)[0].Rahmen; rahmen.Breite != 100 || rahmen.Hoehe != 50 {
			t.Errorf("%+v", rahmen)
		}
	})

	// **Nur eine Länge, keine Kante** — sie beginnt bei null.
	t.Run("nur eine Länge", func(t *testing.T) {
		feld := msepaket.Feldstil{
			Feld: "name", Breite: "80", Hoehe: "40",
			Schrift: msepaket.Schrift{Name: "Beleren", Groesse: "12"},
		}

		plan, befunde := gedeutetMit(t, kartemit("name", "x"), bildstil(feld))
		if len(befunde) > 0 {
			t.Fatalf("%v", befunde)
		}

		if rahmen := ebenen(t, plan)[0].Rahmen; rahmen.Links != 0 || rahmen.Breite != 80 {
			t.Errorf("%+v", rahmen)
		}
	})

	// **Eine Angabe, die eine Zeichenkette ergibt, die sich als Zahl liest.**
	// `card_style.name.content_width` gibt in einigen Vorlagen einen Text.
	t.Run("eine Zahl als Zeichenkette", func(t *testing.T) {
		feld := msepaket.Feldstil{
			Feld: "name", Links: `{ "12" }`, Oben: "0", Breite: "80", Hoehe: "40",
			Schrift: msepaket.Schrift{Name: "Beleren", Groesse: "12"},
		}

		plan, befunde := gedeutetMit(t, kartemit("name", "x"), bildstil(feld))
		if len(befunde) > 0 {
			t.Fatalf("%v", befunde)
		}

		if links := ebenen(t, plan)[0].Rahmen.Links; links != 12 {
			t.Errorf("%v", links)
		}
	})

	// **Eine Symbolschrift ohne Textschrift, die fehlt** — der Symbolzweig ist
	// dann der einzige Weg, und der Befund sagt, was fehlt.
	t.Run("eine fehlende Schrift ohne Textalternative", func(t *testing.T) {
		feld := msepaket.Feldstil{
			Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
			Symbolschrift: msepaket.Symbolangabe{Name: "fehlt", Groesse: "15"},
			ImmerSymbole:  true,
		}

		_, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "R"),
			planvorlage(bildstil(feld), map[string]msepaket.Symbolschrift{}))

		if len(befunde) != 1 {
			t.Fatalf("%d Befunde: %v", len(befunde), befunde)
		}
	})
}

// **Die letzten drei Zweige** — ohne sie stünde neben jedem eine Behauptung.
func TestDieLetztenSymbolzweigeSindBetreten(t *testing.T) {
	// **Ein unbekanntes Zeichen mitten in sonst lauter Symbolen.** Der Wert
	// geht dann nicht den Symbolweg — es sei denn, es gibt keine Textschrift.
	t.Run("ein unbekanntes Zeichen ohne Textalternative", func(t *testing.T) {
		feld := msepaket.Feldstil{
			Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
			Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
			ImmerSymbole:  true,
		}

		// `Q` führt die Probeschrift nicht. Mit Textschrift wird das Feld Text;
		// hier steht keine, also bleibt der Symbolweg — samt Befund.
		_, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "Q"),
			symbolvorlage(feld))

		if len(befunde) != 1 {
			t.Fatalf("%d Befunde: %v", len(befunde), befunde)
		}
	})

	// **Ein Wert aus lauter Leerzeichen ist kein Symbolzug.**
	t.Run("nur Leerzeichen", func(t *testing.T) {
		feld := msepaket.Feldstil{
			Feld: "casting cost", Links: "0", Oben: "0", Breite: "60", Hoehe: "15",
			Symbolschrift: msepaket.Symbolangabe{Name: "mana", Groesse: "15"},
			ImmerSymbole:  true,
		}

		plan, befunde := gedeutetMitVorlage(t, kartemit("casting cost", "   "),
			symbolvorlage(feld))
		if len(befunde) != 1 {
			t.Fatalf("%d Befunde: %v", len(befunde), befunde)
		}

		if len(plan.Ebenen) != 1 {
			t.Errorf("%d Ebenen", len(plan.Ebenen))
		}
	})

	// **Eine Maßangabe, die weder Zahl noch lesbare Zeichenkette ergibt.**
	t.Run("eine Angabe ohne Zahl", func(t *testing.T) {
		feld := msepaket.Feldstil{
			Feld: "name", Links: `{ "links" }`, Oben: "0", Breite: "80", Hoehe: "40",
			Schrift: msepaket.Schrift{Name: "Beleren", Groesse: "12"},
		}

		_, befunde := gedeutetMit(t, kartemit("name", "x"), bildstil(feld))
		if len(befunde) != 1 {
			t.Fatalf("%d Befunde: %v", len(befunde), befunde)
		}
	})
}

// **Eine Symbolschrift nach Namen — für die Ausfuhr** (#658).
//
// `symbols_to_html` nennt `magic-mana-small`, und dort steht kein Stil daneben,
// aus dem sich die Schrift ableiten ließe.
func TestEineSymbolschriftNachNamen(t *testing.T) {
	wurzel := t.TempDir()

	paket := filepath.Join(wurzel, "mana.mse-symbol-font")
	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(paket, "symbol-font"),
		[]byte("mse version: 0.3.9\nimage font size: 135\nsymbol:\n\tcode: R\n\timage: r.png\n"),
		0o600); err != nil {
		t.Fatal(err)
	}

	speicher := service.NeuerStilspeicher(nil, wurzel)

	schrift, da := speicher.Symbolschrift("mana")
	if !da {
		t.Fatal("die Schrift wurde nicht gefunden")
	}

	if schrift.Bildgroesse != 135 || len(schrift.Symbole) != 1 {
		t.Errorf("gelesen wurde %+v", schrift)
	}

	// **Und eine, die es nicht gibt, sagt das.**
	if _, da := speicher.Symbolschrift("gibtsnicht"); da {
		t.Error("eine Schrift ohne Paket wurde gefunden")
	}
}
