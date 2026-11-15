// **`card.notes` über jede Fläche** — Werkstück 6 von Phase 3.7
// (#1168,
// Befund Z1-01).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier die Zusage ist
//
// **Eine Notiz, geschrieben wie jeder andere Feldwert, kommt an jeder Fläche
// als dieselbe an — und im Skript als `card.notes`.**
//
// Bis zum 2026-08-27 hatte `card.notes` zwei Darstellungen: den Feldwert aus
// der Datei und ein Feld `CardRow.Notes`, das **kein Leser dieses Projekts
// füllte**. Welche galt, entschied eine Reihenfolge. **Über MCP gewann die
// leere**: `lesen.karten` gab an jeder Karte `"Notes": ""`, gleich was in ihr
// stand.
package flaechen_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// **Eine Notiz steht danach an jeder Fläche und im Skript.**
func TestEineNotizErreichtJedeFlaeche(t *testing.T) {
	// **Mit der Saat**, weil `card` in der Konsole erst zur Verfügung steht,
	// wenn der Planbauer einen Stil hat — ohne Stilpaket bleibt die
	// Bibliothek und sonst nichts.
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	// **Vor dem Schreiben gezählt** — die Karte trägt schon ein leeres
	// `notes`, weil MSE es an jeder führt.
	vorher := eintraege(t, umgebung, karteID)

	// **Geschrieben wird wie jeder andere Wert.** `notes` ist in der Datei ein
	// Eintrag des `card:`-Blocks; ein eigener Schreibweg wäre eine zweite
	// Antwort auf dieselbe Frage.
	text, istFehler := mcpRuf(t, umgebung, "schreiben.feld", map[string]any{
		"setID": zahl(t, setID), "karteID": zahl(t, karteID),
		"feld": "notes", "wert": "gemerkt",
	})
	if istFehler {
		t.Fatalf("schreiben.feld: %s", text)
	}

	t.Run("MCP nennt sie am Kartensatz", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.karten",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen.karten: %s", text)
		}

		// **Der Feldname steht in der Antwort als `Notes`** — das ist die
		// Gestalt, die `internal/mcp` zusagt.
		if !strings.Contains(text, `"Notes": "gemerkt"`) {
			t.Errorf("Notes fehlt in der Antwort: %s", text)
		}
	})

	t.Run("Kommandozeile liest sie als Skriptwert", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "skript", karteID, "card.notes")

		if strings.TrimSpace(aus) != `"gemerkt"` {
			t.Errorf("Ausgabe %q", aus)
		}
	})

	// **Und `notes` steht genau einmal in der Abbildung.**
	//
	// Das ist die eigentliche Zusage dieses Werkstücks und die einzige, die
	// den doppelten Eintrag sichtbar macht: `card.notes` läse auch bei zwei
	// Einträgen den ersten und wäre grün.
	//
	// **Gezählt wird gegen die Karte vorher**, nicht gegen eine feste Zahl:
	// die Abbildung trägt mehr als die Feldwerte, und wie viel mehr, ist die
	// Sache des Planbauers. Die Zusage ist, dass eine **geschriebene** Notiz
	// die Zahl nicht ändert — vorher stand `notes` schon als leerer Rückfall
	// darin.
	t.Run("und steht genau einmal in der Abbildung", func(t *testing.T) {
		nachher := eintraege(t, umgebung, karteID)

		if nachher != vorher {
			t.Errorf("vorher %d Einträge, nachher %d — `notes` steht doppelt",
				vorher, nachher)
		}
	})

	t.Run("REST ebenso", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/karten/"+karteID+"/skript",
			`{"ausdruck":"card.notes"}`)
		if status != http.StatusOK {
			t.Fatalf("POST skript: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, `\"gemerkt\"`) {
			t.Errorf("Antwort %s", rumpf)
		}
	})
}

// **Und eine Karte ohne Notiz hat eine leere, keine fehlende.**
//
// MSE führt `notes` an jeder Karte (`REFLECT(notes)` in
// `IMPLEMENT_REFLECTION(Card)`, `src/data/card.cpp`, Zeile 91, Fassung
// `f9d9356` vom 2021-07-06); fünf der neunzehn Exportvorlagen des Packs lesen
// es, und `magic.mse-game` zieht daraus die Kartennummer.
func TestEineKarteOhneNotizGibtEinenLeerenText(t *testing.T) {
	// **Mit der Saat**, weil `card` in der Konsole erst zur Verfügung steht,
	// wenn der Planbauer einen Stil hat — ohne Stilpaket bleibt die
	// Bibliothek und sonst nichts.
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	aus := mussKommandozeile(t, umgebung, "skript", karteID, "card.notes")

	if strings.TrimSpace(aus) != `""` {
		t.Errorf("Ausgabe %q, erwartet einen leeren Text", aus)
	}
}

// eintraege zählt die Einträge der Kartenabbildung im Skript.
func eintraege(t *testing.T, umgebung []string, karteID string) int {
	t.Helper()

	aus := strings.TrimSpace(
		mussKommandozeile(t, umgebung, "skript", karteID, "number_of_items(in: card)"))

	zahl, err := strconv.Atoi(aus)
	if err != nil {
		t.Fatalf("keine Zahl: %q", aus)
	}

	return zahl
}
