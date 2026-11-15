package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"runtime"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// pruefe hält den Beleg gegen den heutigen Stand (AK-04).
//
// **Gemessen wird, was der Plan unter `misst` nennt — und der Plan selbst**:
// wer eine Ausnahme ändert, ändert, was gemessen wird. Verglichen wird der
// Inhalt, nicht der Commit ([inhaltLesen]) — **samt Arbeitsbaum und
// Pflichteingaben** ([heuteLesen]).
func pruefe(ctx context.Context, plan laufbeleg.Plan, planpfad string, ohneEingaben bool,
	zu *anzeige, kasten werkzeugkasten,
) int {
	datei, fehler := os.Open(plan.Beleg) // #nosec G304 -- den Ort nennt der Plan
	if errors.Is(fehler, fs.ErrNotExist) {
		zu.sag("%s: kein Beleg unter %s\n", plan.Lauf, plan.Beleg)

		return 1
	}

	if fehler != nil {
		zu.klage("lauf: %v\n", fehler)

		return 2
	}

	beleg, fehler := laufbeleg.Lesen(datei)
	_ = datei.Close() //nolint:errcheck // nur gelesen

	if fehler != nil {
		zu.klage("lauf: %s: %v\n", plan.Beleg, fehler)

		return 2
	}

	heute, fehler := heuteLesen(ctx, kasten, plan, planpfad, ohneEingaben)
	if fehler != nil {
		zu.klage("lauf: was der Lauf misst, ist nicht zu lesen: %v\n", fehler)

		return 2
	}

	einordnung, grund := laufbeleg.Ordne(beleg, heute, kasten.jetzt(), plan.Frist)

	zu.sag("%s: %s · %s — %s\n", plan.Lauf, beleg.Status, einordnung, grund)

	if beleg.Status == laufbeleg.Bestanden && einordnung == laufbeleg.Aktuell {
		return 0
	}

	return 1
}

// heuteLesen erhebt den heutigen Stand: den Inhalt im Index, was der
// Arbeitsbaum dort anders hat, und die Pflichteingaben.
//
// **Die Eingaben werden neu gerechnet, wie beim Messen** — über Namen und
// Inhalt ([laufbeleg.Erhebe]). Das kostet so viel Zeit, wie die Eingaben
// groß sind; eine Prüfung, die nur ihre Zahl hielte, sähe ein neu
// gerastertes Referenzbild derselben Größe nicht. **Mit `ohneEingaben` nicht**
// — und dann sagt die Einordnung es.
func heuteLesen(ctx context.Context, kasten werkzeugkasten, plan laufbeleg.Plan,
	planpfad string, ohneEingaben bool,
) (laufbeleg.Heute, error) {
	inhalt, fehler := inhaltLesen(ctx, kasten, plan.Misst, planpfad)
	if fehler != nil {
		return laufbeleg.Heute{}, fehler
	}

	argumente := append(append([]string{"status", "--porcelain", "--"}, plan.Misst...), planpfad)

	offen, fehler := kasten.git(ctx, argumente...)
	if fehler != nil {
		return laufbeleg.Heute{}, fehler
	}

	heute := laufbeleg.Heute{
		Inhalt: inhalt, OhneEingaben: ohneEingaben,
		Werkzeuge: map[string]string{
			"go": kasten.version(ctx), "plattform": runtime.GOOS + "/" + runtime.GOARCH,
		},
	}

	if !ohneEingaben {
		heute.Eingaben, _ = erhebe(plan, kasten)
		heute.Schalter = schalterLesen(plan, kasten)
	}

	if offen != "" {
		heute.Offen = strings.Split(offen, "\n")
	}

	return heute, nil
}
