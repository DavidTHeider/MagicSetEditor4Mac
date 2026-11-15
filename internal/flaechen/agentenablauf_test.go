// **Der Ablauf, den ein Agent geht** — Werkstück A14 von Phase 5.3
// (#1453),
// Ergänzung vom 2026-09-10.
//
// # Der Satz, für den es diesen Lauf gibt
//
// > Konkrete Agentenabläufe: lokale semantische Suche **ohne**
// > Generierungsaufruf; anschließend auf ausdrücklichen Auftrag
// > quellengebundenen Vorschlag erzeugen und gesondert übernehmen.
//
// **Die beiden Hälften sind zwei Zusagen.** Dass eine Suche nichts erzeugt,
// ist an einem Zähler zu messen; dass die Erzeugung einen **ausdrücklichen**
// Ruf braucht, daran, dass sie ein zweiter ist. Beides in **einer** Sitzung,
// denn dazwischen darf nichts liegen, was der nächste Ruf nicht sähe.
//
// # Und warum er erst seit A13.2a geht
//
// **Weil es den quellengebundenen Vorschlag vorher nicht gab.** A14 hängt an
// dieser Stelle hinter A13; das steht im Issue als Vorgänger und war der
// Grund, A13.2a vorzuziehen.
package flaechen_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Suchen erzeugt nichts. Erzeugen ist ein zweiter Ruf. Übernehmen ein
// dritter.**
func TestEinAgentSuchtErstUndErzeugtDannAufAuftrag(t *testing.T) {
	t.Parallel()

	umgebung, laufzeit := einAgentenbestand(t)
	setID, karteID := eineAbnahmekarte(t, umgebung)

	agent := starteMCPProzess(t, umgebung)

	// **1. Die Suche.** Sie antwortet — ohne Index mit einem Grund, und das
	// ist eine Antwort.
	if antwort := agent.muss("lesen.recherche", map[string]any{
		"fall": "regeln", "frage": "Blocken",
	}); !strings.Contains(antwort, "leere") {
		t.Fatalf("die Recherche nennt keinen Grund: %s", antwort)
	}

	// **Und sie hat kein Modell gefragt.** Das ist die erste Hälfte des
	// Satzes, und sie hängt an diesem Zähler: eine Absage sagte nur, dass
	// nichts gefunden wurde.
	if gefragt := laufzeit.auftraege(); gefragt != 0 {
		t.Errorf("die Suche fragte %d Modelle", gefragt)
	}

	// **2. Der ausdrückliche Auftrag** — jetzt geht einer hinaus.
	agent.muss("schreiben.ki.vorschlag", map[string]any{
		"karteId": karteID, "anbieter": "ollama", "alias": probemodell,
		"auftrag": derAuftrag,
	})

	if gefragt := laufzeit.auftraege(); gefragt != 1 {
		t.Errorf("nach dem Auftrag sind es %d Rufe", gefragt)
	}

	// **Und die Karte ist unberührt** — erzeugen ist nicht übernehmen.
	vorher := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID)
	if vorher == "Aus dem Modell" {
		t.Fatalf("der Vorschlag hat geschrieben")
	}

	// **3. Die Übernahme ist ein dritter Ruf.**
	agent.muss("schreiben.ki.vorschlag", map[string]any{
		"karteId": karteID, "anbieter": "ollama", "alias": probemodell,
		"auftrag": derAuftrag, "uebernehmen": true, "felder": []string{dasFeld},
	})

	if jetzt := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID); jetzt != "Aus dem Modell" {
		t.Errorf("die Übernahme kam nicht an: %q", jetzt)
	}
}

// **Und der Weg über die Einschätzung ist derselbe** — nur ohne Übernahme.
//
// **Das ist der Punkt**: eine Einschätzung hat keinen dritten Ruf, weil es
// ihn nicht gibt. Sie kann nicht in eine Karte wandern, auch nicht
// versehentlich.
func TestEineEinschaetzungHatKeineUebernahme(t *testing.T) {
	t.Parallel()

	umgebung, _ := einAgentenbestand(t)
	agent := starteMCPProzess(t, umgebung)

	for _, name := range werkzeugnamen(t, agent) {
		if strings.Contains(name, "einschaetzung") &&
			name != "schreiben.ki.einschaetzung" {
			t.Errorf("es gibt ein zweites Werkzeug %q", name)
		}
	}
}

// **Und was ein Belegtext sagt, führt niemand aus.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil Belegtexte aus einem Korpus kommen, den jemand anders geschrieben
// hat.** Das Issue nennt es eigens: „MCP-Antworten/Belegtexte enthalten keine
// auszuführenden Instruktionen." Ein Regeltext, in dem „ignoriere deine
// bisherigen Anweisungen" steht, reist sonst durch die Antwort in die
// nächste Runde eines Sprachmodells.
//
// **Was dieser Lauf hält und was nicht**: er hält, dass die Antwort den Text
// als **Feld** führt und nicht als Anweisung — Belege stehen unter
// `belege[].ausschnitt`, und kein Werkzeug führt aus, was dort steht. Was er
// **nicht** halten kann, ist, dass ein Sprachmodell den Text nicht doch als
// Anweisung liest; das entscheidet der Agent, und dagegen hilft kein Wächter
// in diesem Programm.
func TestEinBelegtextIstEinFeldUndKeineAnweisung(t *testing.T) {
	t.Parallel()

	umgebung, _ := einAgentenbestand(t)
	agent := starteMCPProzess(t, umgebung)

	// **Die Recherche gibt Treffer als Felder** — der Text steht unter
	// `ausschnitt`, nicht als freier Satz in der Antwort.
	roh := agent.muss("lesen.recherche", map[string]any{
		"fall": "regeln", "frage": "Blocken",
	})

	var gelesen map[string]json.RawMessage
	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("die Antwort ist kein Verbund: %v — %s", fehler, roh)
	}

	if _, hat := gelesen["treffer"]; !hat {
		t.Errorf("die Antwort führt keine Trefferliste: %s", roh)
	}

	// **Und die Werkzeugliste kennt nichts, was einen Text ausführte.**
	for _, name := range werkzeugnamen(t, agent) {
		for _, verboten := range []string{"ausfuehren.text", "eval", "shell"} {
			if strings.Contains(name, verboten) {
				t.Errorf("es gibt ein Werkzeug %q", name)
			}
		}
	}
}

// werkzeugnamen liest die Namen aus `tools/list`.
func werkzeugnamen(t *testing.T, agent *mcpprozess) []string {
	t.Helper()

	var antwort struct {
		Ergebnis struct {
			Werkzeuge []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}

	if fehler := json.Unmarshal([]byte(agent.werkzeugliste()), &antwort); fehler != nil {
		t.Fatalf("die Werkzeugliste lesen: %v", fehler)
	}

	if len(antwort.Ergebnis.Werkzeuge) == 0 {
		t.Fatalf("die Werkzeugliste ist leer — dann sagt der Lauf nichts")
	}

	namen := make([]string, 0, len(antwort.Ergebnis.Werkzeuge))
	for _, werkzeug := range antwort.Ergebnis.Werkzeuge {
		namen = append(namen, werkzeug.Name)
	}

	return namen
}

// **Und die Art einer Einschätzung ist eine Wahl aus zweien.**
func TestDieEinschaetzungskenntNurZweiArten(t *testing.T) {
	t.Parallel()

	if len(domain.Einschaetzungsarten()) != 2 {
		t.Errorf("es sind %d Arten", len(domain.Einschaetzungsarten()))
	}
}
