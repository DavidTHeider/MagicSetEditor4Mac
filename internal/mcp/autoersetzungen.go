// **Die Autoersetzungen verwalten, über MCP** — Werkstück E von Phase 4.11
// (#1341).
//
// Einordnung nach §16: **erweiternd** — fünf Werkzeuge kommen hinzu, eines
// lesend und vier schreibend.

package mcp

import (
	"context"
	"encoding/json"
)

// Autoersetzungen ist, was MCP für die Regelverwaltung braucht.
type Autoersetzungen interface {
	Autoersetzungen(ctx context.Context, setID int64) (any, error)
	AutoersetzungSetzen(ctx context.Context, setID int64, muster, ersatz string,
		ganzesWort, aktiv bool) error
	AutoersetzungLoeschen(ctx context.Context, setID int64, muster string) error
	AutoersetzungenZuruecksetzen(ctx context.Context, setID int64) error
	AutoersetzungSchalten(ctx context.Context, setID int64, an bool) error
}

// autoersetzungen gibt die fünf Werkzeuge.
func autoersetzungen(regeln Autoersetzungen) (lesend, schreibend []Werkzeug) {
	if regeln == nil {
		return nil, nil
	}

	return []Werkzeug{lesenAutoersetzungen(regeln)}, []Werkzeug{
		schreibenAutoersetzung(regeln),
		loeschenAutoersetzung(regeln),
		zuruecksetzenAutoersetzungen(regeln),
		schaltenAutoersetzung(regeln),
	}
}

// lesenAutoersetzungen nennt die Regeln, die für ein Set gelten.
func lesenAutoersetzungen(regeln Autoersetzungen) Werkzeug {
	return Werkzeug{
		Name: "lesen.autoersetzungen", Art: Lesend,
		Erklaerung: "Nennt die Autoersetzungsregeln, die beim Tippen in " +
			"diesem Set greifen — die des Spielpakets und die eigenen, in " +
			"einer Liste. »eigen« sagt, welche hier angelegt wurde: nur eine " +
			"solche lässt sich löschen. »an« ist der Schalter des ganzen " +
			"Spiels. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"setID": zahl("Das Set, dessen Spiel gefragt ist."),
		}, "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, err := kennung(roh)
			if err != nil {
				return "", err
			}

			liste, fehler := regeln.Autoersetzungen(ctx, setID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(liste)
		},
	}
}

// schreibenAutoersetzung legt eine Regel an oder schreibt sie fort.
func schreibenAutoersetzung(regeln Autoersetzungen) Werkzeug {
	return Werkzeug{
		Name: "schreiben.autoersetzung", Art: Schreibend,
		Erklaerung: "Legt eine Autoersetzungsregel an oder schreibt sie fort. " +
			"Das MUSTER ist der Schlüssel: dasselbe Muster zweimal ist " +
			"dieselbe Regel. Eine Regel des Spielpakets lässt sich damit " +
			"überlagern — ihr Ersatz, ihr Schalter und ihre Wortgrenze; ihr " +
			"Muster gehört dem Paket.",
		Schema: objekt(map[string]any{
			"setID":      zahl("Das Set, dessen Spiel die Regel bekommt."),
			"muster":     text("Der Text, auf den die Regel anspringt."),
			"ersatz":     text("Wodurch er ersetzt wird."),
			"ganzesWort": jaNein("Nur ein ganzes Wort ersetzen."),
			"aktiv":      jaNein("Ob die Regel greift."),
		}, "setID", "muster", "ersatz"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID      int64  `json:"setID"`
				Muster     string `json:"muster"`
				Ersatz     string `json:"ersatz"`
				GanzesWort bool   `json:"ganzesWort"`
				Aktiv      bool   `json:"aktiv"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			if err := regeln.AutoersetzungSetzen(ctx, argumente.SetID, argumente.Muster,
				argumente.Ersatz, argumente.GanzesWort, argumente.Aktiv); err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"gesetzt": argumente.Muster})
		},
	}
}

// loeschenAutoersetzung nimmt eine eigene Regel heraus.
func loeschenAutoersetzung(regeln Autoersetzungen) Werkzeug {
	return Werkzeug{
		Name: "schreiben.autoersetzung.loeschen", Art: Schreibend,
		Erklaerung: "Nimmt eine EIGENE Autoersetzungsregel heraus. Eine des " +
			"Spielpakets lässt sich nicht löschen — sie gehört dem Paket; " +
			"was geht, ist sie mit »aktiv: false« abzuschalten.",
		Schema: objekt(map[string]any{
			"setID":  zahl("Das Set, dessen Spiel die Regel führt."),
			"muster": text("Das Muster der Regel."),
		}, "setID", "muster"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID  int64  `json:"setID"`
				Muster string `json:"muster"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			if err := regeln.AutoersetzungLoeschen(ctx, argumente.SetID, argumente.Muster); err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"geloescht": argumente.Muster})
		},
	}
}

// zuruecksetzenAutoersetzungen vergisst alle Abweichungen eines Spiels.
func zuruecksetzenAutoersetzungen(regeln Autoersetzungen) Werkzeug {
	return Werkzeug{
		Name: "schreiben.autoersetzungen.zuruecksetzen", Art: Schreibend,
		Erklaerung: "Vergisst ALLE eigenen Autoersetzungsregeln dieses Spiels " +
			"und stellt den Schalter auf an. Übrig bleibt, was das Spielpaket " +
			"mitbringt. LÖSCHT — die eigenen Regeln sind danach weg.",
		Schema: objekt(map[string]any{
			"setID": zahl("Das Set, dessen Spiel zurückgesetzt wird."),
		}, "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, err := kennung(roh)
			if err != nil {
				return "", err
			}

			if err := regeln.AutoersetzungenZuruecksetzen(ctx, setID); err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"zurueckgesetzt": setID})
		},
	}
}

// schaltenAutoersetzung legt den globalen Schalter um.
func schaltenAutoersetzung(regeln Autoersetzungen) Werkzeug {
	return Werkzeug{
		Name: "schreiben.autoersetzung.schalter", Art: Schreibend,
		Erklaerung: "Schaltet die Autoersetzung für dieses SPIEL ganz an oder " +
			"aus. Aus heißt: keine Regel greift mehr, auch keine des Pakets.",
		Schema: objekt(map[string]any{
			"setID": zahl("Das Set, dessen Spiel geschaltet wird."),
			"an":    jaNein("Ob die Autoersetzung greift."),
		}, "setID", "an"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID int64 `json:"setID"`
				An    bool  `json:"an"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			if err := regeln.AutoersetzungSchalten(ctx, argumente.SetID, argumente.An); err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"an": argumente.An})
		},
	}
}
