// **Die Einstellungen eines Menschen und seine Auskunft — an jeder Fläche** —
// Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 9, 10 und 11.
//
// # Der Ablauf
//
// Einstellungen lesen → eine setzen → nachsehen, dass sie steht → die
// Auskunft holen und nachsehen, dass sie den Stichtag trägt und **kein
// Geheimnis** nennt.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestDerGanzeEinstellungswegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	if aus := mussKommandozeile(t, umgebung, "nutzereinstellung", "1"); !strings.Contains(
		aus, "barrierefreiheit.kontrast") {
		t.Errorf("die Reihe fehlt: %q", aus)
	}

	aus := mussKommandozeile(t, umgebung, "nutzereinstellung", "1",
		"barrierefreiheit.kontrast", "hoch")
	if !strings.Contains(aus, "hoch") {
		t.Errorf("der Stand fehlt: %q", aus)
	}

	// **Und sie bleibt** — die Antwort ist keine Behauptung über den Bestand.
	if gelesen := mussKommandozeile(t, umgebung, "nutzereinstellung",
		"1"); !strings.Contains(gelesen, "hoch") {
		t.Errorf("die Einstellung blieb nicht: %q", gelesen)
	}

	// **Ein unerlaubter Wert kommt nicht durch.**
	if _, _, code := kommandozeile(t, umgebung, "nutzereinstellung", "1",
		"barrierefreiheit.kontrast", "sehr hoch"); code != 1 {
		t.Errorf("der unerlaubte Wert kam durch (%d)", code)
	}

	pruefeAuskunft(t, mussKommandozeile(t, umgebung, "personendaten", "1"))
}

// **Derselbe Weg über MCP.**
func TestDerGanzeEinstellungswegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	mussMCP(t, umgebung, "schreiben.nutzereinstellung", map[string]any{
		"nutzerID": "1", "schluessel": "barrierefreiheit.bewegung",
		"wert": "reduziert",
	})

	gelesen := mussMCP(t, umgebung, "lesen.nutzereinstellungen",
		map[string]any{"nutzerID": "1"})
	if !strings.Contains(gelesen, "reduziert") {
		t.Errorf("die Einstellung blieb nicht: %s", gelesen)
	}

	roh, istFehler := mcpRuf(t, umgebung, "schreiben.nutzereinstellung", map[string]any{
		"nutzerID": "1", "schluessel": "farbe.des.himmels", "wert": "blau",
	})
	if !istFehler {
		t.Errorf("der unbekannte Schlüssel kam durch: %s", roh)
	}

	pruefeAuskunft(t, mussMCP(t, umgebung, "lesen.personendaten",
		map[string]any{"nutzerID": "1"}))
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeEinstellungswegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	mussREST(t, an, http.MethodPut,
		"/nutzer/1/einstellungen/sicherheit.faktorbeimwechsel",
		`{"wert":"nurAnmelden"}`)

	gelesen := mussREST(t, an, http.MethodGet, "/nutzer/1/einstellungen", "")
	if !strings.Contains(gelesen, "nurAnmelden") {
		t.Errorf("die Einstellung blieb nicht: %s", gelesen)
	}

	status, gemeldet := an.ruf(t, http.MethodPut,
		"/nutzer/1/einstellungen/sicherheit.faktorbeimwechsel", `{"wert":"nie"}`)
	if status < 400 {
		t.Errorf("der unerlaubte Wert kam durch: %d (%s)", status, gemeldet)
	}

	pruefeAuskunft(t, mussREST(t, an, http.MethodGet, "/nutzer/1/personendaten", ""))
}

// pruefeAuskunft hält, was an jeder Fläche gilt.
//
// **Der Stichtag gehört dazu**: eine Auskunft ohne Datum sagt nicht, wann sie
// galt. **Und das Geheimnis nicht**: ein Kennwort ist kein Datum über einen
// Menschen, sondern ein Mittel, ihn zu sein.
func pruefeAuskunft(t *testing.T, roh string) {
	t.Helper()

	var gelesen struct {
		Erhoben string `json:"erhoben"`
		Nutzer  struct {
			Name string `json:"name"`
		} `json:"nutzer"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Auskunft lesen: %v — %s", err, roh)
	}

	if gelesen.Erhoben == "" {
		t.Error("der Stichtag fehlt")
	}

	if gelesen.Nutzer.Name == "" {
		t.Error("der Name fehlt")
	}

	for _, verboten := range []string{"argon2", "otpauth", "$"} {
		if strings.Contains(roh, verboten) {
			t.Errorf("die Auskunft nennt %q", verboten)
		}
	}
}

// **Was ein Profil vom Menschen erbt — an jeder Fläche** — Werkstück 9 von
// Phase 4.12
// (#1351),
// Punkt 9.
//
// # Die Zusage
//
// **Setze es beim Menschen, und es gilt im Profil** — solange dort niemand
// gewählt hat. Und was das Profil wählt, sticht.
//
// **Über die Kommandozeile geprüft, weil sie beides zeigt**: `mse
// einstellungen --json` gibt `wert` **und** `wirksam`. Eine Fläche, die nur
// eines von beiden gäbe, ließe die Frage offen, ob die Verschmelzung
// überhaupt stattfindet.
func TestDieUebersteuerungGiltAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	// **Der Mensch wählt Dunkel.**
	mussKommandozeile(t, umgebung, "nutzereinstellung", "1",
		"erscheinungsbild", "dunkel")

	gelesen := mussKommandozeile(t, umgebung, "einstellungen", "--json")

	if !strings.Contains(gelesen, `"wert": "erben"`) {
		t.Errorf("das Profil hat gewählt, statt zu erben: %s", gelesen)
	}

	if !strings.Contains(gelesen, `"wirksam": "dunkel"`) {
		t.Errorf("der Wert des Menschen gilt nicht: %s", gelesen)
	}

	// **Und die Wahl des Profils sticht ihn.**
	//
	// **Ohne diesen zweiten Teil prüfte der erste nichts über die
	// Richtung**: er hielte auch, wenn immer der Mensch gewönne.
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		"erscheinungsbild", "hell")

	danach := mussKommandozeile(t, umgebung, "einstellungen", "--json")

	if !strings.Contains(danach, `"wirksam": "hell"`) {
		t.Errorf("die Wahl des Profils sticht nicht: %s", danach)
	}
}
