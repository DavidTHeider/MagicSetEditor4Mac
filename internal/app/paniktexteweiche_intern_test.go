package app

// **Durch die Weiche, wie sie im Betrieb schreibt** — Nachlauf über Phase 5
// von 0.9.0, Befund 1 (#2399, Vermerk 0193).
//
// Der erste Fall für JSON schrieb mit einem nackten `slog.NewJSONHandler` und
// ging damit an der Umwandlung vorbei, die im Betrieb zwischen dem Ruf und der
// Zeile steht (`geschwaerzt`). Genau dort lag der Fehler: eine **Gruppe** kam
// als Zeichenkette heraus, `[panik=Kennwort hunter]`, und der Exportfilter
// nahm `Kennwort` und ließ `hunter` stehen. Hier schreibt deshalb die Weiche
// selbst, in beiden Formaten, und gefiltert wird mit dem Filter des Pakets.

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinPaniktextFaelltInJedemFormatUndJederGestalt(t *testing.T) {
	t.Parallel()

	gestalten := []struct {
		name     string
		schreibe func(*slog.Logger)
	}{
		{"als Merkmal", func(l *slog.Logger) { l.Error("brach ab", "panik", "Kennwort hunter") }},
		{"in einer Gruppe", func(l *slog.Logger) {
			l.Error("brach ab", slog.Group("grund", slog.String("panik", "Kennwort hunter")))
		}},
		{"in einer Gruppe in einer Gruppe", func(l *slog.Logger) {
			l.Error("brach ab", slog.Group("grund", slog.Group("grund", slog.String("panik", "Kennwort hunter"))))
		}},
		{"in einer gebundenen Gruppe", func(l *slog.Logger) { l.WithGroup("grund").Error("brach ab", "panik", "Kennwort hunter") }},
		{"gebunden mit With", func(l *slog.Logger) {
			l.With(slog.Group("grund", slog.String("panik", "Kennwort hunter"))).Error("brach ab")
		}},
		{"mehrzeilig", func(l *slog.Logger) { l.Error("brach ab", "panik", "Kennwort\nhunter") }},
		{"mit Anführung", func(l *slog.Logger) { l.Error("brach ab", "panik", `Kennwort "hunter"`) }},
		{"ohne Leerzeichen", func(l *slog.Logger) { l.Error("brach ab", "panik", "hunter") }},
	}

	for _, format := range []string{domain.ProtokollformatText, domain.ProtokollformatJSON} {
		for _, gestalt := range gestalten {
			t.Run(format+" "+gestalt.name, func(t *testing.T) {
				t.Parallel()

				weiche, _, ausgabe := weicheMitPuffern(t, t.TempDir())
				weiche.Einrichten(domain.Protokollregel{Stufe: domain.ProtokollstufeInfo, Ausgabe: true, Format: format})
				gestalt.schreibe(weiche.Logger())

				// **Am Rechner steht der Wert ganz** — die Positivkontrolle:
				// ohne sie hieße ein fehlendes `hunter` unten auch, dass gar
				// nichts geschrieben wurde.
				roh := ausgabe.String()
				if !strings.Contains(roh, "hunter") {
					t.Fatalf("das Protokoll trägt den Wert nicht:\n%s", roh)
				}

				if gefiltert := paniktexteSchwaerzen(roh); strings.Contains(gefiltert, "hunter") {
					t.Errorf("hinaus ginge:\n%s", gefiltert)
				}
			})
		}
	}
}
