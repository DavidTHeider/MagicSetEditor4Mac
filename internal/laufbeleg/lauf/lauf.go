package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// lauf ist das Programm. Es endet mit **0** für einen grünen und aktuellen
// Beleg, mit **1** für jeden anderen und mit **2**, wenn es nicht ausführen
// kann, was verlangt ist.
func lauf(argumente []string, aus, fehlerAus io.Writer, kasten werkzeugkasten) int {
	schalter := flag.NewFlagSet("lauf", flag.ContinueOnError)
	schalter.SetOutput(fehlerAus)

	planpfad := schalter.String("plan", "", "der Plan des Laufs, etwa docs/belege/fremdpakete.plan.json")
	pruefen := schalter.Bool("pruefen", false, "nicht messen, sondern den Beleg gegen den heutigen Stand halten")
	ohneEingaben := schalter.Bool("ohne-eingaben", false,
		"mit -pruefen: die Eingaben nicht neu rechnen — der Beleg heißt dann höchstens »ohne Eingaben«")

	if schalter.Parse(argumente) != nil {
		return 2
	}

	ausgabe := &anzeige{aus: aus, fehler: fehlerAus}

	if *planpfad == "" {
		ausgabe.klage("lauf: -plan fehlt — welcher Lauf soll es sein?\n")

		return 2
	}

	plan, modul, fehler := planLesen(*planpfad)
	if fehler != nil {
		ausgabe.klage("lauf: %v\n", fehler)

		return 2
	}

	ausgabe.modul = modul

	var ergebnis int
	if *pruefen {
		ergebnis = pruefe(context.Background(), plan, *planpfad, *ohneEingaben, ausgabe, kasten)
	} else {
		ergebnis = miss(context.Background(), plan, *planpfad, ausgabe, kasten)
	}

	// **Eine Ausgabe, die nicht ankam, ist ein Fehler des Laufs** — der
	// Beleg steht dann zwar da, aber wer auf den Bildschirm sah, sah ihn nicht.
	if ausgabe.erster != nil {
		ausgabe.klage("lauf: die Ausgabe ließ sich nicht schreiben: %v\n", ausgabe.erster)

		return 2
	}

	return ergebnis
}

// anzeige sind die beiden Ströme, samt dem Modul, das eine Zeile kürzt —
// dieselbe Gestalt wie der `schreiber` in `cmd/mse`.
type anzeige struct {
	aus, fehler io.Writer
	modul       string
	erster      error
}

// sag schreibt auf die Ausgabe und merkt sich den ersten Fehler.
func (a *anzeige) sag(format string, werte ...any) {
	if _, fehler := fmt.Fprintf(a.aus, format, werte...); fehler != nil && a.erster == nil {
		a.erster = fehler
	}
}

// klage schreibt auf den Fehlerstrom.
//
// **Ein Fehler beim Klagen wird nicht gemerkt** — wer sich darüber beschweren
// wollte, bräuchte einen dritten Strom.
func (a *anzeige) klage(format string, werte ...any) {
	_, _ = fmt.Fprintf(a.fehler, format, werte...) //nolint:errcheck // siehe darüber
}

// planLesen liest den Plan gegen die markierten Fälle dieses Baums.
//
// **Der Tag steht im Plan, und die markierten Fälle hängen am Tag** — darum
// wird er zuerst herausgelesen. Was dabei schiefgeht, meldet `LiesPlan`.
func planLesen(pfad string) (laufbeleg.Plan, string, error) {
	inhalt, fehler := os.ReadFile(pfad) // #nosec G304 -- den Plan nennt, wer den Lauf startet
	if fehler != nil {
		return laufbeleg.Plan{}, "", fmt.Errorf("den Plan %s lesen: %w", pfad, fehler)
	}

	var kopf struct {
		Tag string `json:"tag"`
	}

	_ = json.Unmarshal(inhalt, &kopf) //nolint:errcheck // was hier nicht zu lesen ist, sagt LiesPlan

	modul, fehler := laufbeleg.Modul(".")
	if fehler != nil {
		return laufbeleg.Plan{}, "", fehler
	}

	markiert, fehler := laufbeleg.Markierte(".", kopf.Tag)
	if fehler != nil {
		return laufbeleg.Plan{}, "", fehler
	}

	plan, fehler := laufbeleg.LiesPlan(bytes.NewReader(inhalt), modul, markiert)

	return plan, modul, fehler
}
