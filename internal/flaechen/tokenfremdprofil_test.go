// **Ein Token wirkt nur an seinem Profil** — QA-09
// (#1740),
// Befund P2 in `docs/PENTEST.md`.
//
// # Der Angriff
//
// Ein Automat bekommt ein Token für **sein** Profil — „Nachtlauf", Erlaubnis
// `schreiben`. Liegt es in einer Datei, die jemand liest, soll der Schaden an
// diesem Profil enden: „ein Token reicht nie weiter als die Rolle seines
// Profils" (`docs/AUTORISIERUNG.md`). Hier versucht es, sich **im Profil des
// Menschen** ein Token anzulegen, dort Vorgaben zu setzen und die Token dort
// zurückzuziehen — und mit dem erbeuteten Token im fremden Profil zu
// schreiben.
//
// # Woran gemessen
//
// **An der Wirkung** (QA-09.AK-04): ob im fremden Profil ein Token, eine
// Vorgabe oder ein Deck entstand, gelesen mit der Prozessmarke. Und
// derselbe Automat legt sich **im eigenen** Profil ein Token an — sonst hieße
// ein Nein nur, dass der Weg gar nicht ging (§10).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEinTokenWirktNurAnSeinemProfil(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	gezeigt := mussKommandozeile(t, umgebung, "profil", "--json")
	fremd, mensch := profilkennungAus(t, gezeigt), eignerkennung(t, gezeigt)
	eigenes := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Nachtlauf", "--json"))

	automat := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"schreiben"}`,
		"token-anlegen", eigenes))

	an := starteDienst(t, umgebung)
	alsAutomat := an.mitNachweis(automat)

	tokenVorher := tokenzahl(t, an, fremd)
	vorgabenVorher := abgefragt(t, an, "/profile/"+fremd+"/vorgaben")
	nutzerVorher := abgefragt(t, an, "/nutzer")

	// **Angriff 1: ein Token im fremden Profil.**
	status, antwort := alsAutomat.ruf(t, http.MethodPost, "/profile/"+fremd+"/token",
		`{"principal":"Eindringling","bezeichnung":"Hintertür","scope":"schreiben"}`)
	if status < http.StatusBadRequest {
		t.Errorf("ein Token im fremden Profil entstand: %d — %s", status, antwort)

		// **Und was es kann**: im fremden Profil schreiben.
		if beute := klartextAus(t, antwort); beute != "" {
			if status, antwort := an.mitNachweis(beute).ruf(t, http.MethodPost, "/decks",
				`{"name":"Beute","spiel":"magic"}`); status < http.StatusBadRequest {
				t.Errorf("mit dem erbeuteten Token entstand ein Deck: %d — %s", status, antwort)
			}
		}
	}

	// **Angriff 2: eine Vorgabe im fremden Profil.**
	if status, antwort := alsAutomat.ruf(t, http.MethodPut,
		"/profile/"+fremd+"/vorgaben/spiel", `{"werte":["magic"]}`); status < http.StatusBadRequest {
		t.Errorf("eine Vorgabe im fremden Profil wurde gesetzt: %d — %s", status, antwort)
	}

	// **Angriff 3: die Token des fremden Profils zurückziehen.**
	if status, antwort := alsAutomat.ruf(t, http.MethodDelete,
		"/profile/"+fremd+"/token", ""); status < http.StatusBadRequest {
		t.Errorf("die Token des fremden Profils wurden zurückgezogen: %d — %s", status, antwort)
	}

	// **Angriff 4: den Menschen umbenennen und sein Profil löschen** — was
	// einem Menschen gehört, verwaltet kein Automat.
	if status, antwort := alsAutomat.ruf(t, http.MethodPut, "/nutzer/"+mensch,
		`{"name":"Übernommen"}`); status < http.StatusBadRequest {
		t.Errorf("der Mensch wurde umbenannt: %d — %s", status, antwort)
	}

	// **Und die Konten des Menschen liest er auch nicht** (QA-09.AK-03).
	for _, weg := range []string{"/nutzer/" + mensch + "/konten", "/nutzer/" + mensch + "/einstellungen"} {
		if status, antwort := alsAutomat.ruf(t, http.MethodGet, weg, ""); status < http.StatusBadRequest {
			t.Errorf("%s gab einem Automaten Auskunft: %d — %s", weg, status, antwort)
		}
	}

	if status, antwort := alsAutomat.ruf(t, http.MethodDelete,
		"/profile/"+fremd, ""); status < http.StatusBadRequest {
		t.Errorf("das fremde Profil wurde gelöscht: %d — %s", status, antwort)
	}

	// **An der Wirkung**: im fremden Profil ist alles, wie es war.
	if nachher := abgefragt(t, an, "/nutzer"); nachher != nutzerVorher {
		t.Errorf("die Nutzer änderten sich:\nvorher  %s\nnachher %s", nutzerVorher, nachher)
	}

	if nachher := tokenzahl(t, an, fremd); nachher != tokenVorher {
		t.Errorf("im fremden Profil stehen %d Token, vorher %d", nachher, tokenVorher)
	}

	if nachher := abgefragt(t, an, "/profile/"+fremd+"/vorgaben"); nachher != vorgabenVorher {
		t.Errorf("die Vorgaben des fremden Profils änderten sich:\nvorher  %s\nnachher %s",
			vorgabenVorher, nachher)
	}

	// **Die Gegenprobe**: im eigenen Profil darf der Automat sich ein Token
	// einrichten — so steht es in `internal/rest/token.go`.
	if status, antwort := alsAutomat.ruf(t, http.MethodPost, "/profile/"+eigenes+"/token",
		`{"principal":"Rechner","bezeichnung":"Ersatz","scope":"lesen"}`); status != http.StatusCreated {
		t.Errorf("im eigenen Profil entstand kein Token: %d — %s", status, antwort)
	}
}

// tokenzahl zählt die Token eines Profils — mit der Prozessmarke gelesen.
func tokenzahl(t *testing.T, an dienst, profilID string) int {
	t.Helper()

	status, antwort := an.ruf(t, http.MethodGet, "/profile/"+profilID+"/token", "")
	if status != http.StatusOK {
		t.Fatalf("die Token von Profil %s lesen: %d — %s", profilID, status, antwort)
	}

	var gelesen struct {
		Token []json.RawMessage `json:"token"`
	}

	if fehler := json.Unmarshal([]byte(antwort), &gelesen); fehler != nil {
		t.Fatalf("die Token sind kein JSON: %v — %s", fehler, antwort)
	}

	return len(gelesen.Token)
}

// klartextAus liest den Klartext aus einer Anlegeantwort — leer, wenn keiner
// darin steht.
func klartextAus(t *testing.T, antwort string) string {
	t.Helper()

	var gelesen struct {
		Klartext string `json:"klartext"`
	}

	if json.Unmarshal([]byte(antwort), &gelesen) != nil {
		return ""
	}

	return gelesen.Klartext
}
