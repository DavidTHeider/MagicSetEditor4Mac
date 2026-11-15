// **Die schreibenden Turnierwerkzeuge** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turniere.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort wird **gelesen**, hier
// **geschrieben**.

package mcp

import (
	"context"
	"encoding/json"
)

// schreibenTurnierAnlegen legt ein Turnier an.
func schreibenTurnierAnlegen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.anlegen", Art: Schreibend,
		Erklaerung: "Legt ein Turnier an. Ein Turnier gehört zu EINEM Spiel — " +
			"sonst gäbe es kein Format, gegen das sich prüfen ließe. " +
			"»einzelspiel« heißt: ein Match besteht aus einem Spiel; dann " +
			"sagen Spiel- und Gegnerspielquote nichts.",
		Schema: objekt(map[string]any{
			"name":        text("Der Name des Turniers."),
			"spiel":       text("Der Name des Spiels, etwa »magic«."),
			"notiz":       text("Eine Notiz — wahlfrei."),
			"einzelspiel": jaNein("Ein Match aus einem Spiel statt aus dreien."),
			"formatID": zahl("Das Format aus »lesen.formate« — wahlfrei; es muss " +
				"zum selben Spiel gehören."),
		}, "name", "spiel"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Name        string `json:"name"`
				Spiel       string `json:"spiel"`
				Notiz       string `json:"notiz"`
				Einzelspiel bool   `json:"einzelspiel"`
				FormatID    int64  `json:"formatID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			angelegt, fehler := naht.TurnierAnlegen(
				ctx, wahl.Name, wahl.Spiel, wahl.Notiz, wahl.Einzelspiel, wahl.FormatID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(angelegt)
		},
	}
}

// schreibenTurnierLoeschen nimmt ein Turnier samt seinem Verlauf heraus.
func schreibenTurnierLoeschen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.loeschen", Art: Schreibend,
		Erklaerung: "Löscht ein Turnier ENDGÜLTIG — samt Teilnehmern, Runden, " +
			"Paarungen und Ergebnissen. Die Decks der Teilnehmer bleiben.",
		Schema: objekt(map[string]any{
			"turnierID": zahl("Die Kennung des Turniers aus »lesen.turniere«."),
		}, "turnierID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				TurnierID int64 `json:"turnierID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if fehler := naht.TurnierLoeschen(ctx, wahl.TurnierID); fehler != nil {
				return "", fehler
			}

			return "gelöscht", nil
		},
	}
}

// schreibenTurnierAnmelden trägt einen Teilnehmer ein.
func schreibenTurnierAnmelden(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.anmelden", Art: Schreibend,
		Erklaerung: "Meldet einen Teilnehmer an. Wird ein Deck genannt, wird " +
			"seine Liste ABGEZOGEN und nicht verwiesen: eine Deckliste im " +
			"Turnier ist ein eingereichtes Dokument, und eine spätere " +
			"Änderung am Deck erreicht sie nicht mehr.",
		Schema: objekt(map[string]any{
			"turnierID": zahl("Die Kennung des Turniers aus »lesen.turniere«."),
			"name":      text("Der Name des Teilnehmers."),
			"deckID":    zahl("Die Kennung des Decks aus »lesen.decks« — wahlfrei."),
		}, "turnierID", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				TurnierID int64  `json:"turnierID"`
				Name      string `json:"name"`
				DeckID    int64  `json:"deckID"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			angemeldet, fehler := naht.TurnierAnmelden(
				ctx, wahl.TurnierID, wahl.Name, wahl.DeckID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(angemeldet)
		},
	}
}

// schreibenTurnierAussteigen meldet einen Teilnehmer ab — oder wieder an.
func schreibenTurnierAussteigen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.aussteigen", Art: Schreibend,
		Erklaerung: "Meldet einen Teilnehmer ab — er wird nicht mehr gepaart, " +
			"BLEIBT aber in der Rangliste: seine Matchquote geht weiter in die " +
			"Gegnerquote jedes Gegners ein, den er hatte. Wer ihn entfernte, " +
			"veränderte die Rangliste aller anderen. »ausgestiegen: false« " +
			"meldet ihn wieder an.",
		Schema: objekt(map[string]any{
			"turnierID":    zahl("Die Kennung des Turniers aus »lesen.turniere«."),
			"teilnehmerID": zahl("Die Kennung des Teilnehmers aus »lesen.turnier«."),
			"ausgestiegen": jaNein("Wahr meldet ab, falsch wieder an."),
		}, "turnierID", "teilnehmerID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				TurnierID    int64 `json:"turnierID"`
				TeilnehmerID int64 `json:"teilnehmerID"`
				Ausgestiegen bool  `json:"ausgestiegen"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			if fehler := naht.TurnierAussteigen(ctx, wahl.TurnierID,
				wahl.TeilnehmerID, wahl.Ausgestiegen); fehler != nil {
				return "", fehler
			}

			return "gesetzt", nil
		},
	}
}

