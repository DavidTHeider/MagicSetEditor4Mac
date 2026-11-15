package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MTGJSONOeffnen liest eine MTG-JSON-Quelle in den Bestand.
//
// Dieselbe Antwort wie `SetOeffnen` und dieselben Zusagen: ein Abbruch ist kein
// Fehler, eine schon offene Quelle wird gemeldet statt ein zweites Mal angelegt,
// und Befunde schließen den Erfolg nicht aus (§8).
func (a *App) MTGJSONOeffnen(ctx context.Context, pfad, spiel string) OeffnenAntwort {
	ctx = a.imProfil(ctx)

	leer := OeffnenAntwort{Befunde: []string{}, Angelegt: []string{}}

	if pfad == "" {
		leer.Abgebrochen = true

		return leer
	}

	// **Die Regel „eine Quelle, ein Set" steht im Importeur** (#886).
	ergebnis, err := a.importeur.MTGJSON(ctx, pfad, spiel)

	// **Ein Abbruch gibt das Ergebnis mit** (#155). Eine AllSets-Datei führt
	// Hunderte Sets, jedes in einer eigenen Transaktion; was vor dem Fehler
	// durchkam, steht im Bestand. Wer nur den Fehler weiterreicht, lässt Sets
	// zurück, von denen niemand weiß — und wer danach die Liste ansieht, hält
	// sie für vollständig.
	if ergebnis != nil {
		leer.SchonOffen = ergebnis.SchonOffen

		for _, kennung := range ergebnis.Angelegt {
			leer.Angelegt = append(leer.Angelegt, strconv.FormatInt(kennung, 10))
			a.arbeitVermerken(ctx, domain.ArbeitsartSet, strconv.FormatInt(kennung, 10), domain.VorgangEingefuehrt)
		}

		for _, befund := range ergebnis.Befunde {
			leer.Befunde = append(leer.Befunde, befund.String())
		}
	}

	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)
		leer.Unvollstaendig = true

		return leer
	}

	leer.SetID = strconv.FormatInt(ergebnis.SetID, 10)
	leer.Karten = ergebnis.Karten

	return leer
}

// MTGJSONSpiele nennt die Spiele, für die es eine Zuordnung gibt.
//
// **Die Liste kommt aus der Fachbibliothek, nicht von hier.** `internal/app`
// trifft keine fachlichen Entscheidungen (§2); ein fest eingetragenes »magic«
// wäre genau das — und es bliebe stehen, wenn jemand eine zweite Zuordnung
// aufnimmt.
func (a *App) MTGJSONSpiele() []string { return MTGJSONSpiele() }
