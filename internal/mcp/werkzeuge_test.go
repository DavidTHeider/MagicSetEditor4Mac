package mcp_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// werkzeugliste holt `tools/list` und gibt die Werkzeuge.
func werkzeugliste(t *testing.T) []map[string]any {
	t.Helper()

	zeilen := sprich(t, probenbestand(), `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	if len(zeilen) != 1 {
		t.Fatalf("%d Antworten", len(zeilen))
	}

	ergebnis := teilAntwort(t, zeilen[0])

	roh := folge(t, ergebnis["tools"])

	werkzeuge := make([]map[string]any, 0, len(roh))

	for _, eintrag := range roh {
		werkzeug := abbildung(t, eintrag)
		werkzeuge = append(werkzeuge, werkzeug)
	}

	return werkzeuge
}

// **Die Reihen sind getrennt, nicht nur beschriftet** (#143).
//
// Ein Werkzeug, das ein Feld setzt, gehört nicht in dieselbe Reihe wie eines,
// das Karten aufzählt. Die Art steht an drei Stellen — im Namen, im Text und in
// `readOnlyHint` —, und ein Gegenüber erkennt sie an jeder einzelnen.
func TestSchreibendeWerkzeugeSindErkennbar(t *testing.T) {
	werkzeuge := werkzeugliste(t)
	if len(werkzeuge) == 0 {
		t.Fatal("kein einziges Werkzeug")
	}

	lesendeGesehen, schreibendeGesehen := 0, 0

	for _, werkzeug := range werkzeuge {
		name := zeichenkette(t, werkzeug["name"])
		anmerkungen := abbildung(t, werkzeug["annotations"])

		schreibt := strings.HasPrefix(name, "schreiben.")
		liest := strings.HasPrefix(name, "lesen.")

		if !schreibt && !liest {
			t.Errorf("%q sagt im Namen nicht, ob es liest oder schreibt", name)

			continue
		}

		if anmerkungen["readOnlyHint"] != liest {
			t.Errorf("%q: readOnlyHint ist %v", name, anmerkungen["readOnlyHint"])
		}

		// **Ein schreibendes Werkzeug gilt als zerstörend, ein lesendes nicht.**
		//
		// Der erste Anlauf setzte hier überall `false` und begründete es damit,
		// dass nichts ein Set löscht. Das ist wahr und beantwortet die Frage
		// nicht: `schreiben.feld` überschreibt einen Wert und kann ein Feld
		// wegnehmen. Ein Gegenüber, das vor zerstörenden Werkzeugen nachfragt,
		// soll dort nachfragen. Der Zweitblick über Phase 9 hat es angestoßen.
		if anmerkungen["destructiveHint"] != schreibt {
			t.Errorf("%q: destructiveHint ist %v", name, anmerkungen["destructiveHint"])
		}

		if liest {
			lesendeGesehen++

			// Die lesenden stehen **vor** den schreibenden: wer die Liste
			// kürzt, behält damit die harmlosen.
			if schreibendeGesehen > 0 {
				t.Errorf("%q steht hinter einem schreibenden Werkzeug", name)
			}
		} else {
			schreibendeGesehen++

			if !strings.Contains(zeichenkette(t, werkzeug["description"]), "ÄNDERT") {
				t.Errorf("%q sagt im Text nicht, dass es ändert", name)
			}
		}
	}

	if lesendeGesehen == 0 || schreibendeGesehen == 0 {
		t.Errorf("%d lesende, %d schreibende", lesendeGesehen, schreibendeGesehen)
	}
}

// Die lesenden Werkzeuge geben, was der Bestand hergibt.
func TestDieLesendenWerkzeugeAntworten(t *testing.T) {
	faelle := map[string]struct {
		ruf      string
		enthaelt string
	}{
		"Sets": {
			ruf:      `{"name":"lesen.sets","arguments":{}}`,
			enthaelt: "Probeset",
		},
		"Karten": {
			ruf:      `{"name":"lesen.karten","arguments":{"setID":1}}`,
			enthaelt: "Flammenball",
		},
		"Felder": {
			ruf:      `{"name":"lesen.felder","arguments":{"setID":1}}`,
			enthaelt: "rule text",
		},
		"Suche": {
			ruf:      `{"name":"lesen.suche","arguments":{"text":"Flamme"}}`,
			enthaelt: "Flammenball",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			zeilen := sprich(t, probenbestand(),
				`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+fall.ruf+`}`)

			if !strings.Contains(zeilen[0], fall.enthaelt) {
				t.Errorf("Antwort: %s", zeilen[0])
			}
		})
	}
}

