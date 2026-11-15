package main

// **Die Zugriffstoken an der Kommandozeile** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Gegen einen echten Bestand**, wie jeder Lauf in diesem Paket: was hier
// geprüft wird, ist der Weg durch den Composition Root, und ein Doppel
// darunter prüfte den Aufbau statt den Weg.

import (
	"encoding/json"
	"strings"
	"testing"
)

// **Der ganze Weg: anlegen, nennen, ändern, zurückziehen.**
func TestDerTokenwegAnDerZeile(t *testing.T) {
	ort, _, _ := mitKarte(t)

	klartext := strings.TrimSpace(mussEingabe(t, ort,
		`{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"lesen"}`, "token-anlegen", "1"))

	if !strings.HasPrefix(klartext, "mse_") {
		t.Fatalf("der Klartext sieht nicht wie ein Token aus: %q", klartext)
	}

	code, aus, fehler := fahreAn(t, ort, "token", "1", "--json")
	if code != CodeGut {
		t.Fatalf("token: %d (%s)", code, fehler)
	}

	tokenID := ersteTokenkennung(t, aus)

	t.Run("und das Geheimnis steht nicht in der Liste", func(t *testing.T) {
		if strings.Contains(aus, klartext) {
			t.Errorf("die Liste trägt den Klartext: %s", aus)
		}
	})

	mussEingabe(t, ort,
		`{"principal":"Rechner","bezeichnung":"Tagesbau","scope":"schreiben","fassung":"1"}`,
		"token-aendern", tokenID)

	code, aus, fehler = fahreAn(t, ort, "token", "1")
	if code != CodeGut {
		t.Fatalf("token: %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Tagesbau") || !strings.Contains(aus, "schreiben") {
		t.Errorf("die Änderung steht nicht in der Zeile: %q", aus)
	}

	if code, _, fehler := fahreAn(t, ort, "token-loeschen", tokenID); code != CodeGut {
		t.Fatalf("token-loeschen: %d (%s)", code, fehler)
	}

	if _, nachher, _ := fahreAn(t, ort, "token", "1"); strings.Contains(
		nachher, "Tagesbau") {
		t.Errorf("das gelöschte Token steht noch da: %q", nachher)
	}
}

// **Die Gültigkeit steht in einem Wort.**
func TestDieGueltigkeitStehtInDerZeile(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string]struct{ angaben, erwartet string }{
		"ohne Ablauf": {
			`{"principal":"Rechner","bezeichnung":"Ohne","scope":"lesen"}`, "unbefristet",
		},
		"mit Ablauf in der Zukunft": {
			`{"principal":"Rechner","bezeichnung":"Spaeter","scope":"lesen","ablauf":"2099-01-01T00:00:00Z"}`,
			"2099-01-01T00:00:00Z",
		},
		"mit Ablauf in der Vergangenheit": {
			`{"principal":"Rechner","bezeichnung":"Vorbei","scope":"lesen","ablauf":"2020-01-01T00:00:00Z"}`,
			"abgelaufen",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			mussEingabe(t, ort, fall.angaben, "token-anlegen", "1")

			code, aus, fehler := fahreAn(t, ort, "token", "1")
			if code != CodeGut {
				t.Fatalf("token: %d (%s)", code, fehler)
			}

			if !strings.Contains(aus, fall.erwartet) {
				t.Errorf("%q steht nicht in %q", fall.erwartet, aus)
			}
		})
	}
}

// **Jeder der vier Befehle sagt, wie er heißen will.**
func TestDieTokenbefehleWeisenFalscheAufrufeAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string][]string{
		"token ohne Profil":           {"token"},
		"token mit zweien":            {"token", "1", "2"},
		"token-anlegen ohne Profil":   {"token-anlegen"},
		"token-aendern ohne Kennung":  {"token-aendern"},
		"token-loeschen ohne Kennung": {"token-loeschen"},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet war CodeBedienung", code)
			}
		})
	}
}

// **Und ein fachliches Nein ist eines.**
func TestEinTokenAufEinemFehlendenProfil(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string]struct {
		eingabe   string
		argumente []string
	}{
		"anlegen auf ein Profil, das es nicht gibt": {
			`{"principal":"Rechner","bezeichnung":"Nirgends","scope":"lesen"}`,
			[]string{"token-anlegen", "4711"},
		},
		"ändern an einem Token, das es nicht gibt": {
			`{"principal":"Rechner","bezeichnung":"Nirgends","scope":"lesen","fassung":"1"}`,
			[]string{"token-aendern", "4711"},
		},
		"löschen eines Tokens, das es nicht gibt": {
			"", []string{"token-loeschen", "4711"},
		},
		"eine Kennung, die keine Zahl ist": {
			"", []string{"token", "keine-zahl"},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			code, _, _ := fahreAnMitEingabe(t, ort, fall.eingabe, fall.argumente...)
			if code != CodeFachlich {
				t.Errorf("Rückgabewert %d, erwartet war CodeFachlich", code)
			}
		})
	}
}

// **Ein krummer Rumpf ist ein Bedienfehler.**
func TestEinKrummerTokenrumpf(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for name, befehl := range map[string]string{
		"beim Anlegen": "token-anlegen",
		"beim Ändern":  "token-aendern",
	} {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAnMitEingabe(t, ort, "{", befehl, "1"); code == CodeGut {
				t.Errorf("der krumme Rumpf kam durch")
			}
		})
	}
}

