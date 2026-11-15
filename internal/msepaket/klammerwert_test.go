package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Wert mit offener Klammer kommt zurück — und sein Nachbar auch.**
//
// # Der Befund
//
// Gefunden von der Fuzzkampagne (C02,
// #2227),
// `FuzzSymbolzeichnung/a2901ac8792fc369`: ein Teil namens `[` ließ beim
// Wiederlesen seinen Punkt verschwinden. Dieser Zerteiler zählt Klammern, um
// Skriptausdrücke über mehrere Zeilen zu lesen ([klammerbilanz]); ein
// einzeiliger Wert `name: [` öffnet damit einen Ausdruck, und jede folgende
// Zeile **derselben Tiefe** wird zu seiner Fortsetzung. **Dieselbe Klasse in
// einer Setdatei**: eine Karte namens `[` verlor beim Wiederöffnen jedes Feld
// nach ihrem Namen.
//
// MSE2 zählt keine Klammern — es liest einen Wert bis zum Zeilenende
// (`src/util/io/reader.cpp`, Zeilen 207–223, Fassung f9d9356 vom 2021-07-06)
// —, und seinen Block liest es ebenso (Zeilen 258–272). **Der Block ist die
// Form, die beide Leser gleich verstehen**; dieselbe Lösung wie für einen
// Wert mit Umbruch oder führendem Leerraum.
func TestEinWertMitOffenerKlammerKommtZurueck(t *testing.T) {
	t.Parallel()

	for _, wert := range []string{
		"[",
		"{",
		"[a: b",
		"Kosten {2}{R",
		"zu viel ]",
		"[ausgeglichen]",
	} {
		t.Run(wert, func(t *testing.T) {
			t.Parallel()

			baum := &msepaket.Baum{Wurzel: []*msepaket.Knoten{{
				Zeilenart:  msepaket.ArtEintrag,
				Schluessel: "card",
				Kinder: []*msepaket.Knoten{
					{Zeilenart: msepaket.ArtEintrag, Schluessel: "name", Wert: wert},
					{Zeilenart: msepaket.ArtEintrag, Schluessel: "farbe", Wert: "rot"},
				},
			}}}

			geschrieben := msepaket.Schreibe(baum)

			gelesen, err := msepaket.Zerteile("set", geschrieben)
			if err != nil {
				t.Fatalf("zerteilen: %v", err)
			}

			karte := gelesen.Ersten("card")
			if karte == nil {
				t.Fatalf("kein card-Block in\n%s", geschrieben)
			}

			if name := karte.Ersten("name"); name == nil || msepaket.WertOderBlock(name) != wert {
				t.Errorf("der Name kommt nicht zurück aus\n%s", geschrieben)
			}

			if farbe := karte.Ersten("farbe"); farbe == nil || farbe.Wert != "rot" {
				t.Errorf("das Feld nach dem Namen ist verschwunden in\n%s", geschrieben)
			}
		})
	}
}

// **Ein ausgeglichener Wert bleibt auf seiner Zeile** — der Block ist die
// Ausnahme für das, was eine Zeile nicht trägt, nicht die neue Regel.
func TestEinAusgeglichenerKlammerwertBleibtEinzeilig(t *testing.T) {
	t.Parallel()

	baum := &msepaket.Baum{Wurzel: []*msepaket.Knoten{
		{Zeilenart: msepaket.ArtEintrag, Schluessel: "name", Wert: "[ausgeglichen] {2}"},
	}}

	if hat := string(msepaket.Schreibe(baum)); hat != "name: [ausgeglichen] {2}\n" {
		t.Errorf("geschrieben %q", hat)
	}
}
