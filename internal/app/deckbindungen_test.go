package app_test

// **Die Deckbindungen reichen wirklich durch** — Befund 2 des Nachlaufs zum
// Zweitblick über Phase 5.1
// (#1405).
//
// # Warum es diesen Lauf gibt
//
// **Weil `Bindungen.DeckplatzVerschieben` wirkungslos gemacht werden konnte,
// ohne dass ein Lauf rot wurde.** Der Nachlauf hat es probiert: nur die
// Abdeckung sank. Eine Bindung, die niemand auf ihre **Antwort** prüft, ist
// eine Zeile, die aussieht, als täte sie etwas.
//
// **Die Fensterbindung ist die vierte Fläche**, und der Flächenlauf unter
// `internal/flaechen` erreicht sie nicht: er startet Prozesse, und das Fenster
// ist keiner. Hier steht sie.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Bereichswechsel kommt an der Bindung als Antwort zurück.**
func TestDieBindungVerschiebtEinenDeckplatz(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitDecknaht(&deckprobe{deck: domain.Deck{ID: 7, Name: "Aggro"}})

	bindungen := app.NeueBindungen(anwendung)

	antwort := bindungen.DeckplatzVerschieben(t.Context(), "7", "3", "sideboard")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("verschieben: %+v", antwort.Fehler)
	case antwort.Platz == nil:
		t.Fatal("kein Platz in der Antwort")
	case antwort.Platz.ID != "3" || antwort.Platz.Bereich != "sideboard":
		t.Errorf("%+v", antwort.Platz)
	}
}

// **Und ein Befund kommt als Befund zurück.**
func TestDieBindungMeldetEineKaputteKennung(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitDecknaht(&deckprobe{})

	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.DeckplatzVerschieben(t.Context(), "x", "3", "sideboard"); antwort.Fehler == nil ||
		antwort.Fehler.Feld != "deckID" {
		t.Errorf("%+v", antwort.Fehler)
	}
}

// **Ein Fehler ohne Klasse kommt ohne Klasse an** — und das ist Absicht.
//
// `CodeUnbekannt` heißt: wir wissen nicht, was es war. Ihn auf eine Klasse
// abzubilden hieße zu behaupten, der Aufrufer sei schuld — derselbe Satz steht
// bei [Fehler.AlsFehler] in `nahtfehler.go`. Der Weg über die **Werkstatt**
// (MCP und Dienst) muss ihn genauso behandeln.
func TestEinFehlerOhneKlasseBleibtOhneKlasse(t *testing.T) {
	t.Parallel()

	// **Ein geschlossener Bestand gibt einen Fehler, den niemand einordnet** —
	// „sql: database is closed". Genau das ist [CodeUnbekannt].
	bestand := frischerBestand(t)
	if err := bestand.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	_, fehler := app.NeueDeckwerkstatt(bestand).DeckplatzVerschieben(
		t.Context(), 7, 3, "sideboard")

	switch {
	case fehler == nil:
		t.Fatal("der Fehler ging verloren")
	case errors.Is(fehler, domain.ErrValidation),
		errors.Is(fehler, domain.ErrNotFound),
		errors.Is(fehler, domain.ErrConflict):
		t.Errorf("%v trägt eine Klasse, die niemand kennt", fehler)
	case !strings.Contains(fehler.Error(), "unbekannt"):
		t.Errorf("Fehler %q nennt den Code nicht", fehler.Error())
	}
}
