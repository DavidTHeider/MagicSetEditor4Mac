// **Ein Handel an der Kommandozeile** —
// #1637.
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Warum dieselbe Befehlsgruppe wie die Verbuchung
//
// **Weil es ein Vorgang ist und viele Zeitpunkte darin.** `handel` trug schon
// die drei Schritte des Abschlusses; die neun davor daneben in eigene Befehle
// zu legen hieße, eine Zustandsmaschine über zwölf Hilfetexte zu verteilen.
// Dieselbe Überlegung steht im Kopf von `handelsverbuchung.go`, und dieselbe
// trägt `POST /handel/{vorgang}/schritt`.
//
// # Warum die Positionen von der Standardeingabe kommen
//
// **Weil eine Position sieben Felder hat.** Ein eigenes Trennzeichenformat
// dafür wäre eine zweite Gestalt neben dem JSON, das jede andere Fläche
// ohnehin spricht — und beim ersten Feld mit einem Doppelpunkt darin
// unlesbar. `deck-liste` liest aus demselben Grund von dort.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vorgangssicht ist ein Handelsvorgang, wie die Kommandozeile ihn schreibt.
//
// **Eine eigene Gestalt** — die Wails-DTO ist **innen**, diese Fläche
// **außen** (0033, Regel 5). Dieselbe Trennung wie bei `belegsicht`.
type vorgangssicht struct {
	Vorgang  string `json:"vorgang"`
	Revision int    `json:"revision"`
	Abdruck  string `json:"abdruck"`
	Eltern   string `json:"eltern,omitempty"`

	Art       string   `json:"art"`
	Initiator string   `json:"initiator"`
	Parteien  []string `json:"parteien"`

	// Ich ist die Partei dieses Profils — die, deren Stimme Naechster nennt
	// (#2026). Leer, wo es keine Identität gibt.
	Ich     string `json:"ich"`
	Zustand string `json:"zustand"`

	Positionen       []app.HandelspositionDTO `json:"positionen"`
	BewertungGeteilt bool                     `json:"bewertungGeteilt"`
	Zustimmungen     []string                 `json:"zustimmungen"`
	Ende             bool                     `json:"ende"`

	// Die vier aus TRADE-02 (#1716) und der Beschluss aus TRADE-03.
	Beschluss string   `json:"beschluss"`
	Lage      string   `json:"lage"`
	Naechster string   `json:"naechster"`
	Moegliche []string `json:"moegliche"`
	Phase     string   `json:"phase"`
}

// istLeitungsbefehl sagt, ob dieser Befehl über die Leitung geht.
//
// **Die neun Schritte kommen aus der Matrix und nicht aus einer Liste hier**
// — eine zweite Aufzählung wäre beim nächsten Übergang still zu kurz. Ohne
// die drei des Abschlusses: sie bewegen Bestand und Beleg und verlangen eine
// Freigabe; wer sie hier ginge, umginge sie.
//
// **Und vier Befehle daneben, die keine Schritte sind**: eröffnen legt einen
// Vorgang an, vorschlagen legt eine **Revision** an, stand und vorgaenge
// fragen nur — der zweite nach den Handeln mit einem Gegenüber (UX-08.AK-04,
// #1680).
func istLeitungsbefehl(schritt string) bool {
	if schritt == "eroeffnen" || schritt == "vorschlagen" ||
		schritt == "stand" || schritt == "vorgaenge" || schritt == "protokoll" ||
		schritt == "anhalt" ||
		istVerzeichnisbefehl(schritt) {
		return true
	}

	gemeint := domain.Handelsschritt(schritt)

	return domain.IstHandelsschritt(gemeint) &&
		!domain.IstAbschlussschritt(gemeint)
}

