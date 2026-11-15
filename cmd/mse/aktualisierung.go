package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **`mse aktualisierung [--laden]`** —
// #1100,
// Werkstück 7.
//
// # Was hier anders ist als im Fenster
//
// **Die Einstellung wird nicht gelesen.** Sie verhindert, dass die Anwendung
// **von selbst** ins Netz greift (§16, Regel 6); wer diesen Befehl tippt, hat
// gefragt. Der Unterschied ist die **Aufsicht**, nicht die Fähigkeit —
// dieselbe Unterscheidung wie bei `mse zeichnen` gegenüber
// `schreiben.karte.zeichnen`
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
//
// **Und der Fortschritt geht auf den Fehlerstrom.** Auf dem Ziel steht am
// Ende der Pfad und sonst nichts, damit ein Skript ihn weiterreichen kann.

// ladeflagge ist das Argument, das aus dem Nachsehen ein Holen macht.
const ladeflagge = "--laden"

func aktualisierungsbefehl(ctx context.Context, an *schreiber, rest []string) int {
	laden := false

	for _, argument := range rest {
		if argument != ladeflagge {
			an.klage("aktualisierung kennt nur %s.\n%s", ladeflagge, hilfe)

			return CodeBedienung
		}

		laden = true
	}

	auskunft, err := app.NeuesteFassung(ctx)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	an.sag("laufend %s, freigegeben %s\n", auskunft.Laufend, auskunft.Freigegeben)

	if !auskunft.Neuer {
		an.sag("Das ist die laufende Fassung.\n")

		return CodeGut
	}

	an.sag("Eine neuere Fassung ist da: %s\n%s\n",
		auskunft.Freigegeben, auskunft.Seite)

	if !laden {
		return CodeGut
	}

	return holen(ctx, an)
}

// holen lädt das Abbild und nennt seinen Ort.
func holen(ctx context.Context, an *schreiber) int {
	pfad, err := app.FassungLaden(ctx, meldenAuf(an))
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	an.sag("%s\n", pfad)

	return CodeGut
}

// meldenAuf gibt den Fortschrittsmelder für diesen Schreiber.
//
// **Er meldet je angefangenem Zehntel und nicht je Block.** `io.Copy` ruft mit
// 32 KiB; bei 188 MB wären das rund 6.000 Zeilen, und eine Kommandozeile, die
// ihr eigenes Ergebnis wegscrollt, ist keine Auskunft.
func meldenAuf(an *schreiber) func(gelesen, ganz int64) {
	letztes := -1

	return func(gelesen, ganz int64) {
		if ganz <= 0 {
			return
		}

		zehntel := int(gelesen * 10 / ganz)
		if zehntel == letztes {
			return
		}

		letztes = zehntel

		an.klage("%d %%\n", zehntel*10)
	}
}
