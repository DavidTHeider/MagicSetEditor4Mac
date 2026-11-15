package service_test

// **Was ein Sitz an sich selbst ändert** — Leben, Vergiftungsmarken und
// markierter Schaden (G9, #1592).
//
// **Eine eigene Datei, und der Schnitt folgt der Frage** (§2): nebenan steht,
// wie ein **Zug** weitergeht — hier, was ein **Spieler** an seinem Stand
// ändert. Der Anlass war die Zeilengrenze; mit `M1` ging die eine Datei auf
// 451 Zeilen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Leben ändert sich um einen Betrag.**
func TestLebenAendertSichUmEinenBetrag(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, danach, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
			Art: service.BefehlLeben, Nutzlast: map[string]any{"um": -3.0},
		})
	if fehler != nil {
		t.Fatalf("leben: %v", fehler)
	}

	switch {
	case danach.Sitze[1].Leben != 17:
		t.Errorf("der Sitz hat %d Leben", danach.Sitze[1].Leben)
	case ereignisse[0].Oeffentlich["stand"] != 17.0:
		t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
	}

	// **Und der andere Sitz bleibt unberührt** — ohne diese Zeile wäre der
	// Lauf auch grün, wenn der Befehl jedem Sitz drei nähme.
	if danach.Sitze[0].Leben != 20 {
		t.Errorf("der andere Sitz hat %d Leben", danach.Sitze[0].Leben)
	}
}

// **Vergiftungsmarken kommen dazu und stehen im Ereignis.**
func TestVergiftungsmarkenKommenDazu(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, danach, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
			Art: service.BefehlVergiftung, Nutzlast: map[string]any{"um": 2.0},
		})
	if fehler != nil {
		t.Fatalf("vergiftung: %v", fehler)
	}

	switch {
	case danach.Sitze[1].Vergiftung != 2:
		t.Errorf("der Sitz hat %d Marken", danach.Sitze[1].Vergiftung)
	case ereignisse[0].Oeffentlich["stand"] != 2.0:
		t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
	}
}

// **Eine negative Vergiftungsmarke wird abgewiesen** — CR 122.1.
func TestEineNegativeVergiftungsmarkeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art: service.BefehlVergiftung, Nutzlast: map[string]any{"um": -1.0},
	})

	if fehler == nil {
		t.Error("weniger als keine Marke kam durch")
	}
}

// **Ein Befehl ohne Betrag wird abgewiesen.**
func TestEinMarkenbefehlOhneBetragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	for name, nutzlast := range map[string]map[string]any{
		"ohne Feld":  {},
		"keine Zahl": {"um": "drei"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := kasten.Auslegen(eineZugpartie(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art: service.BefehlLeben, Nutzlast: nutzlast,
				})

			if fehler == nil {
				t.Error("der Befehl kam durch")
			}
		})
	}
}

// **Wer nicht am Tisch sitzt, ändert keinen Stand.**
func TestEinFremderAendertKeinenStand(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "niemand", Kennung: "b1",
		Art: service.BefehlLeben, Nutzlast: map[string]any{"um": -3.0},
	})

	if fehler == nil {
		t.Error("ein Fremder hat seinen Stand geändert")
	}
}

// **Markierter Schaden liegt am Objekt** — CR 120.3.
func TestMarkierterSchadenLiegtAmObjekt(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	objekt := partie.Objekte[0].Kennung

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlSchaden,
		Nutzlast: map[string]any{"objekt": objekt, "um": 2.0},
	})
	if fehler != nil {
		t.Fatalf("schaden: %v", fehler)
	}

	if danach.Objekte[0].Schaden != 2 {
		t.Errorf("auf dem Objekt sind %d Schaden", danach.Objekte[0].Schaden)
	}

	// **Und negativer Schaden wird abgewiesen.**
	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b2",
		Art:      service.BefehlSchaden,
		Nutzlast: map[string]any{"objekt": objekt, "um": -1.0},
	}); fehler == nil {
		t.Error("negativer Schaden kam durch")
	}
}

// **Ein Schadensbefehl ohne Betrag wird abgewiesen.**
func TestEinSchadensbefehlOhneBetragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlSchaden,
		Nutzlast: map[string]any{"objekt": partie.Objekte[0].Kennung},
	})

	if fehler == nil {
		t.Error("der Befehl kam durch")
	}
}