// schreibenPaarungsvorschlag zeigt die nächste Runde, ohne sie zu legen.
//
// **Unter den schreibenden, obwohl es nichts schreibt** — und das ist eine
// Entscheidung: es steht in derselben Sitzung wie das Legen, und ein Modell,
// das den Vorschlag unter „lesen" suchte, fände ihn nicht neben dem Werkzeug,
// zu dem er gehört.
func schreibenPaarungsvorschlag(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.vorschlag", Art: Lesend,
		Erklaerung: "Zeigt, wie die nächste Runde gepaart WÜRDE — und legt sie " +
			"NICHT. Gepaart wird nach dem Stand (Schweizer System), eine " +
			"Wiederholung wird vermieden, und bei ungerader Zahl bekommt der " +
			"Letzte ohne Bye einen. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"turnierID": zahl("Die Kennung des Turniers aus »lesen.turniere«."),
		}, "turnierID"),
		Ruf: turnierruf(naht.TurnierPaarungsvorschlag),
	}
}

// schreibenRundeLegen paart die nächste Runde und legt sie an.
func schreibenRundeLegen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.runde", Art: Schreibend,
		Erklaerung: "Paart die nächste Runde und legt sie an — in EINEM Zug, " +
			"damit kein Absturz eine halbe Runde hinterlässt. Geht NICHT, " +
			"solange die laufende Runde offene Ergebnisse hat: sie paarte " +
			"sonst nach einem Stand, den es noch nicht gibt.",
		Schema: objekt(map[string]any{
			"turnierID": zahl("Die Kennung des Turniers aus »lesen.turniere«."),
		}, "turnierID"),
		Ruf: turnierruf(naht.TurnierRundeLegen),
	}
}

// schreibenErgebnis trägt ein Ergebnis ein — oder nimmt es zurück.
func schreibenErgebnis(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.ergebnis", Art: Schreibend,
		Erklaerung: "Trägt den Spielstand einer Paarung ein — als ZAHLEN, " +
			"nicht als Sieger: die Tiebreaker brauchen Spielpunkte. Wer " +
			"nichts angibt, nimmt das Ergebnis ZURÜCK; berichtigen ist am " +
			"Turniertisch der Normalfall.",
		Schema: objekt(map[string]any{
			"turnierID":     zahl("Die Kennung des Turniers aus »lesen.turniere«."),
			"paarungID":     zahl("Die Kennung der Paarung aus »lesen.turnier«."),
			"spieleA":       zahl("Gewonnene Spiele des Erstgenannten."),
			"spieleB":       zahl("Gewonnene Spiele des Zweitgenannten."),
			"unentschieden": zahl("Unentschiedene Spiele."),
			"zuruecknehmen": jaNein("Wahr nimmt das Ergebnis heraus."),
		}, "turnierID", "paarungID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				TurnierID     int64 `json:"turnierID"`
				PaarungID     int64 `json:"paarungID"`
				SpieleA       int   `json:"spieleA"`
				SpieleB       int   `json:"spieleB"`
				Unentschieden int   `json:"unentschieden"`
				Zuruecknehmen bool  `json:"zuruecknehmen"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			var stand map[string]any

			if !wahl.Zuruecknehmen {
				stand = map[string]any{
					"spieleA":       wahl.SpieleA,
					"spieleB":       wahl.SpieleB,
					"unentschieden": wahl.Unentschieden,
				}
			}

			if fehler := naht.TurnierErgebnisSetzen(
				ctx, wahl.TurnierID, wahl.PaarungID, stand); fehler != nil {
				return "", fehler
			}

			return "gesetzt", nil
		},
	}
}
