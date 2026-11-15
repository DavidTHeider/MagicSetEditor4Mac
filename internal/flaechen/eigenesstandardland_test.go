package flaechen_test

// **Ein Deck aus eigenen Standardländern geht durch** —
// #1622.
//
// # Warum dieser Lauf an der Fläche steht und nicht nur im Dienst
//
// **Weil der Befund eine Naht war und keine Rechnung.** `IstStandardland`
// war richtig, `TypzeileText` gab es noch nicht, und die Typzeile kam
// trotzdem nie an: die Prüfung fragte sie **nur** bei Scryfall, und zwei
// Composition Roots (§2) müssen beide verdrahtet sein. Ein Lauf im Dienst
// prüft die Rechnung; dieser prüft, dass sie **jemand aufruft**.

import (
	"encoding/json"
	"strings"
	"testing"
)

// **Zwanzig eigene Gebirge sind kein Verstoß** — CR 100.2a mit CR 205.4a.
func TestZwanzigEigeneGebirgeGehenAnDerFlaecheDurch(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	karte := eineEigeneKarte(t, umgebung, "Basic Land", "Mountain")

	pruefeOhneGleichenbefund(t, deckAusEigenerKarte(t, umgebung, karte))
}

// **Und zwanzig Dryad Arbor sind einer** — CR 205.4c: der Untertyp `Forest`
// macht aus einer Kreatur kein Standardland.
//
// **Die Gegenprobe.** Ohne sie sagte der Lauf darüber nur, dass irgendetwas
// durchgeht — etwa eine Prüfung, die gar nicht mehr zählt (§10).
func TestZwanzigEigeneDryadArborBleibenEinVerstoss(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	karte := eineEigeneKarte(t, umgebung, "Legendary Creature Land",
		"Dryad Forest")

	roh := deckAusEigenerKarte(t, umgebung, karte)

	if !strings.Contains(gleichenbefund(t, roh), "gleiche") {
		t.Errorf("zwanzig gleiche Nichtländer gingen durch: %s", roh)
	}
}

// eineEigeneKarte legt ein Set mit einer Karte an und setzt ihre Typfelder.
func eineEigeneKarte(t *testing.T, umgebung []string, ober, unter string) string {
	t.Helper()

	set := kennungAusFeld(t, mussKommandozeile(t, umgebung,
		"set-neu", "Eigenes", "probe", "probe-schlicht", "--json"))

	karte := kennungAusFeld(t, mussKommandozeile(t, umgebung,
		"karte", set, "anlegen", "--json"))

	mussKommandozeile(t, umgebung, "feld", set, karte, "super type", ober)
	mussKommandozeile(t, umgebung, "feld", set, karte, "sub type", unter)

	return karte
}

// deckAusEigenerKarte baut ein Deck mit zwanzig Stück und prüft es.
func deckAusEigenerKarte(t *testing.T, umgebung []string, karte string) string {
	t.Helper()

	deck := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"deck", "neu", "Länder", "probe"))

	format := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"format", "neu", "konstruiert", "Konstruiert", "probe"))

	mussKommandozeile(t, umgebung, "deck", "format", deck, format)
	mussKommandozeile(t, umgebung, "deck", "legen", deck,
		"--karte="+karte, "20", "hauptdeck")

	// **Nicht `mussKommandozeile`**: `deck pruefen` gibt **1** zurück, wenn
	// das Deck unzulässig ist — und zwanzig Karten sind in einem Format mit
	// Mindestgröße keine sechzig. Der Rückgabewert ist hier eine Aussage
	// über die **Größe**; gefragt ist die über **gleiche Karten**.
	roh, fehler, code := kommandozeile(t, umgebung,
		"deck", "pruefen", deck, "--json")
	if code > 1 {
		t.Fatalf("deck pruefen: %d %s%s", code, roh, fehler)
	}

	return roh
}

// pruefeOhneGleichenbefund sichert zu, dass nichts über gleiche Karten steht.
func pruefeOhneGleichenbefund(t *testing.T, roh string) {
	t.Helper()

	if feld := gleichenbefund(t, roh); feld != "" {
		t.Errorf("die Gebirge wurden beanstandet: %s", roh)
	}
}

// gleichenbefund gibt das Feld des Größenbefunds über gleiche Karten.
func gleichenbefund(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Urteil struct {
			Groessen []struct {
				Feld string `json:"feld"`
			} `json:"groessen"`
		} `json:"urteil"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	for _, befund := range gelesen.Urteil.Groessen {
		if befund.Feld == "gleiche" {
			return befund.Feld
		}
	}

	return ""
}

// kennungAusFeld liest die Kennung aus einer JSON-Antwort.
//
// **Zwei Feldnamen, und das ist kein Versehen**: ein Set antwortet mit `id`,
// eine angelegte Karte mit `karteID`. Wer nur den ersten läse, bekäme eine
// leere Kennung und einen Fehler drei Schritte später.
func kennungAusFeld(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		ID      string `json:"id"`
		KarteID string `json:"karteID"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if gelesen.ID != "" {
		return gelesen.ID
	}

	if gelesen.KarteID == "" {
		t.Fatalf("keine Kennung in der Antwort: %s", roh)
	}

	return gelesen.KarteID
}
