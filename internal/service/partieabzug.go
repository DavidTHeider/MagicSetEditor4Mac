package service

// **Was einen Abzug selbsttragend macht** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Warum das neben dem Aufbau steht und nicht darin
//
// **Weil es eine andere Frage beantwortet.** `partieaufbau.go` sagt, wie aus
// Wünschen eine Partie wird; hier steht, was eine Karte zum **Zeichnen**
// braucht — und das ist genau die Frage, die entsteht, wenn ein Abzug den
// Rechner verlässt, auf dem er entstand. `partieaufbau.go` stand damit bei
// 427 Zeilen; der Schnitt folgt trotzdem der Frage und nicht der Zahl (§2).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// reichereAn trägt Feldwerte und Stil in die Seiten des Abzugs.
//
// # Warum ein Fehler hier keiner ist
//
// **Weil ein Deck auf eine gelöschte Karte zeigen darf** (§8: fremde und alte
// Daten sind der Normalfall). Wer daran scheiterte, machte aus einem Deck mit
// einer verschwundenen Karte einen Tisch, der nicht aufgeht — und die Karte
// steht im Abzug ohnehin mit Name und Menge.
//
// **Was das kostet, gehört ausgesprochen**: ein solcher Abzug reist nicht
// vollständig. Wer ihn verleiht, verleiht eine Karte, die beim Gegenüber ohne
// Feldwerte ankommt; dass sie fehlen, sieht man dort am leeren Bild.
func (a *Partieaufbau) reichereAn(ctx context.Context, abzug *domain.Partiedeck) {
	if a.blaetter == nil {
		return
	}

	stile := map[int64]string{}

	for karte := range abzug.Karten {
		for seite := range abzug.Karten[karte].Seiten {
			a.reichereSeiteAn(ctx, &abzug.Karten[karte].Seiten[seite], stile)
		}
	}
}

// reichereSeiteAn schlägt eine Seite nach — **einmal je Set für den Stil**.
func (a *Partieaufbau) reichereSeiteAn(
	ctx context.Context, seite *domain.Abzugsseite, stile map[int64]string,
) {
	// **Null heißt: ein echter Druck.** Er hat keine Karte in diesem
	// Bestand, und ein Nachschlagen darauf fragte nach der Zeile null.
	if seite.KarteID == 0 {
		return
	}

	karte, fehler := a.blaetter.KarteLesen(ctx, seite.KarteID)
	if fehler != nil {
		return
	}

	seite.Felder = map[string]string{}
	for _, wert := range karte.Values {
		seite.Felder[wert.Field] = wert.Text
	}

	stil, schon := stile[karte.SetID]
	if !schon {
		gefunden, fehler := a.blaetter.Setstil(ctx, karte.SetID)
		if fehler == nil {
			stil = gefunden
		}

		stile[karte.SetID] = stil
	}

	seite.Stil = stil
}
