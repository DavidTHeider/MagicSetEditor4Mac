package flaechen_test

// **Die bekannten Gegenüber über jede Fläche** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass dieselbe Vertrauensentscheidung überall dieselbe ist** (§16). Ein
// Gegenüber, das an einer Fläche bestätigt ist und an der anderen nicht,
// wäre die Gestalt, in der jemand über den Dienst freigibt, was er am
// Fenster abgelehnt hat.
//
// **Und dass das Nein überall dasselbe ist.** Ein falscher Abdruck, ein Recht
// an einem unbestätigten Gegenüber, ein Zweck, den es nicht gibt — jedes davon
// endet, und zwar an jeder Fläche.

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

func TestDieKontakteUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())
	abdruck := fremd.Fingerabdruck()

	// **Vorher kennt niemand jemanden** — an allen drei Flächen.
	//
	// **Einzeln und nicht über den Dreiervergleich**: die Kommandozeile
	// schreibt bei einer leeren Liste **nichts**, und ein gemeinsamer Kern
	// wäre die leere Zeichenkette — die steht in jeder Ausgabe.
	if aus := mussKommandozeile(t, umgebung, "kontakte"); strings.TrimSpace(
		aus) != "" {
		t.Errorf("die Kommandozeile kennt schon jemanden: %q", aus)
	}

	for name, leer := range map[string]string{
		"MCP":    mussMCP(t, umgebung, "lesen.peerkontakte", map[string]any{}),
		"Dienst": mussREST(t, an, http.MethodGet, "/peer/kontakte", ""),
	} {
		if !strings.Contains(strings.ReplaceAll(leer, " ", ""), `"kontakte":[]`) {
			t.Errorf("%s kennt schon jemanden: %s", name, leer)
		}
	}

	// **Eingetragen wird an der Kommandozeile** — und danach sehen ihn alle
	// drei, unbestätigt.
	aufgenommen := mussKommandozeile(
		t, umgebung, "kontakt-aufnehmen", kennung, "Annas Rechner")
	if !strings.Contains(aufgenommen, "unbestaetigt") {
		t.Fatalf("das Aufnehmen gab %q", aufgenommen)
	}

	fuerAlleDreiMit(t, "derselbe Kontakt",
		[3]string{kennung, kennung, kennung},
		func() string { return mussKommandozeile(t, umgebung, "kontakte") },
		func() string { return mussMCP(t, umgebung, "lesen.peerkontakte", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/peer/kontakte", "") })

	// **Ein Recht an einem unbestätigten Gegenüber wird abgewiesen** — und
	// zwar an beiden schreibenden Flächen.
	//
	// **Das ist der Satz des Issues**: „Pairing allein gestattet keinen
	// Import oder Handel."
	if _, _, code := kommandozeile(t, umgebung,
		"zweck-erlauben", kennung, "handel"); code == 0 {
		t.Error("die Kommandozeile gab ein Recht ohne Bestätigung")
	}

	if status, _ := an.ruf(t, http.MethodPut,
		"/peer/kontakte/"+kennung+"/zwecke/handel", ""); status < 400 {
		t.Errorf("der Dienst gab ein Recht ohne Bestätigung: %d", status)
	}

	// **Ein falscher Abdruck bestätigt nichts** — an beiden.
	if _, _, code := kommandozeile(t, umgebung,
		"kontakt-bestaetigen", kennung, "0000-1111"); code == 0 {
		t.Error("die Kommandozeile nahm einen falschen Abdruck an")
	}

	if status, _ := an.ruf(t, http.MethodPost,
		"/peer/kontakte/"+kennung+"/bestaetigung",
		`{"fingerabdruck":"0000-1111"}`); status < 400 {
		t.Errorf("der Dienst nahm einen falschen Abdruck an: %d", status)
	}

	// **Bestätigt wird über den Dienst** — und die Kommandozeile sieht es.
	mussREST(t, an, http.MethodPost, "/peer/kontakte/"+kennung+"/bestaetigung",
		`{"fingerabdruck":"`+abdruck+`"}`)

	if aus := mussKommandozeile(t, umgebung, "kontakte"); !strings.Contains(
		aus, "bestaetigt") {
		t.Fatalf("nach dem Bestätigen steht da %q", aus)
	}

	// **Und jetzt greift die Freigabe** — gegeben an der Kommandozeile,
	// gesehen im Dienst und über MCP.
	mussKommandozeile(t, umgebung, "zweck-erlauben", kennung, "handel")

	// **An MCP und Dienst die Rechte des Kontakts, nicht irgendein
	// `"handel"`** — der Name steht auch in der Aufzählung der möglichen
	// Zwecke, die in derselben Antwort reist (Zweitblick über Phase C,
	// Befund 3; dieselbe Falle, die weiter unten schon einmal zuschnappte).
	if aus := mussKommandozeile(t, umgebung, "kontakte"); !strings.Contains(aus, "darf\thandel") {
		t.Errorf("Kommandozeile: das freigegebene Recht fehlt: %q", aus)
	}

	for flaeche, roh := range map[string]string{
		"MCP":    mussMCP(t, umgebung, "lesen.peerkontakte", map[string]any{}),
		"Dienst": mussREST(t, an, http.MethodGet, "/peer/kontakte", ""),
	} {
		if rechte := rechteVon(t, roh, kennung); !slices.Contains(rechte, "handel") {
			t.Errorf("%s: der Kontakt hat die Rechte %v — „handel“ fehlt", flaeche, rechte)
		}
	}

	// **Und ein Entzug nimmt es zurück** — über den Dienst gegeben, an der
	// Kommandozeile gesehen. Danach wird es wieder gegeben, damit die Sperre
	// darunter etwas zu räumen hat.
	mussREST(t, an, http.MethodDelete,
		"/peer/kontakte/"+kennung+"/zwecke/handel", "")

	if aus := mussKommandozeile(t, umgebung, "kontakte"); strings.Contains(
		aus, "darf\thandel") {
		t.Fatalf("nach dem Entzug steht es noch da: %q", aus)
	}

	// **Und beide Richtungen an beiden schreibenden Flächen**: gegeben über
	// den Dienst, genommen an der Kommandozeile. Eine Freigabe, die nur über
	// eine Fläche zurückzunehmen wäre, wäre an der anderen eine Einbahn.
	mussREST(t, an, http.MethodPut,
		"/peer/kontakte/"+kennung+"/zwecke/handel", "")
	mussKommandozeile(t, umgebung, "zweck-entziehen", kennung, "handel")

	// **Gegen `"rechte":[…]` und nicht gegen `"handel"`** — der Name steht
	// auch in der Aufzählung der möglichen Zwecke, die in derselben Antwort
	// reist. Der erste Anlauf prüfte genau das und war deshalb rot, ohne dass
	// etwas falsch war (§10).
	if aus := mussREST(t, an, http.MethodGet, "/peer/kontakte",
		""); !strings.Contains(aus, `"rechte":[]`) {
		t.Fatalf("nach dem zweiten Entzug steht es noch da: %s", aus)
	}

	mussKommandozeile(t, umgebung, "zweck-erlauben", kennung, "handel")

	// **Eine Sperre nimmt es mit** — gesetzt über MCP, gesehen an den anderen
	// beiden. Das ist die einzige Vertrauensänderung mit Werkzeug, und
	// deshalb wird gerade sie hier gefahren.
	mussMCP(t, umgebung, "schreiben.peerkontakt.sperren",
		map[string]any{"kennung": kennung})

	nachSperre := mussKommandozeile(t, umgebung, "kontakte")
	if !strings.Contains(nachSperre, "gesperrt") ||
		strings.Contains(nachSperre, "darf\thandel") {
		t.Fatalf("nach der Sperre steht da %q", nachSperre)
	}

	// **Und das Entsperren gibt sie nicht zurück** — der Stand ist
	// `unbestaetigt`, nicht `bestaetigt`.
	mussKommandozeile(t, umgebung, "kontakt-entsperren", kennung)

	nachEntsperren := mussKommandozeile(t, umgebung, "kontakte")
	if !strings.Contains(nachEntsperren, "unbestaetigt") ||
		strings.Contains(nachEntsperren, "darf\thandel") {
		t.Fatalf("nach dem Entsperren steht da %q", nachEntsperren)
	}
}

// rechteVon liest die freigegebenen Zwecke eines Kontakts — aus der Gestalt,
// die MCP und Dienst teilen (`{"kontakte":[{"kennung","rechte"}]}`).
func rechteVon(t *testing.T, roh, kennung string) []string {
	t.Helper()

	var antwort struct {
		Kontakte []struct {
			Kennung string   `json:"kennung"`
			Rechte  []string `json:"rechte"`
		} `json:"kontakte"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("kein JSON: %v\n%s", fehler, roh)
	}

	for _, kontakt := range antwort.Kontakte {
		if kontakt.Kennung == kennung {
			return kontakt.Rechte
		}
	}

	t.Fatalf("der Kontakt %s steht nicht in der Antwort:\n%s", kennung, roh)

	return nil
}
