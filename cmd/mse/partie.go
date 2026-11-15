package main

// **Die Partie an der Kommandozeile** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Sechs Befehle, und keiner davon ist ein Schalter
//
// **Anfangen, ansehen, einen Befehl stellen, pausieren, fortsetzen,
// beenden** — dieselben sechs Wege wie im Fenster (§16). Ein
// `partie-stand <ein|aus>` machte aus drei Entscheidungen eine und nähme
// beim Ausschalten noch eine Sitzkennung entgegen.
//
// # Und jeder trägt eine Sitzkennung
//
// **Weil es die Frage „wer sieht das?" ohne sie nicht gibt.** An der
// Kommandozeile ist sie ein Argument statt eines angemeldeten Fensters — und
// das ist der ganze Unterschied. Wer sie wegließe, bekäme die Sicht eines
// Zuschauers und hielte sie für die eines Spielers.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// partieanlegenbefehl legt eine Partie aus zwei Decks an.
func partieanlegenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	// **Die Sitzprofile zählen nicht als Sitze** (UXR-F21, #2041): wer
	// `--sitzprofil=` mitzählte, fände mit einem Sitz und einem Schalter
	// "genug" und bekäme die Abweisung erst eine Schicht tiefer.
	argumente, profile := trenneSitzprofile(rest)

	// **Die Abwahl zählt ebenso wenig als Sitz** (P1a, #2335).
	argumente, abgewaehlt := trenneAbwahl(argumente)

	if len(argumente) < 3 {
		an.klage("partie-anlegen braucht das Spiel und mindestens zwei "+
			"Sitze der Form <kennung>:<name>:<deck>.\n%s", hilfe)

		return CodeBedienung
	}

	wuensche, fehler := sitzwuenscheAus(argumente[1:])
	if fehler == nil {
		fehler = sitzprofileBinden(wuensche, profile)
	}

	if fehler != nil {
		an.klageUeber(fehler)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return partieAusgeben(an, flaeche.PartieUnterRegeln(
				ctx, argumente[0], "", wuensche,
				app.RegelwerkDTO{Abgewaehlt: abgewaehlt}))
		})
}

// partiezeigenbefehl sagt, wie die Partie für einen Sitz aussieht.
//
// **Rein lesend.**
func partiezeigenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("partie-zeigen braucht die Partiekennung und eine "+
			"Sitzkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return partieAusgeben(an, flaeche.Partie(ctx, rest[0], rest[1]))
		})
}

// partiebefehlbefehl stellt einen Spielbefehl.
func partiebefehlbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 4 {
		an.klage("partie-befehl braucht Partie, Sitz, Befehlskennung und "+
			"Art — dazu wahlweise die Nutzlast als JSON.\n%s", hilfe)

		return CodeBedienung
	}

	nutzlast, fehler := nutzlastAus(rest)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			antwort := flaeche.PartieBefehl(ctx, domain.Spielbefehl{
				PartieID: rest[0], Absender: rest[1], Kennung: rest[2],
				Art: rest[3], Nutzlast: nutzlast,
			})

			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			if an.alsJSON {
				return alsJSONSchreiben(an, antwort)
			}

			an.sag("Fassung %s, %d Ereignisse.\n",
				antwort.Fassung, len(antwort.Ereignisse))

			for _, ereignis := range antwort.Ereignisse {
				an.sag("  %s %s\n", ereignis.Folge, ereignis.Art)
			}

			return an.abschluss()
		})
}

// partiestandbefehl pausiert, setzt fort oder beendet.
//
// **Ein Befehl mit drei Worten und nicht drei Befehle**: sie unterscheiden
// sich in einem Wort, und der Weg dahinter ist derselbe. Wer sie trennte,
// hätte dreimal dieselbe Argumentprüfung.
func partiestandbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 3 || !istPartiestand(rest[2]) {
		an.klage("partie-stand braucht Partie, Sitz und eines von "+
			"pausieren, fortsetzen, beenden.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			var antwort app.PartieAntwort

			switch rest[2] {
			case "pausieren":
				antwort = flaeche.PartiePausieren(ctx, rest[0], rest[1])
			case "fortsetzen":
				antwort = flaeche.PartieFortsetzen(ctx, rest[0], rest[1])
			default:
				antwort = flaeche.PartieBeenden(ctx, rest[0], rest[1])
			}

			return partieAusgeben(an, antwort)
		})
}

// istPartiestand sagt, ob dieses Wort einen Standwechsel benennt.
func istPartiestand(wort string) bool {
	switch wort {
	case "pausieren", "fortsetzen", "beenden":
		return true
	}

	return false
}

