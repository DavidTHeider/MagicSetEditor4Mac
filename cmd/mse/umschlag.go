package main

// **Der portable Umschlag an der Kommandozeile** — Werkstück N4 von
// Phase 5.4
// (#1457).
//
// **Drei Befehle und nicht zwei.** Schauen und Übernehmen sind getrennt,
// weil #1457 die fachliche Übernahmefreigabe verlangt — und an der
// Kommandozeile ist sie derselbe zweite Aufruf, den ein Mensch im Fenster
// als zweiten Knopf sieht. Ein `umschlag-uebernehmen`, das zuerst zeigt und
// dann schreibt, nähme die Freigabe entgegen, nachdem sie gegenstandslos
// geworden wäre.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// umschlagschnuerenbefehl legt einen Gegenstand als Umschlag ab.
//
// **Die Art steht hinten und ist wahlfrei** (0033, Regel 6): wer zwei
// Argumente schickt, meint ein Set — so wie bisher. Die Änderung ist
// **erweiternd** (T7, #1472).
func umschlagschnuerenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 || len(rest) > 3 {
		an.klage("umschlag-schnueren braucht eine Kennung, ein Ziel und "+
			"wahlfrei die Art.\n%s", hilfe)

		return CodeBedienung
	}

	art := ""
	if len(rest) == 3 {
		art = rest[2]
	}

	return mitUmschlagflaeche(ctx, bestandsort, an,
		func(ctx context.Context, anwendung *app.App) int {
			return umschlagAusgeben(an,
				anwendung.UmschlagSchnueren(ctx, art, rest[0], rest[1]),
				"Geschnürt nach "+rest[1]+".\n")
		})
}

// umschlagschauenbefehl sagt, was in einem Umschlag steht — **rein lesend**.
func umschlagschauenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("umschlag-schauen braucht genau einen Pfad.\n%s", hilfe)

		return CodeBedienung
	}

	return mitUmschlagflaeche(ctx, bestandsort, an,
		func(ctx context.Context, anwendung *app.App) int {
			return umschlagAusgeben(an, anwendung.UmschlagAnsehen(ctx, rest[0]), "")
		})
}

// umschlaguebernehmenbefehl schreibt einen Umschlag in den Bestand.
func umschlaguebernehmenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("umschlag-uebernehmen braucht genau einen Pfad.\n%s", hilfe)

		return CodeBedienung
	}

	return mitUmschlagflaeche(ctx, bestandsort, an,
		func(ctx context.Context, anwendung *app.App) int {
			antwort := anwendung.UmschlagUebernehmen(ctx, rest[0])
			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			if an.alsJSON {
				return alsJSONSchreiben(an, antwort)
			}

			an.sag("Übernommen als Set %s mit %d Karten.\n", antwort.SetID, antwort.Karten)

			for _, befund := range antwort.Befunde {
				an.sag("  Befund: %s\n", befund)
			}

			return an.abschluss()
		})
}

// mitUmschlagflaeche öffnet den Bestand und baut die kopflose Anwendung.
func mitUmschlagflaeche(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.App) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		anwendung, fehler := app.UmschlagflaecheUeber(db)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return arbeit(ctx, anwendung)
	})
}

// umschlagauskunft ist, was diese Fläche über einen Umschlag schreibt.
//
// **Eine eigene Gestalt und nicht die des Fensters.** `app.UmschlagAntwort`
// trägt `pfad`, `abgebrochen` und `fehler` — drei Felder, die es nur gibt,
// weil dort ein Dialog aufgeht. Hier stünden sie als leere Behauptungen, und
// **§16 verlangt, dass dieselbe Frage überall dieselbe Antwort bekommt**: der
// Flächenlauf vergleicht den Verbund, nicht ein Wort darin.
type umschlagauskunft struct {
	Urheber       string           `json:"urheber,omitempty"`
	Fingerabdruck string           `json:"fingerabdruck,omitempty"`
	Objekte       []umschlagobjekt `json:"objekte"`
}

// umschlagobjekt ist ein Stück im Umschlag.
type umschlagobjekt struct {
	Art         string   `json:"art"`
	Bezeichnung string   `json:"bezeichnung"`
	Groesse     int64    `json:"groesse"`
	Braucht     []string `json:"braucht"`
}

// umschlagAusgeben schreibt eine Umschlagantwort.
//
// **Ein Satz für den Erfolg und `""` für „sag nur, was drin ist"** — beim
// Schauen entsteht nichts, und ein „Fertig." darüber wäre eine Zusage über
// eine Handlung, die es nicht gab.
func umschlagAusgeben(an *schreiber, antwort app.UmschlagAntwort, satz string) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		auskunft := umschlagauskunft{
			Urheber:       antwort.Urheber,
			Fingerabdruck: antwort.Fingerabdruck,
			Objekte:       make([]umschlagobjekt, 0, len(antwort.Objekte)),
		}

		for _, objekt := range antwort.Objekte {
			auskunft.Objekte = append(auskunft.Objekte, umschlagobjekt{
				Art:         objekt.Art,
				Bezeichnung: objekt.Bezeichnung,
				Groesse:     objekt.Groesse,
				Braucht:     objekt.Braucht,
			})
		}

		return alsJSONSchreiben(an, auskunft)
	}

	if satz != "" {
		an.sag("%s", satz)
	}

	if antwort.Urheber != "" {
		an.sag("urheber\t%s\n", antwort.Urheber)
		an.sag("fingerabdruck\t%s\n", antwort.Fingerabdruck)
	}

	for _, objekt := range antwort.Objekte {
		an.sag("objekt\t%s\t%s\t%d\n", objekt.Art, objekt.Bezeichnung, objekt.Groesse)
	}

	return an.abschluss()
}
