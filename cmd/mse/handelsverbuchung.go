package main

// **Die Verbuchung an der Kommandozeile** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum ein Befehl mit einem Schritt und nicht vier Befehle
//
// **Weil es ein Vorgang ist und drei Zeitpunkte darin.** `verbuchen`,
// `abgleichen` und `bestaetigen` sind dieselbe Sache in drei Stadien; wer
// dafür drei Befehle schriebe, verteilte eine Zustandsmaschine über drei
// Hilfetexte.
//
// **Und `journal` steht daneben**, weil es keine Entscheidung ist, sondern
// eine Auskunft.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// belegsicht ist der Beleg, wie die Kommandozeile ihn schreibt.
//
// **Eine eigene Gestalt** — die Wails-DTO ist **innen**, diese Fläche
// **außen** (0033, Regel 5). Dieselbe Trennung wie bei `poststandsicht`.
type belegsicht struct {
	Vorgang      string `json:"vorgang"`
	Zustand      string `json:"zustand"`
	Abdruck      string `json:"abdruck,omitempty"`
	Unterschrift string `json:"unterschrift,omitempty"`
	Beidseitig   bool   `json:"beidseitig"`

	// Gegenzeichner und Gegengezeichnet sagen, **wessen** geprüfter Beleg
	// `beidseitig` deckt und wann er kam
	// (#1650).
	//
	// **Eine erweiternde Änderung** (0033, Regel 6): wer sie nicht liest,
	// liest weiter wie bisher.
	Gegenzeichner   string `json:"gegenzeichner,omitempty"`
	Gegengezeichnet string `json:"gegengezeichnet,omitempty"`

	Gebucht string `json:"gebucht,omitempty"`

	// Befund ist der jüngste abgewiesene Gegenbeleg (#2500) — **erweiternd**
	// wie `gegenzeichner` (0033, Regel 6).
	Befund *befundsicht `json:"befund,omitempty"`
}

// befundsicht ist ein abgewiesener Gegenbeleg, wie die Kommandozeile ihn
// schreibt — `mangel` ist eines aus `domain.DieGegenbelegmaengel`, `grund`
// der Satz der Prüfung.
type befundsicht struct {
	Von       string `json:"von"`
	Abdruck   string `json:"abdruck"`
	Mangel    string `json:"mangel"`
	Grund     string `json:"grund"`
	Empfangen string `json:"empfangen"`
}

// journalsicht ist eine Bewegung, wie die Kommandozeile sie schreibt.
type journalsicht struct {
	ID          int64  `json:"id"`
	Art         string `json:"art"`
	Anlass      string `json:"anlass"`
	Vorgang     string `json:"vorgang,omitempty"`
	Bezug       int64  `json:"bezug,omitempty"`
	Grund       string `json:"grund,omitempty"`
	Printing    string `json:"printing"`
	Bezeichnung string `json:"bezeichnung"`
	Menge       int    `json:"menge"`
	Zustand     string `json:"zustand"`
	Sprache     string `json:"sprache"`
	Finish      string `json:"finish"`
	ExemplarID  int64  `json:"exemplarId,omitempty"`
	Lagerort    string `json:"lagerort,omitempty"`
	Notiz       string `json:"notiz,omitempty"`
	Gebucht     string `json:"gebucht"`
}

// handelsverbuchungbefehl fährt `mse handel <schritt> <vorgang> [wer]`.
func handelsverbuchungbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	// **Die Liste fragt ohne Vorgang** (TRADE-01, #1715).
	if len(rest) > 0 && rest[0] == "liste" {
		return handelslistebefehl(ctx, bestandsort, an, rest[1:])
	}

	if len(rest) < 2 || len(rest) > 4 {
		an.klage("handel nimmt zwei bis vier Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	schritt, vorgang := rest[0], rest[1]

	// **Die Schritte vor dem Abschluss gehen über die Leitung** (T1,
	// #1637).
	// Sie stehen unter demselben Befehl, weil es **ein** Vorgang ist — und
	// sie verzweigen hier, weil sie keine Freigabe nehmen.
	if istLeitungsbefehl(schritt) {
		return handelsleitungbefehl(
			ctx, bestandsort, an, schritt, vorgang, rest[2:])
	}

	wer, freigabe, gut := handelsargumente(an, schritt, rest[2:])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return verbuchungsschritt(
				ctx, db, an, schritt, vorgang, wer, freigabe)
		})
}

// handelsargumente liest, was hinter Schritt und Vorgang steht.
//
// **Je Schritt eine andere Folge** (T10, #1475): `verbuchen` nimmt die
// Freigabe, `bestaetigen` die Partei **und** die Freigabe, `abgleichen`
// nichts — es bindet nichts und braucht deshalb keine.
func handelsargumente(an *schreiber, schritt string, rest []string) (
	wer, freigabe string, gut bool,
) {
	switch schritt {
	case "abgleichen":
		if len(rest) > 0 {
			an.klage("abgleichen nimmt weder Partei noch Freigabe.\n%s", hilfe)

			return "", "", false
		}

		return "", "", true

	case "verbuchen":
		if len(rest) != 1 {
			an.klage("verbuchen nimmt die Freigabe aus dem Fenster.\n%s", hilfe)

			return "", "", false
		}

		return "", rest[0], true

	case "bestaetigen":
		if len(rest) != 2 {
			an.klage("bestaetigen nimmt die Partei und die Freigabe.\n%s",
				hilfe)

			return "", "", false
		}

		return rest[0], rest[1], true
	}

	// **Der unbekannte Schritt endet erst weiter unten** — dort steht, was es
	// gibt, und ein zweiter Satz hier sagte dasselbe in anderen Worten.
	return "", "", true
}

