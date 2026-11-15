// **Eine Vorgabe sieht die anderen Stiloptionen** (#789).
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`default:` wird mit `styling` ausgewertet.**
//
// # Was MSE dort hineinlegt
//
//	ctx.setVariable(SCRIPT_VAR_set,        …);
//	ctx.setVariable(SCRIPT_VAR_game,       …);
//	ctx.setVariable(SCRIPT_VAR_stylesheet, …);
//	ctx.setVariable(SCRIPT_VAR_card_style, …);
//	ctx.setVariable(SCRIPT_VAR_styling,    to_script(&set.stylingDataFor(*stylesheet)));
//
// `src/script/script_manager.cpp`, Zeilen 30–61, Fassung f9d9356 vom
// 2021-07-06. **`styling` sind die gespeicherten Werte des Sets**, nicht die
// gerechneten — und genau die liegen hier vor.
func TestEineVorgabeSiehtDieAnderenStiloptionen(t *testing.T) {
	feld := flaeche("name")
	feld.Schrift = msepaket.Schrift{
		Name: "Helvetica", Groesse: "12", Farbe: "{ styling.schriftfarbe }",
	}

	stil := bildstil(feld)
	stil.Stiloptionen = []domain.StyleField{
		{Name: "hintergrund", Kind: domain.KindChoice, Choices: []string{"hell", "dunkel"}},

		// **Die Vorgabe liest die Nachbaroption.** Ohne `styling` in der
		// Umgebung liefe sie in einen Fehler und fiele auf `initial:` zurück
		// — falsch, aber unsichtbar.
		{
			Name: "schriftfarbe", Kind: domain.KindColor,
			Default: `{ if styling.hintergrund == "dunkel" then "white" else "black" }`,
			Initial: "red",
		},
	}

	plan, befunde := gedeutetMit(t, stilkarte("rot"), stil)

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// `hintergrund` steht ungesetzt auf seiner ersten Wahl, „hell" — die
	// Vorgabe muss daraus Schwarz machen und nicht das rote `initial:`.
	schwarz := render.Farbe{A: 255}
	if gesetzt := textebene(t, plan).Farbe; gesetzt != schwarz {
		t.Errorf("Farbe %+v, erwartet Schwarz — die Vorgabe hat `styling` gesehen",
			gesetzt)
	}
}
