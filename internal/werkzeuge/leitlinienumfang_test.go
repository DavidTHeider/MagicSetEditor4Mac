package werkzeuge_test

// **Die Regeldatei bleibt eine Regeldatei** — ENG-01.AK-04
// (#1726):
// „Detailkonzepte wandern in passende Dokumente statt beide Regeldateien
// weiter mit ganzen Projektplänen zu überladen."
//
// # Wozu dieser Wächter da ist
//
// **Weil eine Regeldatei wächst, ohne dass es jemand beschließt.** Jede
// Regel bekam hier ihren Anlass mit — „der Satz steht hier, weil …" —, und
// das ist richtig; aber am 2026-09-22 waren es 1.658 Zeilen, und rund 290
// davon — gezählt beim Umzug — waren Messungen und Anlässe, die in vier
// Dokumente daneben passen. Wer die Regeln lesen will, liest sie sonst nicht.
//
// **Die Obergrenze ist eine Entscheidung und keine Messung.** Wer sie hebt,
// begründet es im PR — dieselbe Form wie die 400 Zeilen aus §2. **Und sie
// ist knapp gewählt**: nach dem Umzug stand die Datei bei 1.399 Zeilen. Wer
// eine Regel ergänzt, entscheidet damit auch, welcher Anlass dafür umzieht.
//
// # Was er nicht hält
//
// **Ob eine Regel richtig ist, und ob ein Anlass in die Regeldatei gehört
// oder daneben.** Er zählt Zeilen und sieht nach, dass die Dokumente, auf
// die die Regeln verweisen, dastehen.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// leitlinienHoechstens ist die Obergrenze für AGENTS.md.
const leitlinienHoechstens = 1400

// **AGENTS.md bleibt unter der Obergrenze.**
func TestDieLeitlinienBleibenKurz(t *testing.T) {
	t.Parallel()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "AGENTS.md"))
	if fehler != nil {
		t.Fatalf("AGENTS.md lesen: %v", fehler)
	}

	if zeilen := strings.Count(string(roh), "\n"); zeilen > leitlinienHoechstens {
		t.Errorf("AGENTS.md hat %d Zeilen, höchstens %d — Anlässe und "+
			"Messungen gehören in ein Dokument daneben (ENG-01.AK-04)",
			zeilen, leitlinienHoechstens)
	}
}

// **Und was ausgelagert ist, ist verlinkt und da.**
//
// **Die Gegenprobe zur Zeile darüber** (§10): eine Regeldatei wird auch
// dadurch kürzer, dass man Text löscht — und dann fehlt der Anlass, den der
// Umzug bewahren sollte.
func TestDasAusgelagerteIstVerlinkt(t *testing.T) {
	t.Parallel()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "AGENTS.md"))
	if fehler != nil {
		t.Fatalf("AGENTS.md lesen: %v", fehler)
	}

	for _, dokument := range []string{
		"docs/TESTGRENZEN.md", "docs/SICHTPRUEFUNG.md", "docs/PIPELINE.md",
	} {
		if !strings.Contains(string(roh), "("+dokument+")") {
			t.Errorf("AGENTS.md verweist nicht auf %s", dokument)
		}

		if _, fehler := os.Stat(filepath.Join("..", "..", dokument)); fehler != nil {
			t.Errorf("%s fehlt: %v", dokument, fehler)
		}
	}
}
