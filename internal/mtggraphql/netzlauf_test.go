//go:build netzlauf

// **Die GraphQL-Schnittstelle von MTGJSON am Original** — #2049, Entscheidung 1 in `docs/INTEGRATIONEN.md`:
// bis hierher nur gegen selbst geschriebene Antworten geprüft.
//
// **Mit dem Zugang des Auftraggebers**, aus `MSE_NETZLAUF_MTGJSON`. Fehlt er, endet
// der Lauf **rot mit BLOCKED** — nie grün, denn ein übersprungener Lauf
// belegt nichts. **Lesend**: eine Kartensuche nach einem Namen, den es seit 1993 gibt — ohne Marke antwortet der Dienst mit 200 und `Access denied` im Rumpf.
//
//	MSE_NETZLAUF_… go test -count=1 -tags netzlauf -run AmOriginal -v ./internal/mtggraphql/
//
// **Nicht in der Pipeline**: er braucht einen Zugang, den kein zweiter
// Rechner haben muss, und er geht ins Netz (§16).
package mtggraphql_test

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestMTGJSONGraphQLAmOriginal(t *testing.T) {
	marke := os.Getenv("MSE_NETZLAUF_MTGJSON")
	if marke == "" {
		t.Fatal("BLOCKED: MSE_NETZLAUF_MTGJSON fehlt — ohne Marke antwortet die Schnittstelle mit »Access denied«")
	}

	karten, fehler := mtggraphql.NeuerKlient(mtggraphql.Adresse,
		netz.NeuerKlient(netz.UeblicheFristen()), marke).Karten(t.Context(), "Black Lotus")
	if fehler != nil {
		t.Fatalf("Karten: %v", fehler)
	}

	if len(karten) == 0 {
		t.Fatal("MTGJSON findet »Black Lotus« nicht")
	}

	t.Logf("%d Karten, die erste: %+v", len(karten), karten[0])
}
