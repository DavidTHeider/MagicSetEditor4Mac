// **Die Deckwerkstatt für MCP und REST** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Sie ist ein Durchreicher**, und der Lauf prüft genau das: dass jede der
// fünfzehn Methoden ankommt und ihr Ergebnis zurückgibt. Was fachlich gilt,
// prüfen `internal/service` und `internal/flaechen`.
package app_test

import (
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der ganze Weg über die Werkstatt.**
func TestDerDeckwegUeberDieWerkstatt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	werkstatt := app.NeueDeckwerkstatt(frischerBestand(t))

	deck, fehler := werkstatt.DeckAnlegen(probekontext(t), "Aggro", "magic", "")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	deckID := kennungAusWerkstatt(t, deck)

	mussWerkstatt(t)(werkstatt.Decks(probekontext(t), ""))
	mussWerkstatt(t)(werkstatt.Deck(probekontext(t), deckID))
	mussWerkstatt(t)(werkstatt.DeckUmbenennen(probekontext(t), deckID, 1, "Midrange", ""))

	platz, legefehler := werkstatt.DeckplatzLegen(probekontext(t), deckID, 0,
		"a1b2c3d4-0000-4000-8000-00000000000a", 4, "hauptdeck")
	if legefehler != nil {
		t.Fatalf("legen: %v", legefehler)
	}

	format, formatfehler := werkstatt.FormatAusBauart(
		probekontext(t), "konstruiert", "Konstruiert", "magic", "standard")
	if formatfehler != nil {
		t.Fatalf("das Format: %v", formatfehler)
	}

	formatID := kennungAusWerkstatt(t, format)

	mussWerkstatt(t)(werkstatt.Formate(probekontext(t), ""))
	mussWerkstatt(t)(werkstatt.Bauarten(probekontext(t)))
	mussWerkstatt(t)(werkstatt.DeckFormatWaehlen(probekontext(t), deckID, formatID))
	mussWerkstatt(t)(werkstatt.DeckPruefen(probekontext(t), deckID))
	mussWerkstatt(t)(werkstatt.DeckAbgleichen(probekontext(t), deckID))

	// **Kopie und Wunschliste** — DOM-03.AK-02 (#1698).
	mussWerkstatt(t)(werkstatt.DeckKopieren(probekontext(t), deckID, "Aggro II"))
	mussWerkstatt(t)(werkstatt.DeckFehlkartenWuenschen(probekontext(t), deckID))

	// **Aus einem Set, das es nicht gibt** — FMT-04 (#1693): der Fehler kommt
	// als Fehler heraus und nicht als leeres Deck.
	if _, fehler := werkstatt.DeckAusSet(probekontext(t), 999, "", 1); fehler == nil {
		t.Error("aus einem fehlenden Set entstand ein Deck")
	}

	// **Was nicht gesagt wird, bleibt stehen** — die Werkstatt füllt es aus
	// dem vorhandenen Format.
	geaendert, aenderfehler := werkstatt.FormatSetzen(probekontext(t), map[string]any{
		"id": float64(formatID), "fassung": float64(1), "mindestgroesse": float64(40),
	})
	if aenderfehler != nil {
		t.Fatalf("das Format ändern: %v", aenderfehler)
	}

	pruefeErhalten(t, geaendert)

	// **Der Bereichswechsel über die Werkstatt** — Befund 7 des Zweitblicks
	// über Phase 5.1.
	// **Der Platz bekommt dabei eine neue Kennung** — er wird drüben gesetzt
	// und hier gelöscht; wer die alte weiterreicht, greift ins Leere.
	verschoben, verschiebefehler := werkstatt.DeckplatzVerschieben(
		probekontext(t), deckID, kennungAusWerkstatt(t, platz), "sideboard")
	if verschiebefehler != nil {
		t.Fatalf("verschieben: %v", verschiebefehler)
	}

	if err := werkstatt.DeckplatzNehmen(
		probekontext(t), deckID, kennungAusWerkstatt(t, verschoben)); err != nil {
		t.Errorf("nehmen: %v", err)
	}

	if err := werkstatt.DeckLoeschen(probekontext(t), deckID); err != nil {
		t.Errorf("löschen: %v", err)
	}

	if err := werkstatt.FormatLoeschen(probekontext(t), formatID); err != nil {
		t.Errorf("das Format löschen: %v", err)
	}
}

