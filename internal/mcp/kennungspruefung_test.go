package mcp_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Jede Kennung wird geprüft, bevor sie einen Dienst erreicht** —
// #591, §8.
//
// # Was gemessen war, bevor es diesen Wächter gab
//
// Am 2026-08-22 hielten **vier von fünf** Werkzeugen die Zusage nicht:
//
//	schreiben.feld   Fehler=ja   nennt "karteID"=nein
//	lesen.karten     Fehler=nein
//	lesen.felder     Fehler=nein
//	lesen.suche      Fehler=nein
//
// `schreiben.feld` erreichte mit einer Null den **Bestand** und scheiterte dort
// mit „nicht gefunden" — die Prüfung fand also statt, aber am falschen Ort und
// ohne den Feldnamen. Die drei lesenden gaben `[]` zurück: ein Sprachmodell
// schlösse daraus „das Set ist leer" statt „die Kennung ist keine".
//
// # Nur die **verlangten** Kennungen
//
// **Eine Kennung, die nicht in `required` steht, darf null sein.**
// `lesen.suche` sucht ohne `setID` über alle Sets — dort heißt null „nicht
// eingeschränkt" und nicht „keine Kennung". Der erste Anlauf nahm alle
// ID-Felder und brach genau daran; die Unterscheidung stand die ganze Zeit im
// Schema.
//
// # Warum er aus dem Schema liest
//
// **Eine Liste von Namen hielte den Fall, nicht die Klasse.** Ein Werkzeug, das
// morgen dazukommt, stünde nicht darin — und genau so sind die vier hier
// entstanden. Der Wächter nimmt jedes Werkzeug, sucht in seinem Schema die
// Felder, deren Name auf `ID` endet, und prüft jedes einzeln.
//
// # Was er nicht prüft
//
// **Ob die Meldung nützlich ist.** Er sieht, dass der Feldname vorkommt, nicht
// ob der Satz drumherum jemandem hilft. Und er prüft nur **Kennungen**: ein
// leerer Feldname oder ein Suchtext ohne Inhalt sind andere Fragen.
func TestJedeKennungWirdVorDemDienstGeprueft(t *testing.T) {
	geprueft := 0

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
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
		for _, feld := range kennungsfelder(t, werkzeug) {
			geprueft++

			t.Run(werkzeug.Name+"/"+feld, func(t *testing.T) {
				_, err := werkzeug.Ruf(t.Context(),
					argumenteMitNull(t, werkzeug, feld))
				if err == nil {
					t.Fatalf("eine Kennung von null kam durch")
				}

				if !strings.Contains(err.Error(), feld) {
					t.Errorf("die Meldung nennt %q nicht: %v", feld, err)
				}
			})
		}
	}

	// **Ohne diese Zeile wäre ein Werkzeug ohne Kennungen ein Erfolg** — und
	// ein Schema, das sich ändert, machte den Wächter still leer (§10).
	if geprueft < 7 {
		t.Errorf("nur %d Kennungen geprüft — der Wächter sieht die Schemata nicht mehr", geprueft)
	}
}

// kennungsfelder gibt die Schemafelder, deren Name auf `ID` endet.
func kennungsfelder(t *testing.T, werkzeug mcp.Werkzeug) []string {
	t.Helper()

	eigenschaften, da := werkzeug.Schema["properties"].(map[string]any)
	if !da {
		return nil
	}

	verlangt := map[string]bool{}
	if pflicht, da := werkzeug.Schema["required"].([]string); da {
		for _, name := range pflicht {
			verlangt[name] = true
		}
	}

	var aus []string

	for name := range eigenschaften {
		if strings.HasSuffix(name, "ID") && verlangt[name] {
			aus = append(aus, name)
		}
	}

	return aus
}

// argumenteMitNull baut Argumente, in denen genau ein Feld null ist.
//
// **Die übrigen werden brauchbar gefüllt.** Ein Aufruf, der an einem anderen
// Pflichtfeld scheitert, sagte über das geprüfte nichts.
func argumenteMitNull(t *testing.T, werkzeug mcp.Werkzeug, null string) json.RawMessage {
	t.Helper()

	eigenschaften, da := werkzeug.Schema["properties"].(map[string]any)
	if !da {
		t.Fatalf("%s hat kein Schema mit Feldern", werkzeug.Name)
	}

	argumente := map[string]any{}

	for name, roh := range eigenschaften {
		beschreibung, ist := roh.(map[string]any)
		if !ist {
			t.Fatalf("%s: das Feld %q hat keine Beschreibung", werkzeug.Name, name)
		}

		switch {
		case name == null:
			argumente[name] = 0
		case beschreibung["type"] == "integer":
			argumente[name] = 1
		case beschreibung["type"] == "string":
			argumente[name] = "x"
		}
	}

	roh, err := json.Marshal(argumente)
	if err != nil {
		t.Fatalf("Argumente bauen: %v", err)
	}

	return roh
}
