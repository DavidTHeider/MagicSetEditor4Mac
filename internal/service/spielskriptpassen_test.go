package service_test

// **Der Befund eines Kartenskripts am Ereignis des Passens** —
// #1625.
//
// **Herausgezogen aus `spielskriptlauf_test.go`**, weil die Datei an der
// Zeilengrenze stand (§2). Der Schnitt folgt der Frage: dort steht, **was**
// ein Skript bewirkt, hier, **wann** sein Befund an der Fläche ankommt.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Und beim Passen trägt der Strom den Befund** —
// #1625.
//
// **Der Weg, den ein Mensch geht**: er wirkt, alle passen, der Stapel
// verrechnet — und erst dort wird gelesen. Der Lauf darüber prüft die
// Verrechnung unmittelbar; dieser prüft, dass der Befund **an der Fläche**
// ankommt (§10: eine Zusage, die kein Weg einlöst, ist keine).
func TestDerBefundKommtAmEreignisDesPassens(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einSpielskriptstand(t, "a := { a() }\na()")

	_, gewirkt, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlWirken,
		Nutzlast: map[string]any{"objekt": "sitz-1/blitz/0"},
	})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Beide passen** — danach verrechnet der Stapel (CR 117.3b).
	stand := gewirkt

	var letzte []domain.Ereignis

	for nummer, sitz := range []string{"sitz-1", "sitz-2"} {
		ereignisse, danach, gepasst := kasten.Auslegen(stand,
			domain.Spielbefehl{
				PartieID: "p1", Absender: sitz,
				Kennung: fmt.Sprintf("p%d", nummer),
				Art:     service.BefehlPassen,
			})
		if gepasst != nil {
			t.Fatalf("passen (%s): %v", sitz, gepasst)
		}

		stand, letzte = danach, ereignisse
	}

	befund, da := letzte[0].Oeffentlich["befund"].(string)
	if !da || !strings.Contains(befund, "Kartenskript") {
		t.Errorf("der Befund fehlt: %+v", letzte[0].Oeffentlich)
	}
}
