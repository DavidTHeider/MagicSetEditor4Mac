package main

// **Die Partieverwaltung an der Kommandozeile** — PLAY-03.AK-05
// (#1711).
//
// # Vier Befehle, und keiner davon ist ein Schalter
//
// **Löschen, kopieren, ausführen, einlesen** — dieselben vier Wege wie im
// Fenster (§16). Was hier anders ist als dort, ist der **Ort**: das Fenster
// fragt mit einem Dialog danach, hier steht er als Argument. Das ist der
// ganze Unterschied.
//
// # Und keiner trägt eine Sitzkennung
//
// **Weil sie die Partie als Ganzes betreffen und nicht ihre Sicht.** Wer
// eine Partie löscht, löscht sie für jeden; wer sie ausführt, bekommt den
// **vollen** Zustand und nicht die Sicht eines Sitzes — deshalb steht der
// Weg unter derselben Erlaubnis wie der Bestand selbst.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// partieloeschenbefehl nimmt eine Partie aus dem Bestand.
func partieloeschenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("partie-loeschen braucht die Partiekennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return verwaltungAusgeben(an, flaeche.PartieLoeschen(ctx, rest[0]),
				"Partie %s geloescht.%s\n", rest[0])
		})
}

// partiekopierenbefehl legt eine Partie als neuen Übungszustand daneben.
func partiekopierenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("partie-kopieren braucht die Partiekennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return verwaltungAusgeben(an, flaeche.PartieKopieren(ctx, rest[0]),
				"Partie %s kopiert nach %s.\n", rest[0])
		})
}

// partieausfuhrbefehl legt eine Partie als Datei ab.
func partieausfuhrbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("partie-ausfuhr braucht die Partiekennung und einen "+
			"Zielpfad.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return verwaltungAusgeben(an,
				flaeche.PartieAusfuhrSchreiben(ctx, rest[0], rest[1]),
				"Partie %s ausgefuehrt.%s\n", rest[0])
		})
}

// partieeinfuhrbefehl liest eine Partie aus einer Datei.
func partieeinfuhrbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("partie-einfuhr braucht den Pfad der Datei.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return verwaltungAusgeben(an, flaeche.PartieEinfuhrLesen(ctx, rest[0]),
				"Partie aus %s eingelesen als %s.\n", rest[0])
		})
}

// verwaltungAusgeben schreibt den Ausgang — samt Befunden.
//
// **Die Befunde gehen mit** (§8): ein fremdes Feld in der Datei macht die
// Partie nicht unbenutzbar, und es wird trotzdem **genannt**. Wer sie
// verschlucken ließe, ließe jemanden glauben, es sei alles angekommen.
//
// **Alle vier Wege gehen hier durch**, auch die beiden, deren Satz die neue
// Kennung gar nicht braucht: sonst hätten zwei von vier kein `--json`, und
// eine Fläche, die an zwei Wegen anders antwortet als an zwei anderen, ist
// zwei Flächen (§16).
func verwaltungAusgeben(
	an *schreiber, antwort app.PartieverwaltungAntwort, satz, erstes string,
) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag(satz, erstes, antwort.Kennung)

	for _, befund := range antwort.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	return an.abschluss()
}
