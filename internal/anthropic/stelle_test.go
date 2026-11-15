package anthropic_test

// **Zwei namenlose Einträge geben zwei unterscheidbare Befunde** — Werkstück
// A19 von Phase 5.3
// (#1550).
//
// # Der Fund
//
// **Gefunden vom Fuzzer bei der Saat `{"data":[{}]}`.** Ein Eintrag ohne
// `id` fiel schon vorher aus dem Katalog — richtig —, aber sein Befund nannte
// ihn `""`. Zwei solche Einträge gaben zwei **wortgleiche** Sätze, und §8
// verlangt für jeden Fehler die Stelle: „Datei, Zeile, Feld".
//
// **Das ist der Unterschied zwischen einer Meldung und einer Auskunft**: wer
// zwei gleiche Sätze liest, weiß nicht, ob es ein Eintrag war oder zwei.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// **Zwei Einträge ohne Kennung sind auseinanderzuhalten.**
func TestZweiNamenloseEintraegeGebenZweiStellen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write(
				[]byte(`{"data":[{},{}]}`)); fehler != nil {
				t.Errorf("antworten: %v", fehler)
			}
		}))

	defer server.Close()

	modelle, befunde, fehler := klientFuer(server).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	// **Keiner der beiden kommt in den Katalog** — das war schon vorher so,
	// und es ist die Hälfte, die stimmte.
	if len(modelle) != 0 {
		t.Errorf("ein Eintrag ohne Auskunft kam in den Katalog: %+v", modelle)
	}

	if len(befunde) != 2 {
		t.Fatalf("es gab %d Befunde, erwartet waren zwei", len(befunde))
	}

	// **Und sie sind unterscheidbar.** Ohne die Stelle wären beide Sätze
	// gleich, und der Lauf wäre grün, ohne etwas auszusagen.
	if befunde[0].Ort == befunde[1].Ort {
		t.Errorf("beide Befunde nennen dieselbe Stelle %q", befunde[0].Ort)
	}

	for stelle, befund := range befunde {
		if !strings.Contains(befund.Ort, "data[") {
			t.Errorf("der %d. Befund nennt keine Stelle: %+v", stelle, befund)
		}
	}
}
