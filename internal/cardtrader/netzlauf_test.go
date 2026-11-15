//go:build netzlauf

// **CardTrader am Original** — #2049, Entscheidung 1 in `docs/INTEGRATIONEN.md`:
// bis hierher nur gegen selbst geschriebene Antworten geprüft.
//
// **Mit dem Zugang des Auftraggebers**, aus `MSE_NETZLAUF_CARDTRADER`. Fehlt er, endet
// der Lauf **rot mit BLOCKED** — nie grün, denn ein übersprungener Lauf
// belegt nichts. **Lesend und ohne Konto**: die Liste der Spiele — sie belegt Marke, Kopfzeilen und Antwortgestalt, ohne etwas am Konto anzufassen.
//
//	MSE_NETZLAUF_… go test -count=1 -tags netzlauf -run AmOriginal -v ./internal/cardtrader/
//
// **Nicht in der Pipeline**: er braucht einen Zugang, den kein zweiter
// Rechner haben muss, und er geht ins Netz (§16).
package cardtrader_test

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestCardTraderAmOriginal(t *testing.T) {
	marke := os.Getenv("MSE_NETZLAUF_CARDTRADER")
	if marke == "" {
		t.Fatal("BLOCKED: MSE_NETZLAUF_CARDTRADER fehlt — die Marke steht in den Profileinstellungen des CardTrader-Kontos")
	}

	spiele, fehler := cardtrader.NeuerKlient(cardtrader.Vorgabeadresse,
		netz.NeuerKlient(netz.UeblicheFristen()), cardtrader.Zugang{Marke: marke}).Spiele(t.Context())
	if fehler != nil {
		t.Fatalf("Spiele: %v", fehler)
	}

	if len(spiele) == 0 {
		t.Fatal("CardTrader nennt kein einziges Spiel")
	}

	t.Logf("%d Spiele, das erste: %+v", len(spiele), spiele[0])
}
