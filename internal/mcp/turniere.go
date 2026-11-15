// **Turniere über MCP** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// Einordnung nach §16: **erweiternd** — zehn Werkzeuge kommen hinzu, drei
// lesend und sieben schreibend. Keines ändert eine bestehende Fläche.

package mcp

import (
	"context"
	"encoding/json"
)

// Turniere ist, was MCP für Turniere braucht.
type Turniere interface {
	Turniere(ctx context.Context, spiel string) (any, error)
	Turnier(ctx context.Context, turnierID int64) (any, error)
	TurnierRangliste(ctx context.Context, turnierID int64) (any, error)
	TurnierPaarungsvorschlag(ctx context.Context, turnierID int64) (any, error)

	TurnierAnlegen(ctx context.Context, name, spiel, notiz string,
		einzelspiel bool, formatID int64) (any, error)
	TurnierKopieren(ctx context.Context, turnierID int64, name string,
		mitTeilnehmenden bool) (any, error)
	TurnierKorrekturfolgen(ctx context.Context, turnierID, paarungID int64,
		stand map[string]any) (any, error)
	TurnierAbschliessen(ctx context.Context, turnierID int64) error
	TurnierWiedereroeffnen(ctx context.Context, turnierID int64) error
	TurnierProtokoll(ctx context.Context, turnierID int64) (any, error)
	TurnierAusgabe(ctx context.Context, turnierID int64) (any, error)
	TurnierBericht(ctx context.Context, turnierID int64) (any, error)
	TurnierEinlesen(ctx context.Context, inhalt string) (any, error)
	TurnierLoeschen(ctx context.Context, turnierID int64) error
	TurnierAnmelden(ctx context.Context, turnierID int64, name string,
		deckID int64) (any, error)
	TurnierAussteigen(ctx context.Context, turnierID, teilnehmerID int64,
		ausgestiegen bool) error
	TurnierRundeLegen(ctx context.Context, turnierID int64) (any, error)
	TurnierErgebnisSetzen(ctx context.Context, turnierID, paarungID int64,
		stand map[string]any) error
}

// turniere gibt die Werkzeuge — `schreiben.turnier.kopieren` und die drei
// für Datei und Bericht kamen mit DOM-04 (#1699) dazu, **erweiternd** nach
// §16.
func turniere(naht Turniere) (lesend, schreibend []Werkzeug) {
	if naht == nil {
		return nil, nil
	}

	return []Werkzeug{
		lesenTurniere(naht), lesenTurnier(naht), lesenRangliste(naht),
		lesenKorrekturfolgen(naht), lesenProtokoll(naht),
		lesenAusgabe(naht), lesenBericht(naht),
	}, []Werkzeug{
		schreibenTurnierAnlegen(naht), schreibenTurnierKopieren(naht), schreibenEinlesen(naht),
		schreibenAbschliessen(naht), schreibenWiedereroeffnen(naht),
		schreibenTurnierLoeschen(naht),
		schreibenTurnierAnmelden(naht), schreibenTurnierAussteigen(naht),
		schreibenPaarungsvorschlag(naht), schreibenRundeLegen(naht),
		schreibenErgebnis(naht),
	}
}

// lesenTurniere nennt die Turniere eines Spiels.
func lesenTurniere(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "lesen.turniere", Art: Lesend,
		Erklaerung: "Nennt die Turniere — ohne Angabe alle, sonst die eines " +
			"Spiels. Ein Turnier ist ein VERLAUF über die Zeit: Runden, " +
			"Paarungen, Ergebnisse. Die Rangliste wird daraus gerechnet und " +
			"nirgends abgelegt. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"spiel": text("Der Name des Spiels — leer für alle."),
		}),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				Spiel string `json:"spiel"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			liste, fehler := naht.Turniere(ctx, wahl.Spiel)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(liste)
		},
	}
}

// lesenTurnier gibt ein Turnier samt Teilnehmern und Verlauf.
func lesenTurnier(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "lesen.turnier", Art: Lesend,
		Erklaerung: "Gibt ein Turnier samt Teilnehmern, Runden und Paarungen. " +
			"Eine Paarung ohne »b« ist ein BYE — kein fehlender Gegner, " +
			"sondern ein Ergebnis (2–0). Eine Paarung ohne »ergebnis« ist " +
			"noch nicht gespielt und zählt nicht. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"turnierID": zahl("Die Kennung des Turniers aus »lesen.turniere«."),
		}, "turnierID"),
		Ruf: turnierruf(naht.Turnier),
	}
}

// lesenRangliste rechnet den Stand.
func lesenRangliste(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "lesen.turnier.rangliste", Art: Lesend,
		Erklaerung: "Rechnet die Rangliste eines Turniers nach den Magic " +
			"Tournament Rules (§3.1, Appendix C): Matchpunkte, dann " +
			"Gegnermatchquote, dann Spielquote, dann Gegnerspielquote. Jede " +
			"Quote hat eine Untergrenze von 0,33. Ein Bye zählt als 2–0. Wer " +
			"ausgestiegen ist, bleibt in der Liste — seine Quote geht weiter " +
			"in die seiner Gegner ein. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"turnierID": zahl("Die Kennung des Turniers aus »lesen.turniere«."),
		}, "turnierID"),
		Ruf: turnierruf(naht.TurnierRangliste),
	}
}

// turnierruf ist der Ruf für ein Werkzeug, das nur eine Turnierkennung nimmt.
//
// **Einmal statt dreimal** — dieselbe Überlegung wie bei `deckruf`.
func turnierruf(
	tue func(context.Context, int64) (any, error),
) func(context.Context, json.RawMessage) (string, error) {
	return func(ctx context.Context, roh json.RawMessage) (string, error) {
		var wahl struct {
			TurnierID int64 `json:"turnierID"`
		}

		if err := lies(roh, &wahl); err != nil {
			return "", err
		}

		ergebnis, fehler := tue(ctx, wahl.TurnierID)
		if fehler != nil {
			return "", fehler
		}

		return alsJSON(ergebnis)
	}
}
