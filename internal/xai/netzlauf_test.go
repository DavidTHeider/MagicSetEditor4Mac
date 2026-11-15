//go:build netzlauf

// **xAI (Cloudmodelle) am Original** — #2049, Entscheidung 1 in `docs/INTEGRATIONEN.md`:
// bis hierher nur gegen selbst geschriebene Antworten geprüft.
//
// **Mit dem Zugang des Auftraggebers**, aus `MSE_NETZLAUF_XAI`. Fehlt er, endet
// der Lauf **rot mit BLOCKED** — nie grün, denn ein übersprungener Lauf
// belegt nichts. **Nur die Modellliste**: sie kostet nichts und belegt doch Schlüssel, Kopfzeilen und Antwortgestalt. Ein Lauf, der rechnen lässt, kostet Guthaben und bleibt einer eigenen Freigabe vorbehalten.
//
//	MSE_NETZLAUF_… go test -count=1 -tags netzlauf -run AmOriginal -v ./internal/xai/
//
// **Nicht in der Pipeline**: er braucht einen Zugang, den kein zweiter
// Rechner haben muss, und er geht ins Netz (§16).
package xai_test

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

func TestXAIAmOriginal(t *testing.T) {
	schluessel := os.Getenv("MSE_NETZLAUF_XAI")
	if schluessel == "" {
		t.Fatal("BLOCKED: MSE_NETZLAUF_XAI fehlt — ohne Schlüssel nennt xAI keine Modelle")
	}

	modelle, befunde, fehler := xai.NeuerKlient(xai.Vorgabeadresse,
		netz.NeuerKlient(netz.UeblicheFristen()), xai.Zugang{Schluessel: schluessel}).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(modelle) == 0 && len(befunde) == 0 {
		t.Fatal("xAI nennt weder Modelle noch Befunde")
	}

	t.Logf("%d Modelle, %d Befunde", len(modelle), len(befunde))
}
