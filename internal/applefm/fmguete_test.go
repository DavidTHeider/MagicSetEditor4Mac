//go:build fmlauf

// **Die Güte am Testsatz** — AI-03.AK-04
// (#1744):
// „Qualität, Fehlvorschläge und mehrsprachige Beispiele werden an einem
// fachlichen Testsatz bewertet." Gemessen am Gerätemodell, weil es der
// bevorzugte Dienst für genau diese Aufgaben ist (AI-05).
//
// **Der Testsatz ist selbst geschrieben** — acht Karten mit eigenem Inhalt,
// keine fremden Texte (§11). Je Karte werden Name und Flavortext erfragt,
// auf Deutsch und auf Englisch, mit der Anweisung der Anwendung und den
// Sätzen des Formulars. **Kontext und Schema sind nachgebaut** (Z6 des
// Zweitblicks über Phase 6.10): zwei Zeilen `type` und `rule text` statt der
// Textfelder einer Vorlage, das Schema mit dem einen Pflichtfeld — die
// Zahlen gelten für diese Form, und der Beleg sagt es.
//
// **Was der Lauf entscheidet und was nicht**: er zählt, was messbar ist —
// eine Antwort kam, sie war leer, sie wurde abgelehnt, sie hielt die
// Sprache. **Ob ein Name passt, entscheidet er nicht**: die Spalte dafür
// füllt, wer bewertet, und der Beleg nennt ihn.
//
//	go test -tags fmlauf -run TestDieGueteAmTestsatz -v ./internal/applefm/

package applefm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
)

// testkarte ist eine Karte des Testsatzes — Typ und Text, wie sie im
// Auftrag stehen.
type testkarte struct {
	Typ  string `json:"typ"`
	Text string `json:"text"`
}

// testsatzdatei ist `testdaten/ki/testsatz.json` — **seit AI-02 eine Datei
// für zwei Läufe**: denselben Satz fragt `internal/app/modellvergleich_test.go`
// die kleinen Ollama-Modelle, und ein Vergleich mit einer Abschrift wäre ein
// Vergleich mit etwas anderem.
type testsatzdatei struct {
	Karten      []testkarte                  `json:"karten"`
	Auftraege   map[string]map[string]string `json:"auftraege"`
	Kartebisher map[string]string            `json:"kartebisher"`
}

// derTestsatz liest die Datei.
func derTestsatz(t *testing.T) testsatzdatei {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "testdaten", "ki", "testsatz.json"))
	if fehler != nil {
		t.Fatalf("der Testsatz: %v", fehler)
	}

	var satz testsatzdatei
	if fehler := json.Unmarshal(roh, &satz); fehler != nil {
		t.Fatalf("der Testsatz: %v", fehler)
	}

	return satz
}

func TestDieGueteAmTestsatz(t *testing.T) {
	ctx := t.Context()
	satz := derTestsatz(t)

	if einrichtung := applefm.Erkenne(ctx, applefm.Vorhanden, applefm.Exitcode); einrichtung != applefm.EinrichtungBereit {
		t.Fatalf("BLOCKED: das Gerätemodell ist auf diesem Rechner nicht bereit (%s)", einrichtung)
	}

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	dienst := applefm.NeuerDienst(ordner, applefm.Prozessstarter)
	t.Cleanup(func() { _ = dienst.Beende() }) //nolint:errcheck // der Lauf ist vorbei

	klient, fehler := dienst.Klient(ctx)
	if fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	zeilen := make([]string, 0, len(satz.Karten)*4)
	zahlen := map[string]int{}

	// erste ist die erste Frage, wie sie hinausgeht — N6 des Nachlaufs über
	// Phase 6.10: jeder Lauf schreibt sie samt den Grenzen in den Beleg.
	erste, schema := gueteauftrag(satz, satz.Karten[0], "de", "name")
	if form, fehler := json.Marshal(schema); fehler == nil {
		erste += "\n\nSchema: " + string(form)
	}

	for nummer, karte := range satz.Karten {
		for _, sprache := range []string{"de", "en"} {
			for _, feld := range []string{"name", "flavor text"} {
				zeile, art := guetefrage(ctx, t, klient, satz, karte, sprache, feld)
				zeilen = append(zeilen, fmt.Sprintf("| %d | %s | %s | %s", nummer+1, sprache, feld, zeile))
				zahlen[art]++
			}
		}
	}

	gesamt := len(satz.Karten) * 4

	text := fmt.Sprintf(`# Güte des Gerätemodells am Testsatz

**%s** — AI-03.AK-04 (#1744), %d Fragen: acht selbst geschriebene Karten, Name
und Flavortext, Deutsch und Englisch.

go test -tags fmlauf -run TestDieGueteAmTestsatz -v ./internal/applefm/

**Was davon das Formular ist und was nachgebaut**: die Anweisung ist die der
Anwendung (`+"`KIKartenanweisung`"+`), die Auftragssätze sind die des Formulars.
**Kartenkontext und Schema sind nachgebaut**: zwei Zeilen `+"`type`"+` und
`+"`rule text`"+` statt der Textfelder einer echten Vorlage, das Schema mit dem
einen Pflichtfeld. Die Zahlen gelten für diesen Testsatz in dieser Form —
nicht für jede Karte am Formular. Die erste Frage, wörtlich:

%s

| Ausgang | Zahl |
| --- | ---: |
| fertig, Sprache gehalten | %d |
| fertig, Sprache **nicht** gehalten | %d |
| leer | %d |
| abgelehnt | %d |
| Fehler | %d |

**Die letzte Spalte füllt, wer bewertet** — und der Beleg nennt ihn: passt der Vorschlag zur Karte?

| Karte | Sprache | Feld | Ausgang | Antwort | Dauer | passt? |
| ---: | --- | --- | --- | --- | ---: | --- |
%s
`, time.Now().Format("2006-01-02 15:04"), gesamt, eingerueckt(erste), zahlen["gut"], zahlen["sprache"], zahlen["leer"],
		zahlen["abgelehnt"], zahlen["fehler"], strings.Join(zeilen, "\n"))

	ziel := filepath.Join("..", "..", "docs", "belege", "fmguete.md")
	if fehler := os.WriteFile(ziel, []byte(text), 0o600); fehler != nil {
		t.Errorf("Beleg schreiben: %v", fehler)
	}

	t.Log(text)
}

