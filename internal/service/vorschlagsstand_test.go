// **Der Stand eines Vorschlags** —
// #1241.
//
// **Eine eigene Datei**, weil `scryfallabgleich_test.go` mit ihr über 400
// Zeilen ginge (§2). Der Schnitt folgt der Frage: dort steht, **was** ein
// Vorschlag enthält; hier, **woran er gebunden ist**.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Vorschlag ohne Änderungen bindet nichts** —
// #1241.
//
// Es gibt nichts zu bestätigen: `Uebernehmen` schreibt ohnehin nicht.
func TestEinVorschlagOhneAenderungenHatKeinenStand(t *testing.T) {
	t.Parallel()

	if stand := (service.Vorschlag{}).BerechneterStand(); stand != "" {
		t.Errorf("Stand %q, erwartet keinen", stand)
	}
}

// **Und ein geänderter alter Wert ändert ihn.**
func TestEinGeaenderterAlterWertAendertDenStand(t *testing.T) {
	t.Parallel()

	erst := service.Vorschlag{Aenderungen: []service.Feldvorschlag{
		{Feld: "name", Alt: "Feuerball", Neu: "Fireball"},
	}}

	dann := service.Vorschlag{Aenderungen: []service.Feldvorschlag{
		{Feld: "name", Alt: "Feuerkugel", Neu: "Fireball"},
	}}

	if erst.BerechneterStand() == dann.BerechneterStand() {
		t.Error("derselbe Stand für zwei verschiedene Ausgangswerte")
	}
}
