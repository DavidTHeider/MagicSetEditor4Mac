// **Eine bestätigte Vorschau ist an ihren Datenstand gebunden** —
// #1241.
//
// # Der Befund
//
// Die Ersetzenseite bindet ihre Vorschau seit
// #1239 an die
// **Frage**, aus der sie stammt: ändert jemand Suchtext, Ersatz oder ein
// Kästchen, verschwindet der Knopf *Jetzt ersetzen*.
//
// **Was sie nicht band, ist der Datenstand.** Zwischen Vorschau und
// Bestätigung kann eine Karte durch den Editor, eine Rücknahme, den
// REST-Dienst oder den **unabhängigen MCP-Prozess** geändert werden.
// `Ausfuehren` las den Bestand dann neu und rechnete einen neuen Änderungsplan:
// der Knopf bestätigte unter Umständen **andere Treffer als die angezeigten**.
//
// Dass die Ausführung danach die tatsächlich geschriebenen Werte zurückgibt,
// macht die Bestätigung davor nicht gültig.
//
// # Die Klasse
//
// > Ein Bestätigungsschritt muss an **Absicht und Quellzustand** gebunden sein.
//
// # Optimistisch, nicht sperrend
//
// Die Sperre wäre teurer als der Fall: sie hielte den Bestand über die Zeit,
// die jemand zum Lesen einer Vorschau braucht. Der Stand wird stattdessen
// **mitgegeben und beim Ausführen nachgerechnet**.
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineGeaenderteKarteWeistDieBestaetigungAb(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()
	ersetzt := ersetzer(doppel)

	frage := domain.Ersetzungsfrage{Text: "Fliegend", Ersatz: "Flying"}

	schau, err := ersetzt.Vorschau(t.Context(), 1, frage)
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	if schau.Stand == "" {
		t.Fatal("die Vorschau nennt keinen Stand — dann ist nichts zu binden")
	}

	// **Jemand anderes ändert die Karte** — der Editor, eine Rücknahme, der
	// Dienst oder der MCP-Prozess.
	doppel.karten[0].Values[1].Text = "<i>Fliegend</i> und schnell."

	frage.Stand = schau.Stand

	if _, fehler := ersetzt.Ausfuehren(t.Context(), 1, frage); !errors.Is(
		fehler, domain.ErrConflict,
	) {
		t.Errorf("Fehler %v, erwartet ErrConflict — bestätigt wurde eine "+
			"Vorschau, die nicht mehr gilt", fehler)
	}
}

// **Ein unveränderter Stand wird ausgeführt.**
func TestEinUnveraenderterStandWirdAusgefuehrt(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()
	ersetzt := ersetzer(doppel)

	frage := domain.Ersetzungsfrage{Text: "Fliegend", Ersatz: "Flying"}

	schau, err := ersetzt.Vorschau(t.Context(), 1, frage)
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	frage.Stand = schau.Stand

	if _, fehler := ersetzt.Ausfuehren(t.Context(), 1, frage); fehler != nil {
		t.Fatalf("Ausfuehren: %v", fehler)
	}

	if len(doppel.geschrieben) == 0 {
		t.Error("es wurde nichts geschrieben")
	}
}

// **Ohne Stand wird ohne Prüfung ausgeführt.**
//
// Das ist der Weg der Kommandozeile, von MCP und vom Dienst: dort gibt es
// keinen Bestätigungsschritt, den etwas binden könnte. Wer ihn erzwänge,
// bräche drei Flächen für eine Zusage, die dort niemand gibt.
func TestOhneStandWirdOhnePruefungAusgefuehrt(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()

	if _, fehler := ersetzer(doppel).Ausfuehren(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Fliegend", Ersatz: "Flying"}); fehler != nil {
		t.Fatalf("Ausfuehren: %v", fehler)
	}

	if len(doppel.geschrieben) == 0 {
		t.Error("es wurde nichts geschrieben")
	}
}

// **Und eine Karte, die dazukommt, ändert den Stand ebenso.**
//
// Sie trägt einen Treffer, den die Vorschau nicht zeigte — und der ginge sonst
// stillschweigend mit.
func TestEineNeueKarteAendertDenStand(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()
	ersetzt := ersetzer(doppel)

	frage := domain.Ersetzungsfrage{Text: "Fliegend", Ersatz: "Flying"}

	schau, err := ersetzt.Vorschau(t.Context(), 1, frage)
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	doppel.karten = append(doppel.karten, &domain.CardRow{
		ID: 99, SetID: 1, Number: 9,
		Values: []domain.Value{{Field: "rule text", Text: "Fliegend"}},
	})

	frage.Stand = schau.Stand

	if _, fehler := ersetzt.Ausfuehren(t.Context(), 1, frage); !errors.Is(
		fehler, domain.ErrConflict,
	) {
		t.Errorf("Fehler %v, erwartet ErrConflict", fehler)
	}
}
