// **Ein Skriptwert wird als Skript gefärbt** —
// #1181.
//
// Das ist die vierte Art, die den Aufwand lohnt: die Fachlogik einer Vorlage
// steht genau dort.
//
// **Warum der Lauf hier steht und nicht in `msepaket`**: jenes darf
// `msescript` nicht kennen (§2). Was ein Skriptwert **ist**, weiß das
// Paketformat; **wie ein Skript aussieht**, weiß die Sprache. Zusammengesetzt
// wird hier.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinSkriptwertWirdAlsSkriptGefaerbt(t *testing.T) {
	t.Parallel()

	merkmale := service.Paketfarben("\tscript: if x then 1 else 2\n")

	arten := map[string]int{}
	for _, merkmal := range merkmale {
		arten[merkmal.Art]++
	}

	if arten["Schlüsselwort"] < 3 {
		t.Errorf("%d Schlüsselwörter, erwartet mindestens drei "+
			"(if, then, else): %+v", arten["Schlüsselwort"], merkmale)
	}

	if arten["Skriptwert"] > 0 {
		t.Errorf("der Skriptrumpf steht ungefärbt da: %+v", merkmale)
	}
}

// **Die Stelle stimmt.** Der Lexer zählt ab Zeile eins, Spalte eins; hier
// steht der Rumpf mitten in einer Paketdatei.
func TestDieStelleEinesSkriptwertsStimmt(t *testing.T) {
	t.Parallel()

	// `\tscript: if …` — das `if` steht in Spalte 10.
	for _, merkmal := range service.Paketfarben("\tscript: if x then 1 else 2\n") {
		if merkmal.Art == "Schlüsselwort" {
			if merkmal.Zeile != 1 || merkmal.Spalte != 10 {
				t.Errorf("das erste Schlüsselwort steht bei %d:%d, erwartet 1:10",
					merkmal.Zeile, merkmal.Spalte)
			}

			return
		}
	}

	t.Error("kein Schlüsselwort gefunden")
}

// **Ein mehrzeiliger Rumpf wird ab der zweiten Zeile nicht versetzt.**
//
// Dort steht er am Zeilenanfang, und die Spalte des Lexers ist die richtige.
func TestEinMehrzeiligerRumpfWirdNurInDerErstenZeileVersetzt(t *testing.T) {
	t.Parallel()

	merkmale := service.Paketfarben("\tscript:\n\t\tif x then 1\n")

	for _, merkmal := range merkmale {
		if merkmal.Zeile == 2 && merkmal.Spalte < 1 {
			t.Errorf("Merkmal %+v steht vor dem Zeilenanfang", merkmal)
		}
	}
}

// **Ein gewöhnlicher Wert bleibt einer.**
func TestEinGewoehnlicherWertBleibtEiner(t *testing.T) {
	t.Parallel()

	merkmale := service.Paketfarben("game: magic\n")

	if len(merkmale) != 2 || merkmale[1].Art != "Wert" {
		t.Errorf("Merkmale %+v, erwartet Schlüssel und Wert", merkmale)
	}
}

// **Ein Skriptwert, der leer ist, kommt gar nicht erst hierher.**
//
// `msepaket` gibt für einen Schlüssel ohne Wert nur den Schlüssel; dieser Lauf
// hält fest, dass daraus kein leeres Merkmal wird.
func TestEinLeererSkriptwertGibtNurDenSchluessel(t *testing.T) {
	t.Parallel()

	merkmale := service.Paketfarben("\tscript:\n")

	if len(merkmale) != 1 || merkmale[0].Art != "Schlüssel" {
		t.Errorf("Merkmale %+v, erwartet nur den Schlüssel", merkmale)
	}
}

// **Ein Zeilenende im Skriptwert bekommt keine Farbe.**
//
// `a; b` ist in MSE-Script eine Folge, und das `;` ist ein Zeilenende. Ein
// Abschnitt darüber legte sich über den Zeilenwechsel.
func TestEinZeilenendeImSkriptwertBekommtKeineFarbe(t *testing.T) {
	t.Parallel()

	for _, merkmal := range service.Paketfarben("\tscript: 1; 2\n") {
		if merkmal.Art == "Zeilenende" {
			t.Errorf("ein Zeilenende bekam eine Farbe: %+v", merkmal)
		}
	}
}
