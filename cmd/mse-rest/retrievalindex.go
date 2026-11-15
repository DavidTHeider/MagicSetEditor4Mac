package main

// **Der lokale Suchindex am Dienst** — Werkstück A15 von Phase 5.3
// (#1514).

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// indexauskunft reicht die vier Wege an den Dienst.
// indexauskunft bedient die Wege des Suchindex.
//
// **Mit dem Ort des Bestands** — daran hängt die Absprache über die
// Prozesse ([app.Indexsperre]). Ohne ihn baute dieser Dienst neben einem
// `mse indexaufbau` her, gegen dieselbe Datei (A14, #1453).
type indexauskunft struct {
	db   *sql.DB
	pfad string
}

func (i indexauskunft) Indexstand(ctx context.Context) (any, error) {
	antwort := app.RetrievalAnwendungUeber(i.db).Indexstand(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (i indexauskunft) Indexsuche(
	ctx context.Context, frage string, grenze int, filter map[string][]string,
) (any, error) {
	antwort := app.RetrievalAnwendungUeber(i.db).Indexsuche(ctx, frage, grenze,
		app.IndexsuchfilterDTO{
			Quellen: filter["quellen"], Sprachen: filter["sprachen"],
			Sets: filter["sets"], Farben: filter["farben"],
			Farbidentitaet: filter["farbidentitaet"], Typen: filter["typen"],
			ManawertVon: kommazahlAus(filter["manawertVon"]),
			ManawertBis: kommazahlAus(filter["manawertBis"]),
			Format:      erstes(filter["format"]),
			Formatlage:  erstes(filter["formatlage"]),
			Formatstand: erstes(filter["formatstand"]),
		})
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (i indexauskunft) IndexAufbauen(
	ctx context.Context, modell, quelle, revision, datum string,
	fortschreiben bool,
) (any, error) {
	antwort := app.RetrievalAnwendungUeber(i.db).
		MitBestandsort(i.pfad).IndexAufbauen(
		ctx, modell, quelle, revision, datum, fortschreiben)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

// IndexAbbrechen hält einen laufenden Aufbau an.
//
// **Über eine eigene Anwendung, und das ist die Grenze dieser Fläche**: der
// Riegel sitzt in der Verwaltung, und jede Anfrage baut sich hier eine neue.
// Abgebrochen wird damit **nur ein Aufbau, den derselbe Prozess hält** — über
// den Dienst ist das keiner, weil `IndexAufbauen` synchron antwortet.
// Wirksam abgebrochen wird dort, wo der Aufbau läuft: am Fenster über
// `IndexAbbrechen`, auf der Kommandozeile über `^C`.
func (i indexauskunft) IndexAbbrechen(context.Context) (any, error) {
	return map[string]bool{
		"abgebrochen": app.RetrievalAnwendungUeber(i.db).
			MitBestandsort(i.pfad).IndexAbbrechen(),
	}, nil
}

// erstes gibt den ersten Wert — leer, wo keiner steht.
func erstes(werte []string) string {
	if len(werte) == 0 {
		return ""
	}

	return werte[0]
}

// kommazahlAus liest eine Gleitkommazahl — `nil`, wo keine steht.
//
// **Ein Zeiger, weil null ein Wert ist**: ein Land hat den Manawert 0, und
// „kein Filter" ist etwas anderes als „höchstens null".
func kommazahlAus(werte []string) *float64 {
	roh := erstes(werte)
	if roh == "" {
		return nil
	}

	zahl, fehler := strconv.ParseFloat(roh, 64)
	if fehler != nil {
		return nil
	}

	return &zahl
}
