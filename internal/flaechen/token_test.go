// **Zugriffstoken — an jeder Fläche** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkte 12 und 13.
//
// # Der Ablauf
//
// Technisches Profil anlegen → Token anlegen → auflisten → umbenennen →
// zurückziehen. **Und der Lauf, den es nur hier gibt**: ein an der
// Kommandozeile angelegtes Token wird dem **gebauten** Dienst vorgezeigt und
// gilt dort an der Stelle der Prozessmarke.
//
// **Das ist die Abnahme dieses Werkstücks** und nicht „die Tests sind grün":
// dass ein Automat mit einem selbst angelegten Token hereinkommt, sieht man
// erst, wenn zwei Prozesse beteiligt sind.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestDerGanzeTokenwegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	profilID := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Nachtlauf", "--json"))

	// **Erst technisch machen, dann ein Token anlegen.** Die Reihenfolge ist
	// keine Vorschrift des Bestands — sie ist die Reihenfolge, in der ein
	// Mensch das täte, und der Lauf soll den Weg gehen und nicht die Tabelle.
	mitEingabe(t, umgebung,
		`{"name":"Nachtlauf","technisch":true,"fassung":"1"}`,
		"profil-aendern", profilID)

	klartext := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Bogenbau","scope":"lesen"}`, "token-anlegen", profilID))

	if !strings.HasPrefix(klartext, "mse_") {
		t.Fatalf("der Klartext sieht nicht wie ein Token aus: %q", klartext)
	}

	tokenID := tokenkennungAus(t,
		mussKommandozeile(t, umgebung, "token", profilID, "--json"))

	mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Tagesbau","scope":"schreiben","fassung":"1"}`,
		"token-aendern", tokenID)

	aus := mussKommandozeile(t, umgebung, "token", profilID)
	if !strings.Contains(aus, "Tagesbau") || !strings.Contains(aus, "schreiben") {
		t.Errorf("die Änderung steht nicht in der Liste: %q", aus)
	}

	t.Run("und das Geheimnis steht nirgends darin", func(t *testing.T) {
		// **Die Zusage, für die es die gehütete Form gibt.** Wer die Liste
		// liest, findet kein Token, mit dem er hereinkäme.
		if strings.Contains(aus, klartext) {
			t.Errorf("die Liste trägt den Klartext: %q", aus)
		}
	})

	mussKommandozeile(t, umgebung, "token-loeschen", tokenID)

	if nachher := mussKommandozeile(t, umgebung, "token", profilID); strings.Contains(
		nachher, "Tagesbau") {
		t.Errorf("das gelöschte Token steht noch da: %q", nachher)
	}
}

// **Derselbe Weg über MCP.**
func TestDerGanzeTokenwegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	profilID := profilkennungAus(t, mussMCP(t, umgebung, "schreiben.profil.neu",
		map[string]any{"nutzerID": "1", "name": "Nachtlauf"}))

	mussMCP(t, umgebung, "schreiben.profil.aendern", map[string]any{
		"profilID": profilID, "name": "Nachtlauf", "technisch": true,
		"fassung": "1",
	})

	angelegt := mussMCP(t, umgebung, "schreiben.token.anlegen", map[string]any{
		"profilID": profilID, "bezeichnung": "Bogenbau", "principal": "Rechner", "scope": "lesen",
	})

	if !strings.Contains(angelegt, `"klartext"`) || !strings.Contains(angelegt, "mse_") {
		t.Fatalf("die Antwort trägt keinen Klartext: %s", angelegt)
	}

	tokenID := tokenkennungAus(t,
		mussMCP(t, umgebung, "lesen.token", map[string]any{"profilID": profilID}))

	mussMCP(t, umgebung, "schreiben.token.aendern", map[string]any{
		"tokenID": tokenID, "bezeichnung": "Tagesbau", "principal": "Rechner", "scope": "schreiben",
		"fassung": "1",
	})

	mussMCP(t, umgebung, "schreiben.token.loeschen",
		map[string]any{"tokenID": tokenID})

	nachher := mussMCP(t, umgebung, "lesen.token",
		map[string]any{"profilID": profilID})
	if strings.Contains(nachher, "Tagesbau") {
		t.Errorf("das gelöschte Token steht noch da: %s", nachher)
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeTokenwegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	profilID := profilkennungAus(t, mussREST(t, an, http.MethodPost,
		"/nutzer/1/profile", `{"name":"Nachtlauf"}`))

	mussREST(t, an, http.MethodPut, "/profile/"+profilID,
		`{"name":"Nachtlauf","technisch":true,"fassung":"1"}`)

	angelegt := mussREST(t, an, http.MethodPost, "/profile/"+profilID+"/token",
		`{"principal":"Rechner","bezeichnung":"Bogenbau","scope":"lesen"}`)

	if klartext := ausRest(t, angelegt, "klartext"); !strings.HasPrefix(
		klartext, "mse_") {
		t.Fatalf("der Klartext sieht nicht wie ein Token aus: %q", klartext)
	}

	tokenID := tokenkennungAus(t,
		mussREST(t, an, http.MethodGet, "/profile/"+profilID+"/token", ""))

	mussREST(t, an, http.MethodPut, "/token/"+tokenID,
		`{"principal":"Rechner","bezeichnung":"Tagesbau","scope":"schreiben","fassung":"1"}`)
	mussREST(t, an, http.MethodDelete, "/token/"+tokenID, "")

	nachher := mussREST(t, an, http.MethodGet, "/profile/"+profilID+"/token", "")
	if strings.Contains(nachher, "Tagesbau") {
		t.Errorf("das gelöschte Token steht noch da: %s", nachher)
	}
}

