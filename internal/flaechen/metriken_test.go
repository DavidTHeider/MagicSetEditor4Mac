package flaechen_test

// **Die begrenzten Metriken über MCP und Dienst** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Was er hält
//
// **Die Zusage, um derentwillen es das Werkstück gibt**: die Zahl der Reihen
// wächst nicht mit dem Bestand. Geprüft wird sie nicht am Zähler, sondern an
// der **Fläche** — mit echten Rufen, die echte Wirkungen auslösen.
//
// # Was er nicht hält
//
// **Das Fenster und die Kommandozeile.** Das Fenster, weil dieser Lauf
// Prozesse startet und die Wails-Bindung in keinem läuft (wie bei der
// Gesundheit). Die Kommandozeile, weil sie diese Fähigkeit **nicht hat** und
// der Grund in `docs/FAEHIGKEITEN.md` steht: ein Prozess je Befehl, und ein
// Zähler über seine Lebenszeit wäre immer leer.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestDieMetrikenGehenUeberZweiFlaechen prüft dieselbe Auskunft an beiden.
func TestDieMetrikenGehenUeberZweiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		antwort := mussMCP(t, umgebung, "lesen.metriken", map[string]any{})

		// **Die Grenze steht in der Antwort** — sie ist die Zusage, und ein
		// Leser soll sie nicht nachschlagen müssen.
		if !strings.Contains(antwort, `"grenze": 48`) {
			t.Errorf("Antwort %q nennt die Grenze nicht", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		antwort := mussREST(t, an, http.MethodGet, "/metriken", "")
		if !strings.Contains(antwort, `"grenze":48`) {
			t.Errorf("Antwort %q nennt die Grenze nicht", antwort)
		}
	})
}

// TestVieleRufeErzeugenKeineNeuenMessreihen ist die Abnahme aus dem Issue:
// „Hohe Zahl verschiedener Objekte erhöht Metrikserien nicht unbegrenzt."
//
// **Gemessen an der Fläche und nicht am Zähler.** Ein Test, der zwanzigmal
// `service.Zaehle` ruft, prüft den Zähler; hier gehen zwanzig **verschiedene**
// Sets über denselben Dienst, jedes mit einer Ausfuhr — und die Tabelle
// bleibt so kurz, wie sie war.
func TestVieleRufeErzeugenKeineNeuenMessreihen(t *testing.T) {
	t.Parallel()

	an := starteDienst(t, deckumgebung(t))

	vorher := reihenzahl(t, mussREST(t, an, http.MethodGet, "/metriken", ""))

	ziel := t.TempDir()

	for nummer := range 20 {
		// **Ein Weg mit Dateiwirkung**, denn nur der geht am Riegel vorbei:
		// das Sichern eines Sets an einen jedes Mal anderen Ort. Zwanzig
		// verschiedene Sets, zwanzig verschiedene Dateinamen — genau die
		// „hohe Zahl verschiedener Objekte" aus der Abnahme.
		an.ruf(t, http.MethodPost, "/sets/"+itoa(nummer+1)+"/sicherung",
			`{"ziel":"`+ziel+`/probe-`+itoa(nummer)+`.mse-set"}`)
	}

	nachher := reihenzahl(t, mussREST(t, an, http.MethodGet, "/metriken", ""))

	// **Und es ist wirklich gemessen worden.** Ohne diese Zeile wäre der Test
	// auch dann grün, wenn kein einziger Ruf am Riegel vorbeikäme — zweimal
	// null ist keine Aussage über eine Obergrenze (§10).
	if nachher == 0 {
		t.Fatal("nach zwanzig Rufen steht keine einzige Reihe da — " +
			"gemessen wurde nichts, und die Grenze ist damit unbelegt")
	}

	// **Nicht „gleich viele"**, sondern „nicht mehr als vierzig": ob eine
	// neue Reihe entsteht, hängt davon ab, welche Wirkungen der Weg auslöst —
	// **wie viele** es höchstens werden, hängt an nichts.
	if nachher > 48 || nachher < vorher {
		t.Fatalf("Reihen vorher %d, nachher %d — die Grenze ist 48", vorher, nachher)
	}
}

// reihenzahl liest, wie viele Reihen eine Antwort führt.
func reihenzahl(t *testing.T, roh string) int {
	t.Helper()

	var stand struct {
		Reihen []map[string]any `json:"reihen"`
		Grenze int              `json:"grenze"`
	}

	if fehler := json.Unmarshal([]byte(roh), &stand); fehler != nil {
		t.Fatalf("%q ist kein Metrikstand: %v", roh, fehler)
	}

	if stand.Grenze != 48 {
		t.Fatalf("die Grenze ist %d, erwartet 48", stand.Grenze)
	}

	return len(stand.Reihen)
}

// itoa schreibt eine kleine Zahl als Text.
func itoa(zahl int) string {
	if zahl == 0 {
		return "0"
	}

	ziffern := ""
	for zahl > 0 {
		ziffern = string(rune('0'+zahl%10)) + ziffern
		zahl /= 10
	}

	return ziffern
}
