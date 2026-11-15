package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zusatzquelle liest und setzt die Zusatzfelder einer Karte — die Felder,
// die ihr Stil mitbringt (`extra card field:`) und die sich von Hand setzen
// lassen (#2298).
type Zusatzquelle interface {
	Zusatzfelder(ctx context.Context, setID, karteID int64) ([]Zusatzfeld, error)
	ZusatzwertSetzen(ctx context.Context, setID, karteID int64, feld, wert string) (Zusatzvorwert, error)
}

// Zusatzvorwert ist, was vor dem Setzen galt — aus **demselben** Stand, auf
// den geschrieben wurde (Befund 6 des Nachlaufs über E/F): der Block
// `extra data` und darin das Feld.
type Zusatzvorwert struct {
	Block       string
	BlockDa     bool
	Wert        string
	Gespeichert bool
}

// Zusatzfeld ist ein Zusatzfeld samt dem Wert, den die Karte dafür trägt.
type Zusatzfeld struct {
	Name         string   `json:"name"`
	Art          string   `json:"art"`
	Beschreibung string   `json:"beschreibung"`
	Wahlen       []string `json:"wahlen"`
	Wert         string   `json:"wert"`

	// Gespeichert sagt, ob die Karte überhaupt einen Wert trägt — sonst gilt
	// die Vorgabe des Stils.
	Gespeichert bool `json:"gespeichert"`
}

// zusatzwerkzeuge gibt das lesende und das schreibende — **keines ohne
// Naht**, wie bei den Kartenbildern.
func zusatzwerkzeuge(quelle Zusatzquelle) (lesend, schreibend []Werkzeug) {
	if quelle == nil {
		return nil, nil
	}

	return []Werkzeug{zusatzfelderLesen(quelle)}, []Werkzeug{zusatzwertSetzen(quelle)}
}

func zusatzfelderLesen(quelle Zusatzquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.zusatzfelder", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Zusatzfelder des Stils einer Karte, die sich von Hand setzen lassen, " +
			"samt Wert. »gespeichert« sagt, ob die Karte einen trägt — sonst gilt die Vorgabe des Stils.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets, in dem die Karte liegt."),
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
		}, "setID", "karteID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				SetID   int64 `json:"setID"`
				KarteID int64 `json:"karteID"`
			}

			if err := json.Unmarshal(argumente, &wahl); err != nil {
				return "", fmt.Errorf("die Argumente sind kein JSON: %w", err)
			}

			if err := pruefeKennungen(wahl.SetID, wahl.KarteID); err != nil {
				return "", err
			}

			felder, err := quelle.Zusatzfelder(ctx, wahl.SetID, wahl.KarteID)
			if err != nil {
				return "", err
			}

			for i := range felder {
				felder[i].Wahlen = liste(felder[i].Wahlen)
			}

			return alsJSON(map[string]any{"felder": liste(felder)})
		},
	}
}

func zusatzwertSetzen(quelle Zusatzquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.zusatzwert", Art: Schreibend,
		Erklaerung: "ÄNDERT den Wert eines Zusatzfelds an einer Karte. Der Name kommt aus " +
			"»lesen.zusatzfelder«. Gibt den vorigen Wert zurück und sagt, wie es zurückgeht.",
		Schema: objekt(map[string]any{
			"setID":   zahl("Die Kennung des Sets, in dem die Karte liegt."),
			"karteID": zahl("Die Kennung der Karte aus »lesen.karten«."),
			"feld":    text("Der Name des Zusatzfelds aus »lesen.zusatzfelder«."),
			"wert":    text("Der neue Wert. Eine leere Zeichenkette leert das Feld."),
		}, "setID", "karteID", "feld", "wert"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			// **`Wert` als Zeiger**: „fehlt" und „leer" sind zwei Aussagen,
			// wie bei »schreiben.feld«.
			var wahl struct {
				SetID   int64   `json:"setID"`
				KarteID int64   `json:"karteID"`
				Feld    string  `json:"feld"`
				Wert    *string `json:"wert"`
			}

			if err := json.Unmarshal(argumente, &wahl); err != nil {
				return "", fmt.Errorf("die Argumente sind kein JSON: %w", err)
			}

			if err := pruefeKennungen(wahl.SetID, wahl.KarteID); err != nil {
				return "", err
			}

			if strings.TrimSpace(wahl.Feld) == "" {
				return "", domain.NewFieldError("feld", domain.SchluesselPflichtfeld,
					"Ohne den Namen des Zusatzfelds ist nicht zu sagen, was zu ändern wäre.")
			}

			if wahl.Wert == nil {
				return "", domain.NewFieldError("wert", domain.SchluesselPflichtfeld,
					"Es fehlt »wert«. Wer das Feld leeren will, gibt ihn als leere Zeichenkette an.")
			}

			return zusatzwertGesetzt(ctx, quelle, wahl.SetID, wahl.KarteID, wahl.Feld, *wahl.Wert)
		},
	}
}

// zusatzwertGesetzt setzt den neuen Wert und nennt den Rückweg — **aus dem
// Vorwert, den das Setzen selbst gelesen hat**.
//
// Hier wurde vorher getrennt gelesen und dann geschrieben; dazwischen konnte
// eine andere Fläche ändern, und ein Feld, das nicht gesetzt war, bekam einen
// Rückweg über den ganzen Block, ohne den Block zu nennen (Befund 6 des
// Nachlaufs über E/F).
func zusatzwertGesetzt(
	ctx context.Context, quelle Zusatzquelle, setID, karteID int64, feld, wert string,
) (string, error) {
	vorher, fehler := quelle.ZusatzwertSetzen(ctx, setID, karteID, feld, wert)
	if fehler != nil {
		return "", fehler
	}

	antwort := map[string]any{"gesetzt": feld}

	if vorher.Gespeichert {
		antwort["vorher"] = vorher.Wert
		antwort["rueckweg"] = rueckweg(
			"schreiben.zusatzwert mit »feld« = %q und »wert« = %q", feld, vorher.Wert)

		return alsJSON(antwort)
	}

	// **Fehlte der Block ganz, nimmt der Rückweg ihn weg** (J05, #2269): ein
	// leerer Block zurückgeschrieben ließe ein neues, leeres Feld stehen, und
	// der Kartenstand unterscheidet „fehlt" von „leer".
	if !vorher.BlockDa {
		antwort["rueckweg"] = rueckweg(
			"die Karte trug %q nicht — »schreiben.feld« mit »feld« = %q und »entfernen« = true nimmt ihn wieder weg",
			"extra data", "extra data")

		return alsJSON(antwort)
	}

	// **Nicht gesetzt heißt: die Vorgabe galt.** Ein leerer Wert stellte sie
	// nicht wieder her, sondern setzte einen leeren; zurück geht es nur über
	// den Block, wie er vorher war.
	antwort["vorherBlock"] = vorher.Block
	antwort["rueckweg"] = rueckweg(
		"das Feld war nicht gesetzt — die Vorgabe des Stils galt; »schreiben.feld« mit »feld« = %q und »wert« = »vorherBlock« stellt den Block wieder her",
		"extra data")

	return alsJSON(antwort)
}
