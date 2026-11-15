// **Dieselbe Abweisung, nicht irgendeine** —
// #1169,
// Befund 3 des Zweitblicks.
//
// # Der Befund
//
// §16 verlangt „dieselben fachlichen **Fehler**". Neun neue Läufe prüften
// stattdessen nur, dass jede Fläche **irgendwie** scheitert: ein
// Validierungsfehler auf der Kommandozeile, ein `not found` über MCP und ein
// 500 vom Dienst hätten gemeinsam bestanden.
//
// **Und die Läufe davor auch.** Das ist keine Eigenheit der neuen — es ist die
// Klasse, und sie war seit Phase T da. Was fehlte, war ein gemeinsamer Ort,
// an dem „dieselbe Abweisung" einmal definiert ist.
//
// # Was eine Fläche über ihre Abweisung hergibt
//
// | Fläche | Woran die Klasse zu erkennen ist |
// | --- | --- |
// | Kommandozeile | der Rückgabewert: 1 fachlich, 2 Bedienung |
// | MCP | `isError` und der Text — **einen Code gibt es dort nicht** |
// | Dienst | der Status **und** das Feld `code` im Rumpf |
//
// **Zwei von drei tragen die Klasse, die dritte nur den Text.** Deshalb prüft
// [abweisungGleich] bei MCP, dass abgewiesen wurde und der Text denselben
// **Kern** nennt wie die anderen beiden — mehr gibt die Fläche nicht her, und
// das gehört gesagt statt gedeckt.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Die Klassen, die eine Abweisung tragen kann — die Codes von `internal/rest`.
const (
	klasseValidierung   = "validierung"
	klasseNichtGefunden = "nichtGefunden"
	klasseKonflikt      = "konflikt"

	// klasseAbgeschaltet ist ein abgeschalteter Bereich — OPS-10 (#1908).
	klasseAbgeschaltet = "abgeschaltet"

	// klasseErlaubnis ist eine Wirkung, die der Akteur nicht auslösen darf —
	// etwa ein Index, solange die KI-Laufzeit aus ist (C04, #2229).
	klasseErlaubnis = "erlaubnis"
)

// statusFuer gibt den HTTP-Status, den eine Klasse am Dienst hat.
var statusFuer = map[string]int{
	klasseValidierung:   http.StatusBadRequest,
	klasseNichtGefunden: http.StatusNotFound,
	klasseKonflikt:      http.StatusConflict,
	klasseAbgeschaltet:  http.StatusForbidden,
	klasseErlaubnis:     http.StatusForbidden,
}

// abweisung ist ein Fall, der an allen drei Flächen scheitern muss.
type abweisung struct {
	// klasse ist die erwartete fachliche Klasse.
	klasse string

	// kern ist ein Stück Meldungstext, das jede Fläche nennen muss.
	//
	// **Leer heißt: keiner.** Manche Abweisungen tragen eine Meldung, die je
	// Fläche anders formuliert ist — dann bleibt die Klasse die Zusage, und
	// das ist ehrlicher als ein Kern, den man an den Text anpasst, bis er
	// passt.
	//
	// **Pflicht seit dem 2026-08-28.** Er durfte leer sein und wurde dann
	// übersprungen — und **kein einziger** der acht Fälle setzte ihn. Damit
	// prüften Kommandozeile und MCP nur „irgendein Misserfolg", und genau die
	// Gegenprobe, die der Helfer schließen sollte, stand weiter offen: ein
	// Validierungsfehler auf der einen Fläche und ein `nicht gefunden` auf der
	// anderen konnten gemeinsam grün sein. Angestrichen vom Nachlauf über
	// Phase 3.8
	// (#1169,
	// Befund 2) — es ist die Gestalt aus §10: die Korrektur behob den
	// genannten Fall und ließ die Klasse stehen.
	kern string

	// Die drei Aufrufe. Jeder gibt, was seine Fläche über den Ausgang hergibt.
	cli  func() (aus string, code int)
	mcp  func() (text string, istFehler bool)
	rest func() (status int, rumpf string)

	// fenster ist der vierte Aufruf — die Bindung, die das Fenster ruft, mit
	// dem Fehler ihrer Antwort (C04, #2229). **Oder** ohneFenster nennt, warum
	// es ihn nicht gibt; beides leer macht den Fall rot.
	fenster     func() *app.Fehler
	ohneFenster string

	// ohneKommandozeile nennt, warum die Kommandozeile fehlt — etwa, weil
	// sie die Fassung selbst liest und keine veraltete annimmt (C04).
	ohneKommandozeile string

	// ohneMCP nennt, warum MCP fehlt — den Grund aus `FAEHIGKEITEN.md`.
	ohneMCP string
}

