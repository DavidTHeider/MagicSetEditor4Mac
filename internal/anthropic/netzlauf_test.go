//go:build netzlauf

// **Anthropic (Cloudmodelle) am Original** — #2049, Entscheidung 1 in `docs/INTEGRATIONEN.md`:
// bis hierher nur gegen selbst geschriebene Antworten geprüft.
//
// **Mit dem Zugang des Auftraggebers**, aus `MSE_NETZLAUF_ANTHROPIC`. Fehlt er, endet
// der Lauf **rot mit BLOCKED** — nie grün, denn ein übersprungener Lauf
// belegt nichts. **Nur die Modellliste**: sie kostet nichts und belegt doch Schlüssel, Kopfzeilen und Antwortgestalt. Ein Lauf, der rechnen lässt, kostet Guthaben und bleibt einer eigenen Freigabe vorbehalten.
//
//	MSE_NETZLAUF_… go test -count=1 -tags netzlauf -run AmOriginal -v ./internal/anthropic/
//
// **Nicht in der Pipeline**: er braucht einen Zugang, den kein zweiter
// Rechner haben muss, und er geht ins Netz (§16).
package anthropic_test

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestAnthropicAmOriginal(t *testing.T) {
	schluessel := os.Getenv("MSE_NETZLAUF_ANTHROPIC")
	if schluessel == "" {
		t.Fatal("BLOCKED: MSE_NETZLAUF_ANTHROPIC fehlt — ohne Schlüssel nennt Anthropic keine Modelle")
	}

	modelle, befunde, fehler := anthropic.NeuerKlient(anthropic.Vorgabeadresse,
		netz.NeuerKlient(netz.UeblicheFristen()), anthropic.Zugang{Schluessel: schluessel}).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(modelle) == 0 && len(befunde) == 0 {
		t.Fatal("Anthropic nennt weder Modelle noch Befunde")
	}

	t.Logf("%d Modelle, %d Befunde", len(modelle), len(befunde))
}
