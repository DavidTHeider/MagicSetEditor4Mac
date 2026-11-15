package flaechen_test

// **Die Verwaltung einer Partie an jeder Fläche** — PLAY-03.AK-05
// (#1711).
//
// # Wogegen hier geprüft wird
//
// **Dass es die vier Wege an allen Flächen gibt** (§16) — und dass sie an
// jeder **dasselbe** tun. Ein Wächter sieht, dass eine Fähigkeit vorhanden
// ist; dass sie dasselbe bewirkt, sieht nur ein Lauf, der sie dreimal fährt.
//
// **Und dass die Ablehnung dieselbe ist**: eine beendete Partie bleibt an
// jeder Fläche stehen. Wer nur den Erfolgsweg prüft, hat die Hälfte
// ausgelassen (§10).

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// verwaltungsantwort ist der Ausgang, den alle drei Flächen schicken.
type verwaltungsantwort struct {
	Kennung string   `json:"kennung"`
	Inhalt  string   `json:"inhalt"`
	Befunde []string `json:"befunde"`
}

// **Gelöscht wird an jeder Fläche — und eine beendete bleibt an jeder.**
func TestDasLoeschenEinerPartieGiltAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	anlegen := func() string {
		return bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
			"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck,
			"--json")).ID
	}

	cli, mcp, rest := anlegen(), anlegen(), anlegen()

	mussKommandozeile(t, umgebung, "partie-loeschen", cli, "--json")
	mussMCP(t, umgebung, "schreiben.partie.loeschen",
		map[string]any{"partieID": mcp})
	mussREST(t, an, http.MethodDelete, "/partien/"+rest, "")

	// **Und danach ist keine mehr da** — die Liste ist die Gegenprobe.
	if zeilen := zeilenAusCLI(t, mussKommandozeile(
		t, umgebung, "partien", "--json")); len(zeilen) != 0 {
		t.Errorf("es stehen noch %d Partien da", len(zeilen))
	}

	// **Eine beendete bleibt** — an jeder Fläche, und mit einem Satz, der
	// sagt, warum.
	beendet := anlegen()

	mussKommandozeile(t, umgebung, "partie-stand", beendet, "anna", "beenden")

	if _, _, code := kommandozeile(t, umgebung,
		"partie-loeschen", beendet); code == 0 {
		t.Error("die Kommandozeile hat eine beendete Partie gelöscht")
	}

	if _, istFehler := mcpRuf(t, umgebung, "schreiben.partie.loeschen",
		map[string]any{"partieID": beendet}); !istFehler {
		t.Error("MCP hat eine beendete Partie gelöscht")
	}

	if status, rumpf := an.ruf(
		t, http.MethodDelete, "/partien/"+beendet, ""); status < 400 {
		t.Errorf("REST hat eine beendete Partie gelöscht: %d (%s)",
			status, rumpf)
	}
}

// **Kopiert wird an jeder Fläche — und die Kopie läuft, auch aus einer
// beendeten heraus.**
func TestDasKopierenEinerPartieGiltAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	vorlage := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json")).ID

	mussKommandozeile(t, umgebung, "partie-stand", vorlage, "anna", "beenden")

	kopien := map[string]verwaltungsantwort{
		"Kommandozeile": verwaltungAusCLI(t, mussKommandozeile(t, umgebung,
			"partie-kopieren", vorlage, "--json")),
		"MCP": verwaltungAusFlach(t, mussMCP(t, umgebung,
			"schreiben.partie.kopieren", map[string]any{"partieID": vorlage})),
		"REST": verwaltungAusFlach(t, mussREST(t, an, http.MethodPost,
			"/partien/"+vorlage+"/kopie", "")),
	}

	gesehen := map[string]bool{}

	for name, kopie := range kopien {
		switch {
		case kopie.Kennung == "" || kopie.Kennung == vorlage:
			t.Errorf("%s gab die Kennung %q", name, kopie.Kennung)
		case gesehen[kopie.Kennung]:
			t.Errorf("%s gab eine Kennung, die es schon gab", name)
		}

		gesehen[kopie.Kennung] = true
	}

	// **Jede Kopie läuft** — die Vorlage war beendet.
	for _, zeile := range zeilenAusCLI(t, mussKommandozeile(
		t, umgebung, "partien", "--json")) {
		if zeile.ID != vorlage && zeile.Stand != "laeuft" {
			t.Errorf("die Kopie %s steht auf %q", zeile.ID, zeile.Stand)
		}
	}
}