// abweisungGleich verlangt, dass alle drei Flächen **gleich** abweisen.
func abweisungGleich(t *testing.T, was string, fall abweisung) {
	t.Helper()

	aussen := t

	t.Run(was+" — Kommandozeile", func(t *testing.T) {
		if fall.cli == nil {
			if strings.TrimSpace(fall.ohneKommandozeile) == "" {
				t.Fatal("der Fall fährt die Kommandozeile nicht und sagt nicht, warum")
			}

			t.Skipf("ohne Kommandozeile: %s", fall.ohneKommandozeile)
		}

		aus, code := fall.cli()

		if code == 0 {
			t.Fatalf("die Kommandozeile ließ es durch: %s", aus)
		}

		// **Ein fachlicher Fehler und keine Bedienungsklage.** Der
		// Unterschied ist die Aussage: 2 heißt „so ruft man das nicht auf",
		// 1 heißt „der Aufruf war richtig und die Sache geht nicht".
		if code != 1 {
			t.Errorf("der Rückgabewert ist %d — erwartet war 1 (fachlich):\n%s", code, aus)
		}

		pruefeKern(t, "die Kommandozeile", fall, aus)

		// **Und mit `--json` maschinenlesbar** — F03 (#2342): derselbe Aufruf
		// noch einmal; eine Abweisung schreibt nichts.
		if json := mitCode(aussen, fall.cli); !strings.Contains(json, `"code":"`+fall.klasse+`"`) {
			t.Errorf("mit --json nennt die Kommandozeile die Klasse %q nicht:\n%s", fall.klasse, json)
		}
	})

	t.Run(was+" — MCP", func(t *testing.T) {
		if fall.mcp == nil {
			if strings.TrimSpace(fall.ohneMCP) == "" {
				t.Fatal("der Fall fährt MCP nicht und sagt nicht, warum")
			}

			t.Skipf("ohne MCP: %s", fall.ohneMCP)
		}

		text, istFehler := fall.mcp()

		if !istFehler {
			t.Fatalf("MCP ließ es durch: %s", text)
		}

		pruefeKern(t, "MCP", fall, text)

		// **Und strukturiert** — F03 (#2342), in `structuredContent.fehler`.
		strukturiert := mitCode(aussen, func() (string, int) { text, _ := fall.mcp(); return text, 0 })
		if !strings.Contains(strukturiert, `"code":"`+fall.klasse+`"`) {
			t.Errorf("MCP nennt die Klasse %q nicht strukturiert:\n%s", fall.klasse, strukturiert)
		}
	})

	t.Run(was+" — Fenster", func(t *testing.T) {
		switch {
		case fall.fenster == nil && strings.TrimSpace(fall.ohneFenster) == "":
			t.Fatal("der Fall fährt das Fenster nicht und sagt nicht, warum (C04)")
		case fall.fenster == nil:
			t.Skipf("ohne Fenster: %s", fall.ohneFenster)
		}

		fehler := wieWails(fall.fenster)
		if fehler == nil {
			t.Fatal("das Fenster ließ es durch")
		}

		if string(fehler.Code) != fall.klasse {
			t.Errorf("der Code ist %q — erwartet war %q: %s", fehler.Code, fall.klasse, fehler.Meldung)
		}

		pruefeKern(t, "das Fenster", fall, fehler.Meldung)
	})

	t.Run(was+" — Dienst", func(t *testing.T) {
		status, rumpf := fall.rest()

		erwartet, gefuehrt := statusFuer[fall.klasse]
		if !gefuehrt {
			t.Fatalf("die Klasse %q ist nicht geführt", fall.klasse)
		}

		if status != erwartet {
			t.Errorf("der Status ist %d — erwartet war %d für %q:\n%s",
				status, erwartet, fall.klasse, rumpf)
		}

		// **Der Code im Rumpf und nicht nur der Status.** Ein 400 sagt „so
		// nicht"; `"code":"validierung"` sagt, **welche** Regel griff — und
		// das ist die Angabe, nach der ein Aufrufer verzweigt.
		if !strings.Contains(rumpf, `"code":"`+fall.klasse+`"`) {
			t.Errorf("der Rumpf nennt die Klasse %q nicht:\n%s", fall.klasse, rumpf)
		}

		pruefeKern(t, "der Dienst", fall, rumpf)
	})
}

// klassenvorsatz ist der Satzanfang, den eine Fehlerklasse im Klartext trägt.
//
// **Nur für die, die einen haben.** `domain.ErrNotFound` und
// `domain.ErrConflict` schreiben ihn; `ErrValidation` nicht — dort steht der
// fachliche Satz allein. Eine erfundene Zeile für die dritte wäre eine Zusage
// über Text, den es nicht gibt.
var klassenvorsatz = map[string]string{
	klasseNichtGefunden: "nicht gefunden:",
	klasseKonflikt:      "Konflikt mit dem gegenwärtigen Stand:",
	klasseAbgeschaltet:  "dieser Bereich ist abgeschaltet:",
	klasseErlaubnis:     "dieser Nachweis reicht nicht so weit:",
}

// pruefeKern verlangt, dass eine Meldung den gemeinsamen Kern nennt — und den
// Vorsatz ihrer Klasse, sofern die einen hat.
//
// **Der Vorsatz ist die eigentliche Absicherung gegen verwechselte Klassen.**
// Der Kern sagt, dass es um dieselbe Sache geht; der Vorsatz sagt, dass alle
// drei Flächen sie derselben Klasse zuordnen. Ohne ihn konnte die
// Kommandozeile einen Validierungsfehler melden, während der Dienst einen
// Konflikt gab — und beide Läufe waren grün.
func pruefeKern(t *testing.T, wer string, fall abweisung, text string) {
	t.Helper()

	if fall.kern == "" {
		t.Fatalf("%s: dieser Fall nennt keinen Kern — ohne ihn prüft dieser "+
			"Lauf nur, dass irgendetwas schiefging", wer)
	}

	if !strings.Contains(strings.ToLower(text), strings.ToLower(fall.kern)) {
		t.Errorf("%s nennt %q nicht:\n%s", wer, fall.kern, text)
	}

	if vorsatz, hatEinen := klassenvorsatz[fall.klasse]; hatEinen &&
		!strings.Contains(text, vorsatz) {
		t.Errorf("%s trägt den Vorsatz %q der Klasse %q nicht:\n%s",
			wer, vorsatz, fall.klasse, text)
	}
}