// **Und jeder der vier gibt auf Wunsch JSON.**
func TestDieTokenbefehleGebenJSON(t *testing.T) {
	ort, _, _ := mitKarte(t)

	mussEingabe(t, ort, `{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"lesen"}`,
		"token-anlegen", "1")

	_, aus, _ := fahreAn(t, ort, "token", "1", "--json")
	tokenID := ersteTokenkennung(t, aus)

	roh := mussEingabe(t, ort,
		`{"principal":"Rechner","bezeichnung":"Tagesbau","scope":"lesen","fassung":"1"}`,
		"token-aendern", tokenID, "--json")
	if !strings.Contains(roh, `"bezeichnung"`) {
		t.Errorf("die Antwort ist kein Tokenverbund: %s", roh)
	}

	roh = mussEingabe(t, ort, `{"principal":"Rechner","bezeichnung":"Zweites","scope":"lesen"}`,
		"token-anlegen", "1", "--json")
	if !strings.Contains(roh, `"klartext"`) {
		t.Errorf("die Antwort trägt keinen Klartext: %s", roh)
	}

	if code, _, fehler := fahreAn(
		t, ort, "token-loeschen", tokenID, "--json"); code != CodeGut {
		t.Fatalf("token-loeschen: %d (%s)", code, fehler)
	}
}

// mussEingabe fährt einen Befehl mit Eingabe an und bricht bei Fehlschlag ab.
func mussEingabe(t *testing.T, ort, eingabe string, argumente ...string) string {
	t.Helper()

	code, aus, fehler := fahreAnMitEingabe(t, ort, eingabe, argumente...)
	if code != CodeGut {
		t.Fatalf("mse %v: %d (%s)", argumente, code, fehler)
	}

	return aus
}

// ersteTokenkennung liest die erste Kennung aus einer Tokenliste.
func ersteTokenkennung(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Token []struct {
			ID string `json:"id"`
		} `json:"token"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Tokenantwort lesen: %v — %s", err, roh)
	}

	if len(gelesen.Token) == 0 {
		t.Fatalf("die Antwort führt kein Token: %s", roh)
	}

	return gelesen.Token[0].ID
}

// **Rotation, Sammelwiderruf und Ablaufwahlen an der Zeile** — Werkstück O2
// von Phase 5.2 (#1426).
func TestRotationUndSammelwiderrufAnDerZeile(t *testing.T) {
	ort, _, _ := mitKarte(t)

	altes := strings.TrimSpace(mussEingabe(t, ort,
		`{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"lesen"}`, "token-anlegen", "1"))

	_, aus, _ := fahreAn(t, ort, "token", "1", "--json")
	tokenID := ersteTokenkennung(t, aus)

	code, neues, fehler := fahreAn(t, ort, "token-rotieren", tokenID)
	if code != CodeGut || !strings.HasPrefix(strings.TrimSpace(neues), "mse_") ||
		strings.TrimSpace(neues) == altes {
		t.Fatalf("token-rotieren: %d %q (%s)", code, neues, fehler)
	}

	code, wahlen, fehler := fahreAn(t, ort, "token-ablaufwahlen")
	if code != CodeGut || !strings.Contains(wahlen, "tag\t") || !strings.Contains(wahlen, "nie\t") {
		t.Fatalf("token-ablaufwahlen: %d %q (%s)", code, wahlen, fehler)
	}

	code, gezogen, fehler := fahreAn(t, ort, "token-alle-widerrufen", "1")
	if code != CodeGut || strings.TrimSpace(gezogen) != "1" {
		t.Fatalf("token-alle-widerrufen: %d %q (%s)", code, gezogen, fehler)
	}

	// **Und als JSON.**
	for _, argumente := range [][]string{
		{"token-ablaufwahlen", "--json"},
		{"token-alle-widerrufen", "1", "--json"},
	} {
		if code, ausgabe, fehler := fahreAn(t, ort, argumente...); code != CodeGut || !strings.HasPrefix(strings.TrimSpace(ausgabe), "{") {
			t.Errorf("%v: %d %q (%s)", argumente, code, ausgabe, fehler)
		}
	}

	neu := strings.TrimSpace(mussEingabe(t, ort,
		`{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"lesen"}`, "token-anlegen", "1"))
	_, aus, _ = fahreAn(t, ort, "token", "1", "--json")

	if code, rotiert, _ := fahreAn(t, ort, "token-rotieren", ersteTokenkennung(t, aus), "--json"); code != CodeGut ||
		!strings.Contains(rotiert, `"klartext"`) || strings.Contains(rotiert, neu) {
		t.Errorf("token-rotieren --json: %d %s", code, rotiert)
	}
}

// **Falsche Aufrufe und fehlende Ziele.**
func TestDieNeuenTokenbefehleWeisenFalscheAufrufeAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for name, argumente := range map[string][]string{
		"rotieren ohne Kennung":       {"token-rotieren"},
		"widerrufen ohne Profil":      {"token-alle-widerrufen"},
		"ablaufwahlen mit Angabe":     {"token-ablaufwahlen", "zuviel"},
		"rotieren eines Unbekannten":  {"token-rotieren", "4711"},
		"widerrufen mit krummer Zahl": {"token-alle-widerrufen", "x"},
	} {
		t.Run(name, func(t *testing.T) {
			if code, aus, _ := fahreAn(t, ort, argumente...); code == CodeGut {
				t.Errorf("%v ging durch: %s", argumente, aus)
			}
		})
	}
}
