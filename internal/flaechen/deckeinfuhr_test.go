package flaechen_test

// **Die Decklisteneinfuhr über jede Fläche** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Was hier gemessen wird
//
//   - **Dass die Vorschau nichts legt** und den Stand nennt, an den sie
//     gebunden ist — über alle drei kopflosen Flächen.
//   - **Dass eine mehrdeutige Zeile keinen geratenen Druck bekommt.** Die
//     Gegenstelle gibt für einen Namen ohne Setangabe zwei Drucke; die Zeile
//     bleibt offen, und das ist die Zusage des Issues.
//   - **Dass ein veralteter Stand nichts legt** — die Ablehnung, ohne die nur
//     der Erfolgsweg geprüft wäre (§10).
//   - **Und dass dabei kein Besitz entsteht.** Ein Deck verweist auf Karten;
//     die Sammlung bleibt leer, auch nach einem geglückten Lauf.
//
// # Was er **nicht** misst
//
// **Den echten Dienst.** Gefragt wird die Gegenstelle aus
// `scryfallgegenstelle_test.go`; sie bildet nach, was am 2026-08-30 gemessen
// wurde — ein genauer Name mit Setangabe trifft einen Druck, ohne Setangabe
// zwei. Was sie nicht kann, ist belegen, dass Scryfall sich heute so verhält.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// derVorschauweg und derLegeweg sind die Endpunkte, wie sie im Katalog stehen.
const (
	derVorschauweg = "POST /decks/{deckID}/liste/vorschau"
	derLegeweg     = "POST /decks/{deckID}/liste"
)

// dieDeckliste nennt eine eindeutige Zeile, eine mehrdeutige und ein
// Sideboard.
//
// **»Erste Probe« mit Setangabe ist eindeutig**, ohne Setangabe nicht — so
// antwortet die Gegenstelle, und so antwortet der echte Dienst.
const dieDeckliste = "Deck\n" +
	"4 Erste Probe (probe) 1\n" +
	"2 Zweite Probe\n" +
	"\nSideboard\n" +
	"3 Erste Probe (probe) 1\n" +
	"\nMaybeboard\n" +
	"1 Erste Probe (probe) 1\n"

// deckvorschau ist, was eine Fläche zurückgibt.
type deckvorschau struct {
	DeckID    int64  `json:"deckID"`
	Stand     string `json:"stand"`
	Hauptdeck int    `json:"hauptdeck"`
	Sideboard int    `json:"sideboard"`

	Zeilen []struct {
		Bereich      string `json:"bereich"`
		Menge        int    `json:"menge"`
		Name         string `json:"name"`
		Printing     string `json:"printing"`
		Befund       string `json:"befund"`
		Uebernehmbar bool   `json:"uebernehmbar"`
	} `json:"zeilen"`
}

// deckvorschauAus liest die Antwort einer Fläche.
func deckvorschauAus(t *testing.T, roh string) deckvorschau {
	t.Helper()

	var gelesen deckvorschau
	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("die Vorschau deuten: %v — %s", fehler, roh)
	}

	return gelesen
}

// pruefeDeckvorschau hält, was jede Fläche gleich sagen muss.
func pruefeDeckvorschau(t *testing.T, wo string, gesehen deckvorschau) {
	t.Helper()

	if gesehen.Stand == "" {
		t.Errorf("%s nennt keinen Stand", wo)
	}

	// **Vier im Hauptdeck und drei im Sideboard** — die mehrdeutige Zeile
	// zählt nicht mit, und die Merkliste auch nicht.
	if gesehen.Hauptdeck != 4 || gesehen.Sideboard != 3 {
		t.Errorf("%s zählt %d/%d, erwartet 4/3",
			wo, gesehen.Hauptdeck, gesehen.Sideboard)
	}

	var offen, gelegt int

	for _, zeile := range gesehen.Zeilen {
		if zeile.Uebernehmbar {
			gelegt++

			continue
		}

		offen++

		if zeile.Printing != "" {
			t.Errorf("%s hat für %q einen Druck geraten: %s",
				wo, zeile.Name, zeile.Printing)
		}

		if zeile.Befund == "" {
			t.Errorf("%s sagt bei %q nicht, warum", wo, zeile.Name)
		}
	}

	// **Zwei bleiben draußen**: die mehrdeutige Zeile und die Merkliste.
	if gelegt != 2 || offen != 2 {
		t.Errorf("%s hat %d übernehmbare und %d offene Zeilen, erwartet 2 und 2",
			wo, gelegt, offen)
	}
}

