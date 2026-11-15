package service_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Drei Vorlagen desselben Spiels lesen das Spiel einmal** — E03
// (#2240).
//
// Jede Ausfuhr über eine Vorlage fragt zuerst die Liste, und die Liste holte
// für **jede** Vorlage das Spielpaket samt seinen Modulen neu: im vollen Pack
// zweiundzwanzig Vorlagen, für ein `magic`-Set jedes Mal `magic.mse-game` —
// rund 350 MB Zuteilung je Ausfuhr der 70 Archenemy-Karten, für eine Auskunft,
// die für alle dieselbe ist.
//
// **Belegt wird es an der Auskunft selbst**: die drei Vorlagen tragen
// dasselbe Spielmodell, nicht drei gleiche. Gemerkt wird nur innerhalb eines
// Rufs — „nicht gemerkt" über Rufe hinweg gilt weiter (siehe [Vorlagen]).
func TestDreiVorlagenDesselbenSpielsLesenDasSpielEinmal(t *testing.T) {
	wurzel := t.TempDir()

	spiel := filepath.Join(wurzel, "probe.mse-game")
	legeVerzeichnis(t, spiel)
	schreibeDatei(t, filepath.Join(spiel, "game"), "mse version: 2.0.0\nshort name: Probe\n"+
		"init script:\n\tgruss := \"vom Spiel\"\n"+
		"card field:\n\ttype: text\n\tname: name\ncard field:\n\ttype: text\n\tname: rule text\n"+
		"card field:\n\ttype: text\n\tname: flavor text\n")

	for i := range 3 {
		paket := filepath.Join(wurzel, fmt.Sprintf("probe-%d.mse-export-template", i))
		legeVerzeichnis(t, paket)
		schreibeDatei(t, filepath.Join(paket, "export-template"), fmt.Sprintf(
			"mse version: 2.0.0\ngame: probe\nshort name: Probe %d\nfull name: Probe %d\n"+
				"file type: *.txt|*.txt\nscript:\n\tgruss\n", i, i))
	}

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Game: "probe"}}}

	vorlagen, befunde, err := service.NeueVorlagenliste(quelle, wurzel).Vorlagen(t.Context(), 7)
	if err != nil || len(vorlagen) != 3 {
		t.Fatalf("%d Vorlagen, %v, %v", len(vorlagen), befunde, err)
	}

	for i := range vorlagen {
		if vorlagen[i].Spielskript == "" || len(vorlagen[i].Kartenfelder) != 3 {
			t.Fatalf("Vorlage %d: Spielskript %q, %d Kartenfelder", i, vorlagen[i].Spielskript, len(vorlagen[i].Kartenfelder))
		}

		if &vorlagen[i].Kartenfelder[0] != &vorlagen[0].Kartenfelder[0] {
			t.Errorf("Vorlage %d trägt ein eigenes Spielmodell — das Spiel wurde noch einmal gelesen", i)
		}

		// **Gekappt**, damit ein `append` an einer Vorlage nicht in die
		// Felder der anderen schreibt.
		if cap(vorlagen[i].Kartenfelder) != len(vorlagen[i].Kartenfelder) {
			t.Errorf("Vorlage %d: Kapazität %d bei Länge %d", i, cap(vorlagen[i].Kartenfelder), len(vorlagen[i].Kartenfelder))
		}
	}
}
