package main

// **Eine Kartenliste an der Kommandozeile** — Werkstück 2 von Phase 4.5
// (#1292).
//
// # Zwei Dinge gehören an jeden Lauf, der nach draußen greifen könnte (§3)
//
// **Ein eigenes Zuhause und eine umgelenkte Adresse.** Beides einzeln reicht
// nicht — gefunden am 2026-08-24 in
// #888, wo
// ein Lauf zweimal 25 MB von Scryfall geladen hat. `listenumgebung` setzt
// beides.

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// listenkarte ist eine Kartenantwort, wie Scryfall sie schickt — selbst
// geschrieben (§11), auf das Nötige gekürzt.
const listenkarte = `{
  "object": "card",
  "id": "00000000-0000-4000-8000-00000000000%d",
  "name": "%s",
  "lang": "en",
  "layout": "normal",
  "set": "probe",
  "set_name": "Probesatz",
  "collector_number": "%d",
  "rarity": "common",
  "type_line": "Instant"
}`

// listenumgebung stellt eine Gegenstelle auf und lenkt alles dorthin.
//
// **Sie gibt genau einen Treffer, wenn Name und Set genannt sind, und sonst
// zwei** — so verhält sich der echte Dienst (nachgemessen am 2026-08-30, siehe
// `internal/service/listeneinfuhr.go`).
func listenumgebung(t *testing.T) string {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			frage := anfrage.URL.RawQuery

			eine := strings.Contains(frage, "%21%22") &&
				(strings.Contains(frage, "+e%3A") || strings.Contains(frage, "%20e%3A"))

			if strings.Contains(frage, "Gibtesnicht") {
				schreiben.WriteHeader(http.StatusNotFound)
				schreibeAntwort(t, schreiben, `{"object":"error","code":"not_found",`+
					`"status":404,"details":"No cards found."}`)

				return
			}

			if eine {
				schreibeAntwort(t, schreiben, `{"object":"list","has_more":false,`+
					`"total_cards":1,"data":[`+karteMit(1, "Erste Probe", 1)+`]}`)

				return
			}

			schreibeAntwort(t, schreiben, `{"object":"list","has_more":false,`+
				`"total_cards":2,"data":[`+karteMit(1, "Erste Probe", 1)+`,`+
				karteMit(2, "Erste Probe", 2)+`]}`)
		}))

	t.Cleanup(dienst.Close)

	vorher := app.MitScryfalladresse(dienst.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	return eigenesZuhause(t)
}

// karteMit füllt die Vorlage.
func karteMit(kennung int, name string, nummer int) string {
	return fmt.Sprintf(listenkarte, kennung, name, nummer)
}

// schreibeAntwort schreibt und meldet, wenn es nicht geht.
func schreibeAntwort(t *testing.T, schreiben http.ResponseWriter, roh string) {
	t.Helper()

	if _, fehler := schreiben.Write([]byte(roh)); fehler != nil {
		t.Errorf("schreiben: %v", fehler)
	}
}

// fahreAnMitEingabe ist `fahreAn` mit einer Standardeingabe.
func fahreAnMitEingabe(t *testing.T, ort, eingabe string, argumente ...string) (
	code int, aus, fehler string,
) {
	t.Helper()

	var raus, rfehler bytes.Buffer

	code = lauf(probekontext(t), argumente, strings.NewReader(eingabe), &raus, &rfehler, ort)

	return code, raus.String(), rfehler.String()
}

// **`liste-deuten` sagt je Zeile, ob sie eindeutig ist — und schreibt nichts.**
func TestListeDeutenSchreibtNichts(t *testing.T) {
	ort := listenumgebung(t)

	code, aus, fehler := fahreAnMitEingabe(t, ort,
		"2 Erste Probe (probe) 1\n3 Zweite Probe", "liste-deuten")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "=\t2 Erste Probe (probe) 1") {
		t.Errorf("die eindeutige Zeile ist nicht als solche gekennzeichnet: %q", aus)
	}

	if !strings.Contains(aus, "?\t3 Zweite Probe") {
		t.Errorf("die offene Zeile ist nicht als solche gekennzeichnet: %q", aus)
	}

	// **Und die Sammlung ist leer geblieben.** Ohne diese Zeile prüfte der
	// Test die Ausgabe und nicht die Zusage.
	_, sammlung, _ := fahreAn(t, ort, "sammlung", "--json")
	if !strings.Contains(sammlung, `"exemplare": []`) {
		t.Errorf("das Deuten hat geschrieben: %s", sammlung)
	}
}

// **`liste-einlesen` nimmt das Eindeutige — und sagt, was offen bleibt.**
func TestListeEinlesenNimmtDasEindeutige(t *testing.T) {
	ort := listenumgebung(t)

	code, aus, fehler := fahreAnMitEingabe(t, ort,
		"2 Erste Probe (probe) 1\n3 Zweite Probe", "liste-einlesen")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "1 angelegt") || !strings.Contains(aus, "1 offen") {
		t.Errorf("die Ausgabe nennt nicht beides: %q", aus)
	}

	_, sammlung, _ := fahreAn(t, ort, "sammlung", "--json")
	if !strings.Contains(sammlung, "Erste Probe") {
		t.Errorf("das Exemplar steht nicht in der Sammlung: %s", sammlung)
	}
}

// **Ohne eine einzige eindeutige Zeile wird nichts geschrieben — und gesagt.**
func TestListeEinlesenOhneEindeutigeZeile(t *testing.T) {
	ort := listenumgebung(t)

	code, aus, _ := fahreAnMitEingabe(t, ort, "3 Zweite Probe", "liste-einlesen")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d", code)
	}

	if !strings.Contains(aus, "keine eindeutige Zeile") {
		t.Errorf("die Ausgabe sagt es nicht: %q", aus)
	}
}

// **Eine leere Eingabe ist ein fachliches Nein.**
func TestListeDeutenOhneEingabe(t *testing.T) {
	ort := listenumgebung(t)

	if code, _, _ := fahreAnMitEingabe(t, ort, "  \n ", "liste-deuten"); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}
}

// **Ein Name, den es nicht gibt, steht als Befund da.**
func TestListeDeutenMitUnbekanntemNamen(t *testing.T) {
	ort := listenumgebung(t)

	code, aus, _ := fahreAnMitEingabe(t, ort, "1 Gibtesnicht", "liste-deuten")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d", code)
	}

	if !strings.HasPrefix(aus, "!") {
		t.Errorf("die Zeile ist nicht als Befund gekennzeichnet: %q", aus)
	}
}

// **Als JSON geht beides auch.**
func TestListeAlsJSON(t *testing.T) {
	ort := listenumgebung(t)

	_, aus, _ := fahreAnMitEingabe(t, ort,
		"2 Erste Probe (probe) 1", "liste-deuten", "--json")
	if !strings.Contains(aus, `"eindeutig": true`) {
		t.Errorf("die JSON-Ausgabe nennt die Eindeutigkeit nicht: %s", aus)
	}

	_, eingelesen, _ := fahreAnMitEingabe(t, ort,
		"2 Erste Probe (probe) 1", "liste-einlesen", "--json")
	if !strings.Contains(eingelesen, `"offen": 0`) {
		t.Errorf("die JSON-Ausgabe nennt die offenen nicht: %s", eingelesen)
	}
}