// **Eine Änderung gibt den vorigen Wert zurück.**
//
// Das ist die Rücknahme aus [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md):
// eine Bestätigung können wir nicht zusagen — ob ein Gegenüber nachfragt,
// entscheidet es selbst —, eine Rücknahme schon.
func TestEineAenderungGibtDenVorigenWertZurueck(t *testing.T) {
	bestand := probenbestand()

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"name","wert":"Eisball"}}}`)

	if !strings.Contains(zeilen[0], "Flammenball") {
		t.Errorf("der vorige Wert fehlt in der Antwort: %s", zeilen[0])
	}

	if !strings.Contains(zeilen[0], "zuruecknahme") {
		t.Errorf("die Antwort sagt nicht, wie sich das zurücknehmen lässt: %s", zeilen[0])
	}

	if len(bestand.gesetzt) != 1 || bestand.gesetzt[0] != "name=Eisball" {
		t.Errorf("beim Bestand kam an: %v", bestand.gesetzt)
	}
}

// **Ein Feld, das es vorher nicht gab, wird zum Entfernen zurückgenommen.**
//
// Die halbe Zusage war: `vorherDa: false` sagte, dass das Feld fehlte, und der
// genannte Weg zurück war „setze den vorigen Wert" — ein leerer Wert also, und
// eine leere Zeile ist nicht dasselbe wie keine Zeile. Der Zweitblick über
// Phase 9 hat es gefunden.
func TestEinNeuesFeldWirdDurchEntfernenZurueckgenommen(t *testing.T) {
	bestand := probenbestand()

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"rule text","wert":"Wachsamkeit"}}}`)

	if !strings.Contains(zeilen[0], "entfernen") {
		t.Errorf("die Rücknahme nennt nicht das Entfernen: %s", zeilen[0])
	}

	// Und das Entfernen geht auch wirklich.
	zurueck := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"rule text","entfernen":true}}}`)

	if !strings.Contains(zurueck[0], `"isError":false`) {
		t.Errorf("das Entfernen scheiterte: %s", zurueck[0])
	}

	if len(bestand.gesetzt) != 2 || bestand.gesetzt[1] != "rule text entfernt" {
		t.Errorf("beim Bestand kam an: %v", bestand.gesetzt)
	}
}

// Was ein schreibendes Werkzeug nicht ausführen kann, sagt es.
func TestEinUnvollstaendigerSchreibrufWirdBenannt(t *testing.T) {
	faelle := map[string]struct {
		ruf      string
		enthaelt string
	}{
		"ohne Feldnamen": {
			ruf:      `{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"  ","wert":"x"}}`,
			enthaelt: "Ohne Feldnamen",
		},
		"eine Karte, die es nicht gibt": {
			ruf:      `{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":999,"feld":"name","wert":"x"}}`,
			enthaelt: "nicht gefunden",
		},
		"Argumente, die keine sind": {
			ruf:      `{"name":"schreiben.feld","arguments":"kein Objekt"}`,
			enthaelt: "kein JSON",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			zeilen := sprich(t, probenbestand(),
				`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+fall.ruf+`}`)

			if !strings.Contains(zeilen[0], fall.enthaelt) {
				t.Errorf("Antwort: %s", zeilen[0])
			}
		})
	}
}

// Eine fehlende oder unbrauchbare Kennung wird benannt.
func TestEineUnbrauchbareKennungWirdBenannt(t *testing.T) {
	faelle := map[string]struct {
		ruf      string
		enthaelt string
	}{
		"sie fehlt": {
			ruf:      `{"name":"lesen.karten","arguments":{}}`,
			enthaelt: "setID",
		},
		"sie ist keine Zahl": {
			ruf:      `{"name":"lesen.karten","arguments":{"setID":1.5}}`,
			enthaelt: "keine Kennung",
		},
		"die Argumente sind keine": {
			ruf:      `{"name":"lesen.karten","arguments":"drei"}`,
			enthaelt: "kein JSON",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			zeilen := sprich(t, probenbestand(),
				`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+fall.ruf+`}`)

			if !strings.Contains(zeilen[0], fall.enthaelt) {
				t.Errorf("Antwort: %s", zeilen[0])
			}
		})
	}
}

