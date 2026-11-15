package main

// **Decks an der Kommandozeile** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Die fachliche Zusage prüft `internal/flaechen`** — dort läuft derselbe
// Ablauf über alle drei Flächen. Hier stehen die **Bedienfehler** und die
// Gestalt der Ausgabe.

import (
	"strings"
	"testing"
)

// **Der ganze Weg, in einem Lauf.**
//
// **Zusammen und nicht einzeln**: jeder Schritt braucht den vorigen, und
// zwölf Läufe, die jeder ein Deck anlegen, prüften elfmal dasselbe.
func TestDerDeckwegAnDerKommandozeile(t *testing.T) {
	ort, _ := eingefahren(t)

	deckID := strings.TrimSpace(mussLaufen(t, ort, "deck", "neu", "Aggro", "magic"))

	if deckID == "" {
		t.Fatal("das angelegte Deck hat keine Kennung")
	}

	// **Die Liste nennt es.**
	if aus := mussLaufen(t, ort, "decks"); !strings.Contains(aus, "Aggro") {
		t.Errorf("die Liste nennt das Deck nicht: %q", aus)
	}

	// **Und die Liste eines anderen Spiels nicht.**
	if aus := mussLaufen(t, ort, "decks", "vs"); strings.Contains(aus, "Aggro") {
		t.Errorf("ein Magic-Deck stand unter vs: %q", aus)
	}

	fassung := "1"

	mussLaufen(t, ort, "deck", "umbenennen", deckID, fassung, "Midrange", "Notiz")

	if aus := mussLaufen(t, ort, "deck", deckID); !strings.Contains(aus, "Midrange") {
		t.Errorf("das Deck heißt noch anders: %q", aus)
	}

	platzID := strings.TrimSpace(mussLaufen(t, ort, "deck", "legen",
		deckID, "a1b2c3d4-0000-4000-8000-00000000000a", "4"))

	if aus := mussLaufen(t, ort, "deck", deckID); !strings.Contains(aus, "hauptdeck") {
		t.Errorf("der Platz fehlt: %q", aus)
	}

	formatID := strings.TrimSpace(mussLaufen(t, ort,
		"format", "neu", "konstruiert", "Konstruiert", "magic"))

	if aus := mussLaufen(t, ort, "formate"); !strings.Contains(aus, "Konstruiert") {
		t.Errorf("das Format fehlt: %q", aus)
	}

	if aus := mussLaufen(t, ort, "format", "bauarten"); !strings.Contains(aus, "commander") {
		t.Errorf("die Bauarten fehlen: %q", aus)
	}

	mussLaufen(t, ort, "deck", "format", deckID, formatID)

	// **Vier Karten in einem 60-Karten-Format sind ein fachliches Nein.**
	if code, _, _ := fahreAn(t, ort, "deck", "pruefen", deckID); code != CodeFachlich {
		t.Errorf("prüfen gab %d, erwartet %d", code, CodeFachlich)
	}

	// **Und die leere Sammlung ebenso.**
	if code, _, _ := fahreAn(t, ort, "deck", "bestand", deckID); code != CodeFachlich {
		t.Errorf("bestand gab %d, erwartet %d", code, CodeFachlich)
	}

	mussLaufen(t, ort, "format", "setzen",
		formatID, "1", "Konstruiert", "magic", "40", "0", "15", "4", "standard")

	mussLaufen(t, ort, "deck", "nehmen", deckID, platzID)
	mussLaufen(t, ort, "deck", "format", deckID)
	mussLaufen(t, ort, "deck", "loeschen", deckID)
	mussLaufen(t, ort, "format", "loeschen", formatID)
}

// **Und derselbe Weg als JSON.**
//
// **Ein eigener Lauf**, weil `--json` einen anderen Zweig nimmt: die Ausgabe
// entsteht aus dem DTO und nicht aus einer Zeile.
func TestDerDeckwegAlsJSON(t *testing.T) {
	ort, _ := eingefahren(t)

	deckID := kennungAusJSON(t, mussLaufen(t, ort,
		"deck", "neu", "Aggro", "magic", "--json"))

	for _, argumente := range [][]string{
		{"decks", "--json"},
		{"deck", deckID, "--json"},
		{"formate", "--json"},
		{"format", "bauarten", "--json"},
	} {
		if aus := mussLaufen(t, ort, argumente...); !strings.HasPrefix(aus, "{") {
			t.Errorf("%v gab kein JSON: %q", argumente, aus)
		}
	}

	mussLaufen(t, ort, "deck", "legen", deckID,
		"a1b2c3d4-0000-4000-8000-00000000000a", "1", "sideboard", "--json")

	formatID := kennungAusJSON(t, mussLaufen(t, ort,
		"format", "neu", "konstruiert", "K", "magic", "", "--json"))

	mussLaufen(t, ort, "deck", "format", deckID, formatID)

	// **Der Rückgabewert sagt das Urteil, auch als JSON.**
	if code, _, _ := fahreAn(t, ort, "deck", "pruefen", deckID, "--json"); code != CodeFachlich {
		t.Errorf("prüfen gab %d, erwartet %d", code, CodeFachlich)
	}

	if code, _, _ := fahreAn(t, ort, "deck", "bestand", deckID, "--json"); code != CodeFachlich {
		t.Errorf("bestand gab %d, erwartet %d", code, CodeFachlich)
	}
}

