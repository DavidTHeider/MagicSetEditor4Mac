package mcp_test

// **Was an den Token- und Vorgabenwerkzeugen schiefgehen kann** — Werkstücke
// 10 und 15 von Phase 4.12
// (#1351).
//
// **Beide in einer Datei**, weil sie dieselbe Frage stellen: was sagt ein
// Werkzeug, wenn der Rumpf krumm ist oder die Naht klagt. Der Erfolgsweg beider
// steht im Idempotenzlauf — **und der prüft nur ihn**: ein Werkzeug, das jeden
// Fehlschlag als Erfolg meldete, wäre dort grün (§10).

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// mitToken gibt ein Tokenwerkzeug über einer Probe.
func mitToken(t *testing.T, probe *tokenprobe, name string) mcp.Werkzeug {
	t.Helper()

	return werkzeugAus(t, mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(), Token: probe,
	}, name)
}

// mitVorgaben gibt ein Vorgabenwerkzeug über einer Probe.
func mitVorgaben(t *testing.T, probe *vorgabenprobe, name string) mcp.Werkzeug {
	t.Helper()

	return werkzeugAus(t, mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(), Vorgaben: probe,
	}, name)
}

// mitFreigaben gibt ein Freigabenwerkzeug über einer Probe.
func mitFreigaben(t *testing.T, probe *freigabenprobe, name string) mcp.Werkzeug {
	t.Helper()

	return werkzeugAus(t, mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(), Freigaben: probe,
	}, name)
}

