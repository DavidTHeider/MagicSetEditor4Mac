// **Rotation, Widerruf mitten in der Sitzung, Umfang und Ablauf — an den
// gebauten Programmen** — Werkstück O2 von Phase 5.2
// (#1426),
// nach [0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md).
//
// # Warum hier
//
// **Weil die Zusage über Prozesse geht**: ein Token wird an der Kommandozeile
// angelegt, am Dienst vorgezeigt, in einer MCP-Sitzung benutzt und in
// derselben Sitzung zurückgezogen. Kein Test einer Schicht sieht das.
package flaechen_test

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// tokendateivariable ist die Umgebungsvariable, über die `mse-mcp` seine
// Tokendatei findet — dieselbe wie `Tokendateivariable` in `cmd/mse-mcp`.
// Kein Geheimnis, sondern sein Name — deshalb `#nosec G101`.
const tokendateivariable = "MSE_TOKEN_DATEI" // #nosec G101 -- der Name der Variablen

// technischesProfil legt ein technisches Profil an und gibt seine Kennung.
func technischesProfil(t *testing.T, umgebung []string) string {
	t.Helper()

	return profilkennungAus(t,
		mussKommandozeile(t, umgebung, "profil-neu", "1", "Nachtlauf", "--json"))
}

// tokenAnDerZeile legt ein Token an und gibt den Klartext.
func tokenAnDerZeile(t *testing.T, umgebung []string, profilID, scope string) string {
	t.Helper()

	return strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Bogenbau","scope":"`+scope+`"}`,
		"token-anlegen", profilID))
}

// tokendatei schreibt einen Klartext in eine Datei mit den Rechten 0600.
func tokendatei(t *testing.T, klartext string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "mcp-token")
	if err := os.WriteFile(pfad, []byte(klartext+"\n"), 0o600); err != nil {
		t.Fatalf("Tokendatei schreiben: %v", err)
	}

	return pfad
}

// **Rotiert an der Zeile, geprüft am Dienst: das alte Geheimnis gilt in der
// Frist weiter, das neue sofort — und der Sammelwiderruf sperrt beide.**
func TestRotationUndSammelwiderrufUeberZeileUndDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	profilID := technischesProfil(t, umgebung)
	altes := tokenAnDerZeile(t, umgebung, profilID, "lesen")

	tokenID := tokenkennungAus(t, mussKommandozeile(t, umgebung, "token", profilID, "--json"))
	neues := strings.TrimSpace(mussKommandozeile(t, umgebung, "token-rotieren", tokenID))

	if !strings.HasPrefix(neues, "mse_") || neues == altes {
		t.Fatalf("token-rotieren gab %q", neues)
	}

	an := starteDienst(t, umgebung)

	for name, klartext := range map[string]string{"das alte": altes, "das neue": neues} {
		if status, antwort := an.mitNachweis(klartext).ruf(t, http.MethodGet, "/sets", ""); status != http.StatusOK {
			t.Errorf("%s Geheimnis in der Frist: %d — %s", name, status, antwort)
		}
	}

	// **Die Liste nennt Principal, Nutzung und Frist** — und nie ein Geheimnis.
	liste := mussKommandozeile(t, umgebung, "token", profilID, "--json")
	if !strings.Contains(liste, "Rechner") || !strings.Contains(liste, `"vorgaengerBis": "20`) ||
		!strings.Contains(liste, `"letzteNutzung": "20`) || strings.Contains(liste, neues) {
		t.Errorf("die Liste: %s", liste)
	}

	// **Der Dienst rotiert auch** — und antwortet mit 201, wie beim Anlegen.
	status, rotiert := an.ruf(t, http.MethodPost, "/token/"+tokenID+"/rotation", "")
	if status != http.StatusCreated || !strings.Contains(rotiert, `"klartext"`) {
		t.Fatalf("POST /token/{id}/rotation: %d — %s", status, rotiert)
	}

	// **Die Ablaufwahlen an beiden.**
	if wahlen := mussKommandozeile(t, umgebung, "token-ablaufwahlen"); !strings.Contains(wahlen, "jahr\t") {
		t.Errorf("token-ablaufwahlen: %s", wahlen)
	}

	if wahlen := mussREST(t, an, http.MethodGet, "/token/ablaufwahlen", ""); !strings.Contains(wahlen, `"jahr"`) {
		t.Errorf("GET /token/ablaufwahlen: %s", wahlen)
	}

	// **Der Sammelwiderruf sperrt alles.**
	if gezogen := strings.TrimSpace(mussKommandozeile(t, umgebung, "token-alle-widerrufen", profilID)); gezogen != "1" {
		t.Errorf("token-alle-widerrufen zog %q, erwartet 1", gezogen)
	}

	for name, klartext := range map[string]string{"das alte": altes, "das neue": neues} {
		if status, _ := an.mitNachweis(klartext).ruf(t, http.MethodGet, "/sets", ""); status == http.StatusOK {
			t.Errorf("%s Geheimnis kam nach dem Sammelwiderruf noch herein", name)
		}
	}

	if status, antwort := an.ruf(t, http.MethodDelete, "/profile/"+profilID+"/token", ""); status != http.StatusOK ||
		!strings.Contains(antwort, `"gezogen":0`) {
		t.Errorf("DELETE /profile/{id}/token: %d — %s", status, antwort)
	}
}

