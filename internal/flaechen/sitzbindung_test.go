// **Ein Werkzeug, das an einem Sitz spielt, sieht nur diesen Sitz** —
// #1819,
// Befund 2 des Zweitblicks über Phase 6.1.
//
// # Die Abnahme
//
//   - Ein an Sitz A gebundenes Token liest die Sicht von A; `fuer=B` gibt
//     `erlaubnis` (403, an MCP den Satz dazu), der Verlauf ebenso, die
//     Ausfuhr ebenso.
//   - Ein ungebundenes Token verhält sich wie heute (Gegenprobe).
//   - Lauf an MCP und Dienst.
//
// **Die Partie steht im Profil der Token** — ein Token arbeitet in seinem
// Profil, und nur dort gibt es sie.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sitzumgebung legt ein technisches Profil mit einer Partie an und gibt
// zwei Token dazu: eines an den Sitz `anna` gebunden, eines frei.
func sitzumgebung(t *testing.T) (umgebung []string, partieID, gebunden, frei string) {
	t.Helper()

	umgebung = deckumgebung(t)
	profilID := technischesProfil(t, umgebung)
	mussKommandozeile(t, umgebung, "profil", profilID)

	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)
	partieID = bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json")).ID

	gebunden = strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Modell","bezeichnung":"Am Sitz","scope":"schreiben",`+
			`"partie":"`+partieID+`","sitz":"anna"}`,
		"token-anlegen", profilID))
	frei = tokenAnDerZeile(t, umgebung, profilID, "schreiben")

	return umgebung, partieID, gebunden, frei
}

func TestEinGebundenesTokenAmDienst(t *testing.T) {
	t.Parallel()

	umgebung, partieID, gebunden, frei := sitzumgebung(t)
	an := starteDienst(t, umgebung)

	for _, fall := range []struct {
		name   string
		token  string
		weg    string
		status int
	}{
		{"die eigene Sicht", gebunden, "/partien/" + partieID + "?fuer=anna", http.StatusOK},
		{"eine fremde Sicht", gebunden, "/partien/" + partieID + "?fuer=bert", http.StatusForbidden},
		{"der eigene Verlauf", gebunden, "/partien/" + partieID + "/verlauf?fuer=anna", http.StatusOK},
		{"ein fremder Verlauf", gebunden, "/partien/" + partieID + "/verlauf?fuer=bert", http.StatusForbidden},
		{"die Ausfuhr", gebunden, "/partien/" + partieID + "/ausfuhr", http.StatusForbidden},

		// **Die Gegenprobe**: frei geht alles wie bisher.
		{"frei: eine fremde Sicht", frei, "/partien/" + partieID + "?fuer=bert", http.StatusOK},
		{"frei: ein fremder Verlauf", frei, "/partien/" + partieID + "/verlauf?fuer=bert", http.StatusOK},
		{"frei: die Ausfuhr", frei, "/partien/" + partieID + "/ausfuhr", http.StatusOK},
	} {
		status, antwort := an.mitNachweis(fall.token).ruf(t, http.MethodGet, fall.weg, "")
		if status != fall.status {
			t.Errorf("%s: %d, erwartet %d — %s", fall.name, status, fall.status, antwort)
		}
	}

	// **Und ein Befehl** — im Namen des fremden Sitzes abgewiesen, im
	// eigenen ausgeführt.
	for _, befehl := range []struct {
		absender string
		status   int
	}{
		{"bert", http.StatusForbidden},
		{"anna", http.StatusOK},
	} {
		status, antwort := an.mitNachweis(gebunden).ruf(t, http.MethodPost,
			"/partien/"+partieID+"/befehle",
			`{"kennung":"b-`+befehl.absender+`","absender":"`+befehl.absender+
				`","art":"spieler.muenzwurf"}`)
		if status != befehl.status {
			t.Errorf("Befehl als %s: %d, erwartet %d — %s",
				befehl.absender, status, befehl.status, antwort)
		}
	}
}

func TestEinGebundenesTokenAnMCP(t *testing.T) {
	t.Parallel()

	umgebung, partieID, gebunden, frei := sitzumgebung(t)

	fall := func(token, werkzeug string, argumente map[string]any) (string, bool) {
		return mcpRuf(t, append(append([]string{}, umgebung...),
			tokendateivariable+"="+tokendatei(t, token)), werkzeug, argumente)
	}

	if text, istFehler := fall(gebunden, "lesen.partie",
		map[string]any{"partieID": partieID, "fuer": "anna"}); istFehler {
		t.Errorf("die eigene Sicht: %s", text)
	}

	for _, abgewiesen := range []struct {
		name, werkzeug string
		argumente      map[string]any
	}{
		{"eine fremde Sicht", "lesen.partie", map[string]any{"partieID": partieID, "fuer": "bert"}},
		{"ein fremder Verlauf", "lesen.partieverlauf", map[string]any{"partieID": partieID, "fuer": "bert"}},
		{"die Ausfuhr", "lesen.partie.ausfuhr", map[string]any{"partieID": partieID}},
	} {
		// **MCP gibt den Satz und keinen Code** — der Satz von
		// [domain.ErrErlaubnis] ist deshalb der Beleg für die Klasse, und
		// nicht etwa „Partie nicht gefunden".
		text, istFehler := fall(gebunden, abgewiesen.werkzeug, abgewiesen.argumente)
		if !istFehler || !strings.Contains(text, domain.ErrErlaubnis.Error()) {
			t.Errorf("%s: %s", abgewiesen.name, text)
		}
	}

	// **Die Gegenprobe**: frei sieht auch die fremde Sicht.
	if text, istFehler := fall(frei, "lesen.partie",
		map[string]any{"partieID": partieID, "fuer": "bert"}); istFehler {
		t.Errorf("frei: %s", text)
	}
}
