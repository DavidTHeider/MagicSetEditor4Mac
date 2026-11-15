package service_test

// **Ein neues Bild einer Karte verdrängt ihr altes** — E06
// (#2278).
//
// Gemessen im Zwei-Stunden-Lauf vom 2026-10-11: der Go-Heap wuchs über 80
// Minuten gleichmäßig um rund 1,7 MB je Minute, und fast alles davon lag im
// Fach der Spielbilder. Jede Runde ändert eine Karte; ihr Bild bekommt einen
// neuen Schlüssel, und das alte, auf das keine Karte mehr zeigt, blieb liegen,
// bis das Budget es verdrängte. **Der Schlüssel ist der Inhalt** — sieben
// gleiche Karten teilen sich ein Bild —, deshalb zählt das Fach, wie viele
// Karten auf ein Bild zeigen, und wirft es erst weg, wenn es keine mehr ist.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinNeuesBildEinerKarteVerdraengtIhrAltes(t *testing.T) {
	t.Parallel()

	werk, plaene, _, ablage := eineWerkstatt(t)

	alt, fehler := werk.Textur(t.Context(), eineTexturkarte(1), service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	// **Die Karte ändert sich** — ihr Plan ist ein anderer.
	plaene.plan = texturplan("anders.png", "MPlantin")

	neu, fehler := werk.Textur(t.Context(), eineTexturkarte(1), service.SeiteVorne, service.GueteTisch)
	if fehler != nil || neu.Schluessel == alt.Schluessel {
		t.Fatalf("das neue Bild hat denselben Schlüssel oder scheiterte: %v", fehler)
	}

	if _, _, _, da := ablage.Adresse(alt.Schluessel); da {
		t.Error("das alte Bild liegt noch im Fach, obwohl keine Karte mehr darauf zeigt")
	}

	if _, _, _, da := ablage.Adresse(neu.Schluessel); !da {
		t.Error("das neue Bild fehlt")
	}
}

// **Ein Bild, auf das noch eine Karte zeigt, bleibt** — die Gegenprobe: zwei
// gleiche Karten teilen sich eines, und eine davon ändert sich.
func TestEinGeteiltesBildBleibtBisZurLetztenKarte(t *testing.T) {
	t.Parallel()

	werk, plaene, _, ablage := eineWerkstatt(t)

	alt, fehler := werk.Textur(t.Context(), eineTexturkarte(1), service.SeiteVorne, service.GueteTisch)
	if fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if _, fehler := werk.Textur(t.Context(), eineTexturkarte(2), service.SeiteVorne, service.GueteTisch); fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	plaene.plan = texturplan("anders.png", "MPlantin")

	if _, fehler := werk.Textur(t.Context(), eineTexturkarte(1), service.SeiteVorne, service.GueteTisch); fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if _, _, _, da := ablage.Adresse(alt.Schluessel); !da {
		t.Error("das geteilte Bild ist fort, obwohl Karte 2 noch darauf zeigt")
	}

	// **Und mit der letzten geht es** — derselbe Weg für Karte 2.
	if _, fehler := werk.Textur(t.Context(), eineTexturkarte(2), service.SeiteVorne, service.GueteTisch); fehler != nil {
		t.Fatalf("Textur: %v", fehler)
	}

	if _, _, _, da := ablage.Adresse(alt.Schluessel); da {
		t.Error("das Bild liegt noch, obwohl keine Karte mehr darauf zeigt")
	}
}