// **Eine MCP-Sitzung mit Token: der Widerruf wirkt beim nächsten Ruf, in
// derselben Sitzung.** Und ein Token, das nur lesen darf, schreibt nicht.
func TestEinWiderrufWirktInDerLaufendenMCPSitzung(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	profilID := technischesProfil(t, umgebung)

	schreibend := tokenAnDerZeile(t, umgebung, profilID, "schreiben")
	mitToken := append(append([]string{}, umgebung...),
		tokendateivariable+"="+tokendatei(t, schreibend))

	zeilen := mcpSitzung(t, mitToken,
		mcpschritt{"lesen.sets", map[string]any{}},
		mcpschritt{"schreiben.token.widerrufen.alle", map[string]any{"profilID": profilID}},
		mcpschritt{"lesen.sets", map[string]any{}},
	)

	if strings.Contains(zeilen[0], `"isError":true`) {
		t.Fatalf("der erste Ruf mit gültigem Token scheiterte: %s", zeilen[0])
	}

	// **Der Inhalt ist eine JSON-Zeichenkette im JSON** — die Anführungszeichen
	// sind darin maskiert; gesucht wird deshalb das Wort und das Ergebnis.
	if strings.Contains(zeilen[1], `"isError":true`) || !strings.Contains(zeilen[1], "gezogen") {
		t.Fatalf("der Widerruf aus der Sitzung: %s", zeilen[1])
	}

	if !strings.Contains(zeilen[2], `"isError":true`) || !strings.Contains(zeilen[2], "gilt nicht") {
		t.Errorf("der Ruf nach dem Widerruf lief weiter: %s", zeilen[2])
	}

	// **Nur lesen heißt nur lesen** — auch über MCP.
	lesend := tokenAnDerZeile(t, umgebung, profilID, "lesen")
	nurLesend := append(append([]string{}, umgebung...),
		tokendateivariable+"="+tokendatei(t, lesend))

	antworten := mcpSitzung(t, nurLesend,
		mcpschritt{"lesen.sets", map[string]any{}},
		mcpschritt{"schreiben.set.neu", map[string]any{"name": "Neu", "spiel": "probe", "stil": "probe-schlicht"}},
	)

	if strings.Contains(antworten[0], `"isError":true`) {
		t.Errorf("das Lesetoken las nicht: %s", antworten[0])
	}

	if !strings.Contains(antworten[1], `"isError":true`) || !strings.Contains(antworten[1], "lesen") {
		t.Errorf("das Lesetoken schrieb: %s", antworten[1])
	}
}

// **Ein abgelaufenes Token trägt keine Sitzung, und eine Datei, die andere
// lesen dürfen, wird abgewiesen — bevor ein Ruf beantwortet wird.**
func TestAblaufUndOffeneTokendateiSperrenDenMCPProzess(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	profilID := technischesProfil(t, umgebung)

	vorbei := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Vorbei","scope":"lesen","ablauf":"2020-01-01T00:00:00Z"}`,
		"token-anlegen", profilID))

	abgelaufen := append(append([]string{}, umgebung...),
		tokendateivariable+"="+tokendatei(t, vorbei))

	antworten := mcpSitzung(t, abgelaufen, mcpschritt{"lesen.sets", map[string]any{}})
	if !strings.Contains(antworten[0], `"isError":true`) || !strings.Contains(antworten[0], "gilt nicht") {
		t.Errorf("das abgelaufene Token las: %s", antworten[0])
	}

	offen := filepath.Join(t.TempDir(), "offen")
	if err := os.WriteFile(offen, []byte(vorbei), 0o644); err != nil { //nolint:gosec // die Probe braucht eine offene Datei
		t.Fatalf("schreiben: %v", err)
	}

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(t.Context(), gebaut.mcp)
	lauf.Env = append(append([]string{}, umgebung...), tokendateivariable+"="+offen)
	lauf.Stdin = strings.NewReader("")

	var aus bytes.Buffer
	lauf.Stderr = &aus

	if err := lauf.Run(); err == nil || !strings.Contains(aus.String(), "für andere lesbar") {
		t.Errorf("die offene Tokendatei wurde angenommen: %v — %s", err, aus.String())
	}
}