// **Und ein fachliches Nein kommt als Fehler heraus.**
//
// **Das Protokoll trennt Ergebnis und Fehlschlag**: ein Ergebnis mit einem
// Fehler darin läse sich wie ein Erfolg.
func TestEinNeinDerWerkstattIstEinFehler(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	werkstatt := app.NeueDeckwerkstatt(frischerBestand(t))

	for name, ruf := range map[string]func() error{
		"fehlendes Deck": func() error {
			_, err := werkstatt.Deck(probekontext(t), 999)

			return err
		},
		"Deck ohne Namen": func() error {
			_, err := werkstatt.DeckAnlegen(probekontext(t), "", "magic", "")

			return err
		},
		"fehlendes Format": func() error {
			return werkstatt.FormatLoeschen(probekontext(t), 999)
		},
		"erfundene Bauart": func() error {
			_, err := werkstatt.FormatAusBauart(probekontext(t), "turnier", "T", "magic", "")

			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if ruf() == nil {
				t.Error("das Nein kam als Erfolg zurück")
			}
		})
	}
}

// mussWerkstatt gibt eine Prüfung, die einen Ruf der Werkstatt entgegennimmt.
//
// **Als Funktion, die eine Funktion gibt**: Go lässt ein Aufrufergebnis mit
// zwei Werten nur dann als Argumentliste durch, wenn es die **einzige** ist —
// und `t` gehört auch dazu.
func mussWerkstatt(t *testing.T) func(any, error) {
	t.Helper()

	return func(ergebnis any, fehler error) {
		if fehler != nil {
			t.Fatalf("%v", fehler)
		}

		if ergebnis == nil {
			t.Error("das Ergebnis fehlt")
		}
	}
}

// kennungAusWerkstatt liest die Kennung aus einer Antwort der Werkstatt.
func kennungAusWerkstatt(t *testing.T, ergebnis any) int64 {
	t.Helper()

	switch gelesen := ergebnis.(type) {
	case *app.DeckDTO:
		return zahlAus(t, gelesen.ID)

	case *app.DeckplatzDTO:
		return zahlAus(t, gelesen.ID)

	case *app.FormatDTO:
		return zahlAus(t, gelesen.ID)

	// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
	case *app.ZuordnungDTO:
		return zahlAus(t, gelesen.ID)
	}

	t.Fatalf("unerwartete Antwort: %T", ergebnis)

	return 0
}

// pruefeErhalten sichert zu, dass eine ungenannte Zahl stehen blieb.
func pruefeErhalten(t *testing.T, ergebnis any) {
	t.Helper()

	format, ist := ergebnis.(*app.FormatDTO)
	if !ist {
		t.Fatalf("unerwartete Antwort: %T", ergebnis)
	}

	if format.Mindestgroesse != 40 {
		t.Errorf("Mindestgröße %d, erwartet 40", format.Mindestgroesse)
	}

	// **CR 100.4a und 100.2a** — was der Ruf nicht nannte, steht noch.
	if format.Sideboardgroesse != 15 || format.HoechstzahlGleich != 4 {
		t.Errorf("die übrigen Zahlen sind mitgegangen: %+v", format)
	}
}

// zahlAus macht aus einer Kennung als Text eine Zahl.
func zahlAus(t *testing.T, text string) int64 {
	t.Helper()

	var kennung int64

	if _, err := fmt.Sscanf(text, "%d", &kennung); err != nil {
		t.Fatalf("die Kennung %q ist keine Zahl: %v", text, err)
	}

	return kennung
}