// **Eine eigene Karte lässt sich über `--karte=` legen — wenn das Spiel
// stimmt.**
//
// **Das Probeset gehört zum Spiel `probe`**, und genau daran zeigt sich die
// Regel: dasselbe Deck nimmt die Karte an, sobald es zu demselben Spiel
// gehört, und weist sie sonst ab.
func TestEineEigeneKarteUeberDieKommandozeile(t *testing.T) {
	ort, _, karteID := eingefahrenesSet(t)

	fremd := strings.TrimSpace(mussLaufen(t, ort, "deck", "neu", "Fremd", "magic"))

	if code, _, meldung := fahreAn(t, ort, "deck", "legen",
		fremd, "--karte="+karteID, "1"); code != CodeFachlich {
		t.Errorf("die fremde Karte kam durch: %d %s", code, meldung)
	}

	eigen := strings.TrimSpace(mussLaufen(t, ort, "deck", "neu", "Eigen", "probe"))

	if aus := mussLaufen(t, ort, "deck", "legen",
		eigen, "--karte="+karteID, "1"); strings.TrimSpace(aus) == "" {
		t.Error("der Platz hat keine Kennung")
	}

	if aus := mussLaufen(t, ort, "deck", eigen); !strings.Contains(aus, "karte:") {
		t.Errorf("die eigene Karte fehlt: %q", aus)
	}
}

// **Eine Kopie an der Kommandozeile** — DOM-03.AK-02
// (#1698):
// die Kennung der Kopie als Zeile oder als JSON, ein fehlender Name ist ein
// Bedienfehler, ein vergebener ein fachliches Nein.
func TestEinDeckLaesstSichAnDerKommandozeileKopieren(t *testing.T) {
	ort, _ := eingefahren(t)

	deckID := strings.TrimSpace(mussLaufen(t, ort, "deck", "neu", "Aggro", "magic"))

	kopieID := strings.TrimSpace(mussLaufen(t, ort, "deck", "kopieren", deckID, "Aggro II"))
	if kopieID == "" || kopieID == deckID {
		t.Fatalf("die Kopie hat die Kennung %q", kopieID)
	}

	if aus := mussLaufen(t, ort, "deck", kopieID); !strings.Contains(aus, "Aggro II") {
		t.Errorf("die Kopie heißt anders: %q", aus)
	}

	if aus := mussLaufen(t, ort, "deck", "kopieren", deckID, "Aggro III", "--json"); !strings.HasPrefix(aus, "{") {
		t.Errorf("als JSON kam %q", aus)
	}

	if code, _, _ := fahreAn(t, ort, "deck", "kopieren", deckID); code != CodeBedienung {
		t.Errorf("ohne Namen gab es %d, erwartet %d", code, CodeBedienung)
	}

	if code, _, _ := fahreAn(t, ort, "deck", "kopieren", deckID, "Aggro"); code != CodeFachlich {
		t.Errorf("ein vergebener Name gab %d, erwartet %d", code, CodeFachlich)
	}
}

// **Was fehlt, auf die Wunschliste** — DOM-03.AK-02 (#1698): beim ersten Mal
// neu, beim zweiten schon gewünscht, und beides ist ein Erfolg.
func TestFehlkartenKommenAnDerKommandozeileAufDieWunschliste(t *testing.T) {
	ort, _ := eingefahren(t)

	deckID := strings.TrimSpace(mussLaufen(t, ort, "deck", "neu", "Aggro", "magic"))
	mussLaufen(t, ort, "deck", "legen", deckID, "a1b2c3d4-0000-4000-8000-00000000000a", "4")

	if aus := mussLaufen(t, ort, "deck", "wuenschen", deckID); !strings.HasPrefix(aus, "neu\t4×") {
		t.Errorf("beim ersten Mal kam %q", aus)
	}

	if aus := mussLaufen(t, ort, "deck", "wuenschen", deckID); !strings.HasPrefix(aus, "schon gewünscht\t4 von 4") {
		t.Errorf("beim zweiten Mal kam %q", aus)
	}

	if code, _, _ := fahreAn(t, ort, "deck", "wuenschen"); code != CodeBedienung {
		t.Errorf("ohne Kennung gab es %d, erwartet %d", code, CodeBedienung)
	}

	if code, _, _ := fahreAn(t, ort, "deck", "wuenschen", "999"); code != CodeFachlich {
		t.Errorf("ein fehlendes Deck gab %d, erwartet %d", code, CodeFachlich)
	}
}

// mussLaufen ruft `mse` und bricht ab, wenn es nicht glückt.
func mussLaufen(t *testing.T, ort string, argumente ...string) string {
	t.Helper()

	code, aus, fehler := fahreAn(t, ort, argumente...)
	if code != CodeGut {
		t.Fatalf("mse %v: %d %s%s", argumente, code, aus, fehler)
	}

	return aus
}

// kennungAusJSON liest die Kennung aus einer JSON-Antwort.
func kennungAusJSON(t *testing.T, roh string) string {
	t.Helper()

	// **Mit Leerzeichen** — die JSON-Ausgabe dieses Programms ist eingerückt.
	_, nachher, gefunden := strings.Cut(roh, `"id": "`)
	if !gefunden {
		t.Fatalf("die Antwort nennt keine Kennung: %q", roh)
	}

	kennung, _, _ := strings.Cut(nachher, `"`)

	return kennung
}