// handelsleitungbefehl fährt die Schritte vor dem Abschluss.
func handelsleitungbefehl(
	ctx context.Context, bestandsort string, an *schreiber,
	schritt, vorgang string, rest []string,
) int {
	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			flaeche := app.PeerflaecheUeber(db)

			switch schritt {
			case "eroeffnen":
				return handelEroeffnen(ctx, flaeche, an, vorgang, rest)

			case "vorschlagen":
				return handelVorschlagen(ctx, flaeche, an, vorgang, rest)

			case "stand":
				if len(rest) > 0 {
					an.klage("handel stand nimmt nur den Vorgang.\n%s", hilfe)

					return CodeBedienung
				}

				return vorgangSchreiben(an, flaeche.HandelStand(ctx, vorgang))

			case "protokoll":
				// **Wie ein Vorgang dorthin kam** (TRADE-03.AK-03, #1717).
				if len(rest) > 0 {
					an.klage("handel protokoll nimmt nur den Vorgang.\n%s", hilfe)

					return CodeBedienung
				}

				return protokollSchreiben(an, flaeche.HandelProtokoll(ctx, vorgang))

			case "anhalt":
				// **Was die Seiten ungefähr wert sind** (TRADE-02b, #1951).
				if len(rest) > 0 {
					an.klage("handel anhalt nimmt nur den Vorgang.\n%s", hilfe)

					return CodeBedienung
				}

				return anhaltSchreiben(an, flaeche.HandelPreisanhaltVon(ctx, vorgang))

			case "vorgaenge":
				// **Hier steht an der Stelle des Vorgangs das Gegenüber** —
				// derselbe Platz, eine andere Frage.
				if len(rest) > 0 {
					an.klage("handel vorgaenge nimmt nur das Gegenüber.\n%s", hilfe)

					return CodeBedienung
				}

				return vorgaengeSchreiben(an, flaeche.HandelVorgaenge(ctx, vorgang))
			}

			if istVerzeichnisbefehl(schritt) {
				return verzeichnisbefehl(ctx, flaeche, an, schritt, vorgang, rest)
			}

			return handelSchreiten(ctx, flaeche, an, schritt, vorgang, rest)
		})
}

// handelEroeffnen legt einen Vorgang an — die Positionen von der Eingabe.
func handelEroeffnen(
	ctx context.Context, flaeche *app.App, an *schreiber,
	vorgang string, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("handel eroeffnen nimmt die Art und das Gegenüber.\n%s", hilfe)

		return CodeBedienung
	}

	positionen, gelesen := positionenLesen(an)
	if !gelesen {
		return CodeBedienung
	}

	return vorgangSchreiben(an, flaeche.HandelEroeffnen(ctx, app.Handelswunsch{
		Vorgang: vorgang, Art: rest[0], Gegenueber: []string{rest[1]},
		Positionen: positionen,
	}))
}

