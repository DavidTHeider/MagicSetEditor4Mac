//go:build netzlauf

// **Scrydex am Original** — #2049, Entscheidung 1 in `docs/INTEGRATIONEN.md`:
// bis hierher nur gegen selbst geschriebene Antworten geprüft.
//
// **Mit dem Zugang des Auftraggebers**, aus `MSE_NETZLAUF_SCRYDEX_SCHLUESSEL` und `MSE_NETZLAUF_SCRYDEX_TEAM`. Fehlt er, endet
// der Lauf **rot mit BLOCKED** — nie grün, denn ein übersprungener Lauf
// belegt nichts. **Lesend**: eine Kartensuche nach einem Namen, den es seit 1993 gibt.
//
//	MSE_NETZLAUF_… go test -count=1 -tags netzlauf -run AmOriginal -v ./internal/scrydex/
//
// **Nicht in der Pipeline**: er braucht einen Zugang, den kein zweiter
// Rechner haben muss, und er geht ins Netz (§16).
package scrydex_test

import (
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

func TestScrydexAmOriginal(t *testing.T) {
	zugang := scrydex.Zugang{
		Schluessel: os.Getenv("MSE_NETZLAUF_SCRYDEX_SCHLUESSEL"),
		Team:       os.Getenv("MSE_NETZLAUF_SCRYDEX_TEAM"),
	}
	if zugang.Schluessel == "" || zugang.Team == "" {
		t.Fatal("BLOCKED: MSE_NETZLAUF_SCRYDEX_SCHLUESSEL und MSE_NETZLAUF_SCRYDEX_TEAM fehlen — Scrydex verlangt beides")
	}

	karten, fehler := scrydex.NeuerKlient(scrydex.Adresse,
		netz.NeuerKlient(netz.UeblicheFristen()), zugang).Karten(t.Context(), "Black Lotus")
	if fehler != nil {
		t.Fatalf("Karten: %v", fehler)
	}

	if len(karten) == 0 {
		t.Fatal("Scrydex findet »Black Lotus« nicht")
	}

	t.Logf("%d Karten, die erste: %+v", len(karten), karten[0])
}
