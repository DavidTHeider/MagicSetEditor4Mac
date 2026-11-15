// **Die Umsetzung eines Paletteneintrags ohne sein Bild** — Werkstück
// #1391.
//
// # Warum ein interner Lauf
//
// `alsPalettenzeilen` ist paketintern und rein: Einträge hinein, Zeilen
// heraus. Von außen wäre sie nur über eine **Vorlage mit Palette** zu
// erreichen — und die gehört nicht in eine Vorrichtung; §11 hält fremde
// Vorlagen aus dem Repositorium heraus.
//
// **Was der Lauf daneben belegt** (`kopflosenaehte_test.go`): dass der Weg
// vom Bestand bis hierher trägt. Was er nicht belegt, weil das Probeset keine
// Palette mitbringt: dass ein Eintrag richtig umgesetzt wird. Genau das steht
// hier.
package app

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestAlsPalettenzeilenWirftNurDasBildWeg(t *testing.T) {
	zeilen := alsPalettenzeilen([]service.Paletteneintrag{
		{
			Art: "code", Code: "W", Beschriftung: "Weiß",
			Rueckfrage: "",
		},
		{
			Art: "custom", Code: "", Beschriftung: "Eigenes",
			Rueckfrage: "Welches Zeichen?",
		},
		{Art: "line"},
	})

	if len(zeilen) != 3 {
		t.Fatalf("%d Zeilen, erwartet drei", len(zeilen))
	}

	if zeilen[0] != (Palettenzeile{Art: "code", Code: "W", Beschriftung: "Weiß"}) {
		t.Errorf("erste Zeile %+v", zeilen[0])
	}

	// **Die Rückfrage reist mit.** Ohne sie wüsste ein Aufrufer bei einem
	// `custom`-Eintrag nicht, was er fragen soll — und das ist die einzige
	// Auskunft, die dort nicht im Code steht.
	if zeilen[1].Rueckfrage != "Welches Zeichen?" {
		t.Errorf("zweite Zeile %+v", zeilen[1])
	}

	if zeilen[2].Art != "line" {
		t.Errorf("dritte Zeile %+v", zeilen[2])
	}
}

// **Und eine leere Palette gibt eine leere Liste, keine `null`.**
//
// Die Falle aus §3: über die Naht würde aus `nil` ein `null`, und ein
// Aufrufer, der darüber schleift, bekäme einen Fehler statt keiner Einträge.
func TestAlsPalettenzeilenGibtEineLeereListe(t *testing.T) {
	if zeilen := alsPalettenzeilen(nil); zeilen == nil {
		t.Error("nil statt einer leeren Liste")
	}
}
