package main

// **Wer im selben Netz da ist, an der Kommandozeile** — Werkstück N2 von
// Phase 5.4
// (#1455).
//
//	mse nachbarn    wer sich im selben Netz meldet
//
// # Warum es hier nur das Lesen gibt
//
// **Weil eine Ankündigung einen Port braucht**, und den öffnet nur das
// Fenster. Ein Befehl, der ankündigte, müsste bis ⌃C laufen — und wäre damit
// etwas anderes als ein Befehl, der etwas tut und endet. Die Begründung steht
// in `docs/FAEHIGKEITEN.md`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// nachbarnbefehl fragt, wer sich meldet.
func nachbarnbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("nachbarn nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return nachbarnSchreiben(an, app.PeerflaecheUeber(db).Nachbarn(ctx))
	})
}

// nachbarnSchreiben gibt die Antwort aus — eine Zeile je Nachbar, dazu
// eingerückt, was an ihm zu übernehmen wäre.
func nachbarnSchreiben(an *schreiber, antwort app.NachbarnAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, nachbar := range antwort.Nachbarn {
		stand := nachbar.Stand
		if !nachbar.Bekannt {
			stand = "unbekannt"
		}

		// **Die Merkadresse steht dabei, wo sie eine andere ist**
		// (#1633). Sie zweimal zu zeigen, wo sie dieselbe ist, wäre
		// eine Spalte, die nichts sagt.
		gemerkt := nachbar.Merkadresse
		if gemerkt == nachbar.Adresse {
			gemerkt = "-"
		}

		an.sag("%s\t%s\t%s\t%s\t%s\n", stand, nachbar.Fingerabdruck,
			nachbar.Adresse, gemerkt, nachbar.Name)

		// **Eine Adresse zum Übernehmen** (NET-02.NAT-02, #1706) —
		// eingerückt hinter der unveränderten ersten Zeile, mit dem,
		// was die Kontaktliste bisher führt (`-` für nichts).
		if nachbar.AdresseGeaendert {
			bisher := nachbar.GemerkteAdresse
			if bisher == "" {
				bisher = "-"
			}

			an.sag("\tadresse-bisher\t%s\n", bisher)
		}
	}

	// **Die verworfenen Antworten stehen dabei** — sonst sähe „niemand
	// gefunden" wie „nichts gehört" aus.
	if antwort.Verworfen > 0 {
		an.sag("verworfen\t%d\n", antwort.Verworfen)
	}

	return an.abschluss()
}
