package service_test

// **Was der Sandkasten bekommt, ändert er nicht** — Befund beim Bauen von M3
// (#1605),
// gemessen am 2026-09-17.
//
// # Der Befund
//
// **`Sandkasten.kopie` hieß Kopie und war eine halbe.** Sie legte die
// Objektreihe neu an — samt der Zählertabellen darin — und reichte die
// **Sitzreihe** unverändert weiter. Eine Reihe ist ein Zeiger auf einen
// gemeinsamen Speicher: `danach.Sitze[0].Leben += um` schrieb damit in die
// Partie, die der Aufrufer hält.
//
// **Gemessen**: ein `spieler.leben` mit −5 auf eine Partie mit 20 Leben, und
// die **übergebene** Partie stand danach bei 15.
//
// **Warum es bis dahin nicht gestört hat**: der Aufrufer wirft den alten Stand
// auf dem Erfolgsweg ohnehin weg. Es stört auf dem **Fehlerweg** — dort kommt
// ein Fehler zurück und der Stand ist trotzdem verändert (§10: es bleibt kein
// halber Stand zurück) — und ab `M3`, weil `Sitz` mit dem Manavorrat eine
// **Tabelle** trägt und eine geteilte Tabelle noch weiter reicht als eine
// geteilte Reihe.
//
// # Was dieser Lauf nicht prüft
//
// **Ob der Aufrufer den alten Stand braucht.** Er prüft die Zusage des
// Namens: was `kopie` heißt, ist eine.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Befehl lässt die übergebene Partie stehen.**
func TestEinBefehlLaesstDieUebergebenePartieStehen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	vorher := partie.Sitze[0].Leben

	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlLeben, Nutzlast: map[string]any{"um": float64(-5)},
	}); fehler != nil {
		t.Fatalf("leben: %v", fehler)
	}

	if partie.Sitze[0].Leben != vorher {
		t.Errorf("die übergebene Partie steht bei %d statt %d",
			partie.Sitze[0].Leben, vorher)
	}
}

// **Und auch die Tabelle im Sitz bleibt die eigene** — der Manavorrat (M3).
func TestEinBefehlLaesstDenUebergebenenVorratStehen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	partie, fehler := eineZugpartie(t).
		ManaHinzu("sitz-1", domain.ManaRot, 2)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlLeben, Nutzlast: map[string]any{"um": float64(-1)},
	}); fehler != nil {
		t.Fatalf("leben: %v", fehler)
	}

	if zahl := partie.Sitze[0].Manavorrat.Zahl(domain.ManaRot); zahl != 2 {
		t.Errorf("der übergebene Vorrat steht bei %d roten", zahl)
	}
}
