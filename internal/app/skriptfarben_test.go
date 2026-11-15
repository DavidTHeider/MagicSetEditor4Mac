package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Merkmale kommen mit Art, Stelle und Breite.**
func TestDieSkriptfarbenKommenMitStelle(t *testing.T) {
	anwendung := app.NeueApp("")

	antwort := anwendung.Skriptfarben(`card.name # ein Kommentar`)

	if len(antwort.Merkmale) < 4 {
		t.Fatalf("nur %d Merkmale: %+v", len(antwort.Merkmale), antwort.Merkmale)
	}

	erstes := antwort.Merkmale[0]
	if erstes.Art != "Bezeichner" || erstes.Zeile != 1 || erstes.Spalte != 1 || erstes.Breite != 4 {
		t.Errorf("das erste Merkmal ist %+v", erstes)
	}

	letztes := antwort.Merkmale[len(antwort.Merkmale)-1]
	if letztes.Art != "Kommentar" {
		t.Errorf("das letzte Merkmal ist %+v", letztes)
	}
}

// **Ein Zeilenende bekommt keine Farbe.**
//
// Es ist der Umbruch selbst; ein Abschnitt darüber legte sich über den
// Zeilenwechsel und schöbe die folgende Zeile.
func TestEinZeilenendeBekommtKeineFarbe(t *testing.T) {
	anwendung := app.NeueApp("")

	for _, merkmal := range anwendung.Skriptfarben("1\n2;3").Merkmale {
		if merkmal.Art == "Zeilenende" {
			t.Errorf("ein Zeilenende trägt eine Farbe: %+v", merkmal)
		}
	}
}

// **Ein unfertiger Ausdruck kommt trotzdem gefärbt zurück.**
//
// Beim Tippen ist er das die meiste Zeit. Ein Fenster, das dann nichts färbt,
// lässt den Text bei jedem zweiten Anschlag die Farbe verlieren.
func TestEinUnfertigerAusdruckKommtGefaerbtZurueck(t *testing.T) {
	anwendung := app.NeueApp("")

	if antwort := anwendung.Skriptfarben(`card.name + "unfertig`); len(antwort.Merkmale) < 3 {
		t.Errorf("nur %d Merkmale: %+v", len(antwort.Merkmale), antwort.Merkmale)
	}
}

// **Ein leerer Ausdruck gibt eine leere Liste, kein `nil`.**
func TestEinLeererAusdruckGibtEineLeereListe(t *testing.T) {
	if antwort := app.NeueApp("").Skriptfarben(""); antwort.Merkmale == nil {
		t.Error("die Merkmale sind nil statt leer")
	}
}

// **Und die Bindung am Fenster geht denselben Weg.**
func TestDieFarbbindungGehtDenselbenWeg(t *testing.T) {
	antwort := app.NeueBindungen(app.NeueApp("")).Skriptfarben(t.Context(), "1 + 1")

	if len(antwort.Merkmale) != 3 {
		t.Errorf("%d Merkmale: %+v", len(antwort.Merkmale), antwort.Merkmale)
	}
}
