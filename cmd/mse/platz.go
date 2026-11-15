package main

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **`mse platz [freigeben <posten>]`** — Werkstück 4 von Phase 4.13
// (#1352).
//
// # Warum ohne offenen Bestand
//
// **Weil gezählt werden Dateien, keine Zeilen.** Sie liegen auch dann da, wenn
// der Bestand nicht aufgeht — und genau dann will jemand wissen, was Platz
// belegt. `PlatzflaecheUeber` nimmt deshalb den **Ort** und keine Verbindung.
//
// # Und warum es diesen Befehl gibt
//
// **§16**: was das Fenster kann, können die anderen Flächen. Wer über eine
// Verbindung ohne Bildschirm merkt, dass die Platte voll ist, soll nicht erst
// ein Fenster öffnen müssen.
const freigebenwort = "freigeben"

func platzbefehl(ctx context.Context, an *schreiber, bestandsort string, rest []string) int {
	flaeche := app.PlatzflaecheUeber(bestandsort)

	switch {
	case len(rest) == 0:
		return platzzeigen(an, flaeche.Platz(ctx))

	case len(rest) == 2 && rest[0] == freigebenwort:
		return platzzeigen(an, flaeche.PlatzFreigeben(ctx, rest[1]))

	default:
		an.klage("platz nimmt nichts oder »%s <posten>«.\n%s", freigebenwort, hilfe)

		return CodeBedienung
	}
}

// platzzeigen schreibt die Posten — als Text oder als JSON.
func platzzeigen(an *schreiber, antwort app.PlatzAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(app.AlsFehler(antwort.Fehler))

		return CodeFachlich
	}

	nachladenMelden(an, antwort.Nachladefehler)

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, posten := range antwort.Posten {
		vermerk := ""
		if posten.Raeumbar {
			vermerk = "  (räumbar)"
		}

		an.sag("%-16s %10s  %4d Datei(en)%s\n",
			posten.Posten, inBytes(posten.Bytes), posten.Dateien, vermerk)
	}

	return CodeGut
}

// inBytes schreibt eine Zahl so, wie ein Mensch sie liest.
//
// **Tausender und keine Zweierpotenzen.** Der Finder zeigt „127 MB" für
// dieselbe Datei, und eine Zahl, die von seiner abweicht, sät Zweifel an
// beiden.
func inBytes(zahl int64) string {
	const tausend = 1000

	if zahl < tausend {
		return fmt.Sprintf("%d B", zahl)
	}

	wert := float64(zahl)

	for _, einheit := range []string{"kB", "MB", "GB", "TB"} {
		wert /= tausend

		if wert < tausend {
			return fmt.Sprintf("%.1f %s", wert, einheit)
		}
	}

	return fmt.Sprintf("%.1f PB", wert/tausend)
}
