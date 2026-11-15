// **Die Rechnungen des Symboleditors über die Namenstabelle** — Werkstück
// #1391.
//
// # Was hier geprüft wird
//
// **Der Weg über den Namen**: dass jede der acht Rechnungen erreichbar ist,
// dass ihre Argumente ankommen und dass ein Befund als `error` mit seiner
// **Klasse** herauskommt — daran hängen die Rückgabewerte aller drei Flächen
// ohne Fenster.
//
// Dass die Rechnungen selbst stimmen, prüfen die Läufe neben ihnen
// (`symbolgruppe_test.go`, `symbolsymmetrie_test.go`, `symbol_test.go`).
package app_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// rechteck gibt die Teile einer Grundform als JSON.
func rechteck(t *testing.T) string {
	t.Helper()

	ergebnis, fehler := app.SymbolRechnen("lesen.symbol.grundform",
		json.RawMessage(`{"art":"Rechteck"}`))
	if fehler != nil {
		t.Fatalf("Grundform: %v", fehler)
	}

	roh, err := json.Marshal(ergebnis)
	if err != nil {
		t.Fatalf("als JSON: %v", err)
	}

	return string(roh)
}

// **Jede Rechnung ist über ihren Namen erreichbar.**
func TestJedeSymbolrechnungIstUeberIhrenNamenErreichbar(t *testing.T) {
	teile := rechteck(t)

	for name, fall := range map[string]struct {
		was       string
		argumente string
		erwartet  string
	}{
		"Grundform":     {"lesen.symbol.grundform", `{"art":"Rechteck"}`, "Rechteck"},
		"Griff":         {"lesen.symbol.griff", `{"teile":` + teile + `,"teil":0,"punkt":0,"x":0.5,"y":0.5}`, "Rechteck"},
		"Punkt":         {"lesen.symbol.punkt", `{"teile":` + teile + `,"teil":0,"punkt":0,"weise":"einfuegen"}`, "Rechteck"},
		"Punkt löschen": {"lesen.symbol.punkt", `{"teile":` + teile + `,"teil":0,"punkt":0,"weise":"loeschen"}`, "Rechteck"},
		"Reihenfolge":   {"lesen.symbol.reihenfolge", `{"teile":` + teile + `,"von":0,"nach":0}`, "Rechteck"},
		"Symmetrie":     {"lesen.symbol.symmetrie", `{"teile":` + teile + `,"stelle":0,"weise":"anlegen"}`, "Symmetrie"},
		"Zuschnitt": {
			"lesen.symbol.zuschnitt",
			`{"quelleBreite":100,"quelleHoehe":100,"zielBreite":50,"zielHoehe":50,` +
				`"auswahlX":0,"auswahlY":0,"breite":50,"hoehe":50}`,
			"breite",
		},
	} {
		t.Run(name, func(t *testing.T) {
			ergebnis, fehler := app.SymbolRechnen(fall.was, json.RawMessage(fall.argumente))
			if fehler != nil {
				t.Fatalf("%s: %v", fall.was, fehler)
			}

			roh, err := json.Marshal(ergebnis)
			if err != nil {
				t.Fatalf("als JSON: %v", err)
			}

			if !strings.Contains(string(roh), fall.erwartet) {
				t.Errorf("die Antwort nennt %q nicht: %s", fall.erwartet, roh)
			}
		})
	}
}

// **Gruppieren, auflösen und die Kopien einer Symmetrie.**
//
// Sie brauchen mehr als ein Teil beziehungsweise eine echte Symmetrie und
// stehen deshalb neben der Tabelle darüber.
func TestGruppeUndSymmetriekopien(t *testing.T) {
	eins := rechteck(t)
	zwei := "[" + strings.TrimSuffix(strings.TrimPrefix(eins, "["), "]") + "," +
		strings.TrimSuffix(strings.TrimPrefix(eins, "["), "]") + "]"

	gruppiert, fehler := app.SymbolRechnen("lesen.symbol.gruppe",
		json.RawMessage(`{"teile":`+zwei+`,"stellen":[0,1],"weise":"gruppieren"}`))
	if fehler != nil {
		t.Fatalf("gruppieren: %v", fehler)
	}

	alsJSON, err := json.Marshal(gruppiert)
	if err != nil {
		t.Fatalf("als JSON: %v", err)
	}

	if !strings.Contains(string(alsJSON), "Gruppe") {
		t.Errorf("keine Gruppe: %s", alsJSON)
	}

	if _, aufloesefehler := app.SymbolRechnen("lesen.symbol.gruppe",
		json.RawMessage(`{"teile":`+string(alsJSON)+`,"stellen":[0],"weise":"aufloesen"}`),
	); aufloesefehler != nil {
		t.Errorf("auflösen: %v", aufloesefehler)
	}

	symmetrie, fehler := app.SymbolRechnen("lesen.symbol.symmetrie",
		json.RawMessage(`{"teile":`+eins+`,"stelle":0,"weise":"anlegen"}`))
	if fehler != nil {
		t.Fatalf("Symmetrie anlegen: %v", fehler)
	}

	mitSymmetrie, err := json.Marshal(symmetrie)
	if err != nil {
		t.Fatalf("als JSON: %v", err)
	}

	for _, weise := range []string{"setzen", "aufloesen"} {
		if _, weisefehler := app.SymbolRechnen("lesen.symbol.symmetrie",
			json.RawMessage(`{"teile":`+string(mitSymmetrie)+
				`,"stelle":0,"weise":"`+weise+`","kopien":3}`),
		); weisefehler != nil {
			t.Errorf("%s: %v", weise, weisefehler)
		}
	}

	if _, kopienfehler := app.SymbolRechnen("lesen.symbol.symmetriekopien",
		json.RawMessage(`{"teile":`+string(mitSymmetrie)+`,"stelle":0}`),
	); kopienfehler != nil {
		t.Errorf("Kopien: %v", kopienfehler)
	}
}