// sitzwuenscheAus liest die Sitze aus `<kennung>:<name>:<deck>`.
//
// **Die Deckkennung bleibt Text.** Ob sie eine Zahl ist, entscheidet
// `internal/app` — hier stünde sonst eine zweite Prüfung, die bei der ersten
// Änderung von der anderen abwiche.
func sitzwuenscheAus(roh []string) ([]app.Sitzwunsch, error) {
	wuensche := make([]app.Sitzwunsch, 0, len(roh))

	for _, eintrag := range roh {
		teile := strings.Split(eintrag, ":")
		if len(teile) < 3 || len(teile) > 4 {
			return nil, domain.NewFieldError("sitz", "sitzform",
				"Ein Sitz steht als <kennung>:<name>:<deck>[:<commander>] "+
					"da — %q ist keiner.", eintrag)
		}

		wunsch := app.Sitzwunsch{
			Kennung: teile[0], Name: teile[1], DeckID: teile[2],
		}

		// **Ein vierter Teil nennt den Commander** — CR 903.3 (M19,
		// #1620).
		// **Wahlweise**, und das ist der Normalfall: es gibt eine
		// Commandervariante und acht andere Spielweisen.
		if len(teile) > 3 {
			wunsch.Commander = teile[3]
		}

		wuensche = append(wuensche, wunsch)
	}

	return wuensche, nil
}

// nutzlastAus liest die Nutzlast — leer, wo keine dasteht.
func nutzlastAus(rest []string) (map[string]any, error) {
	if len(rest) < 5 || strings.TrimSpace(rest[4]) == "" {
		return map[string]any{}, nil
	}

	nutzlast := map[string]any{}

	if fehler := json.Unmarshal([]byte(rest[4]), &nutzlast); fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Nutzlast lesen")
	}

	return nutzlast, nil
}

// partieAusgeben schreibt eine Partieantwort.
func partieAusgeben(an *schreiber, antwort app.PartieAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	nachladenMelden(an, antwort.Nachladefehler)

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Partie %s (%s), Fassung %s, Sicht für %s.\n",
		antwort.Partie.ID, antwort.Partie.Stand, antwort.Partie.Fassung,
		antwort.Partie.Fuer)

	// **Die Befunde stehen gleich unter der Kopfzeile** (§8, UXR-F21): ein
	// Sitz, dessen Profil es nicht gibt, macht die Partie nicht unbenutzbar,
	// und er wird trotzdem **genannt** — wie am Fenster und wie bei den
	// Wegen der Verwaltung.
	for _, befund := range antwort.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	for _, sitz := range antwort.Partie.Sitze {
		// **Leben immer, Gift und Mana nur, wo es sie gibt** (M3,
		// #1605).
		// Eine Null ist keine Auskunft; „Gift 0" an jeder Zeile wäre Lärm,
		// in dem die eine Marke untergeht, auf die es ankommt.
		an.sag("  Sitz %s — %s, %d Karten, %d Leben%s%s\n",
			sitz.Kennung, sitz.Name, sitz.Karten, sitz.Leben,
			nurWennDa(", Gift %d", sitz.Vergiftung),
			nurWennText(", Mana %s", sitz.Manavorrat))
	}

	// **Eine offene Wahl steht vor der Bilanz** — CR 704.5j (M12,
	// #1614).
	// Solange sie offen ist, geht nichts weiter; wer sie hinter der
	// Objektzahl suchte, hielte die Partie für hängengeblieben.
	for _, wahl := range antwort.Partie.Legendenwahlen {
		an.sag("  %s muss waehlen, welche Legende %q bleibt (CR 704.5j)\n",
			wahl.Sitz, wahl.Name)
	}

	an.sag("  %d Objekte, erlaubt: %s\n",
		len(antwort.Partie.Objekte), strings.Join(antwort.Erlaubt, ", "))

	// **Was ein Klick an einem Objekt täte** — P3 (#2371). Nur die Objekte
	// mit einer Primärhandlung; alle Handlungen trägt `--json`.
	for _, eine := range antwort.Handlungen {
		if eine.Primaer != "" {
			an.sag("  %s: %s%s\n", eine.Objekt, eine.Primaer,
				nurWennText(", Doppelklick %s", eine.Zweit))
		}
	}

	return an.abschluss()
}

// nurWennDa gibt den Satz, wenn die Zahl über null steht — sonst nichts.
func nurWennDa(muster string, zahl int) string {
	if zahl <= 0 {
		return ""
	}

	return fmt.Sprintf(muster, zahl)
}

// nurWennText gibt den Satz, wenn der Text nicht leer ist — sonst nichts.
func nurWennText(muster, text string) string {
	if text == "" {
		return ""
	}

	return fmt.Sprintf(muster, text)
}

// mitPartieflaeche öffnet den Bestand und baut die kopflose Anwendung.
func mitPartieflaeche(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Partieflaeche) int,
) int {
	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return arbeit(ctx, app.PartieflaecheUeber(db))
		})
}
