package mtgjson_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Der Nachlauf hat einen Fall benannt, der bleibt** — und er steht hier,
// damit er nicht als übersehen gilt.
//
// # `loyalty cost 2` an einer wirklich zweiseitigen Karte
//
// [mtgjson.TrenneKartenwerte] verschiebt an einer Karte **mit** zweitem Namen
// jedes Feld mit dem Zusatz ` 2` nach hinten — auch `loyalty cost 2`, das dort
// die zweite Fähigkeit der **Vorderseite** meint.
//
// **Warum das bleibt**: die Unterscheidung braucht die Feldliste des Spiels
// ([domain.Game.Feldreihen]), und der Schreiber hat sie nicht — `Zuordnung`
// kennt den Spielnamen, nicht seine Felder. Sie hineinzureichen hieße, die
// Signatur durch `service.Exporteur` und seinen `SetLeser` zu ziehen, für
// einen Fall, dessen Wirkung nicht messbar ist.
//
// **Und genau das hält dieser Lauf fest**: keines der betroffenen Felder wird
// auf ein MTG-JSON-Feld abgebildet. Was nach hinten wandert, landet in keiner
// geschriebenen Angabe. Der Nachlauf über Phase 3.6 kam zum selben Ergebnis
// und konnte keine Verfälschung belegen.
//
// **Er hält die Aussage, nicht die Behebung.** Wird die Zuordnung erweitert,
// sodass `loyalty cost` doch abgebildet wird, macht er die Pipeline rot — und
// dann ist der Fall zu entscheiden.
func TestDieHilfsfelderBildenKeinMTGJSONFeld(t *testing.T) {
	t.Parallel()

	zuordnung, bekannt := mtgjson.ZuordnungFuer("magic")
	if !bekannt {
		t.Fatal("für magic gibt es keine Zuordnung")
	}

	betroffen := map[string]bool{"loyalty cost": true, "additional credit": true}

	for _, paar := range zuordnung.Felder {
		if betroffen[paar.Feld] {
			t.Errorf("%q wird abgebildet — der benannte Restfall ist keiner mehr", paar.Feld)
		}
	}
}