// handelVorschlagen trägt die nächste Revision ein.
func handelVorschlagen(
	ctx context.Context, flaeche *app.App, an *schreiber,
	vorgang string, rest []string,
) int {
	if len(rest) > 0 {
		an.klage("handel vorschlagen nimmt nur den Vorgang — die Positionen "+
			"kommen von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	positionen, gelesen := positionenLesen(an)
	if !gelesen {
		return CodeBedienung
	}

	return vorgangSchreiben(an,
		flaeche.HandelVorschlagen(ctx, vorgang, positionen))
}

// handelSchreiten geht einen der neun Schritte.
//
// **`--gesehen=<abdruck>` nennt die Revision, die der Schritt meint**
// (#2496): ist der Vorgang inzwischen eine andere, ist das ein Konflikt.
func handelSchreiten(
	ctx context.Context, flaeche *app.App, an *schreiber,
	schritt, vorgang string, rest []string,
) int {
	gesehen := ""
	uebrig := make([]string, 0, len(rest))

	for _, wort := range rest {
		if wert, ist := strings.CutPrefix(wort, "--gesehen="); ist {
			gesehen = wert

			continue
		}

		uebrig = append(uebrig, wort)
	}

	if len(uebrig) > 1 {
		an.klage("handel %s nimmt höchstens die Partei.\n%s", schritt, hilfe)

		return CodeBedienung
	}

	wer := ""
	if len(uebrig) == 1 {
		wer = uebrig[0]
	}

	return vorgangSchreiben(an,
		flaeche.HandelSchritt(ctx, vorgang, schritt, wer, gesehen))
}

// positionenLesen liest die Positionen als JSON von der Standardeingabe.
//
// **Eine leere Eingabe ist keine leere Liste.** Ein Angebot ohne Position
// sagt nicht, worum es geht, und die Domäne weist es ohnehin ab; hier sagt
// die Klage, **wo** es fehlt — an der Eingabe und nicht am Vertrag.
func positionenLesen(an *schreiber) ([]app.HandelspositionDTO, bool) {
	roh, lesefehler := io.ReadAll(an.eingabe)
	if lesefehler != nil {
		an.klageUeber(fmt.Errorf("die Positionen lesen: %w", lesefehler))

		return nil, false
	}

	var positionen []app.HandelspositionDTO

	if fehler := json.Unmarshal(roh, &positionen); fehler != nil {
		an.klageUeber(fmt.Errorf(
			"die Positionen sind kein JSON-Feld: %w", fehler))

		return nil, false
	}

	return positionen, true
}

// vorgangSchreiben gibt einen Handelsvorgang aus.
func vorgangSchreiben(an *schreiber, antwort app.HandelsvorgangAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	sicht := sichtAus(antwort.Vorgang)

	if an.alsJSON {
		return alsJSONSchreiben(an, sicht)
	}

	an.sag("%s\t%d\t%s\t%s\n",
		sicht.Vorgang, sicht.Revision, sicht.Zustand, sicht.Abdruck)

	// **Eingerückt und nach der Zeile, die es schon gab** (TRADE-02, #1716):
	// wer die erste Zeile auswertet, liest sie weiter an derselben Stelle.
	an.sag("\tphase\t%s\n\tlage\t%s\n", sicht.Phase, sicht.Lage)

	if len(sicht.Moegliche) > 0 {
		an.sag("\tschritte\t%s\n", strings.Join(sicht.Moegliche, ","))
	}

	return an.abschluss()
}

// vorgaengeSchreiben gibt die Handel mit einem Gegenüber aus — je Handel
// eine Zeile, in derselben Gestalt wie `handel stand`.
func vorgaengeSchreiben(an *schreiber, antwort app.HandelsvorgaengeAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	sichten := make([]vorgangssicht, 0, len(antwort.Vorgaenge))
	for _, vorgang := range antwort.Vorgaenge {
		sichten = append(sichten, sichtAus(vorgang))
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"vorgaenge": sichten})
	}

	for _, sicht := range sichten {
		an.sag("%s\t%d\t%s\t%s\n",
			sicht.Vorgang, sicht.Revision, sicht.Zustand, sicht.Abdruck)
	}

	return an.abschluss()
}

// sichtAus gibt einer Revision die Gestalt dieser Fläche.
func sichtAus(vorgang app.HandelsvorgangDTO) vorgangssicht {
	return vorgangssicht{
		Vorgang: vorgang.Vorgang, Revision: vorgang.Revision,
		Abdruck: vorgang.Abdruck, Eltern: vorgang.Eltern,
		Art: vorgang.Art, Initiator: vorgang.Initiator,
		Parteien: vorgang.Parteien, Ich: vorgang.Ich, Zustand: vorgang.Zustand,
		Positionen:       vorgang.Positionen,
		BewertungGeteilt: vorgang.BewertungGeteilt,
		Zustimmungen:     vorgang.Zustimmungen,
		Ende:             vorgang.Ende,
		Beschluss:        vorgang.Beschluss, Lage: vorgang.Lage,
		Naechster: vorgang.Naechster, Moegliche: vorgang.Moegliche,
		Phase: vorgang.Phase,
	}
}

// protokollSchreiben gibt die Zustandswechsel eines Vorgangs aus — eine Zeile
// je Wechsel: Revision, von, nach, Zeitpunkt. **Von ist ein Strich**, wenn
// die Revision neu ist: eine leere Spalte verschöbe die übrigen.
func protokollSchreiben(an *schreiber, antwort app.HandelsprotokollAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, zeile := range antwort.Zeilen {
		von := zeile.Von
		if von == "" {
			von = "-"
		}

		an.sag("%d\t%s\t%s\t%s\n", zeile.Revision, von, zeile.Nach, zeile.Zeit)
	}

	return an.abschluss()
}
