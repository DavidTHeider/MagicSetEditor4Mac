package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die beiden Scryfall-Werkzeuge tun, was ihre Namen sagen**
// (#1005).
//
// # Warum zwei und nicht eines
//
// §16, Regel 6: „Nichts wird still übernommen." Für ein Sprachmodell ist das
// die wichtigere Hälfte — es ruft, während niemand hinsieht.
func TestDerVorschlagZeigtUndAendertNichts(t *testing.T) {
	t.Parallel()

	quelle := &abgleichsprobe{}

	text, fehler := rufeAbgleich(t, quelle, "lesen.scryfall.vorschlag",
		`{"setID": 1, "karteID": 7, "name": "Prüfstein"}`)
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	if quelle.uebernommen != 0 {
		t.Errorf("der Vorschlag hat %dmal übernommen", quelle.uebernommen)
	}

	for _, erwartet := range []string{"Aenderungen", "rule text", "Herkunft", "AlterStunden"} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, text)
		}
	}

	// **Die leeren Listen bleiben Listen** — die Falle aus §3.
	//
	// Geprüft wird am **entzifferten** Wert und nicht am Text: die Ausgabe ist
	// eingerückt, und eine Prüfung auf `"Befunde":[]` hinge am Format.
	var gelesen map[string]any
	if err := json.Unmarshal([]byte(text), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v", err)
	}

	if _, istListe := gelesen["Befunde"].([]any); !istListe {
		t.Errorf("die leeren Befunde kamen nicht als Liste: %#v", gelesen["Befunde"])
	}
}

// **Und die Übernahme sagt, wie sie zurückzunehmen ist.**
//
// 0010: ein schreibendes Werkzeug sagt es. Ohne die vorigen Werte müsste ein
// Gegenüber den Vorschlag aufgehoben haben.
func TestDieUebernahmeSagtWieSieZurueckzunehmenIst(t *testing.T) {
	t.Parallel()

	quelle := &abgleichsprobe{}

	text, fehler := rufeAbgleich(t, quelle, "schreiben.scryfall.uebernehmen",
		`{"setID": 1, "karteID": 7, "name": "Prüfstein", "felder": ["rule text"]}`)
	if fehler != nil {
		t.Fatalf("rufen: %v", fehler)
	}

	if quelle.uebernommen != 1 {
		t.Errorf("es wurde %dmal übernommen", quelle.uebernommen)
	}

	for _, erwartet := range []string{"vorher", "zuruecknahme", "schreiben.feld", "Alt"} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, text)
		}
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen, bevor gefragt wird.**
func TestEineUnbrauchbareKennungWirdVorDemAbgleichAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"Vorschlag, Set 0":     `{"setID": 0, "karteID": 7, "name": "x"}`,
		"Vorschlag, Karte 0":   `{"setID": 1, "karteID": 0, "name": "x"}`,
		"Übernahme, Set 0":     `{"setID": 0, "karteID": 7, "name": "x", "felder": []}`,
		"Übernahme, Karte 0":   `{"setID": 1, "karteID": 0, "name": "x", "felder": []}`,
		"Vorschlag, kein JSON": `nicht JSON`,
		"Übernahme, kein JSON": `nicht JSON`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werkzeug := "lesen.scryfall.vorschlag"
			if strings.HasPrefix(name, "Übernahme") {
				werkzeug = "schreiben.scryfall.uebernehmen"
			}

			quelle := &abgleichsprobe{}

			if _, fehler := rufeAbgleich(t, quelle, werkzeug, argumente); fehler == nil {
				t.Error("die unbrauchbare Eingabe kam durch")
			}

			if quelle.gefragt != 0 || quelle.uebernommen != 0 {
				t.Error("es wurde trotzdem gefragt oder übernommen")
			}
		})
	}
}

// **Und ein Fehler des Abgleichs kommt durch.**
func TestEinFehlerDesAbgleichsKommtDurch(t *testing.T) {
	t.Parallel()

	quelle := &abgleichsprobe{fehler: domain.ErrNotFound}

	if _, fehler := rufeAbgleich(t, quelle, "lesen.scryfall.vorschlag",
		`{"setID": 1, "karteID": 7, "name": "x"}`); fehler == nil {
		t.Error("der Fehler kam beim Vorschlag nicht durch")
	}

	if _, fehler := rufeAbgleich(t, quelle, "schreiben.scryfall.uebernehmen",
		`{"setID": 1, "karteID": 7, "name": "x", "felder": []}`); fehler == nil {
		t.Error("der Fehler kam bei der Übernahme nicht durch")
	}
}

// rufeAbgleich findet ein Scryfall-Werkzeug und ruft es.
// **Ein eigener Name**, weil  in  schon
// eine andere Frage beantwortet: dort ein fertiges Werkzeug, hier eines aus
// der Liste über einem Abgleich.
func rufeAbgleich(
	t *testing.T, abgleich mcp.Fremdabgleich, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    abgleich,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("es gibt kein Werkzeug %q", name)

	return "", nil
}

// abgleichsprobe zählt mit, was gefragt und übernommen wurde.
type abgleichsprobe struct {
	gefragt     int
	uebernommen int
	fehler      error
}

func (a *abgleichsprobe) Vorschlag(
	context.Context, int64, int64, string,
) (mcp.Vorschlag, error) {
	a.gefragt++

	if a.fehler != nil {
		return mcp.Vorschlag{}, a.fehler
	}

	return mcp.Vorschlag{
		Aenderungen:  []mcp.Aenderung{{Feld: "rule text", Alt: "Alt", Neu: "Neu"}},
		Herkunft:     "https://beispiel.invalid/karte",
		AlterStunden: 3,
	}, nil
}

func (a *abgleichsprobe) Uebernehmen(
	context.Context, int64, int64, string, []string,
) ([]mcp.Aenderung, error) {
	if a.fehler != nil {
		return nil, a.fehler
	}

	a.uebernommen++

	return []mcp.Aenderung{{Feld: "rule text", Alt: "Alt", Neu: "Neu"}}, nil
}
