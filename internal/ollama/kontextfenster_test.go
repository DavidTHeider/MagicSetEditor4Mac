package ollama_test

// **Das Kontextfenster eines Modells** — AI-01.AK-02
// (#1742):
// „Kontext-/Ressourcenlimits … sind nachvollziehbar". Ollama nennt zwei
// Zahlen, und sie sind nicht dieselbe: die **trainierte** Länge unter
// `model_info` (`<architektur>.context_length`) und das, was die Laufzeit
// **liest**, `num_ctx` unter den Parametern. Gilt, was gelesen wird.
//
// Die Gestalt ist am 2026-10-04 in Ollamas `docs/api.md` nachgelesen (über
// Context7); die Beispiele sind selbst geschrieben (§3).

import (
	"strings"
	"testing"
)

func TestDasKontextfensterIstWasDieLaufzeitLiest(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	// **Mit `num_ctx`**: es gilt, und die trainierte Länge steht im Beleg.
	gemma := modelle[0]
	if gemma.Kontextfenster != 2048 || !strings.Contains(gemma.KontextBeleg, "num_ctx") ||
		!strings.Contains(gemma.KontextBeleg, "131072") {
		t.Errorf("gemma: %d, %q", gemma.Kontextfenster, gemma.KontextBeleg)
	}

	// **Ohne `num_ctx`**: die trainierte Länge, und der Beleg sagt, dass die
	// Laufzeit weniger lesen kann.
	einbetter := modelle[1]
	if einbetter.Kontextfenster != 2048 || !strings.Contains(einbetter.KontextBeleg, "trainiert") {
		t.Errorf("einbetter: %d, %q", einbetter.Kontextfenster, einbetter.KontextBeleg)
	}
}