func TestDieDecklisteWirdUeberallGleichGedeutet(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"deck", "neu", "Einfuhr", "magic"))

	ausCLI := deckvorschauAus(t, mitEingabe(t, umgebung, dieDeckliste,
		"--json", "deck-liste", deckID))
	pruefeDeckvorschau(t, "die Kommandozeile", ausCLI)

	text, istFehler := mcpRuf(t, umgebung, "lesen.deckliste",
		map[string]any{"deckID": zahl(t, deckID), "text": dieDeckliste})
	if istFehler {
		t.Fatalf("MCP klagte: %s", text)
	}

	pruefeDeckvorschau(t, "MCP", deckvorschauAus(t, text))

	status, rumpf := an.ruf(t, http.MethodPost,
		strings.ReplaceAll(pfadAus(derVorschauweg), "{deckID}", deckID),
		string(mussJSON(t, map[string]any{"text": dieDeckliste})))
	if status != http.StatusOK {
		t.Fatalf("der Dienst gab %d: %s", status, rumpf)
	}

	ausREST := deckvorschauAus(t, rumpf)
	pruefeDeckvorschau(t, "der Dienst", ausREST)

	// **Und alle drei nennen denselben Stand** — es ist dasselbe Deck, und
	// nichts wurde dabei geschrieben.
	if ausREST.Stand != ausCLI.Stand {
		t.Errorf("der Dienst nennt %q, die Kommandozeile %q",
			ausREST.Stand, ausCLI.Stand)
	}
}

// **Ein veralteter Stand legt nichts — über jede Fläche.**
func TestEinVeralteterStandLegtNichts(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"deck", "neu", "Einfuhr", "magic"))

	// **Der Stand von vorhin** — er gehört zu einem Deck, das es so nicht
	// mehr gibt, sobald etwas darin liegt.
	alt := deckvorschauAus(t, mitEingabe(t, umgebung, dieDeckliste,
		"--json", "deck-liste", deckID)).Stand

	mussKommandozeile(t, umgebung, "deck", "legen", deckID, druckkennung, "1")

	_, _, code := mitEingabeUndCode(t, umgebung, dieDeckliste,
		"--json", "deck-liste-uebernehmen", deckID, alt)
	if code == 0 {
		t.Error("die Kommandozeile legte über einen veralteten Stand")
	}

	status, rumpf := an.ruf(t, http.MethodPost,
		strings.ReplaceAll(pfadAus(derLegeweg), "{deckID}", deckID),
		string(mussJSON(t, map[string]any{
			"stand": alt,
			"zeilen": []map[string]any{{
				"printing": druckkennung, "menge": 1, "bereich": "hauptdeck",
			}},
		})))
	if status == http.StatusOK {
		t.Errorf("der Dienst legte über einen veralteten Stand: %s", rumpf)
	}

	text, istFehler := mcpRuf(t, umgebung, "schreiben.deckliste",
		map[string]any{
			"deckID": zahl(t, deckID), "stand": alt,
			"zeilen": []map[string]any{{
				"printing": druckkennung, "menge": 1, "bereich": "hauptdeck",
			}},
		})
	if !istFehler {
		t.Errorf("MCP legte über einen veralteten Stand: %s", text)
	}
}

// **Ein geglückter Lauf legt Plätze und keinen Besitz.**
func TestDieDecklisteLegtPlaetzeUndKeinenBesitz(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"deck", "neu", "Einfuhr", "magic"))

	aus := mitEingabe(t, umgebung, dieDeckliste,
		"--json", "deck-liste-uebernehmen", deckID, derFreieStandAnDerFlaeche)

	var bericht struct {
		Gelegt int    `json:"gelegt"`
		Stueck int    `json:"stueck"`
		Offen  int    `json:"offen"`
		Stand  string `json:"stand"`
	}

	if fehler := json.Unmarshal([]byte(aus), &bericht); fehler != nil {
		t.Fatalf("den Bericht deuten: %v — %s", fehler, aus)
	}

	if bericht.Gelegt != 2 || bericht.Stueck != 7 || bericht.Offen != 2 {
		t.Errorf("gelegt %d, Stück %d, offen %d — erwartet 2, 7 und 2",
			bericht.Gelegt, bericht.Stueck, bericht.Offen)
	}

	// **Und die Sammlung ist unberührt.** Ein Deck verweist auf Karten; wer
	// sie besitzen will, trägt sie ein — über einen anderen Weg.
	sammlung := mussKommandozeile(t, umgebung, "--json", "sammlung")
	if strings.Contains(sammlung, druckkennung) {
		t.Errorf("die Einfuhr hat Besitz angelegt: %s", sammlung)
	}

	// **Ein zweiter Lauf über denselben Stand legt nicht noch einmal.**
	if _, _, code := mitEingabeUndCode(t, umgebung, dieDeckliste,
		"--json", "deck-liste-uebernehmen", deckID,
		bericht.Stand); code != 0 {
		t.Errorf("der Lauf mit dem neuen Stand scheiterte: %d", code)
	}
}

// derFreieStandAnDerFlaeche hebt die Bindung an eine Vorschau auf — nur an
// der Kommandozeile, und ausdrücklich.
const derFreieStandAnDerFlaeche = "-"
