// **Die Stilwahl und die Stiloptionen als Werkzeuge**
// (#1107).
//
// # Warum der Wechsel zweistufig bleibt
//
// Ein Wechsel kann Feldwerte ungedeutet lassen und Stiloptionen verlieren. Im
// Fenster steht dazwischen eine Rückfrage; ein Modell hat keine, und deshalb
// muss es wenigstens **fragen** können — `lesen.stilfolgen` sagt, was der
// Wechsel kostet, bevor er stattfindet (§16, Regel 6).

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Stilquelle ist, was die sechs Stilwerkzeuge brauchen.
type Stilquelle interface {
	Stile(ctx context.Context, setID int64) ([]Stil, error)
	Folgen(ctx context.Context, setID int64, paket string) (Wechselfolgen, error)
	Wechseln(ctx context.Context, setID int64, paket string) error
	Optionen(ctx context.Context, setID int64) ([]Stiloption, error)
	OptionSetzen(ctx context.Context, setID int64, feld, wert string) error
	OptionZuruecksetzen(ctx context.Context, setID int64, feld string) error
}

// Stil ist ein wählbarer Stil.
type Stil struct {
	Paket        string `json:"paket"`
	Spiel        string `json:"spiel"`
	Version      string `json:"version"`
	EigenesSpiel bool   `json:"eigenesSpiel"`
	Gewaehlt     bool   `json:"gewaehlt"`
}

// Wechselfolgen sagt, was ein Wechsel kostet.
type Wechselfolgen struct {
	UngedeuteteFelder []string `json:"ungedeuteteFelder"`
	VerloreneOptionen []string `json:"verloreneOptionen"`
	Spielwechsel      bool     `json:"spielwechsel"`
}

// Stiloption ist eine Option des gewählten Stils samt Wert.
type Stiloption struct {
	Name         string `json:"name"`
	Art          string `json:"art"`
	Beschreibung string `json:"beschreibung"`
	Wert         string `json:"wert"`
	Gesetzt      bool   `json:"gesetzt"`
	HatVorgabe   bool   `json:"hatVorgabe"`
	Vorgabe      string `json:"vorgabe"`
}

// stilwerkzeuge gibt die drei lesenden und die drei schreibenden — und die
// der Zusatzfelder, denn auch sie bringt der Stil mit (#2298).
func stilwerkzeuge(quelle Stilquelle, zusatz Zusatzquelle) (lesend, schreibend []Werkzeug) {
	zusatzlesend, zusatzschreibend := zusatzwerkzeuge(zusatz)

	return append([]Werkzeug{stileLesen(quelle), stilfolgenLesen(quelle), stiloptionenLesen(quelle)},
			zusatzlesend...),
		append([]Werkzeug{
			stilWechseln(quelle), stiloptionSetzen(quelle), stiloptionZuruecksetzen(quelle),
		}, zusatzschreibend...)
}

// stileLesen nennt die wählbaren Stile eines Sets.
func stileLesen(quelle Stilquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.stile", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Stile, die für ein Set wählbar sind. »gewaehlt« markiert den aktuellen.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := kennung(argumente)
			if err != nil {
				return "", err
			}

			stile, lesefehler := quelle.Stile(ctx, wahl)
			if lesefehler != nil {
				return "", lesefehler
			}

			return alsJSON(map[string]any{"stile": liste(stile)})
		},
	}
}

// stilfolgenLesen sagt, was ein Wechsel kostet — **bevor** er stattfindet.
func stilfolgenLesen(quelle Stilquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.stilfolgen", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, was ein Stilwechsel kostet: ungedeutete Felder, verlorene Optionen, Spielwechsel. Ändert nichts.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"paket": text("Der Paketname des Zielstils aus »lesen.stile«."),
		}, "setID", "paket"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := stilwahl(argumente)
			if err != nil {
				return "", err
			}

			folgen, lesefehler := quelle.Folgen(ctx, wahl.SetID, wahl.Paket)
			if lesefehler != nil {
				return "", lesefehler
			}

			return alsJSON(map[string]any{
				"ungedeuteteFelder": liste(folgen.UngedeuteteFelder),
				"verloreneOptionen": liste(folgen.VerloreneOptionen),
				"spielwechsel":      folgen.Spielwechsel,
			})
		},
	}
}

// stilWechseln setzt den Stil eines Sets.
func stilWechseln(quelle Stilquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.stil", Art: Schreibend,
		Erklaerung: "ÄNDERT den Stil eines Sets. Frage vorher »lesen.stilfolgen« — der Wechsel kann Feldwerte ungedeutet lassen.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"paket": text("Der Paketname des Zielstils aus »lesen.stile«."),
		}, "setID", "paket"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := stilwahl(argumente)
			if err != nil {
				return "", err
			}

			if fehler := quelle.Wechseln(ctx, wahl.SetID, wahl.Paket); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"gewechselt": wahl.Paket})
		},
	}
}

// stilwahl liest Setkennung und Paketnamen aus den Argumenten.
func stilwahl(argumente json.RawMessage) (struct {
	SetID int64  `json:"setID"`
	Paket string `json:"paket"`
}, error,
) {
	var wahl struct {
		SetID int64  `json:"setID"`
		Paket string `json:"paket"`
	}

	if err := json.Unmarshal(argumente, &wahl); err != nil {
		return wahl, fmt.Errorf("die Argumente sind kein JSON: %w", err)
	}

	return wahl, pruefeKennung("setID", wahl.SetID)
}