// **Und jeder Fehlerweg gibt seine Klasse.**
//
// Daran hängen die Rückgabewerte aller drei Flächen ohne Fenster: die
// Kommandozeile setzt ihren Code, MCP sein `isError`, der Dienst seinen
// Status — und alle drei fragen mit `errors.Is`.
func TestJederSymbolfehlerTraegtSeineKlasse(t *testing.T) {
	teile := rechteck(t)

	for name, fall := range map[string]struct {
		was       string
		argumente string
		nennt     string
	}{
		"kein JSON":            {"lesen.symbol.grundform", `{`, "kein JSON"},
		"Rechnung unbekannt":   {"lesen.symbol.gibtsnicht", `{}`, "gibt es nicht"},
		"Grundform unbekannt":  {"lesen.symbol.grundform", `{"art":"GibtsNicht"}`, "Grundform"},
		"Punktweise unbekannt": {"lesen.symbol.punkt", `{"teile":` + teile + `,"weise":"tanzen"}`, "erlaubt sind"},
		"Gruppenweise":         {"lesen.symbol.gruppe", `{"teile":` + teile + `,"weise":"tanzen"}`, "erlaubt sind"},
		"Symmetrieweise":       {"lesen.symbol.symmetrie", `{"teile":` + teile + `,"weise":"tanzen"}`, "erlaubt sind"},
		"Teil gibt es nicht":   {"lesen.symbol.griff", `{"teile":` + teile + `,"teil":99}`, "Stelle"},
		"Kopien ohne Symmetrie": {
			"lesen.symbol.symmetriekopien", `{"teile":` + teile + `,"stelle":0}`, "Symmetrie",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, fehler := app.SymbolRechnen(fall.was, json.RawMessage(fall.argumente))
			if fehler == nil {
				t.Fatal("der Aufruf ging durch, erwartet war eine Abweisung")
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Klasse fehlt: %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.nennt) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.nennt, fehler)
			}
		})
	}
}

// **Die Namensliste nennt jede Rechnung, die es gibt.**
//
// **In beide Richtungen.** Ein Name in der Liste, den der Rechner nicht
// kennt, wäre eine Fläche ohne Wirkung; einer im Rechner, der fehlt, eine
// Fähigkeit, die niemand findet.
func TestDieNamenslisteNenntJedeRechnung(t *testing.T) {
	namen := app.Symbolrechnungsnamen()

	if len(namen) != 8 {
		t.Errorf("%d Namen, erwartet acht: %v", len(namen), namen)
	}

	for _, name := range namen {
		// **Mit leeren Argumenten**: was hier zählt, ist nicht das Ergebnis,
		// sondern dass der Name **nicht** als unbekannt abgewiesen wird.
		_, fehler := app.SymbolRechnen(name, json.RawMessage(`{}`))
		// **Auf den Satz der unbekannten Rechnung und nicht auf »gibt es
		// nicht«**: den sagt auch eine unbekannte *Weise*, und drei der acht
		// brauchen eine. Der erste Anlauf hielt sie darum für fehlend.
		if fehler != nil && strings.Contains(fehler.Error(), "die Rechnung") {
			t.Errorf("%q steht in der Liste, der Rechner kennt es nicht", name)
		}
	}
}

// **Und eine unlesbare Teilliste sagt es — der Satz, und nicht die Ursache.**
//
// # Was sich hier am 2026-09-06 geändert hat
//
// **Bis dahin verlangte dieser Lauf die Ursache in der Meldung.** Sie kommt
// aus `json.Unmarshal` und lautet etwa „invalid character 'x' at offset 12" —
// ein Go-Text über eine Nutzlast, die das **Fenster** gebaut hat. Für den
// Menschen davor ist er nicht zu gebrauchen, und §8 verbietet ihn: an
// Kommandozeile, MCP und REST steht `Meldung` unübersetzt da.
//
// **Verloren ist er trotzdem nicht.** [domain.MitGrund] hängt ihn in die
// Kette, und `app.Uebersetze` schreibt ihn ins Protokoll, bevor die Meldung
// hinausgeht (Werkstück 1 von Phase 4.13,
// #1352).
func TestSymbolteileUnlesbarTraegtDenGrundVerdeckt(t *testing.T) {
	ursache := errors.New("da war etwas")
	fehler := app.SymbolteileUnlesbar(ursache)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Klasse fehlt: %v", fehler)
	}

	if strings.Contains(fehler.Error(), "da war etwas") {
		t.Errorf("die Ursache steht in der Meldung: %v", fehler)
	}

	// **Und die Gegenprobe**: ohne sie sagte der Lauf darüber nur, dass die
	// Meldung sie nicht nennt — auch dann, wenn sie ganz verloren wäre (§10).
	if grund := domain.GrundVon(fehler); !errors.Is(grund, ursache) {
		t.Errorf("der Grund ging verloren: %v", grund)
	}
}
