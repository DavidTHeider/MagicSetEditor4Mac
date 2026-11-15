// **Der Probelauf an der Kommandozeile** — UXR-F7
// (#2015).
//
// # Derselbe Aufruf, ein Wort mehr
//
//	mse bogen <set> [<ziel.pdf>] --probe
//	mse vorlagen <set> ausfuehren <paket> [<ziel> [feld=wert…]] --probe
//
// **Ein Weg, zwei Ausgänge**: `--probe` geht durch denselben Dienst wie der
// Aufruf ohne das Wort und hört dort auf, wo geschrieben würde. Wer einen Befehl
// absichern will, hängt ein Wort an — dieselben Fehler mit demselben Code und
// demselben Satz, und nichts liegt danach auf der Platte.
//
// Einordnung nach §16: **erweiternd** — ohne das Wort tut jeder Aufruf, was er
// tat.
//
// # Das Ziel ist freiwillig
//
// Ein Probelauf schreibt nichts, also braucht er keinen Ort. Wer einen nennt,
// bekommt die Beigaben unter dem Namen, den der Lauf ihnen gäbe; wer keinen
// nennt, den Namen, den das Fenster vorschlägt.

package main

import (
	"context"
	"database/sql"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probewort ist das freiwillige Wort von `bogen` und `vorlagen … ausfuehren`.
const probewort = "--probe"

// probeAus nimmt `--probe` aus den Argumenten und sagt, ob es dastand.
func probeAus(rest []string) (uebrig []string, probe bool) {
	if !slices.Contains(rest, probewort) {
		return rest, false
	}

	return slices.DeleteFunc(slices.Clone(rest), func(wort string) bool {
		return wort == probewort
	}), true
}

// bogenprobeantwort sagt, wie ein Bogen würde.
type bogenprobeantwort struct {
	// Probelauf unterscheidet die Antwort von der eines gelegten Bogens.
	Probelauf bool `json:"probelauf"`

	Karten          int `json:"karten"`
	Seiten          int `json:"seiten"`
	Spalten         int `json:"spalten"`
	Zeilen          int `json:"zeilen"`
	JeSeite         int `json:"jeSeite"`
	AufLetzterSeite int `json:"aufLetzterSeite"`
	Bildbreite      int `json:"bildbreite"`

	Seitenbreite float64 `json:"seitenbreite"`
	Seitenhoehe  float64 `json:"seitenhoehe"`
	Kartenbreite float64 `json:"kartenbreite"`
	Kartenhoehe  float64 `json:"kartenhoehe"`
	Anschnitt    float64 `json:"anschnitt"`
	Rand         float64 `json:"rand"`

	DPI           int  `json:"dpi"`
	Schnittmarken bool `json:"schnittmarken"`
}

// bogenprobebefehl prüft die Argumente von `bogen … --probe` und führt ihn aus.
func bogenprobebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 || len(rest) > 2 {
		an.klage("bogen --probe braucht die Setkennung und darf ein Ziel nennen.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return bogenProbelauf(ctx, db, an, setID)
	})
}