// werkzeugAus sucht ein Werkzeug in einer Nahtaufstellung.
func werkzeugAus(t *testing.T, naehte mcp.Naehte, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(naehte) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// **Jedes Tokenwerkzeug meldet seinen Fehler.**
func TestJedesTokenwerkzeugMeldetSeinenFehler(t *testing.T) {
	kaputt := errors.New("kaputt")

	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *tokenprobe
	}{
		"lesen: krummer Rumpf":   {"lesen.token", `{`, neueTokenprobe()},
		"anlegen: krummer Rumpf": {"schreiben.token.anlegen", `{`, neueTokenprobe()},
		"ändern: krummer Rumpf":  {"schreiben.token.aendern", `{`, neueTokenprobe()},
		"löschen: krummer Rumpf": {"schreiben.token.loeschen", `{`, neueTokenprobe()},
		"lesen: die Naht klagt": {
			"lesen.token", `{"profilID":"1"}`,
			&tokenprobe{fehler: kaputt},
		},
		"anlegen: die Naht klagt": {
			"schreiben.token.anlegen",
			`{"profilID":"1","bezeichnung":"X","scope":"lesen"}`,
			&tokenprobe{fehler: kaputt},
		},
		"ändern: die Naht klagt": {
			"schreiben.token.aendern",
			`{"tokenID":"1","bezeichnung":"X","scope":"lesen","fassung":"1"}`,
			&tokenprobe{fehler: kaputt},
		},
		"löschen: die Naht klagt": {
			"schreiben.token.loeschen", `{"tokenID":"1"}`,
			&tokenprobe{fehler: kaputt},
		},

		// **Werkstück O2** (#1426): Sammelwiderruf und Ablaufwahlen.
		"sammelwiderruf: krummer Rumpf": {"schreiben.token.widerrufen.alle", `{`, neueTokenprobe()},
		"sammelwiderruf: die Naht klagt": {
			"schreiben.token.widerrufen.alle", `{"profilID":"1"}`,
			&tokenprobe{fehler: kaputt},
		},
		"ablaufwahlen: die Naht klagt": {
			"lesen.token.ablaufwahlen", `{}`, &tokenprobe{fehler: kaputt},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			werkzeug := mitToken(t, fall.probe, fall.werkzeug)

			if _, err := werkzeug.Ruf(
				t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und jedes Vorgabenwerkzeug ebenso.**
// **Sammelwiderruf und Ablaufwahlen tragen ihre Antwort** — Werkstück O2.
func TestSammelwiderrufUndAblaufwahlenUeberMCP(t *testing.T) {
	probe := neueTokenprobe()

	wahlen, err := mitToken(t, probe, "lesen.token.ablaufwahlen").Ruf(t.Context(), json.RawMessage(`{}`))
	if err != nil || !strings.Contains(wahlen, `"tag"`) {
		t.Errorf("Ablaufwahlen: %s, %v", wahlen, err)
	}

	gezogen, err := mitToken(t, probe, "schreiben.token.widerrufen.alle").Ruf(
		t.Context(), json.RawMessage(`{"profilID":"1"}`))
	if err != nil || !strings.Contains(gezogen, `"gezogen": 1`) {
		t.Errorf("Sammelwiderruf: %s, %v", gezogen, err)
	}
}

func TestJedesVorgabenwerkzeugMeldetSeinenFehler(t *testing.T) {
	kaputt := errors.New("kaputt")

	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *vorgabenprobe
	}{
		"lesen: krummer Rumpf":  {"lesen.profilvorgaben", `{`, neueVorgabenprobe()},
		"setzen: krummer Rumpf": {"schreiben.profilvorgabe", `{`, neueVorgabenprobe()},
		"lesen: die Naht klagt": {
			"lesen.profilvorgaben", `{"profilID":"1"}`,
			&vorgabenprobe{fehler: kaputt},
		},
		"setzen: die Naht klagt": {
			"schreiben.profilvorgabe",
			`{"profilID":"1","art":"deck","werte":[]}`,
			&vorgabenprobe{fehler: kaputt},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			werkzeug := mitVorgaben(t, fall.probe, fall.werkzeug)

			if _, err := werkzeug.Ruf(
				t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und jedes schreibende Werkzeug gibt zurück, was es angelegt hat.**
//
// # Warum dieser Lauf sein muss
//
// **Der Idempotenzlauf ruft nur, was Idempotenz zusagt** — und kein
// schreibendes Werkzeug tut das (`TestKeinSchreibendesWerkzeugSagtIdempotenzZu`).
// Ihre **Erfolgswege** waren damit unbetreten: ein Werkzeug, das seine Antwort
// gar nicht baut, wäre grün gewesen (§10).
func TestJedesSchreibendeTokenwerkzeugGibtSeineAntwort(t *testing.T) {
	probe := neueTokenprobe()

	angelegt := rufeRoh(t, mitToken(t, probe, "schreiben.token.anlegen"),
		`{"profilID":"1","bezeichnung":"Bogenbau","scope":"lesen"}`)

	// **Der Klartext steht in dieser einen Antwort** — und sonst nirgends.
	if !strings.Contains(angelegt, `"klartext"`) ||
		!strings.Contains(angelegt, "GEHEIMNIS") {
		t.Errorf("die Antwort trägt keinen Klartext: %s", angelegt)
	}

	geaendert := rufeRoh(t, mitToken(t, probe, "schreiben.token.aendern"),
		`{"tokenID":"1","bezeichnung":"Tagesbau","scope":"schreiben","fassung":"1"}`)
	if !strings.Contains(geaendert, "Tagesbau") {
		t.Errorf("die Änderung fehlt in %s", geaendert)
	}

	geloescht := rufeRoh(t, mitToken(t, probe, "schreiben.token.loeschen"),
		`{"tokenID":"1"}`)
	if !strings.Contains(geloescht, `"geloescht"`) {
		t.Errorf("die Antwort nennt nicht, was weg ist: %s", geloescht)
	}
}

// **Und das Setzen einer Vorgabe gibt den ganzen Stand zurück.**
func TestDasSetzenEinerVorgabeGibtDenStand(t *testing.T) {
	probe := neueVorgabenprobe()

	gesetzt := rufeRoh(t, mitVorgaben(t, probe, "schreiben.profilvorgabe"),
		`{"profilID":"1","art":"deck","werte":["1"]}`)

	// **Alle fünf Arten und nicht die eine** (§7, Regel 2): wer nur die
	// geänderte bekäme, müsste danach noch einmal lesen.
	for _, art := range []string{"format", "spiel", "vorlage", "deck", "lagerort"} {
		if !strings.Contains(gesetzt, `"`+art+`"`) {
			t.Errorf("die Art %q fehlt in %s", art, gesetzt)
		}
	}
}

// **Jedes Freigabenwerkzeug meldet seinen Fehler.**
func TestJedesFreigabenwerkzeugMeldetSeinenFehler(t *testing.T) {
	kaputt := errors.New("kaputt")

	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *freigabenprobe
	}{
		"lesen: krummer Rumpf":  {"lesen.freigaben", `{`, neueFreigabenprobe()},
		"setzen: krummer Rumpf": {"schreiben.freigabe", `{`, neueFreigabenprobe()},
		"löschen: krummer Rumpf": {
			"schreiben.freigabe.loeschen", `{`, neueFreigabenprobe(),
		},
		"lesen: die Naht klagt": {
			"lesen.freigaben", `{"profilID":"1"}`,
			&freigabenprobe{fehler: kaputt},
		},
		"setzen: die Naht klagt": {
			"schreiben.freigabe",
			`{"profilID":"1","nutzerID":"2","rolle":"leser"}`,
			&freigabenprobe{fehler: kaputt},
		},
		"löschen: die Naht klagt": {
			"schreiben.freigabe.loeschen", `{"profilID":"1","nutzerID":"2"}`,
			&freigabenprobe{fehler: kaputt},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			werkzeug := mitFreigaben(t, fall.probe, fall.werkzeug)

			if _, err := werkzeug.Ruf(
				t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und die beiden schreibenden geben den Stand zurück.**
//
// **Der Idempotenzlauf ruft sie nicht** — kein schreibendes Werkzeug sagt
// Idempotenz zu. Ohne diesen Lauf wären ihre Erfolgswege unbetreten (§10).
func TestDieSchreibendenFreigabenwerkzeugeGebenDenStand(t *testing.T) {
	probe := neueFreigabenprobe()

	gesetzt := rufeRoh(t, mitFreigaben(t, probe, "schreiben.freigabe"),
		`{"profilID":"1","nutzerID":"2","rolle":"mitarbeiter"}`)
	if !strings.Contains(gesetzt, "mitarbeiter") {
		t.Errorf("die Rolle fehlt in %s", gesetzt)
	}

	// **Die eigene Rolle steht mit darin** — daran entscheidet ein Modell, ob
	// es überhaupt anbietet.
	if !strings.Contains(gesetzt, `"eigeneRolle"`) {
		t.Errorf("die eigene Rolle fehlt in %s", gesetzt)
	}

	geloescht := rufeRoh(t, mitFreigaben(t, probe, "schreiben.freigabe.loeschen"),
		`{"profilID":"1","nutzerID":"2"}`)
	if strings.Contains(geloescht, "mitarbeiter") {
		t.Errorf("die zurückgezogene Freigabe steht noch in %s", geloescht)
	}
}
