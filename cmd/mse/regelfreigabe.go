package main

// **Die Regelfreigabe an der Kommandozeile** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Warum sie hier steht
//
// **§16**: was das Fenster kann, kann die Kommandozeile. Und das Issue
// verlangt es ausdrücklich — „Turniersperre lässt sich weder über UI noch
// REST/CLI/MCP umgehen". Eine Sperre, die nur das Fenster kennt, ist keine.
//
// # Der Rückgabewert sagt, ob es geht
//
// **`0` für ja, `CodeFachlich` für nein** — dieselbe Form wie bei
// `deck pruefen`. Ein Skript, das eine Partie aufsetzt, fragt vorher und
// bricht ab, statt in einen Fehler zu laufen.

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// regelfreigabebefehl sagt, was durchgesetzt wird und ob es so losgehen darf.
func regelfreigabebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	// **Drei Pflichtangaben, zwei freiwillige** (PLAY-02, #1710): der
	// Modus und die Bereiche, die die Runde selbst übernimmt. Wer sie
	// weglässt, bekommt den Anfängermodus und keine Abwahl.
	if len(rest) < 3 || len(rest) > 5 {
		an.klage("regelfreigabe <spiel> <strenge> <sitze> [modus] [ohne,...]"+
			"\n%s", hilfe)

		return CodeBedienung
	}

	modus, ohne := "", ""

	if len(rest) > 3 {
		modus = rest[3]
	}

	if len(rest) > 4 {
		ohne = rest[4]
	}

	sitze, fehler := strconv.Atoi(rest[2])
	if fehler != nil {
		an.klage("»%s« ist keine Zahl von Sitzen.", rest[2])

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RegelflaecheUeber(db).Regelfreigabe(
			ctx, rest[0], rest[1], sitze, modus, geteilteliste(ohne))
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort.Freigabe); code != 0 {
				return code
			}

			return freigabecode(antwort.Freigabe.Erlaubt)
		}

		schreibeFreigabe(an, antwort.Freigabe)

		return freigabecode(antwort.Freigabe.Erlaubt)
	})
}

// schreibeFreigabe gibt die Freigabe als Zeilen.
//
// **Die Lücken stehen auch bei einem Ja da** — wer „begleitet" bekommt, soll
// sehen, dass „streng" nicht zur Wahl steht.
func schreibeFreigabe(an *schreiber, freigabe app.RegelfreigabeDTO) {
	an.sag("regelwerk\t%s\t%d\twirksam ab %s\n",
		freigabe.Regelwerk, freigabe.Fassung, freigabe.Wirksam)
	an.sag("hoechste\t%s\n", freigabe.Hoechste)

	for _, eintrag := range freigabe.Matrix {
		an.sag("%s\t%s\t%s\n", eintrag.Bereich, eintrag.Grad, eintrag.Quelle)

		for _, grenze := range eintrag.Grenzen {
			an.sag("\tgrenze\t%s\n", grenze)
		}
	}

	for _, hindernis := range freigabe.Hindernisse {
		an.sag("hindernis\t%s\n", hindernis)
	}
}

// freigabecode gibt den Rückgabewert zu einer Freigabe.
func freigabecode(erlaubt bool) int {
	if erlaubt {
		return 0
	}

	return CodeFachlich
}

// geteilteliste zerlegt eine kommagetrennte Angabe.
//
// **Leere Stücke fallen weg**: „zonen," ist ein Bereich und nicht zwei
// (§8) — dieselbe Form wie im Dienst.
func geteilteliste(roh string) []string {
	if strings.TrimSpace(roh) == "" {
		return nil
	}

	stuecke := []string{}

	for _, stueck := range strings.Split(roh, ",") {
		if geglaettet := strings.TrimSpace(stueck); geglaettet != "" {
			stuecke = append(stuecke, geglaettet)
		}
	}

	return stuecke
}
