package app_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Skriptnaht wertet einen Ausdruck aus** —
// #1070.
//
// §15 nennt MSEs eigene Kommandozeile als **Orakel** dieses Projekts; diese
// Naht beantwortet dieselben Fragen über dieses Programm — über
// Kommandozeile, MCP und REST.
func TestDieSkriptnahtWertetAus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	_, karteID := setMitKarte(t, db)

	naht, fehler := app.SkriptnahtUeber(db)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	ergebnis, auswertfehler := naht.Auswerten(probekontext(t), karteID, `"a" + "b"`)
	if auswertfehler != nil {
		t.Fatalf("Auswerten: %v", auswertfehler)
	}

	// **Der Quelltext, nicht der Text** — so zeigt MSE es.
	if ergebnis.Quelltext != `"ab"` || ergebnis.Text != "ab" {
		t.Errorf("Ergebnis = %+v", ergebnis)
	}

	if ergebnis.Art == "" {
		t.Error("die Art fehlt")
	}
}

// **Ein Zerteilfehler kommt durch — mit seiner Stelle.**
func TestEinZerteilfehlerKommtDurchDieNaht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	_, karteID := setMitKarte(t, db)

	naht, fehler := app.SkriptnahtUeber(db)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	_, auswertfehler := naht.Auswerten(probekontext(t), karteID, `1 +`)
	if auswertfehler == nil {
		t.Fatal("der unvollständige Ausdruck kam ohne Fehler durch")
	}

	if !strings.Contains(auswertfehler.Error(), "Zeile") {
		t.Errorf("Fehler %v — ohne Stelle", auswertfehler)
	}
}

// **Ohne Zuhause kommt die Skriptnaht nicht zusammen.**
func TestOhneZuhauseKommtDieSkriptnahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.SkriptnahtUeber(frischerBestand(t)); fehler == nil {
		t.Fatal("kein Fehler ohne Zuhause")
	}
}

// **Die Naht zählt auch auf, was die Umgebung kennt**
// (#1164).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Die Naht ist der Weg für
// Kommandozeile, MCP und REST — sie reicht durch, damit die Namen dort
// dieselben sind wie am Fenster.
//
// **Was dieser Lauf nicht zeigt**: dass eine Vorlage `card` und `set`
// dazulegt. Dieses Set hat kein Stilpaket, also bleibt die Bibliothek (§8) —
// dass die Liste der **Umgebung** folgt und nicht einer festen Aufzählung,
// steht in `internal/service/skriptnamen_test.go` mit einem Stil daneben.
func TestDieSkriptnahtZaehltDieNamenAuf(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	_, karteID := setMitKarte(t, db)

	naht, fehler := app.SkriptnahtUeber(db)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	namen, namensfehler := naht.Namen(probekontext(t), karteID)
	if namensfehler != nil {
		t.Fatalf("Namen: %v", namensfehler)
	}

	// **Ein Name, den es wirklich gibt** — und der Ausdruck daneben belegt
	// es: die Liste ist kein Wunschzettel.
	if !slices.Contains(namen, "length") {
		t.Fatalf("length fehlt in %v", namen)
	}

	if _, laeuft := naht.Auswerten(probekontext(t), karteID, "length"); laeuft != nil {
		t.Errorf("length steht in der Liste, läuft aber nicht: %v", laeuft)
	}
}

// **Eine Abbildung kommt mit ihren Einträgen durch die Naht** — rekursiv,
// und mit `[]` statt `nil` an jedem Blatt (Werkstück U13 von Phase 5.1,
// #1424).
func TestEineAbbildungKommtMitEintraegenDurchDieNaht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	_, karteID := setMitKarte(t, db)

	naht, fehler := app.SkriptnahtUeber(db)
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	ergebnis, auswertfehler := naht.Auswerten(probekontext(t), karteID, `[a: 1, b: [c: "zwei"]]`)
	if auswertfehler != nil {
		t.Fatalf("Auswerten: %v", auswertfehler)
	}

	if len(ergebnis.Eintraege) != 2 || ergebnis.Eintraege[0].Name != "a" || ergebnis.Eintraege[0].Eintraege == nil {
		t.Fatalf("Einträge: %+v", ergebnis.Eintraege)
	}

	if b := ergebnis.Eintraege[1]; len(b.Eintraege) != 1 || b.Eintraege[0].Quelltext != `"zwei"` {
		t.Errorf("b: %+v", b)
	}
}