// **Und die Abnahme: ein Token an der Stelle der Prozessmarke.**
//
// **Angelegt an der Kommandozeile, vorgezeigt am Dienst** — zwei Prozesse,
// ein Bestand. Kein Test einer Schicht sieht das: dort ist die Wache eine
// Schnittstelle und der Klartext eine Zeichenkette, die der Lauf selbst
// erfunden hat.
func TestEinAngelegtesTokenLaesstAmDienstHerein(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	profilID := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Nachtlauf", "--json"))

	nurLesen := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Bogenbau","scope":"lesen"}`, "token-anlegen", profilID))

	darfSchreiben := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Setzer","scope":"schreiben"}`, "token-anlegen", profilID))

	an := starteDienst(t, umgebung)

	t.Run("es liest", func(t *testing.T) {
		status, antwort := an.mitNachweis(nurLesen).ruf(
			t, http.MethodGet, "/sets", "")
		if status != http.StatusOK {
			t.Fatalf("Status %d — %s", status, antwort)
		}
	})

	t.Run("es schreibt nicht", func(t *testing.T) {
		status, antwort := an.mitNachweis(nurLesen).ruf(
			t, http.MethodPost, "/sets/neu", `{"name":"Neu","spiel":"probe","stil":"probe-schlicht"}`)
		if status != http.StatusForbidden {
			t.Fatalf("Status %d, erwartet 403 — %s", status, antwort)
		}

		// **Die eine Ablehnung, die mehr sagt als „nein"**: wer ein gültiges
		// Token hat, ist kein Fremder, und ohne diesen Satz suchte er den
		// Fehler bei seiner Anfrage.
		if !strings.Contains(antwort, "lesen") {
			t.Errorf("die Meldung sagt nicht, woran es lag: %s", antwort)
		}
	})

	t.Run("und das schreibende schreibt", func(t *testing.T) {
		// **Der Beleg, dass die Ablehnung darüber am Scope hing** und nicht
		// daran, dass der Dienst Token grundsätzlich nicht schreiben lässt
		// (§10).
		status, antwort := an.mitNachweis(darfSchreiben).ruf(
			t, http.MethodPost, "/sets/neu", `{"name":"Neu","spiel":"probe","stil":"probe-schlicht"}`)
		if status != http.StatusCreated {
			t.Fatalf("Status %d, erwartet 201 — %s", status, antwort)
		}
	})

	t.Run("und was es anlegt, liegt in SEINEM Profil", func(t *testing.T) {
		// **Das ist der Grund, aus dem es technische Profile gibt** (Punkt 12):
		// wer einen Dienst automatisch bedienen lässt, will nicht, dass er in
		// **seinem** Profil schreibt.
		//
		// **Und es ist der Fall, den keine Gegenprobe darüber erreicht.** Als
		// die Wache den Kontext gar nicht austauschte, blieb jeder andere
		// Lauf hier grün — sie prüfen, **ob** hereingelassen wird, nicht
		// **wohin**.
		mitToken := mussREST(t, an.mitNachweis(darfSchreiben),
			http.MethodGet, "/sets", "")
		if !strings.Contains(mitToken, "Neu") {
			t.Errorf("das angelegte Set fehlt im Profil des Tokens: %s", mitToken)
		}

		mitMarke := mussREST(t, an, http.MethodGet, "/sets", "")
		if strings.Contains(mitMarke, "Neu") {
			t.Errorf("das Set des Automaten steht im gezeigten Profil: %s", mitMarke)
		}
	})

	t.Run("ein erfundenes gilt nicht", func(t *testing.T) {
		status, _ := an.mitNachweis("mse_99_ERFUNDEN").ruf(
			t, http.MethodGet, "/sets", "")
		if status != http.StatusUnauthorized {
			t.Errorf("Status %d, erwartet 401", status)
		}
	})

	t.Run("und die Prozessmarke gilt weiterhin", func(t *testing.T) {
		// **Die verträgliche Hälfte der Änderung** (§16): wer heute eine
		// Marke schickt, merkt nichts.
		if status, antwort := an.ruf(t, http.MethodGet, "/sets", ""); status != http.StatusOK {
			t.Fatalf("Status %d — %s", status, antwort)
		}
	})
}

// tokenkennungAus liest die erste Tokenkennung aus einer Antwort.
func tokenkennungAus(t *testing.T, roh string) string {
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
