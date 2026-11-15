package service_test

// **Der Friedhof ist eine geordnete Zone** — Werkstück G9 von Phase 5.5
// (#1592).
//
// **Eine eigene Datei**, weil `sandkastenschieben_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge — und weil der Schnitt der Frage folgt: dort geht
// es darum, **wie** geschoben wird, hier darum, **wohin** es überhaupt eine
// Stelle gibt.
//
// **CR 400.5** (Fassung 2026-08-19) nennt drei Zonen, deren Reihenfolge sich
// nur ändern lässt, wenn eine Regel oder ein Effekt es erlaubt: Bibliothek,
// **Friedhof** und Stapel. Bis zum 2026-09-17 sagte `Zone.Geordnet` das
// Gegenteil, das gezeichnete Tischbild sagte es richtig, und die Regelmatrix
// führte den Widerspruch als Grenze.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Und auf den Friedhof geht sie, weil er geordnet ist** — CR 400.5.
//
// **Daran hängt der Abwurf im Aufräumschritt** (CR 514.1, CR 404.1): er legt
// eine Karte **oben** auf den Friedhof, und solange der als ungeordnet galt,
// bekam jede Karte darin die Ordnung null.
func TestObenAufDenFriedhofGeht(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, danach, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art: service.BefehlSchieben,
		Nutzlast: map[string]any{
			"objekt": "anna/k1", "zone": "friedhof", "stelle": "oben",
		},
	})
	if fehler != nil {
		t.Fatalf("»oben auf den Friedhof«: %v", fehler)
	}

	for _, objekt := range danach.Objekte {
		if objekt.Kennung != "anna/k1" {
			continue
		}

		switch {
		case objekt.Zone != domain.ZoneFriedhof:
			t.Errorf("die Karte liegt in %q", objekt.Zone)
		case objekt.Ordnung != 1:
			t.Errorf("sie liegt an Stelle %d statt oben", objekt.Ordnung)
		}
	}
}
