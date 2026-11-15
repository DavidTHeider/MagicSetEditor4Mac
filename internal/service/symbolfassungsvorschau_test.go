package service_test

// **Ein Symbol in den Fassungen eines Sets** — DOM-02.AK-02
// (#1697):
// die Größen- und Farbvorschau. Die Fassungen kommen aus Spiel und Stil des
// Sets, und gezeichnet wird über denselben Bildknoten wie auf der Karte.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fassungswurzel legt ein Spiel mit zwei Fassungen und einen Stil mit zwei
// weiteren an — eine davon heißt wie eine des Spiels.
func fassungswurzel(t *testing.T) string {
	t.Helper()

	wurzel := t.TempDir()

	for datei, inhalt := range map[string]string{
		"fass.mse-game/game": "mse version: 2.0.0\nshort name: Fass\n" +
			"set field:\n\ttype: symbol\n\tname: symbol\n" +
			"default set style:\n\tsymbol:\n" +
			"\t\tvariation:\n\t\t\tname: common\n\t\t\tborder radius: 0.10\n\t\t\tfill type: solid\n" +
			"\t\t\tfill color: rgb(0,0,0)\n\t\t\tborder color: rgb(255,255,255)\n" +
			"\t\tvariation:\n\t\t\tname: rare\n\t\t\tfill type: linear gradient\n" +
			"\t\t\tfill color 1: rgb(255,200,0)\n\t\t\tfill color 2: rgb(120,80,0)\n",
		"fass-schlicht.mse-style/style": "mse version: 2.0.0\ngame: fass\nshort name: schlicht\n" +
			"set info style:\n\tsymbol:\n" +
			"\t\tvariation:\n\t\t\tname: invertedcommon\n\t\t\tfill color: rgb(255,255,255)\n" +
			"\t\tvariation:\n\t\t\tname: common\n\t\t\tfill color: rgb(9,9,9)\n",
	} {
		pfad := filepath.Join(wurzel, datei)
		if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	return wurzel
}

func fassungswaehler(t *testing.T, kopf domain.SetHead) *service.Stilwaehler {
	t.Helper()

	return service.NeuerStilwaehler(&koepfequelle{koepfe: []domain.SetHead{kopf}},
		saatkatalog(), &stilablage{}, &feldliste{}, fassungswurzel(t))
}

func fassungsnamen(fassungen []msepaket.Symbolfassung) []string {
	namen := make([]string, 0, len(fassungen))
	for _, fassung := range fassungen {
		namen = append(namen, fassung.Name)
	}

	return namen
}

func TestDieFassungenKommenAusSpielUndStil(t *testing.T) {
	waehler := fassungswaehler(t, domain.SetHead{ID: 1, Game: "fass", Style: "fass-schlicht"})

	fassungen, err := waehler.Symbolfassungen(t.Context(), 1)
	if err != nil || !slices.Equal(fassungsnamen(fassungen), []string{"common", "rare", "invertedcommon", "common"}) {
		t.Fatalf("die Fassungen: %v, %v", fassungsnamen(fassungen), err)
	}

	// **Ohne Spiel und Stil gibt es keine** — und das ist kein Fehler: die
	// Vorschau sagt es dann selbst.
	ohne := fassungswaehler(t, domain.SetHead{ID: 1, Game: "gibtsnicht"})
	if fassungen, err := ohne.Symbolfassungen(t.Context(), 1); err != nil || len(fassungen) != 0 {
		t.Errorf("ohne Pakete: %v, %v", fassungen, err)
	}

	if _, err := waehler.Symbolfassungen(t.Context(), 9); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ohne Set: %v", err)
	}
}

func TestDieVorschauZeichnetJedeFassungEinmal(t *testing.T) {
	bibliothek, _, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)
	bibliothek.MitFassungen(fassungswaehler(t, domain.SetHead{ID: 1, Game: "fass", Style: "fass-schlicht"}))

	plaene, err := bibliothek.Fassungsplaene(t.Context(), krone.ID, 1, 16)
	if err != nil {
		t.Fatalf("die Pläne: %v", err)
	}

	namen := make([]string, 0, len(plaene))
	for _, plan := range plaene {
		namen = append(namen, plan.Name)
	}

	// **Die des Spiels gewinnt bei gleichem Namen** — wie auf der Karte
	// (`ErsteFassung`).
	if !slices.Equal(namen, []string{"common", "rare", "invertedcommon"}) {
		t.Fatalf("die Namen: %v", namen)
	}

	gewoehnlich := plaene[0].Plan
	ebene := gewoehnlich.Ebenen[0]

	if gewoehnlich.Breite != 16 || ebene.Art != render.ArtSymbol || ebene.Bild.Knotenart != "symbol_variation" ||
		len(ebene.Bild.Symbol) != len(msepaket.Standardsymbol().Teile) || ebene.Bild.Zahlen["border_radius"] != 0.10 {
		t.Errorf("der Plan: %+v", gewoehnlich)
	}

	if schwarz := ebene.Bild.Farben["fill_color"]; schwarz != (render.Farbe{A: 255}) {
		t.Errorf("die Füllfarbe: %+v", schwarz)
	}

	if len(plaene[1].Plan.Ebenen[0].Bild.Farben) != 2 {
		t.Errorf("die Verlaufsfarben: %+v", plaene[1].Plan.Ebenen[0].Bild.Farben)
	}
}

func TestDieVorschauOhneFassungenUndMitFehlern(t *testing.T) {
	bibliothek, speicher, _ := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	// **Ohne Quelle der Fassungen** gibt es keine — leer, nicht `nil`.
	if plaene, err := bibliothek.Fassungsplaene(t.Context(), krone.ID, 1, 16); err != nil || plaene == nil || len(plaene) != 0 {
		t.Errorf("ohne Quelle: %v, %v", plaene, err)
	}

	bibliothek.MitFassungen(fassungswaehler(t, domain.SetHead{ID: 1, Game: "fass", Style: "fass-schlicht"}))

	if _, err := bibliothek.Fassungsplaene(t.Context(), 99, 1, 16); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ohne Symbol: %v", err)
	}

	if _, err := bibliothek.Fassungsplaene(t.Context(), krone.ID, 9, 16); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ohne Set: %v", err)
	}

	speicher.symbole[0].Inhalt = "part:\n" + string(make([]byte, msepaket.GrenzeSymboldatei))
	if _, err := bibliothek.Fassungsplaene(t.Context(), krone.ID, 1, 16); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unlesbar: %v", err)
	}
}