// verbuchungsschritt fährt den gewählten Schritt und schreibt seinen Beleg.
func verbuchungsschritt(
	ctx context.Context, db *sql.DB, an *schreiber,
	schritt, vorgang, wer, freigabe string,
) int {
	flaeche := app.PeerflaecheUeber(db)

	var antwort app.VerbuchungAntwort

	switch schritt {
	case "verbuchen":
		antwort = flaeche.Handelsverbuchung(ctx, vorgang, freigabe)

	case "abgleichen":
		antwort = flaeche.Handelsabgleich(ctx, vorgang)

	case "bestaetigen":
		antwort = flaeche.Handelsbestaetigung(ctx, vorgang, wer, freigabe)

	default:
		an.klage("handel kennt verbuchen, abgleichen und bestaetigen.\n%s",
			hilfe)

		return CodeBedienung
	}

	return belegSchreiben(an, antwort)
}

// handelsbelegbefehl fährt `mse handelsbeleg <vorgang>` — Stand und Beleg,
// **ohne** einen Schritt zu gehen (Hüllenprobe `nat04handel`, #1672): wer
// unterschrieben hat, sagte die Kommandozeile bis dahin nur nach einem
// eigenen Schritt.
func handelsbelegbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("handelsbeleg braucht genau einen Vorgang.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return belegSchreiben(an,
				app.PeerflaecheUeber(db).Handelsbeleg(ctx, rest[0]))
		})
}

// belegSchreiben gibt Stand und Beleg aus — nach einem Schritt wie ohne.
func belegSchreiben(an *schreiber, antwort app.VerbuchungAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	sicht := belegsicht{Vorgang: antwort.Vorgang, Zustand: antwort.Zustand}
	if antwort.HatBeleg {
		sicht.Abdruck = antwort.Beleg.Abdruck
		sicht.Unterschrift = antwort.Beleg.Unterschrift
		sicht.Beidseitig = antwort.Beleg.Beidseitig
		sicht.Gegenzeichner = antwort.Beleg.Gegenzeichner
		sicht.Gegengezeichnet = antwort.Beleg.Gegengezeichnet
		sicht.Gebucht = antwort.Beleg.Gebucht

		if befund := antwort.Beleg.Befund; befund != nil {
			sicht.Befund = &befundsicht{
				Von: befund.Von, Abdruck: befund.Abdruck, Mangel: befund.Mangel,
				Grund: befund.Grund, Empfangen: befund.Empfangen,
			}
		}
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, sicht)
	}

	an.sag("%s\t%s\t%v\n", sicht.Vorgang, sicht.Zustand, sicht.Beidseitig)

	return an.abschluss()
}

// handelsjournalbefehl fährt `mse handelsjournal [vorgang]`.
//
// **Ohne Vorgang kommt, was ohne Handel gebucht wurde** — Kauf, Verkauf,
// Geschenk, Einfuhr und Korrektur.
func handelsjournalbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) > 1 {
		an.klage("handelsjournal nimmt höchstens ein Argument.\n%s", hilfe)

		return CodeBedienung
	}

	vorgang := ""
	if len(rest) == 1 {
		vorgang = rest[0]
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			antwort := app.PeerflaecheUeber(db).Handelsjournal(ctx, vorgang)
			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			return journalSchreiben(an, antwort.Zeilen)
		})
}

// journalSchreiben gibt die Bewegungen aus.
func journalSchreiben(an *schreiber, zeilen []app.JournalzeileDTO) int {
	sichten := make([]journalsicht, 0, len(zeilen))

	for _, zeile := range zeilen {
		sichten = append(sichten, journalsicht{
			ID: zeile.ID, Art: zeile.Art, Anlass: zeile.Anlass,
			Vorgang: zeile.Vorgang, Bezug: zeile.Bezug, Grund: zeile.Grund,
			Printing: zeile.Printing, Bezeichnung: zeile.Bezeichnung,
			Menge:   zeile.Menge,
			Zustand: zeile.Zustand, Sprache: zeile.Sprache,
			Finish: zeile.Finish, ExemplarID: zeile.ExemplarID,
			Lagerort: zeile.Lagerort, Notiz: zeile.Notiz,
			Gebucht: zeile.Gebucht,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, sichten)
	}

	for _, sicht := range sichten {
		an.sag("%d\t%s\t%s\t%s\t%d\t%s\n", sicht.ID, sicht.Art, sicht.Anlass,
			sicht.Bezeichnung, sicht.Menge, sicht.Gebucht)
	}

	return an.abschluss()
}
