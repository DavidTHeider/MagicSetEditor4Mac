package storage_test

// **Ein offenes Sideboard überlebt den Bestand** — Befund 1 des Nachlaufs
// über Phase 4.6.
//
// # Der Befund
//
// Die Korrekturrunde führte `Format.SideboardOffen` ein und setzte es für
// `BauartLimitiert`. Das Feld stand aber **nur** in `internal/domain`: die
// Tabelle kannte es nicht, der Leser gab es nicht zurück, das DTO trug es
// nicht. Nach dem ersten Schreiben war wieder der Nullwert `false` in Kraft.
//
// **Beobachtet, nicht erschlossen**: `mse deck pruefen` gab für ein
// limitiertes Deck mit drei Karten im Sideboard aus
//
//	! Das Sideboard hat 3 Karten; "Draft" kennt kein Sideboard.
//
// Damit war die Korrektur aus Befund 2 des Zweitblicks in der laufenden
// Anwendung **wirkungslos** — grün in `internal/domain`, falsch überall
// sonst. Genau die Grenze aus §10: zwei grüne Seiten sind kein Nachweis für
// die Grenze dazwischen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Angelegt, wieder gelesen — und immer noch offen.**
func TestEinOffenesSideboardUeberlebtDasSchreiben(t *testing.T) {
	ablage := storage.NeueFormatablage(frischeDatenbank(t))

	vorlage, gibts := domain.VonBauart(domain.BauartLimitiert)
	if !gibts {
		t.Fatal("die Bauart »limitiert« gibt es nicht")
	}

	vorlage.Name, vorlage.Spiel = "Draft", "magic"

	angelegt, err := ablage.Anlegen(probekontext(t), vorlage)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if !angelegt.SideboardOffen {
		t.Fatal("schon das angelegte Format hat kein offenes Sideboard")
	}

	gelesen, err := ablage.Format(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !gelesen.SideboardOffen {
		t.Error("das gelesene Format hat kein offenes Sideboard mehr — " +
			"die Spalte fehlt oder wird nicht gelesen")
	}

	// **Und in der Liste auch.** Ein Leser, der es kann, und ein Auflister,
	// der es nicht kann, wäre dieselbe Lücke eine Ebene weiter.
	liste, err := ablage.Formate(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("auflisten: %v", err)
	}

	if len(liste) != 1 || !liste[0].SideboardOffen {
		t.Errorf("in der Liste steht %+v", liste)
	}
}

// **Und es überlebt eine Änderung.**
func TestEinOffenesSideboardUeberlebtDasAendern(t *testing.T) {
	ablage := storage.NeueFormatablage(frischeDatenbank(t))

	vorlage, _ := domain.VonBauart(domain.BauartLimitiert)
	vorlage.Name, vorlage.Spiel = "Draft", "magic"

	angelegt, err := ablage.Anlegen(probekontext(t), vorlage)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	angelegt.Name = "Sealed"

	geaendert, err := ablage.Aendern(probekontext(t), angelegt)
	if err != nil {
		t.Fatalf("ändern: %v", err)
	}

	if !geaendert.SideboardOffen {
		t.Error("nach dem Ändern ist das Sideboard nicht mehr offen")
	}

	gelesen, err := ablage.Format(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !gelesen.SideboardOffen {
		t.Error("und aus dem Bestand kommt es auch nicht zurück")
	}
}