// **Ausgeführt und eingelesen wird an jeder Fläche — und die Datei ist
// dieselbe.**
//
// **Zwei Gestalten und eine Bedeutung** (§16): die Kommandozeile bekommt
// einen **Ort**, MCP und REST den **Text**. Was darin steht, ist dasselbe.
func TestDieAusfuhrEinerPartieGiltAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	vorlage := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json")).ID

	ausMCP := verwaltungAusFlach(t, mussMCP(t, umgebung, "lesen.partie.ausfuhr",
		map[string]any{"partieID": vorlage}))
	ausREST := verwaltungAusFlach(t, mussREST(t, an, http.MethodGet,
		"/partien/"+vorlage+"/ausfuhr", ""))

	if ausMCP.Inhalt != ausREST.Inhalt {
		t.Errorf("MCP und REST führen verschieden aus:\n%q\n%q",
			ausMCP.Inhalt, ausREST.Inhalt)
	}

	// **Die Kommandozeile schreibt dieselbe Datei.**
	ort := filepath.Join(t.TempDir(), "partie.msepartie")

	mussKommandozeile(t, umgebung, "partie-ausfuhr", vorlage, ort, "--json")

	rohDatei, lesefehler := os.ReadFile(ort) //nolint:gosec // der Ort kommt aus dem Lauf
	if lesefehler != nil {
		t.Fatalf("die ausgeführte Datei lesen: %v", lesefehler)
	}

	ausDatei := string(rohDatei)

	if ausDatei != ausMCP.Inhalt {
		t.Errorf("die Datei und MCP führen verschieden aus:\n%q\n%q",
			ausDatei, ausMCP.Inhalt)
	}

	// **Und jede Fläche liest sie wieder ein** — als neue Partie.
	eingelesen := map[string]verwaltungsantwort{
		"Kommandozeile": verwaltungAusCLI(t, mussKommandozeile(t, umgebung,
			"partie-einfuhr", ort, "--json")),
		"MCP": verwaltungAusFlach(t, mussMCP(t, umgebung,
			"schreiben.partie.einfuhr", map[string]any{"inhalt": ausDatei})),
		"REST": verwaltungAusFlach(t, mussREST(t, an, http.MethodPost,
			"/partien/einfuhr", einfuhrrumpf(t, ausDatei))),
	}

	for name, antwort := range eingelesen {
		if antwort.Kennung == "" || antwort.Kennung == vorlage {
			t.Errorf("%s gab die Kennung %q", name, antwort.Kennung)
		}
	}

	// **Vier Partien**: die Vorlage und drei Einfuhren.
	if zeilen := zeilenAusCLI(t, mussKommandozeile(
		t, umgebung, "partien", "--json")); len(zeilen) != 4 {
		t.Errorf("es stehen %d Partien da", len(zeilen))
	}
}

// einfuhrrumpf baut den Rumpf der REST-Einfuhr.
func einfuhrrumpf(t *testing.T, inhalt string) string {
	t.Helper()

	roh, fehler := json.Marshal(map[string]string{"inhalt": inhalt})
	if fehler != nil {
		t.Fatalf("den Rumpf bauen: %v", fehler)
	}

	return string(roh)
}

// verwaltungAusCLI liest den Ausgang aus der Hülle der Kommandozeile.
func verwaltungAusCLI(t *testing.T, roh string) verwaltungsantwort {
	t.Helper()

	return verwaltungAusFlach(t, roh)
}

// verwaltungAusFlach liest ihn aus einer flachen Antwort.
func verwaltungAusFlach(t *testing.T, roh string) verwaltungsantwort {
	t.Helper()

	var antwort verwaltungsantwort

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return antwort
}
