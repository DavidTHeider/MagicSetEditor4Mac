// **Der Arbeitskontext an allen vier Flächen** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Warum an den gebauten Programmen
//
// **Weil jede Fläche ihren Kontext selbst ausstattet** — Modus, Profil, bei
// einem Token seinen Umfang — und ein Lauf gegen ein Doppel die Verdrahtung
// nicht sähe. Das Fenster fährt dieselbe Frage in `internal/app`.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// arbeitskontext ist die Antwort, wie jede Fläche sie gibt.
type arbeitskontext struct {
	NutzerID  string `json:"nutzerID"`
	ProfilID  string `json:"profilID"`
	EignerID  string `json:"eignerID"`
	Rolle     string `json:"rolle"`
	NurLesend bool   `json:"nurLesend"`
	Modus     string `json:"modus"`
	Revision  int64  `json:"revision"`
}

// kontextAus liest die Antwort einer Fläche — flach an der Kommandozeile,
// unter `kontext` bei MCP und Dienst.
func kontextAus(t *testing.T, name, roh string, umschlossen bool) arbeitskontext {
	t.Helper()

	var gelesen arbeitskontext

	if umschlossen {
		var umschlag struct {
			Kontext arbeitskontext `json:"kontext"`
		}

		if err := json.Unmarshal([]byte(roh), &umschlag); err != nil {
			t.Fatalf("%s: die Antwort lesen: %v — %s", name, err, roh)
		}

		return umschlag.Kontext
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("%s: die Antwort lesen: %v — %s", name, err, roh)
	}

	return gelesen
}

// **Dieselbe Frage, dieselbe Antwort — und jede Fläche nennt sich selbst.**
func TestDerArbeitskontextAnAllenFlaechen(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	antworten := map[string]arbeitskontext{
		"kommandozeile": kontextAus(t, "die Kommandozeile",
			mussKommandozeile(t, umgebung, "kontext", "--json"), false),
		"mcp": kontextAus(t, "MCP",
			mussMCP(t, umgebung, "lesen.kontext", map[string]any{}), true),
		"dienst": kontextAus(t, "der Dienst",
			mussREST(t, an, http.MethodGet, "/kontext", ""), true),
	}

	revision := antworten["kommandozeile"].Revision

	for modus, kontext := range antworten {
		if kontext.Modus != modus {
			t.Errorf("%s nennt sich %q", modus, kontext.Modus)
		}

		if kontext.Rolle != "verwalter" || kontext.NurLesend {
			t.Errorf("%s: im eigenen Profil %+v — erwartet der Verwalter, nicht nur lesend", modus, kontext)
		}

		if kontext.NutzerID == "" || kontext.NutzerID != kontext.EignerID || kontext.ProfilID == "" {
			t.Errorf("%s: Nutzer %q, Eigner %q, Profil %q", modus, kontext.NutzerID, kontext.EignerID, kontext.ProfilID)
		}

		// **Ein Bestand, ein Stand**: nichts hat sich zwischen den drei Rufen
		// geändert, also sagen alle drei dieselbe Revision.
		if kontext.Revision != revision {
			t.Errorf("%s: Revision %d, die Kommandozeile sagt %d", modus, kontext.Revision, revision)
		}
	}

	// **Und ein Wechsel hebt sie** — sichtbar an jeder Fläche. **In ein
	// anderes Profil**: wer schon dort ist, wird es nicht noch einmal, und der
	// Ruf schreibt dann gar nicht.
	zweites := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Zweites", "--json"))
	mussKommandozeile(t, umgebung, "profil", zweites)

	danach := kontextAus(t, "der Dienst danach", mussREST(t, an, http.MethodGet, "/kontext", ""), true)
	if danach.Revision <= revision {
		t.Errorf("nach dem Wechsel Revision %d, vorher %d", danach.Revision, revision)
	}

	// **Und eine Angabe zu viel ist ein Bedienfehler.**
	if _, _, code := kommandozeile(t, umgebung, "kontext", "zuviel"); code == 0 {
		t.Error("kontext nahm eine Angabe an")
	}
}

// **Ein Token, das nur lesen darf, macht den Kontext nur lesend** — gleich
// welche Rolle das Profil trägt. Das ist die Schnittmenge aus 0078: ein Token
// erweitert nie, was die Rolle hergibt.
func TestEinLesetokenMachtDenKontextNurLesend(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	profilID := profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Nachtlauf", "--json"))

	nurLesen := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Bogenbau","scope":"lesen"}`, "token-anlegen", profilID))

	an := starteDienst(t, umgebung)

	status, antwort := an.mitNachweis(nurLesen).ruf(t, http.MethodGet, "/kontext", "")
	if status != http.StatusOK {
		t.Fatalf("Status %d — %s", status, antwort)
	}

	kontext := kontextAus(t, "der Dienst mit Token", antwort, true)

	if kontext.ProfilID != profilID || kontext.Modus != "dienst" {
		t.Errorf("das Token arbeitet in %q über %q, erwartet %q über dienst", kontext.ProfilID, kontext.Modus, profilID)
	}

	if !kontext.NurLesend || kontext.Rolle != "verwalter" {
		t.Errorf("mit Lesetoken: %+v — erwartet der Verwalter, aber nur lesend", kontext)
	}
}
