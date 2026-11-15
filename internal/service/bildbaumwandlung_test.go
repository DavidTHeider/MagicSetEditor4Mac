package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Bildbaum aus einem Skript kommt als Zeichenbaum im Plan an.**
//
// # Die Naht, um die es geht
//
// `msescript` darf `render` nicht kennen und umgekehrt (§2). Der eine Baum ist
// ein Wert einer Sprache, der andere eine Zeichenanweisung; `service` ist die
// Stelle, an der beide zusammenkommen — und die einzige, die weiß, in welchem
// Paket ein Pfad ohne Schrägstrich zu suchen ist.
//
// Geprüft wird der Weg **ganz**: Feldangabe → Auswertung → Baum → Plan.
func TestEinBildbaumKommtImPlanAn(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "523",
			Bild: `{ recolor_image("rahmen.png", red: rgb(10, 20, 30), anteil: 2, art: "weich") }`,
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	var baum render.Bildbaum

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			baum = ebene.Bild

			break
		}
	}

	if baum.Knotenart != "recolor_image" {
		t.Fatalf("die Knotenart ist %q, erwartet recolor_image", baum.Knotenart)
	}

	// **Der Paketpfad entsteht an dieser Naht**, nicht später: bei einem Baum
	// gäbe es kein „hinterher", sondern viele Blätter.
	if len(baum.Kinder) != 1 || baum.Kinder[0].Baum.Pfad != "/probe.mse-style/rahmen.png" {
		t.Errorf("das Blatt ist %+v, erwartet /probe.mse-style/rahmen.png", baum.Kinder)
	}

	// **Nach Art getrennt und nicht als Text.** Eine Farbe als Zeichenkette
	// wieder zu zerlegen hieße, sie zweimal zu deuten.
	if farbe := baum.Farben["red"]; farbe.R != 10 || farbe.G != 20 || farbe.B != 30 {
		t.Errorf("die Farbe ist %+v, erwartet rgb(10, 20, 30)", farbe)
	}

	if zahl := baum.Zahlen["anteil"]; zahl != 2 {
		t.Errorf("die Zahl ist %v, erwartet 2", zahl)
	}

	if text := baum.Texte["art"]; text != "weich" {
		t.Errorf("der Text ist %q, erwartet weich", text)
	}
}

// **Ein Pfad ohne Skript bleibt ein Blatt.**
//
// Die häufigste Form, und die einzige, die es vor dem Bildbaum gab.
func TestEinePfadangabeWirdEinBlatt(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "523",
			Bild: "rahmen.png",
		}},
	}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild {
			continue
		}

		if !ebene.Bild.IstBlatt() || ebene.Bild.Pfad != "/probe.mse-style/rahmen.png" {
			t.Errorf("der Baum ist %+v, erwartet ein Blatt auf /probe.mse-style/rahmen.png",
				ebene.Bild)
		}

		return
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ein Feld ohne `image:` ergibt keinen Baum.**
//
// Der leere Baum ist „kein Bild" und nicht „ein Bild ohne Pfad": das eine ist
// der Normalfall, das andere wäre ein Mangel.
func TestEinFeldOhneBildangabeErgibtKeinenBaum(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "leer", Links: "0", Oben: "0", Breite: "10", Hoehe: "10",
		}},
	}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			t.Errorf("ein Feld ohne image: hat eine Bildebene ergeben: %+v", ebene.Bild)
		}
	}
}

// **Auch eine Maske ist ein Baum.**
//
// # Der Befund des Zweitblicks
//
// `Bild` wurde zum Baum, `Maske` blieb ein Pfad. Ergab ein `mask:`-Skript eine
// **Verknüpfung**, war deren Wurzel pfadlos — und die Maske verschwand still.
//
// **Bei einer Typmigration bleibt ein gleichartiger Nebenkanal gern im alten
// Skalartyp zurück**, und niemand merkt es, weil der Prüffall nur die einfache
// Form kennt: der Wächter daneben prüfte ein Skript, das unmittelbar einen
// Dateinamen zurückgibt.
func TestAuchEineMaskeIstEinBaum(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "523",
			Bild:  "rahmen.png",
			Maske: `{ recolor_image("maske.png", red: rgb(1, 2, 3)) }`,
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild {
			continue
		}

		if ebene.Maske.Knotenart != "recolor_image" {
			t.Fatalf("die Maske ist %+v, erwartet eine Verknüpfung", ebene.Maske)
		}

		if blaetter := ebene.Maske.Blaetter(); len(blaetter) != 1 ||
			blaetter[0] != "/probe.mse-style/maske.png" {
			t.Errorf("die Blätter der Maske sind %v, erwartet /probe.mse-style/maske.png",
				blaetter)
		}

		return
	}

	t.Fatal("der Plan führt keine Bildebene")
}