// guetefrage stellt eine Frage und ordnet die Antwort ein.
func guetefrage(
	ctx context.Context, t *testing.T, klient *applefm.Klient, satz testsatzdatei, karte testkarte, sprache, feld string,
) (zeile, art string) {
	t.Helper()

	auftrag, schema := gueteauftrag(satz, karte, sprache, feld)

	anfang := time.Now()
	antwort, fehler := klient.Rede(ctx, applefm.Frage{System: anweisung, Auftrag: auftrag, Schema: schema})
	dauer := fmt.Sprintf("%.2f s", time.Since(anfang).Seconds())

	switch {
	case fehler != nil:
		return fmt.Sprintf("Fehler | %s | %s | |", strings.ReplaceAll(fehler.Error(), "|", "/"), dauer), "fehler"
	case antwort.Ausgang == applefm.AusgangAbgelehnt:
		return fmt.Sprintf("abgelehnt | %s | %s | |", antwort.Naeheres, dauer), "abgelehnt"
	}

	var felder map[string]string
	_ = json.Unmarshal([]byte(antwort.Inhalt), &felder) //nolint:errcheck // unlesbar heißt leer

	wert := strings.TrimSpace(felder[feld])
	if wert == "" {
		return fmt.Sprintf("leer | %q | %s | |", antwort.Inhalt, dauer), "leer"
	}

	art = "gut"
	if !haeltSprache(wert, sprache) {
		art = "sprache"
	}

	return fmt.Sprintf("%s | %s | %s | |", map[string]string{"gut": "fertig", "sprache": "fertig, Sprache?"}[art],
		strings.ReplaceAll(wert, "|", "/"), dauer), art
}

// haeltSprache ist eine grobe Probe: ein deutscher Text ohne ein deutsches
// Merkmal oder ein englischer mit Umlauten fällt auf. **Grob**, und die
// Spalte „passt?" ist die eigentliche Prüfung.
func haeltSprache(text, sprache string) bool {
	deutsch := strings.ContainsAny(text, "äöüÄÖÜß") ||
		strings.Contains(" "+strings.ToLower(text)+" ", " der ") ||
		strings.Contains(" "+strings.ToLower(text)+" ", " die ") ||
		strings.Contains(" "+strings.ToLower(text)+" ", " das ") ||
		strings.Contains(" "+strings.ToLower(text)+" ", " und ")

	if sprache == "en" {
		return !strings.ContainsAny(text, "äöüÄÖÜß")
	}

	// **Ein einzelnes Wort** (ein Name) trägt oft kein Merkmal — es zählt,
	// wenn es nicht englisch aussieht.
	return deutsch || utf8.RuneCountInString(text) < 25
}

// gueteauftrag baut Auftrag und Schema einer Frage — eine Stelle für die
// Fragen und für den Beleg.
func gueteauftrag(satz testsatzdatei, karte testkarte, sprache, feld string) (auftrag string, schema map[string]any) {
	auftrag = satz.Auftraege[sprache][feld] + "\n\n" + satz.Kartebisher[sprache] +
		"\ntype: " + karte.Typ + "\nrule text: " + karte.Text
	schema = map[string]any{
		"type": "object", "required": []string{feld},
		"properties": map[string]any{feld: map[string]any{"type": "string"}},
	}

	return auftrag, schema
}

// eingerueckt rückt einen Text um vier Leerzeichen ein — ein Codeblock.
func eingerueckt(text string) string {
	zeilen := strings.Split(text, "\n")
	for stelle, zeile := range zeilen {
		if zeile != "" {
			zeilen[stelle] = "    " + zeile
		}
	}

	return strings.Join(zeilen, "\n")
}
