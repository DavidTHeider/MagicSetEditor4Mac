package service_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ausführer und Mittel in einer Hand** — für die Tests, nicht für den Betrieb.
//
// Seit #736
// nimmt `Fuehre` die Fähigkeiten **einer** Ausfuhr als Wert entgegen, statt sie
// vorher am geteilten Ausführer zu setzen. Das ist im Betrieb genau richtig —
// in einem Test, der beides ohnehin in einem Zug baut, wäre es an jeder
// Aufrufstelle ein zweiter Wert, der nichts erklärt.
//
// **Ein Bündel und keine Hilfsfunktion je Fall**: so bleibt sichtbar, dass die
// Mittel zum Lauf gehören und nicht zum Ausführer.
type ausfuhrprobe struct {
	ausfuehrer *service.Vorlagenausfuehrer
	mittel     service.Ausfuhrmittel
}

func (a ausfuhrprobe) Fuehre(
	ctx context.Context, vorlage msepaket.Exportvorlage,
	karten []*domain.CardRow, setwerte []domain.Value, wahl map[string]string,
) (string, []domain.Befund, error) {
	return a.ausfuehrer.Fuehre(ctx, a.mittel, vorlage, karten, setwerte, wahl)
}

// probelauf bündelt einen frischen Ausführer mit den genannten Mitteln.
func probelauf(mittel service.Ausfuhrmittel) ausfuhrprobe {
	return ausfuhrprobe{ausfuehrer: service.NeuerVorlagenausfuehrer(), mittel: mittel}
}