// bogenProbelauf sagt, wie der Bogen würde, den `bogen` legte.
func bogenProbelauf(ctx context.Context, db *sql.DB, an *schreiber, setID int64) int {
	werk, err := app.Zeichennaht(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	probe, fehler := werk.ProbeBogen(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		// **Derselbe Code wie beim Legen**: ein unbrauchbares Maß ist ein
		// Bedienfehler, ein fehlendes Set ein fachliches Nein.
		return vorlagenCode(fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, bogenprobeantwort{
			Probelauf: true,
			Karten:    probe.Karten, Seiten: probe.Seiten, Spalten: probe.Spalten,
			Zeilen: probe.Zeilen, JeSeite: probe.JeSeite,
			AufLetzterSeite: probe.AufLetzterSeite, Bildbreite: probe.Bildbreite,
			Seitenbreite: probe.Seitenbreite, Seitenhoehe: probe.Seitenhoehe,
			Kartenbreite: probe.Kartenbreite, Kartenhoehe: probe.Kartenhoehe,
			Anschnitt: probe.Anschnitt, Rand: probe.Rand,
			DPI: probe.DPI, Schnittmarken: probe.Schnittmarken,
		})
	}

	an.sag("Probelauf — es wird nichts geschrieben.\n")
	an.sag("%d Karten auf %d Seiten: %d je Seite (%d Spalten × %d Zeilen), "+
		"auf der letzten %d.\n",
		probe.Karten, probe.Seiten, probe.JeSeite, probe.Spalten, probe.Zeilen,
		probe.AufLetzterSeite)
	an.sag("Papier %g × %g mm, Karte %g × %g mm, %d dpi, Rand %g mm, Anschnitt %g mm%s.\n",
		probe.Seitenbreite, probe.Seitenhoehe, probe.Kartenbreite, probe.Kartenhoehe,
		probe.DPI, probe.Rand, probe.Anschnitt, schnittmarkenwort(probe.Schnittmarken))

	return an.abschluss()
}

// schnittmarkenwort hängt an, ob Schnittmarken gesetzt würden.
func schnittmarkenwort(marken bool) string {
	if marken {
		return ", mit Schnittmarken"
	}

	return ", ohne Schnittmarken"
}

// vorlagenprobeantwort sagt, was eine Exportvorlage ergäbe.
type vorlagenprobeantwort struct {
	// Probelauf unterscheidet die Antwort von der einer Ausfuhr.
	Probelauf bool `json:"probelauf"`

	Karten        int      `json:"karten"`
	Dateiname     string   `json:"dateiname"`
	Bytes         int      `json:"bytes"`
	Beigaben      int      `json:"beigaben"`
	BeigabenBytes int      `json:"beigabenBytes"`
	Verzeichnis   string   `json:"verzeichnis"`
	Geschaetzt    bool     `json:"geschaetzt"`
	Befunde       []string `json:"befunde"`
}

// vorlagenprobebefehl prüft die Argumente von `vorlagen … ausfuehren --probe`.
//
// **Ohne Ziel, wenn nach dem Paketnamen nichts mehr steht**: ein Wort danach
// ist das Ziel, und was hinter dem Ziel steht, sind `feld=wert`-Paare — wie
// beim Lauf.
func vorlagenprobebefehl(
	ctx context.Context, bestandsort string, an *schreiber, setID int64, rest []string,
) int {
	if len(rest) < 3 || rest[1] != "ausfuehren" {
		an.klage("--probe gibt es bei »ausfuehren«: vorlagen <set> ausfuehren <paket> "+
			"[<ziel> [feld=wert…]] --probe.\n%s", hilfe)

		return CodeBedienung
	}

	ziel := ""
	if len(rest) >= 4 {
		ziel = rest[3]
	}

	var paare []string
	if len(rest) > 4 {
		paare = rest[4:]
	}

	wahl, brauchbar := wahlAus(an, paare)
	if !brauchbar {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return vorlageProbelauf(ctx, db, an, setID, rest[2], ziel, wahl)
	})
}

// vorlageProbelauf sagt, was `vorlagen … ausfuehren` ergäbe.
func vorlageProbelauf(
	ctx context.Context, db *sql.DB, an *schreiber, setID int64, paket, ziel string,
	wahl map[string]string,
) int {
	weg, err := app.VorlagenwegUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	probe, fehler := weg.Probelauf(ctx, setID, paket, ziel, wahl)
	if fehler != nil {
		an.klageUeber(fehler)

		// **Die Befunde reisen mit** (#209), wie beim Lauf.
		for _, befund := range probe.Befunde {
			an.klage("  %s", befund)
		}

		return vorlagenCode(fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, vorlagenprobeantwort{
			Probelauf: true,
			Karten:    probe.Karten, Dateiname: probe.Dateiname, Bytes: probe.Bytes,
			Beigaben: probe.Beigaben, BeigabenBytes: probe.BeigabenBytes,
			Verzeichnis: probe.Verzeichnis, Geschaetzt: probe.Geschaetzt,
			Befunde: ohneNilTexte(probe.Befunde),
		})
	}

	an.sag("Probelauf — es wird nichts geschrieben.\n")
	an.sag("%d Karten ergeben %s mit %d Bytes.\n", probe.Karten, probe.Dateiname, probe.Bytes)

	if probe.Beigaben > 0 {
		an.sag("Daneben %d Dateien in %s (bekannt: %d Bytes%s).\n",
			probe.Beigaben, probe.Verzeichnis, probe.BeigabenBytes, geschaetztwort(probe.Geschaetzt))
	}

	for _, befund := range probe.Befunde {
		an.sag("  %s\n", befund)
	}

	return an.abschluss()
}

// geschaetztwort sagt, dass die Summe eine Untergrenze ist.
func geschaetztwort(geschaetzt bool) string {
	if geschaetzt {
		return ", mindestens"
	}

	return ""
}
