package schriften

// **Eine Wurzel, die unter einer früheren liegt, wird nicht noch einmal
// gegangen** — #2005.
//
// Gemessen am 2026-10-02 im Leistungsvergleich: seit FMT-02 (#1940) steht das
// Paketverzeichnis vorn in jeder Liste von Paketwurzeln, und das Zeichenwerk
// gibt dieselbe Stelle als Schriftwurzel noch einmal davor. Die Suche ging
// damit `pakete`, `pakete` und `pakete/…/data` — denselben Baum dreimal je
// Schrift, die nicht unter ihrem Dateinamen liegt.

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestEineWurzelUnterEinerFruherenWirdNichtNochEinmalGegangen(t *testing.T) {
	pakete := filepath.Join(t.TempDir(), "pakete")
	darunter := filepath.Join(pakete, "sammlung", "data")
	daneben := filepath.Join(t.TempDir(), "saat")

	if err := os.MkdirAll(darunter, 0o750); err != nil {
		t.Fatal(err)
	}

	// **Ein Name, der nur so anfängt**, liegt nicht darunter.
	aehnlich := pakete + "-zwei"

	suche := Neue(nil).MitPaketen(pakete, "", pakete, darunter, daneben, aehnlich)

	if gegangen := suche.paketwurzeln; !slices.Equal(gegangen, []string{pakete, daneben, aehnlich}) {
		t.Errorf("gegangen wird %q", gegangen)
	}

	// **Der Stand bleibt über jede genannte Wurzel**: ein Paket, das jemand
	// unmittelbar in `data` legt, ändert die Zeit von `data` und nicht die
	// von `pakete`.
	if beobachtet := suche.standwurzeln; !slices.Equal(beobachtet, []string{pakete, darunter, daneben, aehnlich}) {
		t.Errorf("beobachtet wird %q", beobachtet)
	}
}

// **Liegt die tiefere Wurzel vorn, bleibt sie**: die Reihenfolge ist die
// Aussage — was sie trägt, gilt vor dem Rest des Baums darüber.
func TestEineTiefereWurzelVornBleibt(t *testing.T) {
	pakete := filepath.Join(t.TempDir(), "pakete")
	darunter := filepath.Join(pakete, "data")

	if gegangen := Neue(nil).MitPaketen(darunter, pakete).paketwurzeln; !slices.Equal(gegangen, []string{darunter, pakete}) {
		t.Errorf("gegangen wird %q", gegangen)
	}
}

// **Ein Verweis dazwischen hält den Gang auf** — `filepath.WalkDir` folgt
// keinem. Eine Wurzel hinter einem Verweis liegt dem Namen nach darunter und
// wird vom Gang darüber trotzdem nie erreicht.
func TestEineWurzelHinterEinemVerweisBleibt(t *testing.T) {
	pakete := t.TempDir()
	woanders := t.TempDir()

	if err := os.MkdirAll(filepath.Join(woanders, "data"), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(woanders, filepath.Join(pakete, "verweis")); err != nil {
		t.Fatal(err)
	}

	dahinter := filepath.Join(pakete, "verweis", "data")

	if gegangen := Neue(nil).MitPaketen(pakete, dahinter).paketwurzeln; !slices.Equal(gegangen, []string{pakete, dahinter}) {
		t.Errorf("gegangen wird %q", gegangen)
	}
}
