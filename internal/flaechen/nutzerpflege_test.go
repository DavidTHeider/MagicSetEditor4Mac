// **Nutzer und Profile bearbeiten — an jeder Fläche** — Werkstück 5 von
// Phase 4.12
// (#1351).
//
// # Der Ablauf
//
// Nutzer anlegen → umbenennen → zweites Profil anlegen → umbenennen → löschen
// → Nutzer löschen. **Und zweimal dieselbe Ablehnung**: der letzte Nutzer
// bleibt, das letzte Profil bleibt.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestDerGanzePflegewegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	nutzerID := nutzerkennung(t,
		mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json"))

	// **Die Angaben kommen als JSON von der Standardeingabe** — Werkstück 8
	// (#1351): elf Wörter in einer Zeile zählte niemand richtig ab. Die
	// **Fassung** gehört dazu; ohne sie wird abgewiesen, statt eine
	// gleichzeitige Änderung zu überschreiben.
	mitEingabe(t, umgebung, `{"name":"Umbenannt","fassung":"1"}`,
		"nutzer-aendern", nutzerID)

	if aus := mussKommandozeile(t, umgebung, "nutzer"); !strings.Contains(aus, "Umbenannt") {
		t.Errorf("der neue Name fehlt: %q", aus)
	}

	profilID := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", nutzerID, "Turnier", "--json"))

	// **Auch das Profil bekommt seine Angaben als JSON** — seit Werkstück 9
	// (#1351): mit den vier neuen Feldern wären es sieben Wörter im Aufruf.
	mitEingabe(t, umgebung, `{"name":"Legacy","subslogan":"Nur alte Formate","fassung":"1"}`,
		"profil-aendern", profilID)
	mussKommandozeile(t, umgebung, "profil-loeschen", profilID, "Legacy")
	mussKommandozeile(t, umgebung, "nutzer-loeschen", nutzerID, "Umbenannt")

	// **Und der letzte Nutzer bleibt.**
	if _, _, code := kommandozeile(t, umgebung, "nutzer-loeschen", "1", "Ich"); code != 1 {
		t.Errorf("der letzte Nutzer ließ sich löschen (%d)", code)
	}
}

// **Derselbe Weg über MCP.**
func TestDerGanzePflegewegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	nutzerID := nutzerkennung(t, mussMCP(t, umgebung, "schreiben.nutzer.neu",
		map[string]any{"name": "Zweiter"}))

	mussMCP(t, umgebung, "schreiben.nutzer.aendern", map[string]any{
		"nutzerID": nutzerID, "name": "Umbenannt", "fassung": "1",
	})

	profilID := profilkennungAus(t, mussMCP(t, umgebung, "schreiben.profil.neu",
		map[string]any{"nutzerID": nutzerID, "name": "Turnier"}))

	mussMCP(t, umgebung, "schreiben.profil.aendern", map[string]any{
		"profilID": profilID, "name": "Legacy", "fassung": "1",
		"subslogan": "Nur alte Formate",
	})

	mussMCP(t, umgebung, "schreiben.profil.loeschen",
		map[string]any{"profilID": profilID, "name": "Legacy"})

	mussMCP(t, umgebung, "schreiben.nutzer.loeschen",
		map[string]any{"nutzerID": nutzerID, "name": "Umbenannt"})

	// **Und das letzte Profil bleibt.**
	roh, istFehler := mcpRuf(t, umgebung, "schreiben.profil.loeschen",
		map[string]any{"profilID": "1", "name": "Mein Profil"})
	if !istFehler {
		t.Errorf("das letzte Profil ließ sich löschen: %s", roh)
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzePflegewegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	nutzerID := nutzerkennung(t,
		mussREST(t, an, http.MethodPost, "/nutzer", `{"name":"Zweiter"}`))

	mussREST(t, an, http.MethodPut, "/nutzer/"+nutzerID,
		`{"name":"Umbenannt","fassung":"1"}`)

	profilID := profilkennungAus(t, mussREST(t, an, http.MethodPost,
		"/nutzer/"+nutzerID+"/profile", `{"name":"Turnier"}`))

	mussREST(t, an, http.MethodPut, "/profile/"+profilID,
		`{"name":"Legacy","subslogan":"Nur alte Formate","fassung":"1"}`)
	mussREST(t, an, http.MethodDelete, "/profile/"+profilID+"?name=Legacy", "")
	mussREST(t, an, http.MethodDelete, "/nutzer/"+nutzerID+"?name=Umbenannt", "")

	// **Und ein falscher Name löscht nichts.**
	status, antwort := an.ruf(t, http.MethodDelete, "/nutzer/1?name=falsch", "")
	if status < 400 {
		t.Errorf("der falsche Name kam durch: %d (%s)", status, antwort)
	}
}

// nutzerkennung liest die Kennung aus einer Nutzerantwort.
func nutzerkennung(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Nutzer struct {
			ID string `json:"id"`
		} `json:"nutzer"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Nutzerantwort lesen: %v — %s", err, roh)
	}

	return gelesen.Nutzer.ID
}

// profilkennungAus liest die Kennung aus einer Profilantwort.
func profilkennungAus(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Profil struct {
			ID string `json:"id"`
		} `json:"profil"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Profilantwort lesen: %v — %s", err, roh)
	}

	return gelesen.Profil.ID
}