// **Die Wege, die nur ein besonderer Ruf betritt.**
//
// **Eigens gefahren**, weil der Ablauf darüber sie nicht berührt: eine eigene
// Karte statt eines Drucks, ein Format ohne jede genannte Zahl, und eine
// Prüfung, die scheitert.
func TestDieSonderwegeDerWerkstatt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	werkstatt := app.NeueDeckwerkstatt(frischerBestand(t))

	deck, fehler := werkstatt.DeckAnlegen(probekontext(t), "Aggro", "magic", "")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	deckID := kennungAusWerkstatt(t, deck)

	// **Eine eigene Karte, die es nicht gibt** — der andere Zweig von
	// `DeckplatzLegen`, in dem die Kartenkennung als Text reist.
	if _, err := werkstatt.DeckplatzLegen(
		probekontext(t), deckID, 999, "", 1, "hauptdeck"); err == nil {
		t.Error("eine erfundene Karte kam durch")
	}

	// **Ein Format ohne jede Zahl** — `zeichenkette` und `ganzzahl` geben
	// dann ihre Vorgabe, und `ergaenzt` findet nichts zum Übernehmen.
	format, formatfehler := werkstatt.FormatSetzen(probekontext(t), map[string]any{
		"name": "Eigen", "spiel": "magic",
	})
	if formatfehler != nil {
		t.Fatalf("das Format: %v", formatfehler)
	}

	// **Und dasselbe mit einer Kennung, die es nicht gibt**: dann findet
	// `ergaenzt` kein vorhandenes Format und lässt die Nullen stehen.
	if _, err := werkstatt.FormatSetzen(probekontext(t), map[string]any{
		"id": float64(999), "fassung": float64(1), "name": "Nichts", "spiel": "magic",
	}); err == nil {
		t.Error("ein erfundenes Format kam durch")
	}

	// **Ein Deck ohne Format lässt sich nicht prüfen** — eine Frage ohne
	// Gegenstand.
	if _, err := werkstatt.DeckPruefen(probekontext(t), deckID); err == nil {
		t.Error("ein Deck ohne Format kam durch die Prüfung")
	}

	// **Und ein Ruf, der die Untergrenze **nicht** nennt** — sie ist die
	// einzige Zahl, die der Lauf darüber immer mitgibt, und ihr Zweig bliebe
	// sonst unbetreten.
	pruefeUebernommen(t, werkstatt, format)

	if err := werkstatt.FormatLoeschen(
		probekontext(t), kennungAusWerkstatt(t, format)); err != nil {
		t.Errorf("das Format löschen: %v", err)
	}
}

// pruefeUebernommen ändert nur den Namen — jede Zahl kommt aus dem Bestand.
func pruefeUebernommen(t *testing.T, werkstatt *app.Deckwerkstatt, format any) {
	t.Helper()

	kennung := kennungAusWerkstatt(t, format)

	// **Erst eine Zahl setzen, damit es etwas zu übernehmen gibt.**
	if _, err := werkstatt.FormatSetzen(probekontext(t), map[string]any{
		"id": float64(kennung), "fassung": float64(1),
		"name": "Eigen", "spiel": "magic", "mindestgroesse": float64(33),
	}); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	geaendert, fehler := werkstatt.FormatSetzen(probekontext(t), map[string]any{
		"id": float64(kennung), "fassung": float64(2), "name": "Anders",
	})
	if fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	gelesen, ist := geaendert.(*app.FormatDTO)
	if !ist {
		t.Fatalf("unerwartete Antwort: %T", geaendert)
	}

	if gelesen.Mindestgroesse != 33 || gelesen.Spiel != "magic" {
		t.Errorf("die ungenannten Angaben sind mitgegangen: %+v", gelesen)
	}
}

// **Und `ergaenzt` lässt jede genannte Zahl stehen.**
//
// Der Lauf darüber nennt **eine** und prüft, dass die anderen aus dem
// vorhandenen Format kommen. Dieser nennt **alle sieben** — dann darf keine
// übernommen werden, und der Zweig „nicht gesagt" wird kein einziges Mal
// betreten.
func TestEinVollstaendigGenanntesFormatUebernimmtNichts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	werkstatt := app.NeueDeckwerkstatt(frischerBestand(t))

	format, fehler := werkstatt.FormatAusBauart(
		probekontext(t), "commander", "Commander", "magic", "commander")
	if fehler != nil {
		t.Fatalf("das Format: %v", fehler)
	}

	geaendert, aenderfehler := werkstatt.FormatSetzen(probekontext(t), map[string]any{
		"id": float64(kennungAusWerkstatt(t, format)), "fassung": float64(1),
		"name": "Anders", "spiel": "magic", "scryfallname": "duel",
		"mindestgroesse": float64(60), "hoechstgroesse": float64(0),
		"sideboardgroesse": float64(15), "hoechstzahlGleich": float64(4),
	})
	if aenderfehler != nil {
		t.Fatalf("ändern: %v", aenderfehler)
	}

	gelesen, ist := geaendert.(*app.FormatDTO)
	if !ist {
		t.Fatalf("unerwartete Antwort: %T", geaendert)
	}

	// **Nichts von Commander ist geblieben.**
	if gelesen.Mindestgroesse != 60 || gelesen.Hoechstgroesse != 0 ||
		gelesen.Sideboardgroesse != 15 || gelesen.HoechstzahlGleich != 4 ||
		gelesen.Scryfallname != "duel" || gelesen.Name != "Anders" {
		t.Errorf("%+v", gelesen)
	}
}