// **Ein leerer Bestand antwortet mit `[]`, nicht mit `null`.**
//
// Dieselbe Zusage, die `internal/app` der Oberfläche gibt, und aus demselben
// Grund: ein Gegenüber darf darüber laufen, ohne vorher zu prüfen. Ein Modell,
// das `null` bekommt, hat die Wahl zwischen einem Abbruch und einer Vermutung.
//
// Gefunden hat das die Sitzung von Anfang bis Ende gegen einen leeren Bestand.
func TestEinLeererBestandAntwortetMitEinerLeerenListe(t *testing.T) {
	leer := &bestandsprobe{}

	for _, ruf := range []string{
		`{"name":"lesen.sets","arguments":{}}`,
		`{"name":"lesen.karten","arguments":{"setID":1}}`,
		`{"name":"lesen.felder","arguments":{"setID":1}}`,
	} {
		zeilen := sprich(t, leer,
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+ruf+`}`)

		if !strings.Contains(zeilen[0], `[]`) || strings.Contains(zeilen[0], "null") {
			t.Errorf("%s: %s", ruf, zeilen[0])
		}
	}
}

// Jedes Werkzeug nennt ein Schema — auch das ohne Argumente.
func TestJedesWerkzeugNenntEinSchema(t *testing.T) {
	for _, werkzeug := range werkzeugliste(t) {
		schema := abbildung(t, werkzeug["inputSchema"])
		if schema["type"] != "object" {
			t.Errorf("%v: Schema %v", werkzeug["name"], schema)
		}
	}
}

// Die Fehlerwege der lesenden Werkzeuge kommen als Ergebnis zurück.
func TestDieLesendenWerkzeugeMeldenFehler(t *testing.T) {
	bestand := probenbestand()
	bestand.fehler = mcp.ErrKeinWerkzeug

	for _, ruf := range []string{
		`{"name":"lesen.karten","arguments":{"setID":1}}`,
		`{"name":"lesen.felder","arguments":{"setID":1}}`,
		`{"name":"lesen.suche","arguments":{"text":"x"}}`,
		`{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"name","wert":"x"}}`,
	} {
		zeilen := sprich(t, bestand,
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+ruf+`}`)

		if !strings.Contains(zeilen[0], `"isError":true`) {
			t.Errorf("%s: %s", ruf, zeilen[0])
		}
	}
}

// **Ein Schreibfehler wird gemeldet, nicht als Erfolg ausgegeben.**
//
// Ein eigener Weg: `KarteLesen` geht durch, und erst das Schreiben scheitert.
// Wer nur den ersten prüft, hat den zweiten nie ausgeführt — und ein Server,
// der still scheitert, meldet einem Modell eine Änderung, die es nicht gibt.
func TestEinSchreibfehlerWirdGemeldet(t *testing.T) {
	bestand := probenbestand()
	bestand.schreibfehler = mcp.ErrKeinWerkzeug

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"schreiben.feld","arguments":{"setID": 7, "karteID":7,"feld":"name","wert":"x"}}}`)

	if !strings.Contains(zeilen[0], `"isError":true`) {
		t.Errorf("Antwort: %s", zeilen[0])
	}
}

// Die Suche mit Argumenten, die keine sind.
func TestEineSucheMitUnbrauchbarenArgumenten(t *testing.T) {
	zeilen := sprich(t, probenbestand(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"lesen.suche","arguments":"kein Objekt"}}`)

	if !strings.Contains(zeilen[0], "kein JSON") {
		t.Errorf("Antwort: %s", zeilen[0])
	}
}

// Auch »lesen.felder« braucht eine brauchbare Kennung.
func TestFelderBrauchtEineKennung(t *testing.T) {
	zeilen := sprich(t, probenbestand(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":`+
			`{"name":"lesen.felder","arguments":{}}}`)

	if !strings.Contains(zeilen[0], "setID") {
		t.Errorf("Antwort: %s", zeilen[0])
	}
}
